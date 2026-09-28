use serde::Serialize;

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize)]
#[serde(rename_all = "lowercase")]
pub enum ClassKind {
    Class,
    Struct,
    Interface,
    Enum,
    Trait,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize)]
#[serde(rename_all = "lowercase")]
pub enum Visibility {
    Public,
    Private,
    Protected,
    Package,
}

#[derive(Debug, Clone, Serialize)]
pub struct Class {
    pub name: String,
    pub kind: ClassKind,
    pub visibility: Visibility,
    pub extends: Option<String>,
    pub implements: Vec<String>,
    pub line_start: usize,
    pub line_end: usize,
    pub file_path: String,
}

#[derive(Debug, Clone, Serialize)]
pub struct Field {
    pub name: String,
    pub type_name: String,
    pub visibility: Visibility,
    pub class_name: Option<String>,
    pub line: usize,
    pub file_path: String,
}