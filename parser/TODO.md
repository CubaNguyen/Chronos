# 📌 Parser — TODO

> Backlog of remaining work for the **Chronos Rust parser**.

### 🎨 Legend

| Icon | Meaning |
| :---: | :--- |
| 🔴 | **Next** — highest priority, unblocks others |
| 🟡 | **Should do** — important but not blocking |
| 🟢 | **Nice to have** — polish, edge cases |
| ⚪ | **Future** — deferred indefinitely |

---

## 🔴 NEXT — gRPC Wrap (M14)

> **🎯 Goal:** Expose the parser as a gRPC service so the Go core can call it over the network.

> **💡 Why:** This is the **inflection point** of the project. Until this is done, the parser is only usable via CLI. After this, Go core can ingest AST directly.

### 📝 Steps

- [ ] Write `proto/parser.proto` defining `ParserService.ParseFile`
- [ ] Add `tonic` + `prost` to `Cargo.toml`
- [ ] Setup `build.rs` to compile proto at build time
- [ ] Implement gRPC server wrapping `GoParser`
- [ ] Test manually with `grpcurl`
- [ ] Write Go client in `core/internal/parser/`
- [ ] End-to-end test: Go → gRPC → Rust → AST → Go

**⏱️ Estimated effort:** 1 week

---

## 🟡 SHOULD DO

### 🔧 M8 — Extract Interface Methods

**🐛 Problem:** Currently, `type Reader interface { Read(p []byte) (int, error) }` extracts the interface as a `Class` but **not** the `Read` method signature.

**💥 Impact:** Impact analysis can't tell what methods an interface requires. When refactoring a struct, can't check if it still satisfies an interface.

**🛠️ Implementation:**
- In `go.rs`, extend `extract_types` for `ClassKind::Interface`
- Tree-sitter: interface body is `method_elem` nodes (or `method_spec`)
- Extract method name + params + return type — reuse `extract_params` from functions

**🧪 Tests to add:**
- Interface with 1 method
- Interface with multiple methods
- Interface with embedded interface + methods (mixed)

---

### 🔧 M9 — Extract Const/Var Declarations

**🐛 Problem:** Global `const` and `var` are invisible to the parser.

    const MaxRetries = 5
    var GlobalCache = make(map[string]string)

**💥 Impact:** Impact analysis can't track changes to global constants. Changing `MaxRetries` could affect hundreds of files.

**🛠️ Implementation:**
- Add `Const` and `Var` to `ast/` (new file `constant.rs` and `variable.rs`)
- Extract from top-level `const_declaration` and `var_declaration`
- Add to `Ast` as `Vec<Constant>` and `Vec<Variable>`

**🧪 Tests to add:**
- Single const
- Const block `const ( A = 1; B = 2 )`
- Var with initializer
- Var with type inference

---

### 🔧 M11 — Generics Type Parameters

**🐛 Problem:** `type Stack[T any] struct { items []T }` extracts `Stack` but not the type parameter `T`.

**💥 Impact:** Generics are increasingly common in modern Go (1.18+). Without type params, can't distinguish `Stack[int]` from `Stack[string]`.

**🛠️ Implementation:**
- Tree-sitter: type params live in `type_parameter_list`
- Add `type_params: Vec<TypeParam>` to `Class` and `Function`
- `TypeParam { name: String, constraint: String }`

**🧪 Tests to add:**
- Generic struct with 1 type param
- Generic function with 2 type params
- Generic with constraint (`T comparable`)

---

### 🔧 M12 — Parallel Parsing

**🐛 Problem:** Parsing a repo with **10,000 files** is sequential — one file at a time.

**💥 Impact:** Initial indexing of large repos could take minutes.

**🛠️ Implementation:**
- Add `rayon` to `Cargo.toml`
- In `parse_dir`, use `par_iter` to parse files in parallel
- Each file is independent → **embarrassingly parallel**
- Collect results into `Vec<Ast>`

**🧪 Tests:**
- Benchmark: 1000 files sequential vs parallel
- Verify results are identical (order-independent)

---

## 🟢 NICE TO HAVE

### 📊 M13 — Benchmark Suite

**🎯 Goal:** Measure parse speed to know when optimization is needed.

**🛠️ Implementation:**
- Add `criterion` to `[dev-dependencies]`
- Benchmark `GoParser::parse` on:
  - Small file (~100 lines)
  - Medium file (~1000 lines)
  - Large file (~10000 lines)
- Report throughput (files/sec, lines/sec)

---

### 🛡️ Error Recovery

**🐛 Problem:** A file with syntax errors fails entirely. Partial AST is lost.

**💥 Impact:** Real codebases often have files in-progress or with build tags that confuse the parser.

**🛠️ Implementation:**
- Tree-sitter has `ERROR` nodes for syntax errors
- Continue walking the tree — extract what's parseable
- Add `has_errors: bool` to `FileInfo`

---

### 🏷️ Struct Tags

**🐛 Problem:** Go struct tags carry metadata.

    type User struct {
        ID   int    `json:"id" db:"user_id"`
        Name string `json:"name"`
    }

**🛠️ Implementation:**
- Extract `` `json:"id" db:"user_id"` `` as `tags: Vec<String>` in `Field`

---

### 🌍 Multi-language Codebases

**🐛 Problem:** Only **Go** is implemented. Java/Python/Rust/TypeScript providers are stubs.

**🛠️ Implementation (per language):**
1. Add Tree-sitter grammar
2. Implement `LanguageParser` in `src/lang/providers/<lang>.rs`
3. Register in `factory.rs`
4. Add integration tests

**📊 Priority order:**

| # | Language | Reason |
| :-: | :--- | :--- |
| 1 | ☕ **Java** | Most requested in enterprise |
| 2 | 🐍 **Python** | Most AI/ML projects |
| 3 | 📜 **TypeScript** | Most web projects |
| 4 | 🦀 **Rust** | Self-parsing for dogfooding |

---

## ⚪ FUTURE

### 🚀 M14+ — Advanced

- 🔗 **Cross-file symbol resolution:** Currently `callee` is raw (`s.db.Query`). Need to resolve to actual `Function` node
- 💾 **AST caching:** Store parsed AST in disk cache keyed by file hash. Skip re-parse if hash matches
- ⚡ **Incremental parsing:** Tree-sitter supports re-parsing with old tree — faster for large files
- 🧠 **Language server protocol (LSP):** Expose parser as an LSP server so IDEs can use it
- 🌐 **Multi-repo analysis:** Parse dependencies from `go.mod`, `package.json`, etc.

---

## ✅ DONE (Archive)

- [x] 🏗️ **M1** — Cargo.toml + deps
- [x] 🧩 **M2** — Domain model (`ast/`)
- [x] 🔍 **M3** — `detector.rs` with 5 languages
- [x] 🏭 **M4** — `LanguageParser` trait + factory
- [x] 🐹 **M5** — GoParser: imports, classes, fields, functions, methods, calls
- [x] 🔗 **M5.5** — Embedded struct/interface (extends, implements)
- [x] 📦 **M5.6** — Variadic params, closure call handling
- [x] 🧪 **M6** — Test edge cases (generics, variadic, nested, multifile)
- [x] ⌨️ **M7** — CLI with clap
- [x] ✅ **M10** — 15 integration tests for GoParser
- [x] 🐛 **Fix** — file_path on Import, SHA-256 hash

---

## 📖 How to use this file

1. 🎯 Pick the top **🔴 Next** item. Don't jump around.
2. 🏁 Do it **end-to-end**. Code + tests + update this file.
3. 📦 Move done items to **"Done (Archive)"** at the bottom.
4. 🔄 Re-prioritize when a blocker appears (e.g., another module needs a specific feature).
5. ⚠️ Do **not** add items to 🔴 Next without a concrete reason. If it's "nice to have" → it goes to 🟢.