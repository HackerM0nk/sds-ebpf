# SDS eBPF: Complete Guide
## Building End-to-End Architecture Understanding from Runtime Telemetry

**Version**: 1.0
**Last Updated**: 2024-01-01
**Status**: Phase 1 Complete, Phase 2 In Planning

---

## Table of Contents

1. [Executive Summary](#executive-summary)
2. [The Problem We're Solving](#the-problem-were-solving)
3. [Why This Approach is Different](#why-this-approach-is-different)
4. [Architecture Overview](#architecture-overview)
5. [How It Works: Step by Step](#how-it-works-step-by-step)
6. [Value Proposition](#value-proposition)
7. [Use Cases](#use-cases)
8. [Technical Deep Dive](#technical-deep-dive)
9. [Comparison with Existing Tools](#comparison-with-existing-tools)
10. [Roadmap](#roadmap)

---

## Executive Summary

**SDS eBPF** is an open-source observability platform that automatically discovers, maps, and understands your entire distributed system architecture by observing actual runtime behavior using eBPF (Extended Berkeley Packet Filter) technology.

### What Makes It Unique

- **Zero Code Changes**: Uses kernel-level instrumentation - no SDKs, agents in-app, or code modifications
- **Universal Deployment**: Single agent works across VMs, containers, Kubernetes, on-premise, and cloud
- **Runtime Truth**: Captures what actually happens, not what's configured or documented
- **AI-Powered Understanding**: Automatically builds knowledge graphs and enables natural language queries
- **Security-First**: Designed for security and architecture teams to understand their attack surface

### Core Capabilities

1. **📊 Automatic Service Discovery**: Identifies all running services, APIs, and dependencies
2. **🕸️ Dependency Mapping**: Builds complete call graphs and data flow diagrams
3. **🔍 Security Posture**: Maps access patterns, privilege usage, and data flows
4. **🤖 AI-Powered Q&A**: Ask "What services access the payment database?" in plain English
5. **📈 Real-Time Architecture Docs**: Auto-generates and keeps architecture diagrams up-to-date

---

## The Problem We're Solving

### Current State of Infrastructure Understanding

Modern distributed systems have become impossible to understand through traditional means:

#### 1. **Documentation Drift**

```
Reality:
┌─────────────────────────────────────┐
│  Written Docs  │  Actual System    │
├────────────────┼───────────────────┤
│  3 services    │  47 microservices │
│  PostgreSQL    │  + MongoDB        │
│  No Kafka      │  + Redis cluster  │
│  Updated: 2021 │  Current: 2024    │
└─────────────────────────────────────┘
```

**Problem**: Documentation becomes outdated within weeks. No one knows the true architecture.

#### 2. **Configuration ≠ Runtime**

```yaml
# What's deployed (K8s YAML):
apiVersion: v1
kind: Service
name: auth-service
replicas: 3

# What's actually running:
auth-service-v1 (legacy, 1 pod, still handling 20% traffic)
auth-service-v2 (main, 3 pods)
auth-service-canary (experimental, 1 pod, talking to prod DB!)
```

**Problem**: Static configuration doesn't show you what's actually executing or how services interact.

#### 3. **Sprawl Across Environments**

```
Your Infrastructure:
├── AWS: 127 EC2 instances
├── ECS: 43 task definitions
├── EKS: 3 clusters, 890 pods
├── On-Premise: 15 physical servers
├── Azure (Shadow IT): 8 VMs
└── Someone's personal AWS account: ???
```

**Problem**: No unified view. Every environment has its own monitoring, none talk to each other.

#### 4. **Security Blind Spots**

**Questions Security Teams Can't Answer**:
- "Which services can access our customer PII database?"
- "What happens if auth-service is compromised?"
- "Do we have any services running as root that shouldn't be?"
- "What external APIs are we calling and with what data?"

**Problem**: Lack of runtime visibility = inability to assess true security posture.

#### 5. **The Debugging Nightmare**

```
Incident Timeline:
10:00 - Payment failures reported
10:15 - Check payment-service logs (nothing unusual)
10:30 - Check database (queries are fine)
10:45 - Check network (no issues)
11:00 - Discover: auth-service had OOM, restarted
11:15 - Realize: payment-service calls auth-service
11:30 - Still don't know why auth-service OOM'd
12:00 - Root cause: new feature added secret scanning
           which made 1000+ API calls per request
```

**Problem**: No visibility into actual call paths and resource usage patterns.

---

## Why This Approach is Different

### The eBPF Advantage

#### What is eBPF?

eBPF (Extended Berkeley Packet Filter) is a revolutionary Linux kernel technology that allows you to run sandboxed programs in the kernel without changing kernel source code or loading kernel modules.

**Think of it as**: JavaScript for the Linux kernel - safe, portable, and incredibly powerful.

#### Why eBPF for Observability?

| Traditional Approach | eBPF Approach |
|---------------------|---------------|
| Modify application code | Zero code changes |
| Install language-specific SDKs | Universal - works with any language |
| Per-service configuration | Deploy once, observe everything |
| Miss system-level details | See all syscalls, network, files |
| Performance overhead ~5-20% | Overhead <1% (with sparse sampling) |
| Requires app restarts | Attach/detach at runtime |
| Blind to compiled binaries | Observes any process |

### Unique Architectural Decisions

#### 1. **Sparse Sampling (60s ON / 4min OFF)**

**Why**: Continuous tracing is expensive and unnecessary for architecture understanding.

```
Traditional APM:     ████████████████████████ (100% CPU usage)
SDS eBPF (sparse):   ██      ██      ██       (<1% avg CPU)
                     ^60s    ^60s    ^60s
```

**Benefit**:
- Captures enough data to understand architecture (who talks to whom, how often)
- Minimal performance impact
- Scales to thousands of hosts

#### 2. **Multi-Storage Architecture**

We don't force all data into one database. Different query patterns need different stores:

```
ClickHouse:        "Show me all API calls in last hour"
Neo4j:             "What's the shortest path from web-ui to database?"
VictoriaMetrics:   "Plot CPU usage of payment-service"
Weaviate:          "Find services similar to auth-service"
```

**Why This Matters**: Architecture questions require graph queries. Time-series questions need time-series DB. One size doesn't fit all.

#### 3. **LLM-Powered Reasoning (Phase 3)**

Traditional tools: You must know what query to run.

SDS eBPF: Ask questions in plain English.

```
You ask: "What would happen if payments-db goes down?"

SDS responds:
- payment-service would fail (no fallback)
- order-service would also fail (depends on payment-service)
- frontend would show error (calls order-service)
- Impact: 87% of production traffic

Here's the dependency chain:
[Generated diagram showing: frontend → order-service → payment-service → payments-db]
```

---

## Architecture Overview

### High-Level System Design

```
┌─────────────────────────────────────────────────────────────────┐
│                    Your Infrastructure                           │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐       │
│  │ EC2/VMs  │  │ ECS      │  │ EKS/K8s  │  │ On-Prem  │       │
│  │          │  │          │  │          │  │          │       │
│  │ [Agent]  │  │ [Agent]  │  │ [Agent]  │  │ [Agent]  │       │
│  └────┬─────┘  └────┬─────┘  └────┬─────┘  └────┬─────┘       │
└───────┼─────────────┼─────────────┼─────────────┼──────────────┘
        │             │             │             │
        └──────────┬──┴──────┬──────┴──────┬──────┘
                   │         │             │
                   ▼         ▼             ▼
        ┌──────────────────────────────────────────────┐
        │        OpenTelemetry Collector               │
        │        (Normalize & Route)                   │
        └──────────────┬───────────────────────────────┘
                       │
                       ▼
        ┌──────────────────────────────────────────────┐
        │           Kafka / Redpanda                    │
        │           (Event Streaming)                   │
        └──┬────────┬────────┬────────┬────────────────┘
           │        │        │        │
     ┌─────▼──┐ ┌──▼────┐ ┌─▼────┐ ┌▼─────┐
     │ Click  │ │ Neo4j │ │ Vict.│ │Weavi.│
     │ House  │ │(Graph)│ │Metric│ │(Vect)│
     └────────┘ └───────┘ └──────┘ └──────┘
           │        │        │        │
           └────────┴────┬───┴────────┘
                         │
                    ┌────▼──────┐
                    │ ETL       │
                    │ Pipeline  │
                    └────┬──────┘
                         │
                    ┌────▼──────┐
                    │ GraphRAG  │
                    │ + LLM     │
                    └───────────┘
```

### Data Flow Explanation

**Step 1: Collection** (Phase 1 - ✅ Complete)
- eBPF agents run on every host/container
- Capture: process exec/exit, syscalls, network connections, container metadata
- Sparse sampling: 60s collection every 5 minutes
- Output: JSON events streamed to collector

**Step 2: Normalization** (Phase 2 - 🔜 Planning)
- OpenTelemetry Collector receives events from all agents
- Normalizes data format across different environments
- Adds metadata (cluster name, region, environment tags)
- Routes to multiple backends

**Step 3: Storage** (Phase 2)
- **ClickHouse**: Stores raw events for analytical queries
- **Neo4j**: Builds graph of services, processes, dependencies
- **VictoriaMetrics**: Time-series metrics (CPU, memory, request rates)
- **Weaviate**: Vector embeddings for semantic search

**Step 4: Graph Building** (Phase 2)
- ETL pipeline consumes events from Kafka
- Enriches with container/pod metadata (from Docker/K8s API)
- Infers relationships:
  - Process A → Network Connection → Process B = Service Dependency
  - Process → File Access = Data Flow
  - User → Process = Access Pattern
- Writes to Neo4j

**Step 5: AI Reasoning** (Phase 3 - 🔮 Future)
- GraphRAG indexes the knowledge graph
- LLM trained on architecture documentation
- Natural language interface for querying
- Auto-generates diagrams and documentation

---

## How It Works: Step by Step

### From Installation to Insight

#### Step 1: Deploy Agent (5 minutes)

**EC2 / Standalone**:
```bash
curl -sSL https://sds-ebpf.io/install.sh | sudo bash
```

**Kubernetes**:
```bash
kubectl apply -f https://sds-ebpf.io/daemonset.yaml
```

**Docker**:
```yaml
# docker-compose.yml
version: '3'
services:
  sds-agent:
    image: sds-ebpf/agent:latest
    privileged: true
    pid: host
    volumes:
      - /sys/kernel/debug:/sys/kernel/debug:ro
```

#### Step 2: Agent Starts Observing

**What happens**:
1. Agent loads eBPF programs into kernel
2. eBPF programs attach to kernel tracepoints
3. Every process exec, network connection, syscall → event
4. Events buffered in 256KB ring buffer
5. User-space Go collector reads events
6. Events serialized to JSON
7. JSON sent to OTel Collector

**Example Event** (Process Execution):
```json
{
  "type": "process.exec",
  "timestamp": "2024-01-01T12:00:05.123456789Z",
  "pid": 12345,
  "ppid": 1000,
  "uid": 1000,
  "comm": "node",
  "filename": "/usr/bin/node",
  "args": ["node", "server.js"],
  "cgroup_id": 1234567890,
  "container_id": "abc123",
  "pod_name": "payment-service-7d4f8",
  "namespace": "production"
}
```

**Example Event** (Network Connection):
```json
{
  "type": "network.tcp_connect",
  "timestamp": "2024-01-01T12:00:08.777777777Z",
  "pid": 12345,
  "comm": "node",
  "src_addr": "10.0.1.15",
  "dst_addr": "10.0.2.30",
  "dst_port": 5432,
  "container_id": "abc123",
  "pod_name": "payment-service-7d4f8"
}
```

#### Step 3: Data Aggregation (Phase 2)

**Pipeline**:
```
Agent (JSON)
  → OTel Collector (normalize)
  → Kafka (stream)
  → ClickHouse (store events)
  → ETL Service (enrich & transform)
  → Neo4j (build graph)
```

**ETL Enrichment**:
```python
# ETL Service pseudo-code
event = read_from_kafka("sds-events")

# Enrich with container metadata
if event.container_id:
    container = docker_client.containers.get(event.container_id)
    event.container_name = container.name
    event.image = container.image.tags[0]
    event.labels = container.labels

# Enrich with Kubernetes metadata
if event.pod_name:
    pod = k8s_client.read_namespaced_pod(event.pod_name, event.namespace)
    event.service_account = pod.spec.service_account
    event.node_name = pod.spec.node_name
    event.owner = pod.metadata.owner_references[0].name

# Build graph relationships
if event.type == "network.tcp_connect":
    src_service = resolve_service(event.src_addr, event.src_pod)
    dst_service = resolve_service(event.dst_addr, event.dst_port)

    # Create Neo4j relationship
    neo4j.run("""
        MATCH (src:Service {name: $src_service})
        MATCH (dst:Service {name: $dst_service})
        MERGE (src)-[:DEPENDS_ON]->(dst)
        SET relationship.last_seen = $timestamp,
            relationship.count = relationship.count + 1
    """, src_service=src_service, dst_service=dst_service, timestamp=event.timestamp)
```

#### Step 4: Graph Visualization

**Neo4j Query** (Show architecture):
```cypher
// Find all services and their dependencies
MATCH (s:Service)-[d:DEPENDS_ON]->(t:Service)
WHERE s.environment = 'production'
RETURN s, d, t
```

**Result** (Auto-generated diagram):
```
┌──────────┐     ┌─────────────┐     ┌──────────┐
│  web-ui  │────▶│order-service│────▶│payments  │
└──────────┘     └─────────────┘     │-service  │
                        │             └──────────┘
                        │                  │
                        ▼                  ▼
                 ┌─────────────┐     ┌──────────┐
                 │auth-service │     │payments  │
                 └─────────────┘     │-database │
                        │             └──────────┘
                        ▼
                 ┌─────────────┐
                 │ user-db     │
                 └─────────────┘
```

#### Step 5: Query and Analyze

**ClickHouse Query** (Analytical):
```sql
-- Top 10 most active services by network connections
SELECT
    src_service,
    COUNT(*) as connection_count,
    COUNT(DISTINCT dst_service) as unique_destinations
FROM network_events
WHERE timestamp > NOW() - INTERVAL 1 HOUR
GROUP BY src_service
ORDER BY connection_count DESC
LIMIT 10;
```

**Neo4j Query** (Graph Traversal):
```cypher
// Critical path analysis: What breaks if auth-service goes down?
MATCH path = (s:Service)-[:DEPENDS_ON*]->(auth:Service {name: 'auth-service'})
RETURN s.name AS affected_service,
       length(path) AS dependency_depth
ORDER BY dependency_depth
```

**Result**:
```
affected_service    | dependency_depth
--------------------|------------------
user-profile        | 1  (direct dependency)
order-service       | 2  (depends on user-profile)
web-ui              | 3  (depends on order-service)
reporting-job       | 2  (depends on user-profile)
```

**Insight**: If auth-service fails, 4 services are impacted. web-ui has longest dependency chain (3 hops).

#### Step 6: AI-Powered Q&A (Phase 3 - Future)

**Natural Language Query**:
```
User: "Show me all services that access customer PII data"

SDS eBPF (LLM-powered):
1. Query Neo4j for nodes labeled "PII"
2. Find all services with paths to PII nodes
3. Retrieve access patterns from ClickHouse
4. Generate response:

"3 services directly access PII:
1. user-profile-service
   - Reads: users table (name, email, address)
   - Frequency: ~1000 queries/min
   - Authentication: Service account (least privilege ✓)

2. billing-service
   - Reads: users table (name, payment_method)
   - Writes: invoices table
   - Frequency: ~50 queries/min
   - ⚠️ Warning: Runs as root

3. analytics-batch-job
   - Reads: users table (ALL COLUMNS)
   - Frequency: Daily at 2 AM
   - ⚠️ Warning: Exports to S3 bucket (publicly readable!)

[Generated diagram showing data flow from users table]"
```

---

## Value Proposition

### For Security Teams

#### 1. **Attack Surface Mapping**

**Before SDS eBPF**:
- "We think we have 12 internet-facing services"
- Manual audits every 6 months
- Spreadsheets to track

**With SDS eBPF**:
```cypher
// Find all services accepting external traffic
MATCH (s:Service)
WHERE s.ingress_public = true
RETURN s.name, s.exposed_ports, s.last_external_request
```

**Real-Time Dashboard**:
- 23 internet-facing services (9 more than documented!)
- 3 accepting traffic on non-standard ports
- 1 running outdated nginx with known CVE

#### 2. **Privilege Escalation Risk**

**Query**:
```cypher
// Services running as root that shouldn't
MATCH (s:Service)-[:RUNS_AS]->(u:User {uid: 0})
WHERE s.should_be_privileged = false
RETURN s.name, s.reason_for_root
```

**Alert**:
- `legacy-cron-job` runs as root (no reason given)
- `image-processor` runs as root (needs fixing!)
- Recommendation: Use capabilities instead

#### 3. **Blast Radius Analysis**

"If service X is compromised, what can attacker reach?"

```cypher
MATCH path = (compromised:Service {name: 'web-scraper'})-[:DEPENDS_ON*]->(target)
WHERE target.has_sensitive_data = true
RETURN DISTINCT target.name, shortestPath(path)
```

**Result**:
- web-scraper → S3 bucket (contains backup of production DB!)
- Fix: Revoke S3 access, web-scraper shouldn't need it

### For Architecture Teams

#### 1. **Onboarding New Engineers**

**Current Process** (2 weeks):
- Read outdated wiki
- Ask seniors "how does X work?"
- Trial and error

**With SDS eBPF** (2 hours):
```
New Engineer: "Explain how user authentication works"

SDS eBPF:
"User authentication flow:
1. web-ui receives login request
2. Calls auth-service /login endpoint
3. auth-service queries user-db
4. Returns JWT token
5. web-ui stores in cookie

[Interactive diagram with actual code paths]

Want to see the code?
auth-service: /src/services/auth-service/login.go:42"
```

#### 2. **Impact Analysis for Changes**

**Scenario**: We want to upgrade PostgreSQL

**Query**:
```cypher
MATCH (s:Service)-[:CONNECTS_TO]->(db:Database {type: 'PostgreSQL'})
RETURN s.name, s.connection_pool_size, s.query_patterns
```

**Analysis**:
- 7 services connect to PostgreSQL
- 2 use deprecated features (need code changes)
- 1 has connection leak (800 idle connections!)
- Estimated downtime: 15 minutes
- Recommended order: [generated migration plan]

#### 3. **Cost Optimization**

**Unused Services**:
```sql
SELECT service_name, last_request_timestamp
FROM service_activity
WHERE last_request_timestamp < NOW() - INTERVAL 30 DAYS
```

**Result**:
- `experimental-feature-v1`: No traffic in 90 days (can decommission)
- `old-reporting-api`: Replaced by new-api, but still running
- Potential savings: $2,400/month in EC2 costs

### For DevOps/SRE Teams

#### 1. **Faster Incident Response**

**Before**:
```
Alert: 500 errors on /api/checkout
↓ (5 minutes checking logs)
Found: RDS connection timeout
↓ (10 minutes checking RDS)
Found: RDS is fine
↓ (15 minutes checking network)
Found: Security group changed
Root cause: 30 minutes
```

**With SDS eBPF**:
```
Alert: 500 errors on /api/checkout
↓ (Query SDS)
dependency_changed(service='checkout', timestamp='10:15 AM')
  → auth-service deployed
  → auth-service now requires new header
  → checkout-service not updated
Root cause: 30 seconds
```

#### 2. **Capacity Planning**

**Query**: "Which services will hit resource limits first?"

```sql
SELECT service_name,
       current_cpu_usage,
       growth_rate_30d,
       estimated_days_to_limit
FROM service_capacity_forecast
WHERE estimated_days_to_limit < 30
ORDER BY estimated_days_to_limit
```

**Proactive Scaling**:
- `search-service` will hit CPU limit in 12 days
- `recommendation-engine` needs more memory in 20 days
- Auto-scaling configured: ✅

---

## Use Cases

### 1. **Regulatory Compliance (SOC 2, GDPR, HIPAA)**

**Requirement**: Document all systems processing customer data

**Solution**:
```cypher
MATCH (s:Service)-[:PROCESSES]->(d:Data {type: 'PII'})
RETURN s.name, s.data_classification, s.encryption_at_rest
```

**Auto-Generated Compliance Report**:
- List of all services accessing PII: ✅
- Encryption status: ✅
- Access logs: ✅
- Data retention policies: ✅

**Time Saved**: 80 hours/year of manual auditing

### 2. **M&A Due Diligence**

**Scenario**: Acquiring a company, need to understand their tech stack

**Before**: 6 weeks of interviews and document review

**With SDS eBPF**:
1. Deploy agents on acquisition target's infrastructure
2. Run for 1 week
3. Generate complete architecture report

**Report Includes**:
- Full service inventory
- Technology stack (languages, frameworks, databases)
- Dependency graph
- Security vulnerabilities
- Technical debt assessment
- Integration complexity analysis

**Time Saved**: 5 weeks

### 3. **Microservices Migration**

**Goal**: Break monolith into microservices

**Challenge**: Don't know internal dependencies

**Solution**:
1. Deploy SDS agent on monolith server
2. Capture all function calls and database queries
3. Build internal dependency graph
4. Identify bounded contexts

**Example Output**:
```
Monolith Internal Structure:
┌──────────────────────────────────┐
│  Payment Module                   │
│  ├─ Depends on: User Module      │
│  ├─ DB: payments table            │
│  └─ External: Stripe API          │
├──────────────────────────────────┤
│  User Module                      │
│  ├─ Depends on: (nothing)        │
│  ├─ DB: users, sessions           │
│  └─ External: (none)              │
├──────────────────────────────────┤
│  Order Module                     │
│  ├─ Depends on: User, Payment    │
│  ├─ DB: orders, order_items       │
│  └─ External: ShipStation API     │
└──────────────────────────────────┘

Recommended Migration Order:
1. User Module (no dependencies)
2. Payment Module (depends only on User)
3. Order Module (depends on User + Payment)
```

### 4. **Zero Trust Architecture Implementation**

**Goal**: Implement least-privilege access

**Current State**: Everything can talk to everything

**SDS eBPF Analysis**:
```
Actual communication patterns (last 30 days):
- web-ui → auth-service ✓ (expected)
- web-ui → payment-service ✗ (should go through API gateway)
- analytics-job → production-db ✗ (should use read-replica)
- random-script.py → S3 ✗ (unauthorized access)
```

**Generated Network Policies**:
```yaml
# Auto-generated Kubernetes NetworkPolicy
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: web-ui-policy
spec:
  podSelector:
    matchLabels:
      app: web-ui
  policyTypes:
  - Egress
  egress:
  - to:
    - podSelector:
        matchLabels:
          app: auth-service
    ports:
    - protocol: TCP
      port: 8080
  # Only allows proven-necessary connections
```

---

## Technical Deep Dive

### eBPF Programs in Detail

#### 1. **Process Execution Tracer**

**Kernel Hook**: `tracepoint/sched/sched_process_exec`

**C Code** (simplified):
```c
SEC("tp/sched/sched_process_exec")
int handle_exec(struct trace_event_raw_sched_process_exec *ctx) {
    struct process_event event = {};
    struct task_struct *task = (struct task_struct *)bpf_get_current_task();

    event.pid = bpf_get_current_pid_tgid() >> 32;
    event.ppid = BPF_CORE_READ(task, real_parent, tgid);
    event.uid = bpf_get_current_uid_gid() & 0xFFFFFFFF;
    bpf_get_current_comm(&event.comm, sizeof(event.comm));
    bpf_probe_read_str(&event.filename, sizeof(event.filename), ctx->filename);
    event.cgroup_id = get_cgroup_id(task);

    bpf_ringbuf_submit(&events, &event, sizeof(event), 0);
    return 0;
}
```

**What It Captures**:
- PID, PPID (process family tree)
- UID (who ran it)
- Command name
- Executable path
- Container ID (via cgroup)
- Arguments (with privacy filtering)

**Why It Matters**:
- Understand what's actually running (not just what's deployed)
- Detect suspicious process spawns
- Build process execution graph

#### 2. **Network Connection Tracer**

**Kernel Hook**: `kprobe/tcp_connect`

**C Code**:
```c
SEC("kprobe/tcp_connect")
int trace_tcp_connect(struct pt_regs *ctx) {
    struct sock *sk = (struct sock *)PT_REGS_PARM1(ctx);
    struct tcp_event event = {};

    event.pid = bpf_get_current_pid_tgid() >> 32;
    bpf_get_current_comm(&event.comm, sizeof(event.comm));

    struct sock_common *sk_common = &sk->__sk_common;
    event.saddr = BPF_CORE_READ(sk_common, skc_rcv_saddr);
    event.daddr = BPF_CORE_READ(sk_common, skc_daddr);
    event.sport = BPF_CORE_READ(sk_common, skc_num);
    event.dport = bpf_ntohs(BPF_CORE_READ(sk_common, skc_dport));

    bpf_ringbuf_submit(&events, &event, sizeof(event), 0);
    return 0;
}
```

**What It Captures**:
- Source/destination IP:port
- Process making connection
- Timestamp

**Why It Matters**:
- Builds service dependency graph automatically
- Detects anomalous connections
- Maps network topology

#### 3. **Sparse Sampling Logic**

**Go Collector**:
```go
func (c *Collector) Run() {
    ticker := time.NewTicker(5 * time.Minute)

    for {
        // SAMPLING ON: Collect for 60 seconds
        c.enablePrograms()
        time.Sleep(60 * time.Second)
        c.disablePrograms()

        // SAMPLING OFF: Sleep for 4 minutes
        <-ticker.C
    }
}
```

**Benefits**:
- CPU usage: <1% average
- Captures enough data for dependency mapping
- Scalable to 10,000+ hosts

---

## Comparison with Existing Tools

### vs. Traditional APM (Datadog, New Relic, AppDynamics)

| Feature | Traditional APM | SDS eBPF |
|---------|----------------|----------|
| **Instrumentation** | Code changes required | Zero code changes |
| **Language Support** | Java, Python, Node, etc. | **Any language** (compiled, scripting, legacy) |
| **Binary Support** | ❌ No | ✅ Yes (observes syscalls) |
| **Overhead** | 5-20% | <1% (sparse sampling) |
| **Coverage** | Services you instrument | **Everything running** |
| **System Calls** | ❌ No | ✅ Yes |
| **Container Aware** | Limited | ✅ Full (cgroup tracking) |
| **Cost** | $15-100 per host/month | **Open-source** |

### vs. Service Mesh (Istio, Linkerd)

| Feature | Service Mesh | SDS eBPF |
|---------|--------------|----------|
| **Deployment** | Sidecar per pod | DaemonSet per node |
| **Memory Usage** | 100-200 MB per pod | 50 MB per node |
| **Non-K8s Support** | ❌ No | ✅ Yes |
| **Traffic Routing** | ✅ Yes | ❌ No (observation only) |
| **Observability** | HTTP/gRPC only | **All protocols** |
| **System Calls** | ❌ No | ✅ Yes |
| **Purpose** | Traffic management | **Discovery & Understanding** |

### vs. Static Analysis (AWS Config, Azure Resource Graph)

| Feature | Static Analysis | SDS eBPF |
|---------|----------------|----------|
| **Data Source** | Cloud API | **Runtime behavior** |
| **Accuracy** | What's configured | **What's actually running** |
| **Dependencies** | Inferred from config | **Observed from traffic** |
| **Updates** | Lag time | **Real-time** |
| **Cloud Agnostic** | Per-cloud tools | ✅ Yes |
| **On-Premise** | ❌ No | ✅ Yes |

### vs. Network Traffic Analysis (Zeek, Suricata)

| Feature | Network Traffic Analysis | SDS eBPF |
|---------|------------------------|----------|
| **Scope** | Network only | **Process + Network + Files** |
| **Process Context** | ❌ No | ✅ Yes (knows which process) |
| **Encrypted Traffic** | ❌ Can't see | ✅ Sees pre-encryption |
| **System Calls** | ❌ No | ✅ Yes |
| **Deployment** | SPAN port/tap | **Host-based agent** |
| **Overhead** | High (analyzes all packets) | Low (sparse sampling) |

---

## Roadmap

### ✅ Phase 1: Baseline eBPF Agent (COMPLETE)

**Delivered**:
- Sparse-sampling eBPF agent (C + Go)
- Process, syscall, network, container monitoring
- JSON event output
- Docker/Kubernetes deployment support
- Documentation and diagrams

### 🔜 Phase 2: Telemetry Aggregation (Q1 2024)

**Planned**:
- OpenTelemetry Collector integration
- Kafka/Redpanda for event streaming
- ClickHouse for analytical queries
- Neo4j for graph relationships
- VictoriaMetrics for time-series
- ETL pipeline for enrichment

**Outcome**: Unified data platform with multi-storage architecture

### 🔮 Phase 3: Graph ML & LLM Reasoning (Q2 2024)

**Planned**:
- GraphRAG indexing
- LLM training on architecture data
- Natural language query interface
- Automatic service discovery
- Relationship inference

**Outcome**: "ChatGPT for your infrastructure"

### 🔮 Phase 4: Interactive Visualization (Q3 2024)

**Planned**:
- Real-time architecture diagrams
- Interactive exploration (click to drill down)
- Excalidraw/D2 export
- Manim animations for data flow
- Grafana integration

**Outcome**: Living documentation that updates itself

### 🔮 Phase 5: Advanced Use Cases (Q4 2024)

**Planned**:
- Chaos engineering integration (predict blast radius)
- Automated security posture assessment
- Cost optimization recommendations
- Compliance report generation
- A/B test impact analysis

---

## Getting Started

### Prerequisites

- Linux kernel 4.18+ (for eBPF support)
- Root access (for eBPF)
- Docker or Kubernetes (for deployment)

### Quick Start (5 minutes)

**1. Install on EC2/VM**:
```bash
curl -sSL https://sds-ebpf.io/install.sh | sudo bash
```

**2. View collected data**:
```bash
tail -f /var/log/sds-observer/events_*.json | jq .
```

**3. Query events**:
```bash
# Find all network connections
cat /var/log/sds-observer/events_*.json | \
  jq 'select(.type == "network.tcp_connect")'

# Count processes by name
cat /var/log/sds-observer/events_*.json | \
  jq -r 'select(.type == "process.exec") | .comm' | \
  sort | uniq -c | sort -nr
```

### Full Platform (Phase 2)

**Coming Soon**: docker-compose setup for complete stack
```bash
git clone https://github.com/sds-ebpf/platform
cd platform
docker-compose up
```

**Access**:
- Grafana: http://localhost:3000
- Neo4j Browser: http://localhost:7474
- Query API: http://localhost:8080

---

## FAQ

**Q: Does this work with serverless (Lambda, Cloud Functions)?**
A: Not yet. eBPF requires kernel access. Investigating eBPF in Firecracker (Lambda's underlying tech).

**Q: What about Windows?**
A: Windows doesn't have eBPF. Researching alternatives (ETW, eBPF for Windows project).

**Q: Performance impact?**
A: <1% CPU average with sparse sampling (60s/5min). Configurable to your needs.

**Q: Does it capture sensitive data?**
A: No. Filters out credentials, tokens, PII. Only captures metadata (who, what, when, where).

**Q: Can I run this in production?**
A: Phase 1 (agent) is production-ready with sparse sampling. Phase 2+ in development.

**Q: What if my team uses [specific cloud/tool]?**
A: SDS eBPF is cloud-agnostic. Works with any infrastructure. Integrates with existing tools via OpenTelemetry.

---

## Contributing

This is an open-source project. Contributions welcome!

**Areas for Contribution**:
- eBPF program enhancements
- Additional protocol support (HTTP/2, gRPC, etc.)
- UI/visualization
- Documentation
- Testing

**Get Involved**:
- GitHub: https://github.com/sds-ebpf
- Discord: https://discord.gg/sds-ebpf
- Twitter: @sds_ebpf

---

## License

Apache 2.0 License (to be added)

---

## Acknowledgments

Built on the shoulders of giants:
- Linux kernel team (eBPF)
- Cilium team (cilium/ebpf library)
- OpenTelemetry community
- eBPF community

---

**Questions?** Open an issue or join our Discord!

**Ready to understand your infrastructure?** [Get Started →](QUICKSTART.md)