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
		t.Fatalf("newAIReviewChecker() error = %v, want nil", err)
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

func TestAIReviewCheckerMissingCLIIncludesInstallURL(t *testing.T) {
	instance, err := newAIReviewChecker(map[string]interface{}{
		"provider": "ollama",
		"model":    "llama3",
		"prompt":   "review this diff",
	})
	if err != nil {
		t.Fatalf("newAIReviewChecker() error = %v, want nil", err)
	}

	result := instance.Check(CheckContext{Root: t.TempDir(), Diff: "diff --git a/x b/x"})
	if result.Status != Error || !strings.Contains(result.Output, "AI CLI not found:") {
		t.Fatalf("Check() = %+v, want missing CLI message", result)
	}
}
