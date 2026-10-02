#!/usr/bin/env python3

import argparse
import hashlib
import json
from datetime import datetime, timezone
from pathlib import Path


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Write a manifest for a bundled SDS workload demo artifact."
    )
    parser.add_argument("--workload", required=True, help="Workload name")
    parser.add_argument("--namespace", required=True, help="Workload namespace or project name")
    parser.add_argument("--events", required=True, help="Path to canonical NDJSON capture")
    parser.add_argument("--graph", required=True, help="Path to runtime graph snapshot")
    parser.add_argument("--eval-summary", required=True, help="Path to graph evaluation summary")
    parser.add_argument("--validation", required=True, help="Path to artifact validation summary")
    parser.add_argument("--acceptance", required=True, help="Path to copied acceptance contract")
    parser.add_argument("--output", required=True, help="Path to output artifact manifest")
    return parser.parse_args()


def file_metadata(path: Path) -> dict[str, object]:
    digest = hashlib.sha256()
    with path.open("rb") as handle:
        for chunk in iter(lambda: handle.read(1024 * 1024), b""):
            digest.update(chunk)

    return {
        "path": str(path.resolve()),
        "size_bytes": path.stat().st_size,
        "sha256": digest.hexdigest(),
    }


def main() -> int:
    args = parse_args()
    events_path = Path(args.events)
    graph_path = Path(args.graph)
    eval_summary_path = Path(args.eval_summary)
    validation_path = Path(args.validation)
    acceptance_path = Path(args.acceptance)
    output_path = Path(args.output)

    for path in [
        events_path,
        graph_path,
        eval_summary_path,
        validation_path,
        acceptance_path,
    ]:
        if not path.exists():
            raise FileNotFoundError(f"artifact path does not exist: {path}")

    validation = json.loads(validation_path.read_text(encoding="utf-8"))
    eval_summary = json.loads(eval_summary_path.read_text(encoding="utf-8"))
    run_id = events_path.stem

    manifest = {
        "schema_version": "sds.demo_artifact.v1",
        "generated_at": datetime.now(timezone.utc).isoformat().replace("+00:00", "Z"),
        "run_id": run_id,
        "workload": args.workload,
        "namespace": args.namespace,
        "validation_passed": validation.get("passed"),
        "artifacts": {
            "events": file_metadata(events_path),
            "graph": file_metadata(graph_path),
            "eval_summary": file_metadata(eval_summary_path),
            "validation": file_metadata(validation_path),
            "acceptance": file_metadata(acceptance_path),
        },
        "metrics": {
            "precision": eval_summary.get("precision"),
            "recall": eval_summary.get("recall"),
            "f1": eval_summary.get("f1"),
            "namespaced_event_count": validation.get("event_counts", {}).get("namespaced"),
            "namespaced_network_event_count": validation.get("event_counts", {})
            .get("by_kind", {})
            .get("network.tcp_connect", 0),
            "service_node_count": validation.get("graph_counts", {}).get("service_node_count"),
        },
    }

    output_path.parent.mkdir(parents=True, exist_ok=True)
    output_path.write_text(json.dumps(manifest, indent=2, sort_keys=True) + "\n", encoding="utf-8")
    print(
        "wrote demo artifact manifest "
        f"run_id={run_id} validation_passed={manifest['validation_passed']}"
    )
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
