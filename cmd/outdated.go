package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var outdatedCmd = &cobra.Command{
	Use:   "outdated [stack]",
	Short: "Check for newer image versions (not yet implemented)",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: implement registry digest comparison.
		// Planned behavior:
		//   1. Parse image references from compose.yml for each stack
		//   2. Query OCI registry API for latest manifest digest
		//   3. Compare with locally pulled digest (podman image inspect)
		//   4. Print list of services with available updates
		fmt.Println("atrisos outdated: not yet implemented")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(outdatedCmd)
}
