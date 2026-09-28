use crate::error::{ParserError, Result};

/// Các ngôn ngữ mà parser hỗ trợ.
/// Chỉ Go được implement thật ở Phase 1. Các ngôn ngữ khác
/// được khai báo sẵn để detector/factory có thể dispatch.
#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash)]
pub enum Language {
    Go,
    Java,
    Python,
    Rust,
    TypeScript,
}

/// Nhận path file, trả về ngôn ngữ dựa trên extension.
///
/// Không đọc nội dung file — chỉ nhìn extension.
/// Trả lỗi `UnsupportedLanguage` nếu extension không nằm trong danh sách hỗ trợ.
pub fn detect(path: &str) -> Result<Language> {
    let ext = std::path::Path::new(path)
        .extension()
        .and_then(|s| s.to_str())
        .unwrap_or("");

    match ext.to_lowercase().as_str() {
        "go" => Ok(Language::Go),
        "java" => Ok(Language::Java),
        "py" => Ok(Language::Python),
        "rs" => Ok(Language::Rust),
        "ts" | "tsx" => Ok(Language::TypeScript),
        _ => Err(ParserError::UnsupportedLanguage(path.to_string())),
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn detects_go() {
        assert_eq!(detect("main.go").unwrap(), Language::Go);
        assert_eq!(detect("/path/to/user.go").unwrap(), Language::Go);
    }

    #[test]
    fn detects_java() {
        assert_eq!(detect("Main.java").unwrap(), Language::Java);
    }

    #[test]
    fn detects_python() {
        assert_eq!(detect("script.py").unwrap(), Language::Python);
    }

    #[test]
    fn detects_rust() {
        assert_eq!(detect("lib.rs").unwrap(), Language::Rust);
    }

    #[test]
    fn detects_typescript() {
        assert_eq!(detect("app.ts").unwrap(), Language::TypeScript);
        assert_eq!(detect("component.tsx").unwrap(), Language::TypeScript);
    }

    #[test]
    fn detects_uppercase_extension() {
        assert_eq!(detect("Main.GO").unwrap(), Language::Go);
        assert_eq!(detect("Main.Go").unwrap(), Language::Go);
    }

    #[test]
    fn unsupported_extension_returns_error() {
        let result = detect("README.md");
        assert!(matches!(result, Err(ParserError::UnsupportedLanguage(_))));
    }

    #[test]
    fn no_extension_returns_error() {
        assert!(detect("Makefile").is_err());
        assert!(detect("Dockerfile").is_err());
    }
}