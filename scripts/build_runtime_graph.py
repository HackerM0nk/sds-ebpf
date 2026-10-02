#!/usr/bin/env python3

import argparse
import ipaddress
import json
from datetime import datetime, timezone
from pathlib import Path
from typing import Any


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Build a runtime graph snapshot from canonical SDS events.")
    parser.add_argument("--input", required=True, help="Path to canonical NDJSON events")
    parser.add_argument("--output", required=True, help="Path to output graph JSON")
    parser.add_argument("--namespace", help="Optional workload namespace filter")
    return parser.parse_args()


def node_entry(node_type: str, properties: dict[str, Any]) -> dict[str, Any]:
    node_id = properties["id"]
    properties = dict(properties)
    properties.pop("id", None)
    return {"id": node_id, "type": node_type, "properties": properties}


def touch_temporal_properties(properties: dict[str, Any], observed_at: str) -> None:
    if "first_seen" not in properties:
        properties["first_seen"] = observed_at
    properties["last_seen"] = observed_at
    properties["evidence_count"] = properties.get("evidence_count", 0) + 1


def infer_scope(address: str | None) -> str:
    if not address:
        return "external"

    try:
        ip = ipaddress.ip_address(address)
    except ValueError:
        return "external"

    if ip.is_loopback:
        return "loopback"
    if ip.is_private:
        return "internal"
    return "external"


def main() -> int:
    args = parse_args()
    input_path = Path(args.input)
    output_path = Path(args.output)
    namespace_filter = args.namespace

    services: dict[str, dict[str, Any]] = {}
    endpoints: dict[str, dict[str, Any]] = {}
    processes: dict[str, dict[str, Any]] = {}
    edges: dict[tuple[str, str, str], dict[str, Any]] = {}

    with input_path.open("r", encoding="utf-8") as handle:
        for raw in handle:
            raw = raw.strip()
            if not raw:
                continue
            event = json.loads(raw)
            observed_at = event["observed_at"]
            host = event["host"]
            process = event["process"]
            workload = event["workload"]
            if namespace_filter and workload.get("namespace") != namespace_filter:
                continue
            event_obj = event["event"]
            event_data = event_obj["data"]

            service_name = workload["service"]
            service_id = f"service:{service_name}"
            service_props = services.setdefault(
                service_id,
                {
                    "id": service_id,
                    "service": service_name,
                    "namespace": workload.get("namespace"),
                },
            )
            touch_temporal_properties(service_props, observed_at)

            process_id = f"process:{host['id']}:{process['pid']}"
            process_props = processes.setdefault(
                process_id,
                {
                    "id": process_id,
                    "host_id": host["id"],
                    "pid": process["pid"],
                    "ppid": process["ppid"],
                    "comm": process["comm"],
                    "binary": process["binary"],
                    "uid": process["uid"],
                    "service": service_name,
                },
            )
            touch_temporal_properties(process_props, observed_at)

            if event_obj["kind"] == "process.exec" and process["ppid"] > 0:
                parent_id = f"process:{host['id']}:{process['ppid']}"
                parent_props = processes.setdefault(
                    parent_id,
                    {
                        "id": parent_id,
                        "host_id": host["id"],
                        "pid": process["ppid"],
                        "ppid": 0,
                        "comm": "unknown",
                        "binary": "unknown",
                        "uid": process["uid"],
                        "service": service_name,
                    },
                )
                touch_temporal_properties(parent_props, observed_at)

                edge_key = ("SPAWNED_BY", process_id, parent_id)
                edge = edges.setdefault(
                    edge_key,
                    {
                        "kind": "SPAWNED_BY",
                        "src": process_id,
                        "dst": parent_id,
                        "properties": {
                            "source_event_types": ["process.exec"],
                            "confidence": 1.0,
                        },
                    },
                )
                touch_temporal_properties(edge["properties"], observed_at)

            if event_obj["kind"] != "network.tcp_connect":
                continue

            dst_service = event_data.get("dst_service")
            dst_scope = event_data.get("dst_scope")
            inferred_scope = dst_scope or infer_scope(event_data.get("dst_addr"))

            if isinstance(dst_service, str) and dst_service:
                dst_service_id = f"service:{dst_service}"
                dst_props = services.setdefault(
                    dst_service_id,
                    {
                        "id": dst_service_id,
                        "service": dst_service,
                        "namespace": workload.get("namespace"),
                    },
                )
                touch_temporal_properties(dst_props, observed_at)

                edge_key = ("DEPENDS_ON", service_id, dst_service_id)
                edge = edges.setdefault(
                    edge_key,
                    {
                        "kind": "DEPENDS_ON",
                        "src": service_id,
                        "dst": dst_service_id,
                        "properties": {
                            "source_event_types": ["network.tcp_connect"],
                            "confidence": 0.9,
                            "scope": inferred_scope,
                        },
                    },
                )
                touch_temporal_properties(edge["properties"], observed_at)
                continue

            endpoint_name = event_data.get("dst_name") or event_data.get("dst_addr")
            endpoint_port = event_data.get("dst_port")
            endpoint_id = f"endpoint:{endpoint_name}:{endpoint_port}"
            endpoint_props = endpoints.setdefault(
                endpoint_id,
                {
                    "id": endpoint_id,
                    "address": endpoint_name,
                    "port": endpoint_port,
                    "scope": inferred_scope,
                },
            )
            touch_temporal_properties(endpoint_props, observed_at)

            edge_kind = "CONNECTS_TO_EXTERNAL" if inferred_scope == "external" else "CONNECTS_TO_INTERNAL_ENDPOINT"
            edge_key = (edge_kind, service_id, endpoint_id)
            edge = edges.setdefault(
                edge_key,
                {
                    "kind": edge_kind,
                    "src": service_id,
                    "dst": endpoint_id,
                    "properties": {
                        "source_event_types": ["network.tcp_connect"],
                        "confidence": 0.8,
                        "scope": inferred_scope,
                    },
                },
            )
            touch_temporal_properties(edge["properties"], observed_at)

    graph = {
        "schema_version": "sds.graph.v1",
        "generated_at": datetime.now(timezone.utc).isoformat().replace("+00:00", "Z"),
        "nodes": [node_entry("Service", node) for node in services.values()]
        + [node_entry("Endpoint", node) for node in endpoints.values()]
        + [node_entry("ProcessInstance", node) for node in processes.values()],
        "edges": list(edges.values()),
    }

    output_path.parent.mkdir(parents=True, exist_ok=True)
    with output_path.open("w", encoding="utf-8") as handle:
        json.dump(graph, handle, indent=2, sort_keys=True)
        handle.write("\n")

    print(f"wrote runtime graph snapshot to {output_path}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
