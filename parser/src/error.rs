use thiserror::Error;

#[derive(Error, Debug)]
pub enum ParserError {
    #[error("unsupported language for file: {0}")]
    UnsupportedLanguage(String),

    #[error("io error: {0}")]
    Io(#[from] std::io::Error),

    #[error("tree-sitter failed to parse: {0}")]
    TreeSitter(String),

    #[error("invalid utf-8 in file: {0}")]
    InvalidUtf8(String),
}

pub type Result<T> = std::result::Result<T, ParserError>;