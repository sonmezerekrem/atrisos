package restic

import (
	"compress/bzip2"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/sonmezerekrem/atrisos/app/internal/config"
)

const targetVersion = "0.17.3"

// BinPath returns the path where the bundled restic binary lives.
func BinPath() string {
	return filepath.Join(config.Dir(), "bin", "restic")
}

// EnsureInstalled downloads the restic binary for the current OS/arch if it
// doesn't exist or if the installed version doesn't match the target version.
func EnsureInstalled() error {
	bin := BinPath()

	// Check if already installed and working.
	if _, err := os.Stat(bin); err == nil {
		out, err := exec.Command(bin, "version").Output()
		if err == nil && strings.Contains(string(out), targetVersion) {
			return nil
		}
	}

	fmt.Printf("→ downloading restic v%s...\n", targetVersion)

	url := downloadURL()

	// Create bin directory if needed.
	if err := os.MkdirAll(filepath.Dir(bin), 0755); err != nil {
		return fmt.Errorf("creating restic bin dir: %w", err)
	}

	// Download to temp file.
	resp, err := http.Get(url) //nolint:gosec
	if err != nil {
		return fmt.Errorf("downloading restic: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("downloading restic: HTTP %d", resp.StatusCode)
	}

	// Decompress bzip2 stream.
	bzReader := bzip2.NewReader(resp.Body)

	tmp, err := os.CreateTemp("", "restic-*")
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)

	if _, err := io.Copy(tmp, bzReader); err != nil {
		tmp.Close()
		return fmt.Errorf("decompressing restic: %w", err)
	}
	tmp.Close()

	// Move to final path and make executable.
	if err := os.Rename(tmpPath, bin); err != nil {
		// On macOS Rename across devices may fail; fall back to copy.
		if copyErr := copyFile(tmpPath, bin); copyErr != nil {
			return fmt.Errorf("installing restic: %w", copyErr)
		}
	}

	if err := os.Chmod(bin, 0755); err != nil {
		return fmt.Errorf("chmod restic: %w", err)
	}

	fmt.Printf("✓ restic v%s installed\n", targetVersion)
	return nil
}

// Run executes the bundled restic binary with the given arguments,
// streaming output to stdout/stderr.
func Run(args ...string) error {
	cmd := exec.Command(BinPath(), args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Output executes the bundled restic binary and returns captured stdout.
func Output(args ...string) ([]byte, error) {
	return exec.Command(BinPath(), args...).Output()
}

// downloadURL returns the GitHub release download URL for the current OS/arch.
func downloadURL() string {
	goos := runtime.GOOS
	goarch := runtime.GOARCH
	return fmt.Sprintf(
		"https://github.com/restic/restic/releases/download/v%s/restic_%s_%s_%s.bz2",
		targetVersion, targetVersion, goos, goarch,
	)
}

// copyFile copies src to dst (used as fallback when Rename fails).
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}
