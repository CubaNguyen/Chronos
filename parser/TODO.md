# Parser — TODO

## 🔴 Next
- **None.** Parser is feature-complete for Phase 1 MVP.

## 🟡 Should Do (Phase 1.5)

### M8 — Extract Interface Methods
- [ ] Extract method signatures from Go interfaces (currently ignored).
- [ ] Update `ast::Class` or `ast::Function` to link methods to interfaces.

### M9 — Extract Const/Var Declarations
- [ ] Parse top-level `const` and `var` blocks.
- [ ] Add `Constant` and `Variable` entities to AST.

### M11 — Generics Type Parameters
- [ ] Extract type params (`Stack[T any]` -> `T`).
- [ ] Add `type_params` field to `Class` and `Function`.

## 🟢 Nice to Have

### M12 — Parallel Parsing
- [ ] Use `rayon` to parse multiple files concurrently.
- [ ] Benchmark 1k files to measure speedup.

### Error Recovery
- [ ] Partial AST extraction if syntax error is found (instead of failing the whole file).

### Multi-language
- [ ] Implement Java provider.
- [ ] Implement Python provider.

---

## Done (Archive)
- [x] **Phase 1 MVP — F1: Parse Go Code**
- [x] **B1-B4: gRPC Wrap (Rust Server)**
- [x] Reflection enabled.
- [x] Rust CLI + JSON output.