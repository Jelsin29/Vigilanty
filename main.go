package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/jelsin29/vigilanty/cmd"
	"github.com/jelsin29/vigilanty/internal/checker"
)

func main() {
	checker.Init()

	if err := cmd.Execute(); err != nil {
		var exitErr *cmd.ExitError
		if errors.As(err, &exitErr) {
			if exitErr.Err != nil {
				fmt.Fprintf(os.Stderr, "%s\n", exitErr.Err)
			}
			os.Exit(exitErr.Code)
		}

		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
