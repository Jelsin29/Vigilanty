package pipeline

import (
	"fmt"
	"strings"
	"time"

	"github.com/jelsin/vigilanty/internal/checker"
	"github.com/jelsin/vigilanty/internal/config"
)

const (
	ansiReset  = "\033[0m"
	ansiRed    = "\033[31m"
	ansiGreen  = "\033[32m"
	ansiYellow = "\033[33m"
	ansiBlue   = "\033[34m"
)

type Pipeline struct {
	steps    []config.StepConfig
	failFast bool
	verbose  bool
}

func New(cfg *config.Config) *Pipeline {
	if cfg == nil {
		return &Pipeline{}
	}

	steps := make([]config.StepConfig, len(cfg.Pipeline))
	copy(steps, cfg.Pipeline)

	return &Pipeline{
		steps:    steps,
		failFast: cfg.Global.FailFast,
		verbose:  cfg.Global.Verbose,
	}
}

func (p *Pipeline) Run(ctx checker.CheckContext) PipelineResult {
	startedAt := time.Now()
	result := PipelineResult{
		Results: make([]StepResult, 0, len(p.steps)),
		Passed:  true,
	}

	for i, step := range p.steps {
		if step.Enabled != nil && !*step.Enabled {
			result.Results = append(result.Results, StepResult{
				Name: step.Name,
				Result: checker.CheckResult{
					Status: checker.Skipped,
					Output: "step disabled",
				},
			})
			continue
		}

		stepStartedAt := time.Now()
		instance, err := checker.Create(step.Type, checkerConfig(step))
		if err != nil {
			stepResult := StepResult{
				Name: step.Name,
				Result: checker.CheckResult{
					Status:   checker.Error,
					Output:   fmt.Sprintf("create checker: %v", err),
					Duration: time.Since(stepStartedAt),
				},
			}
			result.Results = append(result.Results, stepResult)
			result.Passed = false
			if p.failFast {
				for j := i + 1; j < len(p.steps); j++ {
					result.Results = append(result.Results, StepResult{
						Name: p.steps[j].Name,
						Result: checker.CheckResult{
							Status: checker.Skipped,
							Output: "skipped (previous checker failed)",
						},
					})
				}
				break
			}
			continue
		}

		checkResult := instance.Check(ctx)
		if checkResult.Duration <= 0 {
			checkResult.Duration = time.Since(stepStartedAt)
		}

		stepResult := StepResult{Name: step.Name, Result: checkResult}
		result.Results = append(result.Results, stepResult)

		if checkResult.Status == checker.Failed || checkResult.Status == checker.Error {
			result.Passed = false
			if p.failFast {
				for j := i + 1; j < len(p.steps); j++ {
					result.Results = append(result.Results, StepResult{
						Name: p.steps[j].Name,
						Result: checker.CheckResult{
							Status: checker.Skipped,
							Output: "skipped (previous checker failed)",
						},
					})
				}
				break
			}
		}
	}

	result.Duration = time.Since(startedAt)
	return result
}

func FormatOutput(result PipelineResult, verbose bool) string {
	var builder strings.Builder
	passedCount := 0
	failedAt := ""

	for _, step := range result.Results {
		if step.Result.Status == checker.Passed {
			passedCount++
		}
		if failedAt == "" && (step.Result.Status == checker.Failed || step.Result.Status == checker.Error) {
			failedAt = step.Name
		}

		builder.WriteString(statusColor(step.Result.Status))
		builder.WriteString(statusIcon(step.Result.Status))
		builder.WriteString(" ")
		builder.WriteString(step.Name)
		builder.WriteString(ansiReset)
		builder.WriteString(" (")
		builder.WriteString(step.Result.Duration.String())
		builder.WriteString(")")
		builder.WriteString("\n")

		if shouldPrintOutput(step.Result.Status, verbose) && strings.TrimSpace(step.Result.Output) != "" {
			builder.WriteString(indent(strings.TrimSpace(step.Result.Output), "  "))
			builder.WriteString("\n")
		}
	}

	builder.WriteString("\n")
	if result.Passed {
		builder.WriteString(ansiGreen)
		builder.WriteString(fmt.Sprintf("✓ Pipeline passed (%d/%d checkers) in %s", passedCount, len(result.Results), result.Duration))
	} else {
		builder.WriteString(ansiRed)
		builder.WriteString(fmt.Sprintf("✗ Pipeline failed at %q (%d/%d checkers passed) in %s", failedAt, passedCount, len(result.Results), result.Duration))
	}
	builder.WriteString(ansiReset)

	return builder.String()
}

func shouldPrintOutput(status checker.Status, verbose bool) bool {
	if verbose {
		return true
	}

	switch status {
	case checker.Failed, checker.Error:
		return true
	default:
		return false
	}
}

func checkerConfig(step config.StepConfig) map[string]interface{} {
	merged := make(map[string]interface{}, len(step.Config)+10)
	for key, value := range step.Config {
		merged[key] = value
	}

	if strings.TrimSpace(step.Command) != "" {
		merged["command"] = step.Command
	}
	if strings.TrimSpace(step.Provider) != "" {
		merged["provider"] = step.Provider
	}
	if strings.TrimSpace(step.Prompt) != "" {
		merged["prompt"] = step.Prompt
	}
	if strings.TrimSpace(step.Model) != "" {
		merged["model"] = step.Model
	}
	if strings.TrimSpace(step.Timeout) != "" {
		merged["timeout"] = step.Timeout
	}
	if len(step.Env) > 0 {
		merged["env"] = step.Env
	}
	if step.SkipOnEmptyDiff {
		merged["skip_on_empty_diff"] = step.SkipOnEmptyDiff
	}
	if step.MaxDiffLines > 0 {
		merged["max_diff_lines"] = step.MaxDiffLines
	}
	if strings.TrimSpace(step.PassPattern) != "" {
		merged["pass_pattern"] = step.PassPattern
	}
	if strings.TrimSpace(step.FailPattern) != "" {
		merged["fail_pattern"] = step.FailPattern
	}

	return merged
}

func statusColor(status checker.Status) string {
	switch status {
	case checker.Passed:
		return ansiGreen
	case checker.Failed, checker.Error:
		return ansiRed
	case checker.Skipped:
		return ansiYellow
	default:
		return ansiBlue
	}
}

func statusIcon(status checker.Status) string {
	switch status {
	case checker.Passed:
		return "✓"
	case checker.Failed:
		return "✗"
	case checker.Skipped:
		return "○"
	case checker.Error:
		return "⚠"
	default:
		return "?"
	}
}

func indent(value string, prefix string) string {
	lines := strings.Split(value, "\n")
	for i := range lines {
		lines[i] = prefix + lines[i]
	}
	return strings.Join(lines, "\n")
}
