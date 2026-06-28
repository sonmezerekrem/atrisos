package notify

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestSendEmptyWebhook(t *testing.T) {
	if err := Send("", Payload{Event: EventBackupFailure, Stack: "myapp"}); err != nil {
		t.Errorf("empty webhook should return nil, got %v", err)
	}
}

func TestSendPostsJSON(t *testing.T) {
	var got Payload
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("Content-Type = %q", ct)
		}
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Fatal(err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	ts := time.Date(2026, 6, 28, 12, 0, 0, 0, time.UTC)
	if err := Send(srv.URL, Payload{
		Event:     EventContainerExit,
		Stack:     "web",
		Service:   "api",
		Timestamp: ts,
		Message:   "container stopped",
	}); err != nil {
		t.Fatal(err)
	}
	if got.Event != EventContainerExit || got.Stack != "web" || got.Service != "api" {
		t.Errorf("payload = %+v", got)
	}
	if !got.Timestamp.Equal(ts) {
		t.Errorf("timestamp = %v, want %v", got.Timestamp, ts)
	}
}

func TestSendSetsTimestampWhenZero(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var p Payload
		_ = json.NewDecoder(r.Body).Decode(&p)
		if p.Timestamp.IsZero() {
			t.Error("expected timestamp to be set")
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	if err := Send(srv.URL, Payload{Event: EventCertExpiry, Stack: "x", Message: "expiring"}); err != nil {
		t.Fatal(err)
	}
}

func TestSendNon2xxLogsWarning(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	var buf bytes.Buffer
	old := errorWriter
	errorWriter = &buf
	t.Cleanup(func() { errorWriter = old })

	if err := Send(srv.URL, Payload{Event: EventBackupFailure, Stack: "x", Message: "fail"}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "HTTP 500") {
		t.Errorf("expected warning in stderr buffer, got %q", buf.String())
	}
}

func TestSendNetworkErrorIsBestEffort(t *testing.T) {
	var buf bytes.Buffer
	old := errorWriter
	errorWriter = &buf
	t.Cleanup(func() { errorWriter = old })

	err := Send("http://127.0.0.1:1", Payload{Event: EventBackupFailure, Stack: "x", Message: "fail"})
	if err != nil {
		t.Errorf("network error should return nil, got %v", err)
	}
	if buf.Len() == 0 {
		t.Error("expected warning logged on network error")
	}
}

var _ io.Writer = (*bytes.Buffer)(nil)
