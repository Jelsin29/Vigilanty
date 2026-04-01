package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadValidYAML(t *testing.T) {
	path := writeConfigFile(t, `version: "1"
global:
  fail_fast: true
pipeline:
  - name: go-test
    checker: shell
    command: go test ./...
`)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}

	if len(cfg.Pipeline) != 1 {
		t.Fatalf("len(cfg.Pipeline) = %d, want 1", len(cfg.Pipeline))
	}

	if cfg.Pipeline[0].Type != "shell" {
		t.Fatalf("cfg.Pipeline[0].Type = %q, want shell", cfg.Pipeline[0].Type)
	}
}

func TestLoadInvalidYAML(t *testing.T) {
	path := writeConfigFile(t, "global: [\n")

	if _, err := Load(path); err == nil {
		t.Fatal("Load() error = nil, want error")
	}
}

func TestValidateMissingCheckerType(t *testing.T) {
	cfg := &Config{
		Pipeline: []StepConfig{{Name: "missing-type"}},
	}
	applyDefaults(cfg)

	if err := cfg.Validate(); err == nil {
		t.Fatal("Validate() error = nil, want error")
	}
}

func TestValidateValidConfig(t *testing.T) {
	cfg := &Config{
		Version: "1",
		Global:  GlobalConfig{Timeout: "30s"},
		Pipeline: []StepConfig{{
			Name:    "go-test",
			Type:    "shell",
			Command: "go test ./...",
		}},
	}
	applyDefaults(cfg)

	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate() error = %v, want nil", err)
	}
}

func TestApplyDefaults(t *testing.T) {
	cfg := &Config{
		Pipeline: []StepConfig{{
			Name:    "go-test",
			Checker: "shell",
			Command: "go test ./...",
		}},
	}

	applyDefaults(cfg)

	if cfg.Global.DiffMaxBytes != 256*1024 {
		t.Fatalf("cfg.Global.DiffMaxBytes = %d, want %d", cfg.Global.DiffMaxBytes, 256*1024)
	}

	if cfg.Global.Timeout != "2m" {
		t.Fatalf("cfg.Global.Timeout = %q, want %q", cfg.Global.Timeout, "2m")
	}

	if cfg.Pipeline[0].Timeout != "2m" {
		t.Fatalf("cfg.Pipeline[0].Timeout = %q, want %q", cfg.Pipeline[0].Timeout, "2m")
	}

	if cfg.Pipeline[0].Type != "shell" {
		t.Fatalf("cfg.Pipeline[0].Type = %q, want %q", cfg.Pipeline[0].Type, "shell")
	}

	if cfg.Pipeline[0].Config == nil {
		t.Fatal("cfg.Pipeline[0].Config = nil, want initialized map")
	}
	if len(cfg.Pipeline[0].Config) != 0 {
		t.Fatalf("len(cfg.Pipeline[0].Config) = %d, want 0", len(cfg.Pipeline[0].Config))
	}
}

func TestLoadRejectsUnknownTopLevelKeys(t *testing.T) {
	path := writeConfigFile(t, `version: "1"
settings: {}
pipeline:
  - name: go-test
    checker: shell
    command: go test ./...
`)

	if _, err := Load(path); err == nil {
		t.Fatal("Load() error = nil, want error")
	}
}

func TestValidateRejectsUnsupportedVersion(t *testing.T) {
	cfg := &Config{
		Version: "2",
		Pipeline: []StepConfig{{
			Name:    "go-test",
			Type:    "shell",
			Command: "go test ./...",
		}},
	}
	applyDefaults(cfg)

	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate() error = nil, want error")
	}
	if got, want := err.Error(), "unsupported config version: 2"; got != want {
		t.Fatalf("Validate() error = %q, want %q", got, want)
	}
}

func writeConfigFile(t *testing.T, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), ".vigilanty.yml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}

	return path
}
