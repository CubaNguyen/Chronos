import pytest

from src.chunker import Chunker
from src.embedder import Embedder
from src.store import VectorStore


@pytest.fixture(scope="module")
def embedder():
    return Embedder()


@pytest.fixture
def store(tmp_path):
    """ChromaDB trong tmp dir — không ảnh hưởng data thật."""
    return VectorStore(persist_dir=str(tmp_path / "chroma"))


def _prepare(sample_ast, embedder):
    chunks = Chunker().chunk_ast(sample_ast)
    vectors = embedder.embed_batch([c.text for c in chunks])
    return chunks, vectors


def test_add_and_count(store, sample_ast, embedder):
    chunks, vectors = _prepare(sample_ast, embedder)
    store.add(chunks, vectors)
    assert store.count() == len(chunks)


def test_upsert_idempotent(store, sample_ast, embedder):
    """Add 2 lần cùng data → count không đổi."""
    chunks, vectors = _prepare(sample_ast, embedder)
    store.add(chunks, vectors)
    count1 = store.count()
    store.add(chunks, vectors)
    count2 = store.count()
    assert count1 == count2


def test_search_finds_relevant_chunk(store, sample_ast, embedder):
    chunks, vectors = _prepare(sample_ast, embedder)
    store.add(chunks, vectors)

    qvec = embedder.embed("method that gets user by id")
    results = store.search(qvec, top_k=3)

    assert len(results) > 0
    top = results[0]
    assert top["metadata"]["name"] == "GetUser"
    # score phải trong khoảng hợp lệ
    assert -1.0 <= top["score"] <= 1.0


def test_add_mismatched_lengths_raises(store, sample_ast, embedder):
    chunks, vectors = _prepare(sample_ast, embedder)
    with pytest.raises(ValueError):
        store.add(chunks, vectors[:-1])