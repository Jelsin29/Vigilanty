package tui

import (
	"fmt"
	"strings"
)

func (m Model) viewConfirm() string {
	provider := "claude"
	model := "(none)"
	if selected, ok := m.selectedProviderInfo(); ok {
		provider = providerLabel(selected.Name)
		if strings.TrimSpace(m.selectedModel) != "" {
			model = m.selectedModel
		}
	}

	summary := BorderedBox(strings.Join([]string{
		centerLine(34, "Configuration Summary"),
		fmt.Sprintf("  %s %s", padRight(LabelStyle.Render("Project"), 10), ValueStyle.Render(m.preset)),
		fmt.Sprintf("  %s %s", padRight(LabelStyle.Render("Provider"), 10), ValueStyle.Render(provider)),
		fmt.Sprintf("  %s %s", padRight(LabelStyle.Render("Model"), 10), ValueStyle.Render(model)),
		fmt.Sprintf("  %s %s", padRight(LabelStyle.Render("Include"), 10), ValueStyle.Render(joinPatterns(m.includePatterns))),
		fmt.Sprintf("  %s %s", padRight(LabelStyle.Render("Exclude"), 10), ValueStyle.Render(joinPatterns(m.excludePatterns))),
		fmt.Sprintf("  %s %s", padRight(LabelStyle.Render("AGENTS.md"), 10), ValueStyle.Render(map[bool]string{true: "yes", false: "no"}[m.generateRules])),
	}, "\n"))

	return renderScreen(m.width, "", []string{summary}, "enter: confirm and save • esc: go back")
}
