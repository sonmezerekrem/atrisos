package cmd

import (
	"fmt"

	"github.com/sonmezerekrem/atrisos/internal/stack"
	"github.com/spf13/cobra"
)

var restartCmd = &cobra.Command{
	Use:   "restart <stack>",
	Short: "Stop and start a stack",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		stacks, err := resolveStacks(args, false, "")
		if err != nil {
			return err
		}
		s := stacks[0]
		return restartOne(s)
	},
}

func restartOne(s *stack.Stack) error {
	if err := downOne(s); err != nil {
		return fmt.Errorf("stopping %s: %w", s.Name, err)
	}
	if err := upOne(s); err != nil {
		return fmt.Errorf("starting %s: %w", s.Name, err)
	}
	printSuccess(fmt.Sprintf("%s restarted", s.Name))
	return nil
}

func init() {
	rootCmd.AddCommand(restartCmd)
}
