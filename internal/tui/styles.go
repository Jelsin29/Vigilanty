package tui

import "github.com/charmbracelet/lipgloss"

const (
	PrimaryColor = "#7C3AED"
	SuccessColor = "#22C55E"
	ErrorColor   = "#EF4444"
	MutedColor   = "#6B7280"
)

var (
	TitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(PrimaryColor)).
			Bold(true)

	LabelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(MutedColor)).
			Bold(true)

	ValueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF"))

	SuccessStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(SuccessColor)).
			Bold(true)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ErrorColor)).
			Bold(true)

	MutedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(MutedColor))

	KeyHintStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(MutedColor)).
			Italic(true)

	CheckMark = SuccessStyle.Render("✓")
	CrossMark = ErrorStyle.Render("✗")
)

func BorderedBox(content string) string {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(PrimaryColor)).
		Padding(0, 1).
		Render(content)
}
