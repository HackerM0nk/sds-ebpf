package exporter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/sds-ebpf/agent/pkg/events"
)

type EventExporter interface {
	Export(context.Context, events.CanonicalEvent) error
}

type HTTPExporter struct {
	client   *http.Client
	endpoint string
}

func NewHTTPExporter(endpoint string, client *http.Client) *HTTPExporter {
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Second}
	}

	return &HTTPExporter{
		client:   client,
		endpoint: endpoint,
	}
}

func (e *HTTPExporter) Export(ctx context.Context, event events.CanonicalEvent) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal canonical event: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, e.endpoint, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := e.client.Do(req)
	if err != nil {
		return fmt.Errorf("post event: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	return nil
}
