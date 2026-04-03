package ui

import "sync/atomic"

const (
	Reset  = "\033[0m"
	Bold   = "\033[1m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
	Cyan   = "\033[36m"
	Gray   = "\033[90m"

	// BTTF palette (matches TUI styles)
	BttfFlux  = "\033[38;2;255;111;0m"   // #FF6F00 bright orange
	BttfGold  = "\033[38;2;255;179;0m"   // #FFB300 gold
	BttfAmber = "\033[38;2;212;160;38m"  // #D4A026 muted amber
	BttfSmoke = "\033[38;2;139;131;120m" // #8B8378 warm gray
)

var colorsEnabled atomic.Bool

func init() {
	colorsEnabled.Store(true)
}

func SetColorsEnabled(enabled bool) {
	colorsEnabled.Store(enabled)
}

func Colorize(color string, message string) string {
	if !colorsEnabled.Load() {
		return message
	}
	return color + message + Reset
}

func SuccessText(message string) string {
	return Colorize(Green, message)
}

func WarningText(message string) string {
	return Colorize(Yellow, message)
}

func ErrorText(message string) string {
	return Colorize(Red, message)
}
