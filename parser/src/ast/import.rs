use serde::Serialize;

#[derive(Debug, Clone, Serialize)]
pub struct Import {
    pub module_path: String,
    pub alias: Option<String>,
    pub line: usize,
    pub is_external: bool,
}