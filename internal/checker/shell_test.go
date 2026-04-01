package checker

import (
	"os"
	"strings"
	"testing"
	"time"
)

func TestShellCheckerPasses(t *testing.T) {
	instance, err := newShellChecker(map[string]interface{}{"command": "echo hello"})
	if err != nil {
		t.Fatalf("newShellChecker() error = %v, want nil", err)
	}

	result := instance.Check(CheckContext{Root: t.TempDir()})
	if result.Status != Passed {
		t.Fatalf("Check().Status = %v, want %v", result.Status, Passed)
	}

	if !strings.Contains(result.Output, "hello") {
		t.Fatalf("Check().Output = %q, want to contain hello", result.Output)
	}
}

func TestShellCheckerFailsOnNonZeroExit(t *testing.T) {
	instance, err := newShellChecker(map[string]interface{}{"command": "exit 1"})
	if err != nil {
		t.Fatalf("newShellChecker() error = %v, want nil", err)
	}

	result := instance.Check(CheckContext{Root: t.TempDir()})
	if result.Status != Failed {
		t.Fatalf("Check().Status = %v, want %v", result.Status, Failed)
	}
}

func TestShellCheckerTimeout(t *testing.T) {
	instance, err := newShellChecker(map[string]interface{}{
		"command": "sleep 10",
		"timeout": "1s",
	})
	if err != nil {
		t.Fatalf("newShellChecker() error = %v, want nil", err)
	}

	startedAt := time.Now()
	result := instance.Check(CheckContext{Root: t.TempDir()})
	if result.Status != Error {
		t.Fatalf("Check().Status = %v, want %v", result.Status, Error)
	}

	if !strings.Contains(result.Output, "timed out after 1s") {
		t.Fatalf("Check().Output = %q, want timeout message", result.Output)
	}

	if elapsed := time.Since(startedAt); elapsed >= 10*time.Second {
		t.Fatalf("Check() elapsed = %v, want timeout before 10s", elapsed)
	}
}

func TestShellCheckerSkipsOnEmptyDiff(t *testing.T) {
	instance, err := newShellChecker(map[string]interface{}{
		"command":            "echo hello",
		"skip_on_empty_diff": true,
	})
	if err != nil {
		t.Fatalf("newShellChecker() error = %v, want nil", err)
	}

	result := instance.Check(CheckContext{Root: t.TempDir(), Diff: ""})
	if result.Status != Skipped {
		t.Fatalf("Check().Status = %v, want %v", result.Status, Skipped)
	}
}

func TestShellCheckerAppliesConfiguredEnv(t *testing.T) {
	instance, err := newShellChecker(map[string]interface{}{
		"command": "printf %s \"$VIGILANTY_TEST_ENV\"",
		"env": map[string]interface{}{
			"VIGILANTY_TEST_ENV": "configured",
		},
	})
	if err != nil {
		t.Fatalf("newShellChecker() error = %v, want nil", err)
	}

	_ = os.Unsetenv("VIGILANTY_TEST_ENV")
	result := instance.Check(CheckContext{Root: t.TempDir(), Diff: "diff --git a/x b/x"})
	if result.Status != Passed {
		t.Fatalf("Check().Status = %v, want %v", result.Status, Passed)
	}
	if strings.TrimSpace(result.Output) != "configured" {
		t.Fatalf("Check().Output = %q, want %q", strings.TrimSpace(result.Output), "configured")
	}
}
