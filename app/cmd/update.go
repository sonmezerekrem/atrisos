package cmd

import (
	"fmt"

	"github.com/sonmezerekrem/atrisos/internal/compose"
	"github.com/sonmezerekrem/atrisos/internal/stack"
	"github.com/spf13/cobra"
)

var (
	updatePull   bool
	updateNoPull bool
	updateAll    bool
	updateTag    string
)

var updateCmd = &cobra.Command{
	Use:   "update <stack>",
	Short: "Pull latest images and recreate containers",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if !updateAll && updateTag == "" && len(args) == 0 {
			return fmt.Errorf("specify a stack name, --all, or --tag <tag>")
		}

		stacks, err := resolveStacks(args, updateAll, updateTag)
		if err != nil {
			return err
		}

		var lastErr error
		for _, s := range stacks {
			if err := updateOne(s); err != nil {
				printError(fmt.Sprintf("%s: %v", s.Name, err))
				lastErr = err
				continue
			}
			printSuccess(fmt.Sprintf("%s updated", s.Name))
		}
		return lastErr
	},
}

func updateOne(s *stack.Stack) error {
	printAction(fmt.Sprintf("updating %s...", s.Name))

	doc, err := compose.LoadAndMerge(s.Dir, s.Config)
	if err != nil {
		return fmt.Errorf("loading compose: %w", err)
	}

	runner := compose.NewRunner(s.Dir, s.Name)

	// Pull images unless --no-pull is set.
	if !updateNoPull {
		if err := runner.RunMerged(doc, "pull"); err != nil {
			printWarn(fmt.Sprintf("pull failed for %s: %v", s.Name, err))
		}
	}

	// Recreate containers.
	return runner.RunMerged(doc, "up", "-d")
}

func init() {
	updateCmd.Flags().BoolVar(&updatePull, "pull", false,
		"explicit pull before recreate (default behavior)")
	updateCmd.Flags().BoolVar(&updateNoPull, "no-pull", false,
		"recreate without pulling (e.g. to apply .env changes)")
	updateCmd.Flags().BoolVar(&updateAll, "all", false,
		"update all discovered stacks")
	updateCmd.Flags().StringVar(&updateTag, "tag", "", "update all stacks with this tag")
	rootCmd.AddCommand(updateCmd)
}

