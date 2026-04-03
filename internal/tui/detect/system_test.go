package detect

import (
	"context"
	"os/exec"
	"runtime"
	"testing"
)

func TestDetectSystemReturnsRuntimeInfo(t *testing.T) {
	info := DetectSystem(context.Background())

	if info.OS != runtime.GOOS {
		t.Fatalf("DetectSystem().OS = %q, want %q", info.OS, runtime.GOOS)
	}

	if info.Arch != runtime.GOARCH {
		t.Fatalf("DetectSystem().Arch = %q, want %q", info.Arch, runtime.GOARCH)
	}

	// Tools slice always has git entry, but Found may be false in minimal envs
	if len(info.Tools) == 0 {
		t.Fatal("DetectSystem().Tools is empty")
	}

	gitTool, ok := toolStatusByName(info.Tools, "git")
	if !ok {
		t.Fatal("DetectSystem() did not include git tool")
	}

	// Only assert git details when git is actually on PATH
	if _, err := exec.LookPath("git"); err != nil {
		t.Log("git not in PATH — skipping Found/Version assertions")
		return
	}

	if !gitTool.Found {
		t.Fatal("DetectSystem() reported git as missing")
	}
	if gitTool.Version == "" {
		t.Fatal("DetectSystem() returned empty version for git")
	}
}

func toolStatusByName(tools []ToolStatus, name string) (ToolStatus, bool) {
	for _, tool := range tools {
		if tool.Name == name {
			return tool, true
		}
	}

	return ToolStatus{}, false
}
