package tui

type Step int

const (
	StepWelcome Step = iota
	StepSystemInfo
	StepAIDetect
	StepProviderSelect
	StepSubProviderSelect
	StepModelSelect
	StepPatterns
	StepAgentsMd
	StepConfirm
)

type StepCompleteMsg struct{ Next Step }
