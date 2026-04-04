package cmd

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunJSONFailureKeepsStdoutPureAndStderrSilent(t *testing.T) {
	repoDir := t.TempDir()
	initGitRepo(t, repoDir)
	configPath := filepath.Join(repoDir, ".vigilanty.yml")
	config := `version: "1"
global:
  fail_fast: true
pipeline:
  - name: fail
    checker: shell
    command: sh -c 'printf "failure details\n" >&2; exit 1'
  - name: skipped
    checker: shell
    command: printf 'ok\n'
`
	if err := os.WriteFile(configPath, []byte(config), 0o644); err != nil {
		t.Fatalf("os.WriteFile(%q) error = %v", configPath, err)
	}

	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd() error = %v", err)
	}
	if err := os.Chdir(repoDir); err != nil {
		t.Fatalf("os.Chdir(%q) error = %v", repoDir, err)
	}
	defer func() {
		_ = os.Chdir(oldWD)
	}()

	stdoutReader, stdoutWriter, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() stdout error = %v", err)
	}
	stderrReader, stderrWriter, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() stderr error = %v", err)
	}
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	os.Stdout = stdoutWriter
	os.Stderr = stderrWriter
	defer func() {
		os.Stdout = oldStdout
		os.Stderr = oldStderr
	}()

	rootCmd.SetArgs([]string{"--config", configPath, "run", "--json"})
	defer rootCmd.SetArgs(nil)

	err = Execute()
	_ = stdoutWriter.Close()
	_ = stderrWriter.Close()
	stdout, _ := io.ReadAll(stdoutReader)
	stderr, _ := io.ReadAll(stderrReader)
	_ = stdoutReader.Close()
	_ = stderrReader.Close()

	var exitErr *ExitError
	if err == nil {
		t.Fatal("Execute() error = nil, want ExitError")
	}
	if !errors.As(err, &exitErr) {
		t.Fatalf("Execute() error = %T, want *ExitError", err)
	}
	if exitErr.Code != ExitCheckerFailure {
		t.Fatalf("ExitError.Code = %d, want %d (err=%q stdout=%q stderr=%q)", exitErr.Code, ExitCheckerFailure, err.Error(), string(stdout), string(stderr))
	}
	if got := strings.TrimSpace(string(stderr)); got != "" {
		t.Fatalf("stderr = %q, want empty output", got)
	}
	if strings.Contains(string(stdout), "○ skipped") {
		t.Fatalf("stdout = %q, want pure JSON without live skipped output", string(stdout))
	}

	var payload struct {
		Passed bool `json:"passed"`
		Steps  []struct {
			Name   string `json:"name"`
			Status string `json:"status"`
		} `json:"steps"`
	}
	if err := json.Unmarshal(stdout, &payload); err != nil {
		t.Fatalf("json.Unmarshal(stdout) error = %v\nstdout: %s", err, stdout)
	}
	if payload.Passed {
		t.Fatal("payload.Passed = true, want false")
	}
	if len(payload.Steps) != 2 {
		t.Fatalf("len(payload.Steps) = %d, want 2", len(payload.Steps))
	}
	if payload.Steps[0].Status != "failed" && payload.Steps[0].Status != "error" {
		t.Fatalf("payload.Steps[0].Status = %q, want failed or error", payload.Steps[0].Status)
	}
	if payload.Steps[1].Status != "skipped" {
		t.Fatalf("payload.Steps[1].Status = %q, want skipped", payload.Steps[1].Status)
	}
}

func initGitRepo(t *testing.T, dir string) {
	t.Helper()

	cmd := exec.Command("git", "init")
	cmd.Dir = dir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git init failed: %v\n%s", err, string(output))
	}
}
