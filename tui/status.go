package tui

import (
	"os/exec"
	"strings"
)

// checkTraefikStatus queries Podman for the atrisos_traefik container state.
// Returns "running", "stopped", or "unknown".
func checkTraefikStatus() string {
	cmd := exec.Command("podman", "ps",
		"--filter", "name=atrisos_traefik",
		"--format", "{{.State}}")
	out, err := cmd.Output()
	if err != nil {
		return "unknown"
	}
	s := strings.TrimSpace(string(out))
	if s == "" {
		return "stopped"
	}
	return s
}
