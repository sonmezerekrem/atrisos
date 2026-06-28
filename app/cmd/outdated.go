package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/sonmezerekrem/atrisos/internal/outdated"
	"github.com/sonmezerekrem/atrisos/internal/stack"
	"github.com/spf13/cobra"
)

var outdatedCmd = &cobra.Command{
	Use:   "outdated [stack]",
	Short: "Check for newer image versions available in the registry",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 1 {
			return outdatedOne(args[0])
		}
		return outdatedAll()
	},
}

func outdatedOne(name string) error {
	discovered, err := stack.Discover(rootCfg, rootReg)
	if err != nil {
		return fmt.Errorf("discovering stacks: %w", err)
	}
	s := stack.FindByName(discovered, name)
	if s == nil {
		printError(fmt.Sprintf("stack not found: %s", name))
		os.Exit(2)
	}

	printAction(fmt.Sprintf("checking images for %s...", s.Name))

	updates, err := outdated.CheckStack(s)
	if err != nil {
		return err
	}

	if len(updates) == 0 {
		printSuccess(fmt.Sprintf("%s is up to date", s.Name))
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "STACK\tSERVICE\tIMAGE\tCURRENT\tAVAILABLE")
	for _, u := range updates {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			s.Name, u.Service, u.Image, u.Current, u.Available)
	}
	return w.Flush()
}

func outdatedAll() error {
	discovered, err := stack.Discover(rootCfg, rootReg)
	if err != nil {
		return fmt.Errorf("discovering stacks: %w", err)
	}

	printAction("checking all stacks for image updates...")

	allUpdates, err := outdated.CheckAll(discovered)
	if err != nil {
		return err
	}

	if len(allUpdates) == 0 {
		printSuccess("all stacks are up to date")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "STACK\tSERVICE\tIMAGE\tCURRENT\tAVAILABLE")
	for _, s := range discovered {
		for _, u := range allUpdates[s.Name] {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
				s.Name, u.Service, u.Image, u.Current, u.Available)
		}
	}
	return w.Flush()
}

func init() {
	rootCmd.AddCommand(outdatedCmd)
}
