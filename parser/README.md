# Chronos Parser

Rust sidecar of the Chronos system. Parses source code into a normalized AST and exposes it via gRPC.

## Purpose

Chronos is a Code Knowledge Graph combined with AI. This parser is the **first stage** of the pipeline: it reads source code and produces structured AST that downstream modules consume.

It is a **stateless compute service**. The Go core owns all business logic and databases.

## Status

| Area | Status |
| :--- | :--- |
| **F1 — Parse Go** | ✅ Working (5/5 entities extracted) |
| **gRPC Server** | ✅ Working (port 50051) |
| **gRPC Reflection** | ✅ Working (testable via `grpcurl`) |
| **CLI** | ✅ Working |
| **Java/Python/Rust/TS parsers** | ❌ Not implemented |
| **Parallel parsing** | ❌ Not implemented |
| **Interface methods extract** | ❌ Not implemented |

See [`TODO.md`](./TODO.md) for the full backlog.

## Architecture

       ┌────────────────┐
File ──│ detector.rs    │──▶ Language (Go / Java / Python / Rust / TS)
       └────────────────┘
              │
              ▼
       ┌────────────────┐
       │ factory.rs     │──▶ Box<dyn LanguageParser>
       └────────────────┘
              │
              ▼
       ┌────────────────┐
       │ providers/go.rs│──▶ Tree-sitter parse → walk AST → extract entities
       └────────────────┘
              │
              ▼
       ┌────────────────┐     ┌────────────────┐
       │ ast::Ast       │────▶│ gRPC Server    │──▶ Go Core
       └────────────────┘     └────────────────┘

## gRPC Interface

The server implements `chronos.parser.v1.ParserService`.

- **Method:** `ParseFile(ParseFileRequest) returns (ParseFileResponse)`
- **Input:** `path` (string), `source` (string)
- **Output:** `ast` (protobuf Ast), `error` (string, if parse failed)

## Usage

### 1. Run gRPC Server (Port 50051)

    cargo run --bin parser_server

### 2. Test with grpcurl (Windows PowerShell)

    # List services
    grpcurl -plaintext localhost:50051 list

    # Call method (note the --% to bypass PowerShell parsing)
    grpcurl --% -plaintext -d "{\"path\":\"test.go\",\"source\":\"package main\n\nfunc Add(a, b int) int { return a + b }\"}" localhost:50051 chronos.parser.v1.ParserService/ParseFile

### 3. Run CLI (Direct JSON output)

    cargo run -- parse testdata/sample.go

## Project layout

    parser/
    ├── build.rs                    — Compile proto + generate descriptor
    ├── proto/                      — (Symlink/Copy from root proto/)
    ├── src/
    │   ├── main.rs                 — CLI entry point
    │   ├── bin/parser_server.rs    — gRPC server entry point
    │   ├── lib.rs                  — public API
    │   ├── ast/                    — domain model
    │   ├── lang/                   — language providers
    │   └── grpc/                   — gRPC generated code + conversion + service

## Dependencies

| Crate | Purpose |
| :--- | :--- |
| `tree-sitter` | Parser engine |
| `tonic` / `prost` | gRPC + Protobuf |
| `tokio` | Async runtime |
| `clap` | CLI argument parsing |
| `serde` / `serde_json` | JSON output |
| `sha2` / `hex` | SHA-256 hashing |