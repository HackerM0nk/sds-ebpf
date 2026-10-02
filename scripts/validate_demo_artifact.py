#!/usr/bin/env python3

import argparse
import json
from collections import Counter
from datetime import datetime, timezone
from pathlib import Path
from typing import Any


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Validate a workload-backed SDS demo artifact against acceptance gates."
    )
    parser.add_argument("--events", required=True, help="Path to canonical NDJSON events")
    parser.add_argument("--graph", required=True, help="Path to graph snapshot JSON")
    parser.add_argument("--acceptance", required=True, help="Path to workload acceptance JSON")
    parser.add_argument("--output", required=True, help="Path to output validation summary JSON")
    parser.add_argument("--eval-summary", help="Optional path to evaluation summary JSON")
    return parser.parse_args()


def load_ndjson(path: Path) -> list[dict[str, Any]]:
    records: list[dict[str, Any]] = []
    with path.open("r", encoding="utf-8") as handle:
        for raw in handle:
            raw = raw.strip()
            if not raw:
                continue
            records.append(json.loads(raw))
    return records


def make_check(name: str, passed: bool, actual: Any, expected: Any) -> dict[str, Any]:
    return {
        "name": name,
        "passed": passed,
        "actual": actual,
        "expected": expected,
    }


def main() -> int:
    args = parse_args()
    events_path = Path(args.events)
    graph_path = Path(args.graph)
    acceptance_path = Path(args.acceptance)
    output_path = Path(args.output)

    events = load_ndjson(events_path)
    graph = json.loads(graph_path.read_text(encoding="utf-8"))
    acceptance = json.loads(acceptance_path.read_text(encoding="utf-8"))
    eval_summary = None
    if args.eval_summary:
        eval_summary = json.loads(Path(args.eval_summary).read_text(encoding="utf-8"))

    namespace = acceptance["namespace"]
    namespaced_events = [
        event for event in events if event.get("workload", {}).get("namespace") == namespace
    ]

    namespaced_services = Counter(
        event.get("workload", {}).get("service", "unknown") for event in namespaced_events
    )
    namespaced_kinds = Counter(
        event.get("event", {}).get("kind", "unknown") for event in namespaced_events
    )

    graph_service_ids = {
        node["id"]
        for node in graph.get("nodes", [])
        if node.get("type") == "Service"
    }
    graph_edges = {
        (edge["kind"], edge["src"], edge["dst"])
        for edge in graph.get("edges", [])
    }

    checks: list[dict[str, Any]] = []

    minimums = acceptance.get("minimums", {})
    min_namespaced_event_count = int(minimums.get("namespaced_event_count", 0))
    checks.append(
        make_check(
            "minimum_namespaced_event_count",
            len(namespaced_events) >= min_namespaced_event_count,
            len(namespaced_events),
            {"minimum": min_namespaced_event_count},
        )
    )

    min_namespaced_network_event_count = int(minimums.get("namespaced_network_event_count", 0))
    actual_network_event_count = namespaced_kinds.get("network.tcp_connect", 0)
    checks.append(
        make_check(
            "minimum_namespaced_network_event_count",
            actual_network_event_count >= min_namespaced_network_event_count,
            actual_network_event_count,
            {"minimum": min_namespaced_network_event_count},
        )
    )

    for service in acceptance.get("required_event_services", []):
        actual = namespaced_services.get(service, 0)
        checks.append(
            make_check(
                f"required_event_service:{service}",
                actual > 0,
                actual,
                {"minimum": 1},
            )
        )

    for service in acceptance.get("required_graph_services", []):
        service_id = f"service:{service}"
        checks.append(
            make_check(
                f"required_graph_service:{service}",
                service_id in graph_service_ids,
                service_id in graph_service_ids,
                True,
            )
        )

    for edge in acceptance.get("required_edges", []):
        edge_key = (edge["kind"], edge["src"], edge["dst"])
        checks.append(
            make_check(
                f"required_edge:{edge['src']}->{edge['dst']}",
                edge_key in graph_edges,
                edge_key in graph_edges,
                True,
            )
        )

    validation = {
        "schema_version": "sds.artifact_validation.v1",
        "generated_at": datetime.now(timezone.utc).isoformat().replace("+00:00", "Z"),
        "workload": acceptance.get("workload"),
        "namespace": namespace,
        "events_path": str(events_path),
        "graph_path": str(graph_path),
        "acceptance_path": str(acceptance_path),
        "eval_summary_path": args.eval_summary,
        "event_counts": {
            "total": len(events),
            "namespaced": len(namespaced_events),
            "by_kind": dict(namespaced_kinds),
            "by_service": dict(namespaced_services),
        },
        "graph_counts": {
            "node_count": len(graph.get("nodes", [])),
            "edge_count": len(graph.get("edges", [])),
            "service_node_count": len(graph_service_ids),
        },
        "checks": checks,
        "passed": all(check["passed"] for check in checks),
    }

    if eval_summary is not None:
        validation["eval_summary"] = {
            "precision": eval_summary.get("precision"),
            "recall": eval_summary.get("recall"),
            "f1": eval_summary.get("f1"),
            "true_positive_count": eval_summary.get("true_positive_count"),
            "false_positive_count": eval_summary.get("false_positive_count"),
            "false_negative_count": eval_summary.get("false_negative_count"),
        }

    output_path.parent.mkdir(parents=True, exist_ok=True)
    output_path.write_text(json.dumps(validation, indent=2, sort_keys=True) + "\n", encoding="utf-8")

    print(
        "artifact validation: "
        f"passed={validation['passed']} "
        f"namespaced_events={len(namespaced_events)} "
        f"namespaced_network_events={actual_network_event_count} "
        f"service_nodes={len(graph_service_ids)}"
    )

    return 0 if validation["passed"] else 1


if __name__ == "__main__":
    raise SystemExit(main())
