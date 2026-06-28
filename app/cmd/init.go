package cmd

import (
	"bufio"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"text/template"

	"github.com/sonmezerekrem/atrisos/internal/config"
	"github.com/sonmezerekrem/atrisos/internal/templates"
	"github.com/spf13/cobra"
)

var (
	initTemplateName     string
	initDir              string
	initListTemplates    bool
	initRefreshTemplates bool
)

var initCmd = &cobra.Command{
	Use:         "init [name]",
	Short:       "Create a new stack from a template",
	Annotations: map[string]string{"skipPreRun": "true"},
	RunE:        runInit,
}

func runInit(cmd *cobra.Command, args []string) error {
	if initListTemplates {
		manifest, err := templates.LoadManifest()
		if err != nil {
			return fmt.Errorf("loading templates: %w", err)
		}
		fmt.Printf("%-22s %s\n", "NAME", "DESCRIPTION")
		fmt.Printf("%-22s %s\n", strings.Repeat("-", 22), strings.Repeat("-", 40))
		for _, t := range manifest.Templates {
			fmt.Printf("%-22s %s\n", t.Name, t.Description)
		}
		return nil
	}

	if initRefreshTemplates {
		fmt.Println("→ Refreshing template cache...")
		if err := templates.RefreshCache(); err != nil {
			return fmt.Errorf("refreshing templates: %w", err)
		}
		fmt.Println("✓ Template cache refreshed")
		return nil
	}

	reader := bufio.NewReader(os.Stdin)

	// Determine stack name.
	name := ""
	if len(args) > 0 {
		name = strings.TrimSpace(args[0])
	}
	if name == "" {
		fmt.Print("Stack name: ")
		line, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("reading name: %w", err)
		}
		name = strings.TrimSpace(line)
	}
	if name == "" {
		return fmt.Errorf("stack name is required")
	}

	dirName := slug(name)

	// Load manifest.
	manifest, err := templates.LoadManifest()
	if err != nil {
		return fmt.Errorf("loading templates: %w", err)
	}

	// Pick template.
	chosenTemplate := initTemplateName
	if chosenTemplate == "" {
		fmt.Println("\nAvailable templates:")
		for i, t := range manifest.Templates {
			fmt.Printf("  %d. %s — %s\n", i+1, t.Display, t.Description)
		}
		fmt.Print("\nSelect template [1]: ")
		line, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("reading template choice: %w", err)
		}
		line = strings.TrimSpace(line)
		idx := 1
		if line != "" {
			parsed, err := strconv.Atoi(line)
			if err != nil || parsed < 1 || parsed > len(manifest.Templates) {
				return fmt.Errorf("invalid choice %q: must be 1–%d", line, len(manifest.Templates))
			}
			idx = parsed
		}
		chosenTemplate = manifest.Templates[idx-1].Name
	}

	// Load template metadata.
	meta, err := templates.LoadTemplateMeta(chosenTemplate)
	if err != nil {
		return fmt.Errorf("loading template %q: %w", chosenTemplate, err)
	}

	vars := map[string]interface{}{
		"Name":    name,
		"DirName": dirName,
	}

	fmt.Println()
	for _, p := range meta.Prompts {
		// Pre-generate value if requested (done once, before the retry loop).
		generated := ""
		if p.Generate != "" {
			switch p.Generate {
			case "random_password":
				generated = randomPassword(24)
			case "traefik_me_domain":
				generated = fmt.Sprintf("%s-%s.traefik.me", dirName, randomHex(4))
			}
		}

		for {
			promptStr := p.Label
			switch p.Type {
			case "bool":
				promptStr += " (y/n)"
				if p.Default != "" {
					promptStr += fmt.Sprintf(" [%s]", p.Default)
				}
			case "select":
				if len(p.Options) > 0 {
					promptStr += fmt.Sprintf(" (%s)", strings.Join(p.Options, "/"))
				}
				if p.Default != "" {
					promptStr += fmt.Sprintf(" [%s]", p.Default)
				}
			default:
				if generated != "" {
					promptStr += fmt.Sprintf(" [auto: %s]", generated)
				} else if p.Default != "" {
					promptStr += fmt.Sprintf(" [%s]", p.Default)
				} else if !p.Required {
					promptStr += " (optional)"
				}
			}
			fmt.Print(promptStr + ": ")

			line, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("reading prompt %q: %w", p.Name, err)
			}
			line = strings.TrimSpace(line)

			if line == "" {
				if generated != "" {
					line = generated
				} else if p.Default != "" {
					line = p.Default
				} else if p.Required {
					fmt.Println("  (required — please enter a value)")
					continue
				}
			}

			switch p.Type {
			case "bool":
				lower := strings.ToLower(line)
				vars[p.Name] = lower == "y" || lower == "yes" || lower == "true" || lower == "1"

			case "int":
				if line == "" {
					vars[p.Name] = ""
				} else {
					if _, err := strconv.Atoi(line); err != nil {
						fmt.Println("  (must be an integer)")
						continue
					}
					vars[p.Name] = line
				}

			case "select":
				if len(p.Options) > 0 {
					valid := false
					for _, opt := range p.Options {
						if line == opt {
							valid = true
							break
						}
					}
					if !valid && line != "" {
						fmt.Printf("  (must be one of: %s)\n", strings.Join(p.Options, ", "))
						continue
					}
				}
				vars[p.Name] = line

			default:
				vars[p.Name] = line
			}
			break
		}
	}

	// Derive tls from domain: free/wildcard DNS services don't support real certs.
	if domain, ok := vars["domain"].(string); ok {
		if domain == "" || strings.HasSuffix(domain, ".traefik.me") ||
			strings.HasSuffix(domain, ".nip.io") || strings.HasSuffix(domain, ".sslip.io") {
			vars["tls"] = "false"
		} else {
			vars["tls"] = "true"
		}
	}

	// Determine output directory.
	outDir := initDir
	if outDir == "" {
		cfg, err := config.Load("")
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}
		outDir = filepath.Join(cfg.StacksRoot, dirName)
	}

	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return fmt.Errorf("creating directory %s: %w", outDir, err)
	}

	tmplFiles, err := templates.TemplateFiles(chosenTemplate)
	if err != nil {
		return fmt.Errorf("listing template files for %q: %w", chosenTemplate, err)
	}

	for _, fname := range tmplFiles {
		content, err := templates.ReadTemplateFile(chosenTemplate, fname)
		if err != nil {
			return fmt.Errorf("reading template file %s: %w", fname, err)
		}

		tmpl, err := template.New(fname).Parse(content)
		if err != nil {
			return fmt.Errorf("parsing template %s: %w", fname, err)
		}

		outName := strings.TrimSuffix(fname, ".tmpl")
		outPath := filepath.Join(outDir, outName)

		f, err := os.Create(outPath)
		if err != nil {
			return fmt.Errorf("creating %s: %w", outPath, err)
		}

		if execErr := tmpl.Execute(f, vars); execErr != nil {
			f.Close()
			return fmt.Errorf("rendering %s: %w", fname, execErr)
		}
		f.Close()
	}

	fmt.Printf("\n✓ Stack created at %s\n\n", outDir)
	fmt.Println("Next steps:")
	fmt.Printf("  1. Review %s/.env — passwords have been auto-generated\n", outDir)
	fmt.Printf("  2. Run: atrisos up %s\n", dirName)

	return nil
}

const passwordChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func randomPassword(n int) string {
	b := make([]byte, n)
	for i := range b {
		idx, _ := rand.Int(rand.Reader, big.NewInt(int64(len(passwordChars))))
		b[i] = passwordChars[idx.Int64()]
	}
	return string(b)
}

func randomHex(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func slug(s string) string {
	s = strings.ToLower(s)
	re := regexp.MustCompile(`[^a-z0-9-]+`)
	s = re.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	return s
}

func init() {
	initCmd.Flags().StringVar(&initTemplateName, "template", "",
		"template to use (skips interactive template selection)")
	initCmd.Flags().StringVar(&initDir, "dir", "",
		"output directory (default: <stacks-root>/<name>)")
	initCmd.Flags().BoolVar(&initListTemplates, "list-templates", false,
		"list available templates and exit")
	initCmd.Flags().BoolVar(&initRefreshTemplates, "refresh-templates", false,
		"re-download all templates then exit")
	rootCmd.AddCommand(initCmd)
}
