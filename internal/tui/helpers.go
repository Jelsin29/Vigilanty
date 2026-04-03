package tui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/jelsin29/vigilanty/internal/tui/detect"
)

const cursor = "▸ "

func (m Model) currentModelProvider() (detect.ProviderInfo, int, bool) {
	if m.activeModelProvider < 0 || m.activeModelProvider >= len(m.providers) {
		return detect.ProviderInfo{}, -1, false
	}

	return m.providers[m.activeModelProvider], m.activeModelProvider, true
}

func (m Model) selectedProviderIndices() []int {
	indices := make([]int, 0, len(m.selectedProviders))
	for idx := range m.selectedProviders {
		if !m.selectedProviders[idx] {
			continue
		}
		indices = append(indices, idx)
	}
	sort.Ints(indices)
	return indices
}

func (m Model) selectedProviderNames() []string {
	indices := m.selectedProviderIndices()
	if len(indices) == 0 && m.selectedProvider >= 0 && m.selectedProvider < len(m.providers) {
		indices = []int{m.selectedProvider}
	}
	providers := make([]string, 0, len(indices))
	for _, idx := range indices {
		provider := strings.TrimSpace(m.providers[idx].Name)
		if provider == "" {
			continue
		}
		model := strings.TrimSpace(m.selectedProviderModel[idx])
		if model == "" && idx == m.selectedProvider {
			model = strings.TrimSpace(m.selectedModel)
		}
		if model != "" {
			providers = append(providers, provider+":"+model)
			continue
		}
		providers = append(providers, provider)
	}
	return providers
}

func (m Model) providerHasSubProviders(idx int) bool {
	if idx < 0 || idx >= len(m.providers) {
		return false
	}
	return len(m.providers[idx].SubProviders) > 0
}

func (m Model) activeSubProviderInfo() (detect.SubProvider, bool) {
	provider, _, ok := m.currentModelProvider()
	if !ok || strings.TrimSpace(m.activeSubProvider) == "" {
		return detect.SubProvider{}, false
	}
	for _, subProvider := range provider.SubProviders {
		if strings.TrimSpace(subProvider.ID) == strings.TrimSpace(m.activeSubProvider) {
			return subProvider, true
		}
	}
	return detect.SubProvider{}, false
}

func (m *Model) initializeProviderSelections() {
	if m.providerSelectionsSet {
		return
	}
	for i, provider := range m.providers {
		if provider.Found {
			m.selectedProviders[i] = true
		}
	}
	for i, provider := range m.providers {
		if provider.Found {
			m.providerCursor = i
			break
		}
	}
	m.providerSelectionsSet = true
	m.syncPrimaryProvider()
}

func (m *Model) syncPrimaryProvider() {
	indices := m.selectedProviderIndices()
	if len(indices) == 0 {
		m.selectedProvider = -1
		return
	}
	m.selectedProvider = indices[0]
}

func (m Model) providerMenuSize() int {
	return len(m.providers) + 2
}

func (m Model) providerMenuTarget() string {
	if m.providerCursor >= len(m.providers) {
		if m.providerCursor == len(m.providers) {
			return "continue"
		}
		return "back"
	}
	return "provider"
}

func (m *Model) clampProviderCursor() {
	max := m.providerMenuSize() - 1
	if max < 0 {
		m.providerCursor = 0
		return
	}
	if m.providerCursor < 0 {
		m.providerCursor = 0
		return
	}
	if m.providerCursor > max {
		m.providerCursor = max
	}
}

func (m *Model) toggleProvider(index int) {
	if index < 0 || index >= len(m.providers) {
		return
	}
	if m.selectedProviders[index] {
		delete(m.selectedProviders, index)
		delete(m.selectedProviderModel, index)
		m.providerError = ""
		m.syncPrimaryProvider()
		return
	}
	m.selectedProviders[index] = true
	m.providerError = ""
	m.syncPrimaryProvider()
}

func (m *Model) beginModelFlow() Step {
	m.syncPrimaryProvider()
	for _, idx := range m.selectedProviderIndices() {
		provider := m.providers[idx]
		if !provider.NeedsModel {
			continue
		}
		if strings.TrimSpace(m.selectedProviderModel[idx]) != "" {
			continue
		}
		m.activeModelProvider = idx
		m.selectedModel = ""
		m.activeSubProvider = ""
		m.modelCursor = 0
		m.subProviderCursor = 0
		m.subProviders = append([]detect.SubProvider(nil), provider.SubProviders...)
		m.modelOptions = append([]string(nil), provider.Models...)
		m.modelsDetected = len(provider.Models) > 0 || len(provider.SubProviders) > 0
		m.textInput.SetValue("")
		if m.providerHasSubProviders(idx) {
			return StepSubProviderSelect
		}
		return StepModelSelect
	}
	m.activeModelProvider = -1
	m.activeSubProvider = ""
	m.selectedModel = ""
	m.subProviders = nil
	m.modelOptions = nil
	m.modelsDetected = false
	return StepPatterns
}

func (m *Model) storeCurrentSubProvider(index int) Step {
	if index < 0 || index >= len(m.subProviders) {
		return StepSubProviderSelect
	}
	subProvider := m.subProviders[index]
	m.activeSubProvider = strings.TrimSpace(subProvider.ID)
	m.modelOptions = append([]string(nil), subProvider.Models...)
	m.modelCursor = 0
	m.modelsDetected = true
	m.textInput.SetValue("")
	return StepModelSelect
}

func (m *Model) storeCurrentModel(value string) Step {
	if m.activeModelProvider < 0 || m.activeModelProvider >= len(m.providers) {
		return StepPatterns
	}
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return StepModelSelect
	}
	if strings.TrimSpace(m.activeSubProvider) != "" {
		trimmed = strings.TrimSpace(m.activeSubProvider) + "/" + trimmed
	}
	m.selectedProviderModel[m.activeModelProvider] = trimmed
	m.selectedModel = trimmed
	m.activeSubProvider = ""
	m.subProviders = nil
	m.modelOptions = nil
	m.modelsDetected = false
	return m.beginModelFlow()
}

func providerLabel(name string) string {
	if strings.TrimSpace(name) == "gh" {
		return "gh copilot"
	}
	return name
}

func providerStatus(provider detect.ProviderInfo) string {
	if provider.Found {
		if len(provider.SubProviders) > 0 {
			return fmt.Sprintf("%s found (%d families)", CheckMark, len(provider.SubProviders))
		}
		if len(provider.Models) > 0 {
			return fmt.Sprintf("%s found (%d models)", CheckMark, len(provider.Models))
		}
		return CheckMark + " found"
	}
	if provider.Name == "lmstudio" {
		return CrossMark + " not running"
	}
	return CrossMark + " not found"
}

func joinPatterns(values []string) string {
	if len(values) == 0 {
		return "(none)"
	}
	return strings.Join(values, ", ")
}

func parseCommaList(raw string) []string {
	parts := strings.Split(raw, ",")
	items := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		items = append(items, part)
	}
	return items
}

func mergePatterns(base []string, extra []string) []string {
	merged := append([]string(nil), base...)
	seen := make(map[string]struct{}, len(base)+len(extra))
	for _, item := range merged {
		seen[item] = struct{}{}
	}
	for _, item := range extra {
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		merged = append(merged, item)
	}
	return merged
}

func centerLine(width int, text string) string {
	if width <= 0 {
		return text
	}
	textWidth := lipgloss.Width(text)
	if textWidth >= width {
		return text
	}
	left := (width - textWidth) / 2
	return strings.Repeat(" ", left) + text
}

func renderScreen(width int, title string, body []string, footer string) string {
	lines := make([]string, 0, len(body)+8)
	lines = append(lines, Banner(width), "")
	if title != "" {
		lines = append(lines, TitleStyle.Render(title), "")
	}
	lines = append(lines, body...)
	if footer != "" {
		lines = append(lines, "", KeyHintStyle.Render(footer))
	}
	content := strings.Join(lines, "\n")
	if width > 0 {
		content = lipgloss.NewStyle().MaxWidth(width - 8).Render(content)
	}
	return FrameStyle.Render(content)
}
