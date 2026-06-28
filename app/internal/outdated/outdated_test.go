package outdated

import "testing"

func TestShortDigest(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"docker.io/library/nginx@sha256:abcdef0123456789abcdef0123456789", "abcdef012345"},
		{"sha256:deadbeef0000", "deadbeef0000"},
		{"short", "short"},
		{"verylongstring", "verylongstri"},
	}
	for _, tt := range tests {
		if got := shortDigest(tt.in); got != tt.want {
			t.Errorf("shortDigest(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestRemoteDigestParse(t *testing.T) {
	manifest := `{"manifests":[{"digest":"sha256:abc123def4567890123456789012345678901234567890123456789012345678"}]}`
	got, err := parseManifestDigest(manifest)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "abc123def456" {
		t.Errorf("got %q, want abc123def456", got)
	}
}

func TestRemoteDigestParseErrors(t *testing.T) {
	if _, err := parseManifestDigest(`{"layers":[]}`); err == nil {
		t.Error("expected error for manifest without digest")
	}
}
