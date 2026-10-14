// parser/src/grpc/mod.rs

pub mod convert;
pub mod service;

// Re-export generated proto code from parser.proto
// The module structure MUST match your proto package + service name
pub mod parser {
    tonic::include_proto!("chronos.parser.v1");
}

pub use service::{ParserServiceImpl, ParserServiceServer};