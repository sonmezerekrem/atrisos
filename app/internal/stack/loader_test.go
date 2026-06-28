package stack

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("writeFile %s: %v", path, err)
	}
}

func TestComposeFile(t *testing.T) {
	t.Run("both present returns compose.yml", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, filepath.Join(dir, "compose.yml"), "")
		writeFile(t, filepath.Join(dir, "docker-compose.yml"), "")

		got := ComposeFile(dir)
		want := filepath.Join(dir, "compose.yml")
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("only docker-compose.yml returns it", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, filepath.Join(dir, "docker-compose.yml"), "")

		got := ComposeFile(dir)
		want := filepath.Join(dir, "docker-compose.yml")
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("empty dir returns empty string", func(t *testing.T) {
		dir := t.TempDir()
		if got := ComposeFile(dir); got != "" {
			t.Errorf("got %q, want empty", got)
		}
	})
}

func TestIsStack(t *testing.T) {
	tests := []struct {
		name  string
		files []string
		want  bool
	}{
		{"compose.yml present", []string{"compose.yml"}, true},
		{"docker-compose.yml present", []string{"docker-compose.yml"}, true},
		{"empty dir", nil, false},
		{"unrelated file only", []string{"readme.txt"}, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			for _, f := range tc.files {
				writeFile(t, filepath.Join(dir, f), "")
			}
			if got := IsStack(dir); got != tc.want {
				t.Errorf("IsStack = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestLoadStack(t *testing.T) {
	t.Run("compose.yml and config.yml loads with name from config", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, filepath.Join(dir, "compose.yml"), "services:\n  web:\n    image: nginx\n")
		writeFile(t, filepath.Join(dir, "config.yml"), "name: mystack\ntags:\n  - backend\n")

		s, err := LoadStack(dir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if s.Name != "mystack" {
			t.Errorf("got name %q, want %q", s.Name, "mystack")
		}
		if s.Config.Name != "mystack" {
			t.Errorf("config.Name = %q, want %q", s.Config.Name, "mystack")
		}
		if len(s.Config.Tags) != 1 || s.Config.Tags[0] != "backend" {
			t.Errorf("config.Tags = %v, want [backend]", s.Config.Tags)
		}
	})

	t.Run("docker-compose.yml only name defaults to dir basename", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, filepath.Join(dir, "docker-compose.yml"), "services:\n  app:\n    image: alpine\n")

		s, err := LoadStack(dir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		wantName := filepath.Base(dir)
		if s.Name != wantName {
			t.Errorf("got name %q, want %q", s.Name, wantName)
		}
		if !filepath.IsAbs(s.Dir) {
			t.Errorf("s.Dir %q is not absolute", s.Dir)
		}
	})

	t.Run("both compose.yml and docker-compose.yml prefers compose.yml", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, filepath.Join(dir, "compose.yml"), "services:\n  web:\n    image: nginx\n")
		writeFile(t, filepath.Join(dir, "docker-compose.yml"), "services:\n  web:\n    image: wrong\n")
		writeFile(t, filepath.Join(dir, "config.yml"), "name: preferred\n")

		s, err := LoadStack(dir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if s.Name != "preferred" {
			t.Errorf("got name %q, want preferred", s.Name)
		}
		if got := ComposeFile(dir); got != filepath.Join(dir, "compose.yml") {
			t.Errorf("ComposeFile returned %q, want compose.yml path", got)
		}
	})

	t.Run("compose.yml without config.yml loads with basename as name", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, filepath.Join(dir, "compose.yml"), "services:\n  web:\n    image: nginx\n")

		s, err := LoadStack(dir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		wantName := filepath.Base(dir)
		if s.Name != wantName {
			t.Errorf("got name %q, want %q", s.Name, wantName)
		}
	})

	t.Run("empty directory returns ErrNotAStack", func(t *testing.T) {
		dir := t.TempDir()
		_, err := LoadStack(dir)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		notAStack, ok := err.(*ErrNotAStack)
		if !ok {
			t.Fatalf("expected *ErrNotAStack, got %T: %v", err, err)
		}
		if notAStack.Dir == "" {
			t.Error("ErrNotAStack.Dir should not be empty")
		}
		if got := notAStack.Error(); got == "" || !strings.Contains(got, dir) {
			t.Errorf("Error() = %q, want message containing stack path", got)
		}
	})

	t.Run("invalid config.yml returns parse error", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, filepath.Join(dir, "compose.yml"), "services:\n  web:\n    image: nginx\n")
		writeFile(t, filepath.Join(dir, "config.yml"), "name: [\n")

		if _, err := LoadStack(dir); err == nil {
			t.Fatal("expected yaml parse error")
		}
	})
}
