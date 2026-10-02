#!/usr/bin/env python3

import json
from http.server import BaseHTTPRequestHandler, HTTPServer


CONFIGS = {
    ("catalog-service", "default"): {
        "polar.greeting": "Welcome to the catalog from the SDS demo config server",
    },
    ("catalog-service", "testdata"): {
        "polar.greeting": "Welcome to the catalog from the SDS demo config server",
    },
    ("catalog-service", "prod"): {
        "polar.greeting": "Welcome to the production catalog from the SDS demo config server",
    },
}


def build_payload(application: str, profiles: list[str], label: str | None) -> dict[str, object]:
    source: dict[str, str] = {}
    for profile in profiles:
        source.update(CONFIGS.get(("application", profile), {}))
        source.update(CONFIGS.get((application, profile), {}))

    payload: dict[str, object] = {
        "name": application,
        "profiles": profiles,
        "label": label,
        "version": None,
        "state": None,
        "propertySources": [],
    }
    if source:
        payload["propertySources"].append({"name": "sds-demo-config", "source": source})
    return payload


class Handler(BaseHTTPRequestHandler):
    def do_GET(self) -> None:
        if self.path == "/actuator/health":
            self.respond(200, {"status": "UP"})
            return

        parts = [part for part in self.path.strip("/").split("/") if part]
        if len(parts) < 2:
            self.respond(404, {"error": "not_found"})
            return

        application = parts[0]
        profiles = parts[1].split(",")
        label = parts[2] if len(parts) > 2 else None
        self.respond(200, build_payload(application, profiles, label))

    def log_message(self, format: str, *args) -> None:
        return

    def respond(self, status: int, payload: dict[str, object]) -> None:
        encoded = json.dumps(payload).encode("utf-8")
        self.send_response(status)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(encoded)))
        self.end_headers()
        self.wfile.write(encoded)


def main() -> None:
    server = HTTPServer(("0.0.0.0", 8888), Handler)
    server.serve_forever()


if __name__ == "__main__":
    main()
