use serde::Serialize;
use super::class::Visibility;

#[derive(Debug, Clone, Serialize)]
pub struct Param {
    pub name: String,
    pub type_name: String,
}

#[derive(Debug, Clone, Serialize)]
pub struct Function {
    pub name: String,
    pub class_name: Option<String>,
    pub params: Vec<Param>,
    pub return_type: Option<String>,
    pub visibility: Visibility,
    pub line_start: usize,
    pub line_end: usize,
    pub file_path: String,
}