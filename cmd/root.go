package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	configFile string
	verbose    bool
	Version    = "dev"
)

const (
	ExitSuccess        = 0
	ExitCheckerFailure = 1
	ExitConfigError    = 2
	ExitGitError       = 3
	ExitInternalError  = 4
)

var rootCmd = &cobra.Command{
	Use:               "vigilanty",
	Short:             "Vigilanty runs configurable verification pipelines",
	Long:              "Vigilanty is a configurable pre-commit verification CLI for staged changes.",
	Version:           Version,
	SilenceUsage:      true,
	CompletionOptions: cobra.CompletionOptions{DisableDefaultCmd: true},
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

type ExitError struct {
	Code int
	Err  error
}

func (e *ExitError) Error() string {
	if e == nil || e.Err == nil {
		return ""
	}
	return e.Err.Error()
}

func newExitError(code int, format string, args ...interface{}) error {
	if code <= 0 {
		code = ExitCheckerFailure
	}

	if format == "" {
		return &ExitError{Code: code}
	}

	return &ExitError{
		Code: code,
		Err:  fmt.Errorf(format, args...),
	}
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVar(&configFile, "config", "", "Path to the Vigilanty config file")
	rootCmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "Print verbose checker output")
	rootCmd.Version = Version
	rootCmd.SetVersionTemplate("{{.Name}} {{.Version}}\n")
	rootCmd.AddCommand(newInitCommand())
	rootCmd.AddCommand(newInstallCommand())
	rootCmd.AddCommand(newUninstallCommand())
	rootCmd.AddCommand(newCacheCommand())
	rootCmd.AddCommand(newRunCommand())
	rootCmd.AddCommand(newConfigCommand())
	rootCmd.AddCommand(newVersionCommand())
}
