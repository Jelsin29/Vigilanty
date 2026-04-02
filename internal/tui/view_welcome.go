package tui

import "strings"

func (m Model) viewWelcome() string {
	box := BorderedBox(strings.Join([]string{
		centerLine(60, Banner(m.width)),
		"",
		centerLine(60, TitleStyle.Render("Vigilanty Setup")),
	}, "\n"))

	return renderScreen(m.width, "", []string{box}, "press enter to continue • esc to quit")
}
