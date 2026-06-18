# ARCHITECTURE.md: The `ship` Compiler & Toolchain Specification

## 1. Core Philosophy & Design Axioms

`ship` is a systems programming language and toolchain engineered from first principles for the agentic era. It rejects the accidental complexity of modern toolchains (bloated compilation times, implicit language behaviors, heavy dependencies like LLVM) in favor of **unyielding determinism, raw compilation velocity, and absolute machine-readability.**

### The Three Laws of `ship` Architecture:

1. **Zero Implicit Behavior:** Every allocation, control-flow branch, and error path must be written explicitly in the syntax. There are no exceptions, no hidden macros, no garbage collection, and no implicit type coercion.
2. **Nanosecond Single-Pass Execution:** The toolchain must compile from source to raw machine bytes directly without invoking external assemblers or linkers. No LLVM. No GCC/Clang wrappers.
3. **Machine-First Interface:** The compiler targets silicon (x86_64 binaries) and cognitive systems (AI agents) symmetrically. Diagnostics and metadata graphs are streamed as deterministic structured data payloads.

---

## 2. Directory Layout & Repository Mapping

The repository is organized as a strict, flat monorepo. Dependencies between packages are linear and directional to prevent cyclical imports.

```text
ship/
├── cli/                       # Command Line Interface Workspace Module
│   └── ship/                  # CLI Entry point (main.go) -> Handles flags, orchestrates phases
├── compiler/                  # Core Compiler Workspace Module
│   ├── token/                 # Keywords, operators, literals definitions
│   ├── lexer/                 # Lexical Analysis: Raw bytes -> Token stream
│   ├── ast/                   # Abstract Syntax Tree Node Definitions
│   ├── parser/                # Syntactic Analysis: Token stream -> AST
│   ├── types/                 # Type Verification, Structural Analysis, & Contract Inferences
│   ├── ir/                    # Linear Intermediate Representation (Single-assignment, flat instructions)
│   ├── emitter/               # Direct Binary Generation
│   │   ├── elf/               # x86_64 ELF format encoder (Linux)
│   │   ├── macho/             # x86_64 Mach-O format encoder (macOS)
│   │   ├── pe/                # x86_64 PE format encoder (Windows)
│   │   └── wasm/              # WebAssembly bytecode encoder
│   └── agent/                 # Machine-first diagnostics stream & .shipmap manifest engine
├── lsp/                       # Language Server Protocol Workspace Module (Daemon)
├── std/                       # The Ship Standard Library (Written completely in .ship)
│   ├── core/                  # Bare-metal fundamentals, string primitives, sys-call boundaries
│   ├── memory/                # Standard Manual Allocators (Arena, Page, Heap boundary)
│   └── crypto/                # High-performance cryptographic routines
└── Makefile                   # Pure automation for toolchain self-bootstrapping
```

---

## 3. Compiler Pipeline Architecture

The compilation pipeline is a strictly linear processing chain with no speculative execution loops.

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

---

## 4. Internal Data Structures (Go Specification)

The following precise structures dictate the internal data contracts within the `internal/` subsystem.

### 4.1 Token Subsystem (`internal/token/token.go`)

```go
package token

type Type string

const (
	ILLEGAL   Type = "ILLEGAL"
	EOF       Type = "EOF"
	IDENT     Type = "IDENT"
	INT       Type = "INT"
	STRING    Type = "STRING"
	
	// Keywords
	TYPE      Type = "type"
	STRUCT    Type = "struct"
	FN        Type = "fn"
	CONTRACT  Type = "contract"
	REQUIRE   Type = "require"
	ENSURE    Type = "ensure"
	LET       Type = "let"
	DEFER     Type = "defer"
	ELSE      Type = "else"
	RETURN    Type = "return"
	IF        Type = "if"
)

type Token struct {
	Type    Type
	Literal string
	File    string
	Line    int
	Col     int
}
```

### 4.2 AST Subsystem (`internal/ast/ast.go`)

```go
package ast

import "ship/internal/token"

type Node interface {
	Pos() token.Token
}

type StructField struct {
	Name token.Token
	Type token.Token
}

type StructDecl struct {
	Token token.Token // token.TYPE
	Name  token.Token // Name of the struct
	Fields []StructField
}

type ContractBlock struct {
	Token    token.Token // token.CONTRACT
	Requires []Expression
	Ensures  []Expression
}

type ElseErrBlock struct {
	Token      token.Token // token.ELSE
	ErrIdent   token.Token // The bound error identifier
	Body       []Statement
}

type FuncDecl struct {
	Token      token.Token // token.FN
	Name       token.Token
	Params     []StructField
	ReturnType token.Token
	Contract   *ContractBlock
	Body       []Statement
}

type Statement interface {
	Node
	statementNode()
}

type Expression interface {
	Node
	expressionNode()
}
```

### 4.3 Linear IR Subsystem (`internal/ir/ir.go`)

The IR strips away structural nesting. Control flow is flattened into explicit labels and conditional jumps.

```go
package ir

type OpCode int

const (
	OpStore OpCode = iota
	OpLoad
	OpAdd
	OpSub
	OpCall
	OpLabel
	OpJumpIfZero
	OpRet
)

type Operand struct {
	Type  string
	Value string
}

type Instruction struct {
	Op       OpCode
	Dest     Operand
	Src1     Operand
	Src2     Operand
	Comment  string
}

type Program struct {
	Instructions []Instruction
}
```

---

## 5. The Agentic Interface Protocol (Machine-First)

When running the compiler with the `--agent` flag, standard text logs are suppressed. The toolchain outputs newline-delimited JSON messages to `stdout` / `stderr`.

### 5.1 Diagnostic Error Spec

If parsing, type verification, or contract analysis fails, the output structure must strictly match:

```json
{
  "status": "error",
  "phase": "TYPE_CHECKER",
  "error_code": "CONTRACT_VIOLATION_REQUIRE",
  "target": {
    "file": "std/crypto/secure.ship",
    "function": "parse_secure_slice",
    "line": 14,
    "char": 9
  },
  "context": {
    "ast_node": "ContractBlock",
    "violated_expression": "payload.len > 0",
    "inferred_type": "Slice[U8]"
  }
}
```

### 5.2 The Cognitive Blueprint (`.shipmap`)

Every successful execution of `ship build` drops a `.shipmap` metadata manifest file into the target directory. This layout enables autonomous swarm systems to compute compilation cascades without reprocessing code files line-by-line.

```json
{
  "compiler_version": "0.0.1-go-bootstrap",
  "target_architecture": "x86_64-elf",
  "source_hashes": {
    "std/core/memory.ship": "sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
    "std/crypto/secure.ship": "sha256:8f43434664836817281923192313133649b934ca495991b7852b855aa112344"
  },
  "type_ledger": [
    {
      "name": "Transaction",
      "size_bytes": 48,
      "fields": [
        {"name": "id", "type": "String", "offset": 0},
        {"name": "amount", "type": "U64", "offset": 16},
        {"name": "signer", "type": "Address", "offset": 24}
      ]
    }
  ],
  "dependency_graph": {
    "main.ship": ["std/core/memory.ship", "std/crypto/secure.ship"],
    "std/crypto/secure.ship": ["std/core/memory.ship"]
  }
}
```

---

## 6. Direct x86_64 Emission Specification

The target backend (`internal/emitter/elf/`) compiles directly to machine instructions, bypassing assembly parsing.

### 6.1 Register Allocation Model

* **System Call Rules:** Adheres strictly to System V AMD64 ABI guidelines for Unix-based platforms.
* **Argument Mapping:** Initial function parameters pass through `RDI`, `RSI`, `RDX`, `RCX`, `R8`, `R9`.
* **Return Allocation:** Value arrays and structures return values through `RAX` (and `RDX` for composite structures or error indicators).

### 6.2 The ELF Generation Layout

The binary generation engine builds the ELF file structure segment by segment sequentially via an in-memory byte slice:

```text
+-----------------------------------+
|             ELF Header            | -> Architecture parameters (64-bit, x86_64)
+-----------------------------------+
|      Program Header Table         | -> Defines text and data segments
+-----------------------------------+
|          .text Section            | -> Raw CPU opcodes generated from linear IR
+-----------------------------------+
|          .data Section            | -> Static allocations, constants, and strings
+-----------------------------------+
|      Section Header Table         | -> System link indices (omitted in ultra-stripped builds)
+-----------------------------------+
```

---

## 7. Memory & Standard Library Constraints

1. **No Implicit Heap Allocations:** The keyword `new` or any automatic escaping allocation does not exist.
2. **Explicit Dependency Injection:** If a standard library block requires dynamic memory, the architecture forces the instantiation block to explicitly hand over an allocator reference:
```text
fn compute_hash(allocator: Allocator, input: String) -> Result[Hash, Error]
```

3. **Deterministic Cleanup:** Every allocation cycle must declare a corresponding tracking step via `defer` to ensure continuous stack or frame cleanup at execution limits.

---

This blueprint is complete, rigid, and ready for immediate implementation.