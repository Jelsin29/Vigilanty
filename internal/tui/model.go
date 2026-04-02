package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/jelsin29/vigilanty/internal/tui/detect"
	"github.com/jelsin29/vigilanty/internal/wizard"
)

type Model struct {
	step   Step
	preset string
	width  int
	height int

	sysInfo   detect.SystemInfo
	providers []detect.ProviderInfo

	spinner      spinner.Model
	textInput    textinput.Model
	includeInput textinput.Model
	excludeInput textinput.Model

	selectedProvider int
	providerCursor   int
	modelCursor      int
	rulesCursor      int
	patternField     int
	selectedModel    string
	modelOptions     []string
	includePatterns  []string
	excludePatterns  []string
	generateRules    bool

	sysDetected       bool
	providersDetected bool
	modelsDetected    bool
	cancelled         bool
	done              bool
}

func NewModel(preset string) Model {
	defaults := wizard.DefaultResult(preset)
	spin := spinner.New(spinner.WithSpinner(spinner.MiniDot))
	spin.Style = TitleStyle

	modelInput := newInput("Model: ")
	includeInput := newInput("> ")
	excludeInput := newInput("> ")
	includeInput.Blur()
	excludeInput.Blur()

	return Model{
		step:             StepWelcome,
		preset:           defaults.ProjectType,
		spinner:          spin,
		textInput:        modelInput,
		includeInput:     includeInput,
		excludeInput:     excludeInput,
		selectedProvider: -1,
		includePatterns:  append([]string(nil), defaults.FilePatterns...),
		excludePatterns:  append([]string(nil), defaults.ExcludePatterns...),
		generateRules:    defaults.GenerateRules,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, detectSystemCmd())
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	if spinCmd := m.updateSpinner(msg); spinCmd != nil {
		cmds = append(cmds, spinCmd)
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, tea.Batch(cmds...)
	case tea.KeyMsg:
		if quitCmd, handled := m.handleGlobalKeys(msg); handled {
			cmds = append(cmds, quitCmd)
			return m, tea.Batch(cmds...)
		}
	case SystemInfoMsg:
		m.sysInfo = msg.Info
		m.sysDetected = true
		return m, tea.Batch(cmds...)
	case AIProvidersMsg:
		m.providers = msg.Providers
		m.providersDetected = true
		m.syncProviderCursor()
		return m, tea.Batch(cmds...)
	case ModelsDiscoveredMsg:
		m.modelOptions = msg.Models
		m.modelsDetected = true
		if len(m.modelOptions) == 0 {
			m.textInput.SetValue(m.selectedModel)
			cmds = append(cmds, m.focusManualModel())
		}
		return m, tea.Batch(cmds...)
	case StepCompleteMsg:
		m.step = msg.Next
		cmds = append(cmds, m.enterStep(msg.Next)...)
		return m, tea.Batch(cmds...)
	}

	stepCmd := m.updateCurrentStep(msg)
	if stepCmd != nil {
		cmds = append(cmds, stepCmd)
	}

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	switch m.step {
	case StepWelcome:
		return m.viewWelcome()
	case StepSystemInfo:
		return m.viewSystemInfo()
	case StepAIDetect:
		return m.viewAIDetect()
	case StepProviderSelect:
		return m.viewProviderSelect()
	case StepModelSelect:
		return m.viewModelSelect()
	case StepPatterns:
		return m.viewPatterns()
	case StepAgentsMd:
		return m.viewAgentsMd()
	case StepConfirm:
		return m.viewConfirm()
	default:
		return ""
	}
}

func (m Model) Result() *wizard.InitResult {
	if m.cancelled {
		return nil
	}

	provider := "claude"
	if p, ok := m.selectedProviderInfo(); ok {
		provider = p.Name
		if p.NeedsModel && strings.TrimSpace(m.selectedModel) != "" {
			provider += ":" + strings.TrimSpace(m.selectedModel)
		}
	}

	return &wizard.InitResult{
		ProjectType:     m.preset,
		FilePatterns:    append([]string(nil), m.includePatterns...),
		ExcludePatterns: append([]string(nil), m.excludePatterns...),
		Provider:        provider,
		RulesFile:       "AGENTS.md",
		GenerateRules:   m.generateRules,
	}
}

func newInput(prompt string) textinput.Model {
	input := textinput.New()
	input.Prompt = prompt
	input.PromptStyle = ValueStyle
	input.TextStyle = ValueStyle
	input.PlaceholderStyle = MutedStyle
	return input
}
