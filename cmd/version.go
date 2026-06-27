package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
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
		// TODO: check GitHub for latest version (background, cached 24h)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
