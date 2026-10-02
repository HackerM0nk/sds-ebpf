package identity

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/sds-ebpf/agent/pkg/config"
)

const defaultDockerSocketPath = "/var/run/docker.sock"

// Metadata describes container or workload identity resolved from the runtime.
type Metadata struct {
	ContainerID   string
	ContainerName string
	Service       string
	Namespace     string
}

// Resolver enriches process and address observations with workload identity.
type Resolver interface {
	ResolveProcess(pid uint32) (Metadata, bool)
	ResolveAddress(address string) (Metadata, bool)
}

type nullResolver struct{}

func (nullResolver) ResolveProcess(uint32) (Metadata, bool) {
	return Metadata{}, false
}

func (nullResolver) ResolveAddress(string) (Metadata, bool) {
	return Metadata{}, false
}

// NewResolver creates the best-effort runtime identity resolver for the active
// container runtime. It silently falls back to a no-op resolver when metadata
// sources are unavailable so the collector still works outside Docker.
func NewResolver(cfg *config.Config) Resolver {
	if cfg == nil {
		return nullResolver{}
	}

	switch strings.ToLower(cfg.ContainerRuntime) {
	case "", "docker":
		resolver, err := NewDockerResolver(defaultDockerSocketPath)
		if err == nil {
			return resolver
		}
	}

	return nullResolver{}
}

// DockerResolver resolves Docker Compose metadata from /proc and the Docker API.
type DockerResolver struct {
	client *http.Client

	mu          sync.RWMutex
	containers  map[string]Metadata
	ipIndex     map[string]string
	pidIndex    map[uint32]string
	lastRefresh time.Time
}

type dockerContainerSummary struct {
	ID              string            `json:"Id"`
	Names           []string          `json:"Names"`
	Labels          map[string]string `json:"Labels"`
	NetworkSettings struct {
		Networks map[string]struct {
			IPAddress string `json:"IPAddress"`
		} `json:"Networks"`
	} `json:"NetworkSettings"`
}

func NewDockerResolver(socketPath string) (*DockerResolver, error) {
	if socketPath == "" {
		socketPath = defaultDockerSocketPath
	}

	if _, err := os.Stat(socketPath); err != nil {
		return nil, fmt.Errorf("docker socket unavailable: %w", err)
	}

	transport := &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			dialer := &net.Dialer{Timeout: 2 * time.Second}
			return dialer.DialContext(ctx, "unix", socketPath)
		},
	}

	return &DockerResolver{
		client: &http.Client{
			Transport: transport,
			Timeout:   3 * time.Second,
		},
		containers: make(map[string]Metadata),
		ipIndex:    make(map[string]string),
		pidIndex:   make(map[uint32]string),
	}, nil
}

func (r *DockerResolver) ResolveProcess(pid uint32) (Metadata, bool) {
	r.mu.RLock()
	if containerID, ok := r.pidIndex[pid]; ok {
		if metadata, ok := r.lookupContainerLocked(containerID); ok {
			r.mu.RUnlock()
			return metadata, true
		}
	}
	r.mu.RUnlock()

	containerID, err := containerIDFromPID(pid)
	if err != nil || containerID == "" {
		return Metadata{}, false
	}

	r.mu.RLock()
	if metadata, ok := r.lookupContainerLocked(containerID); ok {
		r.mu.RUnlock()
		r.cachePID(pid, metadata.ContainerID)
		return metadata, true
	}
	r.mu.RUnlock()

	if err := r.refresh(); err != nil {
		return Metadata{}, false
	}

	r.mu.RLock()
	metadata, ok := r.lookupContainerLocked(containerID)
	r.mu.RUnlock()
	if ok {
		r.cachePID(pid, metadata.ContainerID)
	}
	return metadata, ok
}

func (r *DockerResolver) ResolveAddress(address string) (Metadata, bool) {
	if net.ParseIP(address) == nil {
		return Metadata{}, false
	}

	r.mu.RLock()
	if containerID, ok := r.ipIndex[address]; ok {
		if metadata, ok := r.lookupContainerLocked(containerID); ok {
			r.mu.RUnlock()
			return metadata, true
		}
	}
	r.mu.RUnlock()

	if err := r.refresh(); err != nil {
		return Metadata{}, false
	}

	r.mu.RLock()
	defer r.mu.RUnlock()
	if containerID, ok := r.ipIndex[address]; ok {
		return r.lookupContainerLocked(containerID)
	}
	return Metadata{}, false
}

func (r *DockerResolver) refresh() error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://docker/containers/json", nil)
	if err != nil {
		return err
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("docker api status %d", resp.StatusCode)
	}

	var payload []dockerContainerSummary
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return fmt.Errorf("decode docker container list: %w", err)
	}

	containers := make(map[string]Metadata, len(payload))
	ipIndex := make(map[string]string)

	for _, container := range payload {
		metadata := metadataFromContainer(container)
		if metadata.ContainerID == "" {
			continue
		}

		containers[metadata.ContainerID] = metadata
		for _, network := range container.NetworkSettings.Networks {
			if network.IPAddress == "" {
				continue
			}
			ipIndex[network.IPAddress] = metadata.ContainerID
		}
	}

	r.mu.Lock()
	r.containers = containers
	r.ipIndex = ipIndex
	r.lastRefresh = time.Now()
	r.mu.Unlock()
	return nil
}

func (r *DockerResolver) cachePID(pid uint32, containerID string) {
	if containerID == "" {
		return
	}

	r.mu.Lock()
	r.pidIndex[pid] = containerID
	r.mu.Unlock()
}

func (r *DockerResolver) lookupContainerLocked(containerID string) (Metadata, bool) {
	if metadata, ok := r.containers[containerID]; ok {
		return metadata, true
	}

	for knownID, metadata := range r.containers {
		if strings.HasPrefix(knownID, containerID) || strings.HasPrefix(containerID, knownID) {
			return metadata, true
		}
	}

	return Metadata{}, false
}

func metadataFromContainer(container dockerContainerSummary) Metadata {
	name := ""
	if len(container.Names) > 0 {
		name = strings.TrimPrefix(container.Names[0], "/")
	}

	service := container.Labels["com.docker.compose.service"]
	if service == "" {
		service = name
	}

	namespace := container.Labels["com.docker.compose.project"]
	return Metadata{
		ContainerID:   container.ID,
		ContainerName: name,
		Service:       service,
		Namespace:     namespace,
	}
}

func containerIDFromPID(pid uint32) (string, error) {
	procCgroupPath := filepath.Join("/proc", fmt.Sprintf("%d", pid), "cgroup")
	data, err := os.ReadFile(procCgroupPath)
	if err != nil {
		return "", err
	}
	return extractContainerID(string(data)), nil
}

func extractContainerID(contents string) string {
	for _, line := range strings.Split(contents, "\n") {
		if line == "" {
			continue
		}

		parts := strings.Split(line, ":")
		if len(parts) < 3 {
			continue
		}

		path := parts[len(parts)-1]
		for _, segment := range strings.Split(path, "/") {
			candidate := normalizeContainerIDSegment(segment)
			if candidate != "" {
				return candidate
			}
		}
	}

	return ""
}

func normalizeContainerIDSegment(segment string) string {
	if segment == "" {
		return ""
	}

	candidate := strings.TrimSuffix(segment, ".scope")
	for _, prefix := range []string{"docker-", "cri-containerd-", "crio-"} {
		candidate = strings.TrimPrefix(candidate, prefix)
	}

	if len(candidate) < 12 || len(candidate) > 64 {
		return ""
	}

	for _, ch := range candidate {
		switch {
		case ch >= '0' && ch <= '9':
		case ch >= 'a' && ch <= 'f':
		case ch >= 'A' && ch <= 'F':
		default:
			return ""
		}
	}

	return strings.ToLower(candidate)
}
