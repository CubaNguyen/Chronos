use crate::ast::Ast;
use crate::detector::Language;
use crate::error::{ParserError, Result};
use crate::lang::interface::LanguageParser;

pub struct PythonParser;

impl LanguageParser for PythonParser {
    fn language(&self) -> Language {
        Language::Python
    }

    fn parse(&self, _source: &str, _path: &str) -> Result<Ast> {
        Err(ParserError::UnsupportedLanguage("python".to_string()))
    }
}