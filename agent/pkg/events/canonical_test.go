package events

import (
	"testing"
	"time"

	"github.com/sds-ebpf/agent/pkg/collector"
	"github.com/sds-ebpf/agent/pkg/config"
	"github.com/sds-ebpf/agent/pkg/identity"
)

type stubResolver struct {
	processMeta identity.Metadata
	addressMeta identity.Metadata
}

func (s stubResolver) ResolveProcess(uint32) (identity.Metadata, bool) {
	if s.processMeta.Service == "" && s.processMeta.ContainerID == "" {
		return identity.Metadata{}, false
	}
	return s.processMeta, true
}

func (s stubResolver) ResolveAddress(string) (identity.Metadata, bool) {
	if s.addressMeta.Service == "" {
		return identity.Metadata{}, false
	}
	return s.addressMeta, true
}

func TestBuildProcessExecCanonicalEvent(t *testing.T) {
	builder, err := NewBuilderWithResolver(config.Default(), stubResolver{
		processMeta: identity.Metadata{
			ContainerID: "container-123",
			Service:     "api-gateway",
			Namespace:   "demo-stack",
		},
	})
	if err != nil {
		t.Fatalf("NewBuilder() error = %v", err)
	}

	raw := collector.Event{
		Type:      "process.exec",
		Timestamp: time.Date(2026, 3, 21, 12, 0, 0, 0, time.UTC),
		PID:       1234,
		UID:       1000,
		Comm:      "api-gateway",
		Data: map[string]interface{}{
			"ppid":      uint32(1),
			"filename":  "/usr/local/bin/api-gateway",
			"cgroup_id": uint64(1001),
		},
	}

	event := builder.Build(raw)

	if event.SchemaVersion != "sds.event.v1" {
		t.Fatalf("SchemaVersion = %q, want sds.event.v1", event.SchemaVersion)
	}
	if event.Workload.Service != "api-gateway" {
		t.Fatalf("Workload.Service = %q, want api-gateway", event.Workload.Service)
	}
	if event.Workload.Kind != "container" {
		t.Fatalf("Workload.Kind = %q, want container", event.Workload.Kind)
	}
	if event.Workload.ID != "container:container-123" {
		t.Fatalf("Workload.ID = %q", event.Workload.ID)
	}
	if event.Workload.Namespace != "demo-stack" {
		t.Fatalf("Workload.Namespace = %q", event.Workload.Namespace)
	}
	if event.Process.Binary != "/usr/local/bin/api-gateway" {
		t.Fatalf("Process.Binary = %q", event.Process.Binary)
	}
	if event.Process.PPID != 1 {
		t.Fatalf("Process.PPID = %d, want 1", event.Process.PPID)
	}
	if event.Sensor.Name != "sched_process_exec" {
		t.Fatalf("Sensor.Name = %q", event.Sensor.Name)
	}
	if event.Event.Category != "process" {
		t.Fatalf("Event.Category = %q, want process", event.Event.Category)
	}
}

func TestBuildNetworkCanonicalEvent(t *testing.T) {
	builder, err := NewBuilderWithResolver(&config.Config{SamplingOn: 60}, stubResolver{
		processMeta: identity.Metadata{
			ContainerID: "container-5678",
			Service:     "billing-worker",
			Namespace:   "demo-stack",
		},
		addressMeta: identity.Metadata{
			ContainerID: "container-9001",
			Service:     "catalog-service",
			Namespace:   "demo-stack",
		},
	})
	if err != nil {
		t.Fatalf("NewBuilder() error = %v", err)
	}

	raw := collector.Event{
		Type:      "network.tcp_connect",
		Timestamp: time.Date(2026, 3, 21, 12, 0, 1, 0, time.UTC),
		PID:       5678,
		UID:       1000,
		Comm:      "billing-worker",
		Data: map[string]interface{}{
			"dst_addr": "93.184.216.34",
			"dst_port": uint16(443),
		},
	}

	event := builder.Build(raw)

	if event.Sensor.SamplingMode != "continuous" {
		t.Fatalf("Sensor.SamplingMode = %q, want continuous", event.Sensor.SamplingMode)
	}
	if event.Workload.Service != "billing-worker" {
		t.Fatalf("Workload.Service = %q", event.Workload.Service)
	}
	if event.Process.Binary != "billing-worker" {
		t.Fatalf("Process.Binary = %q, want billing-worker", event.Process.Binary)
	}
	if event.Event.Kind != "network.tcp_connect" {
		t.Fatalf("Event.Kind = %q", event.Event.Kind)
	}
	if got := event.Event.Data["dst_service"]; got != "catalog-service" {
		t.Fatalf("dst_service = %v, want catalog-service", got)
	}
	if got := event.Event.Data["dst_scope"]; got != "internal" {
		t.Fatalf("dst_scope = %v, want internal", got)
	}
}
