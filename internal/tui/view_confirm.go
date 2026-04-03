package tui

import (
	"fmt"
)

func (m Model) viewConfirm() string {
	providers := m.selectedProviderNames()
	primary := "claude"
	if len(providers) > 0 {
		primary = providers[0]
	}
	body := []string{
		centerLine(34, "Configuration Summary"),
		"",
		fmt.Sprintf("  %s %s", padRight(LabelStyle.Render("Project"), 10), ValueStyle.Render(m.preset)),
		fmt.Sprintf("  %s %s", padRight(LabelStyle.Render("Primary"), 10), ValueStyle.Render(primary)),
		fmt.Sprintf("  %s %s", padRight(LabelStyle.Render("Providers"), 10), ValueStyle.Render(joinPatterns(providers))),
		fmt.Sprintf("  %s %s", padRight(LabelStyle.Render("Include"), 10), ValueStyle.Render(joinPatterns(m.includePatterns))),
		fmt.Sprintf("  %s %s", padRight(LabelStyle.Render("Exclude"), 10), ValueStyle.Render(joinPatterns(m.excludePatterns))),
		fmt.Sprintf("  %s %s", padRight(LabelStyle.Render("AGENTS.md"), 10), ValueStyle.Render(map[bool]string{true: "yes", false: "no"}[m.generateRules])),
	}
	if len(providers) == 0 {
		body = append(body, "", ErrorStyle.Render("No provider selected."))
	}

	return renderScreen(m.width, "", body, "enter: confirm and save • esc: go back")
}
