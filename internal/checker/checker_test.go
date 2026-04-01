package checker

import (
	"testing"
	"time"
)

func TestStatusString(t *testing.T) {
	tests := []struct {
		name   string
		status Status
		want   string
	}{
		{name: "passed", status: Passed, want: "passed"},
		{name: "failed", status: Failed, want: "failed"},
		{name: "skipped", status: Skipped, want: "skipped"},
		{name: "error", status: Error, want: "error"},
		{name: "unknown", status: Status(99), want: "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.status.String(); got != tt.want {
				t.Fatalf("Status.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCheckResultStoresStatuses(t *testing.T) {
	tests := []struct {
		name   string
		status Status
	}{
		{name: "passed", status: Passed},
		{name: "failed", status: Failed},
		{name: "skipped", status: Skipped},
		{name: "error", status: Error},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CheckResult{
				Status:   tt.status,
				Output:   "checker output",
				Details:  []Finding{{File: "main.go", Line: 10, Message: "problem", Severity: "warning"}},
				Duration: 25 * time.Millisecond,
			}

			if result.Status != tt.status {
				t.Fatalf("CheckResult.Status = %v, want %v", result.Status, tt.status)
			}

			if got := result.Status.String(); got != tt.status.String() {
				t.Fatalf("CheckResult.Status.String() = %q, want %q", got, tt.status.String())
			}

			if result.Output != "checker output" {
				t.Fatalf("CheckResult.Output = %q, want checker output", result.Output)
			}

			if len(result.Details) != 1 {
				t.Fatalf("len(CheckResult.Details) = %d, want 1", len(result.Details))
			}

			if result.Duration <= 0 {
				t.Fatalf("CheckResult.Duration = %v, want > 0", result.Duration)
			}
		})
	}
}
