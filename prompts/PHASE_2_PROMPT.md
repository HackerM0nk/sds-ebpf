# Phase 2 Prompt: Telemetry Aggregation & Storage

**Context:** Phase 1 complete. eBPF agent collects events to JSON files.

---

## Goal

Build an end-to-end telemetry aggregation pipeline that:

1. **Ingests** JSON events from Phase 1 agent
2. **Normalizes** them using OpenTelemetry Collector
3. **Streams** through Kafka/Redpanda
4. **Stores** in multiple backends:
   - ClickHouse (for analytical queries)
   - VictoriaMetrics (for time-series metrics)
   - Neo4j/Memgraph (for graph relationships)
   - Weaviate/LanceDB (for vector embeddings)
5. **Transforms** events into a unified graph schema

---

## Requirements

### 1. OpenTelemetry Collector

**Setup:**
- Receive events via HTTP endpoint (from Phase 1 agent's `--http` flag)
- Parse JSON events
- Add metadata (host, cluster, environment)
- Route to multiple backends

**Output:**
```yaml
receivers:
  otlp:
    protocols:
      http:
        endpoint: 0.0.0.0:4318

processors:
  batch:
    timeout: 10s

  attributes:
    actions:
      - key: cluster
        value: local-dev
        action: upsert

exporters:
  kafka:
    brokers: [localhost:9092]
    topic: sds-events

  clickhouse:
    endpoint: tcp://localhost:9000
    database: observability

  prometheusremotewrite:
    endpoint: http://victoriametrics:8428/api/v1/write
```

### 2. Kafka/Redpanda

**Purpose:** Event streaming and replay capability

**Topics:**
- `sds-events-raw` - Raw events from OTel
- `sds-events-enriched` - After enrichment
- `sds-graph-updates` - Graph mutations

**Setup:**
- Single-node Redpanda for local dev
- Docker Compose setup
- Retention: 7 days

### 3. ClickHouse

**Purpose:** Fast analytical queries on events

**Schema:**
```sql
CREATE TABLE events (
    timestamp DateTime64(9),
    event_type LowCardinality(String),
    pid UInt32,
    uid UInt32,
    comm String,
    data String,  -- JSON for flexible schema
    host String,
    cluster String,
    INDEX idx_type event_type TYPE bloom_filter GRANULARITY 4,
    INDEX idx_comm comm TYPE bloom_filter GRANULARITY 4
) ENGINE = MergeTree()
PARTITION BY toYYYYMMDD(timestamp)
ORDER BY (timestamp, event_type, pid);
```

**Queries to support:**
- Events by time range
- Events by process/service
- Events by type (exec, network, etc.)
- Aggregations (top processes, most connections)

### 4. VictoriaMetrics

**Purpose:** Time-series metrics

**Metrics to derive:**
- `sds_process_starts_total{comm="bash"}` (counter)
- `sds_network_connections_total{dst_port="443"}` (counter)
- `sds_syscall_duration_seconds{syscall="connect"}` (histogram)

**Setup:**
- Single-node VictoriaMetrics
- Scrape metrics from OTel exporter
- Grafana dashboard (optional)

### 5. Neo4j (Graph Database)

**Purpose:** Model architectural relationships

**Graph Schema:**

**Nodes:**
- `Process` (pid, comm, uid)
- `Service` (name, derived from process patterns)
- `NetworkEndpoint` (ip, port)
- `Container` (cgroup_id, name, pod)

**Relationships:**
- `(Process)-[:SPAWNED_BY]->(Process)` (PPID relationship)
- `(Process)-[:RUNS_IN]->(Container)`
- `(Process)-[:CONNECTS_TO]->(NetworkEndpoint)`
- `(Service)-[:DEPENDS_ON]->(Service)` (inferred from network)

**Example Cypher:**
```cypher
CREATE (p:Process {pid: 1234, comm: "curl", uid: 1000})
CREATE (e:NetworkEndpoint {ip: "93.184.216.34", port: 443})
CREATE (p)-[:CONNECTS_TO {timestamp: datetime()}]->(e)
```

### 6. ETL Pipeline

**Implementation:** Python or Go service

**Steps:**
1. Consume from `sds-events-raw` Kafka topic
2. Parse events
3. Enrich with metadata (container names, service labels)
4. Build graph mutations
5. Write to Neo4j
6. Publish to `sds-graph-updates` topic

**Container Enrichment:**
```python
def enrich_with_container_metadata(event):
    cgroup_id = event['data']['cgroup_id']
    # Query Docker/containerd API
    container = docker_client.containers.get(cgroup_id)
    event['container_name'] = container.name
    event['pod_name'] = container.labels.get('io.kubernetes.pod.name')
    return event
```

---

## Deliverables

### Code

```
/
├── collector/
│   └── otel-config.yaml           # OpenTelemetry Collector config
│
├── storage/
│   ├── clickhouse/
│   │   └── schema.sql             # ClickHouse table schemas
│   ├── neo4j/
│   │   └── init.cypher            # Neo4j initialization
│   └── victoriametrics/
│       └── config.yml             # VictoriaMetrics config
│
├── etl/
│   ├── main.py                    # ETL service
│   ├── requirements.txt
│   └── Dockerfile
│
├── docker/
│   └── docker-compose.phase2.yml  # All services
│
└── docs/
    └── PHASE_2.md                 # Phase 2 documentation
```

### Diagrams

- **Mermaid/D2:** Data flow from agent → OTel → Kafka → Stores
- **Graph Schema:** Neo4j node/relationship diagram

### Documentation

- Setup instructions (docker-compose up)
- Example queries for each backend
- Performance benchmarks
- Integration with Phase 1 agent

---

## Testing

### Test Scenario

1. Start Phase 1 agent with `--http http://localhost:4318/v1/traces`
2. Generate workload (curl, process spawns)
3. Query ClickHouse: `SELECT * FROM events ORDER BY timestamp DESC LIMIT 10`
4. Query Neo4j: `MATCH (p:Process)-[:CONNECTS_TO]->(e) RETURN p, e LIMIT 10`
5. View metrics in VictoriaMetrics

### Success Criteria

- ✅ Events flow from agent → OTel → Kafka → Stores
- ✅ ClickHouse has queryable event data
- ✅ Neo4j graph shows process relationships
- ✅ VictoriaMetrics shows time-series metrics
- ✅ Sub-second latency from event to storage

---

## Output Format

Provide:

1. **All code and configs** as listed above
2. **docker-compose.phase2.yml** that starts:
   - OTel Collector
   - Redpanda
   - ClickHouse
   - VictoriaMetrics
   - Neo4j
   - ETL service
3. **Diagrams** (Mermaid + D2)
4. **docs/PHASE_2.md** with setup instructions
5. **Example queries and outputs** for each backend

---

## Open Questions

1. **Graph inference:** How to automatically detect "services" from process patterns?
2. **Deduplication:** How to handle duplicate events in Kafka?
3. **Schema evolution:** How to handle new event types in ClickHouse?

---

**Ready?** Start Phase 2 implementation with the structure above.