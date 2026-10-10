import numpy as np
import pytest

from src.embedder import Embedder


@pytest.fixture(scope="module")
def embedder():
    """Load model 1 lần cho cả module — tiết kiệm ~5s."""
    return Embedder()


def test_dimension(embedder):
    assert embedder.dimension == 384


def test_embed_returns_correct_length(embedder):
    v = embedder.embed("hello world")
    assert len(v) == 384


def test_similar_texts_have_high_similarity(embedder):
    """Câu giống nghĩa → similarity cao hơn câu khác nghĩa."""
    vecs = np.array(
        embedder.embed_batch(
            [
                "func GetUser(id int) (*User, error)",
                "func FindUserByID(id int) (*User, error)",
                "func PrintHello()",
            ]
        )
    )

    def cos(a, b):
        return float(np.dot(a, b) / (np.linalg.norm(a) * np.linalg.norm(b)))

    similar = cos(vecs[0], vecs[1])
    different = cos(vecs[0], vecs[2])
    assert similar > different


def test_embed_batch_empty(embedder):
    assert embedder.embed_batch([]) == []