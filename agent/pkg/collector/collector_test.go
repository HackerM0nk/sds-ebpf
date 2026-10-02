package collector

import (
	"bytes"
	"encoding/binary"
	"testing"
	"time"
)

func TestParseSyscallEventRespectsStructAlignment(t *testing.T) {
	expectedAt := time.Date(2026, 3, 21, 18, 30, 0, 0, time.UTC)
	raw := rawSyscallEvent{
		Type:        EventTypeSyscall,
		Timestamp:   2_500_000_000,
		PID:         4242,
		UID:         1000,
		SyscallID:   42,
		DurationNS:  987654,
		SyscallName: [32]byte{'c', 'o', 'n', 'n', 'e', 'c', 't'},
		Comm:        [16]byte{'c', 'u', 'r', 'l'},
	}

	data := encodeRaw(t, raw)
	collector := &Collector{
		clockOffset: expectedAt.UnixNano() - int64(raw.Timestamp),
	}

	event, err := collector.parseEvent(data)
	if err != nil {
		t.Fatalf("parseEvent() error = %v", err)
	}

	if event.Type != "syscall" {
		t.Fatalf("event.Type = %q, want syscall", event.Type)
	}
	if event.Timestamp.UTC() != expectedAt {
		t.Fatalf("event.Timestamp = %s, want %s", event.Timestamp.UTC(), expectedAt)
	}
	if event.PID != raw.PID {
		t.Fatalf("event.PID = %d, want %d", event.PID, raw.PID)
	}
	if event.UID != raw.UID {
		t.Fatalf("event.UID = %d, want %d", event.UID, raw.UID)
	}
	if event.Comm != "curl" {
		t.Fatalf("event.Comm = %q, want curl", event.Comm)
	}
	if got := event.Data["syscall_name"]; got != "connect" {
		t.Fatalf("syscall_name = %v, want connect", got)
	}
	if got := event.Data["syscall_id"]; got != raw.SyscallID {
		t.Fatalf("syscall_id = %v, want %d", got, raw.SyscallID)
	}
	if got := event.Data["duration_ns"]; got != raw.DurationNS {
		t.Fatalf("duration_ns = %v, want %d", got, raw.DurationNS)
	}
}

func TestParseTCPConnectEventReadsAlignedTail(t *testing.T) {
	expectedAt := time.Date(2026, 3, 21, 18, 31, 0, 0, time.UTC)
	raw := rawTCPEvent{
		Type:      EventTypeTCPConnect,
		Timestamp: 1_750_000_000,
		PID:       99,
		UID:       501,
		SAddr:     0x0100007f,
		DAddr:     0x0201a8c0,
		SPort:     54321,
		DPort:     443,
		CgroupID:  123456789,
		Comm:      [16]byte{'n', 'g', 'i', 'n', 'x'},
	}

	data := encodeRaw(t, raw)
	collector := &Collector{
		clockOffset: expectedAt.UnixNano() - int64(raw.Timestamp),
	}

	event, err := collector.parseEvent(data)
	if err != nil {
		t.Fatalf("parseEvent() error = %v", err)
	}

	if event.Type != "network.tcp_connect" {
		t.Fatalf("event.Type = %q, want network.tcp_connect", event.Type)
	}
	if event.Timestamp.UTC() != expectedAt {
		t.Fatalf("event.Timestamp = %s, want %s", event.Timestamp.UTC(), expectedAt)
	}
	if got := event.Data["src_addr"]; got != "127.0.0.1" {
		t.Fatalf("src_addr = %v, want 127.0.0.1", got)
	}
	if got := event.Data["dst_addr"]; got != "192.168.1.2" {
		t.Fatalf("dst_addr = %v, want 192.168.1.2", got)
	}
	if got := event.Data["src_port"]; got != raw.SPort {
		t.Fatalf("src_port = %v, want %d", got, raw.SPort)
	}
	if got := event.Data["dst_port"]; got != raw.DPort {
		t.Fatalf("dst_port = %v, want %d", got, raw.DPort)
	}
	if got := event.Data["cgroup_id"]; got != raw.CgroupID {
		t.Fatalf("cgroup_id = %v, want %d", got, raw.CgroupID)
	}
}

func TestParseTCPConnectEventSkipsIncompleteSocketState(t *testing.T) {
	raw := rawTCPEvent{
		Type:      EventTypeTCPConnect,
		Timestamp: 1_750_000_000,
		PID:       77,
		UID:       0,
		SAddr:     0x0100007f,
		DAddr:     0,
		SPort:     12345,
		DPort:     0,
		Comm:      [16]byte{'c', 'u', 'r', 'l'},
	}

	collector := &Collector{clockOffset: time.Now().UnixNano() - int64(raw.Timestamp)}
	_, err := collector.parseEvent(encodeRaw(t, raw))
	if err == nil {
		t.Fatal("parseEvent() error = nil, want skip event error")
	}
	if err != errSkipEvent {
		t.Fatalf("parseEvent() error = %v, want %v", err, errSkipEvent)
	}
}

func encodeRaw(t *testing.T, raw interface{}) []byte {
	t.Helper()

	var buf bytes.Buffer
	if err := binary.Write(&buf, binary.LittleEndian, raw); err != nil {
		t.Fatalf("binary.Write() error = %v", err)
	}

	return buf.Bytes()
}
