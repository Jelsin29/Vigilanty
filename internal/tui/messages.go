package tui

import "github.com/jelsin29/vigilanty/internal/tui/detect"

type SystemInfoMsg struct{ Info detect.SystemInfo }

type AIProvidersMsg struct{ Providers []detect.ProviderInfo }

type ModelsDiscoveredMsg struct{ Models []string }
