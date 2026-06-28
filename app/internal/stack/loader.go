package stack

import (
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

// ComposeFile returns the path to the compose file in dir.
// compose.yml takes precedence over docker-compose.yml.
// Returns empty string if neither exists.
func ComposeFile(dir string) string {
	primary := filepath.Join(dir, "compose.yml")
	if _, err := os.Stat(primary); err == nil {
		return primary
	}
	secondary := filepath.Join(dir, "docker-compose.yml")
	if _, err := os.Stat(secondary); err == nil {
		return secondary
	}
	return ""
}

// IsStack returns true if dir contains compose.yml or docker-compose.yml.
func IsStack(dir string) bool {
	return ComposeFile(dir) != ""
}

// LoadStack loads a stack from the given directory. Missing config.yml is
// not an error — defaults are used in that case.
func LoadStack(dir string) (*Stack, error) {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}

	// Compose file must exist.
	if !IsStack(absDir) {
		return nil, &ErrNotAStack{Dir: absDir}
	}

	s := &Stack{Dir: absDir}

	// Load config.yml if present.
	cfgPath := filepath.Join(absDir, "config.yml")
	data, err := os.ReadFile(cfgPath)
	if err == nil {
		if yamlErr := yaml.Unmarshal(data, &s.Config); yamlErr != nil {
			return nil, yamlErr
		}
	} else if !os.IsNotExist(err) {
		return nil, err
	}

	// Set stack name: prefer config name, fall back to directory basename.
	if s.Config.Name != "" {
		s.Name = s.Config.Name
	} else {
		s.Name = filepath.Base(absDir)
	}

	// Load .env if present (for display purposes; podman compose reads it directly).
	envPath := filepath.Join(absDir, ".env")
	if _, err := os.Stat(envPath); err == nil {
		// Ignore errors — .env is optional
		_ = godotenv.Load(envPath)
	}

	return s, nil
}

// ErrNotAStack is returned when a directory is not a valid stack.
type ErrNotAStack struct {
	Dir string
}

func (e *ErrNotAStack) Error() string {
	return "not a stack directory (no compose.yml or docker-compose.yml): " + e.Dir
}
