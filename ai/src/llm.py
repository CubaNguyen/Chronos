# ai/src/llm.py

import atexit
import logging
import os
import warnings
from abc import ABC, abstractmethod
from typing import Generator
from .usage import UsageTracker

import litellm
from dotenv import load_dotenv

load_dotenv()

logger = logging.getLogger(__name__)

os.environ["LITELLM_LOG"] = "ERROR"
os.environ["LITELLM_DISABLE_LOGGING_WORKER"] = "True"

litellm.suppress_debug_info = True
litellm.turn_off_message_logging = True
litellm.drop_params = True
litellm.success_callback = []
litellm.failure_callback = []

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

def _cleanup_litellm() -> None:
    try:
        from litellm.litellm_core_utils.logging_worker import shutdown_logging_worker
        shutdown_logging_worker()
    except Exception:
        pass

atexit.register(_cleanup_litellm)

DEFAULT_MODEL = os.getenv(
    "CHRONOS_LLM_MODEL",
    "gemini/gemini-flash-lite-latest",
)

_env_fallbacks = os.getenv("CHRONOS_LLM_FALLBACKS", "").strip()
if _env_fallbacks:
    _raw_fallbacks = [f.strip() for f in _env_fallbacks.split(",") if f.strip()]
else:
    _raw_fallbacks = [DEFAULT_MODEL]


def _filter_valid_fallbacks(models: list[str]) -> list[str]:
    valid: list[str] = []
    for m in models:
        if m.startswith("groq/") and not os.getenv("GROQ_API_KEY"):
            continue
        if m.startswith("openai/") and not os.getenv("OPENAI_API_KEY"):
            continue
        if m.startswith("anthropic/") and not os.getenv("ANTHROPIC_API_KEY"):
            continue
        valid.append(m)
    return valid

DEFAULT_FALLBACKS = _filter_valid_fallbacks(_raw_fallbacks)


class LLMClient(ABC):
    """Abstract interface cho LLM providers."""

    @abstractmethod
    def generate(self, prompt: str) -> str:
        ...

    @abstractmethod
    def generate_stream(self, prompt: str) -> Generator[str, None, None]:
        ...

    @property
    @abstractmethod
    def model_name(self) -> str:
        ...


class LiteLLMClient(LLMClient):
    """Unified client qua LiteLLM — Gemini, Groq, OpenAI, Ollama."""

    def __init__(
        self,
        model: str = DEFAULT_MODEL,
        fallback_models: list[str] | None = None,
        temperature: float = 0.2,
        max_tokens: int = 1024,
        usage_tracker: UsageTracker | None = None,
    ) -> None:
        self._model = model
        if fallback_models is None:
            self.fallback_models = DEFAULT_FALLBACKS
        else:
            self.fallback_models = fallback_models
        self.temperature = temperature
        self.max_tokens = max_tokens
        self.usage_tracker = usage_tracker or UsageTracker()
        self._verify_credentials()

    @property
    def model_name(self) -> str:
        return self._model

    def generate(self, prompt: str) -> str:
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

        if hasattr(response, "usage") and response.usage:
            self.usage_tracker.record(self._model, response.usage)

        return content.strip()

    def generate_stream(self, prompt: str) -> Generator[str, None, None]:
        logger.debug("LLM Stream call: model=%s, prompt_len=%d", self._model, len(prompt))

        try:
            # 🚨 NO fallbacks here. It breaks sync streams in LiteLLM 1.104.0
            response = litellm.completion(
                model=self._model,
                messages=[{"role": "user", "content": prompt}],
                temperature=self.temperature,
                max_tokens=self.max_tokens,
                num_retries=3,
                stream=True, 
            )
        except Exception as e:
            logger.error("LLM stream call failed: %s", e)
            raise

        for chunk in response:
            content = chunk.choices[0].delta.content
            if content:
                yield content

    def _verify_credentials(self) -> None:
        if self._model.startswith("gemini/") and not os.getenv("GEMINI_API_KEY"):
            raise RuntimeError("GEMINI_API_KEY not set.")
        elif self._model.startswith("groq/") and not os.getenv("GROQ_API_KEY"):
            raise RuntimeError("GROQ_API_KEY not set")
        elif self._model.startswith("openai/") and not os.getenv("OPENAI_API_KEY"):
            raise RuntimeError("OPENAI_API_KEY not set")