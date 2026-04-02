package tui

import "fmt"

func (m Model) viewProviderSelect() string {
	choices := m.providerChoices()
	body := []string{"? Select your AI provider:", ""}
	for i, idx := range choices {
		cursor := "  "
		if i == m.providerCursor {
			cursor = "▸ "
		}
		body = append(body, cursor+providerLabel(m.providers[idx].Name))
	}

	return renderScreen(m.width, "", body, "j/k: navigate • enter: select • esc: back")
}

func (m Model) viewModelSelect() string {
	provider, ok := m.selectedProviderInfo()
	if !ok {
		return renderScreen(m.width, "", []string{"No provider selected."}, "esc: back")
	}

	if !m.modelsDetected {
		body := []string{fmt.Sprintf("? Select model for %s:", providerLabel(provider.Name)), "", m.spinner.View() + " Discovering models..."}
		return renderScreen(m.width, "", body, "esc: back")
	}

	if len(m.modelOptions) == 0 {
		body := []string{
			fmt.Sprintf("? Enter model for %s:", providerLabel(provider.Name)),
			"",
			"  " + m.textInput.View(),
		}
		return renderScreen(m.width, "", body, "enter: confirm • esc: back")
	}

	body := []string{fmt.Sprintf("? Select model for %s:", providerLabel(provider.Name)), ""}
	for i, model := range m.modelOptions {
		cursor := "  "
		if i == m.modelCursor {
			cursor = "▸ "
		}
		body = append(body, cursor+model)
	}

	return renderScreen(m.width, "", body, "j/k: navigate • enter: select • esc: back")
}

func (m Model) viewAgentsMd() string {
	body := []string{"? Generate AGENTS.md with coding standards?", ""}
	options := []string{"Yes", "No"}
	for i, option := range options {
		cursor := "  "
		if i == m.rulesCursor {
			cursor = "▸ "
		}
		body = append(body, cursor+option)
	}

	return renderScreen(m.width, "", body, "j/k: navigate • enter: select • esc: back")
}
