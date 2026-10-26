"""Convert giữa domain models (src/models/core.py) và protobuf messages."""

from src.models.core import Chunk, QueryResult, RAGAnswer
from src.grpc import ai_pb2


# ─── Domain → Proto ──────────────────────────────────────

def chunk_to_proto(chunk: Chunk) -> ai_pb2.Chunk:
    """Chunk domain → protobuf Chunk."""
    # Metadata trong proto là map<string, string>, cần cast mọi value sang str
    metadata_str = {k: str(v) for k, v in chunk.metadata.items()}
    return ai_pb2.Chunk(
        id=chunk.id,
        text=chunk.text,
        source_type=chunk.source_type,
        source_path=chunk.source_path,
        metadata=metadata_str,
    )


def query_result_to_proto(result: QueryResult) -> ai_pb2.ScoredChunk:
    """QueryResult domain → protobuf ScoredChunk."""
    return ai_pb2.ScoredChunk(
        chunk=chunk_to_proto(result.chunk),
        score=float(result.score),
    )


def rag_answer_to_proto(answer: RAGAnswer, no_context: bool = False) -> ai_pb2.QueryResponse:
    """RAGAnswer domain → protobuf QueryResponse."""
    return ai_pb2.QueryResponse(
        answer=answer.answer,
        sources=[query_result_to_proto(r) for r in answer.sources],
        query=answer.query,
        duration_seconds=float(answer.duration_seconds),
        no_context=no_context,
        error="",
    )