package templates

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	repoOwner = "sonmezerekrem"
	repoName  = "atrisos"
	branch    = "main"
)

// cacheDir returns ~/.config/atrisos/templates-cache
func cacheDir() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "atrisos", "templates-cache")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "atrisos", "templates-cache")
}

func manifestURL() string {
	return fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s/templates/manifest.json",
		repoOwner, repoName, branch)
}

func rawURL(relPath string) string {
	return fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s/templates/%s",
		repoOwner, repoName, branch, relPath)
}

// Manifest is the index of available templates.
type Manifest struct {
	Version   string          `json:"version"`
	Templates []TemplateEntry `json:"templates"`
}

// TemplateEntry is one entry in the manifest.
type TemplateEntry struct {
	Name        string `json:"name"`
	Display     string `json:"display"`
	Description string `json:"description"`
}

// TemplateMeta is the content of a template's template.yml file.
type TemplateMeta struct {
	Name        string   `yaml:"name"`
	Display     string   `yaml:"display"`
	Description string   `yaml:"description"`
	Prompts     []Prompt `yaml:"prompts"`
}

// Prompt defines one interactive prompt in the init wizard.
type Prompt struct {
	Name     string   `yaml:"name"`
	Label    string   `yaml:"label"`
	Type     string   `yaml:"type"`     // "string" | "int" | "bool" | "select"
	Default  string   `yaml:"default"`
	Required bool     `yaml:"required"`
	Options  []string `yaml:"options"`  // for type "select"
	Generate string   `yaml:"generate"` // "random_password" | "traefik_me_domain"
}

// httpClient is the shared HTTP client with a 10s timeout.
var httpClient = &http.Client{Timeout: 10 * time.Second}

// download fetches a URL and returns the body bytes.
// Returns an error if the status is not 200; returns a sentinel "not found"
// error for 404 so callers can skip optional files.
func download(url string) ([]byte, error) {
	resp, err := httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("not found: %s", url)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d fetching %s", resp.StatusCode, url)
	}
	return io.ReadAll(resp.Body)
}

// readLocalManifest reads and parses the cached manifest.json.
func readLocalManifest() (*Manifest, error) {
	path := filepath.Join(cacheDir(), "manifest.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parsing cached manifest: %w", err)
	}
	return &m, nil
}

// fetchRemoteManifest downloads and parses the remote manifest with a 5s timeout.
func fetchRemoteManifest() (*Manifest, error) {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(manifestURL())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d fetching remote manifest", resp.StatusCode)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parsing remote manifest: %w", err)
	}
	return &m, nil
}

// LoadManifest loads the local cached manifest. If not present, downloads and
// caches it first. If online and a newer version is available, refreshes the
// full cache. If GitHub is unreachable and a cache exists, uses it silently.
func LoadManifest() (*Manifest, error) {
	manifestPath := filepath.Join(cacheDir(), "manifest.json")

	// No cache at all — must fetch before we can do anything.
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		if err := RefreshCache(); err != nil {
			return nil, fmt.Errorf("fetching templates from GitHub: %w", err)
		}
		return readLocalManifest()
	}

	// Cache exists — try to check for a newer version (best-effort, 5s timeout).
	if remote, err := fetchRemoteManifest(); err == nil {
		if local, lerr := readLocalManifest(); lerr == nil {
			if remote.Version > local.Version {
				// Silently ignore refresh errors; we'll fall back to the cache.
				_ = RefreshCache()
			}
		}
	}

	return readLocalManifest()
}

// fixedTemplateFiles is the set of files we attempt to download per template.
// Files that return 404 are skipped silently.
var fixedTemplateFiles = []string{
	"template.yml",
	"compose.yml.tmpl",
	"config.yml.tmpl",
	".env.tmpl",
	".env.example.tmpl",
}

// RefreshCache re-downloads the manifest and all template files from GitHub.
func RefreshCache() error {
	cd := cacheDir()
	if err := os.MkdirAll(cd, 0o755); err != nil {
		return fmt.Errorf("creating cache dir: %w", err)
	}

	// Download and save manifest.
	data, err := download(manifestURL())
	if err != nil {
		return fmt.Errorf("downloading manifest: %w", err)
	}

	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return fmt.Errorf("parsing manifest: %w", err)
	}
	if err := os.WriteFile(filepath.Join(cd, "manifest.json"), data, 0o644); err != nil {
		return fmt.Errorf("caching manifest: %w", err)
	}

	// Download each template's files.
	for _, entry := range m.Templates {
		dir := filepath.Join(cd, entry.Name)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("creating cache dir for template %s: %w", entry.Name, err)
		}
		for _, fname := range fixedTemplateFiles {
			url := rawURL(entry.Name + "/" + fname)
			fileData, err := download(url)
			if err != nil {
				// Skip missing files (404 or network error) silently.
				continue
			}
			dest := filepath.Join(dir, fname)
			if err := os.WriteFile(dest, fileData, 0o644); err != nil {
				return fmt.Errorf("caching %s/%s: %w", entry.Name, fname, err)
			}
		}
	}

	return nil
}

// LoadTemplateMeta loads and parses the template.yml for a given template name.
// Uses local cache if available; downloads if not.
func LoadTemplateMeta(name string) (*TemplateMeta, error) {
	path := filepath.Join(cacheDir(), name, "template.yml")

	if _, err := os.Stat(path); os.IsNotExist(err) {
		url := rawURL(name + "/template.yml")
		fetched, err := download(url)
		if err != nil {
			return nil, fmt.Errorf("downloading template.yml for %q: %w", name, err)
		}
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return nil, err
		}
		if err := os.WriteFile(path, fetched, 0o644); err != nil {
			return nil, err
		}
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var meta TemplateMeta
	if err := yaml.Unmarshal(data, &meta); err != nil {
		return nil, fmt.Errorf("parsing template.yml for %q: %w", name, err)
	}
	return &meta, nil
}

// TemplateFiles returns the list of .tmpl files for a template from the local
// cache. Convention: any file in the template's cache directory ending in .tmpl.
func TemplateFiles(name string) ([]string, error) {
	dir := filepath.Join(cacheDir(), name)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".tmpl") {
			files = append(files, e.Name())
		}
	}
	return files, nil
}

// ReadTemplateFile returns the content of a cached template file.
// Downloads and caches it first if not already present.
func ReadTemplateFile(name, filename string) (string, error) {
	path := filepath.Join(cacheDir(), name, filename)

	data, err := os.ReadFile(path)
	if err == nil {
		return string(data), nil
	}

	url := rawURL(name + "/" + filename)
	fetched, err := download(url)
	if err != nil {
		return "", fmt.Errorf("downloading %s/%s: %w", name, filename, err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", err
	}
	if err := os.WriteFile(path, fetched, 0o644); err != nil {
		return "", err
	}
	return string(fetched), nil
}
