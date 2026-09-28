pub mod ast;
pub mod detector;
pub mod error;

pub use ast::Ast;
pub use detector::{detect, Language};
pub use error::{ParserError, Result};