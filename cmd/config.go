package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"

	configpkg "github.com/jelsin29/vigilanty/internal/config"
	"github.com/spf13/cobra"
)

func newConfigCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "config",
		Short: "Show the active Vigilanty configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			configPath, err := resolveRunConfigPath()
			if err != nil {
				return err
			}

			cfg, err := configpkg.Load(configPath)
			if err != nil {
				return newExitError(ExitConfigError, "%s", errorText(fmt.Sprintf("error: failed to load config %q: %v", configPath, err)))
			}

			printConfig(os.Stdout, configPath, cfg)
			return nil
		},
	}
}

func printConfig(out io.Writer, source string, cfg *configpkg.Config) {
	if cfg == nil {
		return
	}

	fmt.Fprintf(out, "Configuration\n")
	fmt.Fprintf(out, "  Source:   %s\n", source)
	fmt.Fprintf(out, "  Version:  %s\n", cfg.Version)

	fmt.Fprintf(out, "\nGlobal Settings\n")
	fmt.Fprintf(out, "  %-15s %t\n", "fail_fast:", cfg.Global.FailFast)
	fmt.Fprintf(out, "  %-15s %d\n", "diff_max_bytes:", cfg.Global.DiffMaxBytes)
	fmt.Fprintf(out, "  %-15s %s\n", "timeout:", cfg.Global.Timeout)
	fmt.Fprintf(out, "  %-15s %t\n", "verbose:", cfg.Global.Verbose)

	fmt.Fprintf(out, "\nPipeline (%d steps)\n", len(cfg.Pipeline))
	for index, step := range cfg.Pipeline {
		fmt.Fprintf(out, "  %d. %-16s [%-9s] timeout: %s", index+1, step.Name, stepTypeLabel(step), step.Timeout)

		extra := stepSettings(step)
		if len(extra) > 0 {
			fmt.Fprintf(out, "  %s", strings.Join(extra, "  "))
		}
		fmt.Fprintln(out)
	}
}

func stepTypeLabel(step configpkg.StepConfig) string {
	if strings.TrimSpace(step.Type) != "" {
		return step.Type
	}
	return step.Checker
}

func stepSettings(step configpkg.StepConfig) []string {
	extra := make([]string, 0, 6)

	if strings.TrimSpace(step.Provider) != "" {
		extra = append(extra, fmt.Sprintf("provider: %s", step.Provider))
	}
	if strings.TrimSpace(step.Model) != "" {
		extra = append(extra, fmt.Sprintf("model: %s", step.Model))
	}
	if step.Enabled != nil {
		extra = append(extra, fmt.Sprintf("enabled: %t", *step.Enabled))
	}
	if step.SkipOnEmptyDiff {
		extra = append(extra, "skip_on_empty_diff: true")
	}
	if step.MaxDiffLines > 0 {
		extra = append(extra, fmt.Sprintf("max_diff_lines: %d", step.MaxDiffLines))
	}

	return extra
}
