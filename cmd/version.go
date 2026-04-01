package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func newVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the Vigilanty version",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintf(os.Stdout, "vigilanty v%s\n", Version)
			return nil
		},
	}
}
