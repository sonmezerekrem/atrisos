package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/sonmezerekrem/atrisos/internal/compose"
	"github.com/sonmezerekrem/atrisos/internal/stack"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status [stack]",
	Short: "Show status of stacks",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 1 {
			return statusOne(args[0])
		}
		return statusAll()
	},
}

func statusAll() error {
	stacks, err := stack.Discover(rootCfg, rootReg)
	if err != nil {
		return fmt.Errorf("discovering stacks: %w", err)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tSTATUS\tDOMAINS")
	fmt.Fprintln(w, "────\t──────\t───────")
	for _, s := range stacks {
		statusStr := getStackStatus(s)
		domains := buildDomainList(s)
		fmt.Fprintf(w, "%s\t%s\t%s\n", s.Name, statusStr, domains)
	}
	return w.Flush()
}

func statusOne(name string) error {
	stacks, err := stack.Discover(rootCfg, rootReg)
	if err != nil {
		return fmt.Errorf("discovering stacks: %w", err)
	}

	s := stack.FindByName(stacks, name)
	if s == nil {
		printError(fmt.Sprintf("stack not found: %s", name))
		os.Exit(2)
	}

	fmt.Printf("Name:        %s\n", s.Name)
	fmt.Printf("Directory:   %s\n", s.Dir)
	if s.Config.Description != "" {
		fmt.Printf("Description: %s\n", s.Config.Description)
	}
	if len(s.Config.Tags) > 0 {
		fmt.Printf("Tags:        ")
		for i, t := range s.Config.Tags {
			if i > 0 {
				fmt.Print(", ")
			}
			fmt.Print(t)
		}
		fmt.Println()
	}
	fmt.Println()

	// Run podman compose ps to get container status.
	runner := compose.NewRunner(s.Dir, s.Name)
	composePath := stack.ComposeFile(s.Dir)
	out, err := runner.Output(composePath, "ps")
	if err != nil {
		fmt.Println("Status: (unable to query containers)")
	} else {
		fmt.Println("Containers:")
		fmt.Print(string(out))
	}

	return nil
}

// getStackStatus queries podman to get a brief status for a stack.
func getStackStatus(s *stack.Stack) string {
	runner := compose.NewRunner(s.Dir, s.Name)
	composePath := stack.ComposeFile(s.Dir)
	out, err := runner.Output(composePath, "ps", "--format", "json")
	if err != nil {
		return "unknown"
	}
	if len(out) == 0 || string(out) == "[]\n" || string(out) == "[]" {
		return "stopped"
	}
	return "running"
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
