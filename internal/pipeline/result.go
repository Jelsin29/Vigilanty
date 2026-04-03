package pipeline

import (
	"time"

	"github.com/jelsin29/vigilanty/internal/checker"
)

type PipelineResult struct {
	Results       []StepResult  `json:"results" yaml:"results"`
	Duration      time.Duration `json:"duration" yaml:"duration"`
	Passed        bool          `json:"passed" yaml:"passed"`
	LiveOutputted bool          `json:"live_outputted,omitempty" yaml:"live_outputted,omitempty"`
}

type StepResult struct {
	Name     string              `json:"name" yaml:"name"`
	Type     string              `json:"type,omitempty" yaml:"type,omitempty"`
	Provider string              `json:"provider,omitempty" yaml:"provider,omitempty"`
	Model    string              `json:"model,omitempty" yaml:"model,omitempty"`
	Timeout  string              `json:"timeout,omitempty" yaml:"timeout,omitempty"`
	Files    []string            `json:"files,omitempty" yaml:"files,omitempty"`
	Result   checker.CheckResult `json:"result" yaml:"result"`
	Cached   bool                `json:"cached,omitempty" yaml:"cached,omitempty"`
}
