package pipeline

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jelsin29/vigilanty/internal/checker"
)

const runJSONSchemaVersion = "v1"

type RunJSON struct {
	SchemaVersion string        `json:"schema_version"`
	Passed        bool          `json:"passed"`
	Mode          string        `json:"mode,omitempty"`
	DurationMs    int64         `json:"duration_ms"`
	Files         []string      `json:"files,omitempty"`
	TruncatedDiff bool          `json:"truncated_diff,omitempty"`
	Steps         []RunJSONStep `json:"steps"`
	Summary       string        `json:"summary"`
}

type RunJSONStep struct {
	Name       string            `json:"name"`
	Type       string            `json:"type,omitempty"`
	Status     string            `json:"status"`
	DurationMs int64             `json:"duration_ms"`
	Output     string            `json:"output,omitempty"`
	Details    []checker.Finding `json:"details,omitempty"`
	Provider   string            `json:"provider,omitempty"`
	Model      string            `json:"model,omitempty"`
	Cached     bool              `json:"cached,omitempty"`
	Files      []string          `json:"files,omitempty"`
	Timeout    string            `json:"timeout,omitempty"`
}

type RunJSONMeta struct {
	Mode          string
	Files         []string
	TruncatedDiff bool
}

func JSONResult(result PipelineResult, meta RunJSONMeta) ([]byte, error) {
	payload := RunJSON{
		SchemaVersion: runJSONSchemaVersion,
		Passed:        result.Passed,
		Mode:          strings.TrimSpace(meta.Mode),
		DurationMs:    result.Duration.Milliseconds(),
		Files:         append([]string(nil), meta.Files...),
		TruncatedDiff: meta.TruncatedDiff,
		Steps:         make([]RunJSONStep, 0, len(result.Results)),
	}

	for _, step := range result.Results {
		payload.Steps = append(payload.Steps, RunJSONStep{
			Name:       step.Name,
			Type:       step.Type,
			Status:     step.Result.Status.String(),
			DurationMs: step.Result.Duration.Milliseconds(),
			Output:     strings.TrimSpace(step.Result.Output),
			Details:    append([]checker.Finding(nil), step.Result.Details...),
			Provider:   step.Provider,
			Model:      step.Model,
			Cached:     step.Cached,
			Files:      append([]string(nil), step.Files...),
			Timeout:    step.Timeout,
		})
	}

	payload.Summary = jsonSummary(result)

	encoded, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	return encoded, nil
}

func jsonSummary(result PipelineResult) string {
	passedCount, failedAt := pipelineSummary(result)
	if result.Passed {
		return fmt.Sprintf("Pipeline passed (%d/%d checkers) in %s", passedCount, len(result.Results), result.Duration)
	}

	return fmt.Sprintf("Pipeline failed at %q (%d/%d checkers passed) in %s", failedAt, passedCount, len(result.Results), result.Duration)
}
