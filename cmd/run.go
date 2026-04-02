package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Jelsin29/Vigilanty/internal/checker"
	"github.com/Jelsin29/Vigilanty/internal/config"
	gitpkg "github.com/Jelsin29/Vigilanty/internal/git"
	"github.com/Jelsin29/Vigilanty/internal/pipeline"
	"github.com/spf13/cobra"
)

func newRunCommand() *cobra.Command {
	var noCache bool

	command := &cobra.Command{
		Use:   "run",
		Short: "Run the Vigilanty verification pipeline",
		RunE: func(cmd *cobra.Command, args []string) error {
			_ = noCache

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

			stagedFiles, err := gitpkg.StagedFiles()
			if err != nil {
				return newExitError(ExitGitError, "%s", errorText(fmt.Sprintf("error: failed to list staged files: %v", err)))
			}

			diff, truncated, err := gitpkg.StagedDiffTruncated(cfg.Global.DiffMaxBytes)
			if err != nil {
				return newExitError(ExitGitError, "%s", errorText(fmt.Sprintf("error: failed to read staged diff: %v", err)))
			}

			if truncated {
				fmt.Fprintf(os.Stdout, "%s\n", warningText(fmt.Sprintf("warning: staged diff exceeded %d bytes and was truncated", cfg.Global.DiffMaxBytes)))
			}

			if len(cfg.Pipeline) == 0 {
				fmt.Fprintf(os.Stdout, "%s\n", warningText("warning: pipeline is empty. Add steps to .vigilanty.yml and run again."))
				return nil
			}

			checkCtx := checker.CheckContext{
				Root:        repoRoot,
				Diff:        diff,
				StagedFiles: stagedFiles,
			}

			pipe := pipeline.New(cfg)
			result := pipe.Run(checkCtx)
			fmt.Fprintf(os.Stdout, "%s\n", pipeline.FormatOutput(result, cfg.Global.Verbose))

			if !result.Passed {
				return newExitError(ExitCheckerFailure, "")
			}

			return nil
		},
	}

	command.Flags().BoolVar(&noCache, "no-cache", false, "Accepted for compatibility; ignored in v0.1")
	return command
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
