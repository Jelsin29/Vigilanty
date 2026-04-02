package tui

import "github.com/charmbracelet/lipgloss"

var (
	ColorBase     = lipgloss.Color("#191724")
	ColorSurface  = lipgloss.Color("#1f1d2e")
	ColorOverlay  = lipgloss.Color("#6e6a86")
	ColorText     = lipgloss.Color("#e0def4")
	ColorSubtext  = lipgloss.Color("#908caa")
	ColorLavender = lipgloss.Color("#c4a7e7")
	ColorGreen    = lipgloss.Color("#9ccfd8")
	ColorPeach    = lipgloss.Color("#f6c177")
	ColorRed      = lipgloss.Color("#eb6f92")
	ColorBlue     = lipgloss.Color("#31748f")
	ColorMauve    = lipgloss.Color("#ebbcba")
	ColorTeal     = lipgloss.Color("#9ccfd8")

	FrameStyle = lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder()).
			BorderForeground(ColorLavender).
			Padding(1, 2).
			Foreground(ColorText)

	TitleStyle = lipgloss.NewStyle().
			Foreground(ColorLavender).
			Bold(true)

	LabelStyle = lipgloss.NewStyle().
			Foreground(ColorSubtext).
			Bold(true)

	ValueStyle = lipgloss.NewStyle().
			Foreground(ColorText)

	SelectedStyle = lipgloss.NewStyle().
			Foreground(ColorLavender).
			Bold(true)

	UnselectedStyle = lipgloss.NewStyle().
			Foreground(ColorText)

	SuccessStyle = lipgloss.NewStyle().
			Foreground(ColorGreen).
			Bold(true)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(ColorRed).
			Bold(true)

	MutedStyle = lipgloss.NewStyle().
			Foreground(ColorOverlay)

	HelpStyle = lipgloss.NewStyle().
			Foreground(ColorSubtext)

	KeyHintStyle = HelpStyle.Italic(true)

	CheckMark = SuccessStyle.Render("✓")
	CrossMark = ErrorStyle.Render("✗")
)

func BorderedBox(content string) string {
	return lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(ColorLavender).
		Padding(1, 2).
		Render(content)
}
