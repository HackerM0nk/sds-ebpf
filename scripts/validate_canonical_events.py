#!/usr/bin/env python3

import argparse
import json
from datetime import datetime
from pathlib import Path
from typing import Any


EXPECTED_SCHEMA_VERSION = "sds.event.v1"


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Validate SDS canonical events.")
    parser.add_argument("--input", required=True, help="Path to NDJSON event file")
    return parser.parse_args()


def parse_timestamp(value: str) -> None:
    try:
        datetime.fromisoformat(value.replace("Z", "+00:00"))
    except ValueError as exc:
        raise ValueError(f"invalid observed_at timestamp: {value}") from exc


def require_str(obj: dict[str, Any], key: str) -> str:
    value = obj.get(key)
    if not isinstance(value, str) or not value:
        raise ValueError(f"field '{key}' must be a non-empty string")
    return value


def require_int(obj: dict[str, Any], key: str) -> int:
    value = obj.get(key)
    if not isinstance(value, int) or value < 0:
        raise ValueError(f"field '{key}' must be a non-negative integer")
    return value


def require_obj(obj: dict[str, Any], key: str) -> dict[str, Any]:
    value = obj.get(key)
    if not isinstance(value, dict):
        raise ValueError(f"field '{key}' must be an object")
    return value


def validate_event(event: dict[str, Any], seen_ids: set[str]) -> None:
    if event.get("schema_version") != EXPECTED_SCHEMA_VERSION:
        raise ValueError("schema_version must be 'sds.event.v1'")

    event_id = require_str(event, "event_id")
    if event_id in seen_ids:
        raise ValueError(f"duplicate event_id: {event_id}")
    seen_ids.add(event_id)

    observed_at = require_str(event, "observed_at")
    parse_timestamp(observed_at)

    host = require_obj(event, "host")
    require_str(host, "id")
    require_str(host, "hostname")

    process = require_obj(event, "process")
    require_int(process, "pid")
    require_int(process, "ppid")
    require_str(process, "comm")
    require_str(process, "binary")
    require_int(process, "uid")

    workload = require_obj(event, "workload")
    require_str(workload, "service")

    sensor = require_obj(event, "sensor")
    require_str(sensor, "name")
    require_str(sensor, "sampling_mode")

    event_obj = require_obj(event, "event")
    require_str(event_obj, "kind")
    require_str(event_obj, "category")
    data = event_obj.get("data")
    if not isinstance(data, dict):
        raise ValueError("field 'event.data' must be an object")


def main() -> int:
    args = parse_args()
    input_path = Path(args.input)
    seen_ids: set[str] = set()
    count = 0

    with input_path.open("r", encoding="utf-8") as handle:
        for line_no, raw in enumerate(handle, start=1):
            raw = raw.strip()
            if not raw:
                continue
            try:
                event = json.loads(raw)
            except json.JSONDecodeError as exc:
                raise SystemExit(f"{input_path}:{line_no}: invalid JSON: {exc}") from exc
            try:
                validate_event(event, seen_ids)
            except ValueError as exc:
                raise SystemExit(f"{input_path}:{line_no}: {exc}") from exc
            count += 1

    print(f"validated {count} canonical events from {input_path}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
