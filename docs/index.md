# 🚢 Ship Compiler

Welcome to the official documentation for **Ship**, a systems programming language and toolchain engineered from first principles for the agentic era.

Modern toolchains are plagued by accidental complexity: bloated compilation times, implicit language behaviors, and heavy dependencies like LLVM. Ship rejects this. Instead, it offers **unyielding determinism, raw compilation velocity, and absolute machine-readability**.

## The Philosophy

1. **Zero Implicit Behavior:** Every allocation, control-flow branch, and error path must be written explicitly. No hidden macros, no garbage collection, and no implicit type coercion.
2. **Nanosecond Single-Pass Execution:** The toolchain compiles from source to raw machine bytes directly without invoking external assemblers or linkers.
3. **Machine-First Interface:** The compiler targets silicon (x86_64 binaries) and cognitive systems (AI agents) symmetrically. Diagnostics and metadata graphs are streamed as deterministic structured JSON payloads (`--agent`).

## Features

- **Blazing Fast ELF Generation:** Hand-rolled x86_64 ELF emitter bypassing LLVM/GCC entirely.
- **Agentic Telemetry:** Run `ship --agent file.ship` to output structured JSON error reporting and drop `.shipmap` cognitive blueprints for AI swarms.
- **Contract-Oriented Programming:** Built-in `contract` blocks with `require`/`ensure` clauses evaluated at compile time.
- **Flat IR:** Clean, single-assignment Intermediate Representation for minimal-overhead optimization.
