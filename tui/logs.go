package tui

import (
	"bufio"
	"io"
	"os/exec"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const maxLogLines = 2000

// logsReadyMsg is sent once the log subprocess has started and a line channel
// is available for consumption.
type logsReadyMsg struct {
	ch  chan string
	cmd *exec.Cmd
}

// logModel holds state for the full-screen log streaming panel.
type logModel struct {
	viewport viewport.Model
	lines    []string
	stack    string
	ready    bool
	ch       chan string
	cmd      *exec.Cmd
}

func newLogModel(width, height int, stackName string) logModel {
	vp := viewport.New(width, logViewportHeight(height))
	vp.SetContent("")
	return logModel{
		viewport: vp,
		stack:    stackName,
	}
}

// logViewportHeight returns the viewport height given the terminal height.
func logViewportHeight(h int) int {
	vh := h - 3 // header + status bar + small margin
	if vh < 1 {
		return 1
	}
	return vh
}

// resize updates the viewport dimensions.
func (lm *logModel) resize(width, height int) {
	lm.viewport.Width = width
	lm.viewport.Height = logViewportHeight(height)
}

// addLine appends a log line (capping at maxLogLines) and auto-scrolls.
func (lm *logModel) addLine(line string) {
	lm.lines = append(lm.lines, line)
	if len(lm.lines) > maxLogLines {
		lm.lines = lm.lines[len(lm.lines)-maxLogLines:]
	}
	lm.viewport.SetContent(strings.Join(lm.lines, "\n"))
	lm.viewport.GotoBottom() //nolint:errcheck
}

// view renders the full-screen log panel.
func (lm logModel) view() string {
	title := sectionHeadingStyle.Render("Logs: " + lm.stack)
	help := dimStyle.Render("q/esc close  ↑↓ scroll  pgup/pgdn page")
	return lipgloss.JoinVertical(lipgloss.Left,
		title,
		lm.viewport.View(),
		help,
	)
}

// kill terminates the log subprocess if one is running.
func (lm *logModel) kill() {
	if lm.cmd != nil && lm.cmd.Process != nil {
		_ = lm.cmd.Process.Kill()
		lm.cmd = nil
	}
}

// startLogsCmd starts `podman compose logs -f` for the given stack and
// immediately returns a logsReadyMsg containing the line channel and cmd.
func startLogsCmd(stackDir, stackName, composeFile string) tea.Cmd {
	return func() tea.Msg {
		args := []string{"compose", "--project-name", stackName}
		if composeFile != "" {
			args = append(args, "-f", composeFile)
		}
		args = append(args, "logs", "-f", "--timestamps", "--no-color")

		cmd := exec.Command("podman", args...)
		if stackDir != "" {
			cmd.Dir = stackDir
		}

		stdout, err := cmd.StdoutPipe()
		if err != nil {
			return errMsg{err: err}
		}
		stderr, err := cmd.StderrPipe()
		if err != nil {
			return errMsg{err: err}
		}
		if err := cmd.Start(); err != nil {
			return errMsg{err: err}
		}

		// Discard stderr to prevent blocking.
		go func() { _, _ = io.Copy(io.Discard, stderr) }()

		ch := make(chan string, 100)
		go func() {
			defer close(ch)
			scanner := bufio.NewScanner(stdout)
			for scanner.Scan() {
				ch <- scanner.Text()
			}
			_ = cmd.Wait()
		}()

		return logsReadyMsg{ch: ch, cmd: cmd}
	}
}

// waitLogLine blocks until the next log line is available on ch, then returns
// a logLineMsg (or logDoneMsg when the channel is closed).
func waitLogLine(ch chan string) tea.Cmd {
	return func() tea.Msg {
		line, ok := <-ch
		if !ok {
			return logDoneMsg{}
		}
		return logLineMsg{line: line}
	}
}
