package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"

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

			// check if golangci-lint executable is available, else go install it if go version is >= 1.18
			if err := installLinter(); err != nil {
				fmt.Fprintf(os.Stdout, "%s\n", errorText(fmt.Sprintf("warning: failed to install golangci-lint: %v", err)))
			}

			fmt.Fprintf(os.Stdout, "Next step: stage changes and run 'vigilanty run' to verify the pipeline.\n")
			return nil
		},
	}
}

func installLinter() error {
	if _, err := exec.LookPath("golangci-lint"); err == nil {
		return nil // Already installed
	}

	// check if minor version >= 18
	var minor int
	_, _ = fmt.Sscanf(runtime.Version(), "go1.%d", &minor)

	if minor >= 18 {
		cmd := exec.Command("go", "install", "github.com/golangci/golangci-lint/cmd/golangci-lint@latest")
		cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
		return cmd.Run()
	}

	return fmt.Errorf("go version %s too old; 1.18+ required", runtime.Version())
}
