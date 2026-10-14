// parser/src/bin/parser_server.rs

//! gRPC server cho ParserService.
//!
//! Chạy: cargo run --bin parser-server
//! Port: 50051 (default, override qua env CHRONOS_PARSER_PORT)

use std::net::SocketAddr;

use tonic::transport::Server;
use tonic_reflection::server::Builder as ReflectionBuilder;
use tracing::info;
use tracing_subscriber::EnvFilter;

use chronos_parser::grpc::parser::parser_service_server::ParserServiceServer;
use chronos_parser::grpc::service::ParserServiceImpl;

const DEFAULT_PORT: u16 = 50051;

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    // Setup logging — RUST_LOG=debug để verbose
    tracing_subscriber::fmt()
        .with_env_filter(
            EnvFilter::try_from_default_env().unwrap_or_else(|_| EnvFilter::new("info")),
        )
        .with_target(false)
        .init();

    let port = std::env::var("CHRONOS_PARSER_PORT")
        .ok()
        .and_then(|s| s.parse::<u16>().ok())
        .unwrap_or(DEFAULT_PORT);

    let addr = SocketAddr::from(([0, 0, 0, 0], port));
    let service = ParserServiceImpl;

    // ─── Reflection service ───────────────────────────────
    // Cho phép grpcurl query service definitions (list/describe).
    // Đọc FileDescriptorSet đã sinh bởi build.rs.
    let reflection_service = ReflectionBuilder::configure()
        .register_encoded_file_descriptor_set(
            tonic::include_file_descriptor_set!("chronos_parser_descriptor"),
        )
        .build_v1()?;

    info!("Chronos parser gRPC server listening on {}", addr);

    Server::builder()
        .add_service(reflection_service)
        .add_service(ParserServiceServer::new(service))
        .serve_with_shutdown(addr, shutdown_signal())
        .await?;

    info!("Server shut down cleanly");
    Ok(())
}

/// Chờ Ctrl+C, sau đó trigger graceful shutdown.
async fn shutdown_signal() {
    tokio::signal::ctrl_c()
        .await
        .expect("failed to install Ctrl+C handler");
    info!("Shutdown signal received, stopping server...");
}