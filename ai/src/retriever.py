# ai/src/retriever.py

import logging
from typing import Any

from .embedder import Embedder
from .models import Chunk, QueryResult
from .store import VectorStore

logger = logging.getLogger(__name__)


class Retriever:
    """Query text → list[QueryResult], gộp Embedder + VectorStore."""

    def __init__(self, embedder: Embedder, store: VectorStore) -> None:
        self.embedder = embedder
        self.store = store

    def retrieve(
        self,
        query: str,
        top_k: int = 5,
        filters: dict[str, Any] | None = None,
    ) -> list[QueryResult]:
        """Embed query, search store, trả về QueryResults."""
        qvec = self.embedder.embed(query)
        raw = self.store.search(qvec, top_k=top_k, filters=filters)
        return [self._to_query_result(r) for r in raw]

    def _to_query_result(self, r: dict[str, Any]) -> QueryResult:
        """Convert store dict → QueryResult model."""
        meta = dict(r["metadata"])  # copy để không mutate

        # source_type và source_path đã lưu top-level trong metadata
        source_type = meta.pop("source_type")
        source_path = meta.pop("source_path")

        chunk = Chunk(
            id=r["id"],
            text=r["text"],
            source_type=source_type,
            source_path=source_path,
            metadata=meta,
        )
        return QueryResult(chunk=chunk, score=float(r["score"]))