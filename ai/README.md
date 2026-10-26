# Chronos AI

Python sidecar of the Chronos system. Handles embedding, vector storage, retrieval, and RAG-based question answering.

## Purpose

This AI service implements the **semantic layer**: it turns text (code chunks) into vectors, stores them in ChromaDB, and answers natural-language questions about the codebase with cited evidence.

It is a **stateless compute service**. The Go core owns all databases.

## Status

| Area | Status |
| :--- | :--- |
| **F5 — Embedding** | ✅ Working |
| **F6 — Vector Search** | ✅ Working (ChromaDB) |
| **F7 — Context Query (RAG)** | ✅ Working (LiteLLM + Gemini) |
| **gRPC Server** | ✅ Working (port 50052) |
| **gRPC Reflection** | ✅ Working |
| **CLI: ingest** | ✅ Working |
| **CLI: query** | ❌ Not implemented (gRPC replaces it) |

See [`TODO.md`](./TODO.md) for the full backlog.

## Architecture

        ┌──────────────────┐
JSON ──▶│  chunker.py      │──▶ Chunk[] (text + metadata)
        └──────────────────┘
                 │
                 ▼
        ┌──────────────────┐
        │  embedder.py     │──▶ vector[] (384d)
        │  (MiniLM)        │
        └──────────────────┘
                 │
                 ▼
        ┌──────────────────┐     ┌──────────────────┐
        │  store.py        │     │  gRPC Server     │
        │  (ChromaDB)      │────▶│  (AIService)     │──▶ Go Core
        └──────────────────┘     └──────────────────┘

## gRPC Interface

Implements `chronos.ai.v1.AIService`.

- **Embed(EmbedRequest) -> EmbedResponse:** Text -> 384d vectors.
- **Retrieve(RetrieveRequest) -> RetrieveResponse:** Query -> Top-K chunks.
- **Query(QueryRequest) -> QueryResponse:** RAG pipeline (LLM + citations).

## Usage

### 1. Run gRPC Server (Port 50052)

    # Terminal 1
    python -m src.bin.ai_server

### 2. Test with grpcurl (Terminal 2)

    # List services
    grpcurl -plaintext localhost:50052 list

    # Test Embed
    grpcurl --% -plaintext -d "{\"texts\":[\"func GetUser(id int)\"]}" localhost:50052 chronos.ai.v1.AIService/Embed

## Project layout

    ai/
    ├── src/
    │   ├── bin/ai_server.py        — gRPC server entry
    │   ├── grpc/                   — generated stubs + server impl + converter
    │   ├── models/                 — Pydantic (domain & AST)
    │   ├── llm.py                  — LiteLLM wrapper
    │   ├── chunker.py              — AST -> text chunks
    │   ├── embedder.py             — text -> vector
    │   ├── store.py                — ChromaDB wrapper
    │   ├── retriever.py            — query -> chunks
    │   └── rag.py                  — RAG pipeline

## Dependencies

| Package | Purpose |
| :--- | :--- |
| `sentence-transformers` | Embedding models |
| `chromadb` | Vector DB |
| `litellm` | LLM orchestrator |
| `grpcio` / `grpcio-tools` | gRPC Server |
| `pydantic` | Data validation |