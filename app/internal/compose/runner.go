package compose

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Runner wraps podman compose invocations.
type Runner struct {
	ComposeCmd  string // "podman compose" or "podman-compose"
	ProjectDir  string
	EnvFile     string
	ProjectName string
}

// NewRunner creates a Runner with auto-detected compose command.
func NewRunner(projectDir, projectName string) *Runner {
	envFile := ""
	envPath := projectDir + "/.env"
	if _, err := os.Stat(envPath); err == nil {
		envFile = envPath
	}
	return &Runner{
		ComposeCmd:  DetectComposeCmd(),
		ProjectDir:  projectDir,
		EnvFile:     envFile,
		ProjectName: projectName,
	}
}

// DetectComposeCmd tries "podman compose" first, then "podman-compose".
// Panics with a clear message if neither is found.
func DetectComposeCmd() string {
	cmd := exec.Command("podman", "compose", "version")
	cmd.Stdout = nil
	cmd.Stderr = nil
	if err := cmd.Run(); err == nil {
		return "podman compose"
	}

	cmd2 := exec.Command("podman-compose", "--version")
	cmd2.Stdout = nil
	cmd2.Stderr = nil
	if err := cmd2.Run(); err == nil {
		return "podman-compose"
	}

	fmt.Fprintln(os.Stderr, "error: podman compose not found.")
	fmt.Fprintln(os.Stderr, "Install Podman v4.7+ with built-in compose support, or install podman-compose.")
	os.Exit(4)
	return "" // unreachable
}

// buildArgs constructs the argument list for a compose command.
func (r *Runner) buildArgs(composeFile string, args []string) (string, []string) {
	parts := strings.Fields(r.ComposeCmd)
	binary := parts[0]
	cmdArgs := make([]string, 0, len(parts)-1+10+len(args))
	cmdArgs = append(cmdArgs, parts[1:]...)

	if composeFile != "" {
		cmdArgs = append(cmdArgs, "-f", composeFile)
	}
	if r.ProjectName != "" {
		cmdArgs = append(cmdArgs, "--project-name", r.ProjectName)
	}
	if r.EnvFile != "" {
		cmdArgs = append(cmdArgs, "--env-file", r.EnvFile)
	}
	cmdArgs = append(cmdArgs, args...)

	return binary, cmdArgs
}

// Run runs compose with the given compose file and args, streaming output
// to stdout/stderr.
func (r *Runner) Run(composeFile string, args ...string) error {
	binary, cmdArgs := r.buildArgs(composeFile, args)
	cmd := exec.Command(binary, cmdArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if r.ProjectDir != "" {
		cmd.Dir = r.ProjectDir
	}
	return cmd.Run()
}

// RunMerged writes the ComposeDoc to a temp file, runs compose against it,
// then deletes the temp file.
func (r *Runner) RunMerged(doc ComposeDoc, args ...string) error {
	tmpFile, err := WriteToTemp(doc)
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile)
	return r.Run(tmpFile, args...)
}

// Output runs compose and returns the captured stdout.
func (r *Runner) Output(composeFile string, args ...string) ([]byte, error) {
	binary, cmdArgs := r.buildArgs(composeFile, args)
	cmd := exec.Command(binary, cmdArgs...)
	cmd.Stderr = os.Stderr
	if r.ProjectDir != "" {
		cmd.Dir = r.ProjectDir
	}
	return cmd.Output()
}

// OutputMerged writes the ComposeDoc to a temp file, runs compose and
// captures output, then deletes the temp file.
func (r *Runner) OutputMerged(doc ComposeDoc, args ...string) ([]byte, error) {
	tmpFile, err := WriteToTemp(doc)
	if err != nil {
		return nil, err
	}
	defer os.Remove(tmpFile)
	return r.Output(tmpFile, args...)
}
