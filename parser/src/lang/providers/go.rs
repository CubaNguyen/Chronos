use tree_sitter::{Node, Parser};
use sha2::{Digest, Sha256};
use crate::ast::{
    Ast, Call, Class, ClassKind, Field, FileInfo, Function, Import, Param, Visibility,
};
use crate::detector::Language;
use crate::error::{ParserError, Result};
use crate::lang::interface::LanguageParser;

pub struct GoParser;

impl LanguageParser for GoParser {
    fn language(&self) -> Language {
        Language::Go
    }

    fn parse(&self, source: &str, path: &str) -> Result<Ast> {
        let mut parser = Parser::new();
        let language: tree_sitter::Language = tree_sitter_go::LANGUAGE.into();
        parser
            .set_language(&language)
            .map_err(|e| ParserError::TreeSitter(e.to_string()))?;

        let tree = parser
            .parse(source, None)
            .ok_or_else(|| ParserError::TreeSitter("parse returned None".to_string()))?;

        let root = tree.root_node();

        let mut hasher = Sha256::new();
        hasher.update(source.as_bytes());
        let hash = hex::encode(hasher.finalize());

        let file = FileInfo {
            path: path.to_string(),
            language: "go".to_string(),
            hash,
            size_bytes: source.len(),
            line_count: source.lines().count(),
        };

        let mut ast = Ast {
            file,
            classes: Vec::new(),
            functions: Vec::new(),
            fields: Vec::new(),
            calls: Vec::new(),
            imports: Vec::new(),
        };

        // Walk top-level children của source_file
        let mut cursor = root.walk();
        for child in root.named_children(&mut cursor) {
            match child.kind() {
                "import_declaration" => extract_imports(child, source, path, &mut ast),
                "type_declaration" => extract_types(child, source, path, &mut ast),
                "function_declaration" => extract_function(child, None, source, path, &mut ast),
                "method_declaration" => extract_method(child, source, path, &mut ast),
                _ => {}
            }
        }

        Ok(ast)
    }
}

// ─────────────────────────────────────────────────────────────
// HELPERS
// ─────────────────────────────────────────────────────────────

fn node_text(node: Node, source: &str) -> String {
    node.utf8_text(source.as_bytes())
        .unwrap_or("")
        .to_string()
}

/// Go quy ước: identifier bắt đầu bằng chữ hoa = public, chữ thường = private.
fn go_visibility(name: &str) -> Visibility {
    match name.chars().next() {
        Some(c) if c.is_uppercase() => Visibility::Public,
        _ => Visibility::Private,
    }
}

// ─────────────────────────────────────────────────────────────
// M5.1 — IMPORTS
// ─────────────────────────────────────────────────────────────

fn extract_imports(node: Node, source: &str, path: &str, ast: &mut Ast) {
    let mut cursor = node.walk();
    for child in node.named_children(&mut cursor) {
        match child.kind() {
            "import_spec" => process_import_spec(child, source, path, ast),
            "import_spec_list" => {
                let mut inner = child.walk();
                for spec in child.named_children(&mut inner) {
                    if spec.kind() == "import_spec" {
                        process_import_spec(spec, source, path, ast);
                    }
                }
            }
            _ => {}
        }
    }
}

fn process_import_spec(node: Node, source: &str, path: &str, ast: &mut Ast) {
    let raw_path = node
        .child_by_field_name("path")
        .map(|n| node_text(n, source))
        .unwrap_or_default();

    // Bỏ dấu nháy kép hoặc backtick bao quanh
    let module_path = raw_path
        .trim_matches('"')
        .trim_matches('`')
        .to_string();

    let alias = node
        .child_by_field_name("name")
        .map(|n| node_text(n, source));

    // Heuristic: module path có dấu chấm ở segment đầu = external
    // vd: "fmt" -> stdlib, "github.com/x/y" -> external
    let is_external = module_path
        .split('/')
        .next()
        .map(|s| s.contains('.'))
        .unwrap_or(false);

    ast.imports.push(Import {
        module_path,
        alias,
        line: node.start_position().row + 1,
        is_external,
        file_path: path.to_string(),
    });
}

// ─────────────────────────────────────────────────────────────
// M5.2 + M5.3 — TYPES (struct, interface) + FIELDS
// ─────────────────────────────────────────────────────────────

fn extract_types(node: Node, source: &str, path: &str, ast: &mut Ast) {
    let mut cursor = node.walk();
    for spec in node.named_children(&mut cursor) {
        if spec.kind() != "type_spec" {
            continue;
        }

        let name = spec
            .child_by_field_name("name")
            .map(|n| node_text(n, source))
            .unwrap_or_default();

        let Some(type_node) = spec.child_by_field_name("type") else {
            continue;
        };

        let kind = match type_node.kind() {
            "struct_type" => ClassKind::Struct,
            "interface_type" => ClassKind::Interface,
            _ => continue, // bỏ qua alias type, map type, v.v.
        };

        let class = Class {
            name: name.clone(),
            kind,
            visibility: go_visibility(&name),
            extends: None,
            implements: Vec::new(),
            line_start: spec.start_position().row + 1,
            line_end: spec.end_position().row + 1,
            file_path: path.to_string(),
        };
        ast.classes.push(class);

        match kind {
            ClassKind::Struct => {
                extract_struct_fields(type_node, &name, source, path, ast);
                extract_struct_embeds(type_node, &name, source, ast);
            }
            ClassKind::Interface => {
                extract_interface_embeds(type_node, &name, source, ast);
            }
            _ => {}
        }
    }
}

fn extract_struct_fields(
    struct_node: Node,
    class_name: &str,
    source: &str,
    path: &str,
    ast: &mut Ast,
) {
    let mut cursor = struct_node.walk();
    for child in struct_node.named_children(&mut cursor) {
        if child.kind() != "field_declaration_list" {
            continue;
        }

        let mut inner = child.walk();
        for field in child.named_children(&mut inner) {
            if field.kind() == "field_declaration" {
                process_field(field, class_name, source, path, ast);
            }
        }
    }
}

fn process_field(node: Node, class_name: &str, source: &str, path: &str, ast: &mut Ast) {
    // Go cho phép `a, b int` — nhiều field cùng type
    let mut names: Vec<String> = Vec::new();
    let mut type_name = String::new();

    let mut cursor = node.walk();
    for child in node.named_children(&mut cursor) {
        match child.kind() {
            "field_identifier" => names.push(node_text(child, source)),
            // Các loại type hợp lệ của Go
            "type_identifier"
            | "pointer_type"
            | "qualified_type"
            | "map_type"
            | "array_type"
            | "slice_type"
            | "struct_type"
            | "interface_type"
            | "function_type"
            | "channel_type" => {
                type_name = node_text(child, source);
            }
            _ => {}
        }
    }

    for name in names {
        ast.fields.push(Field {
            name: name.clone(),
            type_name: type_name.clone(),
            visibility: go_visibility(&name),
            class_name: Some(class_name.to_string()),
            line: node.start_position().row + 1,
            file_path: path.to_string(),
        });
    }
}
/// Embedded field trong struct → coi như "extends".
/// `type Admin struct { User; Level int }` → Admin extends User.
fn extract_struct_embeds(struct_node: Node, class_name: &str, source: &str, ast: &mut Ast) {
    let mut cursor = struct_node.walk();
    for child in struct_node.named_children(&mut cursor) {
        if child.kind() != "field_declaration_list" {
            continue;
        }

        let mut inner = child.walk();
        for field in child.named_children(&mut inner) {
            if field.kind() != "field_declaration" {
                continue;
            }

            // Embedded = không có field_identifier
            let mut has_name = false;
            let mut type_name: Option<String> = None;

            let mut f_cursor = field.walk();
            for c in field.named_children(&mut f_cursor) {
                match c.kind() {
                    "field_identifier" => has_name = true,
                    "type_identifier" => {
                        type_name = Some(node_text(c, source));
                    }
                    "pointer_type" => {
                        // *User cũng có thể là embedded
                        let mut p_cursor = c.walk();
                        for p in c.named_children(&mut p_cursor) {
                            if p.kind() == "type_identifier" {
                                type_name = Some(node_text(p, source));
                            }
                        }
                    }
                    _ => {}
                }
            }

            if !has_name {
                if let Some(t) = type_name {
                    // Tìm Class tương ứng trong ast để set `extends`
                    if let Some(class) = ast.classes.iter_mut().find(|c| c.name == class_name) {
                        // Nếu chưa có extends, set. Nếu có rồi, bỏ qua (Go chỉ cho 1 embedded chính).
                        if class.extends.is_none() {
                            class.extends = Some(t);
                        }
                    }
                }
            }
        }
    }
}

/// Embedded interface → push vào `implements`.
/// `type ReadWriter interface { Reader; Writer }` → implements [Reader, Writer].
fn extract_interface_embeds(
    interface_node: Node,
    class_name: &str,
    source: &str,
    ast: &mut Ast,
) {
    let mut embeds: Vec<String> = Vec::new();

    let mut cursor = interface_node.walk();
    for child in interface_node.named_children(&mut cursor) {
        // Trong Go, interface body là `interface_type` chứa các `type_identifier` hoặc method
        // Tree-sitter kind khác nhau tùy version. Cần handle 2 case:
        // - "type_identifier" trực tiếp (embedded interface)
        // - "method_elem" hoặc "method_spec" (method của interface — bỏ qua)
        match child.kind() {
            "type_identifier" => {
                embeds.push(node_text(child, source));
            }
            // Trong một số version tree-sitter-go, interface body là con của "method_elem_list" hoặc tương tự
            _ => {
                let mut inner = child.walk();
                for c in child.named_children(&mut inner) {
                    if c.kind() == "type_identifier" {
                        embeds.push(node_text(c, source));
                    }
                }
            }
        }
    }

    if let Some(class) = ast.classes.iter_mut().find(|c| c.name == class_name) {
        class.implements = embeds;
    }
}
// ─────────────────────────────────────────────────────────────
// M5.4 + M5.5 — FUNCTIONS + METHODS
// ─────────────────────────────────────────────────────────────

/// Free function: `func Add(a, b int) int { ... }`
fn extract_function(
    node: Node,
    class_name: Option<String>,
    source: &str,
    path: &str,
    ast: &mut Ast,
) {
    process_function_like(node, class_name, source, path, ast);
}

/// Method: `func (s *UserService) GetUser(id int) ...`
fn extract_method(node: Node, source: &str, path: &str, ast: &mut Ast) {
    let class_name = extract_receiver_class(node, source);
    process_function_like(node, class_name, source, path, ast);
}

/// Lấy tên class từ receiver của method.
/// `(s *UserService)` -> "UserService"
/// `(u UserService)`  -> "UserService"
fn extract_receiver_class(node: Node, source: &str) -> Option<String> {
    let receiver = node.child_by_field_name("receiver")?;

    // receiver là parameter_list, chứa 1 parameter_declaration
    let mut cursor = receiver.walk();
    for child in receiver.named_children(&mut cursor) {
        if child.kind() != "parameter_declaration" {
            continue;
        }

        let mut inner = child.walk();
        for c in child.named_children(&mut inner) {
            match c.kind() {
                "type_identifier" => return Some(node_text(c, source)),
                "pointer_type" => {
                    // pointer_type chứa type_identifier bên trong
                    let mut ptr_cursor = c.walk();
                    for ptr_child in c.named_children(&mut ptr_cursor) {
                        if ptr_child.kind() == "type_identifier" {
                            return Some(node_text(ptr_child, source));
                        }
                    }
                }
                _ => {}
            }
        }
    }
    None
}

/// Dùng chung cho function và method.
fn process_function_like(
    node: Node,
    class_name: Option<String>,
    source: &str,
    path: &str,
    ast: &mut Ast,
) {
    let name = node
        .child_by_field_name("name")
        .map(|n| node_text(n, source))
        .unwrap_or_default();

    let params = extract_params(node.child_by_field_name("parameters"), source);

    let return_type = node
        .child_by_field_name("result")
        .map(|n| node_text(n, source));

    let visibility = go_visibility(&name);

    let full_name = match &class_name {
        Some(c) => format!("{}.{}", c, name),
        None => name.clone(),
    };

    ast.functions.push(Function {
        name,
        class_name,
        params,
        return_type,
        visibility,
        line_start: node.start_position().row + 1,
        line_end: node.end_position().row + 1,
        file_path: path.to_string(),
    });

    // Đệ quy vào body để extract calls
    if let Some(body) = node.child_by_field_name("body") {
        extract_calls(body, &full_name, source, path, ast);
    }
}

fn extract_params(node: Option<Node>, source: &str) -> Vec<Param> {
    let mut params = Vec::new();
    let Some(node) = node else {
        return params;
    };

    let mut cursor = node.walk();
    for child in node.named_children(&mut cursor) {
        // Parameter thường: `a, b int`
        // Variadic: `nums ...int`
        let is_variadic = match child.kind() {
            "parameter_declaration" => false,
            "variadic_parameter_declaration" => true,
            _ => continue,
        };

        let mut names: Vec<String> = Vec::new();
        let mut type_name = String::new();

        let mut inner = child.walk();
        for c in child.named_children(&mut inner) {
            match c.kind() {
                "identifier" => names.push(node_text(c, source)),
                _ => type_name = node_text(c, source),
            }
        }

        // Variadic: prefix `...` vào type
        if is_variadic {
            type_name = format!("...{}", type_name);
        }

        for name in names {
            params.push(Param {
                name,
                type_name: type_name.clone(),
            });
        }
    }

    params
}

// ─────────────────────────────────────────────────────────────
// M5.6 — CALLS
// ─────────────────────────────────────────────────────────────

fn extract_calls(node: Node, caller: &str, source: &str, path: &str, ast: &mut Ast) {
    if node.kind() == "call_expression" {
        if let Some(func_node) = node.child_by_field_name("function") {
            // Skip closure calls: `func() {...}()`
            // Vì callee là nguyên body closure, vô nghĩa.
            if func_node.kind() != "func_literal" {
                ast.calls.push(Call {
                    caller: caller.to_string(),
                    callee: node_text(func_node, source),
                    line: node.start_position().row + 1,
                    file_path: path.to_string(),
                });
            }
        }
    }

    // Đệ quy xuống tất cả children
    let mut cursor = node.walk();
    for child in node.named_children(&mut cursor) {
        extract_calls(child, caller, source, path, ast);
    }
}