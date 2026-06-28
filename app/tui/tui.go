// Package tui provides the interactive terminal UI for atrisos.
package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/sonmezerekrem/atrisos/app/internal/config"
	"github.com/sonmezerekrem/atrisos/app/internal/registry"
)

// Run launches the atrisos TUI in alternate screen mode.
func Run(cfg *config.Config, reg *registry.Registry) error {
	m := newAppModel(cfg, reg)
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
