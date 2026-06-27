package cmd

import (
	"fmt"

	"github.com/sonmezerekrem/atrisos/internal/compose"
	"github.com/sonmezerekrem/atrisos/internal/stack"
	"github.com/spf13/cobra"
)

var psCmd = &cobra.Command{
	Use:   "ps <stack>",
	Short: "Show containers in a stack",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		stacks, err := resolveStacks(args, false, "")
		if err != nil {
			return err
		}
		s := stacks[0]
		return psOne(s)
	},
}

func psOne(s *stack.Stack) error {
	doc, err := compose.LoadAndMerge(s.Dir, s.Config)
	if err != nil {
		return fmt.Errorf("loading compose: %w", err)
	}
	runner := compose.NewRunner(s.Dir, s.Name)
	return runner.RunMerged(doc, "ps")
}

func init() {
	rootCmd.AddCommand(psCmd)
}
