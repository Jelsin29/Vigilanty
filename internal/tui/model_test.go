package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jelsin29/vigilanty/internal/tui/detect"
)

func TestResultFormatsProviderAndModel(t *testing.T) {
	m := NewModel("go")
	m.providers = []detect.ProviderInfo{{Name: "ollama", Found: true, NeedsModel: true}}
	m.selectedProvider = 0
	m.selectedModel = "llama3:latest"
	m.includePatterns = []string{"*.go"}
	m.excludePatterns = []string{"*_test.go", "vendor/"}
	m.generateRules = true

	result := m.Result()
	if result == nil {
		t.Fatal("Result() returned nil")
	}
	if result.Provider != "ollama:llama3:latest" {
		t.Fatalf("Result().Provider = %q, want %q", result.Provider, "ollama:llama3:latest")
	}
}

func TestProviderSelectionAdvancesToPatternsWhenModelNotNeeded(t *testing.T) {
	m := NewModel("go")
	m.step = StepProviderSelect
	m.providers = []detect.ProviderInfo{{Name: "claude", Found: true}}

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(Model)
	if cmd == nil {
		t.Fatal("expected enter to return a step command")
	}
	msg := cmd()
	stepMsg, ok := msg.(StepCompleteMsg)
	if !ok {
		t.Fatalf("command returned %T, want StepCompleteMsg", msg)
	}
	if stepMsg.Next != StepPatterns {
		t.Fatalf("next step = %v, want %v", stepMsg.Next, StepPatterns)
	}
	if m.selectedProvider != 0 {
		t.Fatalf("selectedProvider = %d, want 0", m.selectedProvider)
	}
}

func TestGoBackFromWelcomeCancels(t *testing.T) {
	m := NewModel("go")
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(Model)
	if !m.cancelled {
		t.Fatal("esc from welcome should cancel the wizard")
	}
	if cmd == nil {
		t.Fatal("expected quit command")
	}
}
