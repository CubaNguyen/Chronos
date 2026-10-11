"""Test nhanh các model để chọn model ổn định."""

import time

from src.llm import LiteLLMClient

MODELS = [
    "gemini/gemini-flash-lite-latest",
    "gemini/gemini-flash-latest",
    "gemini/gemini-3.5-flash",
]

PROMPT = "Say hi in 3 words."

for model in MODELS:
    print(f"\n{'=' * 50}")
    print(f"Testing: {model}")
    print("=" * 50)
    try:
        llm = LiteLLMClient(model=model, fallback_models=[])
        start = time.time()
        answer = llm.generate(PROMPT)
        duration = time.time() - start
        print(f"✅ OK ({duration:.2f}s): {answer}")
    except Exception as e:
        print(f"❌ FAIL: {type(e).__name__}")
        print(f"   {str(e)[:200]}")