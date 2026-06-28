package compose

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunnerBuildArgs(t *testing.T) {
	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env")
	if err := os.WriteFile(envPath, []byte("FOO=bar\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	r := &Runner{
		ComposeCmd:  "podman compose",
		ProjectDir:  dir,
		EnvFile:     envPath,
		ProjectName: "myapp",
	}

	binary, args := r.buildArgs("/tmp/merged-compose.yml", []string{"up", "-d"})
	if binary != "podman" {
		t.Errorf("binary = %q, want podman", binary)
	}

	joined := strings.Join(args, " ")
	for _, want := range []string{"compose", "--env-file", envPath, "-f", "/tmp/merged-compose.yml", "--project-name", "myapp", "up", "-d"} {
		if !strings.Contains(joined, want) {
			t.Errorf("args %q missing %q", joined, want)
		}
	}
}

func TestRunnerBuildArgsWithoutEnvFile(t *testing.T) {
	r := &Runner{
		ComposeCmd:  "podman-compose",
		ProjectDir:  t.TempDir(),
		ProjectName: "demo",
	}

	_, args := r.buildArgs("/tmp/compose.yml", []string{"ps"})
	joined := strings.Join(args, " ")
	if strings.Contains(joined, "--env-file") {
		t.Errorf("unexpected --env-file in args: %s", joined)
	}
	if !strings.Contains(joined, "--project-name demo") {
		t.Errorf("args = %q", joined)
	}
}
