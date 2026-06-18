# 🗺️ Ship Compiler: The Road to 1.0

The `ship` compiler is built for deterministic execution and machine-first telemetry. We are currently at **v0.0.1-go-bootstrap**, having successfully proven out our single-pass architecture, parser, IR lowering, and native x86_64 ELF generation.

This roadmap details the path to a production-ready, self-hosting 1.0 release.

---

## 🟢 Phase 1: The Foundation (Completed)
*Goal: Prove the end-to-end architecture from raw text to executable bytes without external dependencies.*

- [x] Zero-dependency Lexer & Parser
- [x] Abstract Syntax Tree & `contract` nodes
- [x] Strict Type Checker (No implicit coercions)
- [x] Linear Intermediate Representation (IR) Lowering
- [x] Naive x86_64 ELF Direct Emission
- [x] `cmd/ship` Orchestrator (`build`, `run`)
- [x] Agentic Diagnostic Protocol (`--agent`, `.shipmap` generation)

---

## 🟡 Phase 2: Memory Model & Standard Library
*Goal: Implement strict manual memory management and build the core standard library in Ship.*

- [ ] **Memory Allocators (`std/memory`)**
  - Implement Arena, Page, and strict Heap allocators.
  - Enforce explicit `Allocator` injection for dynamic data structures.
- [ ] **Data Structures (`std/core`)**
  - Contiguous Slices, static Strings, and Hash Maps.
- [ ] **Control Flow Expansion**
  - Add `loop`, `break`, and strict bounds-checked iterations.
- [ ] **Advanced Contracts**
  - Expand `require`/`ensure` to validate array bounds and memory offsets at compile-time.

---

## 🟠 Phase 3: The Optimizer & IR Enhancements
*Goal: Generate highly efficient machine code to rival Clang -O3, all in a single pass.*

- [ ] **Static Single Assignment (SSA)**
  - Convert flat IR to SSA form for data flow analysis.
- [ ] **Constant Folding & Propagation**
  - Resolve deterministic math and boolean logic at compile time.
- [ ] **Dead Code Elimination (DCE)**
  - Strip unreachable branches and unused functions.
- [ ] **Register Allocation**
  - Implement Linear Scan Register Allocation for optimal CPU register usage.

---

## 🟣 Phase 4: Target Architectures & ABI
*Goal: Fully support standard operating systems and calling conventions.*

- [ ] **System V AMD64 ABI**
  - Complete register mapping for standard C FFI interoperability.
- [ ] **Mach-O (macOS / ARM64 & x86_64)**
  - Implement native emission for Apple Silicon and Intel Macs.
- [ ] **WebAssembly (WASM)**
  - Compile Ship directly to `.wasm` binaries.
- [ ] **Syscall Integration**
  - Wrap raw Linux/macOS syscalls within `std/sys`.

---

## 🔵 Phase 5: Agentic Evolution
*Goal: Make the compiler an active participant in code repair and swarm intelligence.*

- [ ] **Automated Repair Payloads**
  - Type Checker errors include AST-patch recommendations via JSON.
- [ ] **Cognitive `.shipmap` Graphs**
  - Manifests include structural hash chains and explicit memory offsets for struct fields.
- [ ] **Local Repository Skills**
  - Full `.agents/skills` integration for semantic code generation.

---

## 🚀 Phase 6: Self-Hosting (Ship in Ship)
*Goal: The ultimate test of a systems language.*

- [ ] Rewrite the Lexer in Ship.
- [ ] Rewrite the Parser & Type Checker in Ship.
- [ ] Rewrite the IR & ELF Emitter in Ship.
- [ ] Bootstrapping complete: Retire the Go codebase.

---

## Contributing to the Roadmap
If you are an agent or a human looking to contribute, select an open item from the nearest phase and consult `ARCHITECTURE.md` before writing code.
