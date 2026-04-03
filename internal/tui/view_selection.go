package tui

import (
	"fmt"
)

const scrollWindowSize = 10

// scrollWindow computes the visible window [start, end) for a list of total
// items, keeping cursorIdx visible and the window at most scrollWindowSize.
func scrollWindow(total, cursorIdx int) (start, end int) {
	if total <= scrollWindowSize {
		return 0, total
	}
	// keep cursor roughly centered
	start = cursorIdx - scrollWindowSize/2
	if start < 0 {
		start = 0
	}
	end = start + scrollWindowSize
	if end > total {
		end = total
		start = end - scrollWindowSize
	}
	return start, end
}

func renderScrollList(items []string, cursorIdx int, cursorStr string, selectedStyle, unselectedStyle func(...string) string) []string {
	total := len(items)
	start, end := scrollWindow(total, cursorIdx)

	lines := make([]string, 0, end-start+2)
	if start > 0 {
		lines = append(lines, MutedStyle.Render("  ↑ more"))
	}
	for i := start; i < end; i++ {
		if i == cursorIdx {
			lines = append(lines, cursorStr+selectedStyle(items[i]))
		} else {
			lines = append(lines, "  "+unselectedStyle(items[i]))
		}
	}
	if end < total {
		lines = append(lines, MutedStyle.Render("  ↓ more"))
	}
	return lines
}

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
		var line string
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

	title := fmt.Sprintf("Select model for %s", providerLabel(provider.Name))
	if subProvider, ok := m.activeSubProviderInfo(); ok {
		title = fmt.Sprintf("Select model for %s (%s)", providerLabel(provider.Name), subProvider.Name)
	}

	if !m.modelsDetected {
		body := []string{title, "", m.spinner.View() + " Discovering models..."}
		return renderScreen(m.width, "", body, "esc: back")
	}

	if len(m.modelOptions) == 0 {
		body := []string{
			title,
			"",
			"  " + m.textInput.View(),
		}
		return renderScreen(m.width, "", body, "enter: confirm • esc: back")
	}

	body := []string{title, ""}
	body = append(body, renderScrollList(
		m.modelOptions, m.modelCursor, cursor,
		SelectedStyle.Render, UnselectedStyle.Render,
	)...)

	return renderScreen(m.width, "", body, "j/k: navigate • enter: select • esc: back")
}

func (m Model) viewSubProviderSelect() string {
	provider, _, ok := m.currentModelProvider()
	if !ok {
		return renderScreen(m.width, "", []string{"No provider selected."}, "esc: back")
	}

	body := []string{fmt.Sprintf("Select provider (%s)", providerLabel(provider.Name)), ""}
	subProviderLabels := make([]string, len(m.subProviders))
	for i, sp := range m.subProviders {
		subProviderLabels[i] = fmt.Sprintf("%s (%d models)", sp.Name, sp.ModelCount)
	}
	body = append(body, renderScrollList(
		subProviderLabels, m.subProviderCursor, cursor,
		SelectedStyle.Render, UnselectedStyle.Render,
	)...)

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
