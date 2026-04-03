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
	m.providers = []detect.ProviderInfo{{Name: "opencode", Found: true, NeedsModel: true}}
	m.selectedProvider = 0
	m.selectedModel = "anthropic/claude-sonnet-4"
	m.includePatterns = []string{"*.go"}
	m.excludePatterns = []string{"*_test.go", "vendor/"}
	m.generateRules = true

	result := m.Result()
	if result == nil {
		t.Fatal("Result() returned nil")
	}
	if result.Provider != "opencode:anthropic/claude-sonnet-4" {
		t.Fatalf("Result().Provider = %q, want %q", result.Provider, "opencode:anthropic/claude-sonnet-4")
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
	m.selectedProviders[0] = true
	m.providerCursor = len(m.providers)

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

func TestWelcomeEnterOnQuitCancels(t *testing.T) {
	m := New("go")
	m.welcomeCursor = 1

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(Model)
	if !m.cancelled {
		t.Fatal("enter on quit should cancel the wizard")
	}
	if cmd == nil {
		t.Fatal("expected quit command")
	}
}

func TestProviderSelectSpaceTogglesAndContinueOpensModelStep(t *testing.T) {
	m := New("go")
	m.step = StepProviderSelect
	m.providers = []detect.ProviderInfo{{Name: "claude", Found: true}, {Name: "gh", Found: true, NeedsModel: true, Models: []string{"gpt-4o"}}}
	m.providerSelectionsSet = true

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	m = updated.(Model)
	if cmd != nil {
		t.Fatal("space should not advance steps")
	}
	if !m.selectedProviders[0] {
		t.Fatal("space should select the focused provider")
	}

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = updated.(Model)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	m = updated.(Model)
	if !m.selectedProviders[1] {
		t.Fatal("space should select the second provider")
	}

	m.providerCursor = len(m.providers)
	updated, cmd = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(Model)
	if cmd == nil {
		t.Fatal("continue should return a step command")
	}
	stepMsg, ok := cmd().(StepCompleteMsg)
	if !ok {
		t.Fatalf("command returned %T, want StepCompleteMsg", cmd())
	}
	if stepMsg.Next != StepModelSelect {
		t.Fatalf("next step = %v, want %v", stepMsg.Next, StepModelSelect)
	}
	if m.selectedProvider != 0 {
		t.Fatalf("selectedProvider = %d, want 0", m.selectedProvider)
	}
	if m.activeModelProvider != 1 {
		t.Fatalf("activeModelProvider = %d, want 1", m.activeModelProvider)
	}
}

func TestProviderSelectContinueOpensSubProviderStepWhenAvailable(t *testing.T) {
	m := New("go")
	m.step = StepProviderSelect
	m.providers = []detect.ProviderInfo{
		{Name: "claude", Found: true},
		{Name: "opencode", Found: true, NeedsModel: true, SubProviders: []detect.SubProvider{{Name: "Anthropic", ID: "anthropic", Models: []string{"claude-sonnet-4"}, ModelCount: 1}}},
	}
	m.providerSelectionsSet = true
	m.selectedProviders[1] = true
	m.providerCursor = len(m.providers)

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(Model)
	if cmd == nil {
		t.Fatal("continue should return a step command")
	}
	stepMsg, ok := cmd().(StepCompleteMsg)
	if !ok {
		t.Fatalf("command returned %T, want StepCompleteMsg", cmd())
	}
	if stepMsg.Next != StepSubProviderSelect {
		t.Fatalf("next step = %v, want %v", stepMsg.Next, StepSubProviderSelect)
	}
	if m.activeModelProvider != 1 {
		t.Fatalf("activeModelProvider = %d, want 1", m.activeModelProvider)
	}
	if len(m.subProviders) != 1 || m.subProviders[0].ID != "anthropic" {
		t.Fatalf("subProviders = %#v, want anthropic entry", m.subProviders)
	}
}

func TestSubProviderSelectEnterOpensModelStep(t *testing.T) {
	m := New("go")
	m.step = StepSubProviderSelect
	m.activeModelProvider = 0
	m.subProviders = []detect.SubProvider{{Name: "Anthropic", ID: "anthropic", Models: []string{"claude-sonnet-4", "claude-opus-4"}, ModelCount: 2}}

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(Model)
	if cmd == nil {
		t.Fatal("enter should return a step command")
	}
	stepMsg, ok := cmd().(StepCompleteMsg)
	if !ok {
		t.Fatalf("command returned %T, want StepCompleteMsg", cmd())
	}
	if stepMsg.Next != StepModelSelect {
		t.Fatalf("next step = %v, want %v", stepMsg.Next, StepModelSelect)
	}
	if m.activeSubProvider != "anthropic" {
		t.Fatalf("activeSubProvider = %q, want %q", m.activeSubProvider, "anthropic")
	}
	if len(m.modelOptions) != 2 || m.modelOptions[0] != "claude-sonnet-4" {
		t.Fatalf("modelOptions = %v, want anthropic models", m.modelOptions)
	}
}

func TestStoreCurrentModelFormatsSubProviderSelection(t *testing.T) {
	m := New("go")
	m.providers = []detect.ProviderInfo{{Name: "opencode", Found: true, NeedsModel: true, SubProviders: []detect.SubProvider{{Name: "Anthropic", ID: "anthropic", Models: []string{"claude-sonnet-4"}, ModelCount: 1}}}}
	m.selectedProviders[0] = true
	m.activeModelProvider = 0
	m.activeSubProvider = "anthropic"

	next := m.storeCurrentModel("claude-sonnet-4")
	if next != StepPatterns {
		t.Fatalf("storeCurrentModel() next = %v, want %v", next, StepPatterns)
	}
	if got := m.selectedProviderModel[0]; got != "anthropic/claude-sonnet-4" {
		t.Fatalf("selectedProviderModel[0] = %q, want %q", got, "anthropic/claude-sonnet-4")
	}
}

func TestProviderSelectEnterTogglesFocusedProvider(t *testing.T) {
	m := New("go")
	m.step = StepProviderSelect
	m.providers = []detect.ProviderInfo{{Name: "claude", Found: true}}
	m.providerSelectionsSet = true

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(Model)
	if cmd != nil {
		t.Fatal("enter on provider should not advance steps")
	}
	if !m.selectedProviders[0] {
		t.Fatal("enter should toggle the focused provider")
	}
}

func TestProviderSelectNavigationSkipsUnavailableProviders(t *testing.T) {
	m := New("go")
	m.step = StepProviderSelect
	m.providers = []detect.ProviderInfo{
		{Name: "claude", Found: true},
		{Name: "gemini", Found: false},
		{Name: "ollama", Found: true},
	}
	m.providerSelectionsSet = true
	m.providerCursor = 0

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = updated.(Model)
	if m.providerCursor != 2 {
		t.Fatalf("providerCursor after down = %d, want 2", m.providerCursor)
	}

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m = updated.(Model)
	if m.providerCursor != 0 {
		t.Fatalf("providerCursor after up = %d, want 0", m.providerCursor)
	}
}

func TestInitializeProviderSelectionsStartsOnFirstFoundProvider(t *testing.T) {
	m := New("go")
	m.providers = []detect.ProviderInfo{
		{Name: "claude", Found: false},
		{Name: "gemini", Found: true},
		{Name: "ollama", Found: true},
	}

	m.initializeProviderSelections()

	if m.providerCursor != 1 {
		t.Fatalf("providerCursor = %d, want 1", m.providerCursor)
	}
	if !m.selectedProviders[1] || !m.selectedProviders[2] {
		t.Fatalf("selectedProviders = %v, want providers 1 and 2 selected", m.selectedProviders)
	}
}

func TestStepTransitionsFollowCompletionMessages(t *testing.T) {
	m := New("go")

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(Model)
	if cmd == nil {
		t.Fatal("welcome enter should return a step command")
	}

	updated, _ = m.Update(cmd())
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

	updated, _ = m.Update(cmd())
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

func TestGoBackFromModelSelectReturnsToSubProviderStep(t *testing.T) {
	m := New("go")
	m.step = StepModelSelect
	m.activeSubProvider = "anthropic"

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(Model)
	if cmd == nil {
		t.Fatal("esc should return a step command")
	}
	stepMsg, ok := cmd().(StepCompleteMsg)
	if !ok {
		t.Fatalf("command returned %T, want StepCompleteMsg", cmd())
	}
	if stepMsg.Next != StepSubProviderSelect {
		t.Fatalf("next step = %v, want %v", stepMsg.Next, StepSubProviderSelect)
	}
}
