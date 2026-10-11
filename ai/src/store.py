# ai/src/store.py

import logging
from typing import Any

import chromadb
from chromadb.config import Settings

from .config import CHROMA_COLLECTION, CHROMA_DIR
from .models import Chunk

logger = logging.getLogger(__name__)

# ChromaDB giới hạn batch size — insert quá nhiều 1 lần sẽ lỗi
_BATCH_SIZE = 1000

# ChromaDB metadata chỉ nhận scalar types: str, int, float, bool
# Không nhận None, list, dict.
_VALID_META_TYPES = (str, int, float, bool)


class VectorStore:
    """
    Wrapper quanh ChromaDB persistent client.

    - Không embed — nhận vector từ ngoài.
    - Idempotent — dùng upsert.
    - Cosine similarity — set khi tạo collection.
    """

    def __init__(
        self,
        persist_dir: str = str(CHROMA_DIR),
        collection_name: str = CHROMA_COLLECTION,
    ) -> None:
        self.persist_dir = persist_dir
        self.collection_name = collection_name

        # PersistentClient — tự động load data có sẵn nếu tồn tại
        self.client = chromadb.PersistentClient(
            path=persist_dir,
            settings=Settings(anonymized_telemetry=False),  # tắt telemetry
        )

        # Cosine space — dùng cho embedding vectors
        self.collection = self.client.get_or_create_collection(
            name=collection_name,
            metadata={"hnsw:space": "cosine"},
        )

        logger.info(
            "VectorStore ready: collection=%s, count=%d, dir=%s",
            collection_name,
            self.collection.count(),
            persist_dir,
        )

    # ─── Write ────────────────────────────────────────────

    def add(self, chunks: list[Chunk], embeddings: list[list[float]]) -> None:
        """Upsert chunks + embeddings vào ChromaDB."""
        if len(chunks) != len(embeddings):
            raise ValueError(
                f"chunks ({len(chunks)}) và embeddings ({len(embeddings)}) "
                f"phải cùng độ dài"
            )

        if not chunks:
            return

        ids = [c.id for c in chunks]
        documents = [c.text for c in chunks]
        metadatas = [self._prepare_metadata(c) for c in chunks]

        # Chia batch để tránh ChromaDB max batch size
        total = len(chunks)
        for i in range(0, total, _BATCH_SIZE):
            end = min(i + _BATCH_SIZE, total)
            self.collection.upsert(
                ids=ids[i:end],
                embeddings=embeddings[i:end],
                documents=documents[i:end],
                metadatas=metadatas[i:end],
            )
            logger.info("Upserted %d/%d chunks", end, total)

    # ─── Read ─────────────────────────────────────────────

    def search(
        self,
        query_vector: list[float],
        top_k: int = 5,
        filters: dict[str, Any] | None = None,
    ) -> list[dict[str, Any]]:
        """
        Similarity search.

        Trả về list dict: {id, text, metadata, score}
        score = 1 - distance (cosine similarity, càng cao càng giống).
        """
        result = self.collection.query(
            query_embeddings=[query_vector],
            n_results=top_k,
            where=filters,
            include=["documents", "metadatas", "distances"],
        )

        # ChromaDB wrap mỗi kết quả trong list (vì query có thể có nhiều vector).
        # Ở đây chỉ query 1 vector → lấy [0].
        ids = result["ids"][0]
        documents = result["documents"][0]
        metadatas = result["metadatas"][0]
        distances = result["distances"][0]

        return [
            {
                "id": ids[i],
                "text": documents[i],
                "metadata": metadatas[i],
                "score": 1.0 - distances[i],  # distance → similarity
            }
            for i in range(len(ids))
        ]

    def count(self) -> int:
        """Số lượng chunks hiện có trong collection."""
        return self.collection.count()

    def get_by_ids(self, ids: list[str]) -> list[dict[str, Any]]:
        """Lấy chunks theo ID (không search)."""
        if not ids:
            return []
        result = self.collection.get(ids=ids, include=["documents", "metadatas"])
        return [
            {
                "id": result["ids"][i],
                "text": result["documents"][i],
                "metadata": result["metadatas"][i],
            }
            for i in range(len(result["ids"]))
        ]

    # ─── Admin ────────────────────────────────────────────

    def delete_by_ids(self, ids: list[str]) -> None:
        """Xóa chunks theo ID."""
        if not ids:
            return
        self.collection.delete(ids=ids)
        logger.info("Deleted %d chunks", len(ids))

    def reset(self) -> None:
        """Xóa toàn bộ collection và tạo lại. Dùng cho test."""
        logger.warning("Resetting collection %s", self.collection_name)
        self.client.delete_collection(self.collection_name)
        self.collection = self.client.get_or_create_collection(
            name=self.collection_name,
            metadata={"hnsw:space": "cosine"},
        )

    # ─── Helpers ──────────────────────────────────────────

    def _prepare_metadata(self, chunk: Chunk) -> dict[str, Any]:
        """Validate metadata — ChromaDB chỉ nhận scalar types."""
        meta: dict[str, Any] = {
            "source_type": chunk.source_type,
            "source_path": chunk.source_path,
        }
        for k, v in chunk.metadata.items():
            if v is None:
                meta[k] = ""
            elif isinstance(v, _VALID_META_TYPES):
                meta[k] = v
            else:
                meta[k] = str(v)
        return meta