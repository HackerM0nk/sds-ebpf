# Makefile for SDS eBPF Observer Agent

# Variables
CLANG ?= clang
LLVM_STRIP ?= llvm-strip
BPFTOOL ?= bpftool
GO ?= go

ARCH := $(shell uname -m | sed 's/x86_64/x86/' | sed 's/aarch64/arm64/')
VMLINUX := ebpf/vmlinux.h
BPF_OBJ := ebpf/observer.bpf.o
BPF_SRC := ebpf/observer.bpf.c
GO_SRC := $(shell find . -type f -name '*.go')

OUTPUT_DIR := bin
BINARY := $(OUTPUT_DIR)/sds-observer

# Compiler flags
INCLUDES := -I./ebpf -I/usr/include
CFLAGS := -g -O2 -target bpf -D__TARGET_ARCH_$(ARCH) $(INCLUDES)

.PHONY: all clean build-bpf build-go build docker-build test help

all: build ## Build everything

help: ## Display this help message
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

vmlinux: ## Generate vmlinux.h from BTF
	@echo "Generating vmlinux.h..."
	@if [ -f /sys/kernel/btf/vmlinux ]; then \
		$(BPFTOOL) btf dump file /sys/kernel/btf/vmlinux format c > $(VMLINUX).generated; \
		echo "Generated $(VMLINUX).generated from running kernel"; \
	else \
		echo "Kernel BTF not available. Using minimal vmlinux.h stub."; \
		echo "For full functionality, run on a Linux system with BTF support."; \
	fi

build-bpf: ## Build eBPF programs
	@echo "Building eBPF programs..."
	@mkdir -p $(OUTPUT_DIR)
	$(CLANG) $(CFLAGS) -c $(BPF_SRC) -o $(BPF_OBJ)
	$(LLVM_STRIP) -g $(BPF_OBJ)
	@echo "eBPF object: $(BPF_OBJ)"

build-go: ## Build Go binary (with embedded eBPF)
	@echo "Building Go binary..."
	@mkdir -p $(OUTPUT_DIR)
	cd cmd/observer && $(GO) build -o ../../$(BINARY) .
	@echo "Binary: $(BINARY)"

build: build-bpf build-go ## Build both eBPF and Go components

generate: ## Generate Go bindings from eBPF (requires bpf2go)
	@echo "Generating Go bindings with bpf2go..."
	$(GO) generate ./...

run: build ## Build and run the observer (requires Linux + root)
	@echo "Running observer (requires root privileges)..."
	sudo $(BINARY) --output=/tmp/sds-observer --verbose

docker-build: ## Build Docker image for local testing
	@echo "Building Docker image..."
	docker build -t sds-observer:latest -f ../docker/Dockerfile.agent .

docker-run: docker-build ## Run observer in Docker container
	@echo "Running observer in Docker..."
	docker run --rm --privileged \
		--pid=host \
		-v /sys/kernel/debug:/sys/kernel/debug:ro \
		-v /tmp/sds-observer:/var/log/sds-observer \
		sds-observer:latest

test: ## Run Go tests
	$(GO) test -v ./...

clean: ## Clean build artifacts
	rm -rf $(OUTPUT_DIR)
	rm -f $(BPF_OBJ)
	rm -f $(VMLINUX).generated

lint: ## Run linters
	$(GO) vet ./...
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed, skipping..."; \
	fi

fmt: ## Format code
	$(GO) fmt ./...
	@if command -v gofmt > /dev/null; then \
		gofmt -s -w .; \
	fi

deps: ## Install dependencies
	$(GO) mod download
	$(GO) mod tidy

.DEFAULT_GOAL := help