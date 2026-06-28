package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config holds the global atrisos configuration.
type Config struct {
	StacksRoot string         `yaml:"stacks_root"`
	Update     UpdateDefaults `yaml:"update"`
	Traefik    TraefikConfig  `yaml:"traefik"`
	Backup     BackupDefaults `yaml:"backup"`
	Podman     PodmanConfig   `yaml:"podman"`
	Output     OutputConfig   `yaml:"output"`
}

// UpdateDefaults holds global update behavior settings.
type UpdateDefaults struct {
	DefaultMode string `yaml:"default_mode"` // "manual" | "watch"
}

// TraefikConfig holds Traefik-related configuration.
type TraefikConfig struct {
	ACMEEmail string               `yaml:"acme_email"`
	Network   string               `yaml:"network"`
	HTTPPort  int                  `yaml:"http_port"`
	HTTPSPort int                  `yaml:"https_port"`
	Image     string               `yaml:"image"`
	Dashboard TraefikDashboardConf `yaml:"dashboard"`
}

// TraefikDashboardConf configures the Traefik dashboard exposure.
type TraefikDashboardConf struct {
	Enabled bool   `yaml:"enabled"`
	Host    string `yaml:"host"`
}

// BackupDefaults holds global backup defaults.
type BackupDefaults struct {
	DefaultDestination string `yaml:"default_destination"`
}

// PodmanConfig holds Podman-related configuration.
type PodmanConfig struct {
	ComposeCommand string `yaml:"compose_command"`
	MachineName    string `yaml:"machine_name"`
	MachineCPUs    int    `yaml:"machine_cpus"`
	MachineMemory  string `yaml:"machine_memory"`
}

// OutputConfig controls output formatting.
type OutputConfig struct {
	NoColor         bool   `yaml:"no_color"`
	NoEmoji         bool   `yaml:"no_emoji"`
	TimestampFormat string `yaml:"timestamp_format"`
}

// defaults returns a Config with all defaults filled in.
func defaults() *Config {
	home, _ := os.UserHomeDir()
	return &Config{
		StacksRoot: filepath.Join(home, "atrisos-stacks"),
		Update: UpdateDefaults{
			DefaultMode: "manual",
		},
		Traefik: TraefikConfig{
			Network:   "atrisos_net",
			HTTPPort:  80,
			HTTPSPort: 443,
			Image:     "traefik:v3",
		},
		Backup: BackupDefaults{
			DefaultDestination: filepath.Join(home, "atrisos-backups"),
		},
		Podman: PodmanConfig{
			ComposeCommand: "auto",
			MachineName:    "atrisos",
			MachineCPUs:    2,
			MachineMemory:  "2048",
		},
		Output: OutputConfig{
			TimestampFormat: "2006-01-02 15:04",
		},
	}
}

// Dir returns the atrisos config directory, respecting $XDG_CONFIG_HOME.
func Dir() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "atrisos")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "atrisos")
}

// DefaultPath returns the default path to the global config file.
func DefaultPath() string {
	return filepath.Join(Dir(), "config.yml")
}

// Load reads the config from the given path. If the file doesn't exist,
// it returns defaults without error.
func Load(path string) (*Config, error) {
	cfg := defaults()
	if path == "" {
		path = DefaultPath()
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, err
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	// Fill in any zero-value fields with defaults after unmarshaling.
	d := defaults()
	if cfg.StacksRoot == "" {
		cfg.StacksRoot = d.StacksRoot
	}
	if cfg.Update.DefaultMode == "" {
		cfg.Update.DefaultMode = d.Update.DefaultMode
	}
	if cfg.Traefik.Network == "" {
		cfg.Traefik.Network = d.Traefik.Network
	}
	if cfg.Traefik.HTTPPort == 0 {
		cfg.Traefik.HTTPPort = d.Traefik.HTTPPort
	}
	if cfg.Traefik.HTTPSPort == 0 {
		cfg.Traefik.HTTPSPort = d.Traefik.HTTPSPort
	}
	if cfg.Traefik.Image == "" {
		cfg.Traefik.Image = d.Traefik.Image
	}
	if cfg.Podman.ComposeCommand == "" {
		cfg.Podman.ComposeCommand = d.Podman.ComposeCommand
	}
	if cfg.Podman.MachineName == "" {
		cfg.Podman.MachineName = d.Podman.MachineName
	}
	if cfg.Output.TimestampFormat == "" {
		cfg.Output.TimestampFormat = d.Output.TimestampFormat
	}

	// Expand ~ in StacksRoot.
	if len(cfg.StacksRoot) >= 2 && cfg.StacksRoot[:2] == "~/" {
		home, _ := os.UserHomeDir()
		cfg.StacksRoot = filepath.Join(home, cfg.StacksRoot[2:])
	}

	return cfg, nil
}
