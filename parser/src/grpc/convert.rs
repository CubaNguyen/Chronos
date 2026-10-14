// parser/src/grpc/convert.rs

//! Conversion giữa domain Ast (src/ast/) và proto Ast (generated).

use crate::ast::{
    Ast, Call, Class, ClassKind, Field, FileInfo, Function, Import, Param, Visibility,
};
use crate::grpc::parser as pb;

// ─────────────────────────────────────────────────────────────
// DOMAIN → PROTO
// ─────────────────────────────────────────────────────────────

pub fn ast_to_proto(ast: Ast) -> pb::Ast {
    pb::Ast {
        file: Some(file_to_proto(ast.file)),
        classes: ast.classes.into_iter().map(class_to_proto).collect(),
        functions: ast.functions.into_iter().map(function_to_proto).collect(),
        fields: ast.fields.into_iter().map(field_to_proto).collect(),
        calls: ast.calls.into_iter().map(call_to_proto).collect(),
        imports: ast.imports.into_iter().map(import_to_proto).collect(),
    }
}

fn file_to_proto(f: FileInfo) -> pb::FileInfo {
    pb::FileInfo {
        path: f.path,
        language: f.language,
        hash: f.hash,
        size_bytes: f.size_bytes as i64,
        line_count: f.line_count as i32,
    }
}

fn class_to_proto(c: Class) -> pb::Class {
    pb::Class {
        name: c.name,
        kind: class_kind_to_proto(c.kind) as i32,
        visibility: visibility_to_proto(c.visibility) as i32,
        extends: c.extends,
        implements: c.implements,
        line_start: c.line_start as i32,
        line_end: c.line_end as i32,
        file_path: c.file_path,
    }
}

fn function_to_proto(f: Function) -> pb::Function {
    pb::Function {
        name: f.name,
        class_name: f.class_name,
        params: f.params.into_iter().map(param_to_proto).collect(),
        return_type: f.return_type,
        visibility: visibility_to_proto(f.visibility) as i32,
        line_start: f.line_start as i32,
        line_end: f.line_end as i32,
        file_path: f.file_path,
    }
}

fn param_to_proto(p: Param) -> pb::Param {
    pb::Param {
        name: p.name,
        type_name: p.type_name,
    }
}

fn field_to_proto(f: Field) -> pb::Field {
    pb::Field {
        name: f.name,
        type_name: f.type_name,
        visibility: visibility_to_proto(f.visibility) as i32,
        class_name: f.class_name,
        line: f.line as i32,
        file_path: f.file_path,
    }
}

fn call_to_proto(c: Call) -> pb::Call {
    pb::Call {
        caller: c.caller,
        callee: c.callee,
        line: c.line as i32,
        file_path: c.file_path,
    }
}

fn import_to_proto(i: Import) -> pb::Import {
    pb::Import {
        module_path: i.module_path,
        alias: i.alias,
        line: i.line as i32,
        is_external: i.is_external,
        file_path: i.file_path,
    }
}

fn class_kind_to_proto(k: ClassKind) -> pb::ClassKind {
    match k {
        ClassKind::Class => pb::ClassKind::Class,
        ClassKind::Struct => pb::ClassKind::Struct,
        ClassKind::Interface => pb::ClassKind::Interface,
        ClassKind::Enum => pb::ClassKind::Enum,
        ClassKind::Trait => pb::ClassKind::Trait,
    }
}

fn visibility_to_proto(v: Visibility) -> pb::Visibility {
    match v {
        Visibility::Public => pb::Visibility::Public,
        Visibility::Private => pb::Visibility::Private,
        Visibility::Protected => pb::Visibility::Protected,
        Visibility::Package => pb::Visibility::Package,
    }
}