# ai/src/models/ast.py

from typing import Literal

from pydantic import BaseModel

# ─── Enums (match Rust serde lowercase) ───────────────────

ClassKind = Literal["class", "struct", "interface", "enum", "trait"]
Visibility = Literal["public", "private", "protected", "package"]

# ─── Nested models ────────────────────────────────────────

class FileInfo(BaseModel):
    path: str
    language: str
    hash: str
    size_bytes: int
    line_count: int


class Param(BaseModel):
    name: str
    type_name: str


class Class(BaseModel):
    name: str
    kind: ClassKind
    visibility: Visibility
    extends: str | None
    implements: list[str]
    line_start: int
    line_end: int
    file_path: str


class Field(BaseModel):
    name: str
    type_name: str
    visibility: Visibility
    class_name: str | None
    line: int
    file_path: str


class Function(BaseModel):
    name: str
    class_name: str | None
    params: list[Param]
    return_type: str | None
    visibility: Visibility
    line_start: int
    line_end: int
    file_path: str


class Call(BaseModel):
    caller: str
    callee: str
    line: int
    file_path: str


class Import(BaseModel):
    module_path: str
    alias: str | None
    line: int
    is_external: bool
    file_path: str


# ─── Root ─────────────────────────────────────────────────

class Ast(BaseModel):
    file: FileInfo
    classes: list[Class]
    functions: list[Function]
    fields: list[Field]
    calls: list[Call]
    imports: list[Import]