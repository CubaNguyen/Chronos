use crate::detector::Language;
use crate::error::{ParserError, Result};
use crate::lang::interface::LanguageParser;
use crate::lang::providers::go::GoParser;
use crate::lang::providers::java::JavaParser;
use crate::lang::providers::python::PythonParser;

/// Nhận `Language`, trả về parser tương ứng.
///
/// Phase 1 chỉ Go được implement thật.
/// Java/Python trả về parser nhưng gọi `parse()` sẽ lỗi `Unimplemented`.
/// Rust/TypeScript chưa có provider → trả lỗi ngay từ factory.
pub fn create_parser(lang: Language) -> Result<Box<dyn LanguageParser>> {
    match lang {
        Language::Go => Ok(Box::new(GoParser)),
        Language::Java => Ok(Box::new(JavaParser)),
        Language::Python => Ok(Box::new(PythonParser)),
        Language::Rust => Err(ParserError::UnsupportedLanguage("rust".to_string())),
        Language::TypeScript => Err(ParserError::UnsupportedLanguage("typescript".to_string())),
    }
}