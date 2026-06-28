package templates

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setupCache(t *testing.T) string {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "atrisos")
	t.Setenv("XDG_CONFIG_HOME", dir)
	return filepath.Join(dir, "atrisos", "templates-cache")
}

func TestReadLocalManifest(t *testing.T) {
	cache := setupCache(t)
	if err := os.MkdirAll(cache, 0o755); err != nil {
		t.Fatal(err)
	}
	manifest := `{"version":"2026-01-01T00:00:00Z","templates":[{"name":"basic","display":"Basic","description":"test"}]}`
	if err := os.WriteFile(filepath.Join(cache, "manifest.json"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}

	m, err := readLocalManifest()
	if err != nil {
		t.Fatal(err)
	}
	if m.Version != "2026-01-01T00:00:00Z" {
		t.Errorf("version = %q", m.Version)
	}
	if len(m.Templates) != 1 || m.Templates[0].Name != "basic" {
		t.Errorf("templates = %+v", m.Templates)
	}
}

func TestTemplateFiles(t *testing.T) {
	cache := setupCache(t)
	tmplDir := filepath.Join(cache, "postgres")
	if err := os.MkdirAll(tmplDir, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"compose.yml.tmpl", "config.yml.tmpl", "template.yml"} {
		if err := os.WriteFile(filepath.Join(tmplDir, name), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	files, err := TemplateFiles("postgres")
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 2 {
		t.Fatalf("got %v, want 2 .tmpl files", files)
	}
}

func TestLoadTemplateMetaFromCache(t *testing.T) {
	cache := setupCache(t)
	tmplDir := filepath.Join(cache, "basic")
	if err := os.MkdirAll(tmplDir, 0o755); err != nil {
		t.Fatal(err)
	}
	content := "name: basic\ndisplay: Basic\ndescription: test\nprompts: []\n"
	if err := os.WriteFile(filepath.Join(tmplDir, "template.yml"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	meta, err := LoadTemplateMeta("basic")
	if err != nil {
		t.Fatal(err)
	}
	if meta.Name != "basic" || meta.Display != "Basic" {
		t.Errorf("meta = %+v", meta)
	}
}

func TestManifestURLs(t *testing.T) {
	if !strings.Contains(manifestURL(), "sonmezerekrem/atrisos/main/templates/manifest.json") {
		t.Errorf("unexpected manifestURL: %s", manifestURL())
	}
	if !strings.Contains(rawURL("postgres/template.yml"), "templates/postgres/template.yml") {
		t.Errorf("unexpected rawURL: %s", rawURL("postgres/template.yml"))
	}
}
