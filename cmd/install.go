package cmd

import (
	"fmt"
	"os"

	gitpkg "github.com/Jelsin29/Vigilanty/internal/git"
	"github.com/spf13/cobra"
)

func newInstallCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "install",
		Short: "Install the Vigilanty pre-commit hook",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !gitpkg.IsRepo() {
				return newExitError(1, "%s", errorText("error: not a git repository. Run this command inside your repo."))
			}

			if err := gitpkg.InstallHook("pre-commit"); err != nil {
				return newExitError(1, "%s", errorText(fmt.Sprintf("error: failed to install pre-commit hook: %v", err)))
			}

			fmt.Fprintf(os.Stdout, "%s\n", successText("Installed Vigilanty pre-commit hook"))
			fmt.Fprintf(os.Stdout, "Next step: stage changes and run 'vigilanty run' to verify the pipeline.\n")
			return nil
		},
	}
}
