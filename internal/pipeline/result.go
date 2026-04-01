package pipeline

import (
	"time"

	"github.com/jelsin/vigilanty/internal/checker"
)

type PipelineResult struct {
	Results  []StepResult  `json:"results" yaml:"results"`
	Duration time.Duration `json:"duration" yaml:"duration"`
	Passed   bool          `json:"passed" yaml:"passed"`
}

type StepResult struct {
	Name   string              `json:"name" yaml:"name"`
	Result checker.CheckResult `json:"result" yaml:"result"`
}
