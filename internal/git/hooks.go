package git

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	vigilantyHookStart = "# ---- vigilanty start ----"
	vigilantyHookEnd   = "# ---- vigilanty end ----"
)

func InstallHook(hookType string) error {
	hookType = strings.TrimSpace(hookType)
	if hookType == "" {
		return errors.New("install hook: hook type is required")
	}

	hooksDir, err := HooksDir()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(hooksDir, 0o755); err != nil {
		return fmt.Errorf("create hooks directory %q: %w", hooksDir, err)
	}

	hookPath := filepath.Join(hooksDir, hookType)
	existing, err := os.ReadFile(hookPath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("read hook %q: %w", hookPath, err)
	}

	updated := upsertHookSection(string(existing), renderHookBlock())
	if err := os.WriteFile(hookPath, []byte(updated), 0o755); err != nil {
		return fmt.Errorf("write hook %q: %w", hookPath, err)
	}

	return nil
}

func UninstallHook() error {
	hooksDir, err := HooksDir()
	if err != nil {
		return err
	}

	hookPath := filepath.Join(hooksDir, "pre-commit")
	content, err := os.ReadFile(hookPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("read hook %q: %w", hookPath, err)
	}

	updated, changed := removeHookSection(string(content))
	if !changed {
		return nil
	}

	if err := os.WriteFile(hookPath, []byte(updated), 0o755); err != nil {
		return fmt.Errorf("write hook %q: %w", hookPath, err)
	}

	return nil
}

func IsHookInstalled() bool {
	hooksDir, err := HooksDir()
	if err != nil {
		return false
	}

	content, err := os.ReadFile(filepath.Join(hooksDir, "pre-commit"))
	if err != nil {
		return false
	}

	return strings.Contains(string(content), vigilantyHookStart) && strings.Contains(string(content), vigilantyHookEnd)
}

func renderHookBlock() string {
	return vigilantyHookStart + "\n" +
		"if command -v vigilanty >/dev/null 2>&1; then\n" +
		"  $(pwd)/../vigilanty run\n" +
		"else\n" +
		"  printf '%s\\n' 'vigilanty: command not found' >&2\n" +
		"  exit 1\n" +
		"fi\n" +
		vigilantyHookEnd + "\n"
}

func upsertHookSection(existing string, block string) string {
	if strings.TrimSpace(existing) == "" {
		return "#!/bin/sh\n\n" + block
	}

	updated, changed := replaceHookSection(existing, block)
	if changed {
		return ensureTrailingNewline(updated)
	}

	trimmed := ensureTrailingNewline(existing)
	if !strings.HasSuffix(trimmed, "\n\n") {
		trimmed += "\n"
	}

	return trimmed + block
}

func replaceHookSection(existing string, block string) (string, bool) {
	start := strings.Index(existing, vigilantyHookStart)
	end := strings.Index(existing, vigilantyHookEnd)
	if start == -1 || end == -1 || end < start {
		return existing, false
	}

	end += len(vigilantyHookEnd)
	if end < len(existing) && existing[end] == '\n' {
		end++
	}

	return existing[:start] + block + existing[end:], true
}

func removeHookSection(existing string) (string, bool) {
	updated, changed := replaceHookSection(existing, "")
	if !changed {
		return existing, false
	}

	cleaned := strings.TrimRight(updated, "\n")
	if cleaned == "#!/bin/sh" {
		return cleaned + "\n", true
	}

	if cleaned == "" {
		return "", true
	}

	return cleaned + "\n", true
}

func ensureTrailingNewline(value string) string {
	if strings.HasSuffix(value, "\n") {
		return value
	}
	return value + "\n"
}
