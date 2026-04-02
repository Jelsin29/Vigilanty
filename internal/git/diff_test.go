package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestFilesFromDiffUsesLastBMarker(t *testing.T) {
	diff := strings.Join([]string{
		"diff --git a/dir b/name.txt b/renamed.txt",
		"index 1111111..2222222 100644",
		"--- a/dir b/name.txt",
		"+++ b/renamed.txt",
	}, "\n")

	files := filesFromDiff(diff)
	if len(files) != 1 {
		t.Fatalf("len(filesFromDiff()) = %d, want 1", len(files))
	}
	if files[0] != "renamed.txt" {
		t.Fatalf("filesFromDiff()[0] = %q, want %q", files[0], "renamed.txt")
	}
}

func TestLastCommitDiffWithFilesUsesEmptyTreeForFirstCommit(t *testing.T) {
	repoDir := t.TempDir()
	runGitCommand(t, repoDir, "init")
	writeTestFile(t, filepath.Join(repoDir, "README.md"), "hello\n")
	runGitCommandWithEnv(t, repoDir, []string{
		"GIT_AUTHOR_NAME=Test User",
		"GIT_AUTHOR_EMAIL=test@example.com",
		"GIT_COMMITTER_NAME=Test User",
		"GIT_COMMITTER_EMAIL=test@example.com",
	}, "add", "README.md")
	runGitCommandWithEnv(t, repoDir, []string{
		"GIT_AUTHOR_NAME=Test User",
		"GIT_AUTHOR_EMAIL=test@example.com",
		"GIT_COMMITTER_NAME=Test User",
		"GIT_COMMITTER_EMAIL=test@example.com",
	}, "commit", "-m", "initial")

	diff, files, err := LastCommitDiffWithFiles(repoDir)
	if err != nil {
		t.Fatalf("LastCommitDiffWithFiles() error = %v", err)
	}
	if !strings.Contains(diff, "diff --git a/README.md b/README.md") {
		t.Fatalf("LastCommitDiffWithFiles() diff = %q, want initial commit diff", diff)
	}
	if len(files) != 1 || files[0] != "README.md" {
		t.Fatalf("LastCommitDiffWithFiles() files = %v, want [README.md]", files)
	}
}

func writeTestFile(t *testing.T, path string, content string) {
	t.Helper()

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}
}

func runGitCommandWithEnv(t *testing.T, dir string, env []string, args ...string) {
	t.Helper()

	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), env...)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v failed: %v\n%s", args, err, string(output))
	}
}
