package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const brailleSpace = "⠀"

var eyeLines = []string{
" █████   █████ █████           ███  ████                        █████              ",
"░░███   ░░███ ░░███           ░░░  ░░███                       ░░███               ",
" ░███    ░███  ░███   ███████ ████  ░███   ██████   ████████   ███████   █████ ████",
" ░███    ░███  ░███  ███░░███░░███  ░███  ░░░░░███ ░░███░░███ ░░░███░   ░░███ ░███ ",
" ░░███   ███   ░███ ░███ ░███ ░███  ░███   ███████  ░███ ░███   ░███     ░███ ░███ ",
"  ░░░█████░    ░███ ░███ ░███ ░███  ░███  ███░░███  ░███ ░███   ░███ ███ ░███ ░███ ",
"    ░░███      █████░░███████ █████ █████░░████████ ████ █████  ░░█████  ░░███████ ",
"     ░░░      ░░░░░  ░░░░░███░░░░░ ░░░░░  ░░░░░░░░ ░░░░ ░░░░░    ░░░░░    ░░░░░███ ",
"                     ███ ░███                                             ███ ░███ ",
"                    ░░██████                                             ░░██████  ",
"                     ░░░░░░                                               ░░░░░░   ",
	
}

var gradientColors = []lipgloss.Color{ColorFlux, ColorGold, ColorCircuit, ColorAfterburn}

func centerBrailleLines(lines []string) []string {
	maxWidth := 0
	stripped := make([]string, len(lines))

	for i, line := range lines {
		content := strings.Trim(line, brailleSpace)
		stripped[i] = content
		if width := lipgloss.Width(content); width > maxWidth {
			maxWidth = width
		}
	}

	centered := make([]string, len(lines))
	for i, line := range stripped {
		leftPadding := (maxWidth - lipgloss.Width(line)) / 2
		centered[i] = strings.Repeat(brailleSpace, leftPadding) + line
	}

	return centered
}

var versionStyle = lipgloss.NewStyle().Foreground(ColorSmoke)

func RenderVersion(version string) string {
	return versionStyle.Render(fmt.Sprintf("Vigilanty %s — Pre-commit verification pipeline", version))
}

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
