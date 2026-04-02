package tui

import "fmt"

func (m Model) viewProviderSelect() string {
	body := []string{
		TitleStyle.Render("Select AI Providers"),
		"",
		HelpStyle.Render("Use j/k to move, space to toggle, enter to continue."),
		"",
	}
	for i, provider := range m.providers {
		check := "[ ]"
		style := UnselectedStyle
		if m.selectedProviders[i] {
			check = "[x]"
			style = SelectedStyle
		}
		line := fmt.Sprintf("%s %-12s %s", check, providerLabel(provider.Name), providerStatus(provider))
		if i == m.providerCursor {
			line = cursor + style.Render(line)
		} else {
			line = "  " + style.Render(line)
		}
		body = append(body, line)
	}
	body = append(body, "")
	for i, option := range []string{"Continue", "Back"} {
		idx := len(m.providers) + i
		line := option
		if idx == m.providerCursor {
			line = cursor + SelectedStyle.Render(option)
		} else {
			line = "  " + UnselectedStyle.Render(option)
		}
		body = append(body, line)
	}
	if m.providerError != "" {
		body = append(body, "", ErrorStyle.Render(m.providerError))
	}

	return renderScreen(m.width, "", body, "space: toggle • enter: confirm • esc: back")
}

func (m Model) viewModelSelect() string {
	provider, _, ok := m.currentModelProvider()
	if !ok {
		return renderScreen(m.width, "", []string{"No provider selected."}, "esc: back")
	}

	if !m.modelsDetected {
		body := []string{fmt.Sprintf("Select model for %s", providerLabel(provider.Name)), "", m.spinner.View() + " Discovering models..."}
		return renderScreen(m.width, "", body, "esc: back")
	}

	if len(m.modelOptions) == 0 {
		body := []string{
			fmt.Sprintf("Enter model for %s", providerLabel(provider.Name)),
			"",
			"  " + m.textInput.View(),
		}
		return renderScreen(m.width, "", body, "enter: confirm • esc: back")
	}

	body := []string{fmt.Sprintf("Select model for %s", providerLabel(provider.Name)), ""}
	for i, model := range m.modelOptions {
		prefix := "  "
		if i == m.modelCursor {
			prefix = cursor
			body = append(body, prefix+SelectedStyle.Render(model))
			continue
		}
		body = append(body, prefix+UnselectedStyle.Render(model))
	}

	return renderScreen(m.width, "", body, "j/k: navigate • enter: select • esc: back")
}

func (m Model) viewAgentsMd() string {
	body := []string{"Generate AGENTS.md with coding standards?", ""}
	options := []string{"Yes", "No"}
	for i, option := range options {
		prefix := "  "
		if i == m.rulesCursor {
			prefix = cursor
			body = append(body, prefix+SelectedStyle.Render(option))
			continue
		}
		body = append(body, prefix+UnselectedStyle.Render(option))
	}

	return renderScreen(m.width, "", body, "j/k: navigate • enter: select • esc: back")
}
