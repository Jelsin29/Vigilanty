package checker

import "time"

type Status int

const (
	Passed Status = iota
	Failed
	Skipped
	Error
)

func (s Status) String() string {
	switch s {
	case Passed:
		return "passed"
	case Failed:
		return "failed"
	case Skipped:
		return "skipped"
	case Error:
		return "error"
	default:
		return "unknown"
	}
}

type CheckContext struct {
	Root        string                 `json:"root" yaml:"root"`
	Diff        string                 `json:"diff" yaml:"diff"`
	Config      map[string]interface{} `json:"config" yaml:"config"`
	StagedFiles []string               `json:"staged_files" yaml:"staged_files"`
}

type CheckResult struct {
	Status   Status        `json:"status" yaml:"status"`
	Output   string        `json:"output" yaml:"output"`
	Details  []Finding     `json:"details,omitempty" yaml:"details,omitempty"`
	Duration time.Duration `json:"duration" yaml:"duration"`
}

type Finding struct {
	File     string `json:"file" yaml:"file"`
	Line     int    `json:"line" yaml:"line"`
	Message  string `json:"message" yaml:"message"`
	Severity string `json:"severity" yaml:"severity"`
}

type Checker interface {
	Name() string
	Check(ctx CheckContext) CheckResult
}
