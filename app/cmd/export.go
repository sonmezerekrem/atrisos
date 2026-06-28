package cmd

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/sonmezerekrem/atrisos/internal/stack"
	"github.com/spf13/cobra"
)

var exportOutput string

var exportCmd = &cobra.Command{
	Use:   "export <stack>",
	Short: "Package a stack into a portable .tar.gz",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		stacks, err := resolveStacks(args, false, "")
		if err != nil {
			return err
		}
		s := stacks[0]
		return exportStack(s)
	},
}

var importDir string

var importCmd = &cobra.Command{
	Use:   "import <file>",
	Short: "Extract a stack archive into the stacks root directory",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return importStack(args[0])
	},
}

// exportStack creates a tar.gz of the stack's configuration files.
// .env is intentionally excluded.
func exportStack(s *stack.Stack) error {
	outPath := exportOutput
	if outPath == "" {
		outPath = fmt.Sprintf("./%s.tar.gz", s.Name)
	}

	// Files to include (relative to stack dir).
	candidates := []string{
		"compose.yml",
		"docker-compose.yml",
		"compose.override.yml",
		"config.yml",
		".env.example",
	}

	f, err := os.Create(outPath)
	if err != nil {
		return fmt.Errorf("creating archive: %w", err)
	}
	defer f.Close()

	gw := gzip.NewWriter(f)
	defer gw.Close()

	tw := tar.NewWriter(gw)
	defer tw.Close()

	included := 0
	for _, name := range candidates {
		srcPath := filepath.Join(s.Dir, name)
		if _, err := os.Stat(srcPath); os.IsNotExist(err) {
			continue
		}

		data, err := os.ReadFile(srcPath)
		if err != nil {
			return fmt.Errorf("reading %s: %w", name, err)
		}

		hdr := &tar.Header{
			Name: filepath.Join(s.Name, name),
			Mode: 0o644,
			Size: int64(len(data)),
		}
		if err := tw.WriteHeader(hdr); err != nil {
			return fmt.Errorf("writing tar header for %s: %w", name, err)
		}
		if _, err := tw.Write(data); err != nil {
			return fmt.Errorf("writing %s to archive: %w", name, err)
		}
		included++
	}

	if included == 0 {
		os.Remove(outPath)
		return fmt.Errorf("no files to export from %s", s.Dir)
	}

	printSuccess(fmt.Sprintf("exported %s to %s (%d files)", s.Name, outPath, included))
	return nil
}

// importStack extracts a tar.gz into the appropriate directory.
func importStack(archivePath string) error {
	destDir := importDir
	if destDir == "" {
		destDir = rootCfg.StacksRoot
	}

	f, err := os.Open(archivePath)
	if err != nil {
		return fmt.Errorf("opening archive: %w", err)
	}
	defer f.Close()

	gr, err := gzip.NewReader(f)
	if err != nil {
		return fmt.Errorf("reading gzip: %w", err)
	}
	defer gr.Close()

	tr := tar.NewReader(gr)

	var stackName string
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("reading archive: %w", err)
		}

		// Determine stack name from first path component.
		parts := strings.SplitN(hdr.Name, "/", 2)
		if len(parts) < 2 {
			continue
		}
		if stackName == "" {
			stackName = parts[0]
		}

		outPath := filepath.Join(destDir, hdr.Name)

		// Safety: prevent path traversal.
		if !strings.HasPrefix(filepath.Clean(outPath), filepath.Clean(destDir)) {
			return fmt.Errorf("unsafe path in archive: %s", hdr.Name)
		}

		if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
			return fmt.Errorf("creating directory: %w", err)
		}

		out, err := os.Create(outPath)
		if err != nil {
			return fmt.Errorf("creating file %s: %w", outPath, err)
		}

		if _, err := io.Copy(out, tr); err != nil {
			out.Close()
			return fmt.Errorf("extracting %s: %w", hdr.Name, err)
		}
		out.Close()
	}

	extractedDir := filepath.Join(destDir, stackName)
	printSuccess(fmt.Sprintf("imported to %s", extractedDir))

	// Check if .env.example exists but .env doesn't.
	envExamplePath := filepath.Join(extractedDir, ".env.example")
	envPath := filepath.Join(extractedDir, ".env")
	if _, err := os.Stat(envExamplePath); err == nil {
		if _, err := os.Stat(envPath); os.IsNotExist(err) {
			fmt.Printf("\n")
			printWarn("no .env file found. Create one from the example:")
			fmt.Printf("  cp %s %s\n", envExamplePath, envPath)
			fmt.Printf("  # then edit %s with your actual values\n", envPath)
		}
	}

	return nil
}

func init() {
	exportCmd.Flags().StringVar(&exportOutput, "output", "",
		"output path for the archive (default: ./<stack>.tar.gz)")
	importCmd.Flags().StringVar(&importDir, "dir", "",
		"extract to a specific location (default: stacks root)")
	rootCmd.AddCommand(exportCmd)
	rootCmd.AddCommand(importCmd)
}
