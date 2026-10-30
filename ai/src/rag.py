# ai/src/rag.py

import logging
import time

from .llm import LLMClient
from .models import QueryResult, RAGAnswer
from .retriever import Retriever

logger = logging.getLogger(__name__)

# Ngưỡng similarity tối thiểu — dưới ngưỡng này coi như không có context
DEFAULT_MIN_SCORE = 0.3

# Message cố định khi không đủ context — LLM phải trả về y hệt
NO_CONTEXT_MESSAGE = "I don't have enough context to answer this question."


class RAG:
    """
    RAG pipeline: query → retrieve → build prompt → LLM → answer + sources.
    """

    def __init__(
        self,
        retriever: Retriever,
        llm: LLMClient,
        min_score: float = DEFAULT_MIN_SCORE,
    ) -> None:
        self.retriever = retriever
        self.llm = llm
        self.min_score = min_score

    def query(
        self,
        question: str,
        top_k: int = 5,
        contexts: list[QueryResult] | None = None,
    ) -> RAGAnswer:
        """Full RAG pipeline cho 1 câu hỏi (ưu tiên dùng contexts nếu được truyền từ ngoài)."""
        start = time.perf_counter()

        if contexts is not None:
            relevant = contexts
        else:
            # Fallback nếu không truyền context từ ngoài vào:
            results = self.retriever.retrieve(question, top_k=top_k)
            logger.debug("Retrieved %d chunks for query", len(results))
            relevant = [r for r in results if r.score >= self.min_score]
            logger.debug(
                "%d chunks passed min_score=%.2f", len(relevant), self.min_score
            )

        # Nếu không có chunk nào đủ liên quan -> trả về message cố định
        if not relevant:
            duration = time.perf_counter() - start
            logger.info("No relevant context found for query")
            return RAGAnswer(
                answer=NO_CONTEXT_MESSAGE,
                sources=[],
                query=question,
                duration_seconds=duration,
            )

        # 4. Build prompt
        prompt = self._build_prompt(question, relevant)

        # 5. Call LLM
        answer = self.llm.generate(prompt)

        # 6. Append sources (chỉ khi LLM trả lời thật, không phải NO_CONTEXT)
        if self._is_no_context(answer):
            answer = NO_CONTEXT_MESSAGE
        else:
            answer = self._append_sources(answer, relevant)

        duration = time.perf_counter() - start
        logger.info(
            "RAG query done: %d sources, %.2fs",
            len(relevant),
            duration,
        )

        return RAGAnswer(
            answer=answer,
            sources=relevant,
            query=question,
            duration_seconds=duration,
        )

    # ─── Helpers ──────────────────────────────────────────

    @staticmethod
    def _is_no_context(answer: str) -> bool:
        """Detect nếu LLM trả về NO_CONTEXT_MESSAGE (kể cả có dấu ngoặc kép)."""
        normalized = answer.strip().strip('"').strip("'").strip()
        return normalized == NO_CONTEXT_MESSAGE

    def _build_prompt(self, question: str, chunks: list[QueryResult]) -> str:
        """Build prompt theo template đã chốt."""
        context_lines: list[str] = []
        for i, r in enumerate(chunks, start=1):
            context_lines.append(f"[{i}] {r.chunk.text}")
        context = "\n\n".join(context_lines)

        return f"""You are a code assistant. Answer the question based ONLY on the code context below.

Rules:
- If the context doesn't contain the answer, respond exactly: "{NO_CONTEXT_MESSAGE}"
- Answer concisely in 2-4 sentences.
- Use the same language as the question.
- Reference the code by name when relevant.

Code context:
{context}

Question: {question}

Answer:"""

    def _append_sources(self, answer: str, chunks: list[QueryResult]) -> str:
        """Append danh sách sources vào cuối answer."""
        lines = [answer.rstrip(), "", "Sources:"]
        for i, r in enumerate(chunks, start=1):
            meta = r.chunk.metadata
            name = meta.get("name", "?")
            entity_type = meta.get("entity_type", r.chunk.source_type)
            line_start = meta.get("line_start", "?")
            lines.append(f"[{i}] {name} ({entity_type}) - {r.chunk.source_path}:{line_start}")
        return "\n".join(lines)