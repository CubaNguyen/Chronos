core/
├── cmd/chronos/main.go # Giữ nguyên (chỉ wiring)
├── internal/
│ ├── config/config.go # Giữ nguyên
│ │
│ ├── core/ # (1) TRÁI TIM CỦA APP (Business Logic - DDD)
│ │ ├── pipeline/ # F9: Logic điều phối (Rust -> Python -> DB)
│ │ ├── query/ # F7: Logic Context Query (Vector + Graph + LLM)
│ │ └── ingest/ # F2: Logic Git Ingest
│ │
│ ├── ports/ # (2) CÁC INTERFACE (Hợp đồng)
│ │ ├── parser.go # Interface cho Rust
│ │ ├── ai.go # Interface cho Python
│ │ └── repository.go # Interface cho DB (Postgres, Neo4j)
│ │
│ └── adapters/ # (3) CÁC IMPLEMENTATION (Giao tiếp bên ngoài)
│ ├── grpc/ # Chuyển từ internal/ai và internal/parser sang đây
│ │ ├── ai_client.go # (Trước là internal/ai/client.go)
│ │ └── parser_client.go # (Trước là internal/parser/client.go)
│ ├── http/ # Chuyển từ internal/api sang đây
│ │ └── router.go # (Trước là internal/api/router.go)
│ ├── db/ # F4, F6: Nơi chứa code DB
│ │ ├── postgres.go
│ │ └── neo4j.go
│ └── message/ # F9: NATS consumer/publisher
│ └── nats.go
│
├── pkg/contracts/ # Giữ nguyên (chứa code gen protobuf)
├── chronos.exe
├── go.mod
└── go.sum
