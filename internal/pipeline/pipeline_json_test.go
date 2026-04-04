package pipeline

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jelsin29/vigilanty/internal/checker"
	"github.com/jelsin29/vigilanty/internal/config"
)

func TestJSONResultMatchesGoldenFixture(t *testing.T) {
	result := PipelineResult{
		Passed:   false,
		Duration: 1500 * time.Millisecond,
		Results: []StepResult{
			{Name: "lint", Type: "shell", Timeout: "60s", Files: []string{"main.go"}, Result: checker.CheckResult{Status: checker.Passed, Duration: 100 * time.Millisecond}},
			{Name: "test", Type: "shell", Timeout: "120s", Files: []string{"main.go"}, Result: checker.CheckResult{Status: checker.Failed, Duration: 200 * time.Millisecond, Output: "tests failed\n"}},
			{Name: "review", Type: "ai-review", Provider: "claude", Timeout: "120s", Files: []string{"main.go"}, Result: checker.CheckResult{Status: checker.Skipped}},
			{Name: "security", Type: "shell", Timeout: "90s", Files: []string{"main.go"}, Cached: true, Result: checker.CheckResult{Status: checker.Error, Duration: 300 * time.Millisecond, Output: "checker crashed", Details: []checker.Finding{{File: "main.go", Line: 7, Message: "panic", Severity: "high"}}}},
		},
	}

	payload, err := JSONResult(result, RunJSONMeta{Mode: "ci", Files: []string{"main.go"}, TruncatedDiff: true})
	if err != nil {
		t.Fatalf("JSONResult() error = %v", err)
	}

	goldenPath := filepath.Join("testdata", "run_json_golden.json")
	want, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("os.ReadFile(%q) error = %v", goldenPath, err)
	}

	if strings.TrimSpace(string(payload)) != strings.TrimSpace(string(want)) {
		t.Fatalf("JSONResult() mismatch\n got: %s\nwant: %s", payload, want)
	}
}

func TestJSONResultHandlesEmptySteps(t *testing.T) {
	payload, err := JSONResult(PipelineResult{Passed: true, Duration: 0}, RunJSONMeta{Mode: "staged"})
	if err != nil {
		t.Fatalf("JSONResult() error = %v", err)
	}

	want := `{"schema_version":"v1","passed":true,"mode":"staged","duration_ms":0,"steps":[],"summary":"Pipeline passed (0/0 checkers) in 0s"}`
	if string(payload) != want {
		t.Fatalf("JSONResult() = %s, want %s", payload, want)
	}
}

func TestPipelineQuietModeSuppressesLiveStdout(t *testing.T) {
	checkerType := registerPipelineTestChecker(t, checker.Passed)
	pipe := New(&config.Config{
		Pipeline: []config.StepConfig{{Name: "verify", Type: checkerType}},
	}, Options{Quiet: true})

	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() error = %v", err)
	}
	oldStdout := os.Stdout
	os.Stdout = writer
	defer func() {
		os.Stdout = oldStdout
	}()

	result := pipe.Run(checker.CheckContext{})
	_ = writer.Close()
	output, _ := io.ReadAll(reader)
	_ = reader.Close()

	if result.LiveOutputted {
		t.Fatalf("Run().LiveOutputted = true, want false in quiet mode")
	}
	if strings.TrimSpace(string(output)) != "" {
		t.Fatalf("quiet mode stdout = %q, want empty output", string(output))
	}
}

func TestPipelineQuietModeSuppressesSkippedStdoutAfterFailure(t *testing.T) {
	checkerTypeFail := registerPipelineTestChecker(t, checker.Failed)
	checkerTypeSkipped := registerPipelineTestChecker(t, checker.Passed)
	pipe := New(&config.Config{
		Global: config.GlobalConfig{FailFast: true},
		Pipeline: []config.StepConfig{
			{Name: "fail", Type: checkerTypeFail},
			{Name: "skipped", Type: checkerTypeSkipped},
		},
	}, Options{Quiet: true})

	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() error = %v", err)
	}
	oldStdout := os.Stdout
	os.Stdout = writer
	defer func() {
		os.Stdout = oldStdout
	}()

	result := pipe.Run(checker.CheckContext{})
	_ = writer.Close()
	output, _ := io.ReadAll(reader)
	_ = reader.Close()

	if result.Passed {
		t.Fatal("Run().Passed = true, want false")
	}
	if result.LiveOutputted {
		t.Fatalf("Run().LiveOutputted = true, want false in quiet fail-fast mode")
	}
	if len(result.Results) != 2 {
		t.Fatalf("len(Run().Results) = %d, want 2", len(result.Results))
	}
	if got := result.Results[1].Result.Status; got != checker.Skipped {
		t.Fatalf("second checker status = %v, want %v", got, checker.Skipped)
	}
	if strings.TrimSpace(string(output)) != "" {
		t.Fatalf("quiet fail-fast stdout = %q, want empty output", string(output))
	}
}
