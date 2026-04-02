package pipeline

import (
	"fmt"
	"os"
	"strings"
	"time"

	cachepkg "github.com/jelsin29/vigilanty/internal/cache"
	"github.com/jelsin29/vigilanty/internal/checker"
	"github.com/jelsin29/vigilanty/internal/config"
	"github.com/jelsin29/vigilanty/internal/ui"
)

type Pipeline struct {
	steps    []config.StepConfig
	failFast bool
	verbose  bool
	noCache  bool
	mode     string
}

type Options struct {
	NoCache bool
	Mode    string
}

func New(cfg *config.Config, options Options) *Pipeline {
	if cfg == nil {
		return &Pipeline{}
	}

	steps := make([]config.StepConfig, len(cfg.Pipeline))
	copy(steps, cfg.Pipeline)

	return &Pipeline{
		steps:    steps,
		failFast: cfg.Global.FailFast,
		verbose:  cfg.Global.Verbose,
		noCache:  options.NoCache,
		mode:     options.Mode,
	}
}

func (p *Pipeline) Run(ctx checker.CheckContext) PipelineResult {
	startedAt := time.Now()
	result := PipelineResult{
		Results: make([]StepResult, 0, len(p.steps)),
		Passed:  true,
	}
	liveOutputEnabled := !ui.StdoutIsTTY()

	projectCache, cacheEnabled, filesHash := p.loadCache(ctx)

	for i, step := range p.steps {
		if step.Enabled != nil && !*step.Enabled {
			stepResult := StepResult{
				Name: step.Name,
				Result: checker.CheckResult{
					Status: checker.Skipped,
					Output: "step disabled",
				},
			}
			result.Results = append(result.Results, stepResult)
			printLiveStepResult(stepResult)
			result.LiveOutputted = result.LiveOutputted || liveOutputEnabled
			continue
		}

		stepStartedAt := time.Now()
		cfg := checkerConfig(step)
		configHash := cachepkg.ConfigHash(cfg)

		if cacheEnabled {
			if _, ok := projectCache.Lookup(step.Name, filesHash, configHash); ok {
				stepResult := StepResult{
					Name:   step.Name,
					Cached: true,
					Result: checker.CheckResult{Status: checker.Passed},
				}
				result.Results = append(result.Results, stepResult)
				printLiveStepResult(stepResult)
				result.LiveOutputted = result.LiveOutputted || liveOutputEnabled
				continue
			}
		}

		spinner := ui.NewSpinner(step.Name)
		spinner.Start()
		instance, err := checker.Create(step.Type, cfg)
		if err != nil {
			if cacheEnabled {
				_ = projectCache.Remove(step.Name)
			}

			stepResult := StepResult{
				Name: step.Name,
				Result: checker.CheckResult{
					Status:   checker.Error,
					Output:   fmt.Sprintf("create checker: %v", err),
					Duration: time.Since(stepStartedAt),
				},
			}
			spinner.Stop(statusIcon(stepResult.Result.Status))
			result.LiveOutputted = result.LiveOutputted || liveOutputEnabled
			result.Results = append(result.Results, stepResult)
			result.Passed = false
			if p.failFast {
				appendSkippedResults(&result, p.steps, i+1, liveOutputEnabled)
				break
			}
			continue
		}

		checkResult := instance.Check(ctx)
		if checkResult.Duration <= 0 {
			checkResult.Duration = time.Since(stepStartedAt)
		}
		spinner.Stop(statusIcon(checkResult.Status))
		result.LiveOutputted = result.LiveOutputted || liveOutputEnabled

		if cacheEnabled {
			switch checkResult.Status {
			case checker.Passed:
				_ = projectCache.Store(step.Name, filesHash, configHash)
			case checker.Failed, checker.Error:
				_ = projectCache.Remove(step.Name)
			}
		}

		stepResult := StepResult{Name: step.Name, Result: checkResult}
		result.Results = append(result.Results, stepResult)

		if checkResult.Status == checker.Failed || checkResult.Status == checker.Error {
			result.Passed = false
			if p.failFast {
				appendSkippedResults(&result, p.steps, i+1, liveOutputEnabled)
				break
			}
		}
	}

	result.Duration = time.Since(startedAt)
	return result
}

func FormatOutput(result PipelineResult, verbose bool) string {
	var builder strings.Builder
	writeOutput(&builder, result, verbose, false)
	return builder.String()
}

func FormatSummary(result PipelineResult) string {
	var builder strings.Builder
	writeOutput(&builder, result, false, true)
	return builder.String()
}

func writeOutput(builder *strings.Builder, result PipelineResult, verbose bool, summaryOnly bool) {
	passedCount := 0
	failedAt := ""

	for _, step := range result.Results {
		if step.Result.Status == checker.Passed {
			passedCount++
		}
		if failedAt == "" && (step.Result.Status == checker.Failed || step.Result.Status == checker.Error) {
			failedAt = step.Name
		}

		if summaryOnly {
			if (step.Result.Status == checker.Failed || step.Result.Status == checker.Error) && strings.TrimSpace(step.Result.Output) != "" {
				builder.WriteString("  " + strings.TrimSpace(step.Result.Output) + "\n")
			}
			continue
		}

		builder.WriteString(ui.Colorize(statusColor(step.Result.Status), statusIcon(step.Result.Status)+" "+step.Name))
		builder.WriteString(" (")
		builder.WriteString(stepSummary(step))
		builder.WriteString(")")
		builder.WriteString("\n")

		if shouldPrintOutput(step.Result.Status, verbose) && strings.TrimSpace(step.Result.Output) != "" {
			builder.WriteString(indent(strings.TrimSpace(step.Result.Output), "  "))
			builder.WriteString("\n")
		}
	}

	builder.WriteString("\n")
	if result.Passed {
		builder.WriteString(ui.SuccessText(fmt.Sprintf("✓ Pipeline passed (%d/%d checkers) in %s", passedCount, len(result.Results), result.Duration)))
	} else {
		builder.WriteString(ui.ErrorText(fmt.Sprintf("✗ Pipeline failed at %q (%d/%d checkers passed) in %s", failedAt, passedCount, len(result.Results), result.Duration)))
	}
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
		return ui.Green
	case checker.Failed, checker.Error:
		return ui.Red
	case checker.Skipped:
		return ui.Yellow
	default:
		return ui.Blue
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

func stepSummary(step StepResult) string {
	if step.Cached {
		return "cached"
	}

	if step.Result.Status == checker.Skipped {
		trimmed := strings.TrimSpace(step.Result.Output)
		if strings.HasPrefix(trimmed, "skipped (") && strings.HasSuffix(trimmed, ")") {
			return "skipped — " + strings.TrimSuffix(strings.TrimPrefix(trimmed, "skipped ("), ")")
		}
		if trimmed != "" {
			return trimmed
		}
	}

	return step.Result.Duration.String()
}

func indent(value string, prefix string) string {
	lines := strings.Split(value, "\n")
	for i := range lines {
		lines[i] = prefix + lines[i]
	}
	return strings.Join(lines, "\n")
}

func appendSkippedResults(result *PipelineResult, steps []config.StepConfig, start int, liveOutputEnabled bool) {
	for i := start; i < len(steps); i++ {
		stepResult := StepResult{
			Name: steps[i].Name,
			Result: checker.CheckResult{
				Status: checker.Skipped,
				Output: "skipped — previous step failed",
			},
		}
		result.Results = append(result.Results, stepResult)
		printLiveStepResult(stepResult)
		if liveOutputEnabled {
			result.LiveOutputted = true
		}
	}
}

func printLiveStepResult(step StepResult) {
	if ui.StdoutIsTTY() {
		return
	}

	fmt.Fprintf(os.Stdout, "%s %s (%s)\n", statusIcon(step.Result.Status), step.Name, stepSummary(step))
}

func (p *Pipeline) loadCache(ctx checker.CheckContext) (*cachepkg.Cache, bool, string) {
	if p.noCache || strings.TrimSpace(ctx.Root) == "" {
		return nil, false, ""
	}

	projectCache := cachepkg.New(ctx.Root)
	if err := projectCache.Load(); err != nil {
		fmt.Fprintf(os.Stderr, "warning: cache unavailable: %v\n", err)
		return nil, false, ""
	}

	return projectCache, true, cachepkg.FilesHash(ctx.Root, ctx.StagedFiles, p.mode)
}
