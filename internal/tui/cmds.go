package tui

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jelsin29/vigilanty/internal/tui/detect"
)

func detectSystemCmd() tea.Cmd {
	return func() tea.Msg {
		return SystemInfoMsg{Info: detect.DetectSystem(context.Background())}
	}
}

func detectProvidersCmd() tea.Cmd {
	return func() tea.Msg {
		return AIProvidersMsg{Providers: detect.DetectProviders(context.Background())}
	}
}

func discoverModelsCmd(provider string) tea.Cmd {
	return func() tea.Msg {
		return ModelsDiscoveredMsg{Models: detect.DiscoverModels(context.Background(), provider)}
	}
}
