# ai/src/config.py

import os
from pathlib import Path
from typing import Final

# Project root = ai/  (parent của src/)
PROJECT_ROOT: Final[Path] = Path(__file__).resolve().parent.parent

# ─── Paths ────────────────────────────────────────────────
DATA_DIR: Final[Path] = Path(os.getenv("CHRONOS_DATA_DIR", PROJECT_ROOT / "data"))
ASTS_DIR: Final[Path] = DATA_DIR / "asts"
CHROMA_DIR: Final[Path] = DATA_DIR / "chroma"

# ─── Embedding Model ──────────────────────────────────────
EMBEDDING_MODEL: Final[str] = os.getenv("CHRONOS_EMBEDDING_MODEL", "all-MiniLM-L6-v2")
EMBEDDING_DIM: Final[int] = int(os.getenv("CHRONOS_EMBEDDING_DIM", "384"))

# ─── ChromaDB ─────────────────────────────────────────────
CHROMA_COLLECTION: Final[str] = os.getenv("CHRONOS_CHROMA_COLLECTION", "chronos_code")

# ─── Chunking ─────────────────────────────────────────────
CHUNK_MAX_LINES: Final[int] = int(os.getenv("CHRONOS_CHUNK_MAX_LINES", "50"))
CHUNK_OVERLAP_LINES: Final[int] = int(os.getenv("CHRONOS_CHUNK_OVERLAP_LINES", "5"))


def ensure_dirs() -> None:
    """Tạo các thư mục cần thiết nếu chưa tồn tại."""
    for d in (DATA_DIR, ASTS_DIR, CHROMA_DIR):
        d.mkdir(parents=True, exist_ok=True)