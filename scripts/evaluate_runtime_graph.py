#!/usr/bin/env python3

import argparse
import json
from pathlib import Path


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Evaluate a runtime graph against ground truth edges.")
    parser.add_argument("--graph", required=True, help="Path to graph snapshot JSON")
    parser.add_argument("--ground-truth", required=True, help="Path to expected edge JSON")
    parser.add_argument("--output", required=True, help="Path to output evaluation summary JSON")
    return parser.parse_args()


def compute_f1(precision: float, recall: float) -> float:
    if precision == 0.0 and recall == 0.0:
        return 0.0
    return 2.0 * precision * recall / (precision + recall)


def main() -> int:
    args = parse_args()
    graph_path = Path(args.graph)
    truth_path = Path(args.ground_truth)
    output_path = Path(args.output)

    graph = json.loads(graph_path.read_text(encoding="utf-8"))
    truth = json.loads(truth_path.read_text(encoding="utf-8"))

    predicted_edges = {
        (edge["kind"], edge["src"], edge["dst"])
        for edge in graph.get("edges", [])
        if edge["kind"] in {"DEPENDS_ON", "CONNECTS_TO_EXTERNAL"}
    }
    expected_edges = {
        (edge["kind"], edge["src"], edge["dst"])
        for edge in truth.get("expected_edges", [])
    }

    true_positives = sorted(predicted_edges & expected_edges)
    false_positives = sorted(predicted_edges - expected_edges)
    false_negatives = sorted(expected_edges - predicted_edges)

    precision = len(true_positives) / len(predicted_edges) if predicted_edges else 0.0
    recall = len(true_positives) / len(expected_edges) if expected_edges else 0.0
    f1 = compute_f1(precision, recall)

    summary = {
        "schema_version": "sds.eval.v1",
        "graph_path": str(graph_path),
        "ground_truth_path": str(truth_path),
        "predicted_edge_count": len(predicted_edges),
        "expected_edge_count": len(expected_edges),
        "true_positive_count": len(true_positives),
        "false_positive_count": len(false_positives),
        "false_negative_count": len(false_negatives),
        "precision": precision,
        "recall": recall,
        "f1": f1,
        "true_positives": true_positives,
        "false_positives": false_positives,
        "false_negatives": false_negatives,
    }

    output_path.parent.mkdir(parents=True, exist_ok=True)
    output_path.write_text(json.dumps(summary, indent=2, sort_keys=True) + "\n", encoding="utf-8")

    print(
        "evaluation summary: "
        f"precision={precision:.3f} recall={recall:.3f} f1={f1:.3f} "
        f"tp={len(true_positives)} fp={len(false_positives)} fn={len(false_negatives)}"
    )
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
