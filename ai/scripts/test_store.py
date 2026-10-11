# ai/test_store_manual.py
"""Manual test cho VectorStore — chạy 1 lần để verify flow."""

import json
from pathlib import Path

from src.chunker import Chunker
from src.embedder import Embedder
from src.models import Ast
from src.store import VectorStore


def main() -> None:
    # 1. Load AST
    data = json.loads(Path("data/asts/sample.json").read_text(encoding="utf-8"))
    ast = Ast.model_validate(data)

    # 2. Chunk
    chunker = Chunker()
    chunks = chunker.chunk_ast(ast)
    print(f"Chunks: {len(chunks)}")

    # 3. Embed
    embedder = Embedder()
    texts = [c.text for c in chunks]
    vectors = embedder.embed_batch(texts)
    print(f"Vectors: {len(vectors)} x {len(vectors[0])}d")

    # 4. Store
    store = VectorStore()
    store.reset()
    store.add(chunks, vectors)
    print(f"Count after add: {store.count()}")

    # 5. Search
    query = "function that gets a user by id"
    qvec = embedder.embed(query)
    results = store.search(qvec, top_k=3)

    print()
    print(f"Query: {query!r}")
    print(f"Top {len(results)} results:")
    for r in results:
        name = r["metadata"]["name"]
        etype = r["metadata"]["entity_type"]
        score = r["score"]
        print(f"  score={score:.4f}  {name} ({etype})")


if __name__ == "__main__":
    main()