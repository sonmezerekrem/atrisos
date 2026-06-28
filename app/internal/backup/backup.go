package backup

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/sonmezerekrem/atrisos/internal/restic"
	"github.com/sonmezerekrem/atrisos/internal/stack"
)

// BackupRunConfig holds the parameters for a single backup run.
type BackupRunConfig struct {
	Destination string // restic repo: "s3://..." or local path (~ expanded)
	DryRun      bool
}

// Run performs a backup of a stack's named volumes using the bundled restic.
func Run(s *stack.Stack, cfg *BackupRunConfig) error {
	if err := restic.EnsureInstalled(); err != nil {
		return fmt.Errorf("restic: %w", err)
	}

	dest := expandDest(cfg.Destination)

	// Determine volumes to back up.
	volumes := s.Config.Backup.Volumes
	if len(volumes) == 0 {
		discovered, err := discoverVolumes(s.Name)
		if err != nil {
			return fmt.Errorf("discovering volumes: %w", err)
		}
		volumes = discovered
	}

	if len(volumes) == 0 {
		fmt.Printf("→ no volumes found for stack %s — nothing to back up\n", s.Name)
		return nil
	}

	// Resolve host mount paths for each volume.
	var paths []string
	for _, vol := range volumes {
		mp, err := VolumeMount(vol)
		if err != nil {
			fmt.Fprintf(os.Stderr, "⚠ skipping volume %s: %v\n", vol, err)
			continue
		}
		paths = append(paths, mp)
	}

	if len(paths) == 0 {
		return fmt.Errorf("no mountable volumes found for stack %s", s.Name)
	}

	if cfg.DryRun {
		fmt.Printf("→ dry-run: would back up to %s\n", dest)
		for _, p := range paths {
			fmt.Printf("  %s\n", p)
		}
		return nil
	}

	// Ensure the restic repo is initialized.
	initArgs := []string{"-r", dest, "init"}
	initCmd := exec.Command(restic.BinPath(), initArgs...)
	initCmd.Env = append(os.Environ(), "RESTIC_PASSWORD="+s.Name)
	// Ignore error — repo may already be initialized.
	_ = initCmd.Run()

	// Run the backup.
	args := append([]string{"-r", dest, "backup"}, paths...)
	cmd := exec.Command(restic.BinPath(), args...)
	cmd.Env = append(os.Environ(), "RESTIC_PASSWORD="+s.Name)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("restic backup: %w", err)
	}

	return nil
}

// VolumeMount returns the host mountpoint for a named Podman volume.
func VolumeMount(volumeName string) (string, error) {
	out, err := exec.Command("podman", "volume", "inspect", volumeName,
		"--format", "{{.Mountpoint}}").Output()
	if err != nil {
		return "", fmt.Errorf("podman volume inspect %s: %w", volumeName, err)
	}
	mp := strings.TrimSpace(string(out))
	if mp == "" {
		return "", fmt.Errorf("empty mountpoint for volume %s", volumeName)
	}
	return mp, nil
}

// discoverVolumes finds all named Podman volumes for the given project.
func discoverVolumes(projectName string) ([]string, error) {
	out, err := exec.Command("podman", "volume", "ls",
		"--filter", "label=com.docker.compose.project="+projectName,
		"--format", "json").Output()
	if err != nil {
		return nil, fmt.Errorf("podman volume ls: %w", err)
	}

	var entries []struct {
		Name string `json:"Name"`
	}
	if err := parseVolumeListJSON(string(out), &entries); err != nil {
		return nil, err
	}

	names := make([]string, 0, len(entries))
	for _, e := range entries {
		names = append(names, e.Name)
	}
	return names, nil
}

func parseVolumeListJSON(raw string, entries *[]struct {
	Name string `json:"Name"`
}) error {
	if err := json.Unmarshal([]byte(raw), entries); err != nil {
		return fmt.Errorf("parsing volume list: %w", err)
	}
	return nil
}

// expandDest expands ~ in a destination path and handles S3 URLs unchanged.
func expandDest(dest string) string {
	if strings.HasPrefix(dest, "s3://") {
		return dest
	}
	if len(dest) >= 2 && dest[:2] == "~/" {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, dest[2:])
	}
	return dest
}
