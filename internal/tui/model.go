package tui

import (
	"runtime/debug"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/jelsin29/vigilanty/internal/tui/detect"
)

type Result struct {
	ProjectType     string
	FilePatterns    []string
	ExcludePatterns []string
	Providers       []string
	Provider        string
	RulesFile       string
	GenerateRules   bool
}

type Model struct {
	step    Step
	preset  string
	version string
	width   int
	height  int

	sysInfo   detect.SystemInfo
	providers []detect.ProviderInfo

	spinner      spinner.Model
	textInput    textinput.Model
	includeInput textinput.Model
	excludeInput textinput.Model

	welcomeCursor         int
	providerCursor        int
	modelCursor           int
	rulesCursor           int
	patternField          int
	selectedProvider      int
	activeModelProvider   int
	selectedModel         string
	modelOptions          []string
	includePatterns       []string
	excludePatterns       []string
	selectedProviders     map[int]bool
	selectedProviderModel map[int]string
	providerSelectionsSet bool
	providerError         string
	generateRules         bool

	sysDetected       bool
	providersDetected bool
	modelsDetected    bool
	cancelled         bool
	done              bool
}

func New(preset string) Model {
	defaults := defaultResult(preset)
	spin := spinner.New(spinner.WithSpinner(spinner.MiniDot))
	spin.Style = TitleStyle

	modelInput := newInput("Model: ")
	includeInput := newInput("> ")
	excludeInput := newInput("> ")
	includeInput.Blur()
	excludeInput.Blur()

	return Model{
		step:                  StepWelcome,
		preset:                defaults.ProjectType,
		spinner:               spin,
		textInput:             modelInput,
		includeInput:          includeInput,
		excludeInput:          excludeInput,
		version:               appVersion(),
		selectedProvider:      -1,
		activeModelProvider:   -1,
		selectedProviders:     map[int]bool{},
		selectedProviderModel: map[int]string{},
		includePatterns:       append([]string(nil), defaults.FilePatterns...),
		excludePatterns:       append([]string(nil), defaults.ExcludePatterns...),
		generateRules:         defaults.GenerateRules,
	}
}

func NewModel(preset string) Model {
	return New(preset)
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
		m.initializeProviderSelections()
		m.clampProviderCursor()
		return m, tea.Batch(cmds...)
	case ModelsDiscoveredMsg:
		m.modelOptions = msg.Models
		if m.activeModelProvider >= 0 && m.activeModelProvider < len(m.providers) {
			m.providers[m.activeModelProvider].Models = append([]string(nil), msg.Models...)
		}
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

func (m Model) Result() *Result {
	if m.cancelled {
		return nil
	}

	providers := m.selectedProviderNames()
	provider := "claude"
	if len(providers) > 0 {
		provider = providers[0]
	}

	return &Result{
		ProjectType:     m.preset,
		FilePatterns:    append([]string(nil), m.includePatterns...),
		ExcludePatterns: append([]string(nil), m.excludePatterns...),
		Providers:       providers,
		Provider:        provider,
		RulesFile:       "AGENTS.md",
		GenerateRules:   m.generateRules,
	}
}

func (m Model) Cancelled() bool {
	return m.cancelled
}

func defaultResult(preset string) Result {
	projectType := strings.ToLower(strings.TrimSpace(preset))
	if projectType == "" {
		projectType = "generic"
	}

	includePatterns, excludePatterns := defaultPatternsForPreset(projectType)

	return Result{
		ProjectType:     projectType,
		FilePatterns:    append([]string(nil), includePatterns...),
		ExcludePatterns: append([]string(nil), excludePatterns...),
		Providers:       []string{"claude"},
		Provider:        "claude",
		RulesFile:       "AGENTS.md",
		GenerateRules:   true,
	}
}

func defaultPatternsForPreset(preset string) ([]string, []string) {
	switch strings.ToLower(strings.TrimSpace(preset)) {
	case "go":
		return []string{"*.go"}, []string{"*_test.go", "vendor/"}
	case "typescript":
		return []string{"*.ts", "*.tsx"}, []string{"*.test.ts", "*.spec.ts", "node_modules/"}
	case "node":
		return []string{"*.js", "*.jsx", "*.ts", "*.tsx"}, []string{"*.test.js", "*.spec.js", "node_modules/", "dist/"}
	case "python":
		return []string{"*.py"}, []string{"*_test.py", "test_*", "__pycache__/"}
	case "rust":
		return []string{"*.rs"}, []string{"target/"}
	case "java":
		return []string{"*.java"}, []string{"*Test.java", "build/", "target/"}
	case "dotnet":
		return []string{"*.cs"}, []string{"*Tests.cs", "bin/", "obj/"}
	case "ruby":
		return []string{"*.rb"}, []string{"*_spec.rb", "spec/"}
	case "swift":
		return []string{"*.swift"}, []string{"*Tests.swift", ".build/"}
	case "php":
		return []string{"*.php"}, []string{"*Test.php", "vendor/"}
	default:
		return []string{"*"}, nil
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

func appVersion() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "dev"
	}
	version := strings.TrimSpace(info.Main.Version)
	if version == "" || version == "(devel)" {
		return "dev"
	}
	return version
}
