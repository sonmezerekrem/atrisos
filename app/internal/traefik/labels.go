package traefik

import (
	"crypto/sha256"
	"fmt"
	"regexp"
	"strings"

	"github.com/sonmezerekrem/atrisos/app/internal/stack"
)

var nonAlphanumRE = regexp.MustCompile(`[^a-z0-9]+`)

// sanitize lowercases s and replaces runs of non-alphanumeric characters
// with a single hyphen.
func sanitize(s string) string {
	s = strings.ToLower(s)
	s = nonAlphanumRE.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	return s
}

// RouterName returns the Traefik router name for a service in a stack.
// Format: <sanitized-stackDir>-<sanitized-service>-<hash6>
// When idx > 0 (multiple domains → same service), appends -<idx>.
func RouterName(stackDir, service, stackAbsPath string, idx int) string {
	hash := fmt.Sprintf("%x", sha256.Sum256([]byte(stackAbsPath)))[:6]
	base := fmt.Sprintf("%s-%s-%s", sanitize(stackDir), sanitize(service), hash)
	if idx > 0 {
		return fmt.Sprintf("%s-%d", base, idx)
	}
	return base
}

// GenerateLabels generates the full set of Traefik labels for a domain entry.
//
// idx is used to differentiate multiple domain entries pointing to the same
// service (0 = first entry, 1 = second, etc.).
func GenerateLabels(d stack.DomainConfig, stackDir, stackAbsPath string, idx int) map[string]string {
	name := RouterName(stackDir, d.Service, stackAbsPath, idx)
	labels := map[string]string{}

	labels["traefik.enable"] = "true"

	// Compute the router rule.
	pathPrefix := d.PathPrefix
	if pathPrefix == "" {
		pathPrefix = "/"
	}
	var rule string
	if pathPrefix == "/" {
		rule = fmt.Sprintf("Host(`%s`)", d.Host)
	} else {
		rule = fmt.Sprintf("Host(`%s`) && PathPrefix(`%s`)", d.Host, pathPrefix)
	}

	// Service loadbalancer port (shared across TLS modes).
	labels[fmt.Sprintf("traefik.http.services.%s.loadbalancer.server.port", name)] = fmt.Sprintf("%d", d.Port)

	tls := d.TLS
	if tls == "" {
		tls = "true"
	}

	switch tls {
	case "false":
		// HTTP only — single router.
		labels[fmt.Sprintf("traefik.http.routers.%s.rule", name)] = rule
		labels[fmt.Sprintf("traefik.http.routers.%s.entrypoints", name)] = "web"

	case "staging":
		// HTTPS with staging resolver + HTTP→HTTPS redirect.
		labels[fmt.Sprintf("traefik.http.routers.%s.rule", name)] = rule
		labels[fmt.Sprintf("traefik.http.routers.%s.entrypoints", name)] = "websecure"
		labels[fmt.Sprintf("traefik.http.routers.%s.tls", name)] = "true"
		labels[fmt.Sprintf("traefik.http.routers.%s.tls.certresolver", name)] = "letsencrypt-staging"

		// HTTP → HTTPS redirect router.
		labels[fmt.Sprintf("traefik.http.routers.%s-http.rule", name)] = rule
		labels[fmt.Sprintf("traefik.http.routers.%s-http.entrypoints", name)] = "web"
		labels[fmt.Sprintf("traefik.http.routers.%s-http.middlewares", name)] = "atrisos-https-redirect"

		// HTTPS redirect middleware.
		labels["traefik.http.middlewares.atrisos-https-redirect.redirectscheme.scheme"] = "https"
		labels["traefik.http.middlewares.atrisos-https-redirect.redirectscheme.permanent"] = "true"

	default: // "true" (production)
		// HTTPS with production resolver + HTTP→HTTPS redirect.
		labels[fmt.Sprintf("traefik.http.routers.%s.rule", name)] = rule
		labels[fmt.Sprintf("traefik.http.routers.%s.entrypoints", name)] = "websecure"
		labels[fmt.Sprintf("traefik.http.routers.%s.tls", name)] = "true"
		labels[fmt.Sprintf("traefik.http.routers.%s.tls.certresolver", name)] = "letsencrypt"

		// HTTP → HTTPS redirect router.
		labels[fmt.Sprintf("traefik.http.routers.%s-http.rule", name)] = rule
		labels[fmt.Sprintf("traefik.http.routers.%s-http.entrypoints", name)] = "web"
		labels[fmt.Sprintf("traefik.http.routers.%s-http.middlewares", name)] = "atrisos-https-redirect"

		// HTTPS redirect middleware.
		labels["traefik.http.middlewares.atrisos-https-redirect.redirectscheme.scheme"] = "https"
		labels["traefik.http.middlewares.atrisos-https-redirect.redirectscheme.permanent"] = "true"
	}

	// Append user-configured middlewares to the HTTPS router (or HTTP if tls=false).
	if len(d.Middlewares) > 0 {
		extra := strings.Join(d.Middlewares, ",")
		routerKey := fmt.Sprintf("traefik.http.routers.%s.middlewares", name)
		if existing, ok := labels[routerKey]; ok && existing != "" {
			labels[routerKey] = existing + "," + extra
		} else {
			labels[routerKey] = extra
		}
	}

	return labels
}
