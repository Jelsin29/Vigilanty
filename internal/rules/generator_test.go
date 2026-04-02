package rules

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDiscoverExplicitRulesFile(t *testing.T) {
	dir := t.TempDir()
	rulesPath := filepath.Join(dir, "my-rules.md")
	if err := os.WriteFile(rulesPath, []byte("# Custom Rules\nno bugs"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	content, ok := Discover(dir, "my-rules.md")
	if !ok {
		t.Fatal("Discover() ok = false, want true")
	}
	if !strings.Contains(content, "no bugs") {
		t.Fatalf("Discover() content = %q, want to contain 'no bugs'", content)
	}
}

func TestDiscoverFallsBackToAgentsMD(t *testing.T) {
	dir := t.TempDir()
	agentsPath := filepath.Join(dir, "AGENTS.md")
	if err := os.WriteFile(agentsPath, []byte("# AGENTS\nreview carefully"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	content, ok := Discover(dir, "")
	if !ok {
		t.Fatal("Discover() ok = false, want true")
	}
	if !strings.Contains(content, "review carefully") {
		t.Fatalf("Discover() content = %q, want to contain 'review carefully'", content)
	}
}

func TestDiscoverReturnsEmptyWhenNoFile(t *testing.T) {
	dir := t.TempDir()

	content, ok := Discover(dir, "")
	if ok {
		t.Fatalf("Discover() ok = true, want false (no files exist)")
	}
	if content != "" {
		t.Fatalf("Discover() content = %q, want empty", content)
	}
}

func TestDiscoverExplicitFileTakesPriority(t *testing.T) {
	dir := t.TempDir()

	// both files exist — explicit should win
	os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte("agents content"), 0o644)
	os.WriteFile(filepath.Join(dir, "custom.md"), []byte("custom content"), 0o644)

	content, ok := Discover(dir, "custom.md")
	if !ok {
		t.Fatal("Discover() ok = false, want true")
	}
	if content != "custom content" {
		t.Fatalf("Discover() = %q, want 'custom content'", content)
	}
}

func TestDiscoverAbsoluteRulesFilePath(t *testing.T) {
	dir := t.TempDir()
	absPath := filepath.Join(dir, "abs-rules.md")
	os.WriteFile(absPath, []byte("absolute rules"), 0o644)

	content, ok := Discover("/some/other/root", absPath)
	if !ok {
		t.Fatal("Discover() ok = false, want true for absolute path")
	}
	if content != "absolute rules" {
		t.Fatalf("Discover() = %q, want 'absolute rules'", content)
	}
}

func TestGenerateContainsPresetRules(t *testing.T) {
	content := Generate(GenerateOptions{
		ProjectType:  "go",
		FilePatterns: []string{"*.go"},
	})

	if !strings.Contains(content, "Go Rules") {
		t.Fatal("Generate(go) should contain 'Go Rules'")
	}
	if !strings.Contains(content, "REJECT if") {
		t.Fatal("Generate() missing REJECT section")
	}
	if !strings.Contains(content, "STATUS: PASSED") {
		t.Fatal("Generate() missing response format")
	}
}

func TestGenerateIncludesDetectedTools(t *testing.T) {
	content := Generate(GenerateOptions{
		ProjectType:   "python",
		DetectedTools: []string{"ruff", "mypy"},
	})

	if !strings.Contains(content, "ruff") {
		t.Fatal("Generate() should include detected tool 'ruff'")
	}
	if !strings.Contains(content, "Detected Tools") {
		t.Fatal("Generate() should have Detected Tools section")
	}
}

func TestGenerateGenericPreset(t *testing.T) {
	content := Generate(GenerateOptions{ProjectType: "unknown-lang"})

	if !strings.Contains(content, "Generic Rules") {
		t.Fatalf("Generate(unknown) should fall back to Generic, got: %s", content[:80])
	}
}

func TestDetectToolsFindsConfigs(t *testing.T) {
	dir := t.TempDir()

	// create some config files
	os.WriteFile(filepath.Join(dir, ".golangci.yml"), []byte(""), 0o644)
	os.WriteFile(filepath.Join(dir, ".prettierrc"), []byte("{}"), 0o644)

	tools := DetectTools(dir)

	found := make(map[string]bool)
	for _, tool := range tools {
		found[tool] = true
	}

	if !found["golangci-lint"] {
		t.Error("DetectTools() should find golangci-lint")
	}
	if !found["prettier"] {
		t.Error("DetectTools() should find prettier")
	}
}

func TestDetectToolsEmptyDir(t *testing.T) {
	dir := t.TempDir()
	tools := DetectTools(dir)
	if len(tools) != 0 {
		t.Fatalf("DetectTools() = %v, want empty slice", tools)
	}
}

func TestDetectToolsRuffInPyproject(t *testing.T) {
	dir := t.TempDir()
	pyproject := `[tool.ruff]
line-length = 88
`
	os.WriteFile(filepath.Join(dir, "pyproject.toml"), []byte(pyproject), 0o644)

	tools := DetectTools(dir)
	found := false
	for _, t := range tools {
		if t == "ruff" {
			found = true
		}
	}
	if !found {
		t.Fatal("DetectTools() should detect ruff from pyproject.toml [tool.ruff] section")
	}
}
