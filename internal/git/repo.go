package git

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

func IsRepo() bool {
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	return strings.TrimSpace(string(output)) == "true"
}

func RepoRoot() (string, error) {
	output, err := exec.Command("git", "rev-parse", "--show-toplevel").CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("resolve git repo root: %w: %s", err, strings.TrimSpace(string(output)))
	}

	return strings.TrimSpace(string(output)), nil
}

func HooksDir() (string, error) {
	root, err := RepoRoot()
	if err != nil {
		return "", err
	}

	output, err := exec.Command("git", "rev-parse", "--git-path", "hooks").CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("resolve git hooks dir: %w: %s", err, strings.TrimSpace(string(output)))
	}

	hooksPath := strings.TrimSpace(string(output))
	if filepath.IsAbs(hooksPath) {
		return hooksPath, nil
	}

	return filepath.Join(root, hooksPath), nil
}
