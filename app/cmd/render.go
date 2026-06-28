package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/sonmezerekrem/atrisos/app/internal/compose"
	"github.com/sonmezerekrem/atrisos/app/internal/stack"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var renderDiff bool

var renderCmd = &cobra.Command{
	Use:   "render <stack>",
	Short: "Print the merged compose document that atrisos would pass to podman compose",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		stacks, err := resolveStacks(args, false, "")
		if err != nil {
			return err
		}
		s := stacks[0]

		// Load merged document.
		doc, err := compose.LoadAndMerge(s.Dir, s.Config)
		if err != nil {
			return fmt.Errorf("loading compose: %w", err)
		}

		mergedYAML, err := yaml.Marshal(doc)
		if err != nil {
			return fmt.Errorf("marshaling merged compose: %w", err)
		}

		if !renderDiff {
			fmt.Print(string(mergedYAML))
			return nil
		}

		// Diff mode: compare original vs merged.
		composePath := stack.ComposeFile(s.Dir)
		origData, err := os.ReadFile(composePath)
		if err != nil {
			return fmt.Errorf("reading original compose file: %w", err)
		}

		fmt.Println("--- original compose.yml")
		fmt.Println("+++ merged (with Traefik labels)")
		fmt.Println()
		printSimpleDiff(string(origData), string(mergedYAML))
		return nil
	},
}

// printSimpleDiff prints a simplified diff between two YAML strings.
// Lines present only in 'a' are prefixed with '-'.
// Lines present only in 'b' are prefixed with '+'.
// Lines in both are prefixed with ' '.
func printSimpleDiff(a, b string) {
	aLines := strings.Split(a, "\n")
	bLines := strings.Split(b, "\n")

	// Build a set of original lines.
	aSet := map[string]int{}
	for _, l := range aLines {
		aSet[l]++
	}

	// Build a set of merged lines.
	bSet := map[string]int{}
	for _, l := range bLines {
		bSet[l]++
	}

	// Print removed lines (in a but not b).
	for _, l := range aLines {
		if bSet[l] == 0 {
			if noColorFlag {
				fmt.Println("- " + l)
			} else {
				fmt.Println("\033[31m- " + l + "\033[0m")
			}
		} else {
			bSet[l]-- // consume one occurrence
		}
	}

	// Reset bSet for printing additions.
	bSet = map[string]int{}
	for _, l := range bLines {
		bSet[l]++
	}
	aSet2 := map[string]int{}
	for _, l := range aLines {
		aSet2[l]++
	}

	// Print added lines (in b but not a).
	for _, l := range bLines {
		if aSet2[l] == 0 {
			if noColorFlag {
				fmt.Println("+ " + l)
			} else {
				fmt.Println("\033[32m+ " + l + "\033[0m")
			}
		} else {
			aSet2[l]--
		}
	}
}

func init() {
	renderCmd.Flags().BoolVar(&renderDiff, "diff", false,
		"show diff vs original compose.yml")
	rootCmd.AddCommand(renderCmd)
}
