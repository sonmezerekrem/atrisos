package traefik

import (
	"crypto/sha256"
	"fmt"
	"strings"
	"testing"

	"github.com/sonmezerekrem/atrisos/internal/stack"
)

func expectedHash(absPath string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(absPath)))[:6]
}

func TestRouterName(t *testing.T) {
	t.Run("lowercases inputs", func(t *testing.T) {
		name := RouterName("MyStack", "Web", "/some/path", 0)
		if strings.ContainsAny(name, "ABCDEFGHIJKLMNOPQRSTUVWXYZ") {
			t.Errorf("RouterName returned uppercase chars: %q", name)
		}
	})

	t.Run("sanitizes non-alphanumeric chars to hyphens", func(t *testing.T) {
		name := RouterName("my stack", "my_service", "/some/path", 0)
		// Spaces and underscores should become hyphens.
		if strings.Contains(name, " ") || strings.Contains(name, "_") {
			t.Errorf("RouterName contains unsanitized chars: %q", name)
		}
	})

	t.Run("hash is exactly 6 hex characters", func(t *testing.T) {
		name := RouterName("stack", "svc", "/abs/path", 0)
		// Format: <sanitized-stackDir>-<sanitized-service>-<hash6>
		parts := strings.Split(name, "-")
		hash := parts[len(parts)-1]
		if len(hash) != 6 {
			t.Errorf("hash part %q has length %d, want 6", hash, len(hash))
		}
		for _, c := range hash {
			if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
				t.Errorf("hash part %q contains non-hex char %q", hash, c)
			}
		}
	})

	t.Run("same inputs always produce the same output", func(t *testing.T) {
		a := RouterName("stack", "web", "/abs/path", 0)
		b := RouterName("stack", "web", "/abs/path", 0)
		if a != b {
			t.Errorf("got different names: %q vs %q", a, b)
		}
	})

	t.Run("different stack paths produce different hashes", func(t *testing.T) {
		a := RouterName("stack", "web", "/path/one", 0)
		b := RouterName("stack", "web", "/path/two", 0)
		if a == b {
			t.Errorf("expected different names for different paths, got %q for both", a)
		}
	})

	t.Run("idx greater than zero appends index suffix", func(t *testing.T) {
		base := RouterName("stack", "web", "/abs/path", 0)
		with1 := RouterName("stack", "web", "/abs/path", 1)
		want := base + "-1"
		if with1 != want {
			t.Errorf("got %q, want %q", with1, want)
		}
	})

	t.Run("contains sanitized stackDir and service", func(t *testing.T) {
		name := RouterName("My Stack", "My Service", "/abs/path", 0)
		hash := expectedHash("/abs/path")
		want := "my-stack-my-service-" + hash
		if name != want {
			t.Errorf("got %q, want %q", name, want)
		}
	})
}

func TestGenerateLabels(t *testing.T) {
	const stackDir = "mystack"
	const stackAbsPath = "/stacks/mystack"
	const host = "example.com"
	const svcName = "web"

	routerName := RouterName(stackDir, svcName, stackAbsPath, 0)

	t.Run("TLS true has websecure entrypoint and letsencrypt certresolver", func(t *testing.T) {
		d := stack.DomainConfig{
			Service: svcName,
			Host:    host,
			Port:    8080,
			TLS:     "true",
		}
		labels := GenerateLabels(d, stackDir, stackAbsPath, 0)

		if labels["traefik.enable"] != "true" {
			t.Errorf("traefik.enable = %q, want true", labels["traefik.enable"])
		}

		entryKey := fmt.Sprintf("traefik.http.routers.%s.entrypoints", routerName)
		if labels[entryKey] != "websecure" {
			t.Errorf("%s = %q, want websecure", entryKey, labels[entryKey])
		}

		certKey := fmt.Sprintf("traefik.http.routers.%s.tls.certresolver", routerName)
		if labels[certKey] != "letsencrypt" {
			t.Errorf("%s = %q, want letsencrypt", certKey, labels[certKey])
		}

		tlsKey := fmt.Sprintf("traefik.http.routers.%s.tls", routerName)
		if labels[tlsKey] != "true" {
			t.Errorf("%s = %q, want true", tlsKey, labels[tlsKey])
		}

		// HTTP redirect router should be present.
		httpEntryKey := fmt.Sprintf("traefik.http.routers.%s-http.entrypoints", routerName)
		if labels[httpEntryKey] != "web" {
			t.Errorf("%s = %q, want web", httpEntryKey, labels[httpEntryKey])
		}
	})

	t.Run("TLS false has web entrypoint and no tls labels", func(t *testing.T) {
		d := stack.DomainConfig{
			Service: svcName,
			Host:    host,
			Port:    8080,
			TLS:     "false",
		}
		labels := GenerateLabels(d, stackDir, stackAbsPath, 0)

		entryKey := fmt.Sprintf("traefik.http.routers.%s.entrypoints", routerName)
		if labels[entryKey] != "web" {
			t.Errorf("%s = %q, want web", entryKey, labels[entryKey])
		}

		tlsKey := fmt.Sprintf("traefik.http.routers.%s.tls", routerName)
		if _, ok := labels[tlsKey]; ok {
			t.Errorf("tls label should not be present for TLS=false, got %q", labels[tlsKey])
		}

		certKey := fmt.Sprintf("traefik.http.routers.%s.tls.certresolver", routerName)
		if _, ok := labels[certKey]; ok {
			t.Errorf("certresolver label should not be present for TLS=false")
		}

		// No HTTP redirect router for TLS=false.
		httpMwKey := fmt.Sprintf("traefik.http.routers.%s-http.middlewares", routerName)
		if _, ok := labels[httpMwKey]; ok {
			t.Errorf("http redirect router should not be present for TLS=false")
		}
	})

	t.Run("TLS staging uses letsencrypt-staging certresolver", func(t *testing.T) {
		d := stack.DomainConfig{
			Service: svcName,
			Host:    host,
			Port:    8080,
			TLS:     "staging",
		}
		labels := GenerateLabels(d, stackDir, stackAbsPath, 0)

		certKey := fmt.Sprintf("traefik.http.routers.%s.tls.certresolver", routerName)
		if labels[certKey] != "letsencrypt-staging" {
			t.Errorf("%s = %q, want letsencrypt-staging", certKey, labels[certKey])
		}

		entryKey := fmt.Sprintf("traefik.http.routers.%s.entrypoints", routerName)
		if labels[entryKey] != "websecure" {
			t.Errorf("%s = %q, want websecure", entryKey, labels[entryKey])
		}
	})

	t.Run("custom port sets loadbalancer server port label", func(t *testing.T) {
		d := stack.DomainConfig{
			Service: svcName,
			Host:    host,
			Port:    3000,
			TLS:     "false",
		}
		labels := GenerateLabels(d, stackDir, stackAbsPath, 0)

		portKey := fmt.Sprintf("traefik.http.services.%s.loadbalancer.server.port", routerName)
		if labels[portKey] != "3000" {
			t.Errorf("%s = %q, want 3000", portKey, labels[portKey])
		}
	})

	t.Run("path prefix set includes PathPrefix in rule", func(t *testing.T) {
		d := stack.DomainConfig{
			Service:    svcName,
			Host:       host,
			Port:       8080,
			TLS:        "false",
			PathPrefix: "/api",
		}
		labels := GenerateLabels(d, stackDir, stackAbsPath, 0)

		ruleKey := fmt.Sprintf("traefik.http.routers.%s.rule", routerName)
		if !strings.Contains(labels[ruleKey], "PathPrefix") {
			t.Errorf("rule %q does not contain PathPrefix", labels[ruleKey])
		}
		if !strings.Contains(labels[ruleKey], "/api") {
			t.Errorf("rule %q does not contain /api", labels[ruleKey])
		}
	})

	t.Run("no path prefix uses Host-only rule", func(t *testing.T) {
		d := stack.DomainConfig{
			Service: svcName,
			Host:    host,
			Port:    8080,
			TLS:     "false",
		}
		labels := GenerateLabels(d, stackDir, stackAbsPath, 0)

		ruleKey := fmt.Sprintf("traefik.http.routers.%s.rule", routerName)
		if strings.Contains(labels[ruleKey], "PathPrefix") {
			t.Errorf("rule %q should not contain PathPrefix when path_prefix is empty", labels[ruleKey])
		}
		if !strings.Contains(labels[ruleKey], host) {
			t.Errorf("rule %q does not contain host %q", labels[ruleKey], host)
		}
	})

	t.Run("user middlewares appended to router middlewares", func(t *testing.T) {
		d := stack.DomainConfig{
			Service:     svcName,
			Host:        host,
			Port:        8080,
			TLS:         "false",
			Middlewares: []string{"my-auth", "rate-limit"},
		}
		labels := GenerateLabels(d, stackDir, stackAbsPath, 0)

		mwKey := fmt.Sprintf("traefik.http.routers.%s.middlewares", routerName)
		mw := labels[mwKey]
		if !strings.Contains(mw, "my-auth") {
			t.Errorf("middlewares %q should contain my-auth", mw)
		}
		if !strings.Contains(mw, "rate-limit") {
			t.Errorf("middlewares %q should contain rate-limit", mw)
		}
	})

	t.Run("traefik.enable is always true", func(t *testing.T) {
		d := stack.DomainConfig{Service: svcName, Host: host, Port: 80, TLS: "false"}
		labels := GenerateLabels(d, stackDir, stackAbsPath, 0)
		if labels["traefik.enable"] != "true" {
			t.Errorf("traefik.enable = %q, want true", labels["traefik.enable"])
		}
	})
}
