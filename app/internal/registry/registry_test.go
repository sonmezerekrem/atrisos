package registry

import (
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	t.Run("non-existent file returns empty registry without error", func(t *testing.T) {
		dir := t.TempDir()
		reg, err := Load(dir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if reg == nil {
			t.Fatal("got nil registry")
		}
		if len(reg.Paths()) != 0 {
			t.Errorf("Paths() = %v, want empty", reg.Paths())
		}
	})

	t.Run("valid registry.json is loaded correctly", func(t *testing.T) {
		dir := t.TempDir()
		reg, err := Load(dir)
		if err != nil {
			t.Fatalf("load: %v", err)
		}
		reg.Add("/stacks/myapp")
		if err := reg.Save(); err != nil {
			t.Fatalf("save: %v", err)
		}

		reg2, err := Load(dir)
		if err != nil {
			t.Fatalf("reload: %v", err)
		}
		paths := reg2.Paths()
		if len(paths) != 1 || paths[0] != "/stacks/myapp" {
			t.Errorf("Paths() = %v, want [/stacks/myapp]", paths)
		}
	})
}

func TestAdd(t *testing.T) {
	t.Run("Add path makes it appear in Paths", func(t *testing.T) {
		dir := t.TempDir()
		reg, _ := Load(dir)
		reg.Add("/some/path/mystack")

		paths := reg.Paths()
		if len(paths) != 1 || paths[0] != "/some/path/mystack" {
			t.Errorf("Paths() = %v, want [/some/path/mystack]", paths)
		}
	})

	t.Run("Add duplicate is deduplicated", func(t *testing.T) {
		dir := t.TempDir()
		reg, _ := Load(dir)
		reg.Add("/some/path/mystack")
		reg.Add("/some/path/mystack")
		reg.Add("/some/path/mystack")

		paths := reg.Paths()
		if len(paths) != 1 {
			t.Errorf("Paths() has %d entries, want 1 (dedup): %v", len(paths), paths)
		}
	})

	t.Run("Add multiple different paths", func(t *testing.T) {
		dir := t.TempDir()
		reg, _ := Load(dir)
		reg.Add("/stacks/alpha")
		reg.Add("/stacks/beta")

		paths := reg.Paths()
		if len(paths) != 2 {
			t.Errorf("Paths() has %d entries, want 2: %v", len(paths), paths)
		}
	})
}

func TestRemove(t *testing.T) {
	t.Run("Remove by name removes the matching path", func(t *testing.T) {
		dir := t.TempDir()
		reg, _ := Load(dir)
		reg.Add("/stacks/alpha")
		reg.Add("/stacks/beta")
		reg.Add("/stacks/gamma")

		reg.Remove("beta")

		paths := reg.Paths()
		if len(paths) != 2 {
			t.Errorf("Paths() has %d entries after remove, want 2: %v", len(paths), paths)
		}
		for _, p := range paths {
			if filepath.Base(p) == "beta" {
				t.Errorf("beta should have been removed, but found %q in paths", p)
			}
		}
	})

	t.Run("Remove non-existent name is a no-op", func(t *testing.T) {
		dir := t.TempDir()
		reg, _ := Load(dir)
		reg.Add("/stacks/alpha")

		reg.Remove("does-not-exist")

		paths := reg.Paths()
		if len(paths) != 1 || paths[0] != "/stacks/alpha" {
			t.Errorf("Paths() = %v, want [/stacks/alpha]", paths)
		}
	})

	t.Run("Remove all paths", func(t *testing.T) {
		dir := t.TempDir()
		reg, _ := Load(dir)
		reg.Add("/stacks/alpha")
		reg.Remove("alpha")

		paths := reg.Paths()
		if len(paths) != 0 {
			t.Errorf("Paths() = %v, want empty after removing last entry", paths)
		}
	})
}

func TestSaveAndLoad(t *testing.T) {
	t.Run("Save and Load round-trip preserves all paths", func(t *testing.T) {
		dir := t.TempDir()
		reg, err := Load(dir)
		if err != nil {
			t.Fatalf("initial load: %v", err)
		}

		reg.Add("/stacks/one")
		reg.Add("/stacks/two")
		reg.Add("/stacks/three")

		if err := reg.Save(); err != nil {
			t.Fatalf("save: %v", err)
		}

		reg2, err := Load(dir)
		if err != nil {
			t.Fatalf("reload: %v", err)
		}

		paths := reg2.Paths()
		if len(paths) != 3 {
			t.Fatalf("got %d paths after round-trip, want 3: %v", len(paths), paths)
		}

		want := map[string]bool{
			"/stacks/one":   true,
			"/stacks/two":   true,
			"/stacks/three": true,
		}
		for _, p := range paths {
			if !want[p] {
				t.Errorf("unexpected path %q in round-trip result", p)
			}
		}
	})

	t.Run("Save after Remove persists the removal", func(t *testing.T) {
		dir := t.TempDir()
		reg, _ := Load(dir)
		reg.Add("/stacks/alpha")
		reg.Add("/stacks/beta")
		reg.Save()

		reg.Remove("alpha")
		if err := reg.Save(); err != nil {
			t.Fatalf("save after remove: %v", err)
		}

		reg2, err := Load(dir)
		if err != nil {
			t.Fatalf("reload: %v", err)
		}
		paths := reg2.Paths()
		if len(paths) != 1 || paths[0] != "/stacks/beta" {
			t.Errorf("Paths() = %v after remove+save+load, want [/stacks/beta]", paths)
		}
	})
}

func TestPaths(t *testing.T) {
	t.Run("Paths on fresh empty registry returns empty slice not nil", func(t *testing.T) {
		dir := t.TempDir()
		reg, _ := Load(dir)
		paths := reg.Paths()
		if paths == nil {
			t.Error("Paths() should return empty slice, not nil")
		}
		if len(paths) != 0 {
			t.Errorf("Paths() = %v, want empty", paths)
		}
	})
}
