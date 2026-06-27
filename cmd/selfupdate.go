package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var selfUpdateCmd = &cobra.Command{
	Use:         "self-update",
	Short:       "Download and install the latest atrisos release (not yet implemented)",
	Annotations: map[string]string{"skipPreRun": "true"},
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: implement self-update.
		// Planned behavior:
		//   1. Fetch latest release from GitHub API
		//   2. Download binary for current OS/arch
		//   3. Verify SHA-256 checksum against release manifest
		//   4. Replace running binary in place
		fmt.Println("atrisos self-update: not yet implemented")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(selfUpdateCmd)
}
