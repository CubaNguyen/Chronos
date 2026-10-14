// parser/src/grpc/service.rs

//! Implementation của ParserService gRPC trait.

use tonic::{Request, Response, Status};

use crate::detector::detect;
use crate::grpc::convert::ast_to_proto;
use crate::grpc::parser::{
    parser_service_server::{ParserService, ParserServiceServer as TonicParserServiceServer},
    ParseFileRequest, ParseFileResponse,
};
use crate::lang::factory::create_parser;

pub struct ParserServiceImpl;

#[tonic::async_trait]
impl ParserService for ParserServiceImpl {
    async fn parse_file(
        &self,
        request: Request<ParseFileRequest>,
    ) -> Result<Response<ParseFileResponse>, Status> {
        let req = request.into_inner();

        // Validate input
        if req.path.is_empty() {
            return Err(Status::invalid_argument("path is empty"));
        }
        if req.source.is_empty() {
            return Err(Status::invalid_argument("source is empty"));
        }

        // Detect language từ path
        let lang = match detect(&req.path) {
            Ok(l) => l,
            Err(e) => {
                // Unsupported language → trả response với error field
                // (không phải gRPC error, vì đây là lỗi nghiệp vụ)
                return Ok(Response::new(ParseFileResponse {
                    ast: None,
                    error: format!("unsupported language: {}", e),
                }));
            }
        };

        // Create parser và parse
        let parser = match create_parser(lang) {
            Ok(p) => p,
            Err(e) => {
                return Ok(Response::new(ParseFileResponse {
                    ast: None,
                    error: format!("failed to create parser: {}", e),
                }));
            }
        };

        let ast = match parser.parse(&req.source, &req.path) {
            Ok(a) => a,
            Err(e) => {
                // Parse fail (syntax error...) → trả response với error
                return Ok(Response::new(ParseFileResponse {
                    ast: None,
                    error: format!("parse error: {}", e),
                }));
            }
        };

        // Convert domain Ast → proto Ast
        let proto_ast = ast_to_proto(ast);

        Ok(Response::new(ParseFileResponse {
            ast: Some(proto_ast),
            error: String::new(),
        }))
    }
}

/// Type alias để bin/parser_server.rs dùng.
pub type ParserServiceServer = TonicParserServiceServer<ParserServiceImpl>;