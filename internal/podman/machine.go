package podman

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

// podmanMachineInfo represents a single entry from `podman machine list --format json`.
type podmanMachineInfo struct {
	Name    string `json:"Name"`
	Running bool   `json:"Running"`
}

// EnsureMachine ensures the named Podman machine exists and is running.
// On Linux this is a no-op. On macOS it creates and/or starts the machine.
func EnsureMachine(name string) error {
	if runtime.GOOS != "darwin" {
		return nil
	}

	machines, err := listMachines()
	if err != nil {
		// If we can't list machines, try to proceed anyway.
		return nil
	}

	var found *podmanMachineInfo
	for i := range machines {
		if machines[i].Name == name {
			found = &machines[i]
			break
		}
	}

	if found == nil {
		// Machine doesn't exist — create it.
		fmt.Printf("→ Creating Podman machine %q (this may take a moment)...\n", name)
		cmd := exec.Command("podman", "machine", "init", name)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("podman machine init %q: %w", name, err)
		}
	} else if found.Running {
		// Already running — nothing to do.
		return nil
	}

	// Start the machine.
	fmt.Printf("→ Starting Podman machine %q...\n", name)
	cmd := exec.Command("podman", "machine", "start", name)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("podman machine start %q: %w", name, err)
	}

	return nil
}

// listMachines returns the list of Podman machines.
func listMachines() ([]podmanMachineInfo, error) {
	cmd := exec.Command("podman", "machine", "list", "--format", "json")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	var machines []podmanMachineInfo
	if err := json.Unmarshal(out, &machines); err != nil {
		return nil, err
	}
	return machines, nil
}

// SocketPath returns the Podman socket path for the given machine.
// On Linux: /run/user/<uid>/podman/podman.sock.
// On macOS: parsed from podman machine inspect.
func SocketPath(machineName string) (string, error) {
	switch runtime.GOOS {
	case "linux":
		uid := os.Getuid()
		return fmt.Sprintf("/run/user/%d/podman/podman.sock", uid), nil
	case "darwin":
		return socketPathMacOS(machineName)
	default:
		return "", fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

// socketPathMacOS reads the socket path from podman machine inspect.
func socketPathMacOS(machineName string) (string, error) {
	cmd := exec.Command("podman", "machine", "inspect",
		machineName,
		"--format", "{{.ConnectionInfo.PodmanSocket.Path}}")
	out, err := cmd.Output()
	if err != nil {
		uid := os.Getuid()
		return fmt.Sprintf("/run/user/%s/podman/podman.sock", strconv.Itoa(uid)), nil
	}
	path := strings.TrimSpace(string(out))
	if path == "" {
		return "", fmt.Errorf("empty socket path for machine %q", machineName)
	}
	return path, nil
}
