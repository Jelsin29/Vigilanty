package cmd

import (
	"fmt"
	"os"

	gitpkg "github.com/jelsin/vigilanty/internal/git"
	"github.com/spf13/cobra"
)

func newUninstallCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "uninstall",
		Short: "Remove the Vigilanty pre-commit hook",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !gitpkg.IsRepo() {
				return newExitError(1, "%s", errorText("error: not a git repository. Run this command inside your repo."))
			}

			if err := gitpkg.UninstallHook(); err != nil {
				return newExitError(1, "%s", errorText(fmt.Sprintf("error: failed to uninstall pre-commit hook: %v", err)))
			}

			fmt.Fprintf(os.Stdout, "%s\n", successText("Removed Vigilanty pre-commit hook"))
			fmt.Fprintf(os.Stdout, "Next step: run 'vigilanty install' if you want to enable it again.\n")
			return nil
		},
	}
}
