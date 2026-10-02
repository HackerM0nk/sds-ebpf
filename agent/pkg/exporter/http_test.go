package exporter

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/sds-ebpf/agent/pkg/events"
)

func TestHTTPExporterExport(t *testing.T) {
	var received events.CanonicalEvent

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		if got := r.Header.Get("Content-Type"); got != "application/json" {
			t.Fatalf("Content-Type = %q, want application/json", got)
		}
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		w.WriteHeader(http.StatusAccepted)
	}))
	defer server.Close()

	exporter := NewHTTPExporter(server.URL, server.Client())
	event := events.CanonicalEvent{
		SchemaVersion: "sds.event.v1",
		EventID:       "evt-1",
		ObservedAt:    time.Date(2026, 3, 21, 12, 0, 0, 0, time.UTC),
		Host:          events.Host{ID: "host-1", Hostname: "host-1"},
		Process:       events.Process{PID: 1234, PPID: 1, Comm: "api", Binary: "/bin/api", UID: 1000},
		Workload:      events.Workload{Kind: "container", ID: "cgroup:1001", Service: "api"},
		Sensor:        events.Sensor{Name: "sched_process_exec", SamplingMode: "sparse"},
		Event:         events.Payload{Kind: "process.exec", Category: "process", Data: map[string]interface{}{"filename": "/bin/api"}},
	}

	if err := exporter.Export(context.Background(), event); err != nil {
		t.Fatalf("Export() error = %v", err)
	}

	if received.EventID != event.EventID {
		t.Fatalf("received.EventID = %q, want %q", received.EventID, event.EventID)
	}
}

func TestHTTPExporterExportStatusError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad gateway", http.StatusBadGateway)
	}))
	defer server.Close()

	exporter := NewHTTPExporter(server.URL, server.Client())
	err := exporter.Export(context.Background(), events.CanonicalEvent{
		SchemaVersion: "sds.event.v1",
		EventID:       "evt-1",
		ObservedAt:    time.Date(2026, 3, 21, 12, 0, 0, 0, time.UTC),
		Host:          events.Host{ID: "host-1", Hostname: "host-1"},
		Process:       events.Process{PID: 1, PPID: 0, Comm: "proc", Binary: "proc", UID: 0},
		Workload:      events.Workload{Kind: "process", ID: "process:host-1:1", Service: "proc"},
		Sensor:        events.Sensor{Name: "sched_process_exec", SamplingMode: "continuous"},
		Event:         events.Payload{Kind: "process.exec", Category: "process", Data: map[string]interface{}{}},
	})

	if err == nil {
		t.Fatal("Export() error = nil, want non-nil")
	}
}
