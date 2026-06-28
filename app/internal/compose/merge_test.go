package compose

import (
	"fmt"
	"strings"
	"testing"

	"github.com/sonmezerekrem/atrisos/internal/stack"
	"github.com/sonmezerekrem/atrisos/internal/traefik"
)

// --- MergeOverride tests ---

func TestMergeOverride(t *testing.T) {
	t.Run("base service A plus override service B yields both", func(t *testing.T) {
		base := ComposeDoc{
			"services": map[string]interface{}{
				"a": map[string]interface{}{"image": "alpine"},
			},
		}
		override := ComposeDoc{
			"services": map[string]interface{}{
				"b": map[string]interface{}{"image": "nginx"},
			},
		}
		result := MergeOverride(base, override)
		services, ok := result["services"].(map[string]interface{})
		if !ok {
			t.Fatal("services is not a map")
		}
		if _, ok := services["a"]; !ok {
			t.Error("service 'a' missing from result")
		}
		if _, ok := services["b"]; !ok {
			t.Error("service 'b' missing from result")
		}
	})

	t.Run("override replaces a key in service A", func(t *testing.T) {
		base := ComposeDoc{
			"services": map[string]interface{}{
				"a": map[string]interface{}{"image": "old-image"},
			},
		}
		override := ComposeDoc{
			"services": map[string]interface{}{
				"a": map[string]interface{}{"image": "new-image"},
			},
		}
		result := MergeOverride(base, override)
		services := result["services"].(map[string]interface{})
		svc := services["a"].(map[string]interface{})
		if svc["image"] != "new-image" {
			t.Errorf("image = %q, want new-image", svc["image"])
		}
	})

	t.Run("nested map is deep-merged not replaced", func(t *testing.T) {
		base := ComposeDoc{
			"services": map[string]interface{}{
				"a": map[string]interface{}{
					"environment": map[string]interface{}{
						"FOO": "base-foo",
						"BAR": "base-bar",
					},
				},
			},
		}
		override := ComposeDoc{
			"services": map[string]interface{}{
				"a": map[string]interface{}{
					"environment": map[string]interface{}{
						"FOO": "override-foo",
					},
				},
			},
		}
		result := MergeOverride(base, override)
		services := result["services"].(map[string]interface{})
		svc := services["a"].(map[string]interface{})
		env := svc["environment"].(map[string]interface{})

		if env["FOO"] != "override-foo" {
			t.Errorf("FOO = %q, want override-foo", env["FOO"])
		}
		// BAR should still be present from base.
		if env["BAR"] != "base-bar" {
			t.Errorf("BAR = %q, want base-bar (should survive deep merge)", env["BAR"])
		}
	})

	t.Run("nil override returns base unchanged", func(t *testing.T) {
		base := ComposeDoc{
			"services": map[string]interface{}{
				"a": map[string]interface{}{"image": "alpine"},
			},
		}
		result := MergeOverride(base, nil)
		services, ok := result["services"].(map[string]interface{})
		if !ok {
			t.Fatal("services is not a map")
		}
		svc, ok := services["a"].(map[string]interface{})
		if !ok {
			t.Fatal("service 'a' missing or wrong type")
		}
		if svc["image"] != "alpine" {
			t.Errorf("image = %q, want alpine", svc["image"])
		}
	})

	t.Run("base is not mutated by MergeOverride", func(t *testing.T) {
		base := ComposeDoc{
			"services": map[string]interface{}{
				"a": map[string]interface{}{"image": "original"},
			},
		}
		override := ComposeDoc{
			"services": map[string]interface{}{
				"a": map[string]interface{}{"image": "changed"},
			},
		}
		MergeOverride(base, override)
		// Base should not be mutated.
		svc := base["services"].(map[string]interface{})["a"].(map[string]interface{})
		if svc["image"] != "original" {
			t.Errorf("base was mutated: image = %q, want original", svc["image"])
		}
	})
}

// --- Merge tests ---

const testStackPath = "/stacks/myapp"

func buildDoc(svcNames ...string) ComposeDoc {
	services := map[string]interface{}{}
	for _, name := range svcNames {
		services[name] = map[string]interface{}{}
	}
	return ComposeDoc{"services": services}
}

func getServiceLabels(t *testing.T, doc ComposeDoc, svcName string) map[string]interface{} {
	t.Helper()
	services, ok := doc["services"].(map[string]interface{})
	if !ok {
		t.Fatalf("doc.services is not a map")
	}
	svc, ok := services[svcName].(map[string]interface{})
	if !ok {
		t.Fatalf("service %q not found or wrong type", svcName)
	}
	labels, ok := svc["labels"].(map[string]interface{})
	if !ok {
		t.Fatalf("service %q has no labels map", svcName)
	}
	return labels
}

func TestMerge(t *testing.T) {
	stackDir := "myapp" // filepath.Base(testStackPath)

	t.Run("router name contains stack dir, service name and 6-char hex hash", func(t *testing.T) {
		doc := buildDoc("web")
		cfg := stack.StackConfig{
			Domains: []stack.DomainConfig{
				{Service: "web", Host: "example.com", Port: 80, TLS: "false"},
			},
		}
		result := Merge(doc, cfg, testStackPath)
		labels := getServiceLabels(t, result, "web")

		rn := traefik.RouterName(stackDir, "web", testStackPath, 0)
		ruleKey := fmt.Sprintf("traefik.http.routers.%s.rule", rn)
		if _, ok := labels[ruleKey]; !ok {
			t.Errorf("expected label %q to be present", ruleKey)
		}

		// Router name must contain stack dir, service name, and a 6-char hex hash.
		if !strings.Contains(rn, stackDir) {
			t.Errorf("router name %q does not contain stackDir %q", rn, stackDir)
		}
		if !strings.Contains(rn, "web") {
			t.Errorf("router name %q does not contain service name web", rn)
		}
		parts := strings.Split(rn, "-")
		hash := parts[len(parts)-1]
		if len(hash) != 6 {
			t.Errorf("hash part %q has length %d, want 6", hash, len(hash))
		}
	})

	t.Run("traefik.enable=true label present on service", func(t *testing.T) {
		doc := buildDoc("web")
		cfg := stack.StackConfig{
			Domains: []stack.DomainConfig{
				{Service: "web", Host: "example.com", Port: 80, TLS: "false"},
			},
		}
		result := Merge(doc, cfg, testStackPath)
		labels := getServiceLabels(t, result, "web")

		if labels["traefik.enable"] != "true" {
			t.Errorf("traefik.enable = %v, want true", labels["traefik.enable"])
		}
	})

	t.Run("atrisos_net added to service networks", func(t *testing.T) {
		doc := buildDoc("web")
		cfg := stack.StackConfig{
			Domains: []stack.DomainConfig{
				{Service: "web", Host: "example.com", Port: 80, TLS: "false"},
			},
		}
		result := Merge(doc, cfg, testStackPath)
		services := result["services"].(map[string]interface{})
		svc := services["web"].(map[string]interface{})

		found := false
		switch nets := svc["networks"].(type) {
		case []interface{}:
			for _, n := range nets {
				if n == "atrisos_net" {
					found = true
				}
			}
		case map[string]interface{}:
			_, found = nets["atrisos_net"]
		}
		if !found {
			t.Errorf("atrisos_net not found in service networks: %v", svc["networks"])
		}
	})

	t.Run("atrisos_net declared as external in top-level networks", func(t *testing.T) {
		doc := buildDoc("web")
		cfg := stack.StackConfig{
			Domains: []stack.DomainConfig{
				{Service: "web", Host: "example.com", Port: 80, TLS: "false"},
			},
		}
		result := Merge(doc, cfg, testStackPath)
		topNets, ok := result["networks"].(map[string]interface{})
		if !ok {
			t.Fatal("top-level networks is not a map")
		}
		net, ok := topNets["atrisos_net"]
		if !ok {
			t.Fatal("atrisos_net not declared at top-level networks")
		}
		netMap, ok := net.(map[string]interface{})
		if !ok {
			t.Fatalf("atrisos_net network entry is not a map: %v", net)
		}
		if netMap["external"] != true {
			t.Errorf("atrisos_net external = %v, want true", netMap["external"])
		}
	})

	t.Run("TLS true sets websecure entrypoint", func(t *testing.T) {
		doc := buildDoc("web")
		cfg := stack.StackConfig{
			Domains: []stack.DomainConfig{
				{Service: "web", Host: "example.com", Port: 443, TLS: "true"},
			},
		}
		result := Merge(doc, cfg, testStackPath)
		labels := getServiceLabels(t, result, "web")

		rn := traefik.RouterName(stackDir, "web", testStackPath, 0)
		entryKey := fmt.Sprintf("traefik.http.routers.%s.entrypoints", rn)
		if labels[entryKey] != "websecure" {
			t.Errorf("%s = %v, want websecure", entryKey, labels[entryKey])
		}
	})

	t.Run("TLS false sets web entrypoint and no tls label", func(t *testing.T) {
		doc := buildDoc("web")
		cfg := stack.StackConfig{
			Domains: []stack.DomainConfig{
				{Service: "web", Host: "example.com", Port: 80, TLS: "false"},
			},
		}
		result := Merge(doc, cfg, testStackPath)
		labels := getServiceLabels(t, result, "web")

		rn := traefik.RouterName(stackDir, "web", testStackPath, 0)
		entryKey := fmt.Sprintf("traefik.http.routers.%s.entrypoints", rn)
		if labels[entryKey] != "web" {
			t.Errorf("%s = %v, want web", entryKey, labels[entryKey])
		}

		tlsKey := fmt.Sprintf("traefik.http.routers.%s.tls", rn)
		if _, ok := labels[tlsKey]; ok {
			t.Errorf("tls label should not be present for TLS=false, got %v", labels[tlsKey])
		}
	})

	t.Run("TLS staging sets letsencrypt-staging certresolver", func(t *testing.T) {
		doc := buildDoc("web")
		cfg := stack.StackConfig{
			Domains: []stack.DomainConfig{
				{Service: "web", Host: "example.com", Port: 443, TLS: "staging"},
			},
		}
		result := Merge(doc, cfg, testStackPath)
		labels := getServiceLabels(t, result, "web")

		rn := traefik.RouterName(stackDir, "web", testStackPath, 0)
		certKey := fmt.Sprintf("traefik.http.routers.%s.tls.certresolver", rn)
		if labels[certKey] != "letsencrypt-staging" {
			t.Errorf("%s = %v, want letsencrypt-staging", certKey, labels[certKey])
		}
	})

	t.Run("path_prefix set produces PathPrefix rule", func(t *testing.T) {
		doc := buildDoc("api")
		cfg := stack.StackConfig{
			Domains: []stack.DomainConfig{
				{Service: "api", Host: "example.com", Port: 8080, TLS: "false", PathPrefix: "/api"},
			},
		}
		result := Merge(doc, cfg, testStackPath)
		labels := getServiceLabels(t, result, "api")

		rn := traefik.RouterName(stackDir, "api", testStackPath, 0)
		ruleKey := fmt.Sprintf("traefik.http.routers.%s.rule", rn)
		rule, _ := labels[ruleKey].(string)
		if !strings.Contains(rule, "PathPrefix") {
			t.Errorf("rule %q does not contain PathPrefix", rule)
		}
	})

	t.Run("service not in compose does not crash and gets created", func(t *testing.T) {
		doc := buildDoc("web")
		cfg := stack.StackConfig{
			Domains: []stack.DomainConfig{
				{Service: "missing-svc", Host: "example.com", Port: 80, TLS: "false"},
			},
		}
		// Should not panic.
		result := Merge(doc, cfg, testStackPath)

		// The missing service should now exist in the document.
		services, ok := result["services"].(map[string]interface{})
		if !ok {
			t.Fatal("services is not a map")
		}
		if _, ok := services["missing-svc"]; !ok {
			t.Error("missing-svc should have been created in services map")
		}
		// Original service should still be there.
		if _, ok := services["web"]; !ok {
			t.Error("original service 'web' should still be present")
		}
	})

	t.Run("no domains leaves document unchanged except for selinux", func(t *testing.T) {
		doc := buildDoc("web")
		cfg := stack.StackConfig{Domains: nil}
		result := Merge(doc, cfg, testStackPath)

		// No labels or networks should have been added.
		services := result["services"].(map[string]interface{})
		svc := services["web"].(map[string]interface{})
		if _, ok := svc["labels"]; ok {
			t.Error("labels should not be added when there are no domains")
		}
		if _, ok := svc["networks"]; ok {
			t.Error("networks should not be added when there are no domains")
		}
	})
}
