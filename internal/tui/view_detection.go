package tui

import "fmt"

func (m Model) viewSystemInfo() string {
	if !m.sysDetected {
		body := []string{m.spinner.View() + " Detecting your system..."}
		return renderScreen(m.width, "System Detection", body, "esc: back")
	}

	gitStatus := CrossMark + " missing"
	if len(m.sysInfo.Tools) > 0 && m.sysInfo.Tools[0].Found {
		gitStatus = CheckMark + " found"
	}

	body := []string{
		fmt.Sprintf("  %s %s", padRight(LabelStyle.Render("OS"), 8), ValueStyle.Render(m.sysInfo.OS)),
		fmt.Sprintf("  %s %s", padRight(LabelStyle.Render("Arch"), 8), ValueStyle.Render(m.sysInfo.Arch)),
		fmt.Sprintf("  %s %s", padRight(LabelStyle.Render("Shell"), 8), ValueStyle.Render(emptyFallback(m.sysInfo.Shell, "unknown"))),
		fmt.Sprintf("  %s %s", padRight(LabelStyle.Render("Git"), 8), gitStatus),
	}

	return renderScreen(m.width, "System Detection", body, "enter: continue • esc: back")
}

func (m Model) viewAIDetect() string {
	if !m.providersDetected {
		body := []string{m.spinner.View() + " Detecting AI providers..."}
		return renderScreen(m.width, "AI Provider Detection", body, "esc: back")
	}

	body := make([]string, 0, len(m.providers))
	for _, provider := range m.providers {
		body = append(body, fmt.Sprintf("  %s %s", padRight(ValueStyle.Render(providerLabel(provider.Name)), 12), providerStatus(provider)))
	}

	return renderScreen(m.width, "AI Provider Detection", body, "enter: continue • esc: back")
}

func padRight(value string, width int) string {
	if width <= 0 {
		return value
	}
	missing := width - len(stripANSI(value))
	if missing <= 0 {
		return value
	}
	return value + spaces(missing)
}

func emptyFallback(value string, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
