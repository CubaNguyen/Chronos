from typing import Any, Literal

from pydantic import BaseModel, Field

# ─── Type aliases ─────────────────────────────────────────

SourceType = Literal[
    "function",
    "method",
    "class",
    "field",
    "commit",
    "ticket",
    "doc",
]

# ─── Models ───────────────────────────────────────────────

class Chunk(BaseModel):
    """
    Một đơn vị text đã sẵn sàng để embed.

    `id` nên là SHA-256 của (source_path + line_start + text) để
    đảm bảo deterministic — cùng input → cùng ID → không lưu trùng.
    """

    id: str
    text: str
    source_type: SourceType
    source_path: str
    metadata: dict[str, Any] = Field(default_factory=dict)


class QueryResult(BaseModel):
    """Kết quả trả về từ similarity search."""

    chunk: Chunk
    score: float  # cosine similarity, thường trong [-1, 1]


class IngestStats(BaseModel):
    """Thống kê quá trình ingest, dùng để in ra CLI."""

    total_files: int = 0
    total_chunks: int = 0
    skipped_files: int = 0
    duration_seconds: float = 0.0

class RAGAnswer(BaseModel):
    """Kết quả từ RAG pipeline — câu trả lời + dẫn chứng."""

    answer: str
    sources: list[QueryResult]
    query: str
    duration_seconds: float = 0.0