use crate::ast::Ast;
use crate::detector::Language;
use crate::error::Result;

/// Mọi parser cho một ngôn ngữ cụ thể phải implement trait này.
/// Trait chỉ có 2 method:
/// - `language()`: parser này xử lý ngôn ngữ nào (dùng cho logging/debug).
/// - `parse()`: đọc source code, trả về Ast chuẩn hóa.
pub trait LanguageParser: Send + Sync {
    fn language(&self) -> Language;
    fn parse(&self, source: &str, path: &str) -> Result<Ast>;
}