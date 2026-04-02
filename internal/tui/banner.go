package tui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const brailleSpace = "⠀"

var eyeLines = centerBrailleLines([]string{
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
})

var gradientColors = []lipgloss.Color{ColorMauve, ColorLavender, ColorBlue, ColorTeal, ColorGreen}
var versionGradientStops = []lipgloss.Color{ColorBTTFOrange, ColorBTTFGold, ColorBTTFRed}

func centerBrailleLines(lines []string) []string {
	maxWidth := 0
	trimmed := make([]string, len(lines))

	for i, line := range lines {
		trimmedLine := strings.TrimRight(line, brailleSpace)
		trimmed[i] = trimmedLine
		if width := lipgloss.Width(trimmedLine); width > maxWidth {
			maxWidth = width
		}
	}

	centered := make([]string, len(lines))
	for i, line := range trimmed {
		leftPadding := (maxWidth - lipgloss.Width(line)) / 2
		centered[i] = strings.Repeat(brailleSpace, leftPadding) + line
	}

	return centered
}

func RenderVersion(version string) string {
	text := fmt.Sprintf("Vigilanty %s — Pre-commit verification pipeline", version)
	runes := []rune(text)
	if len(runes) == 0 {
		return ""
	}

	var b strings.Builder
	b.Grow(len(text) * 12)

	for i, r := range runes {
		color := interpolateGradient(versionGradientStops, i, len(runes))
		b.WriteString(lipgloss.NewStyle().Foreground(color).Bold(true).Render(string(r)))
	}

	return b.String()
}

func interpolateGradient(stops []lipgloss.Color, index, total int) lipgloss.Color {
	if len(stops) == 0 {
		return ColorLavender
	}
	if len(stops) == 1 || total <= 1 {
		return stops[0]
	}

	position := float64(index) / float64(total-1)
	segments := len(stops) - 1
	scaled := position * float64(segments)
	segment := int(scaled)
	if segment >= segments {
		segment = segments - 1
	}
	localT := scaled - float64(segment)

	from := mustHexColor(string(stops[segment]))
	to := mustHexColor(string(stops[segment+1]))

	return lipgloss.Color(fmt.Sprintf("#%02X%02X%02X",
		lerpChannel(from[0], to[0], localT),
		lerpChannel(from[1], to[1], localT),
		lerpChannel(from[2], to[2], localT),
	))
}

func mustHexColor(value string) [3]int {
	trimmed := strings.TrimPrefix(value, "#")
	if len(trimmed) != 6 {
		panic("unsupported color format: " + value)
	}

	return [3]int{
		mustParseHex(trimmed[0:2]),
		mustParseHex(trimmed[2:4]),
		mustParseHex(trimmed[4:6]),
	}
}

func mustParseHex(value string) int {
	parsed, err := strconv.ParseInt(value, 16, 0)
	if err != nil {
		panic(err)
	}
	return int(parsed)
}

func lerpChannel(from, to int, t float64) int {
	return int(float64(from) + (float64(to-from) * t))
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
