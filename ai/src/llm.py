# ai/src/llm.py

import atexit
import logging
import os
import warnings
from abc import ABC, abstractmethod
from .usage import UsageTracker

import litellm
from dotenv import load_dotenv

load_dotenv()

# ─── Logger ──────────────────────────────────────────────
logger = logging.getLogger(__name__)

# ─── Suppress LiteLLM noise ──────────────────────────────
os.environ["LITELLM_LOG"] = "ERROR"
os.environ["LITELLM_DISABLE_LOGGING_WORKER"] = "True"

litellm.suppress_debug_info = True
litellm.turn_off_message_logging = True
litellm.drop_params = True
litellm.success_callback = []
litellm.failure_callback = []

# Suppress loggers
for name in [
    "LiteLLM",
    "litellm",
    "LiteLLM.vertex_and_google_ai_studio_gemini",
    "asyncio",
]:
    logging.getLogger(name).setLevel(logging.CRITICAL)
    logging.getLogger(name).propagate = False

warnings.filterwarnings("ignore", category=DeprecationWarning, module="litellm")
warnings.filterwarnings("ignore", category=RuntimeWarning, module="litellm")


# ─── Cleanup on exit ─────────────────────────────────────
def _cleanup_litellm() -> None:
    """Cancel LiteLLM background async tasks before exit."""
    try:
        from litellm.litellm_core_utils.logging_worker import shutdown_logging_worker
        shutdown_logging_worker()
    except Exception:
        pass

atexit.register(_cleanup_litellm)


# ─── Defaults từ env ─────────────────────────────────────
# Dùng alias "latest" để tránh Google deprecate model.
# Nếu muốn model cụ thể: gemini/gemini-3.5-flash-lite, gemini/gemini-2.5-pro...
DEFAULT_MODEL = os.getenv(
    "CHRONOS_LLM_MODEL",
    "gemini/gemini-flash-lite-latest",
)

# Đọc fallbacks từ env (comma-separated).
# Nếu không set → dùng chính DEFAULT_MODEL làm fallback (an toàn nhất).
_env_fallbacks = os.getenv("CHRONOS_LLM_FALLBACKS", "").strip()
if _env_fallbacks:
    _raw_fallbacks = [f.strip() for f in _env_fallbacks.split(",") if f.strip()]
else:
    _raw_fallbacks = [DEFAULT_MODEL]


def _filter_valid_fallbacks(models: list[str]) -> list[str]:
    """Loại bỏ fallback không có API key tương ứng."""
    valid: list[str] = []
    for m in models:
        if m.startswith("groq/") and not os.getenv("GROQ_API_KEY"):
            logger.warning("Skipping Groq fallback — GROQ_API_KEY not set")
            continue
        if m.startswith("openai/") and not os.getenv("OPENAI_API_KEY"):
            logger.warning("Skipping OpenAI fallback — OPENAI_API_KEY not set")
            continue
        if m.startswith("anthropic/") and not os.getenv("ANTHROPIC_API_KEY"):
            logger.warning("Skipping Anthropic fallback — ANTHROPIC_API_KEY not set")
            continue
        valid.append(m)
    return valid


DEFAULT_FALLBACKS = _filter_valid_fallbacks(_raw_fallbacks)


# ─── Interface ────────────────────────────────────────────
class LLMClient(ABC):
    """Abstract interface cho LLM providers."""

    @abstractmethod
    def generate(self, prompt: str) -> str:
        """Nhận prompt, trả về text response."""
        ...

    @property
    @abstractmethod
    def model_name(self) -> str:
        ...


# ─── LiteLLM implementation ───────────────────────────────
class LiteLLMClient(LLMClient):
    """Unified client qua LiteLLM — Gemini, Groq, OpenAI, Ollama."""

    def __init__(
        self,
        model: str = DEFAULT_MODEL,
        fallback_models: list[str] | None = None,
        temperature: float = 0.2,
        max_tokens: int = 1024,
        usage_tracker: UsageTracker | None = None,   # ← THÊM
    ) -> None:
        self._model = model
        # fallback_models=None → dùng DEFAULT_FALLBACKS (đã filter)
        # fallback_models=[]   → KHÔNG fallback
        if fallback_models is None:
            self.fallback_models = DEFAULT_FALLBACKS
        else:
            self.fallback_models = fallback_models
        self.temperature = temperature
        self.max_tokens = max_tokens
        self.usage_tracker = usage_tracker or UsageTracker()   # ← THÊM
        self._verify_credentials()

    @property
    def model_name(self) -> str:
        return self._model

    def generate(self, prompt: str) -> str:
        """Gọi LLM, trả về text. Tự retry + fallback."""
        logger.debug("LLM call: model=%s, prompt_len=%d", self._model, len(prompt))

        try:
            response = litellm.completion(
                model=self._model,
                messages=[{"role": "user", "content": prompt}],
                temperature=self.temperature,
                max_tokens=self.max_tokens,
                num_retries=3,
                fallbacks=self.fallback_models,
            )
        except Exception as e:
            logger.error("LLM call failed: %s", e)
            raise

        content = response.choices[0].message.content or ""

        # ─── Track usage ─────────────────────────────────
        if hasattr(response, "usage") and response.usage:
            self.usage_tracker.record(self._model, response.usage)
            usage = response.usage
            logger.info(
                "LLM usage: in=%d, out=%d, total=%d tokens",
                usage.prompt_tokens,
                usage.completion_tokens,
                usage.total_tokens,
            )

        return content.strip()

    def _verify_credentials(self) -> None:
        """Fail sớm nếu thiếu API key."""
        if self._model.startswith("gemini/") and not os.getenv("GEMINI_API_KEY"):
            raise RuntimeError(
                "GEMINI_API_KEY not set. "
                "Get one at https://aistudio.google.com/app/apikey"
            )
        elif self._model.startswith("groq/") and not os.getenv("GROQ_API_KEY"):
            raise RuntimeError("GROQ_API_KEY not set")
        elif self._model.startswith("openai/") and not os.getenv("OPENAI_API_KEY"):
            raise RuntimeError("OPENAI_API_KEY not set")