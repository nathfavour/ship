# Architecture & Pipeline

The Ship compilation pipeline is a strictly linear processing chain with no speculative execution loops. It is designed to go from raw bytes to raw executable machine code without touching an intermediate file system step.

```text
[Raw .ship Source Files]
           │
           ▼
     1. Lexer Subsystem   -----> Outputs stream of token.Token
           │
           ▼
    2. Parser Subsystem   -----> Validates grammar, constructs ast.File structures
           │
           ▼
    3. Type/Contract Check -----> Resolves types, validates compile-time `contract` blocks
           │
           ▼
      4. IR Lowering      -----> flattens AST into linear ir.Instruction stream
           │
           ▼
  5. Machine Target Emitter ----> Outputs raw bytes to emitter/elf or emitter/wasm
```

## Subsystem Details

### Lexer (`internal/lexer`)
Converts raw string data into identifiable tokens without allocating excessive intermediary strings.

### Parser (`internal/parser`)
Validates grammar structures and builds the Abstract Syntax Tree (AST). It natively understands `contract` clauses and rejects fundamentally ambiguous nested flows.

### Type Checker (`internal/types`)
Ship enforces rigid static typing. This phase guarantees all expressions align. Notably, it also parses `contract` expressions (e.g. `require: x > 0`) evaluating truth constants at compile time.

### Intermediate Representation (`internal/ir`)
The AST is lowered into a completely flat instruction set composed of opcodes like `OpLoad`, `OpStore`, `OpJump`, and `OpAdd`. This stage mimics machine assembly but maintains architecture agnosticism.

### ELF Emitter (`internal/emitter/elf`)
The IR is translated directly into x86_64 Sys-V opcodes. Instead of linking via GCC, Ship hand-rolls the 64-bit ELF headers, text sections, and byte arrays in memory, writing the final executable instantly.
