package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/sonmezerekrem/atrisos/internal/stack"
	"github.com/spf13/cobra"
)

var (
	listFormat string
	listTag    string
)

// listStackInfo is the JSON-serializable summary of a stack.
type listStackInfo struct {
	Name        string   `json:"name"`
	Dir         string   `json:"dir"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
	Domains     []string `json:"domains"`
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all discovered stacks",
	RunE: func(cmd *cobra.Command, args []string) error {
		stacks, err := stack.Discover(rootCfg, rootReg)
		if err != nil {
			return fmt.Errorf("discovering stacks: %w", err)
		}

		if listTag != "" {
			stacks = stack.FilterByTag(stacks, listTag)
		}

		switch listFormat {
		case "json":
			return printListJSON(stacks)
		case "plain":
			for _, s := range stacks {
				fmt.Println(s.Name)
			}
		default: // table
			return printListTable(stacks)
		}
		return nil
	},
}

func printListTable(stacks []*stack.Stack) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tDESCRIPTION\tDOMAINS\tTAGS")
	fmt.Fprintln(w, "────\t───────────\t───────\t────")
	for _, s := range stacks {
		domains := buildDomainList(s)
		tags := strings.Join(s.Config.Tags, ",")
		desc := s.Config.Description
		if len(desc) > 40 {
			desc = desc[:37] + "..."
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", s.Name, desc, domains, tags)
	}
	return w.Flush()
}

func printListJSON(stacks []*stack.Stack) error {
	var infos []listStackInfo
	for _, s := range stacks {
		var domains []string
		for _, d := range s.Config.Domains {
			domains = append(domains, d.Host)
		}
		infos = append(infos, listStackInfo{
			Name:        s.Name,
			Dir:         s.Dir,
			Description: s.Config.Description,
			Tags:        s.Config.Tags,
			Domains:     domains,
		})
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(infos)
}

func buildDomainList(s *stack.Stack) string {
	var hosts []string
	for _, d := range s.Config.Domains {
		hosts = append(hosts, d.Host)
	}
	if len(hosts) == 0 {
		return "-"
	}
	return strings.Join(hosts, ",")
}

func init() {
	listCmd.Flags().StringVar(&listFormat, "format", "table",
		"output format: table, json, plain")
	listCmd.Flags().StringVar(&listTag, "tag", "",
		"filter stacks by tag")
	rootCmd.AddCommand(listCmd)
}
