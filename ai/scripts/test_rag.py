"""Manual test cho RAG pipeline."""

from src.embedder import Embedder
from src.llm import LiteLLMClient
from src.rag import RAG
from src.retriever import Retriever
from src.store import VectorStore


def main() -> None:
    print("🔧 Loading embedder (may take 3-8s first time)...")
    embedder = Embedder()
    store = VectorStore()
    retriever = Retriever(embedder=embedder, store=store)
    llm = LiteLLMClient()
    rag = RAG(retriever=retriever, llm=llm)

    print(f"✅ Ready. Store has {store.count()} chunks.")
    print(f"✅ LLM model: {llm.model_name}")

    queries = [
        "What methods does UserService have?",
        "What is the signature of GetUser?",
        "What is the capital of France?",  # ngoài scope → test no-context
    ]

    for q in queries:
        print(f"\n{'=' * 60}")
        print(f"Query: {q}")
        print("=" * 60)
        result = rag.query(q, top_k=3)
        print(result.answer)
        print(
            f"\n[duration: {result.duration_seconds:.2f}s, "
            f"sources: {len(result.sources)}]"
        )


if __name__ == "__main__":
    main()