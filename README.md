# 🚢 Ship: The Agentic Era Systems Compiler

**Ship** is a systems programming language and toolchain engineered from first principles for the agentic era.

Modern toolchains are plagued by accidental complexity: bloated compilation times, implicit language behaviors, and heavy dependencies like LLVM. Ship rejects this. Instead, it offers **unyielding determinism, raw compilation velocity, and absolute machine-readability**.

> 💡 **AnyIsland Ready:** Run `anyisland install nathfavour/ship` to compile and expose the Ship toolchain instantly.

## The Three Laws of `ship`

1. **Zero Implicit Behavior:** Every allocation, control-flow branch, and error path must be written explicitly. No hidden macros, no garbage collection, and no implicit type coercion.
2. **Nanosecond Single-Pass Execution:** The toolchain compiles from source to raw machine bytes directly without invoking external assemblers or linkers.
3. **Machine-First Interface:** The compiler targets silicon (x86_64 binaries) and cognitive systems (AI agents) symmetrically. Diagnostics and metadata graphs are streamed as deterministic structured JSON payloads (`--agent`).

## Features

- **Blazing Fast ELF Generation:** Hand-rolled x86_64 ELF emitter bypassing LLVM/GCC entirely.
- **Agentic Telemetry:** Run `ship --agent file.ship` to output structured JSON error reporting and drop `.shipmap` cognitive blueprints for AI swarms.
- **Contract-Oriented Programming:** Built-in `contract` blocks with `require`/`ensure` clauses evaluated at compile time.
- **Flat IR:** Clean, single-assignment Intermediate Representation for minimal-overhead optimization.

## Quick Start

### Using AnyIsland (Recommended)
You'll need to install the [AnyIsland Package Manager](https://github.com/nathfavour/anyisland) first. Once AnyIsland is set up, installing Ship is instant:
```bash
anyisland install nathfavour/ship
```

### Manual Installation
Make sure you have Go 1.25.5 or later installed.
```bash
git clone https://github.com/nathfavour/ship.git
cd ship
go build -o ship ./cmd/ship

ship run file.ship
```

## Language Syntax at a Glance

Ship's syntax is strictly typed, explicit, and C-like. Here is the entire language structure in a nutshell:

```rust
// 1. Structs (Strict linear memory alignment)
type Transaction struct {
    amount int
    recipient string
}

// 2. Functions & Strong Typing
fn process(x int, y int) -> int {
    
    // 3. Compile-Time Contracts
    contract {
        require: x > 0
        ensure: x != 100
    }

    // 4. Variables & Control Flow
    let sum = x + y;
    if (sum > 50) {
        return sum;
    } else {
        return 0;
    }
}
```

For the complete breakdown, view the [Syntax Reference Guide](https://github.com/nathfavour/ship/blob/master/docs/syntax.md).

## Architecture

See [ARCHITECTURE.md](./ARCHITECTURE.md) for the definitive compiler and toolchain specification.

## Contributing

We welcome contributions! Please see [CONTRIBUTING.md](./CONTRIBUTING.md) for guidelines on how to get started.

## Benchmarks

See [BENCHMARKS.md](./BENCHMARKS.md) for performance metrics against other modern compilers.
