package pipeline

import (
	"fmt"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Jelsin29/Vigilanty/internal/checker"
	"github.com/Jelsin29/Vigilanty/internal/config"
)

var pipelineCheckerSeq atomic.Uint64

func TestPipelineAllPassingCheckers(t *testing.T) {
	checkerTypeA := registerPipelineTestChecker(t, checker.Passed)
	checkerTypeB := registerPipelineTestChecker(t, checker.Passed)

	pipe := New(&config.Config{
		Pipeline: []config.StepConfig{
			{Name: "first", Type: checkerTypeA},
			{Name: "second", Type: checkerTypeB},
		},
	})

	result := pipe.Run(checker.CheckContext{})
	if !result.Passed {
		t.Fatal("Run().Passed = false, want true")
	}
	if len(result.Results) != 2 {
		t.Fatalf("len(Run().Results) = %d, want 2", len(result.Results))
	}
}

func TestPipelineFailFastStopsAfterFailure(t *testing.T) {
	checkerTypeFail, failCalls := registerPipelineTestCheckerWithCounter(t, checker.Failed)
	checkerTypeNever, neverCalls := registerPipelineTestCheckerWithCounter(t, checker.Passed)

	pipe := New(&config.Config{
		Global: config.GlobalConfig{FailFast: true},
		Pipeline: []config.StepConfig{
			{Name: "first", Type: checkerTypeFail},
			{Name: "second", Type: checkerTypeNever},
		},
	})

	result := pipe.Run(checker.CheckContext{})
	if result.Passed {
		t.Fatal("Run().Passed = true, want false")
	}
	if len(result.Results) != 2 {
		t.Fatalf("len(Run().Results) = %d, want 2", len(result.Results))
	}
	if got := failCalls.Load(); got != 1 {
		t.Fatalf("failed checker calls = %d, want 1", got)
	}
	if got := neverCalls.Load(); got != 0 {
		t.Fatalf("second checker calls = %d, want 0", got)
	}
	if got := result.Results[1].Result.Status; got != checker.Skipped {
		t.Fatalf("second checker status = %v, want %v", got, checker.Skipped)
	}
}

func TestPipelineEmptyPasses(t *testing.T) {
	result := New(&config.Config{}).Run(checker.CheckContext{})
	if !result.Passed {
		t.Fatal("Run().Passed = false, want true")
	}
	if len(result.Results) != 0 {
		t.Fatalf("len(Run().Results) = %d, want 0", len(result.Results))
	}
}

func TestPipelineErrorStopsExecution(t *testing.T) {
	checkerTypeError, errorCalls := registerPipelineTestCheckerWithCounter(t, checker.Error)
	checkerTypeNever, neverCalls := registerPipelineTestCheckerWithCounter(t, checker.Passed)

	pipe := New(&config.Config{
		Global: config.GlobalConfig{FailFast: true},
		Pipeline: []config.StepConfig{
			{Name: "first", Type: checkerTypeError},
			{Name: "second", Type: checkerTypeNever},
		},
	})

	result := pipe.Run(checker.CheckContext{})
	if result.Passed {
		t.Fatal("Run().Passed = true, want false")
	}
	if len(result.Results) != 2 {
		t.Fatalf("len(Run().Results) = %d, want 2", len(result.Results))
	}
	if got := errorCalls.Load(); got != 1 {
		t.Fatalf("error checker calls = %d, want 1", got)
	}
	if got := neverCalls.Load(); got != 0 {
		t.Fatalf("second checker calls = %d, want 0", got)
	}
	if got := result.Results[1].Result.Status; got != checker.Skipped {
		t.Fatalf("second checker status = %v, want %v", got, checker.Skipped)
	}
}

func TestFormatOutputIncludesSummaryCounts(t *testing.T) {
	output := FormatOutput(PipelineResult{
		Passed:   false,
		Duration: time.Second,
		Results: []StepResult{
			{Name: "lint", Result: checker.CheckResult{Status: checker.Passed, Duration: time.Millisecond}},
			{Name: "test", Result: checker.CheckResult{Status: checker.Failed, Duration: time.Millisecond}},
			{Name: "ai-review", Result: checker.CheckResult{Status: checker.Skipped}},
		},
	}, false)

	if !strings.Contains(output, `✗ Pipeline failed at "test" (1/3 checkers passed) in 1s`) {
		t.Fatalf("FormatOutput() = %q, want failure summary", output)
	}
	if !strings.Contains(output, "✓ lint") {
		t.Fatalf("FormatOutput() = %q, want passed icon", output)
	}
	if !strings.Contains(output, "✗ test") {
		t.Fatalf("FormatOutput() = %q, want failed icon", output)
	}
	if !strings.Contains(output, "○ ai-review") {
		t.Fatalf("FormatOutput() = %q, want skipped icon", output)
	}
}

type pipelineTestChecker struct {
	status checker.Status
	calls  *atomic.Int32
}

func (p *pipelineTestChecker) Name() string {
	return "pipeline-test-checker"
}

func (p *pipelineTestChecker) Check(ctx checker.CheckContext) checker.CheckResult {
	if p.calls != nil {
		p.calls.Add(1)
	}
	return checker.CheckResult{Status: p.status}
}

func registerPipelineTestChecker(t *testing.T, status checker.Status) string {
	t.Helper()
	checkerType, _ := registerPipelineTestCheckerWithCounter(t, status)
	return checkerType
}

func registerPipelineTestCheckerWithCounter(t *testing.T, status checker.Status) (string, *atomic.Int32) {
	t.Helper()

	checkerType := fmt.Sprintf("%s-%s-%d", t.Name(), status.String(), pipelineCheckerSeq.Add(1))
	calls := &atomic.Int32{}
	checker.Register(checkerType, func(cfg map[string]interface{}) (checker.Checker, error) {
		return &pipelineTestChecker{status: status, calls: calls}, nil
	})

	return checkerType, calls
}
