package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestIsRepo(t *testing.T) {
	repoDir := t.TempDir()
	runGitCommand(t, repoDir, "init")

	withWorkingDir(t, repoDir, func() {
		if !IsRepo() {
			t.Fatal("IsRepo() = false, want true")
		}
	})
}

func TestIsRepoReturnsFalseOutsideGitDirectory(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "not-a-repo")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("os.MkdirAll() error = %v", err)
	}

	withWorkingDir(t, dir, func() {
		if IsRepo() {
			t.Fatal("IsRepo() = true, want false")
		}
	})
}

func withWorkingDir(t *testing.T, dir string, fn func()) {
	t.Helper()

	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd() error = %v", err)
	}

	if err := os.Chdir(dir); err != nil {
		t.Fatalf("os.Chdir() error = %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(wd); err != nil {
			t.Fatalf("restore working directory: %v", err)
		}
	})

	fn()
}

func runGitCommand(t *testing.T, dir string, args ...string) {
	t.Helper()

	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v failed: %v\n%s", args, err, string(output))
	}
}
