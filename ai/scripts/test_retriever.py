"""Manual test cho Retriever."""

from src.embedder import Embedder
from src.retriever import Retriever
from src.store import VectorStore


def main() -> None:
    embedder = Embedder()
    store = VectorStore()
    retriever = Retriever(embedder, store)

    queries = [
        "function that gets a user by id",
        "interface for reading data",
        "add two numbers",
    ]

    for q in queries:
        print(f"\n=== Query: {q!r} ===")
        results = retriever.retrieve(q, top_k=3)
        for r in results:
            meta = r.chunk.metadata
            print(
                f"  score={r.score:.4f}  "
                f"{meta.get('name')} ({r.chunk.source_type})"
            )


if __name__ == "__main__":
    main()