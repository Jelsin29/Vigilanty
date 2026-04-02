package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) updateSpinner(msg tea.Msg) tea.Cmd {
	spin, cmd := m.spinner.Update(msg)
	m.spinner = spin
	return cmd
}

func (m *Model) handleGlobalKeys(msg tea.KeyMsg) (tea.Cmd, bool) {
	switch msg.String() {
	case "ctrl+c":
		m.cancelled = true
		return tea.Quit, true
	case "q":
		m.cancelled = true
		return tea.Quit, true
	case "esc":
		return m.goBack(), true
	}

	return nil, false
}

func (m *Model) updateCurrentStep(msg tea.Msg) tea.Cmd {
	switch m.step {
	case StepWelcome:
		return m.updateWelcome(msg)
	case StepSystemInfo:
		return m.updateSystemInfo(msg)
	case StepAIDetect:
		return m.updateAIDetect(msg)
	case StepProviderSelect:
		return m.updateProviderSelect(msg)
	case StepModelSelect:
		return m.updateModelSelect(msg)
	case StepPatterns:
		return m.updatePatterns(msg)
	case StepAgentsMd:
		return m.updateAgentsMd(msg)
	case StepConfirm:
		return m.updateConfirm(msg)
	default:
		return nil
	}
}

func (m *Model) updateWelcome(msg tea.Msg) tea.Cmd {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return nil
	}

	switch key.String() {
	case "j", "down":
		if m.welcomeCursor < 1 {
			m.welcomeCursor++
		}
	case "k", "up":
		if m.welcomeCursor > 0 {
			m.welcomeCursor--
		}
	case "enter":
		if m.welcomeCursor == 1 {
			m.cancelled = true
			return tea.Quit
		}
		return nextStepCmd(StepSystemInfo)
	}

	return nil
}

func (m *Model) updateSystemInfo(msg tea.Msg) tea.Cmd {
	key, ok := msg.(tea.KeyMsg)
	if !ok || key.Type != tea.KeyEnter || !m.sysDetected {
		return nil
	}

	return nextStepCmd(StepAIDetect)
}

func (m *Model) updateAIDetect(msg tea.Msg) tea.Cmd {
	key, ok := msg.(tea.KeyMsg)
	if !ok || key.Type != tea.KeyEnter || !m.providersDetected {
		return nil
	}

	return nextStepCmd(StepProviderSelect)
}

func (m *Model) updateProviderSelect(msg tea.Msg) tea.Cmd {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return nil
	}

	switch key.String() {
	case "j", "down":
		if m.providerCursor < m.providerMenuSize()-1 {
			m.providerCursor++
		}
	case "k", "up":
		if m.providerCursor > 0 {
			m.providerCursor--
		}
	case "space", " ":
		if m.providerMenuTarget() != "provider" {
			return nil
		}
		m.toggleProvider(m.providerCursor)
	case "enter":
		switch m.providerMenuTarget() {
		case "provider":
			return nil
		case "back":
			return nextStepCmd(StepAIDetect)
		}
		if len(m.selectedProviderIndices()) == 0 {
			m.providerError = "Select at least one provider to continue."
			return nil
		}
		m.providerError = ""
		return nextStepCmd(m.beginModelFlow())
	}

	return nil
}

func (m *Model) updateModelSelect(msg tea.Msg) tea.Cmd {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return m.updateManualModelInput(msg)
	}

	if len(m.modelOptions) == 0 {
		if key.Type == tea.KeyEnter {
			next := m.storeCurrentModel(m.textInput.Value())
			if next == StepModelSelect {
				return nil
			}
			return nextStepCmd(next)
		}
		return m.updateManualModelInput(msg)
	}

	switch key.String() {
	case "j", "down":
		if m.modelCursor < len(m.modelOptions)-1 {
			m.modelCursor++
		}
	case "k", "up":
		if m.modelCursor > 0 {
			m.modelCursor--
		}
	case "enter":
		return nextStepCmd(m.storeCurrentModel(m.modelOptions[m.modelCursor]))
	}

	return nil
}

func (m *Model) updatePatterns(msg tea.Msg) tea.Cmd {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return m.updatePatternInputs(msg)
	}

	switch key.String() {
	case "j", "down":
		m.patternField = 1
		return m.focusPatternField()
	case "k", "up":
		m.patternField = 0
		return m.focusPatternField()
	case "enter":
		if m.patternField == 0 {
			m.patternField = 1
			return m.focusPatternField()
		}
		m.applyPatternInputs()
		return nextStepCmd(StepAgentsMd)
	default:
		return m.updatePatternInputs(msg)
	}
}

func (m *Model) updateAgentsMd(msg tea.Msg) tea.Cmd {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return nil
	}

	switch key.String() {
	case "j", "down":
		if m.rulesCursor < 1 {
			m.rulesCursor++
		}
	case "k", "up":
		if m.rulesCursor > 0 {
			m.rulesCursor--
		}
	case "space", " ":
		if m.rulesCursor == 0 {
			m.rulesCursor = 1
		} else {
			m.rulesCursor = 0
		}
	case "enter":
		m.generateRules = m.rulesCursor == 0
		return nextStepCmd(StepConfirm)
	}

	m.generateRules = m.rulesCursor == 0
	return nil
}

func (m *Model) updateConfirm(msg tea.Msg) tea.Cmd {
	key, ok := msg.(tea.KeyMsg)
	if !ok || key.Type != tea.KeyEnter {
		return nil
	}

	m.done = true
	return tea.Quit
}

func (m *Model) updateManualModelInput(msg tea.Msg) tea.Cmd {
	input, cmd := m.textInput.Update(msg)
	m.textInput = input
	return cmd
}

func (m *Model) updatePatternInputs(msg tea.Msg) tea.Cmd {
	if m.patternField == 0 {
		input, cmd := m.includeInput.Update(msg)
		m.includeInput = input
		return cmd
	}

	input, cmd := m.excludeInput.Update(msg)
	m.excludeInput = input
	return cmd
}

func (m *Model) goBack() tea.Cmd {
	switch m.step {
	case StepWelcome:
		m.cancelled = true
		return tea.Quit
	case StepSystemInfo:
		return nextStepCmd(StepWelcome)
	case StepAIDetect:
		return nextStepCmd(StepSystemInfo)
	case StepProviderSelect:
		return nextStepCmd(StepAIDetect)
	case StepModelSelect:
		return nextStepCmd(StepProviderSelect)
	case StepPatterns:
		if m.activeModelProvider >= 0 {
			return nextStepCmd(StepModelSelect)
		}
		return nextStepCmd(StepProviderSelect)
	case StepAgentsMd:
		return nextStepCmd(StepPatterns)
	case StepConfirm:
		return nextStepCmd(StepAgentsMd)
	default:
		return nil
	}
}

func (m *Model) enterStep(step Step) []tea.Cmd {
	switch step {
	case StepAIDetect:
		if m.providersDetected {
			return nil
		}
		return []tea.Cmd{detectProvidersCmd()}
	case StepModelSelect:
		p, idx, ok := m.currentModelProvider()
		if !ok || !p.NeedsModel {
			return []tea.Cmd{nextStepCmd(StepPatterns)}
		}
		if existing := strings.TrimSpace(m.selectedProviderModel[idx]); existing != "" {
			m.selectedModel = existing
			m.modelOptions = append([]string(nil), p.Models...)
			m.modelsDetected = len(m.modelOptions) > 0
			return nil
		}
		m.modelOptions = append([]string(nil), p.Models...)
		m.modelsDetected = len(m.modelOptions) > 0
		m.textInput.SetValue("")
		if len(m.modelOptions) > 0 {
			return nil
		}
		return []tea.Cmd{discoverModelsCmd(p.Name)}
	case StepPatterns:
		return []tea.Cmd{m.focusPatternField()}
	case StepAgentsMd:
		return nil
	}

	return nil
}

func (m *Model) focusPatternField() tea.Cmd {
	m.includeInput.Blur()
	m.excludeInput.Blur()
	if m.patternField == 0 {
		return m.includeInput.Focus()
	}
	return m.excludeInput.Focus()
}

func (m *Model) focusManualModel() tea.Cmd {
	m.textInput.SetCursor(len(m.textInput.Value()))
	return m.textInput.Focus()
}

func (m *Model) applyPatternInputs() {
	m.includePatterns = mergePatterns(m.includePatterns, parseCommaList(m.includeInput.Value()))
	m.excludePatterns = mergePatterns(m.excludePatterns, parseCommaList(m.excludeInput.Value()))
	m.includeInput.SetValue("")
	m.excludeInput.SetValue("")
	if m.rulesCursor > 1 {
		m.rulesCursor = 1
	}
	m.generateRules = m.rulesCursor == 0
}

func nextStepCmd(next Step) tea.Cmd {
	return func() tea.Msg { return StepCompleteMsg{Next: next} }
}
