package cmd

import (
	"fmt"

	"github.com/sonmezerekrem/atrisos/app/internal/compose"
	"github.com/sonmezerekrem/atrisos/app/internal/scheduler"
	"github.com/sonmezerekrem/atrisos/app/internal/stack"
	"github.com/spf13/cobra"
)

var (
	downVolumes bool
	downAll     bool
	downTag     string
)

var downCmd = &cobra.Command{
	Use:   "down <stack>",
	Short: "Stop a stack and remove its containers",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if !downAll && downTag == "" && len(args) == 0 {
			return fmt.Errorf("specify a stack name, --all, or --tag <tag>")
		}

		stacks, err := resolveStacks(args, downAll, downTag)
		if err != nil {
			return err
		}

		var lastErr error
		for _, s := range stacks {
			if err := downOne(s); err != nil {
				printError(fmt.Sprintf("%s: %v", s.Name, err))
				lastErr = err
				continue
			}
			printSuccess(fmt.Sprintf("%s is down", s.Name))

			// Remove scheduler units after stopping.
			if err := scheduler.RemoveAutoStart(s); err != nil {
				printWarn(fmt.Sprintf("removing auto-start scheduler: %v", err))
			}
			if err := scheduler.RemoveBackupTimer(s); err != nil {
				printWarn(fmt.Sprintf("removing backup scheduler: %v", err))
			}
		}
		return lastErr
	},
}

func downOne(s *stack.Stack) error {
	printAction(fmt.Sprintf("stopping %s...", s.Name))

	doc, err := compose.LoadAndMerge(s.Dir, s.Config)
	if err != nil {
		return fmt.Errorf("loading compose: %w", err)
	}

	runner := compose.NewRunner(s.Dir, s.Name)
	args := []string{"down"}
	if downVolumes {
		args = append(args, "-v")
	}

	return runner.RunMerged(doc, args...)
}

func init() {
	downCmd.Flags().BoolVar(&downVolumes, "volumes", false,
		"also remove named volumes (destructive)")
	downCmd.Flags().BoolVar(&downAll, "all", false, "stop all discovered stacks")
	downCmd.Flags().StringVar(&downTag, "tag", "", "stop all stacks with this tag")
	rootCmd.AddCommand(downCmd)
}
