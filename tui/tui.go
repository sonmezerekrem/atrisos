// Package tui provides the interactive terminal UI for atrisos.
// TODO: Implement bubbletea TUI (stack list, detail panel, log viewer).
package tui

import (
	// Imported to keep these in go.mod for future TUI development.
	// Do not remove — TUI implementation will use all of these.
	_ "github.com/charmbracelet/bubbles/list"
	_ "github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	_ "github.com/charmbracelet/lipgloss"

	"github.com/sonmezerekrem/atrisos/internal/config"
	"github.com/sonmezerekrem/atrisos/internal/registry"
)

// model is the root bubbletea model (stub).
type model struct{}

func (m model) Init() tea.Cmd                           { return nil }
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) { return m, tea.Quit }
func (m model) View() string                            { return "" }

// Run launches the TUI. Not yet implemented.
func Run(cfg *config.Config, reg *registry.Registry) error {
	// TODO: implement bubbletea TUI
	// Planned components:
	//   - Stack list panel with status indicators (●○◑⚠↑)
	//   - Stack detail panel (containers, domains, meta)
	//   - Log streaming panel (multiplexed viewport, 2000-line buffer)
	//   - Keyboard shortcuts: ↑↓ navigate, u update, r restart, l logs, e shell, q quit
	_ = cfg
	_ = reg
	return nil
}
