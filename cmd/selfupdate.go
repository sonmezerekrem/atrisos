package cmd

import (
	"fmt"

	"github.com/sonmezerekrem/atrisos/internal/selfupdate"
	"github.com/spf13/cobra"
)

var selfUpdateVersion string

var selfUpdateCmd = &cobra.Command{
	Use:         "self-update",
	Short:       "Download and install the latest atrisos release",
	Annotations: map[string]string{"skipPreRun": "true"},
	RunE: func(cmd *cobra.Command, args []string) error {
		target := selfUpdateVersion
		if target == "" {
			printAction("fetching latest version...")
			target = selfupdate.LatestVersion()
			if target == "" {
				return fmt.Errorf("could not determine latest version (check your internet connection)")
			}
			printAction(fmt.Sprintf("latest version is %s", target))
		}

		if err := selfupdate.Update(target); err != nil {
			return fmt.Errorf("self-update: %w", err)
		}
		return nil
	},
}

func init() {
	selfUpdateCmd.Flags().StringVar(&selfUpdateVersion, "version", "",
		"install a specific version (e.g. v0.3.0); default: latest")
	rootCmd.AddCommand(selfUpdateCmd)
}
