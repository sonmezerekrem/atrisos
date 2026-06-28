package scheduler

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"text/template"

	"github.com/sonmezerekrem/atrisos/internal/stack"
)

// InstallAutoStart creates and enables the OS scheduler unit for a stack's
// auto-start behavior. No-op if auto_start is false.
func InstallAutoStart(s *stack.Stack, atrisosPath string) error {
	if !s.Config.AutoStart {
		return nil
	}
	switch runtime.GOOS {
	case "linux":
		return linuxInstallAutoStart(s, atrisosPath)
	case "darwin":
		return darwinInstallAutoStart(s, atrisosPath)
	}
	return nil
}

// RemoveAutoStart removes the auto-start unit for a stack.
func RemoveAutoStart(s *stack.Stack) error {
	switch runtime.GOOS {
	case "linux":
		return linuxRemoveUnit(linuxAutoStartUnit(s.Name))
	case "darwin":
		return darwinRemovePlist(darwinAutoStartPlist(s.Name))
	}
	return nil
}

// InstallBackupTimer creates and enables the OS scheduler unit for a stack's
// backup schedule. No-op if backup is not enabled or schedule is empty.
func InstallBackupTimer(s *stack.Stack, atrisosPath string) error {
	if !s.Config.Backup.Enabled || s.Config.Backup.Schedule == "" {
		return nil
	}
	switch runtime.GOOS {
	case "linux":
		return linuxInstallBackupTimer(s, atrisosPath)
	case "darwin":
		return darwinInstallBackupTimer(s, atrisosPath)
	}
	return nil
}

// RemoveBackupTimer removes the backup timer unit for a stack.
// No-op if backup was never enabled.
func RemoveBackupTimer(s *stack.Stack) error {
	if !s.Config.Backup.Enabled {
		return nil
	}
	switch runtime.GOOS {
	case "linux":
		svcUnit := linuxBackupServiceUnit(s.Name)
		timerUnit := linuxBackupTimerUnit(s.Name)
		if err := linuxRemoveUnit(timerUnit); err != nil {
			return err
		}
		return linuxRemoveUnit(svcUnit)
	case "darwin":
		return darwinRemovePlist(darwinBackupPlist(s.Name))
	}
	return nil
}

// ── Linux ─────────────────────────────────────────────────────────────────────

func systemdDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "systemd", "user")
}

func linuxAutoStartUnit(name string) string {
	return "atrisos-" + name + ".service"
}

func linuxBackupServiceUnit(name string) string {
	return "atrisos-backup-" + name + ".service"
}

func linuxBackupTimerUnit(name string) string {
	return "atrisos-backup-" + name + ".timer"
}

const linuxAutoStartServiceTmpl = `[Unit]
Description=atrisos stack: {{.Name}}
After=default.target

[Service]
Type=oneshot
ExecStart={{.AtrisosPath}} up {{.Name}}
RemainAfterExit=yes

[Install]
WantedBy=default.target
`

const linuxBackupServiceTmpl = `[Unit]
Description=atrisos backup: {{.Name}}

[Service]
Type=oneshot
ExecStart={{.AtrisosPath}} backup {{.Name}}
`

const linuxBackupTimerTmpl = `[Unit]
Description=atrisos backup timer: {{.Name}}

[Timer]
OnCalendar={{.OnCalendar}}
Persistent=true

[Install]
WantedBy=timers.target
`

func linuxInstallAutoStart(s *stack.Stack, atrisosPath string) error {
	dir := systemdDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating systemd user dir: %w", err)
	}

	unitPath := filepath.Join(dir, linuxAutoStartUnit(s.Name))
	data := struct {
		Name        string
		AtrisosPath string
	}{s.Name, atrisosPath}

	if err := writeTemplate(unitPath, linuxAutoStartServiceTmpl, data); err != nil {
		return err
	}

	return systemctlUser("enable", "--now", linuxAutoStartUnit(s.Name))
}

func linuxInstallBackupTimer(s *stack.Stack, atrisosPath string) error {
	dir := systemdDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating systemd user dir: %w", err)
	}

	data := struct {
		Name        string
		AtrisosPath string
	}{s.Name, atrisosPath}

	svcPath := filepath.Join(dir, linuxBackupServiceUnit(s.Name))
	if err := writeTemplate(svcPath, linuxBackupServiceTmpl, data); err != nil {
		return err
	}

	onCal, err := cronToOnCalendar(s.Config.Backup.Schedule)
	if err != nil {
		fmt.Fprintf(os.Stderr, "⚠ %v — using OnCalendar=daily\n", err)
		onCal = "daily"
	}

	timerData := struct {
		Name       string
		OnCalendar string
	}{s.Name, onCal}

	timerPath := filepath.Join(dir, linuxBackupTimerUnit(s.Name))
	if err := writeTemplate(timerPath, linuxBackupTimerTmpl, timerData); err != nil {
		return err
	}

	return systemctlUser("enable", "--now", linuxBackupTimerUnit(s.Name))
}

func linuxRemoveUnit(unit string) error {
	unitPath := filepath.Join(systemdDir(), unit)
	if _, err := os.Stat(unitPath); os.IsNotExist(err) {
		return nil
	}
	_ = systemctlUser("disable", "--now", unit)
	if err := os.Remove(unitPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("removing %s: %w", unitPath, err)
	}
	return nil
}

func systemctlUser(args ...string) error {
	cmd := exec.Command("systemctl", append([]string{"--user"}, args...)...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// cronToOnCalendar converts a 5-field cron expression to a systemd OnCalendar
// value. For v1 only simple minute+hour daily patterns are supported.
func cronToOnCalendar(cron string) (string, error) {
	fields := strings.Fields(cron)
	if len(fields) != 5 {
		return "", fmt.Errorf("unsupported cron expression %q (expected 5 fields)", cron)
	}
	minute, hour := fields[0], fields[1]
	dom, month, dow := fields[2], fields[3], fields[4]

	// Only support simple daily cron: dom/month/dow all wildcard.
	if dom != "*" || month != "*" || dow != "*" {
		return "", fmt.Errorf("complex cron %q not supported", cron)
	}

	// Validate hour and minute are plain integers or *.
	hourStr := "*"
	if hour != "*" {
		h, err := strconv.Atoi(hour)
		if err != nil || h < 0 || h > 23 {
			return "", fmt.Errorf("invalid hour in cron %q", cron)
		}
		hourStr = fmt.Sprintf("%02d", h)
	}
	minStr := "00"
	if minute != "*" {
		m, err := strconv.Atoi(minute)
		if err != nil || m < 0 || m > 59 {
			return "", fmt.Errorf("invalid minute in cron %q", cron)
		}
		minStr = fmt.Sprintf("%02d", m)
	}

	return fmt.Sprintf("*-*-* %s:%s:00", hourStr, minStr), nil
}

// ── macOS ─────────────────────────────────────────────────────────────────────

func launchAgentsDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "Library", "LaunchAgents")
}

func darwinAutoStartPlist(name string) string {
	return filepath.Join(launchAgentsDir(), "io.atrisos."+name+".plist")
}

func darwinBackupPlist(name string) string {
	return filepath.Join(launchAgentsDir(), "io.atrisos.backup."+name+".plist")
}

const darwinAutoStartPlistTmpl = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>io.atrisos.{{.Name}}</string>
    <key>ProgramArguments</key>
    <array>
        <string>{{.AtrisosPath}}</string>
        <string>up</string>
        <string>{{.Name}}</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
</dict>
</plist>
`

const darwinBackupPlistTmpl = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>io.atrisos.backup.{{.Name}}</string>
    <key>ProgramArguments</key>
    <array>
        <string>{{.AtrisosPath}}</string>
        <string>backup</string>
        <string>{{.Name}}</string>
    </array>
    <key>StartCalendarInterval</key>
    <dict>
        <key>Hour</key>
        <integer>{{.Hour}}</integer>
        <key>Minute</key>
        <integer>{{.Minute}}</integer>
    </dict>
</dict>
</plist>
`

func darwinInstallAutoStart(s *stack.Stack, atrisosPath string) error {
	if err := os.MkdirAll(launchAgentsDir(), 0755); err != nil {
		return fmt.Errorf("creating LaunchAgents dir: %w", err)
	}

	plistPath := darwinAutoStartPlist(s.Name)
	data := struct {
		Name        string
		AtrisosPath string
	}{s.Name, atrisosPath}

	if err := writeTemplate(plistPath, darwinAutoStartPlistTmpl, data); err != nil {
		return err
	}

	return launchctlLoad(plistPath)
}

func darwinInstallBackupTimer(s *stack.Stack, atrisosPath string) error {
	if err := os.MkdirAll(launchAgentsDir(), 0755); err != nil {
		return fmt.Errorf("creating LaunchAgents dir: %w", err)
	}

	fields := strings.Fields(s.Config.Backup.Schedule)
	hour, minute := 2, 0
	if len(fields) >= 2 {
		if h, err := strconv.Atoi(fields[1]); err == nil {
			hour = h
		}
		if m, err := strconv.Atoi(fields[0]); err == nil {
			minute = m
		}
	}

	plistPath := darwinBackupPlist(s.Name)
	data := struct {
		Name        string
		AtrisosPath string
		Hour        int
		Minute      int
	}{s.Name, atrisosPath, hour, minute}

	if err := writeTemplate(plistPath, darwinBackupPlistTmpl, data); err != nil {
		return err
	}

	return launchctlLoad(plistPath)
}

func darwinRemovePlist(plistPath string) error {
	// Best-effort unload.
	_ = launchctlUnload(plistPath)

	if err := os.Remove(plistPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("removing %s: %w", plistPath, err)
	}
	return nil
}

func launchctlLoad(plistPath string) error {
	cmd := exec.Command("launchctl", "load", plistPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func launchctlUnload(plistPath string) error {
	cmd := exec.Command("launchctl", "unload", plistPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// ── helpers ───────────────────────────────────────────────────────────────────

func writeTemplate(path, tmplStr string, data interface{}) error {
	tmpl, err := template.New("").Parse(tmplStr)
	if err != nil {
		return fmt.Errorf("parsing template: %w", err)
	}

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("creating %s: %w", path, err)
	}
	defer f.Close()

	return tmpl.Execute(f, data)
}
