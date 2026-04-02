package checker

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"sort"
	"strings"
	"time"
)

const defaultShellTimeout = 60 * time.Second

type ShellChecker struct {
	command         string
	timeout         time.Duration
	env             map[string]string
	skipOnEmptyDiff bool
}

func newShellChecker(cfg map[string]interface{}) (Checker, error) {
	command, ok := cfg["command"].(string)
	if !ok || strings.TrimSpace(command) == "" {
		return nil, fmt.Errorf("shell checker command is required and must be a string")
	}

	timeout, err := durationConfigValue(cfg, "timeout", defaultShellTimeout)
	if err != nil {
		return nil, fmt.Errorf("shell checker timeout: %w", err)
	}

	env, err := envConfigValue(cfg, "env")
	if err != nil {
		return nil, fmt.Errorf("shell checker env: %w", err)
	}

	skipOnEmptyDiff, err := boolConfigValue(cfg, "skip_on_empty_diff", false)
	if err != nil {
		return nil, fmt.Errorf("shell checker skip_on_empty_diff: %w", err)
	}

	return &ShellChecker{
		command:         command,
		timeout:         timeout,
		env:             env,
		skipOnEmptyDiff: skipOnEmptyDiff,
	}, nil
}

func (c *ShellChecker) Name() string {
	return "shell"
}

func (c *ShellChecker) Check(ctx CheckContext) CheckResult {
	startedAt := time.Now()

	if c.skipOnEmptyDiff && strings.TrimSpace(ctx.Diff) == "" {
		return CheckResult{
			Status:   Skipped,
			Output:   "skipped (empty diff)",
			Duration: time.Since(startedAt),
		}
	}

	execCtx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	cmd := exec.CommandContext(execCtx, shellProgram(), shellArgs(c.command)...)
	configureCommand(cmd)
	if strings.TrimSpace(ctx.Root) != "" {
		cmd.Dir = ctx.Root
	}
	cmd.Env = mergedEnv(c.env)

	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &output

	done := make(chan struct{})
	started := make(chan struct{})
	go func() {
		select {
		case <-execCtx.Done():
			<-started
			if errors.Is(execCtx.Err(), context.DeadlineExceeded) {
				_ = terminateCommand(cmd)
			}
		case <-done:
		}
	}()

	if err := cmd.Start(); err != nil {
		close(started)
		close(done)
		return CheckResult{
			Status: Error,
			Output: fmt.Sprintf("failed to start command: %v", err),
		}
	}
	close(started)

	err := cmd.Wait()
	close(done)
	result := CheckResult{
		Output:   output.String(),
		Duration: time.Since(startedAt),
	}

	if err == nil {
		result.Status = Passed
		return result
	}

	if errors.Is(execCtx.Err(), context.DeadlineExceeded) {
		result.Status = Error
		if strings.TrimSpace(result.Output) == "" {
			result.Output = fmt.Sprintf("shell checker timed out after %s", c.timeout)
		} else {
			result.Output = strings.TrimSpace(result.Output) + fmt.Sprintf("\ncommand timed out after %s", c.timeout)
		}
		return result
	}

	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		result.Status = Failed
		return result
	}

	result.Status = Error
	if strings.TrimSpace(result.Output) == "" {
		result.Output = fmt.Sprintf("run shell command: %v", err)
		return result
	}
	result.Output = strings.TrimSpace(result.Output) + fmt.Sprintf("\nrun shell command: %v", err)
	return result
}

func shellProgram() string {
	if runtime.GOOS == "windows" {
		return "cmd"
	}

	return "sh"
}

func shellArgs(command string) []string {
	if runtime.GOOS == "windows" {
		return []string{"/c", command}
	}

	return []string{"-c", command}
}

func durationConfigValue(cfg map[string]interface{}, key string, fallback time.Duration) (time.Duration, error) {
	raw, ok := cfg[key]
	if !ok || raw == nil {
		return fallback, nil
	}

	switch value := raw.(type) {
	case time.Duration:
		if value <= 0 {
			return 0, fmt.Errorf("%q must be greater than zero", key)
		}
		return value, nil
	case string:
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			return fallback, nil
		}

		parsed, err := time.ParseDuration(trimmed)
		if err != nil {
			return 0, fmt.Errorf("parse %q duration %q: %w", key, value, err)
		}
		if parsed <= 0 {
			return 0, fmt.Errorf("%q must be greater than zero", key)
		}
		return parsed, nil
	default:
		return 0, fmt.Errorf("%q must be a duration string", key)
	}
}

func boolConfigValue(cfg map[string]interface{}, key string, fallback bool) (bool, error) {
	raw, ok := cfg[key]
	if !ok || raw == nil {
		return fallback, nil
	}

	value, ok := raw.(bool)
	if !ok {
		return false, fmt.Errorf("%q must be a boolean", key)
	}

	return value, nil
}

func envConfigValue(cfg map[string]interface{}, key string) (map[string]string, error) {
	raw, ok := cfg[key]
	if !ok || raw == nil {
		return nil, nil
	}

	switch typed := raw.(type) {
	case map[string]string:
		cloned := make(map[string]string, len(typed))
		for envKey, envValue := range typed {
			cloned[envKey] = envValue
		}
		return cloned, nil
	case map[string]interface{}:
		cloned := make(map[string]string, len(typed))
		for envKey, envValue := range typed {
			text, ok := envValue.(string)
			if !ok {
				return nil, fmt.Errorf("%q values must be strings", key)
			}
			cloned[envKey] = text
		}
		return cloned, nil
	default:
		return nil, fmt.Errorf("%q must be a map of strings", key)
	}
}

func mergedEnv(extra map[string]string) []string {
	env := append([]string{}, os.Environ()...)
	if len(extra) == 0 {
		return env
	}

	keys := make([]string, 0, len(extra))
	for key := range extra {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		env = append(env, fmt.Sprintf("%s=%s", key, extra[key]))
	}

	return env
}
