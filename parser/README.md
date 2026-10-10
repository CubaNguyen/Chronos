# 🧠 Chronos Parser

> Rust sidecar of the Chronos system. Parses source code into a normalized AST that the Go core can ingest into a Knowledge Graph.

---

## 🎯 Purpose

Chronos is a **Code Knowledge Graph** combined with AI. It answers two questions:

| # | Question | Feature |
| :-: | :--- | :--- |
| 1 | *"Why is this code like this?"* | 🔍 Context Tracing |
| 2 | *"If I change this, what breaks?"* | 💥 Impact Analysis |

This parser is the **first stage** of the pipeline: it reads source code and produces structured AST that downstream modules (Graph, Vector, AI) consume.

---

## 📊 Status

| Area | Status |
| :--- | :--- |
| 🟢 Go parser | Working (5/5 entities extracted) |
| 🔴 Java parser | Not implemented |
| 🔴 Python parser | Not implemented |
| 🔴 Rust parser | Not implemented |
| 🔴 TypeScript parser | Not implemented |
| 🟢 CLI | Working |
| 🟡 gRPC server | Not implemented (next milestone) |
| 🔴 Parallel parsing | Not implemented |
| 🔴 Interface methods | Not extracted |
| 🔴 Const/var declarations | Not extracted |
| 🔴 Generics type params | Not extracted |

> 📌 See **[TODO.md](./TODO.md)** for the full backlog.

---

## ⚙️ How it works

    File
      │
      ▼
    ┌────────────────┐
    │  detector.rs   │──▶  Language (Go / Java / Python / Rust / TS)
    └────────────────┘
      │
      ▼
    ┌────────────────┐
    │  factory.rs    │──▶  Box<dyn LanguageParser>
    └────────────────┘
      │
      ▼
    ┌─────────────────────────┐
    │  providers/go.rs        │──▶  Tree-sitter parse → walk AST → extract entities
    └─────────────────────────┘
      │
      ▼
    ┌────────────────┐
    │   ast::Ast     │──▶  normalized output (JSON)
    └────────────────┘

---

## 🌳 AST Model

The output is **normalized** — same structure regardless of source language.

    Ast
    ├── file: FileInfo           — path, language, SHA-256 hash, size, line count
    ├── classes: Vec<Class>      — struct, interface, class, enum, trait
    ├── functions: Vec<Function> — free functions + methods (distinguished by class_name)
    ├── fields: Vec<Field>       — struct/class fields (flat list, class_name points to owner)
    ├── calls: Vec<Call>         — caller → callee edges (raw names, resolved by Go core)
    └── imports: Vec<Import>     — module path, alias, is_external

**💡 Design principle:** Flat, not nested. Everything is a top-level list. Relationships are expressed via foreign-key-like fields (`class_name`, `caller`, `file_path`). This maps **1:1 to Neo4j** nodes/edges with zero transformation logic.

> 📂 See `src/ast/` for the full schema.

---

## 🖥️ CLI Usage

**Parse a single file (pretty JSON):**

    cargo run -- parse testdata/sample.go

**Parse a single file (compact JSON, single line):**

    cargo run -- parse testdata/sample.go --compact

**Parse a directory recursively (pretty JSON array):**

    cargo run -- parse testdata/

**Help:**

    cargo run -- --help
    cargo run -- parse --help

### 🎬 Behavior

- ✅ Files with unsupported extensions are **skipped silently**
- ⚠️ Files that fail to parse log a warning to stderr but **don't crash**
- 🚫 Hidden files/dirs (`.git`, `.env`) and build dirs (`target/`, `node_modules/`) are skipped
- 📤 Exit code `0` on success, `1` on error

---

## 🛠️ Development

### 📋 Prerequisites

- 🦀 Rust **1.75+** (install via [rustup](https://rustup.rs))
- 🔧 C compiler (`cc`) for compiling Tree-sitter grammars

### 🚀 Setup

    cd parser
    cargo build
    cargo test

### 📁 Project layout

    parser/
    ├── src/
    │   ├── main.rs                  — CLI entry point
    │   ├── lib.rs                   — public API
    │   ├── error.rs                 — ParserError type
    │   ├── detector.rs              — extension → Language
    │   ├── ast/                     — domain model (language-agnostic)
    │   │   ├── file.rs              — FileInfo
    │   │   ├── class.rs             — Class, ClassKind, Field, Visibility
    │   │   ├── function.rs          — Function, Param
    │   │   ├── call.rs              — Call
    │   │   └── import.rs            — Import
    │   └── lang/
    │       ├── interface.rs         — LanguageParser trait
    │       ├── factory.rs           — Language → parser dispatch
    │       └── providers/
    │           ├── go.rs            — Go parser (Tree-sitter)
    │           ├── java.rs          — stub
    │           └── python.rs        — stub
    ├── tests/
    │   └── go_parser.rs             — integration tests
    ├── testdata/                    — sample Go files for manual testing
    ├── Cargo.toml
    ├── README.md                    — this file
    └── TODO.md                      — remaining work

### 🧪 Running tests

    cargo test                    # all tests
    cargo test --test go_parser   # only parser integration tests
    cargo test detector           # only detector tests

### ➕ Adding a new language

1. Add Tree-sitter grammar to `Cargo.toml`:

       tree-sitter-java = "0.23"

2. Implement `LanguageParser` in `src/lang/providers/java.rs`
3. Register in `src/lang/factory.rs`
4. Add integration tests in `tests/java_parser.rs`

---

## 📦 Dependencies

| Crate | Purpose |
| :--- | :--- |
| 🌲 `tree-sitter` | Parser engine |
| 🐹 `tree-sitter-go` | Go grammar |
| ⌨️ `clap` | CLI argument parsing |
| 📄 `serde` + `serde_json` | JSON output |
| ⚠️ `thiserror` | Error type definition |
| 🔐 `sha2` + `hex` | SHA-256 hashing |
| 🚶 `walkdir` | Recursive directory traversal |

---

## 🔗 Links

- 📐 Architecture doc: `../../docs/02-architecture.md`
- 🗺️ Roadmap: `../../docs/05-roadmap.md`
- 📌 TODO: `TODO.md`