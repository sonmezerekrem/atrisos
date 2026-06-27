package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/sonmezerekrem/atrisos/internal/compose"
	"github.com/sonmezerekrem/atrisos/internal/stack"
	"github.com/spf13/cobra"
)

var execCmd = &cobra.Command{
	Use:   "exec <stack> <service> -- <command...>",
	Short: "Run a command inside a running container",
	// Disable arg parsing after "--"
	DisableFlagParsing: false,
	Args:               cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			return fmt.Errorf("usage: atrisos exec <stack> <service> -- <command...>")
		}
		stackName := args[0]
		service := args[1]

		// Everything after "--" is the command.
		var cmdArgs []string
		for i, a := range args {
			if a == "--" && i+1 < len(args) {
				cmdArgs = args[i+1:]
				break
			}
		}
		if len(cmdArgs) == 0 && len(args) > 2 {
			cmdArgs = args[2:]
		}
		if len(cmdArgs) == 0 {
			return fmt.Errorf("no command specified after --")
		}

		stacks, err := resolveStacks([]string{stackName}, false, "")
		if err != nil {
			return err
		}
		s := stacks[0]
		return execInContainer(s, service, cmdArgs, false)
	},
}

var shellCmd = &cobra.Command{
	Use:   "shell <stack> <service>",
	Short: "Open an interactive shell in a running container",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		stacks, err := resolveStacks([]string{args[0]}, false, "")
		if err != nil {
			return err
		}
		s := stacks[0]
		service := args[1]

		// Try /bin/bash first, fall back to /bin/sh.
		err = execInContainer(s, service, []string{"/bin/bash"}, true)
		if err != nil {
			return execInContainer(s, service, []string{"/bin/sh"}, true)
		}
		return nil
	},
}

// execInContainer runs a command in the container via podman compose exec.
func execInContainer(s *stack.Stack, service string, cmdArgs []string, interactive bool) error {
	doc, err := compose.LoadAndMerge(s.Dir, s.Config)
	if err != nil {
		return fmt.Errorf("loading compose: %w", err)
	}

	tmpFile, err := compose.WriteToTemp(doc)
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile)

	runner := compose.NewRunner(s.Dir, s.Name)
	parts := strings.Fields(runner.ComposeCmd)
	binary := parts[0]

	// Build compose exec args.
	composeArgs := make([]string, 0, 20)
	composeArgs = append(composeArgs, parts[1:]...)
	if runner.EnvFile != "" {
		composeArgs = append(composeArgs, "--env-file", runner.EnvFile)
	}
	composeArgs = append(composeArgs, "-f", tmpFile, "--project-name", s.Name, "exec")

	// Check if we're in a TTY for -it flags.
	fi, statErr := os.Stdin.Stat()
	isTTY := statErr == nil && (fi.Mode()&os.ModeCharDevice) != 0
	if interactive || isTTY {
		composeArgs = append(composeArgs, "-it")
	}

	composeArgs = append(composeArgs, service)
	composeArgs = append(composeArgs, cmdArgs...)

	c := exec.Command(binary, composeArgs...)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.Dir = s.Dir
	return c.Run()
}

func init() {
	rootCmd.AddCommand(execCmd)
	rootCmd.AddCommand(shellCmd)
}
