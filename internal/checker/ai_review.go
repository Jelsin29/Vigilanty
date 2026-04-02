package checker

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/jelsin29/vigilanty/internal/rules"
)

const defaultAIReviewTimeout = 120 * time.Second

var (
	defaultAIPassPatterns = []string{`(?i)STATUS:\s*PASSED`, `(?i)\bPASS\b`}
	defaultAIFailPatterns = []string{`(?i)STATUS:\s*FAILED`, `(?i)\bFAIL\b`}
	aiInstallURLs         = map[string]string{
		"claude":   "https://docs.anthropic.com/en/docs/claude-code",
		"gemini":   "https://github.com/google-gemini/gemini-cli",
		"ollama":   "https://ollama.com/download",
		"codex":    "https://platform.openai.com/docs/codex/cli",
		"opencode": "https://github.com/sst/opencode",
		"lmstudio": "https://lmstudio.ai/",
		"github":   "https://cli.github.com/",
	}

	claudeOutputFormatSupport sync.Once
	claudeHasOutputFormat     bool
)

type AIReviewChecker struct {
	provider        string
	model           string
	prompt          string
	rulesFile       string
	timeout         time.Duration
	maxDiffLines    int
	skipOnEmptyDiff bool
	passPatterns    []string
	failPatterns    []string
	passRegexps     []*regexp.Regexp
	failRegexps     []*regexp.Regexp
}

func newAIReviewChecker(cfg map[string]interface{}) (Checker, error) {
	var model, prompt, rulesFile, provider string
	var ok bool

	if provider, ok = cfg["provider"].(string); !ok || strings.TrimSpace(provider) == "" {
		return nil, fmt.Errorf("ai review provider is required and must be a string")
	}
	if model, ok = cfg["model"].(string); !ok {
		model = ""
	}
	if prompt, ok = cfg["prompt"].(string); !ok {
		prompt = ""
	}
	if rulesFile, ok = cfg["rules_file"].(string); !ok {
		rulesFile = ""
	}
	timeout, err := durationConfigValue(cfg, "timeout", defaultAIReviewTimeout)
	if err != nil {
		return nil, fmt.Errorf("ai review timeout: %w", err)
	}

	maxDiffLines, err := intConfigValue(cfg, "max_diff_lines", 500)
	if err != nil {
		return nil, fmt.Errorf("ai review max_diff_lines: %w", err)
	}

	skipOnEmptyDiff, err := boolConfigValue(cfg, "skip_on_empty_diff", false)
	if err != nil {
		return nil, fmt.Errorf("ai review skip_on_empty_diff: %w", err)
	}

	passPatterns := listStringConfigValue(cfg, "pass_pattern", defaultAIPassPatterns)
	failPatterns := listStringConfigValue(cfg, "fail_pattern", defaultAIFailPatterns)
	passRegexps, err := compilePatterns(passPatterns, "pass_pattern")
	if err != nil {
		return nil, err
	}
	failRegexps, err := compilePatterns(failPatterns, "fail_pattern")
	if err != nil {
		return nil, err
	}

	provider = strings.ToLower(strings.TrimSpace(provider))
	switch provider {
	case "claude", "gemini", "ollama", "codex", "opencode", "lmstudio", "github":
	default:
		return nil, fmt.Errorf("unsupported provider %q", provider)
	}

	if provider == "ollama" && model == "" {
		return nil, fmt.Errorf("ai review model: %q is required for provider %q", "model", provider)
	}

	return &AIReviewChecker{
		provider:        provider,
		model:           model,
		prompt:          prompt,
		rulesFile:       rulesFile,
		timeout:         timeout,
		maxDiffLines:    maxDiffLines,
		skipOnEmptyDiff: skipOnEmptyDiff,
		passPatterns:    passPatterns,
		failPatterns:    failPatterns,
		passRegexps:     passRegexps,
		failRegexps:     failRegexps,
	}, nil
}

func (c *AIReviewChecker) Name() string {
	return "ai-review"
}

func (c *AIReviewChecker) Check(ctx CheckContext) CheckResult {
	startedAt := time.Now()

	if c.skipOnEmptyDiff && strings.TrimSpace(ctx.Diff) == "" {
		return CheckResult{
			Status:   Skipped,
			Output:   "skipped (empty diff)",
			Duration: time.Since(startedAt),
		}
	}

	if _, err := exec.LookPath(c.provider); err != nil {
		return CheckResult{
			Status:   Error,
			Output:   fmt.Sprintf("AI CLI not found: %s. Install it from %s", c.provider, aiInstallURLs[c.provider]),
			Duration: time.Since(startedAt),
		}
	}

	prompt, err := c.buildPrompt(ctx)
	if err != nil {
		return CheckResult{
			Status:   Error,
			Output:   fmt.Sprintf("build AI review prompt: %v", err),
			Duration: time.Since(startedAt),
		}
	}

	execCtx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	cmd := exec.CommandContext(execCtx, c.provider, c.providerArgs()...)
	cmd.Stdin = strings.NewReader(prompt)
	if strings.TrimSpace(ctx.Root) != "" {
		cmd.Dir = ctx.Root
	}

	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &output

	err = cmd.Run()
	response := output.String()
	result := CheckResult{
		Output:   response,
		Duration: time.Since(startedAt),
	}

	if errors.Is(execCtx.Err(), context.DeadlineExceeded) {
		result.Status = Error
		trimmed := strings.TrimSpace(result.Output)
		if trimmed == "" {
			result.Output = fmt.Sprintf("AI review timed out after %s", c.timeout)
		} else {
			result.Output = trimmed + fmt.Sprintf("\nAI review timed out after %s", c.timeout)
		}
		return result
	}

	if err != nil {
		result.Status = Error
		if strings.TrimSpace(result.Output) == "" {
			result.Output = fmt.Sprintf("execute AI review with %s: %v", c.provider, err)
		} else {
			result.Output = strings.TrimSpace(result.Output) + fmt.Sprintf("\nexecute AI review with %s: %v", c.provider, err)
		}
		return result
	}

	result.Output = stripMarkdown(strings.TrimSpace(result.Output))

	hasPass := matchesAny(c.passRegexps, result.Output)
	hasFail := matchesAny(c.failRegexps, result.Output)

	switch {
	case hasPass && hasFail:
		result.Status = Error
		result.Output = strings.TrimSpace(result.Output) + "\nambiguous response: matched both pass and fail patterns"
		return result
	case hasPass:
		result.Status = Passed
		return result
	case hasFail:
		result.Status = Failed
		return result
	case !hasPass && !hasFail:
		result.Status = Error
		if strings.TrimSpace(result.Output) == "" {
			result.Output = fmt.Sprintf("ambiguous response: no pass/fail patterns matched regex patterns %s for pass and %s for fail", c.passPatterns, c.failPatterns)
		} else {
			result.Output = fmt.Sprintf("ambiguous response: no pass/fail patterns matched. Regex Pattern (%s, %s, %v, %v)\nOutput: %s", c.passPatterns, c.failPatterns, hasPass, hasFail, strings.TrimSpace(result.Output))
		}
		return result
	}

	panic("unreachable")
}

func (c *AIReviewChecker) buildPrompt(ctx CheckContext) (string, error) {
	rulesContent, hasRules := rules.Discover(ctx.Root, c.rulesFile)
	truncatedDiff := truncateDiffLines(ctx.Diff, c.maxDiffLines)
	if strings.TrimSpace(truncatedDiff) == "" {
		truncatedDiff = "(empty diff)"
	}

	if hasRules {
		var builder strings.Builder
		builder.WriteString("## Code Review Rules\n")
		builder.WriteString(strings.TrimSpace(rulesContent))
		builder.WriteString("\n\n## Task\n")
		builder.WriteString(strings.TrimSpace(c.prompt))
		builder.WriteString("\n\n## Diff to Review\n")
		builder.WriteString(truncatedDiff)
		if !strings.HasSuffix(truncatedDiff, "\n") {
			builder.WriteString("\n")
		}
		return builder.String(), nil
	}

	var builder strings.Builder
	builder.WriteString("Review the staged changes and return a final verdict using STATUS: PASSED or STATUS: FAILED.\n")

	if strings.TrimSpace(c.prompt) != "" {
		builder.WriteString("\nAdditional Instructions:\n")
		builder.WriteString(strings.TrimSpace(c.prompt))
		builder.WriteString("\n")
	}

	builder.WriteString("\nGit Diff:\n")
	builder.WriteString(truncatedDiff)
	if !strings.HasSuffix(truncatedDiff, "\n") {
		builder.WriteString("\n")
	}

	return builder.String(), nil
}

// providerArgs returns CLI flags for the provider. The prompt itself
// gets piped through stdin to avoid hitting ARG_MAX on large diffs.
func (c *AIReviewChecker) providerArgs() []string {
	switch c.provider {
	case "claude":
		args := []string{"-p", "-"}
		if claudeSupportsOutputFormat() {
			args = append(args, "--output-format", "text")
		}
		return args
	case "gemini":
		return []string{"-p", "-"}
	case "ollama":
		return []string{"run", c.model}
	default:
		return []string{"-p", "-"}
	}
}

func claudeSupportsOutputFormat() bool {
	claudeOutputFormatSupport.Do(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		output, err := exec.CommandContext(ctx, "claude", "--help").CombinedOutput()
		if err == nil && strings.Contains(string(output), "--output-format") {
			claudeHasOutputFormat = true
		}
	})

	return claudeHasOutputFormat
}

func matchesAny(patterns []*regexp.Regexp, text string) bool {
	for _, re := range patterns {
		if re.MatchString(text) {
			return true
		}
	}
	return false
}

// stripANSI removes ANSI escape sequences from CLI output.
// We only strip escapes here — markdown bold/italic removal was eating
// legit characters like Go pointers (*T) and underscored identifiers.
var ansiRe = regexp.MustCompile(`\x1b\[[0-9;?]*[a-zA-Z]|\x1b\].*?\x1b\\|\x1b[^[\]]`)

func stripMarkdown(s string) string {
	s = ansiRe.ReplaceAllString(s, "")
	// only strip paired markdown emphasis, not bare chars
	s = strings.ReplaceAll(s, "**", "")
	s = strings.ReplaceAll(s, "__", "")
	s = strings.ReplaceAll(s, "\u00a0", " ")
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	return s
}

func listStringConfigValue(cfg map[string]interface{}, key string, fallback []string) []string {
	raw, ok := cfg[key]
	if !ok || raw == nil {
		return fallback
	}

	switch value := raw.(type) {
	case string:
		value = strings.TrimSpace(value)
		if value == "" {
			return fallback
		}
		return []string{value}
	case []string:
		if len(value) == 0 {
			return fallback
		}
		return value
	case []interface{}:
		patterns := make([]string, 0, len(value))
		for _, item := range value {
			text, ok := item.(string)
			if !ok {
				continue
			}
			text = strings.TrimSpace(text)
			if text == "" {
				continue
			}
			patterns = append(patterns, text)
		}
		if len(patterns) == 0 {
			return fallback
		}
		return patterns
	default:
		return fallback
	}
}

func compilePatterns(patterns []string, key string) ([]*regexp.Regexp, error) {
	compiled := make([]*regexp.Regexp, 0, len(patterns))
	for _, pattern := range patterns {
		re, err := regexp.Compile(pattern)
		if err != nil {
			return nil, fmt.Errorf("ai review %s: %w", key, err)
		}
		compiled = append(compiled, re)
	}
	return compiled, nil
}

func intConfigValue(cfg map[string]interface{}, key string, fallback int) (int, error) {
	raw, ok := cfg[key]
	if !ok || raw == nil {
		return fallback, nil
	}

	switch value := raw.(type) {
	case int:
		return value, nil
	case int64:
		return int(value), nil
	case float64:
		if value != float64(int(value)) {
			return 0, fmt.Errorf("%q must be an integer", key)
		}
		return int(value), nil
	default:
		return 0, fmt.Errorf("%q must be an integer", key)
	}
}

func truncateDiffLines(diff string, limit int) string {
	if limit <= 0 {
		return diff
	}

	lines := strings.Split(diff, "\n")
	if len(lines) <= limit {
		return diff
	}

	return strings.Join(lines[:limit], "\n") + fmt.Sprintf("\n[diff truncated at %d lines]", limit)
}
