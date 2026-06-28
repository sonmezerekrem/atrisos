package traefik

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

// EnsureNetwork creates the Podman network if it doesn't already exist.
func EnsureNetwork(name string) error {
	cmd := exec.Command("podman", "network", "inspect", name)
	cmd.Stdout = nil
	cmd.Stderr = nil
	if err := cmd.Run(); err == nil {
		// Network already exists.
		return nil
	}

	// Create the network.
	create := exec.Command("podman", "network", "create", name)
	create.Stdout = os.Stdout
	create.Stderr = os.Stderr
	if err := create.Run(); err != nil {
		return fmt.Errorf("creating podman network %q: %w", name, err)
	}
	return nil
}

// PodmanSocketPath returns the Podman socket path.
// On Linux: /run/user/<uid>/podman/podman.sock
// On macOS: parsed from podman machine inspect.
func PodmanSocketPath(machineName string) (string, error) {
	switch runtime.GOOS {
	case "linux":
		uid := os.Getuid()
		return fmt.Sprintf("/run/user/%d/podman/podman.sock", uid), nil
	case "darwin":
		return podmanSocketMacOS(machineName)
	default:
		return "", fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

// podmanSocketMacOS parses the socket path from podman machine inspect.
func podmanSocketMacOS(machineName string) (string, error) {
	cmd := exec.Command("podman", "machine", "inspect",
		machineName,
		"--format", "{{.ConnectionInfo.PodmanSocket.Path}}")
	out, err := cmd.Output()
	if err != nil {
		// Fall back to user-specific runtime path.
		uid := os.Getuid()
		return fmt.Sprintf("/run/user/%s/podman/podman.sock", strconv.Itoa(uid)), nil
	}
	path := strings.TrimSpace(string(out))
	if path == "" {
		return "", fmt.Errorf("podman machine inspect returned empty socket path for machine %q", machineName)
	}
	return path, nil
}
