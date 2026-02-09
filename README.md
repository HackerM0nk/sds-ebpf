# WIP: SDS eBPF: Graph-based Observability & Reasoning System

**Open-source platform for end-to-end observability and AI-powered reasoning across distributed systems**

[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](LICENSE)
[![eBPF](https://img.shields.io/badge/powered%20by-eBPF-orange.svg)](https://ebpf.io/)

---

## 🎯 Vision

Build a complete observability platform that:

1. **Collects** sparse telemetry from hosts and Kubernetes using eBPF
2. **Aggregates** data through open-source pipelines (OpenTelemetry, Kafka, ClickHouse)
3. **Builds** a unified knowledge graph of system architecture
4. **Reasons** using GraphRAG and LLMs to:
   - Explain components and dependencies
   - Generate live architecture diagrams
   - Auto-document infrastructure
   - Answer natural language questions about security, performance, and architecture

---

## 🏗️ Architecture Overview

<img width="1844" height="836" alt="image" src="https://github.com/user-attachments/assets/683b6da0-a599-4ad8-afa2-a7c686173f63" />


---

## 🚀 Quick Start

### Prerequisites (macOS)

```bash
brew install colima docker docker-compose
colima start --cpu 4 --memory 8
```

### Run

```bash
git clone <your-repo-url> sds-ebpf
cd sds-ebpf
docker build -t sds-observer:latest -f docker/Dockerfile.agent .
cd docker && docker-compose up
```

### Verify

```bash
tail -f docker/output/events_*.json | jq .
```

**See full instructions:** [Quick Start Guide](docs/QUICKSTART.md)

---

## 📦 Current Status

### ✅ Phase 1: Baseline eBPF Agent (COMPLETE)

**What's Built:**
- Sparse-sampling eBPF agent (60s on / 4min off)
- Monitors: processes, syscalls, network connections, containers
- User-space collector in Go
- JSON event output
- Docker setup for local macOS development

**Documentation:**
- [Phase 1 Details](docs/PHASE_1.md)
- [macOS Local Development](docs/LOCAL_DEV_macOS.md)
- [Architecture Diagrams](diagrams/)

**Try it:**
```bash
cd agent
make build
sudo ./bin/sds-observer --output /tmp/sds-observer --verbose
```

---

### 🔜 Phase 2: Telemetry Aggregation (NEXT)

**Planned:**
- OpenTelemetry Collector integration
- Kafka/Redpanda for streaming
- ClickHouse for event queries
- VictoriaMetrics for metrics
- Neo4j/Memgraph for graph storage
- ETL pipeline to build unified graph schema

---

### 🔮 Phase 3-5: ML, Visualization, Documentation (FUTURE)

- **Phase 3:** GraphRAG indexing, LLM training on architecture data
- **Phase 4:** Interactive diagrams with LLM explanations
- **Phase 5:** Auto-generated documentation and testing guides

---

## 📂 Project Structure

```
sds-ebpf/
├── agent/                  # eBPF agent (C + Go)
│   ├── ebpf/               # eBPF programs (C)
│   │   ├── observer.bpf.c  # Main eBPF program
│   │   └── vmlinux.h       # Kernel type definitions
│   ├── cmd/observer/       # Go entry point
│   ├── pkg/                # Go packages
│   │   ├── collector/      # Event collection & parsing
│   │   └── config/         # Configuration
│   ├── Makefile            # Build system
│   └── go.mod              # Go dependencies
│
├── collector/              # (Phase 2) OTel collector configs
├── storage/                # (Phase 2) Database schemas
├── ml/                     # (Phase 3) RAG & LLM scripts
├── diagrams/               # Architecture diagrams
│   ├── phase1-architecture.mmd       # Mermaid
│   ├── phase1-architecture.d2        # D2
│   └── phase1-sparse-sampling.mmd    # Sequence diagram
│
├── docs/                   # Documentation
│   ├── PHASE_1.md          # Phase 1 complete guide
│   ├── LOCAL_DEV_macOS.md  # macOS development
│   └── QUICKSTART.md       # Quick start
│
└── docker/                 # Docker & docker-compose
    ├── Dockerfile.agent    # Agent container
    └── docker-compose.yml  # Local testing setup
```

---

## 🧩 Key Features

### Phase 1 Features

| Feature | Status | Description |
|---------|--------|-------------|
| **Sparse Sampling** | ✅ | Collect for 60s every 5min (configurable) |
| **Process Monitoring** | ✅ | Exec, exit, PID, PPID, UID, GID |
| **Syscall Tracing** | ✅ | Sample key syscalls with duration |
| **Network Monitoring** | ✅ | TCP connections with src/dst IP:port |
| **Container Awareness** | ✅ | Cgroup IDs for correlation |
| **JSON Output** | ✅ | Structured events to files |
| **HTTP Export** | ✅ | Optional endpoint for streaming (prep for Phase 2) |
| **Docker Support** | ✅ | Local testing on macOS via Colima/Docker Desktop |
| **CO-RE** | ✅ | Portable across kernel versions (using cilium/ebpf) |

---

## 🛠️ Technology Stack

### Phase 1 (Current)

| Component | Technology | Purpose |
|-----------|------------|---------|
| **eBPF Programs** | C (libbpf) | Kernel instrumentation |
| **User-space Collector** | Go 1.21+ | Event processing |
| **eBPF Library** | cilium/ebpf | CO-RE support |
| **Output Format** | JSON (NDJSON) | Structured events |
| **Container Runtime** | Docker | Local development |

### Phase 2 (Planned)

| Component | Technology | Purpose |
|-----------|------------|---------|
| **Telemetry Pipeline** | OpenTelemetry Collector | Normalize & route events |
| **Streaming** | Kafka / Redpanda | Event streaming |
| **Event Store** | ClickHouse | Fast analytical queries |
| **Metrics** | VictoriaMetrics | Time-series data |
| **Graph DB** | Neo4j / Memgraph | Relationship modeling |
| **Vector DB** | Weaviate / LanceDB | Embeddings for RAG |

---

## 📊 Performance

**Measured on Ubuntu 22.04, 4 vCPUs, 8GB RAM:**

| Metric | Sampling ON (60s) | Sampling OFF (240s) |
|--------|-------------------|---------------------|
| CPU Usage | 2-5% | <0.1% |
| Memory | 50-80 MB | 20 MB |
| Events/sec | 50-200 | 0 |
| Disk I/O | 1-5 MB/min | 0 |

**Scalability:** Tested with 500+ concurrent processes without drops.

---

## 🤝 Contributing

This is an open-source project. Contributions welcome!

### Development Setup

1. **Fork and clone**
2. **Set up local environment:** See [macOS Dev Guide](docs/LOCAL_DEV_macOS.md)
3. **Make changes**
4. **Test locally**
5. **Submit PR**

### Areas for Contribution

- eBPF program enhancements (more protocols, better filtering)
- Go collector optimizations
- Documentation improvements
- Phase 2-5 implementation
- Performance benchmarking
- Kubernetes manifests

---

## 📚 Documentation

### Quick Start
- **[Quick Start](docs/QUICKSTART.md)** - Get running in 5 minutes
- **[Complete Guide](docs/COMPLETE_GUIDE.md)** - Full platform documentation with use cases

### Technical Documentation
- **[Phase 1 Guide](docs/PHASE_1.md)** - eBPF agent implementation details
- **[macOS Development](docs/LOCAL_DEV_macOS.md)** - Local setup on macOS
- **[Phase 1 Summary](docs/PHASE_1_SUMMARY.md)** - Implementation summary

### Architecture & Diagrams
- **[Diagrams Guide](docs/DIAGRAMS.md)** - All Excalidraw architecture diagrams
- **[Instrumentation Overview](diagrams/01-instrumentation-overview.excalidraw)** - How we instrument different environments
- **[End-to-End Pipeline](diagrams/end-to-end-pipeline.excalidraw)** - Complete data flow
- **[Phase 2 Architecture](diagrams/phase2-aggregation-architecture.excalidraw)** - Aggregation & storage layer

---

## 🔐 Security & Privacy

This is a **defensive security tool** for observability and monitoring.

- ✅ Monitors system activity for understanding architecture
- ✅ No data exfiltration (local storage by default)
- ✅ Open-source and auditable
- ❌ Not for malicious use
- ❌ Does not harvest credentials or sensitive data

**Note:** Always review your organization's policies before deploying observability tools.

---

## 📝 License

Apache 2.0 License (to be added)

---

## 🙏 Acknowledgments

Built with:
- [eBPF](https://ebpf.io/)
- [cilium/ebpf](https://github.com/cilium/ebpf)
- [libbpf](https://github.com/libbpf/libbpf)

---

## 📬 Contact

Questions or feedback? Open an issue or reach out!

---

**Status:** Phase 1 Complete ✅ | Phase 2 In Planning 🔜
