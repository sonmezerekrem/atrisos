package stack

import (
	"os"
	"path/filepath"
	"sort"

	"github.com/sonmezerekrem/atrisos/internal/config"
	"github.com/sonmezerekrem/atrisos/internal/registry"
)

// Discover finds all stacks from the configured root directory and registered
// extra paths. Stacks are deduplicated by absolute path and sorted by name.
func Discover(cfg *config.Config, reg *registry.Registry) ([]*Stack, error) {
	seen := map[string]bool{}
	var candidates []string

	// Walk stacks root one level deep.
	root := cfg.StacksRoot
	entries, err := os.ReadDir(root)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		absPath := filepath.Join(root, e.Name())
		if !seen[absPath] && IsStack(absPath) {
			seen[absPath] = true
			candidates = append(candidates, absPath)
		}
	}

	// Merge in registered extra paths (each is a direct stack dir).
	for _, p := range reg.Paths() {
		absPath, err := filepath.Abs(p)
		if err != nil {
			continue
		}
		if !seen[absPath] && IsStack(absPath) {
			seen[absPath] = true
			candidates = append(candidates, absPath)
		}
	}

	// Load each candidate.
	var stacks []*Stack
	for _, dir := range candidates {
		s, err := LoadStack(dir)
		if err != nil {
			// Skip directories that fail to load rather than aborting discovery.
			continue
		}
		stacks = append(stacks, s)
	}

	// Sort by name.
	sort.Slice(stacks, func(i, j int) bool {
		return stacks[i].Name < stacks[j].Name
	})

	return stacks, nil
}

// FilterByTag returns the subset of stacks whose Config.Tags contains the
// given tag.
func FilterByTag(stacks []*Stack, tag string) []*Stack {
	var result []*Stack
	for _, s := range stacks {
		for _, t := range s.Config.Tags {
			if t == tag {
				result = append(result, s)
				break
			}
		}
	}
	return result
}

// FindByName returns the first stack whose name matches the given name,
// searching discovered stacks. Returns nil if not found.
func FindByName(stacks []*Stack, name string) *Stack {
	for _, s := range stacks {
		if s.Name == name {
			return s
		}
	}
	return nil
}
