package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jelsin29/vigilanty/internal/tui/detect"
)

func TestNewUsesPresetDefaults(t *testing.T) {
	m := New("go")

	if m.preset != "go" {
		t.Fatalf("preset = %q, want %q", m.preset, "go")
	}
	if m.step != StepWelcome {
		t.Fatalf("step = %v, want %v", m.step, StepWelcome)
	}
	if len(m.includePatterns) != 1 || m.includePatterns[0] != "*.go" {
		t.Fatalf("includePatterns = %v, want [*.go]", m.includePatterns)
	}
	if len(m.excludePatterns) != 2 || m.excludePatterns[0] != "*_test.go" || m.excludePatterns[1] != "vendor/" {
		t.Fatalf("excludePatterns = %v, want [*_test.go vendor/]", m.excludePatterns)
	}
}

func TestResultFormatsProviderAndModel(t *testing.T) {
	m := New("go")
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

func TestResultReturnsNilWhenCancelled(t *testing.T) {
	m := New("go")
	m.cancelled = true

	if result := m.Result(); result != nil {
		t.Fatalf("Result() = %#v, want nil", result)
	}
	if !m.Cancelled() {
		t.Fatal("Cancelled() = false, want true")
	}
}

func TestProviderSelectionAdvancesToPatternsWhenModelNotNeeded(t *testing.T) {
	m := New("go")
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

func TestStepTransitionsFollowCompletionMessages(t *testing.T) {
	m := New("go")

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(Model)
	if cmd == nil {
		t.Fatal("welcome enter should return a step command")
	}

	updated, cmd = m.Update(cmd())
	m = updated.(Model)
	if m.step != StepSystemInfo {
		t.Fatalf("step = %v, want %v", m.step, StepSystemInfo)
	}

	m.sysDetected = true
	updated, cmd = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(Model)
	if cmd == nil {
		t.Fatal("system info enter should return a step command")
	}

	updated, cmd = m.Update(cmd())
	m = updated.(Model)
	if m.step != StepAIDetect {
		t.Fatalf("step = %v, want %v", m.step, StepAIDetect)
	}
}

func TestGoBackFromWelcomeCancels(t *testing.T) {
	m := New("go")
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(Model)
	if !m.cancelled {
		t.Fatal("esc from welcome should cancel the wizard")
	}
	if cmd == nil {
		t.Fatal("expected quit command")
	}
}
