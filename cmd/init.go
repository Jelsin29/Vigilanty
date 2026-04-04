package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/jelsin29/vigilanty/internal/config"
	"github.com/jelsin29/vigilanty/internal/rules"
	"github.com/jelsin29/vigilanty/internal/wizard"
	"github.com/spf13/cobra"
)

func newInitCommand() *cobra.Command {
	var force bool
	var preset string
	var noInteractive bool

	command := &cobra.Command{
		Use:   "init",
		Short: "Create a .vigilanty.yml config file",
		RunE: func(cmd *cobra.Command, args []string) error {
			configPath := filepath.Join(".vigilanty.yml")
			selectedPreset := preset
			var result *wizard.InitResult

			stdinIsTTY, err := stdinIsTTY()
			if err != nil {
				return newExitError(1, "%s", errorText(fmt.Sprintf("error: %v", err)))
			}

			if strings.TrimSpace(selectedPreset) == "" {
				selectedPreset = detectProjectType()
			}

			interactive := stdinIsTTY && !noInteractive

			var content string
			if interactive {
				var runErr error
				result, runErr = wizard.Run(selectedPreset)
				if runErr != nil {
					if errors.Is(runErr, wizard.ErrCancelled) {
						fmt.Fprintln(os.Stdout, "Setup cancelled.")
						return nil
					}
					return newExitError(1, "%s", errorText(fmt.Sprintf("error: %v", runErr)))
				}

				content, err = config.ConfigYAMLForWizardResult(result)
			} else {
				content, err = config.ConfigYAMLForPreset(selectedPreset)
			}
			if err != nil {
				return newExitError(1, "%s", errorText(fmt.Sprintf("error: %v", err)))
			}

			if err := printPresetPreview(os.Stdout, selectedPreset, content); err != nil {
				return newExitError(1, "%s", errorText(fmt.Sprintf("error: %v", err)))
			}

			if _, err := os.Stat(configPath); err == nil {
				if !force {
					confirmed, confirmErr := confirmOverwrite(configPath)
					if confirmErr != nil {
						return newExitError(1, "%s", errorText(fmt.Sprintf("error: %v", confirmErr)))
					}
					if !confirmed {
						fmt.Fprintf(os.Stdout, "%s\n", warningText("warning: keeping existing .vigilanty.yml"))
						return nil
					}
				}
			} else if !errors.Is(err, os.ErrNotExist) {
				return newExitError(1, "%s", errorText(fmt.Sprintf("error: cannot inspect %s: %v", configPath, err)))
			}

			if err := os.WriteFile(configPath, []byte(content), 0o644); err != nil {
				return newExitError(1, "%s", errorText(fmt.Sprintf("error: cannot create %s: %v", configPath, err)))
			}

			if result != nil && result.GenerateRules {
				root, err := os.Getwd()
				if err != nil {
					return newExitError(1, "%s", errorText(fmt.Sprintf("error: cannot inspect project root: %v", err)))
				}

				rulesPath := strings.TrimSpace(result.RulesFile)
				if rulesPath == "" {
					rulesPath = "AGENTS.md"
				}

				rulesContent := rules.Generate(rules.GenerateOptions{
					ProjectType:     result.ProjectType,
					FilePatterns:    result.FilePatterns,
					ExcludePatterns: result.ExcludePatterns,
					DetectedTools:   rules.DetectTools(root),
				})
				if err := os.WriteFile(rulesPath, []byte(rulesContent), 0o644); err != nil {
					return newExitError(1, "%s", errorText(fmt.Sprintf("error: cannot create %s: %v", rulesPath, err)))
				}
				fmt.Fprintf(os.Stdout, "%s\n", successText(fmt.Sprintf("Created %s", rulesPath)))
			}

			fmt.Fprintf(os.Stdout, "%s\n", successText("Created .vigilanty.yml"))
			printInitNextSteps(os.Stdout)
			return nil
		},
	}

	command.Flags().BoolVar(&force, "force", false, "Overwrite .vigilanty.yml without prompting")
	command.Flags().BoolVar(&noInteractive, "no-interactive", false, "Skip the interactive setup wizard")
	command.Flags().StringVar(&preset, "preset", "", "Scaffold template preset: go, node, typescript, python, rust, java, dotnet, ruby, swift, php, generic")
	return command
}

func printPresetPreview(out io.Writer, preset string, content string) error {
	cfg, err := config.LoadFromReader(strings.NewReader(content))
	if err != nil {
		return fmt.Errorf("parse generated preset: %w", err)
	}

	presetName := strings.TrimSpace(preset)
	if presetName == "" {
		presetName = detectProjectTypeName(content, cfg)
	}

	steps := make([]string, 0, len(cfg.Pipeline))
	for _, step := range cfg.Pipeline {
		label := step.Name
		if strings.TrimSpace(step.Command) != "" {
			label = fmt.Sprintf("%s — %s", step.Name, step.Command)
		} else if pipeline := strings.TrimSpace(step.Provider); pipeline != "" {
			label = fmt.Sprintf("%s — %s", step.Name, pipeline)
		}
		steps = append(steps, label)
	}

	fmt.Fprintf(out, "Preset preview (%s):\n", presetName)
	for _, step := range steps {
		fmt.Fprintf(out, "  - %s\n", step)
	}
	fmt.Fprintln(out)
	return nil
}

func detectProjectTypeName(content string, cfg *config.Config) string {
	if cfg == nil {
		return "detected"
	}

	for _, step := range cfg.Pipeline {
		switch step.Name {
		case "golangci-lint":
			return "go"
		case "eslint":
			return "node"
		case "ruff":
			return "python"
		}
	}

	if strings.TrimSpace(content) == "" {
		return "detected"
	}

	return "detected"
}

func printInitNextSteps(out io.Writer) {
	fmt.Fprintf(out, "Next Steps:\n")
	fmt.Fprintf(out, "  1. Review .vigilanty.yml and adjust the generated steps if needed\n")
	fmt.Fprintf(out, "  2. Run 'vigilanty install' to add the pre-commit hook\n")
	fmt.Fprintf(out, "  3. Run 'vigilanty run' to verify your current staged changes\n")
	fmt.Fprintf(out, "  4. Use 'vigilanty run --json' when you need machine-readable CI output\n")
}

func detectProjectType() string {
	languages := []map[string][]string{
		{"typescript": {"tsconfig.json"}},
		{"go": {"go.mod"}},
		{"rust": {"Cargo.toml"}},
		{"swift": {"Package.swift"}},
		{"dotnet": {"*.csproj", "*.sln"}},
		{"java": {"pom.xml", "build.gradle", "build.gradle.kts"}},
		{"ruby": {"Gemfile"}},
		{"php": {"composer.json"}},
		{"python": {"requirements.txt", "pyproject.toml", "setup.py"}},
		{"node": {"package.json"}},
	}

	for _, lang := range languages {
		for preset, files := range lang {
			for _, file := range files {
				if fileExists(file) {
					fmt.Fprintf(os.Stdout, "Detected %s project (found %s), using '%s' preset\n", preset, file, preset)
					return preset
				}
			}
		}
	}

	fmt.Fprintf(os.Stdout, "No project type detected, using 'generic' preset\n")
	return "generic"
}

func fileExists(path string) bool {
	if _, err := os.Stat(path); err == nil {
		return true
	}
	if matches, err := filepath.Glob(path); err == nil && len(matches) > 0 {
		return true
	}
	return false
}

func confirmOverwrite(path string) (bool, error) {
	isTTY, err := stdinIsTTY()
	if err != nil {
		return false, err
	}
	if !isTTY {
		return false, fmt.Errorf("%s already exists. Re-run with --force to overwrite in non-interactive mode", path)
	}

	fmt.Fprintf(os.Stdout, "%s", warningText(fmt.Sprintf("warning: %s already exists. Overwrite? [y/N]: ", path)))
	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false, fmt.Errorf("read confirmation: %w", err)
	}

	answer := strings.ToLower(strings.TrimSpace(response))
	return answer == "y" || answer == "yes", nil
}

func stdinIsTTY() (bool, error) {
	info, err := os.Stdin.Stat()
	if err != nil {
		return false, fmt.Errorf("inspect stdin: %w", err)
	}

	return (info.Mode() & os.ModeCharDevice) != 0, nil
}
