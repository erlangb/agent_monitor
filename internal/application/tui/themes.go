package tui

// Claude-inspired palette — warm coral accent on dark terminal.
//
//	coral(209) · white(255) · silver(245) · dim(238) · border(236) · error(167)

import "github.com/charmbracelet/lipgloss"

var (
	// Runner
	runnerTitleStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("255"))
	runnerDescStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	runnerExampleStyle  = lipgloss.NewStyle().Italic(true).Foreground(lipgloss.Color("238"))
	runnerYouStyle      = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#018281"))
	runnerHintStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("238"))
	runnerThinkingStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#018281")).Blink(true)
	runnerErrorStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("167"))
	answerStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("255"))
	answerBorderStyle   = lipgloss.NewStyle().
				BorderStyle(lipgloss.NormalBorder()).
				BorderLeft(true).
				BorderForeground(lipgloss.Color("#018281")).
				PaddingLeft(1)
	replDividerStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("236"))
	tokenSummaryStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("238"))

	// Inspect
	inspectBadgeStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("255")).Background(lipgloss.Color("#018281")).Padding(0, 1)
	inspectAccentStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#018281"))
	inspectDividerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("236"))
	inspectStepStyle    = lipgloss.NewStyle().
				BorderStyle(lipgloss.NormalBorder()).
				BorderLeft(true).
				BorderForeground(lipgloss.Color("#018281")).
				PaddingLeft(1)
)

var (
	// Selector
	selectorTitleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#018281")).Padding(0, 1).MarginBottom(1)
	selectedItemStyle  = lipgloss.NewStyle().
				Bold(true).
				PaddingLeft(1).
				BorderStyle(lipgloss.NormalBorder()).
				BorderLeft(true).
				BorderForeground(lipgloss.Color("#018281")).
				Foreground(lipgloss.Color("255"))
	normalItemStyle   = lipgloss.NewStyle().PaddingLeft(3).Foreground(lipgloss.Color("245"))
	selectedDescStyle = lipgloss.NewStyle().PaddingLeft(5).Foreground(lipgloss.Color("245"))
	normalDescStyle   = lipgloss.NewStyle().PaddingLeft(5).Foreground(lipgloss.Color("238"))
	selectorHintStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("238"))
	selectorBoxStyle  = lipgloss.NewStyle().
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("236")).
				Padding(1, 2)
)