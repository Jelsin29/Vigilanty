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
	liveOutputEnabled := true // spinners and AI review sections output live in all modes

	projectCache, cacheEnabled, filesHash := p.loadCache(ctx)

	for i, step := range p.steps {
		if step.Enabled != nil && !*step.Enabled {
			stepResult := StepResult{
				Name:     step.Name,
				Type:     step.Type,
				Provider: step.Provider,
				Model:    step.Model,
				Timeout:  step.Timeout,
				Files:    append([]string(nil), ctx.StagedFiles...),
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
					Name:     step.Name,
					Type:     step.Type,
					Provider: step.Provider,
					Model:    step.Model,
					Timeout:  step.Timeout,
					Files:    append([]string(nil), ctx.StagedFiles...),
					Cached:   true,
					Result:   checker.CheckResult{Status: checker.Passed},
				}
				result.Results = append(result.Results, stepResult)
				printLiveStepResult(stepResult)
				result.LiveOutputted = result.LiveOutputted || liveOutputEnabled
				continue
			}
		}

		ttyAIReview := ui.StdoutIsTTY() && isAIReview(step)

		if ttyAIReview {
			printAIReviewStart(step, ctx.StagedFiles)
		}

		var spinner *ui.Spinner
		if !ttyAIReview {
			spinner = ui.NewSpinner(step.Name)
			spinner.Start()
		}

		instance, err := checker.Create(step.Type, cfg)
		if err != nil {
			if cacheEnabled {
				_ = projectCache.Remove(step.Name)
			}

			stepResult := StepResult{
				Name:     step.Name,
				Type:     step.Type,
				Provider: step.Provider,
				Model:    step.Model,
				Timeout:  step.Timeout,
				Files:    append([]string(nil), ctx.StagedFiles...),
				Result: checker.CheckResult{
					Status:   checker.Error,
					Output:   fmt.Sprintf("create checker: %v", err),
					Duration: time.Since(stepStartedAt),
				},
			}
			if spinner != nil {
				spinner.Stop(statusIcon(stepResult.Result.Status))
			}
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
		if spinner != nil {
			spinner.Stop(statusIcon(checkResult.Status))
		}
		result.LiveOutputted = result.LiveOutputted || liveOutputEnabled

		if ttyAIReview {
			printAIReviewResult(checkResult)
		}

		if cacheEnabled {
			switch checkResult.Status {
			case checker.Passed:
				_ = projectCache.Store(step.Name, filesHash, configHash)
			case checker.Failed, checker.Error:
				_ = projectCache.Remove(step.Name)
			}
		}

		stepResult := StepResult{
			Name:     step.Name,
			Type:     step.Type,
			Provider: step.Provider,
			Model:    step.Model,
			Timeout:  step.Timeout,
			Files:    append([]string(nil), ctx.StagedFiles...),
			Result:   checkResult,
		}
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
	if ui.StdoutIsTTY() {
		// In TTY mode, everything was printed live. Just the summary bar.
		separator := "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
		builder.WriteString(ui.Colorize(ui.BttfAmber, separator))
		builder.WriteString("\n")
		passedCount, failedAt := pipelineSummary(result)
		if result.Passed {
			builder.WriteString(ui.SuccessText(fmt.Sprintf("✓ Pipeline passed (%d/%d checkers) in %s", passedCount, len(result.Results), result.Duration)))
		} else {
			builder.WriteString(ui.ErrorText(fmt.Sprintf("✗ Pipeline failed at %q (%d/%d checkers passed) in %s", failedAt, passedCount, len(result.Results), result.Duration)))
		}
		builder.WriteString("\n")
		builder.WriteString(ui.Colorize(ui.BttfAmber, separator))
		return builder.String()
	}
	writeOutput(&builder, result, false, true)
	return builder.String()
}

func writeOutput(builder *strings.Builder, result PipelineResult, verbose bool, summaryOnly bool) {
	if ui.StdoutIsTTY() && !summaryOnly {
		writeTTYOutput(builder, result, verbose)
		return
	}

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

func writeTTYOutput(builder *strings.Builder, result PipelineResult, verbose bool) {
	nonAISteps := make([]StepResult, 0, len(result.Results))
	var aiStep *StepResult

	for i := range result.Results {
		step := result.Results[i]
		if IsAIReviewStep(config.StepConfig{Type: step.Type}) {
			copied := step
			aiStep = &copied
			continue
		}
		nonAISteps = append(nonAISteps, step)
	}

	if len(nonAISteps) > 0 {
		builder.WriteString("Pipeline:\n")
		for _, step := range nonAISteps {
			builder.WriteString("  ")
			builder.WriteString(ui.Colorize(statusColor(step.Result.Status), statusIcon(step.Result.Status)+" "+step.Name))
			builder.WriteString(" (")
			builder.WriteString(stepSummary(step))
			builder.WriteString(")\n")
			if shouldPrintOutput(step.Result.Status, verbose) && strings.TrimSpace(step.Result.Output) != "" {
				builder.WriteString(indent(strings.TrimSpace(step.Result.Output), "    "))
				builder.WriteString("\n")
			}
		}
	}

	if aiStep != nil {
		if len(nonAISteps) > 0 {
			builder.WriteString("\n")
		}
		writeAIReviewOutput(builder, *aiStep)
	}

	separator := "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	builder.WriteString("\n")
	builder.WriteString(ui.Colorize(ui.BttfAmber, separator))
	builder.WriteString("\n")
	passedCount, failedAt := pipelineSummary(result)
	if result.Passed {
		builder.WriteString(ui.SuccessText(fmt.Sprintf("✓ Pipeline passed (%d/%d checkers) in %s", passedCount, len(result.Results), result.Duration)))
	} else {
		builder.WriteString(ui.ErrorText(fmt.Sprintf("✗ Pipeline failed at %q (%d/%d checkers passed) in %s", failedAt, passedCount, len(result.Results), result.Duration)))
	}
	builder.WriteString("\n")
	builder.WriteString(ui.Colorize(ui.BttfAmber, separator))
}

func writeAIReviewOutput(builder *strings.Builder, step StepResult) {
	builder.WriteString("AI Review:\n")
	builder.WriteString("  ")
	builder.WriteString(ui.Colorize(ui.BttfFlux, fmt.Sprintf("ℹ️  Sending to %s for review (timeout: %s)...", formatProvider(step), step.Timeout)))
	builder.WriteString("\n\n")
	builder.WriteString("  Files to review:\n")
	for _, file := range step.Files {
		builder.WriteString("    - ")
		builder.WriteString(file)
		builder.WriteString("\n")
	}
	builder.WriteString("\n")

	if strings.TrimSpace(step.Result.Output) != "" {
		builder.WriteString(indent(strings.TrimSpace(step.Result.Output), "  "))
		builder.WriteString("\n")
	}

	switch step.Result.Status {
	case checker.Passed:
		builder.WriteString("\n  ")
		builder.WriteString(ui.SuccessText("✅ CODE REVIEW PASSED"))
		builder.WriteString("\n")
	case checker.Failed:
		builder.WriteString("\n  ")
		builder.WriteString(ui.ErrorText("❌ CODE REVIEW FAILED"))
		builder.WriteString("\n")
	case checker.Error:
		builder.WriteString("\n  ")
		builder.WriteString(ui.WarningText("⚠️  CODE REVIEW ERROR"))
		builder.WriteString("\n")
	}
}

func pipelineSummary(result PipelineResult) (int, string) {
	passedCount := 0
	failedAt := ""

	for _, step := range result.Results {
		if step.Result.Status == checker.Passed {
			passedCount++
		}
		if failedAt == "" && (step.Result.Status == checker.Failed || step.Result.Status == checker.Error) {
			failedAt = step.Name
		}
	}

	return passedCount, failedAt
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
	if strings.TrimSpace(step.RulesFile) != "" {
		merged["rules_file"] = step.RulesFile
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
			Name:     steps[i].Name,
			Type:     steps[i].Type,
			Provider: steps[i].Provider,
			Model:    steps[i].Model,
			Timeout:  steps[i].Timeout,
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

// IsAIReviewStep reports whether a step is configured as an AI review checker.
func IsAIReviewStep(step config.StepConfig) bool {
	return isAIReview(step)
}

func isAIReview(step config.StepConfig) bool {
	t := strings.TrimSpace(step.Type)
	if t == "" {
		t = strings.TrimSpace(step.Checker)
	}
	return t == "ai-review" || t == "ai" || t == "prompt"
}

func printAIReviewStart(step config.StepConfig, files []string) {
	fmt.Fprintln(os.Stdout, "\nAI Review:")
	fmt.Fprintf(os.Stdout, "  %s\n", ui.Colorize(ui.BttfFlux, fmt.Sprintf("ℹ️  Sending to %s for review (timeout: %s)...", formatProviderStep(step), step.Timeout)))
	fmt.Fprintln(os.Stdout)
	fmt.Fprintln(os.Stdout, "  Files to review:")
	for _, file := range files {
		fmt.Fprintf(os.Stdout, "    - %s\n", file)
	}
	fmt.Fprintln(os.Stdout)
}

func printAIReviewResult(result checker.CheckResult) {
	output := strings.TrimSpace(result.Output)
	if output != "" {
		fmt.Fprintln(os.Stdout, output)
		fmt.Fprintln(os.Stdout)
	}

	switch result.Status {
	case checker.Passed:
		fmt.Fprintf(os.Stdout, "%s\n", ui.SuccessText("✅ CODE REVIEW PASSED"))
	case checker.Failed:
		fmt.Fprintf(os.Stdout, "%s\n", ui.ErrorText("❌ CODE REVIEW FAILED"))
		fmt.Fprintln(os.Stdout)
		fmt.Fprintln(os.Stdout, "Fix the violations listed above before committing.")
	case checker.Error:
		fmt.Fprintf(os.Stdout, "%s\n", ui.WarningText("⚠️  CODE REVIEW ERROR"))
	}
	fmt.Fprintln(os.Stdout)
}

func formatProvider(step StepResult) string {
	if strings.TrimSpace(step.Model) == "" {
		return step.Provider
	}
	return fmt.Sprintf("%s (%s)", step.Provider, step.Model)
}

func formatProviderStep(step config.StepConfig) string {
	if strings.TrimSpace(step.Model) == "" {
		return step.Provider
	}
	return fmt.Sprintf("%s (%s)", step.Provider, step.Model)
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
