package cmd

import "github.com/jelsin29/vigilanty/internal/ui"

func successText(message string) string {
	return ui.SuccessText(message)
}

func warningText(message string) string {
	return ui.WarningText(message)
}

func errorText(message string) string {
	return ui.ErrorText(message)
}
