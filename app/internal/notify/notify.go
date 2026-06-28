package notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// errorWriter is where best-effort warning messages go.
var errorWriter io.Writer = os.Stderr

// Event represents a notification event type.
type Event string

const (
	EventContainerExit Event = "container_exit"
	EventBackupFailure Event = "backup_failure"
	EventCertExpiry    Event = "cert_expiry"
)

// Payload is the JSON body sent to the webhook endpoint.
type Payload struct {
	Event     Event     `json:"event"`
	Stack     string    `json:"stack"`
	Service   string    `json:"service,omitempty"`
	Timestamp time.Time `json:"timestamp"`
	Message   string    `json:"message"`
}

// Send POSTs a JSON payload to the webhook URL. Returns nil if webhookURL is
// empty. Non-2xx responses are logged to stderr but not returned as errors.
func Send(webhookURL string, p Payload) error {
	if webhookURL == "" {
		return nil
	}

	if p.Timestamp.IsZero() {
		p.Timestamp = time.Now().UTC()
	}

	body, err := json.Marshal(p)
	if err != nil {
		return fmt.Errorf("notify: marshaling payload: %w", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(webhookURL, "application/json", bytes.NewReader(body))
	if err != nil {
		// Best-effort — log and return nil.
		fmt.Fprintf(errorWriter, "⚠ webhook: %v\n", err)
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		fmt.Fprintf(errorWriter, "⚠ webhook: HTTP %d from %s\n", resp.StatusCode, webhookURL)
	}

	return nil
}
