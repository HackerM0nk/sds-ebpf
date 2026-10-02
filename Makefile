PYTHON ?= python3
DOCKER_COMPOSE ?= docker-compose
POLAR_PROJECT ?= sds-polar-demo
POLAR_COMPOSE_FILE ?= eval/workloads/polar-bookshop/docker-compose.yml

.PHONY: agent-go-test agent-go-build validate-sample build-sample-graph eval-sample research-loop clean-eval polar-compose-config polar-up polar-down polar-scenario polar-validate polar-demo

agent-go-test:
	cd agent && go test ./...

agent-go-build:
	cd agent && mkdir -p bin && go build -o bin/observer ./cmd/observer

validate-sample:
	$(PYTHON) scripts/validate_canonical_events.py --input eval/sample/canonical_events.ndjson

build-sample-graph:
	mkdir -p eval/output
	$(PYTHON) scripts/build_runtime_graph.py --input eval/sample/canonical_events.ndjson --output eval/output/runtime_graph.json

eval-sample:
	bash scripts/run_eval_loop.sh

research-loop: agent-go-test agent-go-build eval-sample

clean-eval:
	rm -rf eval/output

polar-compose-config:
	$(DOCKER_COMPOSE) -f $(POLAR_COMPOSE_FILE) -p $(POLAR_PROJECT) config >/dev/null

polar-up:
	$(DOCKER_COMPOSE) -f $(POLAR_COMPOSE_FILE) -p $(POLAR_PROJECT) up -d

polar-down:
	$(DOCKER_COMPOSE) -f $(POLAR_COMPOSE_FILE) -p $(POLAR_PROJECT) down -v --remove-orphans

polar-scenario:
	bash eval/workloads/polar-bookshop/scenarios/smoke.sh

polar-validate:
	bash eval/workloads/polar-bookshop/validate_latest.sh

polar-demo:
	bash eval/workloads/polar-bookshop/run_demo.sh
