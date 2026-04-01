package cmd

const (
	ansiReset  = "\033[0m"
	ansiRed    = "\033[31m"
	ansiGreen  = "\033[32m"
	ansiYellow = "\033[33m"
)

func successText(message string) string {
	return ansiGreen + message + ansiReset
}

func warningText(message string) string {
	return ansiYellow + message + ansiReset
}

func errorText(message string) string {
	return ansiRed + message + ansiReset
}
