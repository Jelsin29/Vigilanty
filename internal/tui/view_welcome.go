package tui

func (m Model) viewWelcome() string {
	items := []string{"Setup project", "Quit"}
	body := []string{
		centerLine(60, Banner(m.width)),
		"",
		centerLine(60, RenderVersion(m.version)),
		"",
	}
	for i, item := range items {
		line := "  " + item
		if i == m.welcomeCursor {
			line = cursor + SelectedStyle.Render(item)
		} else {
			line = "  " + UnselectedStyle.Render(item)
		}
		body = append(body, line)
	}

	return renderScreen(m.width, "", body, "j/k: navigate • enter: select • q: quit")
}
