package selfupdate

import "testing"

func TestSameVersion(t *testing.T) {
	tests := []struct {
		a, b string
		want bool
	}{
		{"0.3.0", "v0.3.0", true},
		{"v0.3.0", "0.3.0", true},
		{"v0.3.0", "v0.3.0", true},
		{"0.2.2", "v0.3.0", false},
		{"dev", "v0.3.0", false},
	}
	for _, tt := range tests {
		if got := SameVersion(tt.a, tt.b); got != tt.want {
			t.Errorf("SameVersion(%q, %q) = %v, want %v", tt.a, tt.b, got, tt.want)
		}
	}
}
