package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/sonmezerekrem/atrisos/app/internal/podman"
)

// containerDot returns a colored status dot for a single container.
func containerDot(c podman.ContainerInfo) string {
	if c.Health == "unhealthy" {
		return colorRed.Render("●")
	}
	switch c.Status {
	case "running":
		return colorGreen.Render("●")
	case "exited", "dead":
		return colorDim.Render("○")
	default:
		return colorYellow.Render("◑")
	}
}

// RenderDetail renders the right-panel stack detail component.
//
// Parameters:
//   - s: selected StackStatus (may be nil if no stack is selected)
//   - width, height: available inner dimensions for the panel
func RenderDetail(s *podman.StackStatus, width, height int) string {
	if s == nil {
		return "\n" + dimStyle.Render("  select a stack to see details")
	}

	var sb strings.Builder

	// Name heading + divider.
	sb.WriteString(lipgloss.NewStyle().Bold(true).Render(s.Stack.Name))
	sb.WriteString("\n")
	divLen := width - 4
	if divLen < 1 {
		divLen = 1
	}
	sb.WriteString(dimStyle.Render(strings.Repeat("─", divLen)))
	sb.WriteString("\n\n")

	// Status line.
	runCount := 0
	for _, c := range s.Containers {
		if c.Status == "running" {
			runCount++
		}
	}
	var stateStr string
	switch s.State {
	case "running":
		stateStr = colorGreen.Render(fmt.Sprintf("running (%d container(s))", len(s.Containers)))
	case "partial":
		stateStr = colorYellow.Render(fmt.Sprintf("partial (%d/%d running)", runCount, len(s.Containers)))
	case "stopped":
		stateStr = colorDim.Render("stopped")
	default:
		stateStr = colorDim.Render("unknown")
	}
	sb.WriteString(fmt.Sprintf("%-10s %s\n", "Status:", stateStr))

	// Domain(s).
	if len(s.Stack.Config.Domains) > 0 {
		for i, d := range s.Stack.Config.Domains {
			scheme := "https"
			if d.TLS == "false" {
				scheme = "http"
			}
			url := fmt.Sprintf("%s://%s", scheme, d.Host)
			label := ""
			if i == 0 {
				label = "Domain:"
			}
			sb.WriteString(fmt.Sprintf("%-10s %s\n", label, url))
		}
	}
	sb.WriteString("\n")

	// Containers table.
	if len(s.Containers) > 0 {
		sb.WriteString(sectionHeadingStyle.Render("Containers"))
		sb.WriteString("\n")
		for _, c := range s.Containers {
			dot := containerDot(c)
			health := c.Health
			if health == "" {
				health = "–"
			}
			age := c.Started
			if age == "" {
				age = "–"
			}
			sb.WriteString(fmt.Sprintf("  %-16s %s %-10s %-10s %s\n",
				c.Service, dot, c.Status, health, age))
		}
		sb.WriteString("\n")
	}

	// Meta key-values.
	if len(s.Stack.Config.Meta) > 0 {
		sb.WriteString(sectionHeadingStyle.Render("Meta"))
		sb.WriteString("\n")
		for k, v := range s.Stack.Config.Meta {
			sb.WriteString(fmt.Sprintf("  %s: %s\n", k, v))
		}
		sb.WriteString("\n")
	}

	// Description.
	if s.Stack.Config.Description != "" {
		sb.WriteString(dimStyle.Render(s.Stack.Config.Description))
		sb.WriteString("\n\n")
	}

	// Key hints.
	hints := keyHintStyle.Render("[u]update") + " " +
		keyHintStyle.Render("[r]restart") + " " +
		keyHintStyle.Render("[l]logs") + " " +
		keyHintStyle.Render("[x]down") + " " +
		keyHintStyle.Render("[e]shell")
	sb.WriteString(hints)

	// Trim to height to avoid overflow.
	result := sb.String()
	lines := strings.Split(result, "\n")
	if len(lines) > height {
		lines = lines[:height]
	}
	return strings.Join(lines, "\n")
}
