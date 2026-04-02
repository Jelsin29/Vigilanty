package ui

import "sync/atomic"

const (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
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
