package tui

import (
	"context"
	"erlangb/agentmonitor/internal/usecase"
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/erlangb/agentmeter"
	"github.com/erlangb/agentmeter/reasoning"
)

const (
	maxViewportWidth  = 120
	maxReplHeight     = 20
	maxInspectHeight  = 30
)

type runnerResultMsg struct {
	answer string
	err    error
}

// RunnerModel runs a single use case: accepts input, shows a spinner while
// thinking, and renders the answer inside a thin bordered box.
// When the user types /inspect it switches to the step-by-step inspect view.
type RunnerModel struct {
	ctx     context.Context
	uc      usecase.UseCase
	meter   *agentmeter.Meter
	input   textinput.Model
	spinner spinner.Model

	// REPL state
	question string
	answer   string
	errMsg   string
	thinking bool

	// Run tracking: each user query maps to a slice of History() snapshots.
	// runStarts[i] is the first snapshot index in meter.History() for query i.
	// The slice for query i is History()[runStarts[i] : runStarts[i+1]].
	runStarts      []int
	prevHistoryLen int

	// layout
	width  int
	height int

	// inspect state
	inspecting  bool
	inspectRun  int // 0-based index into runStarts (= which user query)
	inspectStep int // 0-based snapshot index within that query's History() slice
	viewport    viewport.Model
}

func newRunnerModel(ctx context.Context, uc usecase.UseCase, meter *agentmeter.Meter) RunnerModel {
	ti := textinput.New()
	ti.Placeholder = "Ask something… (esc to go back, /inspect to browse steps)"
	ti.CharLimit = 1024
	ti.Focus()

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("5"))

	vp := viewport.New(80, 20)

	return RunnerModel{ctx: ctx, uc: uc, meter: meter, input: ti, spinner: sp, viewport: vp}
}

func (m RunnerModel) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, m.spinner.Tick)
}

func (m RunnerModel) Update(msg tea.Msg) (RunnerModel, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewport.Width = min(msg.Width, maxViewportWidth)
		m.input.Width = min(msg.Width, maxViewportWidth)
		if m.inspecting {
			m.viewport.Height = max(3, msg.Height-6)
		} else {
			m.viewport.Height = max(3, msg.Height-15)
		}

	case runnerResultMsg:
		m.thinking = false
		if msg.err != nil {
			m.errMsg = msg.err.Error()
			m.answer = ""
		} else {
			m.answer = msg.answer
			m.errMsg = ""
			m.runStarts = append(m.runStarts, m.prevHistoryLen)
			m.prevHistoryLen = len(m.meter.History())
			m.viewport.Height = max(3, m.height-15)
			m.viewport.SetContent(wrapContent(answerStyle.Render(m.answer), min(m.width, maxViewportWidth)))
			m.viewport.GotoTop()
		}
		return m, m.input.Focus()

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case tea.KeyMsg:
		// Inspect modal intercepts all keys first.
		if m.inspecting {
			return m.updateInspect(msg)
		}
		if m.thinking {
			if msg.String() == "ctrl+c" {
				return m, tea.Quit
			}
			return m, nil
		}
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "esc":
			m.meter.ClearHistory()
			m.runStarts = nil
			m.prevHistoryLen = 0
			m.answer = ""
			m.question = ""
			m.errMsg = ""
			return m, func() tea.Msg { return RunnerDoneMsg{SwitchRequested: true} }
		case "enter":
			query := strings.TrimSpace(m.input.Value())
			if query == "" {
				return m, nil
			}
			if strings.HasPrefix(query, "/inspect") {
				return m.handleInspectCmd(query)
			}
			m.question = query
			m.answer = ""
			m.errMsg = ""
			m.input.Reset()
			m.thinking = true
			return m, tea.Batch(m.spinner.Tick, runUCCmd(m.ctx, m.uc, query))
		}
	}

	// Forward scroll events to the viewport while inspecting.
	if m.inspecting {
		var cmd tea.Cmd
		m.viewport, cmd = m.viewport.Update(msg)
		return m, cmd
	}

	if !m.thinking {
		var cmds []tea.Cmd
		if m.answer != "" {
			var vpCmd tea.Cmd
			m.viewport, vpCmd = m.viewport.Update(msg)
			cmds = append(cmds, vpCmd)
		}
		var inputCmd tea.Cmd
		m.input, inputCmd = m.input.Update(msg)
		cmds = append(cmds, inputCmd)
		return m, tea.Batch(cmds...)
	}
	return m, nil
}

// handleInspectCmd parses "/inspect [N]" and opens the inspect view.
// N is 1-based user-query index. If omitted, the last query is selected.
func (m RunnerModel) handleInspectCmd(query string) (RunnerModel, tea.Cmd) {
	if len(m.runStarts) == 0 {
		m.input.Reset()
		return m, nil
	}

	runIdx := len(m.runStarts) - 1 // default: last user query
	if arg := strings.TrimSpace(strings.TrimPrefix(query, "/inspect")); arg != "" {
		if n, err := strconv.Atoi(arg); err == nil && n >= 1 && n <= len(m.runStarts) {
			runIdx = n - 1
		}
	}

	snaps := m.runSnapshots(runIdx)
	m.input.Reset()
	if len(snaps) == 0 {
		return m, nil
	}
	m.inspecting = true
	m.inspectRun = runIdx
	m.inspectStep = 0
	m.viewport.Height = max(3, m.height-6)
	m.viewport.SetContent(wrapContent(renderStep(snaps[0]), min(m.width, maxViewportWidth)))
	m.viewport.GotoTop()
	return m, nil
}

// updateInspect handles all key events while the inspect modal is open.
func (m RunnerModel) updateInspect(msg tea.KeyMsg) (RunnerModel, tea.Cmd) {
	switch msg.String() {
	case "esc", "q":
		m.inspecting = false
		m.viewport.Height = max(3, m.height-15)
		if m.answer != "" {
			m.viewport.SetContent(wrapContent(answerStyle.Render(m.answer), min(m.width, maxViewportWidth)))
		}
		return m, nil

	case "right", "l", "n":
		snaps := m.runSnapshots(m.inspectRun)
		if m.inspectStep+1 < len(snaps) {
			m.inspectStep++
			m.viewport.SetContent(wrapContent(renderStep(snaps[m.inspectStep]), min(m.width, maxViewportWidth)))
			m.viewport.GotoTop()
		}
		return m, nil

	case "left", "h", "p":
		if m.inspectStep > 0 {
			snaps := m.runSnapshots(m.inspectRun)
			m.inspectStep--
			m.viewport.SetContent(wrapContent(renderStep(snaps[m.inspectStep]), min(m.width, maxViewportWidth)))
			m.viewport.GotoTop()
		}
		return m, nil

	case "tab":
		if m.inspectRun+1 < len(m.runStarts) {
			m.inspectRun++
			m.inspectStep = 0
			if snaps := m.runSnapshots(m.inspectRun); len(snaps) > 0 {
				m.viewport.SetContent(wrapContent(renderStep(snaps[0]), min(m.width, maxViewportWidth)))
				m.viewport.GotoTop()
			}
		}
		return m, nil

	case "shift+tab":
		if m.inspectRun > 0 {
			m.inspectRun--
			m.inspectStep = 0
			if snaps := m.runSnapshots(m.inspectRun); len(snaps) > 0 {
				m.viewport.SetContent(wrapContent(renderStep(snaps[0]), min(m.width, maxViewportWidth)))
				m.viewport.GotoTop()
			}
		}
		return m, nil

	default:
		var cmd tea.Cmd
		m.viewport, cmd = m.viewport.Update(msg)
		return m, cmd
	}
}

func (m RunnerModel) View() string {
	if m.inspecting {
		return m.inspectView()
	}
	return m.replView()
}

// replView renders the normal question/answer REPL screen.
func (m RunnerModel) replView() string {
	w := min(m.width, maxViewportWidth)
	div := replDividerStyle.Render(strings.Repeat("─", max(w, 40)))

	// ── top: header + question ───────────────────────────────────────────
	var top strings.Builder
	top.WriteString(div + "\n")
	top.WriteString(runnerTitleStyle.Render("[ "+m.uc.Name()+" ]") + "  " + runnerDescStyle.Render(m.uc.Description()) + "\n")
	if m.question == "" {
		if ex := m.uc.ExampleInput(); ex != "" {
			top.WriteString(runnerExampleStyle.Render("eg: "+ex) + "\n")
		}
	}
	top.WriteString("\n")
	if m.question != "" {
		top.WriteString(runnerYouStyle.Render("You:") + "  " + lipgloss.NewStyle().Width(max(w-6, 0)).Render(m.question) + "\n")
		top.WriteString("\n")
	}

	// ── bottom: input + hint (always pinned) ─────────────────────────────
	hint := "enter to ask  •  esc back to selector"
	if m.answer != "" {
		hint += "  •  ↑↓ scroll answer"
	}
	if len(m.runStarts) > 0 {
		hint += fmt.Sprintf("  •  /inspect %d to browse steps", len(m.runStarts))
	}

	// ── compose ──────────────────────────────────────────────────────────
	switch {
	case m.thinking:
		return top.String() +
			m.spinner.View() + runnerThinkingStyle.Render(" thinking…") + "\n\n" +
			div + "\n" +
			m.input.View() + "\n" +
			runnerHintStyle.Render(hint)

	case m.errMsg != "":
		return top.String() +
			runnerErrorStyle.Render("Error: "+m.errMsg) + "\n\n" +
			div + "\n" +
			m.input.View() + "\n" +
			runnerHintStyle.Render(hint)

	case m.answer != "":
		lastRunSnaps := m.runSnapshots(len(m.runStarts) - 1)
		return top.String() +
			div + "\n" +
			answerBorderStyle.Render(m.viewport.View()) + "\n" +
			div + "\n" +
			tokenSummaryStyle.Render(tokenSummaryLine(lastRunSnaps)) + "\n" +
			div + "\n" +
			m.input.View() + "\n" +
			runnerHintStyle.Render(hint)

	default:
		return top.String() +
			div + "\n" +
			m.input.View() + "\n" +
			runnerHintStyle.Render(hint)
	}
}

// inspectView renders the step-by-step inspect modal (replaces the full screen).
func (m RunnerModel) inspectView() string {
	if len(m.runStarts) == 0 {
		return runnerHintStyle.Render("no run data — press esc")
	}
	snaps := m.runSnapshots(m.inspectRun)

	divider := inspectDividerStyle.Render(strings.Repeat("─", max(min(m.width, maxViewportWidth), 40)))

	badge := inspectBadgeStyle.Render("Inspect")
	runPart := inspectAccentStyle.Render(fmt.Sprintf("Run %d/%d", m.inspectRun+1, len(m.runStarts)))
	stepPart := inspectAccentStyle.Render(fmt.Sprintf("Step %d/%d", m.inspectStep+1, len(snaps)))
	title := badge + runnerHintStyle.Render(" — ") + runPart + runnerHintStyle.Render("  ·  ") + stepPart

	keys := runnerHintStyle.Render("↑↓ scroll  ·  ← → steps  ·  ⇥ next run  ·  ⇤ prev run  ·  esc close")

	return divider + "\n" + title + "\n" + keys + "\n" + divider + "\n" + m.viewport.View() + "\n" + divider
}

// runSnapshots returns the slice of History() snapshots that belong to user
// query runIdx. Each snapshot corresponds to one agent-chain execution.
func (m RunnerModel) runSnapshots(runIdx int) []agentmeter.Snapshot {
	hist := m.meter.History()
	start := m.runStarts[runIdx]
	end := len(hist)
	if runIdx+1 < len(m.runStarts) {
		end = m.runStarts[runIdx+1]
	}
	return hist[start:end]
}

// tokenSummaryLine aggregates token usage across all snapshots in a run.
func tokenSummaryLine(snaps []agentmeter.Snapshot) string {
	p, buf := reasoning.NewBufferedPrinter()
	p.PrintHistoryTokenSummary(snaps)
	return buf.String()
}

// renderStep renders a full agent-chain snapshot (one "step" in the inspect view).
func renderStep(snap agentmeter.Snapshot) string {
	p, buf := reasoning.NewBufferedPrinter(reasoning.WithMaxContentLen(3000))
	p.Print(snap)
	return inspectStepStyle.Render(buf.String())
}

// wrapContent word-wraps s to width using lipgloss, so viewport content never
// overflows the terminal horizontally.
func wrapContent(s string, width int) string {
	if width <= 0 {
		return s
	}
	return lipgloss.NewStyle().Width(min(width, maxViewportWidth)).Render(s)
}

func runUCCmd(ctx context.Context, uc usecase.UseCase, input string) tea.Cmd {
	return func() tea.Msg {
		answer, err := uc.Run(ctx, input)
		return runnerResultMsg{answer: answer, err: err}
	}
}