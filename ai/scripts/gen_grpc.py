"""Generate Python gRPC code từ proto files.

Chạy: python scripts/gen_grpc.py
"""

import subprocess
import sys
from pathlib import Path

# Project root = ai/  (parent của scripts/)
AI_ROOT = Path(__file__).resolve().parent.parent
CHRONOS_ROOT = AI_ROOT.parent
PROTO_DIR = CHRONOS_ROOT / "proto"
OUT_DIR = AI_ROOT / "src" / "grpc"


def main() -> None:
    proto_files = list(PROTO_DIR.glob("*.proto"))
    if not proto_files:
        print(f"No .proto files found in {PROTO_DIR}", file=sys.stderr)
        sys.exit(1)

    print(f"Generating {len(proto_files)} proto file(s) -> {OUT_DIR}")

    cmd = [
        sys.executable, "-m", "grpc_tools.protoc",
        f"-I={PROTO_DIR}",
        f"--python_out={OUT_DIR}",
        f"--grpc_python_out={OUT_DIR}",
        *[str(p) for p in proto_files],
    ]

    subprocess.run(cmd, check=True)

    # Fix relative imports trong file _grpc.py
    for grpc_file in OUT_DIR.glob("*_pb2_grpc.py"):
        content = grpc_file.read_text(encoding="utf-8")
        # import ai_pb2 -> from . import ai_pb2
        fixed = content.replace(
            "import ai_pb2 as ai__pb2",
            "from . import ai_pb2 as ai__pb2"
        ).replace(
            "import parser_pb2 as parser__pb2",
            "from . import parser_pb2 as parser__pb2"
        )
        if fixed != content:
            grpc_file.write_text(fixed, encoding="utf-8")
            print(f"  Fixed import in {grpc_file.name}")

    print("Done.")


if __name__ == "__main__":
    main()