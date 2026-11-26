# Build stage
FROM rust:latest AS builder

RUN apt-get update && apt-get install -y --no-install-recommends \
    protobuf-compiler \
    build-essential \
    pkg-config \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /workspace

# Copy proto schema definition
COPY proto ./proto

# Copy parser source code
COPY parser ./parser

WORKDIR /workspace/parser

# Compile release binary for parser-server
RUN cargo build --release --bin parser-server

# Production runtime stage
FROM debian:stable-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=builder /workspace/parser/target/release/parser-server /app/parser-server

ENV CHRONOS_PARSER_PORT=50051 \
    RUST_LOG=info

EXPOSE 50051

CMD ["/app/parser-server"]
