package traefik

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/sonmezerekrem/atrisos/app/internal/config"
)

// Manager manages the shared Traefik compose stack.
type Manager struct {
	cfg *config.Config
	dir string // ~/.config/atrisos/traefik/
}

// NewManager creates a Manager for the given config.
func NewManager(cfg *config.Config) *Manager {
	return &Manager{
		cfg: cfg,
		dir: filepath.Join(config.Dir(), "traefik"),
	}
}

// EnsureStarted checks if Traefik is running and starts it if not.
func (m *Manager) EnsureStarted() error {
	status, err := m.Status()
	if err != nil {
		return err
	}
	if status == "running" {
		return nil
	}
	return m.Start()
}

// Start starts the managed Traefik instance.
func (m *Manager) Start() error {
	// Ensure the shared network exists.
	if err := EnsureNetwork(m.cfg.Traefik.Network); err != nil {
		return fmt.Errorf("traefik: %w", err)
	}

	// Check that the required ports are not already in use.
	if err := m.checkPort(m.cfg.Traefik.HTTPPort); err != nil {
		return err
	}
	if err := m.checkPort(m.cfg.Traefik.HTTPSPort); err != nil {
		return err
	}

	// Get the Podman socket path.
	socketPath, err := PodmanSocketPath(m.cfg.Podman.MachineName)
	if err != nil {
		return fmt.Errorf("traefik: resolving podman socket: %w", err)
	}

	// Write compose files.
	if err := m.writeComposeFiles(socketPath); err != nil {
		return fmt.Errorf("traefik: writing compose files: %w", err)
	}

	// Run podman compose up -d.
	return m.runCompose("up", "-d")
}

// Stop stops the managed Traefik instance.
func (m *Manager) Stop() error {
	return m.runCompose("down")
}

// Restart stops then starts Traefik.
func (m *Manager) Restart() error {
	if err := m.Stop(); err != nil {
		return err
	}
	return m.Start()
}

// Status returns "running", "stopped", or "not found".
func (m *Manager) Status() (string, error) {
	cmd := exec.Command("podman", "ps",
		"--filter", "name=atrisos_traefik",
		"--filter", "status=running",
		"--format", "{{.Names}}")
	out, err := cmd.Output()
	if err != nil {
		return "not found", nil
	}
	output := strings.TrimSpace(string(out))
	if output == "" {
		return "stopped", nil
	}
	return "running", nil
}

// checkPort does a TCP dial to see if the port is already bound.
func (m *Manager) checkPort(port int) error {
	addr := fmt.Sprintf("127.0.0.1:%d", port)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		// Port not in use — good.
		return nil
	}
	conn.Close()
	return fmt.Errorf("port %d is already in use; Traefik cannot bind to it (exit 5)", port)
}

// writeComposeFiles writes compose.yml and .env into the Traefik dir.
func (m *Manager) writeComposeFiles(socketPath string) error {
	if err := os.MkdirAll(m.dir, 0o755); err != nil {
		return err
	}

	// Write .env
	envContent := fmt.Sprintf("ACME_EMAIL=%s\nPODMAN_SOCKET=%s\n",
		m.cfg.Traefik.ACMEEmail, socketPath)
	if err := os.WriteFile(filepath.Join(m.dir, ".env"), []byte(envContent), 0o600); err != nil {
		return err
	}

	// Write compose.yml
	composePath := filepath.Join(m.dir, "compose.yml")
	composeContent := m.generateComposeYAML()
	return os.WriteFile(composePath, []byte(composeContent), 0o644)
}

// generateComposeYAML returns the Traefik compose.yml content.
func (m *Manager) generateComposeYAML() string {
	cfg := m.cfg.Traefik
	network := cfg.Network
	httpPort := cfg.HTTPPort
	httpsPort := cfg.HTTPSPort
	image := cfg.Image

	return fmt.Sprintf(`services:
  traefik:
    image: %s
    container_name: atrisos_traefik
    command:
      - "--providers.docker=true"
      - "--providers.docker.network=%s"
      - "--providers.docker.exposedbydefault=false"
      - "--entrypoints.web.address=:%d"
      - "--entrypoints.websecure.address=:%d"
      - "--certificatesresolvers.letsencrypt.acme.httpchallenge=true"
      - "--certificatesresolvers.letsencrypt.acme.httpchallenge.entrypoint=web"
      - "--certificatesresolvers.letsencrypt.acme.email=${ACME_EMAIL}"
      - "--certificatesresolvers.letsencrypt.acme.storage=/letsencrypt/acme.json"
      - "--certificatesresolvers.letsencrypt-staging.acme.httpchallenge=true"
      - "--certificatesresolvers.letsencrypt-staging.acme.httpchallenge.entrypoint=web"
      - "--certificatesresolvers.letsencrypt-staging.acme.email=${ACME_EMAIL}"
      - "--certificatesresolvers.letsencrypt-staging.acme.storage=/letsencrypt/acme-staging.json"
      - "--certificatesresolvers.letsencrypt-staging.acme.caserver=https://acme-staging-v02.api.letsencrypt.org/directory"
      - "--api.dashboard=true"
      - "--log.level=INFO"
    ports:
      - "%d:%d"
      - "%d:%d"
    volumes:
      - ${PODMAN_SOCKET}:/var/run/docker.sock:ro
      - letsencrypt:/letsencrypt
    networks:
      - %s
    restart: unless-stopped

networks:
  %s:
    external: true

volumes:
  letsencrypt:
`, image, network, httpPort, httpsPort,
		httpPort, httpPort, httpsPort, httpsPort,
		network, network)
}

// detectComposeCmd detects the available compose command without importing
// the compose package (to avoid circular imports).
func detectComposeCmd() string {
	cmd := exec.Command("podman", "compose", "version")
	cmd.Stdout = nil
	cmd.Stderr = nil
	if err := cmd.Run(); err == nil {
		return "podman compose"
	}
	return "podman-compose"
}

// runCompose runs a podman compose command in the Traefik directory.
func (m *Manager) runCompose(args ...string) error {
	composeFile := filepath.Join(m.dir, "compose.yml")
	envFile := filepath.Join(m.dir, ".env")
	composeCmd := detectComposeCmd()
	parts := strings.Fields(composeCmd)
	binary := parts[0]

	// Build: [subcommands...] [--env-file env] -f compose.yml --project-name name [args...]
	cmdArgs := make([]string, 0, len(parts)+8+len(args))
	cmdArgs = append(cmdArgs, parts[1:]...)
	if _, err := os.Stat(envFile); err == nil {
		cmdArgs = append(cmdArgs, "--env-file", envFile)
	}
	cmdArgs = append(cmdArgs, "-f", composeFile, "--project-name", "atrisos-traefik")
	cmdArgs = append(cmdArgs, args...)

	cmd := exec.Command(binary, cmdArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = m.dir
	return cmd.Run()
}
