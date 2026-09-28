# Chronos — Overview

> **Tagline:** Understand your codebase like the person who wrote it.

## ⚠️ Problem
Large codebases are archaeological sites without a diary. Developers spend hours answering two questions:
*   *"Why is this code like this?"* — the business reason, the ticket, the constraint, the tradeoff.
*   *"If I change this, what breaks?"* — the blast radius across files, services, and tests.

Existing tools (Git blame, Jira, IDE, ChatGPT) work in isolation. Git knows history but not tickets. Jira knows tickets but not code. IDEs know code but not history. **Nothing connects them.**

---

## 💡 Solution
Chronos is a **Code Knowledge Graph** combined with AI. It ingests Git history, tickets, and source code into a unified graph, then lets developers query it in natural language — with cited evidence from commits, PRs, and issues.

**Two core capabilities:**
1.  **Context Tracing** — Ask *"Why?"* and get an answer backed by commit hashes, PR links, and ticket IDs.
2.  **Impact Analysis** — Ask *"What breaks?"* and get the full blast radius: affected files, functions, tests, and APIs.

---

## 🏗️ Architecture
**Modular Monolith with Polyglot Sidecars.**

*   **Go Core:** Owns all business logic and databases (Neo4j, pgvector, Redis, NATS). Contains modules: `api`, `ingest`, `graph`, `vector`, `worker`, `ai`, `parser`, `queue`, `mcp`.
*   **Rust Parser Sidecar:** Stateless. Parses source code into AST using Tree-sitter. Queries graph for impact analysis. Communicates via gRPC.
*   **Python AI Sidecar:** Stateless. Handles embedding and LLM orchestration (RAG). Communicates via gRPC.
*   **TypeScript Web UI:** Interactive graph visualization and chat interface.

**System Design Rules:**
*   **Data ownership:** Only the Go core accesses databases. Rust and Python sidecars are pure compute — they receive input, return output, and hold no persistent state. This keeps them easily scalable and restartable.
*   **Communication:** In-process interfaces within Go core. gRPC between Go ↔ Rust and Go ↔ Python. NATS (JetStream) for async job processing.

---

## 🚀 Core Features (Phase 1 — MVP)

| # | Feature | Language |
|---|---|---|
| **F1** | Parse Source Code | Rust |
| **F2** | Ingest Git History | Go |
| **F3** | Ingest Tickets | Go |
| **F4** | Build Knowledge Graph | Go |
| **F5** | Embedding | Python |
| **F6** | Vector Search | Go + Python |
| **F7** | Context Query | Python + Go |
| **F8** | Impact Analysis | Rust + Go |
| **F9** | Background Worker | Go |
| **F10**| API Gateway | Go |
| **F11**| CLI Tool | Go |
| **F12**| Web UI | TypeScript |
| **F13**| MCP Server | Go |

---

## 🔮 Extended Features (Phase 2 — LLMOps & RAG Optimization)

| # | Feature | Purpose |
|---|---|---|
| **P2-1** | LLM Observability & Cost Tracking | Monitor tokens, latency, prompt/response history |
| **P2-2** | Semantic Caching | Cache similar queries, cut LLM cost 50-70% |
| **P2-3** | Model Serving & Infrastructure | Host models via Ollama / vLLM / Triton |
| **P2-4** | Re-ranking | Filter context before sending to LLM |
| **P2-5** | RAG Evaluation | Automated scoring with Ragas / TruLens |

---

## 🚫 Non-Goals
*   Not an auto-fixer. Chronos advises, humans decide.
*   Not an IDE replacement.
*   Not a public repo analyzer. Only repos you have access to.
*   No fine-tuning or custom model training in Phase 1.

---

## 🎯 Target Users
*   **Senior/Mid developers** refactoring safely.
*   **New developers** onboarding to a large codebase.
*   **Tech leads** auditing architectural decisions.
*   **AI coding agents** (Claude Code, Cursor) needing accurate context.

---

## 🧭 Guiding Principles
*   **In-process first, extract later.** Build functionality in Go. Only split into Rust/Python sidecars when there's a concrete need (performance, ecosystem, or missing capability).
*   **Adapter over hard-code.** Every external source (Jira, GitHub, Linear) is plugged via an adapter interface. No vendor lock-in.
*   **Graph owns truth.** The Knowledge Graph is the single source of truth. Vector DB and LLM are auxiliary.
*   **Evidence over speculation.** Every AI answer cites its source: commit hash, PR link, ticket ID.

---

## 📊 Status
*   [x] **Phase 0** — Setup complete (repo structure, toolchain).
*   [ ] **Phase 1** — F1 (Rust parser) in progress.