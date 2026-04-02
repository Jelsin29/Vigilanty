package checker

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
)

const defaultAIReviewTimeout = 120 * time.Second

var (
	defaultAIPassPatterns = []string{`(?i)STATUS:\s*PASSED`, `(?i)\bPASS\b`}
	defaultAIFailPatterns = []string{`(?i)STATUS:\s*FAILED`, `(?i)\bFAIL\b`}
	aiInstallURLs         = map[string]string{
		"claude": "https://docs.anthropic.com/en/docs/claude-code",
		"gemini": "https://github.com/google-gemini/gemini-cli",
		"ollama": "https://ollama.ai/download",
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

	provider = strings.ToLower(strings.TrimSpace(provider))
	switch provider {
	case "claude", "gemini", "ollama":
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

	cmd := exec.CommandContext(execCtx, c.provider, c.providerArgs(prompt)...)
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

	hasPass := matchesAny(c.passPatterns, result.Output)
	hasFail := matchesAny(c.failPatterns, result.Output)

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
			result.Output = fmt.Sprintf("\nambiguous response: no pass/fail patterns matched. Regex Pattern (%s, %s, %v, %v)\nOutput: %s", c.passPatterns, c.failPatterns, hasPass, hasFail, strings.TrimSpace(result.Output))
		}
		return result
	}

	return result
}

func (c *AIReviewChecker) buildPrompt(ctx CheckContext) (string, error) {
	rules, err := c.loadRules(ctx.Root)
	if err != nil {
		return "", err
	}

	var builder strings.Builder
	builder.WriteString("Review the staged changes and return a final verdict using STATUS: PASSED or STATUS: FAILED.\n")

	if rules != "" {
		builder.WriteString("\nRules:\n")
		builder.WriteString(rules)
		if !strings.HasSuffix(rules, "\n") {
			builder.WriteString("\n")
		}
	}

	if strings.TrimSpace(c.prompt) != "" {
		builder.WriteString("\nAdditional Instructions:\n")
		builder.WriteString(strings.TrimSpace(c.prompt))
		builder.WriteString("\n")
	}

	builder.WriteString("\nGit Diff:\n")
	truncatedDiff := truncateDiffLines(ctx.Diff, c.maxDiffLines)
	if strings.TrimSpace(truncatedDiff) == "" {
		builder.WriteString("(empty diff)\n")
	} else {
		builder.WriteString(truncatedDiff)
		if !strings.HasSuffix(truncatedDiff, "\n") {
			builder.WriteString("\n")
		}
	}

	return builder.String(), nil
}

func (c *AIReviewChecker) loadRules(root string) (string, error) {
	if c.rulesFile != "" {
		content, err := os.ReadFile(resolveRulesPath(root, c.rulesFile))
		if err != nil {
			return "", fmt.Errorf("read rules file %q: %w", c.rulesFile, err)
		}
		return strings.TrimSpace(string(content)), nil
	}

	path, ok := discoverAgentsFile(root)
	if !ok {
		return "", nil
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read AGENTS.md %q: %w", path, err)
	}

	return strings.TrimSpace(string(content)), nil
}

func (c *AIReviewChecker) providerArgs(prompt string) []string {
	switch c.provider {
	case "claude":
		args := []string{"-p", prompt}
		if claudeSupportsOutputFormat() {
			args = append(args, "--output-format", "text")
		}
		return args
	case "gemini":
		return []string{"-p", prompt}
	case "ollama":
		return []string{"run", c.model, prompt}
	default:
		return []string{"-p", prompt}
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

func matchesAny(patterns []string, text string) bool {
	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		if re.MatchString(text) {
			return true
		}
	}
	return false
}

func stripMarkdown(s string) string {
	ansi := regexp.MustCompile(`\x1b\[[0-9;?]*[a-zA-Z]|\x1b[^[]`) // THIS CAUSES A TON OF ISSUES AND IMPOSSIBLE TO DEBUG
	s = ansi.ReplaceAllString(s, "")
	s = strings.ReplaceAll(s, "**", "")
	s = strings.ReplaceAll(s, "__", "")
	s = strings.ReplaceAll(s, "*", "")
	s = strings.ReplaceAll(s, "_", "")
	s = strings.ReplaceAll(s, "\u00a0", " ")
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	return s
}

func listStringConfigValue(cfg map[string]interface{}, key string, fallback []string) []string {
	val, ok := cfg[key].([]string)
	if !ok || len(val) == 0 {
		return fallback
	}
	return val
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

func resolveRulesPath(root string, candidate string) string {
	if filepath.IsAbs(candidate) {
		return candidate
	}
	if strings.TrimSpace(root) == "" {
		return candidate
	}
	return filepath.Join(root, candidate)
}

func discoverAgentsFile(root string) (string, bool) {
	searchRoots := []string{}
	if strings.TrimSpace(root) != "" {
		searchRoots = append(searchRoots, root)
	}

	homeDir, err := os.UserHomeDir()
	if err == nil {
		searchRoots = append(searchRoots, filepath.Join(homeDir, ".config", "opencode"))
	}

	for _, start := range searchRoots {
		for dir := start; dir != ""; dir = filepath.Dir(dir) {
			candidate := filepath.Join(dir, "AGENTS.md")

			if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
				return candidate, true
			}

			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
		}
	}

	return "", false
}
