# Quick Start Guide

Get the SDS eBPF Observer running in **under 5 minutes**.

## For macOS Users

### Prerequisites

```bash
# Install Colima and Docker
brew install colima docker docker-compose

# Start Colima
colima start --cpu 4 --memory 8
```

### Run

```bash
# Clone and navigate
git clone <your-repo-url> sds-ebpf
cd sds-ebpf

# Build and start
docker build -t sds-observer:latest -f docker/Dockerfile.agent .
cd docker && mkdir -p output
docker-compose up
```

### Verify

```bash
# In another terminal
tail -f docker/output/events_*.json | jq .
```

You should see events like:

```json
{"type":"process.exec","timestamp":"2024-01-01T12:00:05Z","pid":1234,...}
{"type":"network.tcp_connect","timestamp":"2024-01-01T12:00:06Z","pid":5678,...}
```

**Success!** 🎉 Your observer is collecting telemetry.

---

## For Linux Users

### Prerequisites

```bash
# Ubuntu/Debian
sudo apt-get install -y clang llvm libbpf-dev linux-headers-$(uname -r) golang-1.21 make

# RHEL/Fedora
sudo dnf install -y clang llvm libbpf-devel kernel-devel golang make
```

### Build

```bash
cd agent
make build
```

### Run

```bash
# Run with sudo (eBPF requires root)
sudo ./bin/sds-observer --output /tmp/sds-observer --verbose
```

### Verify

```bash
# In another terminal
tail -f /tmp/sds-observer/events_*.json | jq .
```

---

## What's Happening?

The observer is:

1. **Every 5 minutes:**
   - Collecting for 60 seconds
   - Monitoring: processes, syscalls, network connections, containers
   - Writing events to JSON files

2. **Then sleeping for 4 minutes** (low overhead)

3. **Repeating the cycle**

This "sparse sampling" approach provides observability without continuous overhead.

---

## Next Steps

- 📖 Read [Phase 1 Documentation](PHASE_1.md) for details
- 🍎 See [macOS Development Guide](LOCAL_DEV_macOS.md) for advanced setup
- 🔍 Query your events:
  ```bash
  cat output/events_*.json | jq -r '.type' | sort | uniq -c
  ```

- 🚀 Ready for more? Proceed to **Phase 2** for data aggregation and graph building

---

## Troubleshooting

### "Permission denied"
→ Run with `sudo` (eBPF requires root privileges)

### "BTF not available"
→ Expected in Docker. Using minimal vmlinux.h stub (sufficient for Phase 1)

### No events being captured
→ Check if observer is running:
```bash
docker exec sds-observer ps aux | grep sds-observer
```

### More help
→ See [Troubleshooting in Phase 1 docs](PHASE_1.md#troubleshooting)