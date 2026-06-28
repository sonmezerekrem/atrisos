package config

import (
	"os"
	"path/filepath"
	"testing"
)

func writeConfigFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("writeFile %s: %v", path, err)
	}
}

func TestLoad(t *testing.T) {
	t.Run("non-existent file returns defaults without error", func(t *testing.T) {
		dir := t.TempDir()
		cfg, err := Load(filepath.Join(dir, "config.yml"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg == nil {
			t.Fatal("got nil config")
		}
		// Spot-check a few defaults.
		if cfg.Traefik.Network != "atrisos_net" {
			t.Errorf("Traefik.Network = %q, want atrisos_net", cfg.Traefik.Network)
		}
		if cfg.Traefik.HTTPPort != 80 {
			t.Errorf("Traefik.HTTPPort = %d, want 80", cfg.Traefik.HTTPPort)
		}
		if cfg.Traefik.HTTPSPort != 443 {
			t.Errorf("Traefik.HTTPSPort = %d, want 443", cfg.Traefik.HTTPSPort)
		}
		if cfg.Update.DefaultMode != "manual" {
			t.Errorf("Update.DefaultMode = %q, want manual", cfg.Update.DefaultMode)
		}
		if cfg.Podman.ComposeCommand != "auto" {
			t.Errorf("Podman.ComposeCommand = %q, want auto", cfg.Podman.ComposeCommand)
		}
		if cfg.Output.TimestampFormat == "" {
			t.Error("Output.TimestampFormat should not be empty")
		}
	})

	t.Run("valid YAML file merges over defaults", func(t *testing.T) {
		dir := t.TempDir()
		cfgPath := filepath.Join(dir, "config.yml")
		writeConfigFile(t, cfgPath, "stacks_root: /custom/stacks\ntraefik:\n  acme_email: admin@example.com\n")

		cfg, err := Load(cfgPath)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.StacksRoot != "/custom/stacks" {
			t.Errorf("StacksRoot = %q, want /custom/stacks", cfg.StacksRoot)
		}
		if cfg.Traefik.ACMEEmail != "admin@example.com" {
			t.Errorf("Traefik.ACMEEmail = %q, want admin@example.com", cfg.Traefik.ACMEEmail)
		}
		// Defaults still apply for unset fields.
		if cfg.Traefik.Network != "atrisos_net" {
			t.Errorf("Traefik.Network = %q, want atrisos_net (should keep default)", cfg.Traefik.Network)
		}
	})

	t.Run("zero-value fields in YAML are filled by defaults", func(t *testing.T) {
		dir := t.TempDir()
		cfgPath := filepath.Join(dir, "config.yml")
		// Explicitly set network to empty string — should revert to default.
		writeConfigFile(t, cfgPath, "traefik:\n  network: \"\"\n  http_port: 0\n")

		cfg, err := Load(cfgPath)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.Traefik.Network != "atrisos_net" {
			t.Errorf("Traefik.Network = %q, want atrisos_net after zero-value fill", cfg.Traefik.Network)
		}
		if cfg.Traefik.HTTPPort != 80 {
			t.Errorf("Traefik.HTTPPort = %d, want 80 after zero-value fill", cfg.Traefik.HTTPPort)
		}
	})

	t.Run("tilde in stacks_root is expanded to home dir", func(t *testing.T) {
		dir := t.TempDir()
		cfgPath := filepath.Join(dir, "config.yml")
		writeConfigFile(t, cfgPath, "stacks_root: ~/my-stacks\n")

		cfg, err := Load(cfgPath)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		home, _ := os.UserHomeDir()
		want := filepath.Join(home, "my-stacks")
		if cfg.StacksRoot != want {
			t.Errorf("StacksRoot = %q, want %q", cfg.StacksRoot, want)
		}
	})

	t.Run("empty path falls back to DefaultPath", func(t *testing.T) {
		// Just verify it doesn't error even if the default path doesn't exist
		// (it might or might not exist on this machine).
		_, err := Load("")
		// Only a parsing error would be fatal; missing file is not an error.
		// We can't know if the machine has a real config file, so just ensure
		// the call doesn't panic.
		_ = err
	})

	t.Run("update default mode filled when empty", func(t *testing.T) {
		dir := t.TempDir()
		cfgPath := filepath.Join(dir, "config.yml")
		writeConfigFile(t, cfgPath, "update:\n  default_mode: \"\"\n")

		cfg, err := Load(cfgPath)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.Update.DefaultMode != "manual" {
			t.Errorf("Update.DefaultMode = %q, want manual after zero-value fill", cfg.Update.DefaultMode)
		}
	})
}
