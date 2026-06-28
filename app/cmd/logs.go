package cmd

import (
	"fmt"
	"os"
	"strconv"

	"github.com/sonmezerekrem/atrisos/internal/compose"
	"github.com/sonmezerekrem/atrisos/internal/stack"
	"github.com/spf13/cobra"
)

var (
	logsService    string
	logsLines      int
	logsFollow     bool
	logsNoFollow   bool
	logsTimestamps bool
)

var logsCmd = &cobra.Command{
	Use:   "logs <stack>",
	Short: "Tail logs from a stack",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		stacks, err := resolveStacks(args, false, "")
		if err != nil {
			return err
		}
		s := stacks[0]
		return logsOne(s, cmd)
	},
}

func logsOne(s *stack.Stack, cmd *cobra.Command) error {
	doc, err := compose.LoadAndMerge(s.Dir, s.Config)
	if err != nil {
		return fmt.Errorf("loading compose: %w", err)
	}

	runner := compose.NewRunner(s.Dir, s.Name)
	logsArgs := []string{"logs"}

	// Determine follow mode: default to true if stdout is a TTY.
	follow := logsFollow
	if !logsNoFollow && !cmd.Flags().Changed("follow") {
		// Auto-detect TTY.
		fi, err := os.Stdout.Stat()
		if err == nil && (fi.Mode()&os.ModeCharDevice) != 0 {
			follow = true
		}
	}
	if logsNoFollow {
		follow = false
	}

	if follow {
		logsArgs = append(logsArgs, "-f")
	}
	if logsLines > 0 {
		logsArgs = append(logsArgs, "--tail", strconv.Itoa(logsLines))
	}
	if logsTimestamps {
		logsArgs = append(logsArgs, "--timestamps")
	}
	if logsService != "" {
		logsArgs = append(logsArgs, logsService)
	}

	return runner.RunMerged(doc, logsArgs...)
}

func init() {
	logsCmd.Flags().StringVar(&logsService, "service", "",
		"show logs for a specific service")
	logsCmd.Flags().IntVar(&logsLines, "lines", 50,
		"number of lines to show from the end")
	logsCmd.Flags().BoolVar(&logsFollow, "follow", false,
		"keep streaming (default: true if TTY)")
	logsCmd.Flags().BoolVar(&logsNoFollow, "no-follow", false,
		"print and exit")
	logsCmd.Flags().BoolVar(&logsTimestamps, "timestamps", false,
		"include timestamps")
	rootCmd.AddCommand(logsCmd)
}
