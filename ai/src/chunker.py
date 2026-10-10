# ai/src/chunker.py

import hashlib
import logging

from .models import Ast, Chunk, Class, Field, Function, Param

logger = logging.getLogger(__name__)

# Ngưỡng cảnh báo chunk quá dài (model all-MiniLM có max ~256 tokens ~ 1000-1200 chars)
# 2000 chars là ngưỡng an toàn để warn sớm.
_WARN_CHARS = 2000


class Chunker:
    """Convert Ast → list[Chunk] bằng cách synthesize text từ structured fields."""

    # ─── Public API ───────────────────────────────────────

    def chunk_ast(self, ast: Ast) -> list[Chunk]:
        """Convert 1 Ast thành list Chunks (classes + functions)."""
        chunks: list[Chunk] = []
        for cls in ast.classes:
            chunks.append(self._make_class_chunk(cls, ast))
        for func in ast.functions:
            chunks.append(self._make_function_chunk(func, ast))
        return chunks

    def chunk_asts(self, asts: list[Ast]) -> list[Chunk]:
        """Convert nhiều Ast cùng lúc."""
        chunks: list[Chunk] = []
        for ast in asts:
            chunks.extend(self.chunk_ast(ast))
        return chunks

    # ─── Function chunk ───────────────────────────────────

    def _make_function_chunk(self, func: Function, ast: Ast) -> Chunk:
        is_method = func.class_name is not None
        source_type = "method" if is_method else "function"

        # Header: "Method GetUser in class UserService" hoặc "Function Add"
        if is_method:
            header = f"Method {func.name} in class {func.class_name}"
        else:
            header = f"Function {func.name}"

        # Signature: "GetUser(id int) (*User, error)" — không prefix class vì đã biết
        sig = self._format_signature(func)
        text = f"{header}\nSignature: {sig}"

        full_name = self._full_name(func)
        chunk_id = self._make_id(ast.file.path, source_type, full_name, func.line_start)

        self._warn_if_long(chunk_id, text)

        return Chunk(
            id=chunk_id,
            text=text,
            source_type=source_type,  # "function" hoặc "method" — tùy có class_name
            source_path=ast.file.path,
            metadata={
                "file_path": ast.file.path,
                "language": ast.file.language,
                "line_start": func.line_start,
                "line_end": func.line_end,
                "entity_type": source_type,  # "function" hoặc "method"
                "class_name": func.class_name or "",
                "name": func.name,
                "visibility": func.visibility,
            },
        )

    # ─── Class chunk ──────────────────────────────────────

    def _make_class_chunk(self, cls: Class, ast: Ast) -> Chunk:
        lines: list[str] = []

        # Header: "Struct UserService" / "Interface Reader"
        lines.append(f"{cls.kind.capitalize()} {cls.name}")
        lines.append(f"File: {ast.file.path}")

        if cls.extends:
            lines.append(f"Extends: {cls.extends}")

        if cls.implements:
            lines.append(f"Implements: {', '.join(cls.implements)}")

        # Fields của class này
        cls_fields = [f for f in ast.fields if f.class_name == cls.name]
        if cls_fields:
            lines.append(f"Fields: {', '.join(self._format_field(f) for f in cls_fields)}")

        # Method signatures của class này
        cls_methods = [f for f in ast.functions if f.class_name == cls.name]
        if cls_methods:
            lines.append("Methods:")
            for m in cls_methods:
                lines.append(f"  - {self._format_signature(m)}")

        text = "\n".join(lines)

        chunk_id = self._make_id(ast.file.path, "class", cls.name, cls.line_start)
        self._warn_if_long(chunk_id, text)

        return Chunk(
            id=chunk_id,
            text=text,
            source_type="class",
            source_path=ast.file.path,
            metadata={
                "file_path": ast.file.path,
                "language": ast.file.language,
                "line_start": cls.line_start,
                "line_end": cls.line_end,
                "entity_type": cls.kind,  # "struct", "interface", "class"...
                "class_name": cls.name,
                "name": cls.name,
                "visibility": cls.visibility,
                "extends": cls.extends or "",
                "implements": ",".join(cls.implements),
            },
        )

    # ─── Formatting helpers ───────────────────────────────

    def _format_signature(self, func: Function) -> str:
        """`GetUser(id int, name string) (*User, error)`"""
        params = self._format_params(func.params)
        ret = f" {func.return_type}" if func.return_type else ""
        return f"{func.name}({params}){ret}"

    def _format_params(self, params: list[Param]) -> str:
        if not params:
            return ""
        return ", ".join(f"{p.name} {p.type_name}" for p in params)

    def _format_field(self, f: Field) -> str:
        """`db *sql.DB`"""
        return f"{f.name} {f.type_name}"

    def _full_name(self, func: Function) -> str:
        """`UserService.GetUser` cho method, `Add` cho free function."""
        if func.class_name:
            return f"{func.class_name}.{func.name}"
        return func.name

    # ─── ID + warning ────────────────────────────────────

    def _make_id(self, file_path: str, source_type: str, full_name: str, line: int) -> str:
        """SHA-256 deterministic ID — cùng input → cùng ID."""
        raw = f"{file_path}|{source_type}|{full_name}|{line}"
        return hashlib.sha256(raw.encode("utf-8")).hexdigest()

    def _warn_if_long(self, chunk_id: str, text: str) -> None:
        if len(text) > _WARN_CHARS:
            logger.warning(
                "Chunk %s is %d chars (> %d) — may exceed model max length",
                chunk_id[:12],
                len(text),
                _WARN_CHARS,
            )