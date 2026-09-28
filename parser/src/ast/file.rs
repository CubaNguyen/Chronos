use serde::Serialize;

#[derive(Debug, Clone, Serialize)]
pub struct FileInfo {
    pub path: String,
    pub language: String,
    pub hash: String,
    pub size_bytes: usize,
    pub line_count: usize,
}