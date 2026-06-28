package backup

import (
	"path/filepath"
	"testing"
)

func TestExpandDest(t *testing.T) {
	t.Run("s3 url unchanged", func(t *testing.T) {
		got := expandDest("s3://bucket/path")
		if got != "s3://bucket/path" {
			t.Errorf("got %q", got)
		}
	})

	t.Run("tilde path expands", func(t *testing.T) {
		home := t.TempDir()
		t.Setenv("HOME", home)
		got := expandDest("~/backups/myapp")
		want := filepath.Join(home, "backups/myapp")
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("absolute path unchanged", func(t *testing.T) {
		got := expandDest("/var/backups")
		if got != "/var/backups" {
			t.Errorf("got %q", got)
		}
	})
}

func TestParseVolumeListJSON(t *testing.T) {
	var entries []struct {
		Name string `json:"Name"`
	}
	if err := parseVolumeListJSON(`[{"Name":"myapp_db_data"},{"Name":"myapp_uploads"}]`, &entries); err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 || entries[0].Name != "myapp_db_data" {
		t.Errorf("got %+v", entries)
	}
	if err := parseVolumeListJSON(`not json`, &entries); err == nil {
		t.Error("expected parse error")
	}
}
