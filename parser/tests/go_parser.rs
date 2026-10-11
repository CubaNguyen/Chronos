//! Integration tests cho GoParser.
//!
//! Mỗi test tạo source code Go inline (không dùng file testdata),
//! gọi parser, rồi assert kết quả.

use chronos_parser::ast::Ast;
use chronos_parser::detector::detect;
use chronos_parser::lang::factory::create_parser;

/// Helper: parse source code Go, trả về Ast.
/// Panic nếu detect/factory/parse fail — vì test chỉ chạy happy path.
fn parse_go(source: &str) -> Ast {
    let lang = detect("test.go").expect("detect should work for .go");
    let parser = create_parser(lang).expect("factory should create GoParser");
    parser
        .parse(source, "test.go")
        .expect("parse should succeed")
}

// ─────────────────────────────────────────────────────────────
// Đợt 1: Basic entities
// ─────────────────────────────────────────────────────────────

#[test]
fn parses_empty_file() {
    let ast = parse_go("package main\n");

    assert_eq!(ast.classes.len(), 0, "empty file should have no classes");
    assert_eq!(ast.functions.len(), 0, "empty file should have no functions");
    assert_eq!(ast.fields.len(), 0, "empty file should have no fields");
    assert_eq!(ast.calls.len(), 0, "empty file should have no calls");
    assert_eq!(ast.imports.len(), 0, "empty file should have no imports");

    assert_eq!(ast.file.language, "go");
    assert_eq!(ast.file.path, "test.go");
}

#[test]
fn parses_simple_function() {
    let src = r#"
package main

func Add(a, b int) int {
    return a + b
}
"#;
    let ast = parse_go(src);

    assert_eq!(ast.functions.len(), 1);
    let f = &ast.functions[0];
    assert_eq!(f.name, "Add");
    assert_eq!(f.class_name, None);
    assert_eq!(f.return_type, Some("int".to_string()));
    assert_eq!(f.params.len(), 2);
    assert_eq!(f.params[0].name, "a");
    assert_eq!(f.params[0].type_name, "int");
    assert_eq!(f.params[1].name, "b");
    assert_eq!(f.params[1].type_name, "int");
}

#[test]
fn parses_struct_with_fields() {
    let src = r#"
package main

type User struct {
    ID    int
    Name  string
    Email string
}
"#;
    let ast = parse_go(src);

    assert_eq!(ast.classes.len(), 1);
    let c = &ast.classes[0];
    assert_eq!(c.name, "User");
    assert_eq!(c.extends, None);
    assert_eq!(c.implements.len(), 0);

    assert_eq!(ast.fields.len(), 3);
    assert_eq!(ast.fields[0].name, "ID");
    assert_eq!(ast.fields[0].type_name, "int");
    assert_eq!(ast.fields[0].class_name, Some("User".to_string()));
    assert_eq!(ast.fields[1].name, "Name");
    assert_eq!(ast.fields[1].type_name, "string");
    assert_eq!(ast.fields[2].name, "Email");
    assert_eq!(ast.fields[2].type_name, "string");
}

#[test]
fn parses_interface() {
    let src = r#"
package main

type Reader interface {
    Read(p []byte) (int, error)
}
"#;
    let ast = parse_go(src);

    assert_eq!(ast.classes.len(), 1);
    let c = &ast.classes[0];
    assert_eq!(c.name, "Reader");
    assert_eq!(c.implements.len(), 0);

    // Interface methods hiện tại chưa extract — sẽ làm ở M8
    // Test này chỉ verify interface được nhận diện là class.
}

// ─────────────────────────────────────────────────────────────
// Đợt 2: Methods, imports, variadic, calls
// ─────────────────────────────────────────────────────────────

#[test]
fn parses_method_with_pointer_receiver() {
    let src = r#"
package main

type UserService struct {
    db string
}

func (s *UserService) GetUser(id int) string {
    return ""
}
"#;
    let ast = parse_go(src);

    assert_eq!(ast.classes.len(), 1);
    assert_eq!(ast.functions.len(), 1);

    let m = &ast.functions[0];
    assert_eq!(m.name, "GetUser");
    assert_eq!(m.class_name, Some("UserService".to_string()));
    assert_eq!(m.params.len(), 1);
    assert_eq!(m.params[0].name, "id");
    assert_eq!(m.params[0].type_name, "int");
    assert_eq!(m.return_type, Some("string".to_string()));
}

#[test]
fn parses_method_with_value_receiver() {
    let src = r#"
package main

type Counter struct {
    n int
}

func (c Counter) Value() int {
    return c.n
}
"#;
    let ast = parse_go(src);

    assert_eq!(ast.functions.len(), 1);
    let m = &ast.functions[0];
    assert_eq!(m.name, "Value");
    // Value receiver (không có *) vẫn phải extract đúng class_name
    assert_eq!(m.class_name, Some("Counter".to_string()));
}

#[test]
fn parses_imports_with_alias_and_external() {
    let src = r#"
package main

import (
    "fmt"
    "database/sql"
    mycache "github.com/myorg/cache"
)
"#;
    let ast = parse_go(src);

    assert_eq!(ast.imports.len(), 3);

    // Import 1: stdlib, không alias
    assert_eq!(ast.imports[0].module_path, "fmt");
    assert_eq!(ast.imports[0].alias, None);
    assert_eq!(ast.imports[0].is_external, false);

    // Import 2: stdlib, không alias
    assert_eq!(ast.imports[1].module_path, "database/sql");
    assert_eq!(ast.imports[1].is_external, false);

    // Import 3: external, có alias
    assert_eq!(ast.imports[2].module_path, "github.com/myorg/cache");
    assert_eq!(ast.imports[2].alias, Some("mycache".to_string()));
    assert_eq!(ast.imports[2].is_external, true);
}

#[test]
fn parses_variadic_parameter() {
    let src = r#"
package main

func Sum(nums ...int) int {
    return 0
}
"#;
    let ast = parse_go(src);

    assert_eq!(ast.functions.len(), 1);
    let f = &ast.functions[0];
    assert_eq!(f.params.len(), 1);
    assert_eq!(f.params[0].name, "nums");
    assert_eq!(f.params[0].type_name, "...int");
}

#[test]
fn parses_function_calls() {
    let src = r#"
package main

func Outer() {
    Inner()
    fmt.Println("hi")
}

func Inner() {}
"#;
    let ast = parse_go(src);

    assert_eq!(ast.functions.len(), 2);

    // 2 calls: Inner() và fmt.Println()
    assert_eq!(ast.calls.len(), 2);

    let call1 = &ast.calls[0];
    assert_eq!(call1.caller, "Outer");
    assert_eq!(call1.callee, "Inner");

    let call2 = &ast.calls[1];
    assert_eq!(call2.caller, "Outer");
    assert_eq!(call2.callee, "fmt.Println");
}

// ─────────────────────────────────────────────────────────────
// Đợt 3: Edge cases
// ─────────────────────────────────────────────────────────────

#[test]
fn skips_closure_call_but_keeps_inner_calls() {
    let src = r#"
package main

func Outer() {
    Inner()
    func() {
        DeepNested()
    }()
}

func Inner() {}
func DeepNested() {}
"#;
    let ast = parse_go(src);

    // Có 3 functions: Outer, Inner, DeepNested
    assert_eq!(ast.functions.len(), 3);

    // Calls: chỉ có 2, closure call bị skip.
    // Inner (trực tiếp) và DeepNested (bên trong closure, gán caller = Outer).
    assert_eq!(ast.calls.len(), 2, "closure call itself should be skipped");

    let calls: Vec<(&str, &str)> = ast
        .calls
        .iter()
        .map(|c| (c.caller.as_str(), c.callee.as_str()))
        .collect();

    assert!(calls.contains(&("Outer", "Inner")));
    assert!(calls.contains(&("Outer", "DeepNested")));

    // Đảm bảo không có call nào có callee chứa "func" (closure body)
    for c in &ast.calls {
        assert!(
            !c.callee.contains("func"),
            "callee should not be a closure body, got: {}",
            c.callee
        );
    }
}

#[test]
fn parses_embedded_struct() {
    let src = r#"
package main

type Base struct {
    ID int
}

type Derived struct {
    Base
    Name string
}
"#;
    let ast = parse_go(src);

    assert_eq!(ast.classes.len(), 2);

    let base = ast.classes.iter().find(|c| c.name == "Base").unwrap();
    assert_eq!(base.extends, None);

    let derived = ast.classes.iter().find(|c| c.name == "Derived").unwrap();
    assert_eq!(derived.extends, Some("Base".to_string()));

    // Derived chỉ có field Name (Base là embedded, không phải field thường)
    let derived_fields: Vec<&str> = ast
        .fields
        .iter()
        .filter(|f| f.class_name.as_deref() == Some("Derived"))
        .map(|f| f.name.as_str())
        .collect();
    assert_eq!(derived_fields, vec!["Name"]);
}

#[test]
fn parses_embedded_interface() {
    let src = r#"
package main

type Reader interface {
    Read(p []byte) (int, error)
}

type Writer interface {
    Write(p []byte) (int, error)
}

type ReadWriter interface {
    Reader
    Writer
}
"#;
    let ast = parse_go(src);

    let rw = ast
        .classes
        .iter()
        .find(|c| c.name == "ReadWriter")
        .unwrap();

    assert_eq!(rw.implements.len(), 2);
    assert!(rw.implements.contains(&"Reader".to_string()));
    assert!(rw.implements.contains(&"Writer".to_string()));
}

#[test]
fn parses_generic_struct() {
    let src = r#"
package main

type Stack[T any] struct {
    items []T
}
"#;
    let ast = parse_go(src);

    assert_eq!(ast.classes.len(), 1);
    let c = &ast.classes[0];
    assert_eq!(c.name, "Stack");

    // Field `items` type `[]T` — type param `T` chưa extract (M11),
    // nhưng field vẫn phải được extract.
    assert_eq!(ast.fields.len(), 1);
    assert_eq!(ast.fields[0].name, "items");
    assert_eq!(ast.fields[0].type_name, "[]T");
}

#[test]
fn computes_consistent_hash() {
    let src1 = "package main\n\nfunc A() {}\n";
    let src2 = "package main\n\nfunc A() {}\n";
    let src3 = "package main\n\nfunc B() {}\n";

    let ast1 = parse_go(src1);
    let ast2 = parse_go(src2);
    let ast3 = parse_go(src3);

    // Cùng content → cùng hash
    assert_eq!(ast1.file.hash, ast2.file.hash);
    // Khác content → khác hash
    assert_ne!(ast1.file.hash, ast3.file.hash);
    // SHA-256 hex luôn 64 ký tự
    assert_eq!(ast1.file.hash.len(), 64);
}

#[test]
fn extracts_visibility_from_case() {
    let src = r#"
package main

type UserService struct {
    PublicField  string
    privateField string
}

func ExportedFunc() {}
func privateFunc()  {}

func (s *UserService) PublicMethod()  {}
func (s *UserService) privateMethod() {}
"#;
    let ast = parse_go(src);

    // Field visibility
    let pub_field = ast.fields.iter().find(|f| f.name == "PublicField").unwrap();
    let priv_field = ast.fields.iter().find(|f| f.name == "privateField").unwrap();
    assert_eq!(pub_field.visibility, chronos_parser::ast::Visibility::Public);
    assert_eq!(priv_field.visibility, chronos_parser::ast::Visibility::Private);

    // Function visibility
    let exp_fn = ast.functions.iter().find(|f| f.name == "ExportedFunc").unwrap();
    let priv_fn = ast.functions.iter().find(|f| f.name == "privateFunc").unwrap();
    assert_eq!(exp_fn.visibility, chronos_parser::ast::Visibility::Public);
    assert_eq!(priv_fn.visibility, chronos_parser::ast::Visibility::Private);

    // Method visibility
    let pub_m = ast.functions.iter().find(|f| f.name == "PublicMethod").unwrap();
    let priv_m = ast.functions.iter().find(|f| f.name == "privateMethod").unwrap();
    assert_eq!(pub_m.visibility, chronos_parser::ast::Visibility::Public);
    assert_eq!(priv_m.visibility, chronos_parser::ast::Visibility::Private);
}