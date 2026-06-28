package tui

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/sonmezerekrem/atrisos/internal/config"
	"github.com/sonmezerekrem/atrisos/internal/notify"
	"github.com/sonmezerekrem/atrisos/internal/outdated"
	"github.com/sonmezerekrem/atrisos/internal/podman"
	"github.com/sonmezerekrem/atrisos/internal/registry"
	"github.com/sonmezerekrem/atrisos/internal/stack"
)

// panelMode distinguishes which panel is currently displayed.
type panelMode int

const (
	panelList panelMode = iota
	panelLogs
)

// ---- Message types ----

type stacksLoadedMsg struct{ statuses []*podman.StackStatus }
type traefikStatusMsg struct{ status string }
type refreshTickMsg   struct{}
type logLineMsg       struct{ line string }
type logDoneMsg       struct{}
type errMsg           struct{ err error }
type outdatedResultMsg struct {
	updates map[string]bool // stack name → has updates
}

// ---- AppModel ----

// AppModel is the root bubbletea model for atrisos.
type AppModel struct {
	cfg            *config.Config
	reg            *registry.Registry
	stacks         []*podman.StackStatus
	filtered       []*podman.StackStatus
	cursor         int
	panel          panelMode
	filterText     string
	filtering      bool
	logModel       logModel
	width          int
	height         int
	traefik        string
	lastRefresh    time.Time
	err            error
	outdatedUpdates map[string]bool // stack name → has image updates
	prevStates      map[string]string // "<stackName>/<service>" → previous container status
}

func newAppModel(cfg *config.Config, reg *registry.Registry) AppModel {
	return AppModel{
		cfg:     cfg,
		reg:     reg,
		traefik: "unknown",
		panel:   panelList,
	}
}

// Init returns the initial batch of startup commands.
func (m AppModel) Init() tea.Cmd {
	return tea.Batch(
		loadStacksCmd(m.cfg, m.reg),
		loadTraefikStatusCmd(),
		tea.Tick(10*time.Second, func(t time.Time) tea.Msg {
			return refreshTickMsg{}
		}),
	)
}

// loadStacksCmd discovers stacks and queries Podman for live status.
func loadStacksCmd(cfg *config.Config, reg *registry.Registry) tea.Cmd {
	return func() tea.Msg {
		stacks, err := stack.Discover(cfg, reg)
		if err != nil {
			return errMsg{err: err}
		}
		statuses, err := podman.GetAllStatus(stacks)
		if err != nil {
			return errMsg{err: err}
		}
		return stacksLoadedMsg{statuses: statuses}
	}
}

// loadTraefikStatusCmd queries for the Traefik container state.
func loadTraefikStatusCmd() tea.Cmd {
	return func() tea.Msg {
		return traefikStatusMsg{status: checkTraefikStatus()}
	}
}

// checkOutdatedCmd runs an async check for image updates across all stacks.
func checkOutdatedCmd(stacks []*podman.StackStatus) tea.Cmd {
	return func() tea.Msg {
		stackList := make([]*stack.Stack, 0, len(stacks))
		for _, s := range stacks {
			stackList = append(stackList, s.Stack)
		}
		updates, _ := outdated.CheckAll(stackList)
		result := make(map[string]bool, len(updates))
		for name := range updates {
			result[name] = true
		}
		return outdatedResultMsg{updates: result}
	}
}

// applyFilter rebuilds m.filtered from m.stacks using m.filterText.
func (m *AppModel) applyFilter() {
	if m.filterText == "" {
		m.filtered = m.stacks
		return
	}
	lower := strings.ToLower(m.filterText)
	result := make([]*podman.StackStatus, 0)
	for _, s := range m.stacks {
		if strings.Contains(strings.ToLower(s.Stack.Name), lower) {
			result = append(result, s)
		}
	}
	m.filtered = result
}

// selectedStack returns the currently highlighted StackStatus, or nil.
func (m AppModel) selectedStack() *podman.StackStatus {
	if len(m.filtered) == 0 || m.cursor >= len(m.filtered) {
		return nil
	}
	return m.filtered[m.cursor]
}

// panelWidths returns (leftWidth, rightWidth) for the two-column layout.
func panelWidths(totalWidth int) (int, int) {
	leftW := totalWidth * 3 / 10
	if leftW < 20 {
		leftW = 20
	}
	if leftW > totalWidth-10 {
		leftW = totalWidth - 10
	}
	return leftW, totalWidth - leftW
}

// execAtrisos builds an *exec.Cmd to call the atrisos binary with the given args.
func execAtrisos(args ...string) *exec.Cmd {
	return exec.Command("atrisos", args...)
}

// Update handles all bubbletea messages.
func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.logModel.resize(m.width, m.height)
		return m, nil

	case stacksLoadedMsg:
		// Detect unexpected container exits and fire webhooks.
		for _, ss := range msg.statuses {
			if ss.Stack.Config.Notify.Webhook == "" {
				continue
			}
			for _, c := range ss.Containers {
				key := ss.Stack.Name + "/" + c.Service
				prevStatus, hasPrev := m.prevStates[key]
				if hasPrev && prevStatus == "running" && (c.Status == "exited" || c.Status == "dead") {
					webhook := ss.Stack.Config.Notify.Webhook
					stackName := ss.Stack.Name
					service := c.Service
					go notify.Send(webhook, notify.Payload{ //nolint:errcheck
						Event:     notify.EventContainerExit,
						Stack:     stackName,
						Service:   service,
						Timestamp: time.Now(),
						Message:   fmt.Sprintf("container %s/%s exited unexpectedly", stackName, service),
					})
				}
			}
		}
		// Update previous container states.
		if m.prevStates == nil {
			m.prevStates = make(map[string]string)
		}
		for _, ss := range msg.statuses {
			for _, c := range ss.Containers {
				m.prevStates[ss.Stack.Name+"/"+c.Service] = c.Status
			}
		}

		m.stacks = msg.statuses
		m.applyFilter()
		if m.cursor >= len(m.filtered) && len(m.filtered) > 0 {
			m.cursor = len(m.filtered) - 1
		}
		if len(m.filtered) == 0 {
			m.cursor = 0
		}
		m.lastRefresh = time.Now()
		m.err = nil

		var cmds []tea.Cmd
		cmds = append(cmds, tea.Tick(10*time.Second, func(t time.Time) tea.Msg {
			return refreshTickMsg{}
		}))
		if m.outdatedUpdates == nil {
			cmds = append(cmds, checkOutdatedCmd(m.stacks))
		}
		return m, tea.Batch(cmds...)

	case traefikStatusMsg:
		m.traefik = msg.status
		return m, nil

	case refreshTickMsg:
		return m, tea.Batch(
			loadStacksCmd(m.cfg, m.reg),
			loadTraefikStatusCmd(),
		)

	case logsReadyMsg:
		m.logModel.ch = msg.ch
		m.logModel.cmd = msg.cmd
		m.logModel.ready = true
		return m, waitLogLine(msg.ch)

	case logLineMsg:
		m.logModel.addLine(msg.line)
		return m, waitLogLine(m.logModel.ch)

	case logDoneMsg:
		return m, nil

	case outdatedResultMsg:
		m.outdatedUpdates = msg.updates
		return m, nil

	case errMsg:
		m.err = msg.err
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

// handleKey dispatches keyboard events to the appropriate panel handler.
func (m AppModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.panel == panelLogs {
		return m.handleLogKey(msg)
	}
	if m.filtering {
		return m.handleFilterKey(msg)
	}
	return m.handleListKey(msg)
}

func (m AppModel) handleLogKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc":
		m.logModel.kill()
		m.panel = panelList
	case "ctrl+c":
		m.logModel.kill()
		return m, tea.Quit
	case "up", "k":
		m.logModel.viewport.LineUp(1) //nolint:errcheck
	case "down", "j":
		m.logModel.viewport.LineDown(1) //nolint:errcheck
	case "pgup":
		m.logModel.viewport.HalfViewUp() //nolint:errcheck
	case "pgdown":
		m.logModel.viewport.HalfViewDown() //nolint:errcheck
	}
	return m, nil
}

func (m AppModel) handleFilterKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		m.filtering = false
	case "esc":
		m.filtering = false
		m.filterText = ""
		m.applyFilter()
		m.cursor = 0
	case "backspace":
		if len(m.filterText) > 0 {
			m.filterText = m.filterText[:len(m.filterText)-1]
			m.applyFilter()
			m.cursor = 0
		}
	default:
		if len(msg.Runes) > 0 {
			m.filterText += string(msg.Runes)
			m.applyFilter()
			m.cursor = 0
		}
	}
	return m, nil
}

func (m AppModel) handleListKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit

	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}

	case "down", "j":
		if m.cursor < len(m.filtered)-1 {
			m.cursor++
		}

	case "/":
		m.filtering = true

	case "l":
		if s := m.selectedStack(); s != nil {
			composeFile := stack.ComposeFile(s.Stack.Dir)
			m.logModel = newLogModel(m.width, m.height, s.Stack.Name)
			m.panel = panelLogs
			return m, startLogsCmd(s.Stack.Dir, s.Stack.Name, composeFile)
		}

	case "u", "s":
		if s := m.selectedStack(); s != nil {
			cmd := execAtrisos("up", s.Stack.Name)
			return m, tea.ExecProcess(cmd, func(err error) tea.Msg {
				return refreshTickMsg{}
			})
		}

	case "r":
		if s := m.selectedStack(); s != nil {
			cmd := execAtrisos("restart", s.Stack.Name)
			return m, tea.ExecProcess(cmd, func(err error) tea.Msg {
				return refreshTickMsg{}
			})
		}

	case "x":
		if s := m.selectedStack(); s != nil {
			cmd := execAtrisos("down", s.Stack.Name)
			return m, tea.ExecProcess(cmd, func(err error) tea.Msg {
				return refreshTickMsg{}
			})
		}

	case "e":
		if s := m.selectedStack(); s != nil && len(s.Containers) > 0 {
			firstService := s.Containers[0].Service
			cmd := execAtrisos("shell", s.Stack.Name, firstService)
			return m, tea.ExecProcess(cmd, func(err error) tea.Msg {
				return nil
			})
		}
	}
	return m, nil
}

// View renders the complete TUI.
func (m AppModel) View() string {
	if m.width == 0 {
		return "Loading…"
	}

	header := m.renderHeader()

	if m.panel == panelLogs {
		return lipgloss.JoinVertical(lipgloss.Left,
			header,
			m.logModel.view(),
		)
	}

	leftW, rightW := panelWidths(m.width)
	panelH := m.height - 2 // header + status bar
	if panelH < 3 {
		panelH = 3
	}
	innerH := panelH - 2 // subtract border top+bottom
	if innerH < 1 {
		innerH = 1
	}

	leftContent := RenderList(m.filtered, m.cursor, m.filterText, leftW-2, innerH, m.outdatedUpdates)
	leftPanel := panelBorderStyle.
		Width(leftW - 2).
		Height(innerH).
		Render(leftContent)

	rightContent := RenderDetail(m.selectedStack(), rightW-2, innerH)
	rightPanel := panelBorderStyle.
		Width(rightW - 2).
		Height(innerH).
		Render(rightContent)

	body := lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, rightPanel)
	statusBar := m.renderStatusBar()

	return lipgloss.JoinVertical(lipgloss.Left, header, body, statusBar)
}

// renderHeader builds the one-line header.
func (m AppModel) renderHeader() string {
	traefikStr := m.traefik
	var traefikStyled string
	switch traefikStr {
	case "running":
		traefikStyled = colorGreen.Render("traefik: " + traefikStr)
	case "stopped":
		traefikStyled = colorRed.Render("traefik: " + traefikStr)
	default:
		traefikStyled = colorDim.Render("traefik: " + traefikStr)
	}

	title := headerStyle.Render("atrisos")
	pad := m.width - lipgloss.Width(title) - lipgloss.Width(traefikStyled) - 1
	if pad < 1 {
		pad = 1
	}
	return title + strings.Repeat(" ", pad) + traefikStyled
}

// renderStatusBar builds the bottom status bar.
func (m AppModel) renderStatusBar() string {
	if m.filtering {
		return statusBarStyle.Render("filter: " + m.filterText + "█  enter accept  esc clear")
	}
	if m.err != nil {
		return colorRed.Padding(0, 1).Render(fmt.Sprintf("error: %v", m.err))
	}
	age := ""
	if !m.lastRefresh.IsZero() {
		age = fmt.Sprintf("  (updated %s ago)", time.Since(m.lastRefresh).Truncate(time.Second))
	}
	return statusBarStyle.Render(
		"↑↓/jk navigate  / filter  l logs  u up  r restart  x down  e shell  q quit" + age,
	)
}
