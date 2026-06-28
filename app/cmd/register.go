package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/sonmezerekrem/atrisos/app/internal/stack"
	"github.com/spf13/cobra"
)

var registerCmd = &cobra.Command{
	Use:   "register <path>",
	Short: "Register a stack directory outside the root directory",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := args[0]
		absPath, err := filepath.Abs(path)
		if err != nil {
			return fmt.Errorf("resolving path: %w", err)
		}

		if !stack.IsStack(absPath) {
			return fmt.Errorf("%s is not a stack directory (no compose.yml or docker-compose.yml)", absPath)
		}

		rootReg.Add(absPath)
		if err := rootReg.Save(); err != nil {
			return fmt.Errorf("saving registry: %w", err)
		}

		s, _ := stack.LoadStack(absPath)
		name := filepath.Base(absPath)
		if s != nil {
			name = s.Name
		}

		printSuccess(fmt.Sprintf("registered %s (%s)", name, absPath))
		return nil
	},
}

var unregisterCmd = &cobra.Command{
	Use:   "unregister <stack>",
	Short: "Remove a stack from the registry",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		// Try to find the stack by name in extra paths.
		found := false
		for _, p := range rootReg.Paths() {
			base := filepath.Base(p)
			// Also check if it matches the stack name from config.yml.
			s, err := stack.LoadStack(p)
			if err == nil && (s.Name == name || base == name) {
				found = true
				break
			} else if base == name {
				found = true
				break
			}
		}

		if !found {
			printWarn(fmt.Sprintf("stack %q not found in registry", name))
		}

		rootReg.Remove(name)
		if err := rootReg.Save(); err != nil {
			return fmt.Errorf("saving registry: %w", err)
		}

		printSuccess(fmt.Sprintf("unregistered %s", name))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(registerCmd)
	rootCmd.AddCommand(unregisterCmd)
}
