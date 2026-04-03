package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/jelsin29/vigilanty/internal/checker"
	"github.com/jelsin29/vigilanty/internal/config"
	gitpkg "github.com/jelsin29/vigilanty/internal/git"
	"github.com/jelsin29/vigilanty/internal/pipeline"
	"github.com/jelsin29/vigilanty/internal/ui"
	"github.com/spf13/cobra"
)

func newRunCommand() *cobra.Command {
	var noCache bool
	var prMode bool
	var baseBranch string
	var ciMode bool

	command := &cobra.Command{
		Use:   "run",
		Short: "Run the Vigilanty verification pipeline",
		RunE: func(cmd *cobra.Command, args []string) error {
			if prMode && ciMode {
				return newExitError(ExitConfigError, "%s", errorText("error: cannot use both --pr-mode and --ci"))
			}
			if baseBranch != "" && !prMode {
				return newExitError(ExitConfigError, "%s", errorText("error: --base requires --pr-mode"))
			}

			if !gitpkg.IsRepo() {
				return newExitError(ExitGitError, "%s", errorText("error: not a git repository. Run this command inside your repo."))
			}

			repoRoot, err := gitpkg.RepoRoot()
			if err != nil {
				return newExitError(ExitGitError, "%s", errorText(fmt.Sprintf("error: failed to resolve git repository root: %v", err)))
			}

			cfg, err := loadRunConfig()
			if err != nil {
				return err
			}

			if verbose {
				cfg.Global.Verbose = true
			}

			if ui.StdoutIsTTY() {
				printRunBanner(cfg, noCache)
			}

			if ciMode && !ui.StdoutIsTTY() {
				ui.SetColorsEnabled(false)
				defer ui.SetColorsEnabled(true)
			}

			if ciMode {
				if ciName, detected := gitpkg.DetectCI(); detected {
					fmt.Fprintf(os.Stdout, "Running in CI mode (%s detected)\n", ciName)
				}
			}

			diff, changedFiles, truncated, err := resolveRunDiff(repoRoot, cfg.Global.DiffMaxBytes, prMode, baseBranch, ciMode)
			if err != nil {
				return err
			}

			if truncated {
				fmt.Fprintf(os.Stdout, "%s\n", warningText(fmt.Sprintf("warning: diff exceeded %d bytes and was truncated", cfg.Global.DiffMaxBytes)))
			}

			if len(cfg.Pipeline) == 0 {
				fmt.Fprintf(os.Stdout, "%s\n", warningText("warning: pipeline is empty. Add steps to .vigilanty.yml and run again."))
				return nil
			}

			checkCtx := checker.CheckContext{
				Root:        repoRoot,
				Diff:        diff,
				StagedFiles: changedFiles,
			}

			runMode := "staged"
			if prMode {
				runMode = "pr"
			} else if ciMode {
				runMode = "ci"
			}

			pipe := pipeline.New(cfg, pipeline.Options{NoCache: noCache, Mode: runMode})
			result := pipe.Run(checkCtx)
			if result.LiveOutputted {
				fmt.Fprintf(os.Stdout, "%s\n", pipeline.FormatSummary(result))
			} else {
				fmt.Fprintf(os.Stdout, "%s\n", pipeline.FormatOutput(result, cfg.Global.Verbose))
			}

			if !result.Passed {
				return newExitError(ExitCheckerFailure, "")
			}

			return nil
		},
	}

	command.Flags().BoolVar(&noCache, "no-cache", false, "Bypass checker cache for this run")
	command.Flags().BoolVar(&prMode, "pr-mode", false, "Review full PR diff against base branch")
	command.Flags().StringVar(&baseBranch, "base", "", "Base branch for PR mode (default: auto-detect main/master)")
	command.Flags().BoolVar(&ciMode, "ci", false, "Review last commit changes (CI/CD mode)")
	return command
}

func printRunBanner(cfg *config.Config, noCache bool) {
	if cfg == nil {
		return
	}

	fmt.Fprintf(os.Stdout, "%s\n\n", ui.RunBanner(Version))

	if step, ok := findAIReviewStep(cfg.Pipeline); ok {
		provider := step.Provider
		if strings.TrimSpace(step.Model) != "" {
			provider = fmt.Sprintf("%s (%s)", provider, step.Model)
		}
		fmt.Fprintln(os.Stdout, ui.Colorize(ui.BttfFlux, fmt.Sprintf("ℹ️  Provider: %s", provider)))

		if strings.TrimSpace(step.RulesFile) != "" {
			fmt.Fprintln(os.Stdout, ui.Colorize(ui.BttfFlux, fmt.Sprintf("ℹ️  Rules file: %s", step.RulesFile)))
		}
	}

	if len(cfg.Global.FilePatterns) > 0 {
		fmt.Fprintln(os.Stdout, ui.Colorize(ui.BttfFlux, fmt.Sprintf("ℹ️  File patterns: %s", strings.Join(cfg.Global.FilePatterns, ", "))))
	}
	if len(cfg.Global.ExcludePatterns) > 0 {
		fmt.Fprintln(os.Stdout, ui.Colorize(ui.BttfFlux, fmt.Sprintf("ℹ️  Exclude: %s", strings.Join(cfg.Global.ExcludePatterns, ", "))))
	}

	cacheState := "enabled"
	if noCache {
		cacheState = "disabled"
	}
	fmt.Fprintln(os.Stdout, ui.Colorize(ui.BttfFlux, fmt.Sprintf("ℹ️  Cache: %s", cacheState)))
	fmt.Fprintln(os.Stdout)
}

func findAIReviewStep(steps []config.StepConfig) (config.StepConfig, bool) {
	for _, step := range steps {
		if pipeline.IsAIReviewStep(step) {
			return step, true
		}
	}

	return config.StepConfig{}, false
}

func resolveRunDiff(repoRoot string, maxBytes int, prMode bool, baseBranch string, ciMode bool) (string, []string, bool, error) {
	if prMode {
		diff, files, err := gitpkg.PRDiffWithFiles(repoRoot, baseBranch)
		if err != nil {
			return "", nil, false, newExitError(ExitGitError, "%s", errorText(fmt.Sprintf("error: failed to read PR diff: %v", err)))
		}

		truncatedDiff, truncated := truncateDiff(diff, maxBytes)
		return truncatedDiff, files, truncated, nil
	}

	if ciMode {
		diff, files, err := gitpkg.LastCommitDiffWithFiles(repoRoot)
		if err != nil {
			return "", nil, false, newExitError(ExitGitError, "%s", errorText(fmt.Sprintf("error: failed to read last commit diff: %v", err)))
		}

		truncatedDiff, truncated := truncateDiff(diff, maxBytes)
		return truncatedDiff, files, truncated, nil
	}

	stagedFiles, err := gitpkg.StagedFiles(repoRoot)
	if err != nil {
		return "", nil, false, newExitError(ExitGitError, "%s", errorText(fmt.Sprintf("error: failed to list staged files: %v", err)))
	}

	diff, truncated, err := gitpkg.StagedDiffTruncated(repoRoot, maxBytes)
	if err != nil {
		return "", nil, false, newExitError(ExitGitError, "%s", errorText(fmt.Sprintf("error: failed to read staged diff: %v", err)))
	}

	return diff, stagedFiles, truncated, nil
}

func truncateDiff(diff string, maxBytes int) (string, bool) {
	if maxBytes <= 0 || len(diff) <= maxBytes {
		return diff, false
	}

	truncated := diff[:maxBytes]
	for !utf8.ValidString(truncated) && len(truncated) > 0 {
		truncated = truncated[:len(truncated)-1]
	}

	return truncated + "\n... diff truncated ...\n", true
}

func loadRunConfig() (*config.Config, error) {
	configPath, err := resolveRunConfigPath()
	if err != nil {
		return nil, err
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		return nil, newExitError(ExitConfigError, "%s", errorText(fmt.Sprintf("error: failed to load config %q: %v", configPath, err)))
	}

	return cfg, nil
}

func resolveRunConfigPath() (string, error) {
	if configFile != "" {
		return configFile, nil
	}

	localPath := filepath.Join(".vigilanty.yml")
	if _, err := os.Stat(localPath); err == nil {
		return localPath, nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return "", newExitError(ExitConfigError, "%s", errorText(fmt.Sprintf("error: cannot inspect %s: %v", localPath, err)))
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", newExitError(ExitConfigError, "%s", errorText(fmt.Sprintf("error: failed to resolve home directory: %v", err)))
	}

	globalPath := filepath.Join(homeDir, ".config", "vigilanty", "config.yml")
	if _, err := os.Stat(globalPath); err == nil {
		return globalPath, nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return "", newExitError(ExitConfigError, "%s", errorText(fmt.Sprintf("error: cannot inspect %s: %v", globalPath, err)))
	}

	return "", newExitError(ExitConfigError, "%s", errorText("error: no Vigilanty config file found. Run 'vigilanty init' to create .vigilanty.yml."))
}
