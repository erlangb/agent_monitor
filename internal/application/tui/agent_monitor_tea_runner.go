package tui

import (
	"context"
	"erlangb/agentmonitor/internal/application"
	"erlangb/agentmonitor/internal/usecase"
	"log/slog"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/erlangb/agentmeter"
)

// screen represents the active child screen.
type screen int

const (
	screenSelector screen = iota // use-case list
	screenRunner                  // active use-case REPL
)

// RunnerDoneMsg is emitted by the runner sub-model when it exits.
type RunnerDoneMsg struct {
	SwitchRequested bool
}

// AgentMonitorTeaRunner is the single top-level tea.Model passed to tea.NewProgram.
// It owns the full application lifecycle and delegates to child sub-models.
// AgentMeter and UseCases are held directly — no callback closures.
type AgentMonitorTeaRunner struct {
	ctx      context.Context
	screen   screen
	width    int
	height   int
	meter    *agentmeter.Meter
	useCases []usecase.UseCase

	// sub-models
	selector   SelectorModel
	runner     RunnerModel
	teaProgram *tea.Program
}

// NewAgentMonitorTeaRunner returns an application.Runner that drives a bubbletea TUI.
func NewAgentMonitorTeaRunner() application.Runner {
	return &AgentMonitorTeaRunner{}
}

func (m *AgentMonitorTeaRunner) Run(ctx context.Context, useCases []usecase.UseCase, meter *agentmeter.Meter) error {
	slog.Info("running agent monitor")
	m.ctx = ctx
	m.meter = meter
	m.useCases = useCases
	m.teaProgram = tea.NewProgram(m, tea.WithAltScreen())

	_, err := m.teaProgram.Run()
	return err
}

func (m *AgentMonitorTeaRunner) Close() {
	if m.teaProgram != nil {
		m.teaProgram.Quit()
	}
}

// Init initializes the selector and returns its Init command.
func (m *AgentMonitorTeaRunner) Init() tea.Cmd {
	m.selector = NewSelectorModel(m.useCases)
	return m.selector.Init()
}

func (m *AgentMonitorTeaRunner) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	// User picked a use case → open runner.
	case SelectorDoneMsg:
		if msg.UC == nil {
			return m, tea.Quit
		}
		m.screen = screenRunner
		m.runner = newRunnerModel(m.ctx, msg.UC, m.meter)
		m.runner.width = m.width
		m.runner.height = m.height
		m.runner.viewport.Width = min(m.width, maxViewportWidth)
		m.runner.viewport.Height = max(3, m.height-15)
		return m, m.runner.Init()

	// Runner exited → switch back to selector or quit.
	case RunnerDoneMsg:
		if msg.SwitchRequested {
			m.screen = screenSelector
			m.selector = NewSelectorModel(m.useCases)
			return m, m.selector.Init()
		}
		return m, tea.Quit
	}

	// Delegate to the active sub-model.
	switch m.screen {
	case screenSelector:
		updated, cmd := m.selector.Update(msg)
		m.selector = updated
		return m, cmd

	case screenRunner:
		updated, cmd := m.runner.Update(msg)
		m.runner = updated
		return m, cmd
	}

	return m, nil
}

func (m *AgentMonitorTeaRunner) View() string {
	switch m.screen {
	case screenSelector:
		return m.selector.View()
	case screenRunner:
		return m.runner.View()
	}
	return ""
}