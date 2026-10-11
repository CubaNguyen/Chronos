"""Profile từng phase của RAG pipeline để tìm bottleneck."""

import time
from typing import Any, Callable, TypeVar

from src.embedder import Embedder
from src.llm import LiteLLMClient
from src.rag import RAG
from src.retriever import Retriever
from src.store import VectorStore

T = TypeVar("T")


def timed(label: str, fn: Callable[[], T]) -> T:
    """Đo thời gian 1 function call."""
    start = time.perf_counter()
    result = fn()
    duration = time.perf_counter() - start
    print(f"⏱️  {label}: {duration:.3f}s")
    return result


def main() -> None:
    print("=" * 60)
    print("PROFILE RAG PIPELINE")
    print("=" * 60)

    # ─── Setup ────────────────────────────────────────────
    print("\n📦 SETUP:")
    embedder = timed("  Load Embedder", lambda: Embedder())
    store = timed("  Init VectorStore", lambda: VectorStore())
    retriever = timed("  Init Retriever", lambda: Retriever(embedder, store))
    llm = timed("  Init LLM", lambda: LiteLLMClient())
    rag = timed("  Init RAG", lambda: RAG(retriever, llm))

    print(f"\n  📊 Store count: {store.count()} chunks")
    print(f"  📊 LLM model:   {llm.model_name}")

    # ─── Query ────────────────────────────────────────────
    question = "What methods does UserService have?"
    print(f"\n🔍 QUERY: {question}")

    # Phase 1: Embed query (model đã load → nhanh)
    qvec = timed("  Embed query", lambda: embedder.embed(question))

    # Phase 2: Retrieve
    results = timed(
        "  Retrieve top-k",
        lambda: store.search(qvec, top_k=5),
    )
    print(f"     → {len(results)} chunks (top score: {results[0]['score']:.3f})")

    # Phase 3: Full RAG query (bao gồm LLM)
    print(f"\n🔍 FULL RAG QUERY:")
    result = timed("  RAG query (retrieve + LLM)", lambda: rag.query(question, top_k=3))
    print(f"     → {result.answer[:100]}...")

    # ─── Summary ──────────────────────────────────────────
    print("\n" + "=" * 60)
    print("SUMMARY")
    print("=" * 60)
    print(f"Total RAG duration: {result.duration_seconds:.2f}s")


if __name__ == "__main__":
    main()