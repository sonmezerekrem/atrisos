package traefik

import (
	"strings"
	"testing"

	"github.com/sonmezerekrem/atrisos/internal/config"
)

func TestGenerateComposeYAML(t *testing.T) {
	m := NewManager(&config.Config{
		Traefik: config.TraefikConfig{
			ACMEEmail: "admin@example.com",
			Network:   "atrisos_net",
			HTTPPort:  8080,
			HTTPSPort: 8443,
			Image:     "docker.io/library/traefik:v3.3",
		},
	})

	yaml := m.generateComposeYAML()

	checks := []string{
		"docker.io/library/traefik:v3.3",
		"container_name: atrisos_traefik",
		"atrisos_net",
		"8080:8080",
		"8443:8443",
		"letsencrypt-staging",
		"${ACME_EMAIL}",
		"${PODMAN_SOCKET}",
	}
	for _, want := range checks {
		if !strings.Contains(yaml, want) {
			t.Errorf("compose YAML missing %q", want)
		}
	}
}
