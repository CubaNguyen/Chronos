# 🧠 Chronos AI

> Python sidecar of the Chronos system. Handles embedding, vector storage, retrieval, and RAG-based question answering over code.

---

## 🎯 Purpose

Chronos is a Code Knowledge Graph combined with AI. This AI service implements the **semantic layer**: it turns text (code chunks, commit messages, tickets) into vectors, stores them, and answers natural-language questions about the codebase with cited evidence.

It is a **stateless compute service**. The Go core owns all databases — this service only computes embeddings and orchestrates LLM calls.

---

## 📊 Status

| Area | Status |
| :--- | :--- |
| 🟢 **F5 — Embedding** | Working (`all-MiniLM-L6-v2`, 384d) |
| 🟢 **F6 — Vector Search** | Working (ChromaDB, cosine) |
| 🟢 **F7 — Context Query (RAG)** | Working (LiteLLM + Gemini) |
| 🟢 **CLI: ingest** | Working |
| 🔴 **CLI: query** | Not implemented yet |
| 🟡 **gRPC server** | Not implemented (next milestone) |
| 🔴 **Semantic Caching** | Phase 2 |
| 🔴 **Re-ranking** | Phase 2 |
| 🔴 **RAG Evaluation** | Phase 2 |
| 🔴 **LLM Observability** | Phase 2 |

> 📌 See [TODO.md](./TODO.md) for the full backlog.

---

## 🏗️ Architecture

    ┌──────────────────┐
    │ JSON AST files   │────▶│ chunker.py │──▶ Chunk[] (text + metadata)
    │ (from Rust)      │     └──────────────────┘
    └──────────────────┘              │
                                      ▼
                            ┌──────────────────┐
                            │  embedder.py     │──▶ vector[] (384d)
                            │  (MiniLM)        │
                            └──────────────────┘
                                      │
                                      ▼
                            ┌──────────────────┐
                            │   store.py       │──▶ ChromaDB
                            │  (persistent)    │    (data/chroma/)
                            └──────────────────┘
                                      │
                              ┌───────┘ query time
                              ▼
                    ┌──────────────┐   ┌──────────────┐   ┌──────────────┐
                    │  embedder    │──▶│  retriever   │──▶│   rag.py     │
                    │ (query→vec)  │   │  (top-k)     │   │ (LLM + cite) │
                    └──────────────┘   └──────────────┘   └──────────────┘
                                                                  │
                                                                  ▼
                                                        RAGAnswer (text + sources)

---

## ⚙️ How it works

### 🧬 F5 — Embedding

Turns text into **384-dimensional vectors** using `all-MiniLM-L6-v2` (Sentence-Transformers). Model runs on CPU, ~80MB, cached locally after first load.

### 🔍 F6 — Vector Search

Stores vectors in **ChromaDB** (persistent, embedded) with **cosine similarity**. `upsert` semantics — same chunk ID → no duplicate on re-ingest.

### 🤖 F7 — Context Query (RAG)

Pipeline:

    query → embed → retrieve top-K → filter by min_score 
          → build prompt → LLM → append citations

If no chunks pass `min_score=0.3`, returns a fixed *"I don't have enough context"* message **without calling the LLM** (saves cost).

---

## 🚀 Setup

### 📋 Prerequisites

- 🐍 Python **3.10+** (tested on 3.13)
- 🔧 C compiler for some transitive deps (`torch`, `onnxruntime`)

### 📦 Install

    cd ai
    python -m venv .venv
    .venv\Scripts\activate          # Windows
    # source .venv/bin/activate     # Linux/macOS
    pip install -r requirements.txt

### 🔐 Configure

Create `ai/.env`:

    GEMINI_API_KEY=your_key_here
    CHRONOS_LLM_MODEL=gemini/gemini-1.5-flash
    CHRONOS_LLM_FALLBACKS=gemini/gemini-2.0-flash,groq/llama-3.1-70b-versatile

> 🔑 Get a free API key at https://aistudio.google.com/app/apikey

---

## 🎬 Usage

### 1️⃣ Ingest AST JSON into ChromaDB

    # Single file
    python -m src.ingest data/asts/sample.json

    # Whole directory
    python -m src.ingest data/asts/

    # Reset collection first, then ingest
    python -m src.ingest data/asts/ --reset

    # Chunk only (skip embedding), fast debug
    python -m src.ingest data/asts/ --no-embed

### 2️⃣ Test the RAG pipeline

    python -m scripts.test_rag

### 3️⃣ Generate AST JSON (from parser)

    cd ../parser
    cargo run -- parse testdata/sample.go > ../ai/data/asts/sample.json

> ⚠️ **Windows note:** Always use `Out-File -Encoding utf8` when redirecting to a file, or use `cmd` instead of PowerShell. PowerShell's `>` writes UTF-16 LE by default, which breaks JSON parsing.

---

## 📁 Project layout

    ai/
    ├── .venv/                      — virtual environment (gitignored)
    ├── .env                        — API keys (gitignored)
    ├── pyproject.toml
    ├── requirements.txt
    ├── README.md
    ├── TODO.md
    │
    ├── src/
    │   ├── __init__.py
    │   ├── config.py               — env vars, paths, model names
    │   ├── llm.py                  — LLM client wrapper (LiteLLM)
    │   ├── chunker.py              — AST → text chunks
    │   ├── embedder.py             — F5: text → vector
    │   ├── store.py                — F6: ChromaDB wrapper
    │   ├── retriever.py            — F6: query → top-k chunks
    │   ├── rag.py                  — F7: RAG pipeline
    │   ├── ingest.py               — CLI: JSON → chunks → vectors → DB
    │   ├── query.py                — CLI: question → answer (not yet)
    │   │
    │   └── models/
    │       ├── __init__.py
    │       ├── core.py             — Chunk, QueryResult, RAGAnswer, IngestStats
    │       └── ast.py              — Pydantic mirror of Rust parser output
    │
    ├── scripts/
    │   ├── __init__.py
    │   ├── test_embedder.py        — manual test: embed similarity
    │   ├── test_store.py           — manual test: ingest + search
    │   ├── test_retriever.py       — manual test: query → results
    │   ├── test_llm.py             — manual test: LLM call
    │   └── test_rag.py             — manual test: full RAG pipeline
    │
    ├── data/
    │   ├── asts/                   — JSON files from Rust parser
    │   └── chroma/                 — ChromaDB persistence
    │
    └── tests/                      — pytest tests (TODO)

---

## 📦 Dependencies

| Package | Purpose |
| :--- | :--- |
| 🧠 `sentence-transformers` | Embedding models (`all-MiniLM-L6-v2`) |
| 🔥 `torch` | Backend for embedding (CPU) |
| 💾 `chromadb` | Vector database (embedded, persistent) |
| 🤖 `litellm` | Unified LLM API (Gemini, Groq, Ollama...) |
| 🔑 `python-dotenv` | Load `.env` file |
| ✅ `pydantic` | Data validation & serialization |
| 🔢 `numpy` | Vector math (transitive) |

---

## 💡 Design decisions

### Why ChromaDB (not Qdrant/pgvector)?

**Learning phase.** ChromaDB is embedded (no server), pure Python, zero setup. The `VectorStore` interface is designed so swapping to pgvector (when Go core owns the DB) requires only a new implementation — no changes to callers.

### Why `all-MiniLM-L6-v2` (not CodeBERT)?

**Speed + size.** 80MB, runs on CPU in milliseconds. CodeBERT is heavier and slower for marginal quality gain. Can swap later via `EMBEDDING_MODEL` env var.

### Why LiteLLM (not native SDK)?

**One interface, many providers.** Switching from Gemini to Groq to Ollama is a string change. Built-in retry + fallback. No vendor lock-in.

### Why `min_score=0.3`?

**Conservative threshold.** Below 0.3 means the query is likely unrelated to any chunk. Skip LLM call → save cost + latency. Tunable.

### Why flat chunk metadata?

ChromaDB only accepts **scalar** metadata types (`str/int/float/bool`), no nested dict. Chunker flattens everything.

---

## 🔗 Links

- 📐 Architecture doc: `../../docs/02-architecture.md`
- 🗺️ Roadmap: `../../docs/05-roadmap.md`
- 🦀 Parser README: `../parser/README.md`
- 📌 TODO: `TODO.md`