package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jelsin29/vigilanty/internal/config"
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
			interactive := false

			stdinIsTTY, err := stdinIsTTY()
			if err != nil {
				return newExitError(1, "%s", errorText(fmt.Sprintf("error: %v", err)))
			}

			if strings.TrimSpace(selectedPreset) == "" {
				selectedPreset = detectProjectType()
				if stdinIsTTY && !noInteractive {
					interactive = true
				} else {
					printDetectedPresetMessage(selectedPreset)
				}
			}

			var content string
			if interactive {
				result, runErr := wizard.Run(selectedPreset)
				if runErr != nil {
					return newExitError(1, "%s", errorText(fmt.Sprintf("error: %v", runErr)))
				}

				content, err = config.ConfigYAMLForWizardResult(result)
			} else {
				content, err = config.ConfigYAMLForPreset(selectedPreset)
			}
			if err != nil {
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

			fmt.Fprintf(os.Stdout, "%s\n", successText("Created .vigilanty.yml"))
			fmt.Fprintf(os.Stdout, "Next steps:\n")
			fmt.Fprintf(os.Stdout, "  1. Edit .vigilanty.yml to define your pipeline\n")
			fmt.Fprintf(os.Stdout, "  2. Run 'vigilanty install' to add the pre-commit hook\n")
			fmt.Fprintf(os.Stdout, "  3. Run 'vigilanty run' to verify the current staged changes\n")
			return nil
		},
	}

	command.Flags().BoolVar(&force, "force", false, "Overwrite .vigilanty.yml without prompting")
	command.Flags().BoolVar(&noInteractive, "no-interactive", false, "Skip the interactive setup wizard")
	command.Flags().StringVar(&preset, "preset", "", "Scaffold template preset: go, node, typescript, python, rust, java, dotnet, ruby, swift, php, generic")
	return command
}

func detectProjectType() string {
	if fileExists("tsconfig.json") {
		return "typescript"
	}

	if fileExists("go.mod") {
		return "go"
	}

	if fileExists("Cargo.toml") {
		return "rust"
	}

	if fileExists("Package.swift") {
		return "swift"
	}

	if hasDotnetProjectFiles() {
		return "dotnet"
	}

	if fileExists("pom.xml") || fileExists("build.gradle") || fileExists("build.gradle.kts") {
		return "java"
	}

	if fileExists("Gemfile") {
		return "ruby"
	}

	if fileExists("composer.json") {
		return "php"
	}

	if fileExists("requirements.txt") || fileExists("pyproject.toml") || fileExists("setup.py") {
		return "python"
	}

	if fileExists("package.json") {
		return "node"
	}

	return "generic"
}

func printDetectedPresetMessage(preset string) {
	switch preset {
	case "typescript":
		fmt.Fprintf(os.Stdout, "Detected TypeScript project (found tsconfig.json), using 'typescript' preset\n")
	case "go":
		fmt.Fprintf(os.Stdout, "Detected Go project (found go.mod), using 'go' preset\n")
	case "rust":
		fmt.Fprintf(os.Stdout, "Detected Rust project (found Cargo.toml), using 'rust' preset\n")
	case "swift":
		fmt.Fprintf(os.Stdout, "Detected Swift project (found Package.swift), using 'swift' preset\n")
	case "dotnet":
		switch {
		case hasMatchingFile("*.csproj"):
			fmt.Fprintf(os.Stdout, "Detected .NET project (found .csproj file), using 'dotnet' preset\n")
		case hasMatchingFile("*.sln"):
			fmt.Fprintf(os.Stdout, "Detected .NET project (found .sln file), using 'dotnet' preset\n")
		default:
			fmt.Fprintf(os.Stdout, "Detected .NET project, using 'dotnet' preset\n")
		}
	case "java":
		switch {
		case fileExists("pom.xml"):
			fmt.Fprintf(os.Stdout, "Detected Java project (found pom.xml), using 'java' preset\n")
		case fileExists("build.gradle"):
			fmt.Fprintf(os.Stdout, "Detected Java project (found build.gradle), using 'java' preset\n")
		case fileExists("build.gradle.kts"):
			fmt.Fprintf(os.Stdout, "Detected Java project (found build.gradle.kts), using 'java' preset\n")
		default:
			fmt.Fprintf(os.Stdout, "Detected Java project, using 'java' preset\n")
		}
	case "ruby":
		fmt.Fprintf(os.Stdout, "Detected Ruby project (found Gemfile), using 'ruby' preset\n")
	case "php":
		fmt.Fprintf(os.Stdout, "Detected PHP project (found composer.json), using 'php' preset\n")
	case "node":
		fmt.Fprintf(os.Stdout, "Detected Node project (found package.json), using 'node' preset\n")
	case "python":
		switch {
		case fileExists("requirements.txt"):
			fmt.Fprintf(os.Stdout, "Detected Python project (found requirements.txt), using 'python' preset\n")
		case fileExists("pyproject.toml"):
			fmt.Fprintf(os.Stdout, "Detected Python project (found pyproject.toml), using 'python' preset\n")
		case fileExists("setup.py"):
			fmt.Fprintf(os.Stdout, "Detected Python project (found setup.py), using 'python' preset\n")
		default:
			fmt.Fprintf(os.Stdout, "Detected Python project, using 'python' preset\n")
		}
	default:
		fmt.Fprintf(os.Stdout, "No project type detected, using 'generic' preset\n")
	}
}

func hasDotnetProjectFiles() bool {
	return hasMatchingFile("*.csproj") || hasMatchingFile("*.sln")
}

func hasMatchingFile(pattern string) bool {
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return false
	}

	return len(matches) > 0
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
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
