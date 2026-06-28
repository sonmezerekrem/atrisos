package selfupdate

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"sync"
	"time"
)

const githubAPI = "https://api.github.com/repos/sonmezerekrem/atrisos/releases/latest"
const downloadBase = "https://github.com/sonmezerekrem/atrisos/releases/download"

var (
	cachedVersion   string
	cachedAt        time.Time
	cacheMu         sync.Mutex
	cacheTTL        = 24 * time.Hour
)

// LatestVersion fetches the latest atrisos release tag from GitHub API.
// Cached in memory for 24 hours. Returns "" on error.
func LatestVersion() string {
	cacheMu.Lock()
	defer cacheMu.Unlock()

	if cachedVersion != "" && time.Since(cachedAt) < cacheTTL {
		return cachedVersion
	}

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest(http.MethodGet, githubAPI, nil)
	if err != nil {
		return ""
	}
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := client.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ""
	}

	var release struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return ""
	}

	cachedVersion = release.TagName
	cachedAt = time.Now()
	return cachedVersion
}

// Update downloads the release binary for targetVersion, replaces the current
// binary, and prints success. targetVersion should be a tag like "v0.3.0".
func Update(targetVersion string) error {
	if targetVersion == "" {
		return fmt.Errorf("no version specified")
	}

	currentBin, err := os.Executable()
	if err != nil {
		return fmt.Errorf("finding current binary: %w", err)
	}

	url := fmt.Sprintf("%s/%s/atrisos-%s-%s",
		downloadBase, targetVersion, runtime.GOOS, runtime.GOARCH)

	fmt.Printf("→ downloading Atrisos %s...\n", targetVersion)

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Get(url) //nolint:gosec
	if err != nil {
		return fmt.Errorf("downloading release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("downloading release: HTTP %d", resp.StatusCode)
	}

	// Write to a temp file in the same directory as the current binary.
	tmp, err := os.CreateTemp("", "atrisos-update-*")
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)

	if _, err := io.Copy(tmp, resp.Body); err != nil {
		tmp.Close()
		return fmt.Errorf("writing update: %w", err)
	}
	tmp.Close()

	if err := os.Chmod(tmpPath, 0755); err != nil {
		return fmt.Errorf("chmod on update: %w", err)
	}

	// Atomic replace: try Rename first, fall back to copy+delete on macOS
	// across device boundaries.
	if err := os.Rename(tmpPath, currentBin); err != nil {
		if copyErr := copyReplace(tmpPath, currentBin); copyErr != nil {
			return fmt.Errorf("replacing binary: %w", copyErr)
		}
	}

	fmt.Printf("✓ Atrisos updated to %s\n", targetVersion)
	return nil
}

// copyReplace copies src to dst, removing dst first.
func copyReplace(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	if err := os.Remove(dst); err != nil && !os.IsNotExist(err) {
		return err
	}

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}
