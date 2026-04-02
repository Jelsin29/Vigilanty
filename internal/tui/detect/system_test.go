package detect

import (
	"context"
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
	if len(info.Tools) == 0 {
		t.Fatal("DetectSystem().Tools is empty")
	}

	if len(info.Tools) != 1 {
		t.Fatalf("DetectSystem().Tools len = %d, want 1", len(info.Tools))
	}

	gitTool, ok := toolStatusByName(info.Tools, "git")
	if !ok {
		t.Fatal("DetectSystem() did not include git tool")
	}
	if !gitTool.Found {
		t.Fatal("DetectSystem() reported git as missing")
	}
	if gitTool.Name != "git" {
		t.Fatalf("ToolStatus.Name = %q, want %q", gitTool.Name, "git")
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
