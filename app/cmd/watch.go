package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/sonmezerekrem/atrisos/app/internal/compose"
	"github.com/sonmezerekrem/atrisos/app/internal/stack"
	"github.com/spf13/cobra"
)

var watchCmd = &cobra.Command{
	Use:   "watch <stack>",
	Short: "Watch a stack directory and re-apply on changes",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		stacks, err := resolveStacks(args, false, "")
		if err != nil {
			return err
		}
		s := stacks[0]
		return watchStack(s)
	},
}

func watchStack(s *stack.Stack) error {
	// Start the stack first.
	printAction(fmt.Sprintf("starting %s before watching...", s.Name))
	if err := upOne(s); err != nil {
		return fmt.Errorf("initial start failed: %w", err)
	}

	// Create watcher.
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("creating watcher: %w", err)
	}
	defer watcher.Close()

	// Watch specific files in the stack directory.
	watchFiles := []string{
		"compose.yml", "docker-compose.yml", "compose.override.yml",
		".env", "config.yml",
	}
	for _, f := range watchFiles {
		path := s.Dir + "/" + f
		if _, err := os.Stat(path); err == nil {
			if err := watcher.Add(path); err != nil {
				printWarn(fmt.Sprintf("cannot watch %s: %v", f, err))
			}
		}
	}

	printSuccess(fmt.Sprintf("watching %s (Ctrl+C to stop)", s.Dir))

	// Signal handling for graceful stop.
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Debounce state.
	var (
		lastEvent time.Time
		pending   bool
	)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return nil
			}
			if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) || event.Has(fsnotify.Remove) {
				printAction(fmt.Sprintf("change detected: %s", event.Name))
				lastEvent = time.Now()
				pending = true
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
			printWarn(fmt.Sprintf("watch error: %v", err))

		case <-ticker.C:
			if pending && time.Since(lastEvent) >= 500*time.Millisecond {
				pending = false
				printAction("re-applying changes...")
				// Reload stack config in case config.yml changed.
				reloaded, err := stack.LoadStack(s.Dir)
				if err != nil {
					printWarn(fmt.Sprintf("reload failed: %v", err))
					continue
				}
				doc, err := compose.LoadAndMerge(reloaded.Dir, reloaded.Config)
				if err != nil {
					printWarn(fmt.Sprintf("compose merge failed: %v", err))
					continue
				}
				runner := compose.NewRunner(reloaded.Dir, reloaded.Name)
				if err := runner.RunMerged(doc, "up", "-d"); err != nil {
					printWarn(fmt.Sprintf("compose up failed: %v", err))
				} else {
					printSuccess("changes applied")
				}
			}

		case <-sigChan:
			printAction(fmt.Sprintf("stopping watch (stack %s still running)", s.Name))
			return nil
		}
	}
}

func init() {
	rootCmd.AddCommand(watchCmd)
}
