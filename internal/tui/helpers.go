package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/jelsin29/vigilanty/internal/tui/detect"
)

func (m Model) selectedProviderInfo() (detect.ProviderInfo, bool) {
	if m.selectedProvider < 0 || m.selectedProvider >= len(m.providers) {
		return detect.ProviderInfo{}, false
	}

	return m.providers[m.selectedProvider], true
}

func (m Model) providerChoices() []int {
	found := make([]int, 0, len(m.providers))
	all := make([]int, 0, len(m.providers))
	for i, provider := range m.providers {
		all = append(all, i)
		if provider.Found {
			found = append(found, i)
		}
	}
	if len(found) > 0 {
		return found
	}
	return all
}

func (m *Model) syncProviderCursor() {
	choices := m.providerChoices()
	if len(choices) == 0 {
		m.providerCursor = 0
		return
	}
	for i, idx := range choices {
		if idx == m.selectedProvider {
			m.providerCursor = i
			return
		}
	}
	m.providerCursor = 0
}

func providerLabel(name string) string {
	if strings.TrimSpace(name) == "gh" {
		return "gh copilot"
	}
	return name
}

func providerStatus(provider detect.ProviderInfo) string {
	if provider.Found {
		if len(provider.Models) > 0 {
			return fmt.Sprintf("%s found (%d models)", CheckMark, len(provider.Models))
		}
		return CheckMark + " found"
	}
	if provider.Name == "lmstudio" {
		return CrossMark + " not running"
	}
	return CrossMark + " not found"
}

func joinPatterns(values []string) string {
	if len(values) == 0 {
		return "(none)"
	}
	return strings.Join(values, ", ")
}

func parseCommaList(raw string) []string {
	parts := strings.Split(raw, ",")
	items := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		items = append(items, part)
	}
	return items
}

func mergePatterns(base []string, extra []string) []string {
	merged := append([]string(nil), base...)
	seen := make(map[string]struct{}, len(base)+len(extra))
	for _, item := range merged {
		seen[item] = struct{}{}
	}
	for _, item := range extra {
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		merged = append(merged, item)
	}
	return merged
}

func centerLine(width int, text string) string {
	if width <= 0 {
		return text
	}
	textWidth := lipgloss.Width(text)
	if textWidth >= width {
		return text
	}
	left := (width - textWidth) / 2
	return strings.Repeat(" ", left) + text
}

func renderScreen(width int, title string, body []string, footer string) string {
	lines := []string{TitleStyle.Render(title), ""}
	lines = append(lines, body...)
	if footer != "" {
		lines = append(lines, "", KeyHintStyle.Render(footer))
	}
	return strings.Join(lines, "\n")
}
