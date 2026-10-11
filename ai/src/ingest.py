# ai/src/ingest.py

import argparse
import json
import logging
import sys
import time
from pathlib import Path

from .chunker import Chunker
from .config import ensure_dirs
from .embedder import Embedder
from .models import Ast, IngestStats
from .store import VectorStore

logger = logging.getLogger(__name__)


def discover_ast_files(path: Path) -> list[Path]:
    """Nếu path là dir → glob *.json; nếu là file → [path]."""
    if path.is_file():
        return [path]
    if path.is_dir():
        return sorted(path.glob("*.json"))
    raise FileNotFoundError(f"Path not found: {path}")


def load_ast(path: Path) -> Ast:
    """Load + validate 1 file JSON thành Ast."""
    data = json.loads(path.read_text(encoding="utf-8"))
    return Ast.model_validate(data)


def main() -> int:
    parser = argparse.ArgumentParser(
        description="Ingest AST JSON files into ChromaDB"
    )
    parser.add_argument("path", help="JSON file or directory of JSON files")
    parser.add_argument("--reset", action="store_true", help="Reset collection first")
    parser.add_argument("--no-embed", action="store_true", help="Skip embedding step")
    parser.add_argument("-v", "--verbose", action="store_true", help="Verbose output")
    args = parser.parse_args()

    logging.basicConfig(
        level=logging.DEBUG if args.verbose else logging.INFO,
        format="%(levelname)s %(name)s: %(message)s",
    )

    ensure_dirs()

    start = time.perf_counter()
    stats = IngestStats()

    # 1. Discover files
    path = Path(args.path)
    try:
        files = discover_ast_files(path)
    except FileNotFoundError as e:
        print(f"error: {e}", file=sys.stderr)
        return 1

    print(f"Found {len(files)} AST file(s)")

    # 2. Load ASTs
    asts: list[Ast] = []
    for f in files:
        try:
            asts.append(load_ast(f))
            stats.total_files += 1
        except Exception as e:
            logger.warning("Failed to load %s: %s", f, e)
            stats.skipped_files += 1

    if not asts:
        print("No ASTs loaded. Exiting.", file=sys.stderr)
        return 1

    print(f"Loaded {stats.total_files} AST(s), skipped {stats.skipped_files}")

    # 3. Chunk
    chunker = Chunker()
    chunks = chunker.chunk_asts(asts)
    stats.total_chunks = len(chunks)
    print(f"Generated {stats.total_chunks} chunks")

    if args.verbose:
        for c in chunks:
            print(f"  [{c.source_type:8s}] {c.metadata.get('name', '?'):20s} {c.id[:12]}")

    # 4. Embed (optional)
    if args.no_embed:
        stats.duration_seconds = time.perf_counter() - start
        _print_stats(stats)
        return 0

    print("Loading embedding model...")
    embedder = Embedder()

    print(f"Embedding {len(chunks)} chunks...")
    texts = [c.text for c in chunks]
    vectors = embedder.embed_batch(texts)

    # 5. Store
    print("Storing in ChromaDB...")
    store = VectorStore()
    if args.reset:
        store.reset()
    store.add(chunks, vectors)

    # 6. Stats
    stats.duration_seconds = time.perf_counter() - start
    _print_stats(stats)
    print(f"Collection size: {store.count()} chunks")

    return 0 if stats.skipped_files == 0 else 1


def _print_stats(stats: IngestStats) -> None:
    print()
    print("=" * 50)
    print(f"  Files loaded:    {stats.total_files}")
    print(f"  Files skipped:   {stats.skipped_files}")
    print(f"  Chunks created:  {stats.total_chunks}")
    print(f"  Duration:        {stats.duration_seconds:.2f}s")
    print("=" * 50)


if __name__ == "__main__":
    sys.exit(main())