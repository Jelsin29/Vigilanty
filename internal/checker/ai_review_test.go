package checker

import (
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
	args := c.providerArgs()
	// should use stdin marker "-" instead of the prompt itself
	if len(args) < 2 || args[1] != "-" {
		t.Fatalf("providerArgs() = %v, want '-' as stdin marker", args)
	}
}

func TestProviderArgsOllama(t *testing.T) {
	c := &AIReviewChecker{provider: "ollama", model: "llama3"}
	args := c.providerArgs()
	if args[0] != "run" || args[1] != "llama3" {
		t.Fatalf("providerArgs() = %v, want [run llama3]", args)
	}
	// prompt should NOT be in args (it goes through stdin)
	if len(args) > 2 {
		t.Fatalf("providerArgs() has too many args, prompt should go via stdin: %v", args)
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

// this step will be done in ai-review
// func TestAIReviewCheckerMissingCLIIncludesInstallURL(t *testing.T) {
// 	skipIfNoLlama3(t)
// 	instance, err := newAIReviewChecker(map[string]interface{}{
// 		"provider": "ollama",
// 		"model":    "llama3",
// 		"prompt":   "review this diff",
// 		"timeout":  "2m",
// 	})
// 	if err != nil {
// 		t.Fatalf("newAIReviewChecker() error = %v, want nil", err)
// 	}

// 	cmd := exec.Command("ollama", "show", "llama3")
// 	if err := cmd.Run(); err != nil {
// 		t.Fatalf("llama3 model not available in ollama, skipping test. To install, follow instructions at https://ollama.com/docs/installation")
// 	}

// 	result := instance.Check(CheckContext{Root: t.TempDir(), Diff: "diff --git a/x b/x"})
// 	if result.Status == Error || !strings.Contains(result.Output, "AI CLI not found:") {
// 		t.Fatalf("Check() = %+v", result)
// 	}
// }

// func skipIfNoLlama3(t *testing.T) {
// 	t.Helper()
// 	if _, err := exec.LookPath("ollama"); err != nil {
// 		t.Fatalf("ollama CLI not found, skipping test")
// 	}
// 	cmd := exec.Command("ollama", "show", "llama3")
// 	if err := cmd.Run(); err != nil {
// 		t.Fatalf("llama3 model not available in ollama, skipping test.\nNEED TO INSTALL `llama3` for OLLAMA.")
// 	}
// }
