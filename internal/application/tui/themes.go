package tui

// Claude-inspired palette — warm coral accent on dark terminal.
//
//	coral(209) · white(255) · silver(245) · dim(238) · border(236) · error(167)

import "github.com/charmbracelet/lipgloss"

var primary = lipgloss.Color("#018281")
var secondary = lipgloss.Color("238")
var white = lipgloss.Color("255")
var red = lipgloss.Color("167")

var (
	// Runner
	runnerTitleStyle    = lipgloss.NewStyle().Bold(true).Foreground(primary)
	runnerDescStyle     = lipgloss.NewStyle().Foreground(secondary)
	runnerExampleStyle  = lipgloss.NewStyle().Italic(true).Foreground(secondary)
	runnerYouStyle      = lipgloss.NewStyle().Bold(true).Foreground(primary)
	runnerHintStyle     = lipgloss.NewStyle().Foreground(secondary)
	runnerThinkingStyle = lipgloss.NewStyle().Foreground(primary).Blink(true)
	runnerErrorStyle    = lipgloss.NewStyle().Foreground(red)
	answerStyle         = lipgloss.NewStyle().Foreground(white)
	answerBorderStyle   = lipgloss.NewStyle().
				BorderStyle(lipgloss.NormalBorder()).
				BorderLeft(true).
				BorderForeground(primary).
				PaddingLeft(1)
	replDividerStyle  = lipgloss.NewStyle().Foreground(secondary)
	tokenSummaryStyle = lipgloss.NewStyle().Foreground(white)

	// Inspect
	inspectBadgeStyle   = lipgloss.NewStyle().Bold(true).Foreground(white).Background(primary).Padding(0, 1)
	inspectAccentStyle  = lipgloss.NewStyle().Bold(true).Foreground(primary)
	inspectDividerStyle = lipgloss.NewStyle().Foreground(secondary)
	inspectStepStyle    = lipgloss.NewStyle().
				BorderStyle(lipgloss.NormalBorder()).
				BorderLeft(true).
				BorderForeground(primary).
				PaddingLeft(1)
)

var (
	// Selector
	selectorTitleStyle = lipgloss.NewStyle().Bold(true).Foreground(primary).Padding(0, 1).MarginBottom(1)
	selectedItemStyle  = lipgloss.NewStyle().
				Bold(true).
				PaddingLeft(1).
				BorderStyle(lipgloss.NormalBorder()).
				BorderLeft(true).
				BorderForeground(primary).
				Foreground(white)
	normalItemStyle   = lipgloss.NewStyle().PaddingLeft(3).Foreground(primary)
	selectedDescStyle = lipgloss.NewStyle().PaddingLeft(5).Foreground(primary)
	normalDescStyle   = lipgloss.NewStyle().PaddingLeft(5).Foreground(secondary)
	selectorHintStyle = lipgloss.NewStyle().Foreground(secondary)
	selectorBoxStyle  = lipgloss.NewStyle().
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(secondary).
				Padding(1, 2)
)
