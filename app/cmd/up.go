package cmd

import (
	"fmt"
	"os"

	"github.com/sonmezerekrem/atrisos/app/internal/compose"
	"github.com/sonmezerekrem/atrisos/app/internal/scheduler"
	"github.com/sonmezerekrem/atrisos/app/internal/stack"
	"github.com/sonmezerekrem/atrisos/app/internal/traefik"
	"github.com/spf13/cobra"
)

var (
	upPull  bool
	upBuild bool
	upAll   bool
	upTag   string
)

var upCmd = &cobra.Command{
	Use:   "up <stack>",
	Short: "Start a stack (and Traefik if not running)",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if !upAll && upTag == "" && len(args) == 0 {
			return fmt.Errorf("specify a stack name, --all, or --tag <tag>")
		}

		stacks, err := resolveStacks(args, upAll, upTag)
		if err != nil {
			return err
		}

		// Ensure Traefik is started.
		mgr := traefik.NewManager(rootCfg)
		printAction("ensuring Traefik is running...")
		if err := mgr.EnsureStarted(); err != nil {
			printError(fmt.Sprintf("Traefik: %v", err))
			os.Exit(5)
		}

		atrisosExe, _ := os.Executable()

		var lastErr error
		for _, s := range stacks {
			if err := upOne(s); err != nil {
				printError(fmt.Sprintf("%s: %v", s.Name, err))
				lastErr = err
				continue
			}
			printSuccess(fmt.Sprintf("%s is up", s.Name))

			// Wire scheduler units after successful start.
			if s.Config.AutoStart {
				if err := scheduler.InstallAutoStart(s, atrisosExe); err != nil {
					printWarn(fmt.Sprintf("auto-start scheduler: %v", err))
				}
			}
			if s.Config.Backup.Enabled {
				if err := scheduler.InstallBackupTimer(s, atrisosExe); err != nil {
					printWarn(fmt.Sprintf("backup scheduler: %v", err))
				}
			}
		}
		return lastErr
	},
}

func upOne(s *stack.Stack) error {
	printAction(fmt.Sprintf("starting %s...", s.Name))

	doc, err := compose.LoadAndMerge(s.Dir, s.Config)
	if err != nil {
		return fmt.Errorf("loading compose: %w", err)
	}

	runner := compose.NewRunner(s.Dir, s.Name)
	args := []string{"up", "-d"}
	if upPull {
		args = append(args, "--pull", "always")
	}
	if upBuild {
		args = append(args, "--build")
	}

	return runner.RunMerged(doc, args...)
}

func init() {
	upCmd.Flags().BoolVar(&upPull, "pull", false, "pull latest images before starting")
	upCmd.Flags().BoolVar(&upBuild, "build", false, "build images from Dockerfile before starting")
	upCmd.Flags().BoolVar(&upAll, "all", false, "start all discovered stacks")
	upCmd.Flags().StringVar(&upTag, "tag", "", "start all stacks with this tag")
	rootCmd.AddCommand(upCmd)
}

// resolveStacks returns the target stack list based on args / --all / --tag.
func resolveStacks(args []string, all bool, tag string) ([]*stack.Stack, error) {
	if all || tag != "" {
		discovered, err := stack.Discover(rootCfg, rootReg)
		if err != nil {
			return nil, fmt.Errorf("discovering stacks: %w", err)
		}
		if tag != "" {
			discovered = stack.FilterByTag(discovered, tag)
		}
		return discovered, nil
	}

	if len(args) == 0 {
		return nil, fmt.Errorf("no stack specified")
	}

	discovered, err := stack.Discover(rootCfg, rootReg)
	if err != nil {
		return nil, fmt.Errorf("discovering stacks: %w", err)
	}
	name := args[0]
	s := stack.FindByName(discovered, name)
	if s == nil {
		printError(fmt.Sprintf("stack not found: %s", name))
		os.Exit(2)
	}
	return []*stack.Stack{s}, nil
}
