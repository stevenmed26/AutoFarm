# Makefile for AutoFarm
# Usage examples:
#   make proto
#   make build
#   make run-orchestrator
#   make run-node
#   make run-api
#   make docker-up
#   make docker-down

GO        ?= go
PROTOC    ?= protoc
COMPOSE   ?= docker-compose

PROTO_DIR := internal/proto
PROTO_FILES := $(wildcard $(PROTO_DIR)/*.proto)

BIN_DIR   := bin

API_BIN          := $(BIN_DIR)/api
ORCHESTRATOR_BIN := $(BIN_DIR)/orchestrator
NODE_BIN         := $(BIN_DIR)/node

.PHONY: all proto build clean \
        run-api run-orchestrator run-node \
        docker-up docker-down docker-logs

all: build

## === Protobuf generation ===

proto:
	@echo "Generating gRPC code from .proto files..."
	@$(PROTOC) \
		-I $(PROTO_DIR) \
		--go_out=. \
		--go-grpc_out=. \
		$(PROTO_FILES)
	@echo "Done."

## === Build binaries ===

build: $(API_BIN) $(ORCHESTRATOR_BIN) $(NODE_BIN)
	@echo "All services built."

$(API_BIN):
	@echo "Building API..."
	@mkdir -p $(BIN_DIR)
	@$(GO) build -o $(API_BIN) ./cmd/api

$(ORCHESTRATOR_BIN):
	@echo "Building Orchestrator..."
	@mkdir -p $(BIN_DIR)
	@$(GO) build -o $(ORCHESTRATOR_BIN) ./cmd/orchestrator

$(NODE_BIN):
	@echo "Building Node Worker..."
	@mkdir -p $(BIN_DIR)
	@$(GO) build -o $(NODE_BIN) ./cmd/node

## === Local runs (without Docker) ===

run-api:
	@echo "Running API..."
	@API_HTTP_ADDR=":8080" \
	ORCHESTRATOR_GRPC_ADDR="localhost:50051" \
	$(GO) run ./cmd/api

run-orchestrator:
	@echo "Running Orchestrator..."
	@ORCHESTRATOR_GRPC_ADDR=":50051" \
	WORKER_GRPC_ADDR="localhost:50052" \
	$(GO) run ./cmd/orchestrator

run-node:
	@echo "Running Node Worker..."
	@NODE_GRPC_ADDR=":50052" \
	$(GO) run ./cmd/node

## === Docker helpers ===

docker-up:
	@echo "Starting docker stack..."
	@$(COMPOSE) up --build

docker-down:
	@echo "Stopping docker stack..."
	@$(COMPOSE) down

docker-logs:
	@echo "Tailing docker logs (Ctrl+C to exit)..."
	@$(COMPOSE) logs -f

## === Cleanup ===

clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BIN_DIR)
