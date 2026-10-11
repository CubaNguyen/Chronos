use crate::ast::Ast;
use crate::detector::Language;
use crate::error::{ParserError, Result};
use crate::lang::interface::LanguageParser;

pub struct JavaParser;

impl LanguageParser for JavaParser {
    fn language(&self) -> Language {
        Language::Java
    }

    fn parse(&self, _source: &str, _path: &str) -> Result<Ast> {
        Err(ParserError::UnsupportedLanguage("java".to_string()))
    }
}