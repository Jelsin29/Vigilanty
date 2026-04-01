package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jelsin/vigilanty/internal/config"
	"github.com/spf13/cobra"
)

func newInitCommand() *cobra.Command {
	var force bool
	var preset string

	command := &cobra.Command{
		Use:   "init",
		Short: "Create a .vigilanty.yml config file",
		RunE: func(cmd *cobra.Command, args []string) error {
			configPath := filepath.Join(".vigilanty.yml")

			content, err := config.ConfigYAMLForPreset(preset)
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
			} else if !os.IsNotExist(err) {
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
	command.Flags().StringVar(&preset, "preset", "", "Scaffold template preset: go, node, python, generic")
	return command
}

func confirmOverwrite(path string) (bool, error) {
	if info, err := os.Stdin.Stat(); err != nil {
		return false, fmt.Errorf("inspect stdin: %w", err)
	} else if (info.Mode() & os.ModeCharDevice) == 0 {
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
