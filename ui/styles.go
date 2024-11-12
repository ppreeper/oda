package ui

import "github.com/charmbracelet/lipgloss"

// ========= Banner =========
// func cText(color, msg string) string {
// 	return color + msg + "{{ .AnsiColor.Default }}"
// }

// ========= LipGloss =========

var HeaderStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#EEEEEE")).
	Bold(true)

var EvenRowStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#777777"))

var OddRowStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#AAAAAA"))

var StepStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#EEEEEE")).
	Bold(true)

var SubStepStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#228822")).
	Bold(true)

var WarningStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#FF8800")).
	Bold(true)

var ErrorStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#EE0000")).
	Bold(true)
