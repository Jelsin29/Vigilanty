//go:build windows

package checker

import "os/exec"

func configureCommand(cmd *exec.Cmd) {}

func terminateCommand(cmd *exec.Cmd) error {
	if cmd == nil || cmd.Process == nil {
		return nil
	}

	return cmd.Process.Kill()
}
