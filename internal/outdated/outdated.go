package outdated

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/sonmezerekrem/atrisos/internal/compose"
	"github.com/sonmezerekrem/atrisos/internal/stack"
)

// ImageUpdate describes an available update for a single service image.
type ImageUpdate struct {
	Service   string
	Image     string
	Current   string // local digest (short, 12 chars)
	Available string // remote digest (short, 12 chars)
}

// CheckStack checks all images used by a stack for available updates.
// Returns only entries where the remote digest differs from the local digest.
func CheckStack(s *stack.Stack) ([]ImageUpdate, error) {
	doc, err := compose.LoadAndMerge(s.Dir, s.Config)
	if err != nil {
		return nil, fmt.Errorf("loading compose for %s: %w", s.Name, err)
	}

	services := compose.GetServices(doc)
	if len(services) == 0 {
		return nil, nil
	}

	var updates []ImageUpdate
	for svcName, svcVal := range services {
		svc, ok := svcVal.(map[string]interface{})
		if !ok {
			continue
		}
		image, _ := svc["image"].(string)
		if image == "" {
			continue
		}

		local, err := localDigest(image)
		if err != nil {
			// Image not pulled locally — skip.
			continue
		}

		remote, err := remoteDigest(image)
		if err != nil {
			// Can't reach registry — skip.
			if verboseMode() {
				fmt.Fprintf(os.Stderr, "⚠ skipping %s/%s: %v\n", s.Name, svcName, err)
			}
			continue
		}

		if local != remote && remote != "" {
			updates = append(updates, ImageUpdate{
				Service:   svcName,
				Image:     image,
				Current:   local,
				Available: remote,
			})
		}
	}

	return updates, nil
}

// CheckAll checks all stacks and returns a map of stack name → updates.
func CheckAll(stacks []*stack.Stack) (map[string][]ImageUpdate, error) {
	result := make(map[string][]ImageUpdate)
	for _, s := range stacks {
		updates, err := CheckStack(s)
		if err != nil {
			fmt.Fprintf(os.Stderr, "⚠ %s: %v\n", s.Name, err)
			continue
		}
		if len(updates) > 0 {
			result[s.Name] = updates
		}
	}
	return result, nil
}

// localDigest returns the short digest of the locally pulled image.
func localDigest(image string) (string, error) {
	out, err := exec.Command("podman", "image", "inspect", image,
		"--format", "{{index .RepoDigests 0}}").Output()
	if err != nil {
		return "", fmt.Errorf("podman image inspect %s: %w", image, err)
	}
	return shortDigest(strings.TrimSpace(string(out))), nil
}

// remoteDigest returns the short digest from the remote registry via
// podman manifest inspect (no pull required).
func remoteDigest(image string) (string, error) {
	out, err := exec.Command("podman", "manifest", "inspect", image).Output()
	if err != nil {
		return "", fmt.Errorf("podman manifest inspect %s: %w", image, err)
	}
	// The manifest output is JSON; look for the first "digest" field.
	raw := string(out)
	const needle = `"digest"`
	idx := strings.Index(raw, needle)
	if idx < 0 {
		return "", fmt.Errorf("no digest in manifest for %s", image)
	}
	rest := raw[idx+len(needle):]
	// rest looks like: : "sha256:abc..."
	colonIdx := strings.Index(rest, `"`)
	if colonIdx < 0 {
		return "", fmt.Errorf("malformed digest in manifest for %s", image)
	}
	rest = rest[colonIdx+1:]
	endIdx := strings.Index(rest, `"`)
	if endIdx < 0 {
		return "", fmt.Errorf("malformed digest in manifest for %s", image)
	}
	return shortDigest(rest[:endIdx]), nil
}

// shortDigest returns the first 12 characters after "sha256:" in a digest
// string (e.g. "repo@sha256:abcdef..." → "abcdef...").
func shortDigest(digest string) string {
	const prefix = "sha256:"
	idx := strings.Index(digest, prefix)
	if idx < 0 {
		// Not a sha256 digest — return as-is truncated.
		if len(digest) > 12 {
			return digest[:12]
		}
		return digest
	}
	after := digest[idx+len(prefix):]
	if len(after) > 12 {
		return after[:12]
	}
	return after
}

func verboseMode() bool {
	return os.Getenv("ATRISOS_VERBOSE") != ""
}
