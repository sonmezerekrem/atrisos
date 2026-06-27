package stack

// Stack represents a discovered and parsed atrisos stack.
type Stack struct {
	Dir    string      // absolute path to stack directory
	Name   string      // from config.yml name, or dir basename if empty
	Config StackConfig // parsed config.yml content
}

// StackConfig is the parsed content of a stack's config.yml.
type StackConfig struct {
	Name        string            `yaml:"name"`
	Description string            `yaml:"description"`
	Tags        []string          `yaml:"tags"`
	Meta        map[string]string `yaml:"meta"`
	Domains     []DomainConfig    `yaml:"domains"`
	Update      StackUpdateConf   `yaml:"update"`
	AutoStart   bool              `yaml:"auto_start"`
	Backup      BackupConfig      `yaml:"backup"`
	Notify      NotifyConfig      `yaml:"notify"`
}

// DomainConfig describes how a service is exposed via Traefik.
type DomainConfig struct {
	Service     string   `yaml:"service"`
	Host        string   `yaml:"host"`
	Port        int      `yaml:"port"`
	PathPrefix  string   `yaml:"path_prefix"`
	TLS         string   `yaml:"tls"`        // "true" | "staging" | "false" — default "true"
	Middlewares []string `yaml:"middlewares"`
}

// StackUpdateConf holds per-stack update configuration.
type StackUpdateConf struct {
	Mode string `yaml:"mode"` // "manual" | "watch" | "" (inherit global)
}

// BackupConfig holds per-stack backup configuration.
type BackupConfig struct {
	Enabled     bool     `yaml:"enabled"`
	Schedule    string   `yaml:"schedule"`
	Destination string   `yaml:"destination"`
	Volumes     []string `yaml:"volumes"`
}

// NotifyConfig holds per-stack webhook notification configuration.
type NotifyConfig struct {
	Webhook string `yaml:"webhook"`
}
