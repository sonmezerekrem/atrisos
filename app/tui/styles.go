package tui

import "github.com/charmbracelet/lipgloss"

var (
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Padding(0, 1)

	selectedItemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.AdaptiveColor{Light: "#006400", Dark: "#00cc00"}).
				Bold(true)

	normalItemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#333333", Dark: "#cccccc"})

	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#888888", Dark: "#555555"})

	panelBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.AdaptiveColor{Light: "#aaaaaa", Dark: "#444444"})

	statusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#555555", Dark: "#888888"}).
			Padding(0, 1)

	sectionHeadingStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.AdaptiveColor{Light: "#333333", Dark: "#dddddd"})

	keyHintStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#007777", Dark: "#00cccc"})
)

// Colored styles for status indicators.
var (
	colorGreen  = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#008000", Dark: "#00cc00"})
	colorYellow = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#cc8800", Dark: "#ffcc00"})
	colorRed    = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#cc0000", Dark: "#ff4444"})
	colorCyan   = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#007777", Dark: "#00cccc"})
	colorBlue   = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#0000cc", Dark: "#4488ff"})
	colorDim    = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#aaaaaa", Dark: "#555555"})
)
