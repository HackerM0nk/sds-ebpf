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

## 🧭 Project Posture

This repository is currently an **early-stage research artifact scaffold**, not a finished observability platform.

**What exists today:**
- Initial eBPF program prototypes
- Go collector skeleton and local runtime wrapper
- Architecture, roadmap, and research design docs
- Docker-based local development setup

**What is still missing:**
- Real end-to-end eBPF loading and event ingestion path
- Stable graph-building pipeline and query layer
- Reproducible evaluation harness and benchmark results
- Release-quality packaging for a publishable artifact

**Build direction:**
- [Vision](docs/VISION.md)
- [Artifact Spec v0.1](docs/ARTIFACT_SPEC_V0_1.md)
- [Engineering Loop](docs/ENGINEERING_LOOP.md)
- [Execution Plan](docs/EXECUTION_PLAN.md)
- [Roadmap](ROADMAP.md)
- [Research Blueprint](docs/RESEARCH_BLUEPRINT.md)
- [Novelty Map](docs/NOVELTY_MAP.md)
- [Workload Plan](eval/WORKLOAD_PLAN.md)
- [Polar Bookshop Demo Artifact](eval/workloads/polar-bookshop/README.md)

---

## 🚀 Quick Start

### First Demo Artifact

The fastest way to exercise SDS as an artifact, not just a prototype collector, is the first workload-backed demo:

```bash
make polar-demo
```

That path:
- starts a bounded observer capture,
- runs the Polar Bookshop microservice stack,
- drives a deterministic traffic scenario,
- builds a namespace-filtered runtime graph,
- and evaluates the inferred dependencies against workload-specific ground truth.

See [Polar Bookshop Demo Artifact](eval/workloads/polar-bookshop/README.md) for details.

**Note:** The current quick-start path is for the prototype scaffold and is expected to evolve during Milestone 0 and Milestone 1 hardening.

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
tail -f docker/output/events_*.ndjson | jq .
```

**See full instructions:** [Quick Start Guide](docs/QUICKSTART.md)

---

## 📦 Current Status

### 🧪 Prototype Scaffold (CURRENT)

**Implemented in some form:**
- Draft eBPF programs for process, syscall, and TCP visibility
- Go-based collector and sparse-sampling control loop
- NDJSON output path and Docker-based local development setup
- Broad system design docs and diagrams for the full platform

**Not yet artifact-grade:**
- Real bpf2go-based loader and attach flow
- Stable event schema with provenance and versioning
- Graph construction pipeline from runtime events
- Reproducible benchmarks, tests, and release process

**Use this repo as:**
- A foundation for a serious systems artifact
- The implementation base for a research paper
- A staged GitHub project with milestones rather than a finished product

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
├── .github/workflows/      # CI for the research loop
├── agent/                  # eBPF agent prototype (C + Go)
│   ├── ebpf/               # eBPF programs and kernel type stubs
│   ├── cmd/observer/       # Go entry point
│   ├── pkg/                # Collector and config packages
│   ├── Makefile            # Build targets
│   └── go.mod              # Go module definition
├── diagrams/               # Architecture and pipeline diagrams
├── docker/                 # Local container setup
├── docs/                   # Vision, blueprint, spec, and workflow docs
├── eval/                   # Sample traces, ground truth, and eval outputs
├── prompts/                # Planning prompts for later phases
├── schema/                 # Canonical event and graph model contracts
├── scripts/                # Validation, graph build, and eval automation
├── Makefile                # Root engineering loop targets
└── ROADMAP.md              # Artifact and GitHub project roadmap
```

---

## 🧩 Key Features

### Current Prototype Features

| Feature | Status | Description |
|---------|--------|-------------|
| **Sparse Sampling Loop** | 🧪 | Control loop exists in Go, but needs validation against real sensors |
| **Process Monitoring** | 🧪 | Draft eBPF program exists; end-to-end ingestion path is incomplete |
| **Syscall Tracing** | 🧪 | Prototype sensor exists; not yet artifact-grade |
| **Network Monitoring** | 🧪 | TCP connect prototype exists in eBPF source |
| **Container Awareness** | 🧪 | Cgroup-oriented fields are present, enrichment pipeline is missing |
| **JSON Output** | 🧪 | NDJSON-style local output path exists in prototype form |
| **HTTP Export** | 🚧 | Stubbed in code, not implemented end to end |
| **Docker Support** | 🧪 | Local dev scaffolding exists, but reproducibility still needs work |
| **CO-RE Portability** | 🚧 | Intended design target; current `vmlinux.h` path is still provisional |

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
