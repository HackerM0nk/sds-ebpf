# SDS eBPF Architecture Diagrams

Interactive Excalidraw diagrams for the SDS eBPF observability platform.

## How to View

All diagrams are in **Excalidraw format** (`.excalidraw` files) and can be viewed:

1. **Online**: Visit [excalidraw.com](https://excalidraw.com) and open the file
2. **VS Code**: Install the [Excalidraw extension](https://marketplace.visualstudio.com/items?itemName=pomdtr.excalidraw-editor)
3. **CLI**: Use `excalidraw-cli` to export to PNG/SVG

---

## Phase 1: Baseline eBPF Agent

### 📊 Agent Architecture
**File**: [`phase1-agent-architecture.excalidraw`](../diagrams/phase1-agent-architecture.excalidraw)

**Shows:**
- Linux Kernel Space with eBPF programs
  - 📊 Process Exec/Exit Monitor
  - 🔧 Syscall Tracer (10% sampling)
  - 🌐 TCP Connect Monitor
  - 🐳 Container Metadata Collector
- Ring Buffer (256KB) for efficient event streaming
- User Space Go Collector
  - ⏱️ Sparse Sampling Controller (60s ON / 4m OFF)
  - 📦 Event Parser & Enricher
  - 🔄 JSON Serializer
  - 💾 File Writer
- Local Storage (JSON event files)

**Key Features:**
- Color-coded components (kernel=blue, userspace=orange, storage=purple)
- Data flow arrows showing event streaming
- Sparse sampling control logic

---

## End-to-End Pipeline

### 🔄 Complete Data Flow
**File**: [`end-to-end-pipeline.excalidraw`](../diagrams/end-to-end-pipeline.excalidraw)

**Shows:**
- **Data Collection Layer**
  - ☸️ Kubernetes Cluster with multiple nodes
  - 🖥️ Nodes running pods (auth-service, payment-api, user-db, api-gateway)
  - 🔍 SDS Agents (DaemonSet deployment)

- **Aggregation Layer**
  - 📡 OpenTelemetry Collector (normalize & route)
  - 🚀 Apache Kafka/Redpanda (event streaming)

- **Storage Layer**
  - ⚡ ClickHouse (analytical queries)
  - 📊 VictoriaMetrics (time-series metrics)
  - 🕸️ Neo4j (graph relationships)
  - 🧠 Weaviate (vector embeddings)

- **Visualization Layer**
  - 📈 Grafana dashboards

- **AI Reasoning Layer**
  - 🤖 GraphRAG & LLM
  - Natural language queries
  - Auto-documentation
  - Architecture visualization

**Key Features:**
- Proper distributed system symbols (Kafka, ClickHouse, Neo4j logos/colors)
- Multi-storage architecture
- ETL pipeline for graph building
- Phase 2 integration ready

---

## Phase 2: Telemetry Aggregation

### 📡 Aggregation Architecture
**File**: [`phase2-aggregation-architecture.excalidraw`](../diagrams/phase2-aggregation-architecture.excalidraw)

**Shows:**
- **5-Stage Pipeline:**
  1. 1️⃣ Data Collection (SDS Agents)
  2. 2️⃣ Normalization (OpenTelemetry Collector)
  3. 3️⃣ Multi-Storage (Kafka → ClickHouse, Neo4j, VictoriaMetrics, Weaviate)
  4. 4️⃣ ETL Transform (enrichment & graph building)
  5. 5️⃣ Query & Analyze

- **ETL Service** (Python/Go)
  - Consume from Kafka
  - Enrich with Docker/Kubernetes metadata
  - Build graph relationships
  - Cache queries in Redis

- **Query Layer**
  - Event Queries (ClickHouse): By time, service, aggregations
  - Graph Queries (Neo4j): Dependencies, relationships, path finding
  - Metrics Queries (VictoriaMetrics): Time-series, dashboards, alerts
  - Semantic Search (Weaviate): Similarity, embeddings

- **External Services**
  - 🐳 Docker API (container metadata)
  - ⚡ Redis (cache layer)

**Key Features:**
- Complete Phase 2 architecture
- ETL enrichment flow with metadata sources
- Multi-database query patterns
- Real distributed system symbols (Redis, Docker, Kafka colors/icons)

---

## Symbol Legend

### Colors Used

| Color | Component Type | Examples |
|-------|---------------|----------|
| 🟢 Green | Data Collection | eBPF Agents, SDS Observer |
| 🟠 Orange | Processing | OpenTelemetry Collector |
| ⚫ Black | Streaming | Kafka, Redpanda |
| 🟡 Yellow | Analytics DB | ClickHouse |
| 🔴 Red | Metrics | VictoriaMetrics, Redis |
| 🔵 Blue | Containers | Docker, Kubernetes, Neo4j |
| 🟣 Purple | AI/ML | ETL Service, LLM, GraphRAG |
| 🟢 Teal | Vectors | Weaviate |
| 🔷 Cyan | Query Layer | API interfaces |

### Icons Used

- 🔍 **Search/Monitoring**: SDS Agents, Observability
- 📡 **Communication**: OpenTelemetry, data transmission
- 🚀 **Speed/Streaming**: Kafka, high-throughput systems
- ⚡ **Fast Queries**: ClickHouse, VictoriaMetrics
- 🕸️ **Graph**: Neo4j, relationship databases
- 🧠 **Intelligence**: Vector DB, AI systems
- 📊 **Analytics**: Metrics, dashboards
- 📈 **Visualization**: Grafana
- 🤖 **AI**: LLM, GraphRAG
- 🐳 **Containers**: Docker, Pods
- ☸️ **Orchestration**: Kubernetes
- 🖥️ **Infrastructure**: Nodes, servers
- 🔄 **Processing**: ETL, transformation
- ⏱️ **Time**: Sampling, scheduling
- 💾 **Storage**: Filesystems, databases

---

## Exporting Diagrams

### To PNG

```bash
# Install excalidraw CLI
npm install -g @excalidraw/cli

# Export Phase 1 architecture
excalidraw-cli diagrams/phase1-agent-architecture.excalidraw \
  -o diagrams/phase1-agent-architecture.png \
  --width 2000

# Export end-to-end pipeline
excalidraw-cli diagrams/end-to-end-pipeline.excalidraw \
  -o diagrams/end-to-end-pipeline.png \
  --width 3000

# Export Phase 2 architecture
excalidraw-cli diagrams/phase2-aggregation-architecture.excalidraw \
  -o diagrams/phase2-aggregation-architecture.png \
  --width 3000
```

### To SVG (Vector Graphics)

```bash
excalidraw-cli diagrams/phase1-agent-architecture.excalidraw \
  -o diagrams/phase1-agent-architecture.svg
```

### To PDF

```bash
# First convert to SVG, then use cairosvg
excalidraw-cli diagrams/end-to-end-pipeline.excalidraw \
  -o diagrams/end-to-end-pipeline.svg

pip install cairosvg
cairosvg diagrams/end-to-end-pipeline.svg \
  -o diagrams/end-to-end-pipeline.pdf
```

---

## Editing Diagrams

### Online (Easiest)

1. Go to [excalidraw.com](https://excalidraw.com)
2. Click "Open" → Select the `.excalidraw` file
3. Edit freely
4. Export: Menu → Save → `.excalidraw`

### VS Code (Recommended for Development)

1. Install [Excalidraw Extension](https://marketplace.visualstudio.com/items?itemName=pomdtr.excalidraw-editor)
2. Open any `.excalidraw` file in VS Code
3. Edit visually with full tool support
4. Auto-saves on changes

### Libraries for Symbols

Excalidraw has built-in libraries with icons for:
- Cloud providers (AWS, Azure, GCP)
- Databases (Redis, MongoDB, PostgreSQL)
- Container tech (Docker, Kubernetes)
- Messaging (Kafka, RabbitMQ)
- Monitoring (Prometheus, Grafana)

**To add libraries:**
1. Open Excalidraw
2. Click library icon (books icon)
3. Browse → Select tech stack libraries

---

## Diagram Update Log

| Date | Diagram | Changes |
|------|---------|---------|
| 2024-01-01 | Phase 1 Architecture | Initial creation with full agent workflow |
| 2024-01-01 | End-to-End Pipeline | Complete Kubernetes → Storage → AI flow |
| 2024-01-01 | Phase 2 Aggregation | ETL pipeline with proper system symbols |

---

## Contributing

When adding new diagrams:

1. Use consistent color scheme (see legend above)
2. Include proper icons/symbols for distributed systems
3. Add clear data flow arrows
4. Label all components
5. Export to PNG for documentation
6. Update this README

---

## Alternative Formats (Legacy)

For reference, we also maintain:
- **Mermaid**: `*.mmd` files (text-based, Git-friendly)
- **D2**: `*.d2` files (declarative diagrams)

These are kept for CI/CD automation but Excalidraw is **primary** for visual clarity.

---

## Questions?

See:
- [Phase 1 Documentation](PHASE_1.md#architecture)
- [Phase 2 Prompt](../prompts/PHASE_2_PROMPT.md)
- [Excalidraw Documentation](https://docs.excalidraw.com/)