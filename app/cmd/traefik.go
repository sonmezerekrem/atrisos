package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/sonmezerekrem/atrisos/internal/config"
	"github.com/sonmezerekrem/atrisos/internal/traefik"
	"github.com/spf13/cobra"
)

var traefikCmd = &cobra.Command{
	Use:   "traefik",
	Short: "Manage the shared Traefik instance",
}

var traefikUpCmd = &cobra.Command{
	Use:   "up",
	Short: "Start the managed Traefik instance",
	RunE: func(cmd *cobra.Command, args []string) error {
		mgr := traefik.NewManager(rootCfg)
		printAction("starting Traefik...")
		if err := mgr.Start(); err != nil {
			printError(fmt.Sprintf("Traefik start failed: %v", err))
			os.Exit(5)
		}
		printSuccess("Traefik is running")
		return nil
	},
}

var traefikDownCmd = &cobra.Command{
	Use:   "down",
	Short: "Stop the managed Traefik instance",
	RunE: func(cmd *cobra.Command, args []string) error {
		mgr := traefik.NewManager(rootCfg)
		printAction("stopping Traefik...")
		if err := mgr.Stop(); err != nil {
			printError(fmt.Sprintf("Traefik stop failed: %v", err))
			os.Exit(5)
		}
		printSuccess("Traefik stopped")
		return nil
	},
}

var traefikRestartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restart the managed Traefik instance",
	RunE: func(cmd *cobra.Command, args []string) error {
		mgr := traefik.NewManager(rootCfg)
		printAction("restarting Traefik...")
		if err := mgr.Restart(); err != nil {
			printError(fmt.Sprintf("Traefik restart failed: %v", err))
			os.Exit(5)
		}
		printSuccess("Traefik restarted")
		return nil
	},
}

var traefikStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show Traefik container status",
	RunE: func(cmd *cobra.Command, args []string) error {
		mgr := traefik.NewManager(rootCfg)
		status, err := mgr.Status()
		if err != nil {
			return fmt.Errorf("checking Traefik status: %w", err)
		}
		fmt.Printf("Traefik: %s\n", status)
		return nil
	},
}

var traefikLogsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Tail Traefik logs",
	RunE: func(cmd *cobra.Command, args []string) error {
		traefikDir := filepath.Join(config.Dir(), "traefik")
		composeFile := filepath.Join(traefikDir, "compose.yml")
		envFile := filepath.Join(traefikDir, ".env")

		binary, cmdArgs := buildTraefikComposeArgs(envFile, composeFile,
			"logs", "-f", "--tail", "100")
		c := exec.Command(binary, cmdArgs...)
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		c.Dir = traefikDir
		return c.Run()
	},
}

var traefikDashboardCmd = &cobra.Command{
	Use:   "dashboard",
	Short: "Print the Traefik dashboard URL",
	RunE: func(cmd *cobra.Command, args []string) error {
		url := "http://localhost:8080/dashboard/"
		if rootCfg.Traefik.Dashboard.Enabled && rootCfg.Traefik.Dashboard.Host != "" {
			url = fmt.Sprintf("https://%s/dashboard/", rootCfg.Traefik.Dashboard.Host)
		}
		fmt.Println(url)
		return nil
	},
}

// buildTraefikComposeArgs constructs the compose command args for the traefik stack.
func buildTraefikComposeArgs(envFile, composeFile string, args ...string) (string, []string) {
	composeCmd := detectComposeCmd()
	parts := strings.Fields(composeCmd)
	binary := parts[0]

	cmdArgs := make([]string, 0, 20)
	cmdArgs = append(cmdArgs, parts[1:]...)
	if _, err := os.Stat(envFile); err == nil {
		cmdArgs = append(cmdArgs, "--env-file", envFile)
	}
	cmdArgs = append(cmdArgs, "-f", composeFile, "--project-name", "atrisos-traefik")
	cmdArgs = append(cmdArgs, args...)
	return binary, cmdArgs
}

// detectComposeCmd detects the available compose command (private helper).
func detectComposeCmd() string {
	c := exec.Command("podman", "compose", "version")
	c.Stdout = nil
	c.Stderr = nil
	if err := c.Run(); err == nil {
		return "podman compose"
	}
	return "podman-compose"
}

func init() {
	traefikCmd.AddCommand(traefikUpCmd)
	traefikCmd.AddCommand(traefikDownCmd)
	traefikCmd.AddCommand(traefikRestartCmd)
	traefikCmd.AddCommand(traefikStatusCmd)
	traefikCmd.AddCommand(traefikLogsCmd)
	traefikCmd.AddCommand(traefikDashboardCmd)
	rootCmd.AddCommand(traefikCmd)
}
