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

	goTool, ok := toolStatusByName(info.Tools, "go")
	if !ok {
		t.Fatal("DetectSystem() did not include go tool")
	}
	if !goTool.Found {
		t.Fatal("DetectSystem() reported go as missing")
	}
	if goTool.Name != "go" {
		t.Fatalf("ToolStatus.Name = %q, want %q", goTool.Name, "go")
	}
	if goTool.Version == "" {
		t.Fatal("DetectSystem() returned empty version for go")
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
