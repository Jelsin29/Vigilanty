package detect

import (
	"bufio"
	"context"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"
)

type SystemInfo struct {
	OS    string
	Arch  string
	Shell string
	Tools []ToolStatus
}

type ToolStatus struct {
	Name    string
	Found   bool
	Version string
}

var (
	lookPath       = exec.LookPath
	commandContext = exec.CommandContext
	getenv         = os.Getenv
)

func DetectSystem(ctx context.Context) SystemInfo {
	info := SystemInfo{
		OS:    runtime.GOOS,
		Arch:  runtime.GOARCH,
		Shell: getenv("SHELL"),
		Tools: make([]ToolStatus, 0, 5),
	}

	tools := []string{"go", "node", "git", "curl", "python3"}
	results := make([]ToolStatus, len(tools))

	var wg sync.WaitGroup
	for i, name := range tools {
		if ctx.Err() != nil {
			break
		}

		wg.Add(1)
		go func(i int, name string) {
			defer wg.Done()
			results[i] = detectTool(ctx, name)
		}(i, name)
	}
	wg.Wait()

	for _, tool := range results {
		if tool.Name == "" {
			continue
		}
		info.Tools = append(info.Tools, tool)
	}

	return info
}

func detectTool(ctx context.Context, name string) ToolStatus {
	status := ToolStatus{Name: name}
	if ctx.Err() != nil {
		return status
	}

	path, err := lookPath(name)
	if err != nil {
		return status
	}
	status.Found = true
	status.Version = detectCommandVersion(ctx, path)

	return status
}

func detectCommandVersion(ctx context.Context, binary string) string {
	if ctx.Err() != nil {
		return ""
	}

	for _, args := range [][]string{{"--version"}, {"version"}} {
		vctx, cancel := context.WithTimeout(ctx, 2*time.Second)
		cmd := commandContext(vctx, binary, args...)
		out, err := cmd.CombinedOutput()
		cancel()
		if err != nil {
			continue
		}

		line := firstLine(string(out))
		if line != "" {
			return line
		}
	}

	return ""
}

func firstLine(raw string) string {
	s := bufio.NewScanner(strings.NewReader(raw))
	if !s.Scan() {
		return ""
	}

	return strings.TrimSpace(s.Text())
}
