package tui

import (
	"erlangb/agentmonitor/internal/usecase"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// SelectorDoneMsg is emitted when the user confirms a selection or quits.
// UC is nil when the user quits without selecting.
type SelectorDoneMsg struct {
	UC usecase.UseCase
}

// SelectorModel is a bubbletea sub-model for use-case selection.
// It does not call tea.Quit — it emits SelectorDoneMsg so the parent model can react.
type SelectorModel struct {
	useCases []usecase.UseCase
	cursor   int
}

func NewSelectorModel(useCases []usecase.UseCase) SelectorModel {
	return SelectorModel{useCases: useCases}
}

func (m SelectorModel) Init() tea.Cmd { return nil }

func (m SelectorModel) Update(msg tea.Msg) (SelectorModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.useCases)-1 {
				m.cursor++
			}
		case "enter":
			uc := m.useCases[m.cursor]
			return m, func() tea.Msg { return SelectorDoneMsg{UC: uc} }
		case "ctrl+c", "q":
			return m, func() tea.Msg { return SelectorDoneMsg{UC: nil} }
		}
	}
	return m, nil
}

func (m SelectorModel) View() string {
	var sb strings.Builder
	sb.WriteString(selectorTitleStyle.Render("AgentSelector"))
	sb.WriteString("\n")
	for i, uc := range m.useCases {
		if i == m.cursor {
			sb.WriteString(selectedItemStyle.Render("▶  " + uc.Name()))
			sb.WriteString("\n")
			sb.WriteString(selectedDescStyle.Render(uc.Description()))
		} else {
			sb.WriteString(normalItemStyle.Render("   " + uc.Name()))
			sb.WriteString("\n")
			sb.WriteString(normalDescStyle.Render(uc.Description()))
		}
		sb.WriteString("\n\n")
	}
	sb.WriteString(selectorHintStyle.Render("↑/↓  navigate • enter  select • q  quit"))
	return selectorBoxStyle.Render(sb.String())
}