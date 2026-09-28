.PHONY: help setup infra-up infra-down core parser ai web proto

help:
	@echo "Chronos - Available commands:"
	@echo "  make setup       - Setup môi trường dev"
	@echo "  make infra-up    - Khởi động Neo4j, Postgres, Redis, NATS"
	@echo "  make infra-down  - Dừng infra"
	@echo "  make core        - Chạy Go core"
	@echo "  make parser      - Chạy Rust parser"
	@echo "  make ai          - Chạy Python AI"
	@echo "  make web         - Chạy Web UI"
	@echo "  make proto       - Generate code từ .proto"

setup:
	@bash scripts/setup.sh

infra-up:
	docker compose -f deploy/docker-compose.infra.yml up -d

infra-down:
	docker compose -f deploy/docker-compose.infra.yml down

core:
	cd core && go run ./cmd/chronos

parser:
	cd parser && cargo run

ai:
	cd ai && source .venv/bin/activate && uvicorn app.main:app --reload

web:
	cd web && npm run dev

proto:
	@bash scripts/gen-proto.sh