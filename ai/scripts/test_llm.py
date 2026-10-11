"""Manual test cho LLM client."""

from src.llm import LiteLLMClient


def main() -> None:
    client = LiteLLMClient()
    print(f"Model: {client.model_name}")
    print()

    prompt = "In one sentence, what is Rust programming language?"
    print(f"Prompt: {prompt}")
    print()

    answer = client.generate(prompt)
    print(f"Answer: {answer}")


if __name__ == "__main__":
    main()