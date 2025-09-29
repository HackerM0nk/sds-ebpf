# Local Development on macOS

Complete guide for developing and testing the SDS eBPF Observer on macOS.

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Setup Options](#setup-options)
3. [Quick Start](#quick-start)
4. [Development Workflow](#development-workflow)
5. [Troubleshooting](#troubleshooting)

---

## Prerequisites

### Required Tools

```bash
# Install Homebrew (if not already installed)
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

# Install required tools
brew install go docker docker-compose colima
```

### Optional Tools (for visualization)

```bash
# For viewing Mermaid diagrams
brew install mermaid-cli

# For viewing D2 diagrams
brew install d2
```

---

## Setup Options

### Option 1: Colima (Recommended for M1/M2/M3 Macs)

Colima provides a lightweight Docker runtime with better performance on Apple Silicon.

#### Initial Setup

```bash
# Start Colima with required settings for eBPF
colima start \
  --cpu 4 \
  --memory 8 \
  --disk 50 \
  --vm-type=vz \
  --mount-type=virtiofs \
  --mount /Users:/Users:w

# Verify Colima is running
colima status
```

#### Configure Docker Context

```bash
# Set Docker to use Colima
docker context use colima

# Verify
docker ps
```

#### Build and Run

```bash
cd /path/to/sds-ebpf

# Build the Docker image
docker build -t sds-observer:latest -f docker/Dockerfile.agent .

# Run with docker-compose
cd docker
mkdir -p output
docker-compose up

# View events
tail -f output/events_*.json | jq .
```

---

### Option 2: Docker Desktop

#### Initial Setup

1. Download and install [Docker Desktop for Mac](https://www.docker.com/products/docker-desktop/)

2. **Enable required features:**
   - Open Docker Desktop settings
   - Go to "Resources" → "Advanced"
   - Allocate at least:
     - CPUs: 4
     - Memory: 8 GB
     - Disk: 50 GB

3. **Apply and restart**

#### Build and Run

```bash
cd /path/to/sds-ebpf

# Build and run
cd docker
docker-compose up --build

# View events in another terminal
tail -f output/events_*.json | jq .
```

---

### Option 3: Kind (Kubernetes in Docker)

For testing Kubernetes deployments locally.

#### Setup Kind Cluster

```bash
# Install Kind
brew install kind kubectl

# Create cluster with extra mounts (for eBPF)
cat <<EOF | kind create cluster --name sds-observability --config=-
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
  extraMounts:
  - hostPath: /sys/kernel/debug
    containerPath: /sys/kernel/debug
  - hostPath: /sys/fs/bpf
    containerPath: /sys/fs/bpf
EOF

# Verify cluster
kubectl cluster-info --context kind-sds-observability
```

#### Deploy Observer

```bash
# Build image
cd agent
make docker-build

# Load into Kind
kind load docker-image sds-observer:latest --name sds-observability

# Deploy (will create in next phase)
# kubectl apply -f deploy/k8s/daemonset.yaml
```

---

## Quick Start

### 1. Clone and Setup

```bash
# Clone repository
git clone https://github.com/your-org/sds-ebpf.git
cd sds-ebpf

# Start Colima (if not already running)
colima start --cpu 4 --memory 8
```

### 2. Build and Run

```bash
# Build Docker image
docker build -t sds-observer:latest -f docker/Dockerfile.agent .

# Start observer
cd docker
mkdir -p output
docker-compose up
```

### 3. Verify Collection

In another terminal:

```bash
# Watch events being collected
watch -n 1 'ls -lh docker/output/ | tail -5'

# View events
cd docker/output
cat events_*.json | jq -r '.type' | sort | uniq -c

# Expected output:
#   15 network.tcp_connect
#   10 process.exec
#   10 process.exit
#    5 syscall
```

### 4. Stop

```bash
# Stop observer (in the docker-compose terminal)
Ctrl+C

# Or from another terminal
cd docker
docker-compose down
```

---

## Development Workflow

### Typical Development Cycle

```bash
# 1. Edit code
vim agent/ebpf/observer.bpf.c
# or
vim agent/pkg/collector/collector.go

# 2. Rebuild Docker image
docker build -t sds-observer:latest -f docker/Dockerfile.agent .

# 3. Restart observer
cd docker
docker-compose down
docker-compose up

# 4. Test changes
tail -f output/events_*.json | jq .
```

### Fast Iteration (Go-only changes)

If you're only changing Go code (not eBPF):

```bash
# Build binary on host (requires Go 1.21+)
cd agent
go build -o bin/sds-observer ./cmd/observer

# Run in Docker with host binary mounted
docker run --rm --privileged \
  --pid=host \
  -v $(pwd)/bin/sds-observer:/usr/local/bin/sds-observer \
  -v /sys/kernel/debug:/sys/kernel/debug:ro \
  -v /tmp/sds-observer:/var/log/sds-observer \
  ubuntu:22.04 \
  /usr/local/bin/sds-observer --output /var/log/sds-observer --verbose
```

### Debugging

#### View Container Logs

```bash
docker logs sds-observer --follow
```

#### Exec into Container

```bash
docker exec -it sds-observer bash

# Inside container:
ps aux | grep sds-observer
ls -lh /var/log/sds-observer/
cat /var/log/sds-observer/events_*.json | head -10
```

#### Check eBPF Programs Loaded

```bash
docker exec -it sds-observer bash -c "bpftool prog list"
```

---

## Architecture Visualization

### View Mermaid Diagrams

```bash
# Generate PNG from Mermaid
mmdc -i diagrams/phase1-architecture.mmd -o diagrams/phase1-architecture.png

# Open in browser
open diagrams/phase1-architecture.png
```

### View D2 Diagrams

```bash
# Generate SVG from D2
d2 diagrams/phase1-architecture.d2 diagrams/phase1-architecture.svg

# Open in browser
open diagrams/phase1-architecture.svg
```

---

## Troubleshooting

### Colima Issues

#### Colima won't start

```bash
# Check status
colima status

# View logs
colima logs

# Try resetting
colima delete
colima start --cpu 4 --memory 8 --vm-type=vz --mount-type=virtiofs
```

#### Docker context issues

```bash
# List contexts
docker context ls

# Switch to Colima
docker context use colima

# Test
docker ps
```

### Docker Desktop Issues

#### Container fails with "permission denied"

**Solution:** Ensure Docker Desktop has "Use kernel networking for UDP" disabled:
1. Open Docker Desktop
2. Settings → Resources → Network
3. Uncheck "Use kernel networking for UDP"
4. Apply & Restart

#### Slow performance

**Solution:** Increase resource allocation:
1. Docker Desktop → Settings → Resources
2. Increase CPUs to 4-6
3. Increase Memory to 8-12 GB
4. Apply & Restart

### eBPF-Specific Issues

#### "BTF not available" in container

This is expected in Docker containers running older kernels. The minimal vmlinux.h stub in `agent/ebpf/vmlinux.h` provides basic type definitions.

**For full BTF support:**
1. Use a host with kernel 5.4+
2. Generate vmlinux.h on that host:
   ```bash
   bpftool btf dump file /sys/kernel/btf/vmlinux format c > agent/ebpf/vmlinux.h
   ```

#### Events not being captured

**Check:**
1. Container has `--privileged` flag:
   ```bash
   docker inspect sds-observer | grep Privileged
   # Should show: "Privileged": true
   ```

2. `/sys/kernel/debug` is mounted:
   ```bash
   docker exec sds-observer ls /sys/kernel/debug
   ```

3. Observer is running:
   ```bash
   docker exec sds-observer ps aux | grep sds-observer
   ```

### General macOS Issues

#### "Architecture mismatch" errors (M1/M2/M3)

If building fails with architecture errors:

```bash
# Build for linux/amd64 (if your cluster runs x86)
docker buildx build --platform linux/amd64 \
  -t sds-observer:latest \
  -f docker/Dockerfile.agent .

# Or build for linux/arm64 (for ARM-based clusters)
docker buildx build --platform linux/arm64 \
  -t sds-observer:latest \
  -f docker/Dockerfile.agent .
```

#### "No space left on device"

```bash
# Clean Docker system
docker system prune -a --volumes

# If using Colima, increase disk size
colima stop
colima start --cpu 4 --memory 8 --disk 100
```

---

## Performance Tips

### 1. Use Colima with VZ (fastest)

```bash
colima start --vm-type=vz --mount-type=virtiofs
```

This uses macOS Virtualization.framework (faster than QEMU).

### 2. Limit Log Output

When not debugging, disable verbose mode:

```yaml
# In docker/docker-compose.yml
command: ["--output", "/var/log/sds-observer", "--sampling-on", "60", "--sampling-off", "240"]
```

### 3. Tune Sampling Parameters

For faster iteration during development:

```bash
# Shorter sampling cycles (30s on / 30s off)
docker run ... sds-observer:latest \
  --sampling-on 30 \
  --sampling-off 30 \
  --output /var/log/sds-observer
```

---

## Next Steps

Once you've verified the observer works locally:

1. ✅ Review collected events in `docker/output/events_*.json`
2. ✅ Experiment with sampling parameters
3. ✅ Generate workload and observe captures
4. → Move to **Phase 2**: Set up telemetry aggregation pipeline

---

## Resources

- [Colima Documentation](https://github.com/abiosoft/colima)
- [Docker Desktop for Mac](https://docs.docker.com/desktop/mac/install/)
- [Kind Quick Start](https://kind.sigs.k8s.io/docs/user/quick-start/)
- [eBPF on macOS (limitations)](https://ebpf.io/what-is-ebpf#development-toolchains)

---

**Have issues?** File a bug report with:
- macOS version: `sw_vers`
- Docker version: `docker version`
- Colima version (if used): `colima version`
- Error logs from `docker logs sds-observer`