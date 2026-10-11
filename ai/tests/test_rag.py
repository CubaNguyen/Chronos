import pytest

from src.chunker import Chunker
from src.embedder import Embedder
from src.rag import NO_CONTEXT_MESSAGE, RAG
from src.retriever import Retriever
from src.store import VectorStore


@pytest.fixture(scope="module")
def embedder():
    return Embedder()


@pytest.fixture
def rag(tmp_path, sample_ast, embedder, mock_llm):
    """RAG pipeline đầy đủ với store trong tmp dir + mock LLM."""
    store = VectorStore(persist_dir=str(tmp_path / "chroma"))
    chunks = Chunker().chunk_ast(sample_ast)
    vectors = embedder.embed_batch([c.text for c in chunks])
    store.add(chunks, vectors)

    retriever = Retriever(embedder, store)
    return RAG(retriever, mock_llm)


def test_no_context_returns_fixed_message(rag, mock_llm):
    """Query rác → không pass min_score → fallback, không gọi LLM."""
    result = rag.query("what is the capital of france?", top_k=3)

    assert result.answer == NO_CONTEXT_MESSAGE
    assert result.sources == []
    # Verify LLM không được gọi
    assert mock_llm.last_prompt is None


def test_with_context_calls_llm(rag, mock_llm):
    """Query liên quan → gọi LLM, append sources."""
    result = rag.query("method that gets user by id", top_k=3)

    # Mock LLM được gọi
    assert mock_llm.last_prompt is not None
    assert "GetUser" in mock_llm.last_prompt
    # Answer chứa mock response + sources
    assert mock_llm.response in result.answer
    assert "Sources:" in result.answer
    assert len(result.sources) > 0


def test_query_has_duration(rag):
    result = rag.query("what is the capital of france?")
    assert result.duration_seconds >= 0