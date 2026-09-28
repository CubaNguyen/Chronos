pub mod call;
pub mod class;
pub mod file;
pub mod function;
pub mod import;

pub use call::Call;
pub use class::{Class, ClassKind, Field, Visibility};
pub use file::FileInfo;
pub use function::{Function, Param};
pub use import::Import;

use serde::Serialize;

#[derive(Debug, Serialize)]
pub struct Ast {
    pub file: FileInfo,
    pub classes: Vec<Class>,
    pub functions: Vec<Function>,
    pub fields: Vec<Field>,
    pub calls: Vec<Call>,
    pub imports: Vec<Import>,
}