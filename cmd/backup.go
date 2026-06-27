package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var backupCmd = &cobra.Command{
	Use:   "backup <stack>",
	Short: "Manually trigger a backup (not yet implemented)",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: implement backup using bundled restic binary.
		// Planned behavior:
		//   1. Download restic binary to ~/.config/atrisos/bin/restic if not present
		//   2. Verify checksum
		//   3. Stop target volumes temporarily, snapshot, restart
		//   4. Run restic backup to stack.Config.Backup.Destination
		//   5. Notify via webhook if configured
		fmt.Println("atrisos backup: not yet implemented")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(backupCmd)
}
