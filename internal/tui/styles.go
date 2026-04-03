package tui

import "github.com/charmbracelet/lipgloss"

var (
	// BTTF warm palette — all in the amber/gold/orange family
	ColorFlux      = lipgloss.Color("#FF6F00") // bright orange — eye gradient, accents
	ColorGold      = lipgloss.Color("#FFB300") // gold — titles, selected items
	ColorAmber     = lipgloss.Color("#D4A026") // muted amber — frame border
	ColorCircuit   = lipgloss.Color("#FF1744") // red — errors, eye gradient tail
	ColorAfterburn = lipgloss.Color("#E53935") // deep red — secondary error
	ColorChrome    = lipgloss.Color("#E8E0D0") // warm white — main text
	ColorAlloy     = lipgloss.Color("#B8AD9E") // warm silver — labels
	ColorSmoke     = lipgloss.Color("#8B8378") // warm gray — hints, muted
	ColorReadout   = lipgloss.Color("#00E676") // neon green — success (universal)

	FrameStyle = lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder()).
			BorderForeground(ColorAmber).
			Padding(1, 2).
			Foreground(ColorChrome)

	TitleStyle = lipgloss.NewStyle().
			Foreground(ColorGold).
			Bold(true)

	LabelStyle = lipgloss.NewStyle().
			Foreground(ColorAlloy).
			Bold(true)

	ValueStyle = lipgloss.NewStyle().
			Foreground(ColorChrome)

	SelectedStyle = lipgloss.NewStyle().
			Foreground(ColorGold).
			Bold(true)

	UnselectedStyle = lipgloss.NewStyle().
			Foreground(ColorChrome)

	SuccessStyle = lipgloss.NewStyle().
			Foreground(ColorReadout).
			Bold(true)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(ColorCircuit).
			Bold(true)

	MutedStyle = lipgloss.NewStyle().
			Foreground(ColorSmoke)

	HelpStyle = lipgloss.NewStyle().
			Foreground(ColorSmoke)

	KeyHintStyle = HelpStyle.Italic(true)

	CheckMark = SuccessStyle.Render("✓")
	CrossMark = ErrorStyle.Render("✗")
)

func BorderedBox(content string) string {
	return lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(ColorAmber).
		Padding(1, 2).
		Foreground(ColorChrome).
		Render(content)
}
