"""Shared fixtures cho toàn bộ test suite."""

import pytest

from src.llm import LLMClient
from src.models import (
    Ast,
    Call,
    Class,
    Field,
    FileInfo,
    Function,
    Import,
    Param,
)


# ─── Mock LLM ─────────────────────────────────────────────

class MockLLMClient(LLMClient):
    """Fake LLM trả về string cố định — không gọi API, không tốn tiền."""

    def __init__(self, response: str = "Mock answer.") -> None:
        self.response = response
        self.last_prompt: str | None = None
        self.usage_tracker = MockUsageTracker()  # ← thêm

    @property
    def model_name(self) -> str:
        return "mock"

    def generate(self, prompt: str) -> str:
        self.last_prompt = prompt
        return self.response


class MockUsageTracker:
    """Fake usage tracker — không track gì."""

    def summary(self) -> str:
        return "Calls: 0 | Tokens: 0 in + 0 out | Cost: $0.000000"

@pytest.fixture
def mock_llm() -> MockLLMClient:
    return MockLLMClient()


# ─── Sample AST ───────────────────────────────────────────

@pytest.fixture
def sample_ast() -> Ast:
    """
    Ast tối giản — 1 struct có 1 field + 2 methods, 1 free function.

    Không load từ file → test nhanh, không phụ thuộc testdata/.
    """
    return Ast(
        file=FileInfo(
            path="test.go",
            language="go",
            hash="dummy",
            size_bytes=100,
            line_count=20,
        ),
        classes=[
            Class(
                name="UserService",
                kind="struct",
                visibility="public",
                extends=None,
                implements=[],
                line_start=3,
                line_end=10,
                file_path="test.go",
            ),
        ],
        functions=[
            Function(
                name="GetUser",
                class_name="UserService",
                params=[Param(name="id", type_name="int")],
                return_type="(*User, error)",
                visibility="public",
                line_start=5,
                line_end=7,
                file_path="test.go",
            ),
            Function(
                name="helper",
                class_name="UserService",
                params=[],
                return_type=None,
                visibility="private",
                line_start=9,
                line_end=10,
                file_path="test.go",
            ),
            Function(
                name="Add",
                class_name=None,
                params=[
                    Param(name="a", type_name="int"),
                    Param(name="b", type_name="int"),
                ],
                return_type="int",
                visibility="public",
                line_start=13,
                line_end=15,
                file_path="test.go",
            ),
        ],
        fields=[
            Field(
                name="db",
                type_name="*sql.DB",
                visibility="private",
                class_name="UserService",
                line=4,
                file_path="test.go",
            ),
        ],
        calls=[],
        imports=[],
    )