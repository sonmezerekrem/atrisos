package tui

import (
	"fmt"
	"strings"

	"github.com/sonmezerekrem/atrisos/app/internal/podman"
)

// statusDot returns a colored status indicator for a stack.
func statusDot(s *podman.StackStatus) string {
	// Unhealthy trumps everything.
	for _, c := range s.Containers {
		if c.Health == "unhealthy" {
			return colorRed.Render("⚠")
		}
	}
	switch s.State {
	case "running":
		return colorGreen.Render("●")
	case "partial":
		return colorYellow.Render("◑")
	case "stopped":
		return colorDim.Render("○")
	default:
		return colorDim.Render("○")
	}
}

// updateDot returns "↑" (blue) if an image update is available. Stubbed as
// always false per the spec.
func updateDot() string {
	return "" // stub: always no update available
}

// RenderList renders the left-panel stack list component.
//
// Parameters:
//   - stacks: the (possibly filtered) list of StackStatus to display
//   - cursor: index of the selected item in stacks
//   - filter: current filter string (empty = no filter active)
//   - width, height: available inner dimensions for the panel
//   - outdatedUpdates: map of stack name → has image updates available
func RenderList(stacks []*podman.StackStatus, cursor int, filter string, width, height int, outdatedUpdates map[string]bool) string {
	var sb strings.Builder

	// Heading
	sb.WriteString(sectionHeadingStyle.Render("STACKS"))
	sb.WriteString("\n")

	if filter != "" {
		sb.WriteString(dimStyle.Render("/ " + filter))
		sb.WriteString("\n")
	}
	sb.WriteString("\n")

	if len(stacks) == 0 {
		sb.WriteString(dimStyle.Render("  no stacks found"))
		return sb.String()
	}

	// Find longest name for alignment.
	maxLen := 0
	for _, s := range stacks {
		if len(s.Stack.Name) > maxLen {
			maxLen = len(s.Stack.Name)
		}
	}

	for i, s := range stacks {
		dot := statusDot(s)
		upd := ""
		if outdatedUpdates[s.Stack.Name] {
			upd = colorBlue.Render("↑")
		}
		name := fmt.Sprintf("%-*s", maxLen, s.Stack.Name)

		var line string
		if i == cursor {
			arrow := colorGreen.Render("▶")
			styled := selectedItemStyle.Render(name)
			line = fmt.Sprintf("%s %s %s%s", arrow, styled, dot, upd)
		} else {
			var styled string
			if s.State == "stopped" {
				styled = dimStyle.Render(name)
			} else {
				styled = normalItemStyle.Render(name)
			}
			line = fmt.Sprintf("  %s %s%s", styled, dot, upd)
		}

		sb.WriteString(line)
		if i < len(stacks)-1 {
			sb.WriteString("\n")
		}
	}

	// Trim to height to avoid overflowing the panel border.
	result := sb.String()
	lines := strings.Split(result, "\n")
	if len(lines) > height {
		lines = lines[:height]
	}
	return strings.Join(lines, "\n")
}
