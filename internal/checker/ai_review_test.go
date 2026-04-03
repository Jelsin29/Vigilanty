package checker

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestAIReviewCheckerSkipsOnEmptyDiff(t *testing.T) {
	instance, err := newAIReviewChecker(map[string]interface{}{
		"provider":           "claude",
		"prompt":             "review this diff",
		"skip_on_empty_diff": true,
	})
	if err != nil {
		t.Fatalf("newAIReviewChecker() error = %s", err.Error())
	}

	result := instance.Check(CheckContext{Root: t.TempDir(), Diff: ""})
	if result.Status != Skipped {
		t.Fatalf("Check().Status = %v, want %v", result.Status, Skipped)
	}
}

func TestTruncateDiffLines(t *testing.T) {
	diff := strings.Join([]string{"1", "2", "3", "4"}, "\n")
	got := truncateDiffLines(diff, 2)
	if want := "1\n2\n[diff truncated at 2 lines]"; got != want {
		t.Fatalf("truncateDiffLines() = %q, want %q", got, want)
	}
}

func TestTruncateDiffLinesNoTruncation(t *testing.T) {
	diff := "line1\nline2"
	got := truncateDiffLines(diff, 10)
	if got != diff {
		t.Fatalf("truncateDiffLines() truncated when it shouldn't have")
	}
}

func TestTruncateDiffLinesZeroLimit(t *testing.T) {
	diff := "anything"
	got := truncateDiffLines(diff, 0)
	if got != diff {
		t.Fatalf("truncateDiffLines(0) should return full diff")
	}
}

func TestStripMarkdownPreservesAsterisks(t *testing.T) {
	// Go pointer syntax should survive stripping
	input := "variable *User is a pointer to User"
	got := stripMarkdown(input)
	if !strings.Contains(got, "*User") {
		t.Fatalf("stripMarkdown() ate a bare asterisk: %q", got)
	}
}

func TestStripMarkdownRemovesBoldPairs(t *testing.T) {
	input := "this is **bold** text"
	got := stripMarkdown(input)
	if strings.Contains(got, "**") {
		t.Fatalf("stripMarkdown() should remove **: %q", got)
	}
	if !strings.Contains(got, "bold") {
		t.Fatalf("stripMarkdown() lost the word 'bold': %q", got)
	}
}

func TestStripMarkdownANSI(t *testing.T) {
	// simulate ANSI color code
	input := "\x1b[31mERROR\x1b[0m something"
	got := stripMarkdown(input)
	if strings.Contains(got, "\x1b") {
		t.Fatalf("stripMarkdown() should remove ANSI escapes: %q", got)
	}
	if !strings.Contains(got, "ERROR") {
		t.Fatalf("stripMarkdown() lost content after stripping ANSI: %q", got)
	}
}

func TestProviderArgsClaude(t *testing.T) {
	c := &AIReviewChecker{provider: "claude"}
	args, usesStdin := c.providerArgs("test prompt")
	if !usesStdin {
		t.Fatal("claude should use stdin")
	}
	if len(args) < 2 || args[1] != "-" {
		t.Fatalf("providerArgs() = %v, want '-' as stdin marker", args)
	}
}

func TestProviderArgsOllama(t *testing.T) {
	c := &AIReviewChecker{provider: "ollama", model: "llama3"}
	args, usesStdin := c.providerArgs("test prompt")
	if !usesStdin {
		t.Fatal("ollama should use stdin")
	}
	if args[0] != "run" || args[1] != "llama3" {
		t.Fatalf("providerArgs() = %v, want [run llama3]", args)
	}
	if len(args) > 2 {
		t.Fatalf("providerArgs() has too many args, prompt should go via stdin: %v", args)
	}
}

func TestProviderArgsOpencode(t *testing.T) {
	c := &AIReviewChecker{provider: "opencode", model: "openai/gpt-5.4"}
	args, usesStdin := c.providerArgs("review this diff")
	if usesStdin {
		t.Fatal("opencode should NOT use stdin, prompt goes as positional arg")
	}
	if args[0] != "run" {
		t.Fatalf("providerArgs()[0] = %q, want 'run'", args[0])
	}
	hasModel := false
	for i, arg := range args {
		if arg == "--model" && i+1 < len(args) && args[i+1] == "openai/gpt-5.4" {
			hasModel = true
		}
	}
	if !hasModel {
		t.Fatalf("providerArgs() missing --model flag: %v", args)
	}
	if args[len(args)-1] != "review this diff" {
		t.Fatalf("providerArgs() missing prompt as last arg: %v", args)
	}
}

func TestProviderArgsCodex(t *testing.T) {
	c := &AIReviewChecker{provider: "codex"}
	args, usesStdin := c.providerArgs("review this")
	if usesStdin {
		t.Fatal("codex should NOT use stdin")
	}
	if args[0] != "exec" || args[1] != "review this" {
		t.Fatalf("providerArgs() = %v, want [exec 'review this']", args)
	}
}

func TestNewAIReviewCheckerSupportsWizardProviders(t *testing.T) {
	tests := []struct {
		name     string
		provider string
	}{
		{name: "codex", provider: "codex"},
		{name: "opencode", provider: "opencode"},
		{name: "lmstudio", provider: "lmstudio"},
		{name: "github", provider: "github"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := newAIReviewChecker(map[string]interface{}{
				"provider": tt.provider,
			})
			if err != nil {
				t.Fatalf("newAIReviewChecker() error = %v, want nil", err)
			}
		})
	}
}

func TestNewAIReviewCheckerRequiresProvider(t *testing.T) {
	_, err := newAIReviewChecker(map[string]interface{}{
		"prompt": "review",
	})
	if err == nil {
		t.Fatal("newAIReviewChecker() without provider should fail")
	}
}

func TestNewAIReviewCheckerOllamaRequiresModel(t *testing.T) {
	_, err := newAIReviewChecker(map[string]interface{}{
		"provider": "ollama",
		"prompt":   "review",
	})
	if err == nil {
		t.Fatal("newAIReviewChecker(ollama) without model should fail")
	}
}

func TestNewAIReviewCheckerUsesStringPatterns(t *testing.T) {
	instance, err := newAIReviewChecker(map[string]interface{}{
		"provider":     "claude",
		"pass_pattern": `(?i)all good`,
		"fail_pattern": `(?i)all bad`,
	})
	if err != nil {
		t.Fatalf("newAIReviewChecker() error = %v", err)
	}

	c := instance.(*AIReviewChecker)
	if len(c.passPatterns) != 1 || c.passPatterns[0] != `(?i)all good` {
		t.Fatalf("passPatterns = %v, want custom string wrapped in slice", c.passPatterns)
	}
	if len(c.failPatterns) != 1 || c.failPatterns[0] != `(?i)all bad` {
		t.Fatalf("failPatterns = %v, want custom string wrapped in slice", c.failPatterns)
	}
	if len(c.passRegexps) != 1 || len(c.failRegexps) != 1 {
		t.Fatal("expected compiled regexps for custom patterns")
	}
}

func TestNewAIReviewCheckerUsesInterfaceSlicePatterns(t *testing.T) {
	instance, err := newAIReviewChecker(map[string]interface{}{
		"provider":     "claude",
		"pass_pattern": []interface{}{`(?i)all good`, 42, `(?i)ship it`},
		"fail_pattern": []interface{}{`(?i)all bad`, nil},
	})
	if err != nil {
		t.Fatalf("newAIReviewChecker() error = %v", err)
	}

	c := instance.(*AIReviewChecker)
	if len(c.passPatterns) != 2 || c.passPatterns[0] != `(?i)all good` || c.passPatterns[1] != `(?i)ship it` {
		t.Fatalf("passPatterns = %v, want only string entries", c.passPatterns)
	}
	if len(c.failPatterns) != 1 || c.failPatterns[0] != `(?i)all bad` {
		t.Fatalf("failPatterns = %v, want only string entries", c.failPatterns)
	}
}

func TestAIInstallURLsCoverAllProviders(t *testing.T) {
	providers := []string{"claude", "gemini", "ollama", "codex", "opencode", "lmstudio", "github"}
	for _, provider := range providers {
		if strings.TrimSpace(aiInstallURLs[provider]) == "" {
			t.Fatalf("aiInstallURLs[%q] is empty", provider)
		}
	}
}

func TestBuildPrompt(t *testing.T) {
	tests := []struct {
		name      string
		rulesFile string
		setup     func(t *testing.T, dir string)
		want      []string
		notWant   []string
	}{
		{
			name:      "with rules file",
			rulesFile: "AGENTS.md",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				path := filepath.Join(dir, "AGENTS.md")
				if err := os.WriteFile(path, []byte("# Rules\nno panics"), 0o644); err != nil {
					t.Fatalf("WriteFile() error = %v", err)
				}
			},
			want: []string{"CODING STANDARDS", "no panics", "ADDITIONAL INSTRUCTIONS", "review this", "DIFF TO REVIEW", "diff --git", "STATUS: PASSED", "STATUS: FAILED"},
		},
		{
			name:    "without rules file",
			setup:   func(t *testing.T, dir string) {},
			want:    []string{"You are a code reviewer", "ADDITIONAL INSTRUCTIONS", "review this", "DIFF TO REVIEW", "Begin your response now:"},
			notWant: []string{"CODING STANDARDS"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			tt.setup(t, dir)

			c := &AIReviewChecker{
				prompt:       "review this",
				rulesFile:    tt.rulesFile,
				maxDiffLines: 10,
			}

			got, err := c.buildPrompt(CheckContext{Root: dir, Diff: "diff --git a/x b/x"})
			if err != nil {
				t.Fatalf("buildPrompt() error = %v", err)
			}

			for _, want := range tt.want {
				if !strings.Contains(got, want) {
					t.Fatalf("buildPrompt() = %q, want to contain %q", got, want)
				}
			}
			for _, notWant := range tt.notWant {
				if strings.Contains(got, notWant) {
					t.Fatalf("buildPrompt() = %q, should not contain %q", got, notWant)
				}
			}
		})
	}
}

func TestMatchesAny(t *testing.T) {
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)status:\s*passed`),
		regexp.MustCompile(`(?i)\bpass\b`),
	}

	tests := []struct {
		name string
		text string
		want bool
	}{
		{name: "matches first", text: "STATUS: PASSED", want: true},
		{name: "matches second", text: "pass", want: true},
		{name: "no match", text: "FAILED", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := matchesAny(patterns, tt.text); got != tt.want {
				t.Fatalf("matchesAny() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAIReviewCheckerAmbiguousResponse(t *testing.T) {
	binDir := t.TempDir()
	cmdPath := filepath.Join(binDir, "codex")
	script := "#!/bin/sh\ncat >/dev/null\nprintf 'STATUS: PASSED\nSTATUS: FAILED\n'\n"
	if err := os.WriteFile(cmdPath, []byte(script), 0o755); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	oldPath := os.Getenv("PATH")
	if err := os.Setenv("PATH", binDir+string(os.PathListSeparator)+oldPath); err != nil {
		t.Fatalf("Setenv(PATH) error = %v", err)
	}
	defer os.Setenv("PATH", oldPath)

	instance, err := newAIReviewChecker(map[string]interface{}{
		"provider": "codex",
		"prompt":   "review this diff",
	})
	if err != nil {
		t.Fatalf("newAIReviewChecker() error = %v", err)
	}

	result := instance.Check(CheckContext{Root: t.TempDir(), Diff: "diff --git a/x b/x"})
	if result.Status != Error {
		t.Fatalf("Check().Status = %v, want %v", result.Status, Error)
	}
	if !strings.Contains(result.Output, "ambiguous response") {
		t.Fatalf("Check().Output = %q, want ambiguous response message", result.Output)
	}
	if strings.HasPrefix(result.Output, "\n") {
		t.Fatalf("Check().Output = %q, should not start with a newline", result.Output)
	}
}
