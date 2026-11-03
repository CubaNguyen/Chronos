.PHONY: help setup infra-up infra-down core parser ai web fe proto

help:
	@echo "Chronos - Available commands:"
	@echo "  make setup       - Setup môi trường dev"
	@echo "  make infra-up    - Khởi động Neo4j, Postgres, NATS"
	@echo "  make infra-down  - Dừng infra"
	@echo "  make core        - Chạy Go core"
	@echo "  make parser      - Chạy Rust parser"
	@echo "  make ai          - Chạy Python AI"
	@echo "  make fe          - Chạy React UI (fe)"
	@echo "  make proto       - Generate code từ .proto"

setup:
	@bash scripts/setup.sh

infra-up:
	docker compose up -d

infra-down:
	docker compose down

core:
	cd core && go run ./cmd/chronos

parser:
	cd parser && cargo run

ai:
	cd ai && source .venv/bin/activate && uvicorn app.main:app --reload

fe:
	cd fe && npm run dev

proto:
	@bash scripts/gen-proto.sh
