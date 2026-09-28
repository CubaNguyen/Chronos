use crate::ast::{Ast, FileInfo};
use crate::detector::Language;
use crate::error::Result;
use crate::lang::interface::LanguageParser;

pub struct GoParser;

impl LanguageParser for GoParser {
    fn language(&self) -> Language {
        Language::Go
    }

    fn parse(&self, source: &str, path: &str) -> Result<Ast> {
        // TODO (M5): Dùng Tree-sitter để parse thật.
        // Hiện tại chỉ trả về Ast rỗng để factory/skeleton compile được.
        let file = FileInfo {
            path: path.to_string(),
            language: "go".to_string(),
            hash: String::new(),
            size_bytes: source.len(),
            line_count: source.lines().count(),
        };

        Ok(Ast {
            file,
            classes: Vec::new(),
            functions: Vec::new(),
            fields: Vec::new(),
            calls: Vec::new(),
            imports: Vec::new(),
        })
    }
}