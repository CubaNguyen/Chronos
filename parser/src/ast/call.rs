use serde::Serialize;

#[derive(Debug, Clone, Serialize)]
pub struct Call {
    pub caller: String,
    pub callee: String,
    pub line: usize,
    pub file_path: String,
}