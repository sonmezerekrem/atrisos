package registry

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Registry holds extra registered stack paths outside the root directory.
type Registry struct {
	ExtraPaths []string `json:"extra_paths"`
	path       string   // unexported: set on load
}

// Load reads the registry from the given config directory. If the file
// doesn't exist, it returns an empty registry without error.
func Load(cfgDir string) (*Registry, error) {
	path := filepath.Join(cfgDir, "registry.json")
	reg := &Registry{path: path}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return reg, nil
		}
		return nil, err
	}

	if err := json.Unmarshal(data, reg); err != nil {
		return nil, err
	}
	reg.path = path
	return reg, nil
}

// Save writes the registry to disk. Creates the file if it doesn't exist.
func (r *Registry) Save() error {
	if err := os.MkdirAll(filepath.Dir(r.path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(r.path, data, 0o644)
}

// Add registers an absolute path to the registry (idempotent).
func (r *Registry) Add(absPath string) {
	for _, p := range r.ExtraPaths {
		if p == absPath {
			return
		}
	}
	r.ExtraPaths = append(r.ExtraPaths, absPath)
}

// Remove removes an entry by directory base name. Matches the first entry
// whose filepath.Base equals the given name.
func (r *Registry) Remove(name string) {
	filtered := r.ExtraPaths[:0]
	for _, p := range r.ExtraPaths {
		if filepath.Base(p) != name {
			filtered = append(filtered, p)
		}
	}
	r.ExtraPaths = filtered
}

// Paths returns all registered extra paths.
func (r *Registry) Paths() []string {
	if r.ExtraPaths == nil {
		return []string{}
	}
	return r.ExtraPaths
}
