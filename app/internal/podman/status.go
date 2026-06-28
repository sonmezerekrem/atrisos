package podman

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/sonmezerekrem/atrisos/internal/stack"
)

// ContainerInfo holds live container state for one container in a stack.
type ContainerInfo struct {
	ID      string
	Name    string
	Service string
	Status  string // "running", "exited", "created", "paused"
	Health  string // "healthy", "unhealthy", "starting", "" (no healthcheck)
	Started string // age string
}

// StackStatus combines a Stack with its live container state.
type StackStatus struct {
	Stack      *stack.Stack
	Containers []ContainerInfo
	State      string // "running" | "partial" | "stopped" | "unknown"
}

// podmanContainerJSON represents a single entry from `podman ps --format json`.
type podmanContainerJSON struct {
	ID        string            `json:"Id"`
	Names     []string          `json:"Names"`
	State     string            `json:"State"`
	Labels    map[string]string `json:"Labels"`
	StartedAt int64             `json:"StartedAt"`
	Health    *podmanHealth     `json:"Health"`
}

type podmanHealth struct {
	Status string `json:"Status"`
}

// GetStackContainers queries Podman for containers belonging to the given
// compose project and returns their status.
func GetStackContainers(projectName string) ([]ContainerInfo, error) {
	cmd := exec.Command("podman", "ps", "-a", "--format", "json",
		"--filter", "label=com.docker.compose.project="+projectName)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("podman ps: %w", err)
	}

	trimmed := strings.TrimSpace(string(out))
	if trimmed == "" || trimmed == "null" || trimmed == "[]" {
		return nil, nil
	}

	var raw []podmanContainerJSON
	if err := json.Unmarshal([]byte(trimmed), &raw); err != nil {
		return nil, fmt.Errorf("parsing podman ps output: %w", err)
	}

	containers := make([]ContainerInfo, 0, len(raw))
	for _, r := range raw {
		name := ""
		if len(r.Names) > 0 {
			name = strings.TrimPrefix(r.Names[0], "/")
		}
		health := ""
		if r.Health != nil {
			health = r.Health.Status
		}
		started := ""
		if r.StartedAt > 0 {
			started = formatAge(time.Unix(r.StartedAt, 0))
		}
		containers = append(containers, ContainerInfo{
			ID:      r.ID,
			Name:    name,
			Service: r.Labels["com.docker.compose.service"],
			Status:  r.State,
			Health:  health,
			Started: started,
		})
	}
	return containers, nil
}

// formatAge returns a human-readable duration since t.
func formatAge(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return fmt.Sprintf("%ds", int(d.Seconds()))
	case d < time.Hour:
		return fmt.Sprintf("%dm", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh", int(d.Hours()))
	default:
		return fmt.Sprintf("%dd", int(d.Hours()/24))
	}
}

// AggregateState computes the overall state from a container list.
func AggregateState(containers []ContainerInfo) string {
	if len(containers) == 0 {
		return "stopped"
	}
	running := 0
	for _, c := range containers {
		if c.Status == "running" {
			running++
		}
	}
	switch {
	case running == len(containers):
		return "running"
	case running == 0:
		return "stopped"
	default:
		return "partial"
	}
}

// GetStackStatus returns a StackStatus for a single stack.
func GetStackStatus(s *stack.Stack) (*StackStatus, error) {
	containers, err := GetStackContainers(s.Name)
	if err != nil {
		return &StackStatus{Stack: s, State: "unknown"}, err
	}
	return &StackStatus{
		Stack:      s,
		Containers: containers,
		State:      AggregateState(containers),
	}, nil
}

// GetAllStatus returns StackStatus for a slice of stacks (runs sequentially).
func GetAllStatus(stacks []*stack.Stack) ([]*StackStatus, error) {
	result := make([]*StackStatus, 0, len(stacks))
	for _, s := range stacks {
		ss, err := GetStackStatus(s)
		if err != nil {
			result = append(result, &StackStatus{Stack: s, State: "unknown"})
			continue
		}
		result = append(result, ss)
	}
	return result, nil
}
