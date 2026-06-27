package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sonmezerekrem/atrisos/internal/compose"
	"github.com/sonmezerekrem/atrisos/internal/stack"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var validateAll bool

var validateCmd = &cobra.Command{
	Use:   "validate <stack>",
	Short: "Validate a stack's configuration files",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if !validateAll && len(args) == 0 {
			return fmt.Errorf("specify a stack name or --all")
		}

		var stacks []*stack.Stack
		if validateAll {
			var err error
			stacks, err = stack.Discover(rootCfg, rootReg)
			if err != nil {
				return fmt.Errorf("discovering stacks: %w", err)
			}
		} else {
			var err error
			stacks, err = resolveStacks(args, false, "")
			if err != nil {
				return err
			}
		}

		allOK := true
		for _, s := range stacks {
			errs, warns := validateStack(s)
			if len(errs) > 0 {
				allOK = false
				printError(fmt.Sprintf("%s: %d error(s)", s.Name, len(errs)))
				for _, e := range errs {
					fmt.Fprintf(os.Stderr, "  ✗ %s\n", e)
				}
			} else {
				printSuccess(fmt.Sprintf("%s: valid", s.Name))
			}
			for _, w := range warns {
				printWarn(fmt.Sprintf("%s: %s", s.Name, w))
			}
		}

		if !allOK {
			os.Exit(3)
		}
		return nil
	},
}

// validateStack checks a stack's configuration and returns errors and warnings.
func validateStack(s *stack.Stack) (errors []string, warnings []string) {
	// 1. config.yml parses without error — we already loaded it, so just confirm.
	cfgPath := filepath.Join(s.Dir, "config.yml")
	if _, err := os.Stat(cfgPath); err == nil {
		data, err := os.ReadFile(cfgPath)
		if err != nil {
			errors = append(errors, fmt.Sprintf("cannot read config.yml: %v", err))
		} else {
			var cfg stack.StackConfig
			if err := yaml.Unmarshal(data, &cfg); err != nil {
				errors = append(errors, fmt.Sprintf("config.yml parse error: %v", err))
			}
		}
	}

	// 2. compose.yml exists and parses.
	composePath := stack.ComposeFile(s.Dir)
	if composePath == "" {
		errors = append(errors, "no compose.yml or docker-compose.yml found")
		return
	}
	data, err := os.ReadFile(composePath)
	if err != nil {
		errors = append(errors, fmt.Sprintf("cannot read compose file: %v", err))
		return
	}
	var composeDoc compose.ComposeDoc
	if err := yaml.Unmarshal(data, &composeDoc); err != nil {
		errors = append(errors, fmt.Sprintf("compose file parse error: %v", err))
		return
	}

	// 3. Each domain service exists in compose services.
	services := compose.GetServices(composeDoc)
	for i, d := range s.Config.Domains {
		if d.Service == "" {
			errors = append(errors, fmt.Sprintf("domains[%d]: service is required", i))
			continue
		}
		if services != nil {
			if _, exists := services[d.Service]; !exists {
				errors = append(errors, fmt.Sprintf("domains[%d]: service %q not found in compose services", i, d.Service))
			}
		}

		// 4. Host is non-empty and (for TLS) not empty.
		if d.Host == "" {
			errors = append(errors, fmt.Sprintf("domains[%d]: host is required", i))
		} else {
			tls := d.TLS
			if tls == "" {
				tls = "true"
			}
			if tls != "false" {
				// Warn if host looks like a bare IP.
				if isIPAddress(d.Host) {
					warnings = append(warnings, fmt.Sprintf("domains[%d]: host %q looks like an IP address — ACME cannot issue certs for IPs", i, d.Host))
				}
			}
		}

		// 5. Port is 1-65535.
		if d.Port < 1 || d.Port > 65535 {
			errors = append(errors, fmt.Sprintf("domains[%d]: port %d is out of range (1-65535)", i, d.Port))
		}
	}

	// 6. .env file exists (warning, not error).
	envPath := filepath.Join(s.Dir, ".env")
	if _, err := os.Stat(envPath); os.IsNotExist(err) {
		warnings = append(warnings, ".env file not found (stack may fail if compose.yml references variables)")
	}

	// 7. backup.schedule is valid cron if backup.enabled.
	if s.Config.Backup.Enabled && s.Config.Backup.Schedule == "" {
		errors = append(errors, "backup.enabled is true but backup.schedule is empty")
	} else if s.Config.Backup.Enabled && s.Config.Backup.Schedule != "" {
		if !isValidCron(s.Config.Backup.Schedule) {
			errors = append(errors, fmt.Sprintf("backup.schedule %q is not a valid 5-field cron expression", s.Config.Backup.Schedule))
		}
	}

	return
}

// isIPAddress checks if a string looks like an IP address.
func isIPAddress(host string) bool {
	// Simple check: if all parts are digits and dots, it's an IPv4.
	parts := strings.Split(host, ".")
	if len(parts) == 4 {
		allDigits := true
		for _, p := range parts {
			for _, c := range p {
				if c < '0' || c > '9' {
					allDigits = false
					break
				}
			}
		}
		return allDigits
	}
	// IPv6: contains colons.
	return strings.Contains(host, ":")
}

// isValidCron checks that s has exactly 5 space-separated fields.
func isValidCron(s string) bool {
	fields := strings.Fields(s)
	return len(fields) == 5
}

func init() {
	validateCmd.Flags().BoolVar(&validateAll, "all", false, "validate all discovered stacks")
	rootCmd.AddCommand(validateCmd)
}
