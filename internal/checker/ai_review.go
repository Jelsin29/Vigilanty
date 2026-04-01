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
	passPatterns    []*regexp.Regexp
	failPatterns    []*regexp.Regexp
}

func newAIReviewChecker(cfg map[string]interface{}) (Checker, error) {
	provider, err := stringConfigValue(cfg, "provider")
	if err != nil {
		return nil, fmt.Errorf("ai review provider: %w", err)
	}

	timeout, err := durationConfigValue(cfg, "timeout", defaultAIReviewTimeout)
	if err != nil {
		return nil, fmt.Errorf("ai review timeout: %w", err)
	}

	prompt, err := optionalStringConfigValue(cfg, "prompt")
	if err != nil {
		return nil, fmt.Errorf("ai review prompt: %w", err)
	}

	rulesFile, err := optionalStringConfigValue(cfg, "rules_file")
	if err != nil {
		return nil, fmt.Errorf("ai review rules_file: %w", err)
	}

	model, err := optionalStringConfigValue(cfg, "model")
	if err != nil {
		return nil, fmt.Errorf("ai review model: %w", err)
	}

	maxDiffLines, err := intConfigValue(cfg, "max_diff_lines", 500)
	if err != nil {
		return nil, fmt.Errorf("ai review max_diff_lines: %w", err)
	}

	skipOnEmptyDiff, err := boolConfigValue(cfg, "skip_on_empty_diff", false)
	if err != nil {
		return nil, fmt.Errorf("ai review skip_on_empty_diff: %w", err)
	}

	passPatterns, err := regexConfigValue(cfg, "pass_pattern", "pass_patterns", defaultAIPassPatterns)
	if err != nil {
		return nil, fmt.Errorf("ai review pass_pattern: %w", err)
	}

	failPatterns, err := regexConfigValue(cfg, "fail_pattern", "fail_patterns", defaultAIFailPatterns)
	if err != nil {
		return nil, fmt.Errorf("ai review fail_pattern: %w", err)
	}

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

	cliName := c.providerCLIName()
	if _, err := exec.LookPath(cliName); err != nil {
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

	cmd := exec.CommandContext(execCtx, cliName, c.providerArgs(prompt)...)
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
		if strings.TrimSpace(result.Output) == "" {
			result.Output = fmt.Sprintf("AI review timed out after %s", c.timeout)
		} else {
			result.Output = strings.TrimSpace(result.Output) + fmt.Sprintf("\nAI review timed out after %s", c.timeout)
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

	hasPass := matchesAny(c.passPatterns, response)
	hasFail := matchesAny(c.failPatterns, response)

	switch {
	case hasPass && hasFail:
		result.Status = Error
		result.Output = strings.TrimSpace(result.Output) + "\nambiguous response: matched both pass and fail patterns"
	case hasPass:
		result.Status = Passed
	case hasFail:
		result.Status = Failed
	default:
		result.Status = Error
		if strings.TrimSpace(result.Output) == "" {
			result.Output = "ambiguous response: no pass/fail patterns matched"
		} else {
			result.Output = strings.TrimSpace(result.Output) + "\nambiguous response: no pass/fail patterns matched"
		}
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

func (c *AIReviewChecker) providerCLIName() string {
	return c.provider
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

func matchesAny(patterns []*regexp.Regexp, value string) bool {
	for _, pattern := range patterns {
		if pattern.MatchString(value) {
			return true
		}
	}
	return false
}

func regexListConfigValue(cfg map[string]interface{}, key string, fallback []string) ([]*regexp.Regexp, error) {
	raw, ok := cfg[key]
	if !ok || raw == nil {
		return compileRegexes(fallback)
	}

	var values []string
	switch typed := raw.(type) {
	case string:
		if strings.TrimSpace(typed) != "" {
			values = []string{typed}
		}
	case []string:
		values = typed
	case []interface{}:
		values = make([]string, 0, len(typed))
		for _, item := range typed {
			text, ok := item.(string)
			if !ok {
				return nil, fmt.Errorf("%q entries must be strings", key)
			}
			values = append(values, text)
		}
	default:
		return nil, fmt.Errorf("%q must be a string or list of strings", key)
	}

	if len(values) == 0 {
		return compileRegexes(fallback)
	}

	return compileRegexes(values)
}

func regexConfigValue(cfg map[string]interface{}, singularKey string, listKey string, fallback []string) ([]*regexp.Regexp, error) {
	if raw, ok := cfg[singularKey]; ok && raw != nil {
		value, ok := raw.(string)
		if !ok {
			return nil, fmt.Errorf("%q must be a string", singularKey)
		}
		if strings.TrimSpace(value) == "" {
			return compileRegexes(fallback)
		}
		return compileRegexes([]string{value})
	}

	return regexListConfigValue(cfg, listKey, fallback)
}

func compileRegexes(values []string) ([]*regexp.Regexp, error) {
	patterns := make([]*regexp.Regexp, 0, len(values))
	for _, value := range values {
		compiled, err := regexp.Compile(value)
		if err != nil {
			return nil, fmt.Errorf("compile regex %q: %w", value, err)
		}
		patterns = append(patterns, compiled)
	}
	return patterns, nil
}

func optionalStringConfigValue(cfg map[string]interface{}, key string) (string, error) {
	raw, ok := cfg[key]
	if !ok || raw == nil {
		return "", nil
	}

	value, ok := raw.(string)
	if !ok {
		return "", fmt.Errorf("%q must be a string", key)
	}

	return strings.TrimSpace(value), nil
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
			if fileExists(candidate) {
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

func fileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}
