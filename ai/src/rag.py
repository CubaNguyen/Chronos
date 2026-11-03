# ai/src/rag.py

import logging
import time

from .llm import LLMClient
from .models import QueryResult, RAGAnswer
from .retriever import Retriever

logger = logging.getLogger(__name__)

DEFAULT_MIN_SCORE = 0.3
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

    # 🚀 STREAMING METHOD
    def query_stream(self, question: str, top_k: int = 5, contexts: list[QueryResult] | None = None):
        """
        Streaming version of the RAG pipeline.
        Yields chunks of the LLM answer as they are generated.
        """
        # 1. Retrieve relevant chunks
        if contexts is not None:
            relevant = contexts
        else:
            results = self.retriever.retrieve(question, top_k=top_k)
            relevant = [r for r in results if r.score >= self.min_score]

        # 2. Fallback if no context
        if not relevant:
            logger.info("No relevant context found for query")
            yield NO_CONTEXT_MESSAGE
            return

        # 3. Build prompt
        prompt = self._build_prompt(question, relevant)

        # 4. Call LLM with streaming enabled
        try:
            for chunk in self.llm.generate_stream(prompt): 
                if chunk:
                    yield chunk
        except Exception as e:
            logger.exception("Streaming failed")
            yield f"Error: {str(e)}"

    # BLOCKING METHOD (Giữ nguyên để dùng cho CLI nếu cần)

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
            results = self.retriever.retrieve(question, top_k=top_k)
            logger.debug("Retrieved %d chunks for query", len(results))
            relevant = [r for r in results if r.score >= self.min_score]
            logger.debug(
                "%d chunks passed min_score=%.2f", len(relevant), self.min_score
            )

        if not relevant:
            duration = time.perf_counter() - start
            logger.info("No relevant context found for query")
            return RAGAnswer(
                answer=NO_CONTEXT_MESSAGE,
                sources=[],
                query=question,
                duration_seconds=duration,
            )

        prompt = self._build_prompt(question, relevant)
        answer = self.llm.generate(prompt)

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
        normalized = answer.strip().strip('"').strip("'").strip()
        return normalized == NO_CONTEXT_MESSAGE

    def _build_prompt(self, question: str, chunks: list[QueryResult]) -> str:
        context_lines: list[str] = []
        for i, r in enumerate(chunks, start=1):
            # 🚀 THIS LINE IS MISSING IN YOUR CODE
            context_lines.append(f"[{i}] File: {r.chunk.source_path}\n{r.chunk.text}")
        context = "\n\n".join(context_lines)

        return f"""You are a code assistant. Answer the question based ONLY on the code context below.

Rules:
- If the context doesn't contain the answer, say "I don't have enough context to answer this question."
- Answer concisely in 2-4 sentences.
- Use the same language as the question.

Code context:
{context}

Question: {question}

Answer:"""

    def _append_sources(self, answer: str, chunks: list[QueryResult]) -> str:
        lines = [answer.rstrip(), "", "Sources:"]
        for i, r in enumerate(chunks, start=1):
            meta = r.chunk.metadata
            name = meta.get("name", "?")
            entity_type = meta.get("entity_type", r.chunk.source_type)
            line_start = meta.get("line_start", "?")
            lines.append(f"[{i}] {name} ({entity_type}) - {r.chunk.source_path}:{line_start}")
        return "\n".join(lines)
