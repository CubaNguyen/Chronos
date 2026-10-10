# ai/src/usage.py

"""Track cumulative LLM usage (tokens, cost) trong 1 session."""

from dataclasses import dataclass, field
from threading import Lock
from typing import Any


# ─── Pricing (USD per 1M tokens) ─────────────────────────
# Cập nhật khi provider đổi giá.
# Gemini free tier = $0, nhưng để estimate nếu upgrade lên paid.
PRICING: dict[str, dict[str, float]] = {
    # Gemini (Google AI Studio)
    "gemini/gemini-flash-lite-latest": {"input": 0.075, "output": 0.30},
    "gemini/gemini-flash-latest":      {"input": 0.075, "output": 0.30},
    "gemini/gemini-3.5-flash":         {"input": 0.075, "output": 0.30},
    "gemini/gemini-3.5-flash-lite":    {"input": 0.075, "output": 0.30},
    "gemini/gemini-3.8-flash":         {"input": 0.10,  "output": 0.40},
    "gemini/gemini-2.5-pro":           {"input": 1.25,  "output": 5.00},
    # Groq
    "groq/llama-3.1-70b-versatile":    {"input": 0.59,  "output": 0.79},
    "groq/llama-3.1-8b-instant":       {"input": 0.05,  "output": 0.08},
    # OpenAI
    "gpt-4o-mini":                     {"input": 0.15,  "output": 0.60},
    "gpt-4o":                          {"input": 2.50,  "output": 10.00},
}


@dataclass
class UsageTracker:
    """Track cumulative LLM usage. Thread-safe."""

    total_calls: int = 0
    total_input_tokens: int = 0
    total_output_tokens: int = 0
    total_cost_usd: float = 0.0
    _lock: Lock = field(default_factory=Lock)

    def record(self, model: str, usage: Any) -> None:
        """Record 1 LLM call. `usage` là object từ litellm response."""
        if usage is None:
            return

        with self._lock:
            self.total_calls += 1
            self.total_input_tokens += getattr(usage, "prompt_tokens", 0) or 0
            self.total_output_tokens += getattr(usage, "completion_tokens", 0) or 0
            self.total_cost_usd += self._estimate_cost(model, usage)

    def _estimate_cost(self, model: str, usage: Any) -> float:
        """Estimate cost in USD từ pricing table."""
        if model not in PRICING:
            return 0.0

        price = PRICING[model]
        input_tokens = getattr(usage, "prompt_tokens", 0) or 0
        output_tokens = getattr(usage, "completion_tokens", 0) or 0

        return (
            input_tokens * price["input"] / 1_000_000
            + output_tokens * price["output"] / 1_000_000
        )

    def summary(self) -> str:
        """Human-readable summary."""
        return (
            f"Calls: {self.total_calls} | "
            f"Tokens: {self.total_input_tokens} in + "
            f"{self.total_output_tokens} out | "
            f"Cost: ${self.total_cost_usd:.6f}"
        )

    def reset(self) -> None:
        """Reset counters."""
        with self._lock:
            self.total_calls = 0
            self.total_input_tokens = 0
            self.total_output_tokens = 0
            self.total_cost_usd = 0.0