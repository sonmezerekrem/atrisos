package compose

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/sonmezerekrem/atrisos/internal/stack"
	"github.com/sonmezerekrem/atrisos/internal/traefik"
	"gopkg.in/yaml.v3"
)

// Merge injects Traefik labels and the atrisos network into the compose
// document for each domain entry in cfg. The original file is never modified.
func Merge(doc ComposeDoc, cfg stack.StackConfig, stackPath string) ComposeDoc {
	if len(cfg.Domains) == 0 {
		return doc
	}

	networkName := "atrisos_net"
	stackDir := filepath.Base(stackPath)

	// Ensure top-level services map exists.
	if doc["services"] == nil {
		doc["services"] = map[string]interface{}{}
	}
	services, ok := doc["services"].(map[string]interface{})
	if !ok {
		services = map[string]interface{}{}
		doc["services"] = services
	}

	// Track how many times each service appears in the domains list.
	// This handles the case where two domains point to the same service.
	serviceCount := map[string]int{}

	for _, d := range cfg.Domains {
		idx := serviceCount[d.Service]
		serviceCount[d.Service]++

		// Generate Traefik labels for this domain entry.
		labels := traefik.GenerateLabels(d, stackDir, stackPath, idx)

		// Get or create the service map.
		var svc map[string]interface{}
		if existing, ok := services[d.Service].(map[string]interface{}); ok {
			svc = existing
		} else {
			svc = map[string]interface{}{}
			services[d.Service] = svc
		}

		// Merge labels into service (preserve existing labels).
		var existingLabels map[string]interface{}
		switch v := svc["labels"].(type) {
		case map[string]interface{}:
			existingLabels = v
		case map[interface{}]interface{}:
			// YAML sometimes gives us this type.
			existingLabels = make(map[string]interface{}, len(v))
			for k, val := range v {
				existingLabels = appendLabel(existingLabels, fmt.Sprintf("%v", k), fmt.Sprintf("%v", val))
			}
		default:
			existingLabels = map[string]interface{}{}
		}
		for k, v := range labels {
			existingLabels[k] = v
		}
		svc["labels"] = existingLabels

		// Add atrisos_net to service networks (preserve existing).
		svc["networks"] = mergeNetworkList(svc["networks"], networkName)
	}

	// Add atrisos_net as an external network at the top level.
	if doc["networks"] == nil {
		doc["networks"] = map[string]interface{}{}
	}
	topNetworks, ok := doc["networks"].(map[string]interface{})
	if !ok {
		topNetworks = map[string]interface{}{}
		doc["networks"] = topNetworks
	}
	if _, exists := topNetworks[networkName]; !exists {
		topNetworks[networkName] = map[string]interface{}{"external": true}
	}

	return doc
}

// appendLabel is a helper used only during label conversion.
func appendLabel(m map[string]interface{}, k, v string) map[string]interface{} {
	m[k] = v
	return m
}

// mergeNetworkList ensures networkName is in the service's networks list
// while preserving all existing entries.
func mergeNetworkList(existing interface{}, networkName string) interface{} {
	// Build set of existing network names.
	existingNames := map[string]bool{}
	switch v := existing.(type) {
	case []interface{}:
		for _, n := range v {
			if s, ok := n.(string); ok {
				existingNames[s] = true
			}
		}
	case map[string]interface{}:
		// Networks can be a map keyed by network name.
		result := make(map[string]interface{}, len(v))
		for k, val := range v {
			result[k] = val
		}
		if _, ok := result[networkName]; !ok {
			result[networkName] = nil
		}
		return result
	case nil:
		// No existing networks — add "default" + atrisos_net.
		return []interface{}{"default", networkName}
	}

	if !existingNames[networkName] {
		switch v := existing.(type) {
		case []interface{}:
			return append(v, networkName)
		}
	}
	return existing
}

// MergeOverride deep-merges the override document on top of the base document.
// Maps are merged recursively; slices are replaced (not appended) — same
// semantics as docker compose override files.
func MergeOverride(base, override ComposeDoc) ComposeDoc {
	result := deepCopyMap(base)
	deepMerge(result, override)
	return result
}

// deepMerge merges src into dst in place.
func deepMerge(dst, src map[string]interface{}) {
	for k, srcVal := range src {
		if dstVal, exists := dst[k]; exists {
			// Both maps: recurse.
			dstMap, dstIsMap := dstVal.(map[string]interface{})
			srcMap, srcIsMap := srcVal.(map[string]interface{})
			if dstIsMap && srcIsMap {
				deepMerge(dstMap, srcMap)
				continue
			}
		}
		// Otherwise: override (includes slice replacement).
		dst[k] = srcVal
	}
}

// deepCopyMap creates a shallow-deep copy of a map.
func deepCopyMap(m map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{}, len(m))
	for k, v := range m {
		switch val := v.(type) {
		case map[string]interface{}:
			result[k] = deepCopyMap(val)
		default:
			result[k] = v
		}
	}
	return result
}

// LoadAndMerge reads the compose file from stackDir, optionally applies
// compose.override.yml, then injects Traefik labels and networks.
func LoadAndMerge(stackDir string, cfg stack.StackConfig) (ComposeDoc, error) {
	// Determine compose file path.
	composePath := stack.ComposeFile(stackDir)
	if composePath == "" {
		return nil, fmt.Errorf("no compose.yml or docker-compose.yml found in %s", stackDir)
	}

	// Parse compose.yml.
	data, err := os.ReadFile(composePath)
	if err != nil {
		return nil, fmt.Errorf("reading compose file: %w", err)
	}
	var doc ComposeDoc
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("parsing compose file: %w", err)
	}
	if doc == nil {
		doc = ComposeDoc{}
	}

	// Check for compose.override.yml.
	overridePath := filepath.Join(stackDir, "compose.override.yml")
	if overrideData, err := os.ReadFile(overridePath); err == nil {
		var overrideDoc ComposeDoc
		if err := yaml.Unmarshal(overrideData, &overrideDoc); err != nil {
			return nil, fmt.Errorf("parsing compose.override.yml: %w", err)
		}
		if overrideDoc != nil {
			doc = MergeOverride(doc, overrideDoc)
		}
	}

	// Inject Traefik labels and networks.
	absDir, err := filepath.Abs(stackDir)
	if err != nil {
		return nil, err
	}
	doc = Merge(doc, cfg, absDir)

	return doc, nil
}

// WriteToTemp serializes the compose document to a temp file and returns
// its path. The caller is responsible for deleting the file.
func WriteToTemp(doc ComposeDoc) (string, error) {
	f, err := os.CreateTemp("", "atrisos-compose-*.yml")
	if err != nil {
		return "", fmt.Errorf("creating temp compose file: %w", err)
	}
	defer f.Close()

	if err := yaml.NewEncoder(f).Encode(doc); err != nil {
		os.Remove(f.Name())
		return "", fmt.Errorf("writing temp compose file: %w", err)
	}

	return f.Name(), nil
}
