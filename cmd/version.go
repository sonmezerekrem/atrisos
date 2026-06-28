package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"

	"github.com/sonmezerekrem/atrisos/internal/selfupdate"
)

// Version is set at build time via -ldflags.
var Version = "dev"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Annotations: map[string]string{"skipPreRun": "true"},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("atrisos %s\n", Version)
		fmt.Printf("OS/Arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
		fmt.Printf("Go:      %s\n", runtime.Version())
		latest := selfupdate.LatestVersion()
		if latest != "" && Version != "dev" {
			if latest != Version {
				fmt.Printf("Latest:  %s  → run `atrisos self-update` to upgrade\n", latest)
			} else {
				fmt.Printf("Latest:  %s (up to date)\n", latest)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
