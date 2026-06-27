package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/sonmezerekrem/atrisos/internal/backup"
	"github.com/sonmezerekrem/atrisos/internal/stack"
	"github.com/spf13/cobra"
)

var backupDryRun bool

var backupCmd = &cobra.Command{
	Use:   "backup <stack>",
	Short: "Manually trigger a backup of a stack's volumes using restic",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		discovered, err := stack.Discover(rootCfg, rootReg)
		if err != nil {
			return fmt.Errorf("discovering stacks: %w", err)
		}
		s := stack.FindByName(discovered, name)
		if s == nil {
			printError(fmt.Sprintf("stack not found: %s", name))
			os.Exit(2)
		}

		dest := s.Config.Backup.Destination
		if dest == "" {
			dest = filepath.Join(rootCfg.Backup.DefaultDestination, s.Name)
		}

		printAction(fmt.Sprintf("backing up %s → %s", s.Name, dest))

		if err := backup.Run(s, &backup.BackupRunConfig{
			Destination: dest,
			DryRun:      backupDryRun,
		}); err != nil {
			return fmt.Errorf("backup failed: %w", err)
		}

		if !backupDryRun {
			printSuccess(fmt.Sprintf("%s backed up", s.Name))
		}
		return nil
	},
}

func init() {
	backupCmd.Flags().BoolVar(&backupDryRun, "dry-run", false,
		"print what would be backed up without running restic")
	rootCmd.AddCommand(backupCmd)
}
