# 📌 AI — TODO

> Backlog of remaining work for the Chronos AI sidecar.

### 🎨 Legend

| Icon | Meaning |
| :---: | :--- |
| 🔴 | **Next** — highest priority, unblocks others |
| 🟡 | **Should do** — important but not blocking |
| 🟢 | **Nice to have** — polish, edge cases |
| ⚪ | **Future** — deferred indefinitely |

---

## 🔴 NEXT

### 1. CLI: `query.py`

**🎯 Goal:** Wrap RAG pipeline thành CLI sạch — thay vì chạy `scripts/test_rag.py`.

**💡 Why:** F7 chạy được nhưng chỉ qua script. CLI = user-friendly interface.

**🛠️ Implementation:**

- [ ] `src/query.py` — argparse với positional query + flags
- [ ] Flags: `--top-k`, `--min-score`, `--json`, `--no-citations`
- [ ] Output đẹp (dùng `rich` hoặc plain text)
- [ ] Print `usage_tracker.summary()` cuối mỗi query

**📝 Usage:**

    python -m src.query "What methods does UserService have?"
    python -m src.query "..." --top-k 5 --no-citations
    python -m src.query "..." --json

**⏱️ Estimated effort:** 2 hours.

---

### 2. Pytest Unit Tests

**🐛 Problem:** Chỉ có manual smoke tests. Không CI-ready.

**🛠️ Implementation:**

- [ ] `tests/conftest.py` — fixtures (temp ChromaDB, mock embedder, mock LLM)
- [ ] `tests/test_chunker.py` — chunk count, text format, deterministic IDs
- [ ] `tests/test_embedder.py` — dimension, similarity của related texts
- [ ] `tests/test_store.py` — upsert idempotency, search correctness
- [ ] `tests/test_retriever.py` — top-1 correct cho known queries
- [ ] `tests/test_rag.py` — mock LLM, verify prompt + citation format

**🎯 Target:** ~80% coverage.

**⏱️ Estimated effort:** 1 day.

---

## 🟡 SHOULD DO

### 3. gRPC Wrap (Phase 2 start)

**🎯 Goal:** Expose AI service (embed, retrieve, query) qua gRPC → Go core gọi được.

**💡 Why:** Inflection point. Sau khi xong → Go core integrate được AI.

**🛠️ Steps:**

- [ ] Write `proto/ai.proto` — 3 methods:
  - `Embed(EmbedRequest) → EmbedResponse`
  - `Retrieve(RetrieveRequest) → RetrieveResponse`
  - `Query(QueryRequest) → QueryResponse`
- [ ] Add `grpcio` + `grpcio-tools` to `requirements.txt`
- [ ] Generate Python stubs
- [ ] Implement gRPC server wrapping `Embedder`, `Retriever`, `RAG`
- [ ] Test with `grpcurl`
- [ ] Go client trong `core/internal/ai/`
- [ ] End-to-end test

**⏱️ Estimated effort:** 1 week.

---

### 4. Multi-language Chunking

**🐛 Problem:** Chunker giả định Go syntax. Java/Python methods sẽ chunk sai.

**🛠️ Implementation:**

- [ ] Verify chunker **không phụ thuộc language** — chỉ dùng Ast fields
- [ ] Adjust text format per language nếu cần
- [ ] Test với Java/Python parser output (khi có)

---

### 5. Chunk Metadata Enrichment

**💡 Ideas:**

- [ ] `complexity` — số params, số calls bên trong function
- [ ] `is_exported` — boolean, từ visibility
- [ ] `file_extension` — `.go`, `.java`
- [ ] `repo_name` — khi multi-repo

---

## 🟢 NICE TO HAVE

### 6. Semantic Caching (P2-2)

**🐛 Problem:** Câu hỏi lặp → gọi LLM lại → tốn tiền + chậm.

**🛠️ Implementation:**

- [ ] Redis Stack hoặc Qdrant làm semantic cache
- [ ] Trước khi gọi LLM: embed query → search cache
- [ ] Similarity > 0.95 → trả cached answer
- [ ] Invalidate khi repo có commit mới

**📊 Impact:** Tiết kiệm 50-70% chi phí LLM.

---

### 7. LLM Observability (P2-1)

**🐛 Problem:** Không biết token usage, latency, prompt history.

**🛠️ Implementation:**

- [ ] Tích hợp Langfuse (self-hosted) hoặc Arize Phoenix
- [ ] Log mỗi LLM call: model, tokens, latency, cost
- [ ] Dashboard phân tích bottleneck

**📊 Impact:** Debug dễ, optimize cost.

---

### 8. Re-ranking (P2-4)

**🐛 Problem:** Top-K từ vector search có thể nhiễu → tràn context + hallucination.

**🛠️ Implementation:**

- [ ] BGE Reranker (local) hoặc Cohere Rerank (API)
- [ ] Reranker chấm điểm → chọn top 3-5 liên quan nhất
- [ ] Chỉ đưa top sau rerank vào prompt

**📊 Impact:** Tăng độ chính xác, giảm hallucination.

---

### 9. RAG Evaluation (P2-5)

**🐛 Problem:** Không biết RAG đúng hay sai. Không có số liệu để improve.

**🛠️ Implementation:**

- [ ] Tạo Golden Dataset (20-50 câu hỏi + đáp án mẫu)
- [ ] Dùng Ragas hoặc TruLens để chấm điểm:
  - **Faithfulness** — answer có bám sát context không
  - **Answer Relevance** — answer có đúng câu hỏi không
  - **Context Precision** — retrieved chunks có đúng không
- [ ] Chạy eval mỗi khi đổi prompt/model

**📊 Impact:** Đo lường khoa học, không "cảm tính".

---

### 10. Model Serving (P2-3)

**🐛 Problem:** Đang dùng Gemini API. Muốn host local để zero cost + private.

**🛠️ Implementation:**

- [ ] Ollama cho dev — `ollama pull qwen2.5-coder:7b`
- [ ] vLLM cho production nếu có GPU
- [ ] Đổi qua env: `CHRONOS_LLM_MODEL=ollama/qwen2.5-coder:7b`

**📊 Impact:** Không phụ thuộc API provider, không rate limit.

---

## ⚪ FUTURE

### Multi-modal Embedding

**💡 Idea:** Embed cả code, docs, và diagram.

**⚠️ Challenge:** Cần model multi-modal (CLIP hoặc Gemini embedding).

---

### Incremental Re-embedding

**🐛 Problem:** File thay đổi → phải re-ingest cả file.

**💡 Idea:** So sánh hash từng chunk → chỉ re-embed chunk thay đổi.

**⚠️ Challenge:** Cần track hash per-chunk trong ChromaDB metadata. Breaking change.

---

### Hybrid Search

**💡 Idea:** Vector search (semantic) + BM25 (keyword).

**💡 Why:** Query có tên function cụ thể (vd: `GetUser`) → BM25 chính xác hơn.

**⚠️ Challenge:** ChromaDB không support hybrid. Cần chuyển sang Weaviate hoặc Qdrant.

---

### Conversation History

**💡 Idea:** Multi-turn chat — user hỏi follow-up.

**🛠️ Implementation:**

- [ ] Lưu conversation vào PostgreSQL (Go core owns)
- [ ] Pass history vào prompt
- [ ] Query rewriting để LLM hiểu ngữ cảnh

**⚠️ Note:** Cross-cutting concern — thuộc Go core, không phải AI sidecar.

---

## ✅ DONE (Archive)

- [x] 🧬 **F5** — Embedder (`all-MiniLM-L6-v2`, 384d)
- [x] 🔍 **F6** — VectorStore (ChromaDB, cosine, upsert)
- [x] 🔍 **F6** — Retriever (embed query → top-k chunks)
- [x] 🤖 **F7** — RAG pipeline (retrieve + prompt + LLM + citation)
- [x] ⌨️ **CLI** — `ingest.py` (JSON → chunks → vectors → ChromaDB)
- [x] 📦 **Models** — Chunk, QueryResult, RAGAnswer, IngestStats
- [x] 🧠 **LLM** — LiteLLM wrapper với retry + fallback
- [x] ✅ **Pydantic** validation cho JSON từ Rust parser
- [x] 🔧 **A1** — Fix 3 warnings (Deprecation, Task destroyed, em dash)
- [x] 📖 **A2** — README.md
- [x] 💰 **A4** — UsageTracker (tokens + cost)
- [x] ⚡ **A5** — Performance: 23s → 1.6s/query (đổi model `gemini-flash-lite-latest`)

---

## 📖 How to use this file

1. 🎯 Pick the top **🔴 Next** item. Don't jump around.
2. 🏁 Do it **end-to-end**. Code + tests + update this file.
3. 📦 Move done items to **"Done (Archive)"** at the bottom.
4. 🔄 Re-prioritize when a blocker appears.
5. ⚠️ Do **not** add items to 🔴 Next without a concrete reason. If it's "nice to have" → 🟢.
6. 🚫 Phase 2 items (P2-X) stay in 🟢 until Phase 1 MVP is fully done.