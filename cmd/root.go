package cmd

import (
	"fmt"
	"os"
	"runtime"

	"github.com/sonmezerekrem/atrisos/internal/config"
	"github.com/sonmezerekrem/atrisos/internal/podman"
	"github.com/sonmezerekrem/atrisos/internal/registry"
	"github.com/sonmezerekrem/atrisos/tui"
	"github.com/spf13/cobra"
)

// Package-level state populated by PersistentPreRunE and used by subcommands.
var (
	rootCfg     *config.Config
	rootReg     *registry.Registry
	verboseFlag bool
	noColorFlag bool
	noEmojiFlag bool
	rootFlag    string
	cfgFileFlag string
)

// ANSI color codes.
const (
	colorReset  = "\033[0m"
	colorGreen  = "\033[32m"
	colorRed    = "\033[31m"
	colorYellow = "\033[33m"
	colorCyan   = "\033[36m"
)

// Output helpers.

func printSuccess(msg string) {
	if noEmojiFlag {
		fmt.Println("[ok]", msg)
	} else if noColorFlag {
		fmt.Println("✓", msg)
	} else {
		fmt.Println(colorGreen+"✓"+colorReset, msg)
	}
}

func printError(msg string) {
	if noEmojiFlag {
		fmt.Fprintln(os.Stderr, "[err]", msg)
	} else if noColorFlag {
		fmt.Fprintln(os.Stderr, "✗", msg)
	} else {
		fmt.Fprintln(os.Stderr, colorRed+"✗"+colorReset, msg)
	}
}

func printAction(msg string) {
	if noEmojiFlag {
		fmt.Println("[->]", msg)
	} else if noColorFlag {
		fmt.Println("→", msg)
	} else {
		fmt.Println(colorCyan+"→"+colorReset, msg)
	}
}

func printWarn(msg string) {
	if noEmojiFlag {
		fmt.Fprintln(os.Stderr, "[warn]", msg)
	} else if noColorFlag {
		fmt.Fprintln(os.Stderr, "⚠", msg)
	} else {
		fmt.Fprintln(os.Stderr, colorYellow+"⚠"+colorReset, msg)
	}
}

// exitWithCode prints an error and exits with the given code.
func exitWithCode(code int, format string, args ...interface{}) {
	printError(fmt.Sprintf(format, args...))
	os.Exit(code)
}

// rootCmd is the top-level cobra command.
var rootCmd = &cobra.Command{
	Use:   "atrisos",
	Short: "Manage Podman Compose stacks with automatic Traefik routing",
	Long: `atrisos — CLI + TUI tool for managing Podman Compose stacks with
automatic Traefik routing. Run with no arguments to launch the TUI.`,
	SilenceErrors: true,
	SilenceUsage:  true,
	RunE: func(cmd *cobra.Command, args []string) error {
		// No subcommand → launch TUI.
		return tui.Run(rootCfg, rootReg)
	},
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Skip pre-run for commands that don't need config/runtime.
		if cmd.Annotations["skipPreRun"] == "true" {
			return nil
		}

		// Apply environment variable overrides (flags take precedence via cobra).
		if os.Getenv("ATRISOS_NO_COLOR") != "" || os.Getenv("NO_COLOR") != "" {
			noColorFlag = true
		}
		if os.Getenv("ATRISOS_VERBOSE") != "" {
			verboseFlag = true
		}

		// Load config.
		cfgPath := cfgFileFlag
		if envCfg := os.Getenv("ATRISOS_CONFIG"); envCfg != "" && cfgPath == "" {
			cfgPath = envCfg
		}
		cfg, err := config.Load(cfgPath)
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		// Apply --root flag or ATRISOS_ROOT env.
		if rootFlag != "" {
			cfg.StacksRoot = rootFlag
		} else if envRoot := os.Getenv("ATRISOS_ROOT"); envRoot != "" {
			cfg.StacksRoot = envRoot
		}

		// Apply --no-color / --no-emoji from config if not set via flag.
		if cfg.Output.NoColor {
			noColorFlag = true
		}
		if cfg.Output.NoEmoji {
			noEmojiFlag = true
		}

		rootCfg = cfg

		// Load registry.
		reg, err := registry.Load(config.Dir())
		if err != nil {
			return fmt.Errorf("loading registry: %w", err)
		}
		rootReg = reg

		// On macOS, ensure the Podman machine is running.
		if runtime.GOOS == "darwin" {
			if err := podman.EnsureMachine(cfg.Podman.MachineName); err != nil {
				printWarn(fmt.Sprintf("podman machine: %v", err))
			}
		}

		return nil
	},
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFileFlag, "config", "",
		"path to global config file (default: ~/.config/atrisos/config.yml)")
	rootCmd.PersistentFlags().StringVar(&rootFlag, "root", "",
		"override the stacks root directory for this invocation")
	rootCmd.PersistentFlags().BoolVarP(&verboseFlag, "verbose", "v", false,
		"show verbose output including raw Podman commands")
	rootCmd.PersistentFlags().BoolVar(&noColorFlag, "no-color", false,
		"disable color output")
	rootCmd.PersistentFlags().BoolVar(&noEmojiFlag, "no-emoji", false,
		"disable emoji status indicators")
}
