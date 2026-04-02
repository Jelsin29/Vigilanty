package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var eyeLines = []string{
	"⠀⠀⠀⠀⡀⣤⠠⠤⠤⠄⣤⣀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀",
	"⠀⢀⣴⠿⠊⠁⠀⠀⠀⠀⠀⠉⠳⣦⡀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀",
	"⢠⡿⠁⠀⠀⠀⠀⠀⠀⠀⠀⠀⠢⡊⠱⡄⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀",
	"⣿⠳⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠞⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀",
	"⡏⡇⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢰⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀",
	"⣥⢁⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢀⣼⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀",
	"⠱⣷⠈⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢀⣾⠂⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀",
	"⠀⠘⢷⡤⡀⠀⠀⠀⠀⠀⠀⣀⣴⢿⣿⡦⢀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀",
	"⠀⠀⠀⠈⠚⠶⠦⠤⠤⠴⠲⠛⠁⠀⠘⢹⣶⣶⡄⡀⠀⠀⠀⠀⠀⠀⠀",
	"⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠻⣿⣮⣺⣄⠀⠀⠀⠀⠀⠀",
	"⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠈⠻⣿⣯⣷⣄⡀⠀⠀⠀",
	"⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠈⠻⣿⣿⣿⣦⡀⠀",
	"⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠈⠙⢿⣿⣝⣄",
	"⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠈⠛⠟⠿",
}

var gradientColors = []lipgloss.Color{ColorMauve, ColorLavender, ColorBlue, ColorTeal, ColorGreen}

func RenderLogo() string {
	total := len(eyeLines)
	bands := len(gradientColors)
	if total == 0 || bands == 0 {
		return ""
	}

	lines := make([]string, 0, total)
	for i, line := range eyeLines {
		bandIdx := (i * bands) / total
		if bandIdx >= bands {
			bandIdx = bands - 1
		}
		lines = append(lines, lipgloss.NewStyle().Foreground(gradientColors[bandIdx]).Render(line))
	}

	return strings.Join(lines, "\n")
}

func Banner(width int) string {
	if width < 40 {
		return "[ VIGILANTY ]"
	}
	return RenderLogo()
}
