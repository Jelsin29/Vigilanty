package ui

import (
	"strings"
	"unicode/utf8"
)

var runBannerBorder = "▀▀▀▀▀ ▀▀▀▀▀ ▀▀▀▀▀ ▀▀▀▀▀ ▀▀▀▀▀ ▀▀▀▀▀ ▀▀▀▀▀ ▀▀▀▀▀ ▀▀▀▀▀"

var runBannerArt = []string{
	"  ██  ██ ▄▄  ▄▄▄▄ ▄▄ ▄▄     ▄▄▄  ▄▄  ▄▄ ▄▄▄▄▄▄ ▄▄ ▄▄",
	"  ██▄▄██ ██ ██ ▄▄ ██ ██    ██▀██ ███▄██   ██   ▀███▀",
}

// lastArtLine is the final line of the art, with the version appended on the right.
const lastArtLinePrefix = "   ▀██▀  ██ ▀███▀ ██ ██▄▄▄ ██▀██ ██ ▀██   ██     █"

// RunBanner returns the ASCII art banner for the CLI run output.
// The version string is centered between the art and the bottom border.
func RunBanner(version string) string {
	lines := make([]string, 0, 8)

	// Top border (2 lines)
	lines = append(lines, Colorize(BttfAmber, runBannerBorder))
	lines = append(lines, Colorize(BttfAmber, runBannerBorder))

	// Art (first lines)
	for _, line := range runBannerArt {
		lines = append(lines, Colorize(BttfGold, line))
	}

	// Last art line with version on the right
	borderWidth := utf8.RuneCountInString(runBannerBorder)
	prefixWidth := utf8.RuneCountInString(lastArtLinePrefix)
	gap := borderWidth - prefixWidth - utf8.RuneCountInString(version)
	if gap < 2 {
		gap = 2
	}
	lastLine := lastArtLinePrefix + strings.Repeat(" ", gap) + version
	lines = append(lines, Colorize(BttfGold, lastLine))

	// Bottom border (2 lines)
	lines = append(lines, Colorize(BttfAmber, runBannerBorder))
	lines = append(lines, Colorize(BttfAmber, runBannerBorder))

	return strings.Join(lines, "\n")
}

func centerText(value string, width int) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return strings.Repeat(" ", width)
	}

	padding := width - utf8.RuneCountInString(trimmed)
	if padding <= 0 {
		return trimmed
	}

	left := padding / 2
	right := padding - left
	return strings.Repeat(" ", left) + trimmed + strings.Repeat(" ", right)
}
