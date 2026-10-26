"""Entry point cho AI gRPC server."""

import logging
import sys

from src.grpc.server import serve

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s [%(levelname)s] %(name)s: %(message)s",
    stream=sys.stdout,
)


def main() -> None:
    port = 50052
    serve(port=port)


if __name__ == "__main__":
    main()