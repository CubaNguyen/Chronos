# ai/src/embedder.py

from sentence_transformers import SentenceTransformer

from .config import EMBEDDING_DIM, EMBEDDING_MODEL


class Embedder:
    """
    Wrapper quanh Sentence-Transformers.

    Load model 1 lần khi khởi tạo, giữ trong RAM.
    Sau đó gọi embed/embed_batch nhiều lần không cần load lại.
    """

    def __init__(self, model_name: str = EMBEDDING_MODEL) -> None:
        self.model_name = model_name
        self._model = SentenceTransformer(model_name)

        # Verify dimension match config — fail sớm nếu mismatch
        actual_dim = self._model.get_embedding_dimension()
        if actual_dim != EMBEDDING_DIM:
            raise ValueError(
                f"Model {model_name} has dim={actual_dim}, "
                f"config expects {EMBEDDING_DIM}. "
                f"Update EMBEDDING_DIM in config.py."
            )

    @property
    def dimension(self) -> int:
        return EMBEDDING_DIM

    def embed(self, text: str) -> list[float]:
        """Embed 1 đoạn text, trả về vector float."""
        vec = self._model.encode(text, convert_to_numpy=True)
        return vec.tolist()

    def embed_batch(self, texts: list[str]) -> list[list[float]]:
        """Embed nhiều text cùng lúc — nhanh hơn loop embed() nhiều lần."""
        if not texts:
            return []
        vecs = self._model.encode(
            texts,
            convert_to_numpy=True,
            show_progress_bar=len(texts) > 100,  # hiện progress bar nếu batch lớn
            batch_size=32,
        )
        return [v.tolist() for v in vecs]