package events

import (
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	"github.com/sds-ebpf/agent/pkg/collector"
	"github.com/sds-ebpf/agent/pkg/config"
	"github.com/sds-ebpf/agent/pkg/identity"
)

const schemaVersion = "sds.event.v1"

type Host struct {
	ID       string `json:"id"`
	Hostname string `json:"hostname"`
}

type Process struct {
	PID    uint32 `json:"pid"`
	PPID   uint32 `json:"ppid"`
	Comm   string `json:"comm"`
	Binary string `json:"binary"`
	UID    uint32 `json:"uid"`
}

type Workload struct {
	Kind      string `json:"kind,omitempty"`
	ID        string `json:"id,omitempty"`
	Service   string `json:"service"`
	Namespace string `json:"namespace,omitempty"`
	CgroupID  string `json:"cgroup_id,omitempty"`
}

type Sensor struct {
	Name         string `json:"name"`
	SamplingMode string `json:"sampling_mode"`
}

type Payload struct {
	Kind     string                 `json:"kind"`
	Category string                 `json:"category"`
	Data     map[string]interface{} `json:"data"`
}

type CanonicalEvent struct {
	SchemaVersion string    `json:"schema_version"`
	EventID       string    `json:"event_id"`
	ObservedAt    time.Time `json:"observed_at"`
	Host          Host      `json:"host"`
	Process       Process   `json:"process"`
	Workload      Workload  `json:"workload"`
	Sensor        Sensor    `json:"sensor"`
	Event         Payload   `json:"event"`
}

type Builder struct {
	host         Host
	samplingMode string
	sequence     atomic.Uint64
	resolver     identity.Resolver
}

func NewBuilder(cfg *config.Config) (*Builder, error) {
	return NewBuilderWithResolver(cfg, identity.NewResolver(cfg))
}

func NewBuilderWithResolver(cfg *config.Config, resolver identity.Resolver) (*Builder, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return nil, fmt.Errorf("resolve hostname: %w", err)
	}

	samplingMode := "continuous"
	if cfg != nil && cfg.SamplingOff > 0 {
		samplingMode = "sparse"
	}

	return &Builder{
		host: Host{
			ID:       hostname,
			Hostname: hostname,
		},
		samplingMode: samplingMode,
		resolver:     resolver,
	}, nil
}

func (b *Builder) Build(raw collector.Event) CanonicalEvent {
	observedAt := raw.Timestamp.UTC()
	if observedAt.IsZero() {
		observedAt = time.Now().UTC()
	}

	binary := binaryName(raw)
	data := maps.Clone(raw.Data)
	cgroupID := stringify(data["cgroup_id"])
	service := serviceName(raw, binary)
	namespace := ""
	workloadKind := "process"
	workloadID := fmt.Sprintf("process:%s:%d", b.host.ID, raw.PID)

	if b.resolver != nil {
		if metadata, ok := b.resolver.ResolveProcess(raw.PID); ok {
			if metadata.Service != "" {
				service = metadata.Service
			}
			if metadata.Namespace != "" {
				namespace = metadata.Namespace
			}
			if metadata.ContainerID != "" {
				workloadKind = "container"
				workloadID = "container:" + metadata.ContainerID
			}
		}

		if raw.Type == "network.tcp_connect" {
			if metadata, ok := b.resolver.ResolveAddress(stringValue(data, "dst_addr")); ok {
				if metadata.Service != "" {
					data["dst_service"] = metadata.Service
				}
				if metadata.Namespace != "" {
					data["dst_namespace"] = metadata.Namespace
				}
				data["dst_scope"] = "internal"
			}
		}
	}

	if cgroupID != "" {
		workloadKind = "container"
		if !strings.HasPrefix(workloadID, "container:") {
			workloadID = "cgroup:" + cgroupID
		}
	}

	return CanonicalEvent{
		SchemaVersion: schemaVersion,
		EventID:       b.nextID(observedAt),
		ObservedAt:    observedAt,
		Host:          b.host,
		Process: Process{
			PID:    raw.PID,
			PPID:   uint32Value(raw.Data, "ppid"),
			Comm:   defaultString(raw.Comm, service),
			Binary: binary,
			UID:    raw.UID,
		},
		Workload: Workload{
			Kind:      workloadKind,
			ID:        workloadID,
			Service:   service,
			Namespace: namespace,
			CgroupID:  cgroupID,
		},
		Sensor: Sensor{
			Name:         sensorName(raw.Type),
			SamplingMode: b.samplingMode,
		},
		Event: Payload{
			Kind:     raw.Type,
			Category: category(raw.Type),
			Data:     data,
		},
	}
}

func (b *Builder) nextID(observedAt time.Time) string {
	sequence := b.sequence.Add(1)
	return fmt.Sprintf("evt-%s-%06d", observedAt.Format("20060102T150405.000000000Z07:00"), sequence)
}

func sensorName(eventType string) string {
	switch eventType {
	case "process.exec":
		return "sched_process_exec"
	case "process.exit":
		return "sched_process_exit"
	case "network.tcp_connect":
		return "tcp_connect"
	case "syscall":
		return "sys_exit"
	default:
		return "unknown"
	}
}

func category(eventType string) string {
	if prefix, _, ok := strings.Cut(eventType, "."); ok {
		return prefix
	}
	return eventType
}

func serviceName(raw collector.Event, binary string) string {
	if raw.Comm != "" {
		return raw.Comm
	}
	if binary != "" && binary != "unknown" {
		return filepath.Base(binary)
	}
	return "unknown"
}

func binaryName(raw collector.Event) string {
	if filename := stringValue(raw.Data, "filename"); filename != "" {
		return filename
	}
	if raw.Comm != "" {
		return raw.Comm
	}
	return "unknown"
}

func defaultString(value, fallback string) string {
	if value != "" {
		return value
	}
	return fallback
}

func stringValue(data map[string]interface{}, key string) string {
	if data == nil {
		return ""
	}
	if value, ok := data[key].(string); ok {
		return value
	}
	return ""
}

func uint32Value(data map[string]interface{}, key string) uint32 {
	if data == nil {
		return 0
	}
	switch value := data[key].(type) {
	case uint32:
		return value
	case uint64:
		return uint32(value)
	case int:
		return uint32(value)
	case int32:
		return uint32(value)
	case int64:
		return uint32(value)
	case float64:
		return uint32(value)
	default:
		return 0
	}
}

func stringify(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	case uint64:
		return fmt.Sprintf("%d", v)
	case uint32:
		return fmt.Sprintf("%d", v)
	case int:
		return fmt.Sprintf("%d", v)
	case int64:
		return fmt.Sprintf("%d", v)
	case float64:
		return fmt.Sprintf("%.0f", v)
	default:
		return ""
	}
}
