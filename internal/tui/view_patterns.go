package tui

func (m Model) viewPatterns() string {
	body := []string{
		"File patterns to review",
		"",
		"  " + LabelStyle.Render("Include:") + " " + ValueStyle.Render(joinPatterns(m.includePatterns)),
		"  Add more (comma-separated, enter to keep):",
		"  " + m.includeInput.View(),
		"",
		"  " + LabelStyle.Render("Exclude:") + " " + ValueStyle.Render(joinPatterns(m.excludePatterns)),
		"  Add more (comma-separated, enter to keep):",
		"  " + m.excludeInput.View(),
	}

	return renderScreen(m.width, "", body, "enter: continue • esc: back")
}
