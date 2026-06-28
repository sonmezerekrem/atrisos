package scheduler

import "testing"

func TestCronToOnCalendar(t *testing.T) {
	tests := []struct {
		cron    string
		want    string
		wantErr bool
	}{
		{"0 3 * * *", "*-*-* 03:00:00", false},
		{"30 14 * * *", "*-*-* 14:30:00", false},
		{"* * * * *", "*-*-* *:00:00", false},
		{"0 3 1 * *", "", true},
		{"bad cron", "", true},
		{"60 3 * * *", "", true},
	}
	for _, tt := range tests {
		got, err := cronToOnCalendar(tt.cron)
		if tt.wantErr {
			if err == nil {
				t.Errorf("cronToOnCalendar(%q) expected error", tt.cron)
			}
			continue
		}
		if err != nil {
			t.Errorf("cronToOnCalendar(%q): %v", tt.cron, err)
			continue
		}
		if got != tt.want {
			t.Errorf("cronToOnCalendar(%q) = %q, want %q", tt.cron, got, tt.want)
		}
	}
}

func TestLinuxUnitNames(t *testing.T) {
	if got := linuxAutoStartUnit("myapp"); got != "atrisos-myapp.service" {
		t.Errorf("linuxAutoStartUnit = %q", got)
	}
	if got := linuxBackupTimerUnit("myapp"); got != "atrisos-backup-myapp.timer" {
		t.Errorf("linuxBackupTimerUnit = %q", got)
	}
}
