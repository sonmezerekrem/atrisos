package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:         "init [name]",
	Short:       "Create a new stack from a template (not yet implemented)",
	Annotations: map[string]string{"skipPreRun": "true"},
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: implement template fetching from GitHub and interactive wizard.
		// Planned behavior:
		//   1. Fetch template list from github.com/sonmezerekrem/atrisos/templates/ (main branch)
		//   2. Cache locally in ~/.config/atrisos/templates-cache/
		//   3. Interactive prompts: name, template, domain, port, backup options
		//   4. Generate compose.yml, config.yml, .env.example in stacks root
		fmt.Println("atrisos init: template system not yet implemented")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
