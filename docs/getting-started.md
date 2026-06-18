# Getting Started

Ship is designed to be frictionless to install and run. You can grab the latest toolchain using AnyIsland or build it manually from source.

## Installation

### Method 1: AnyIsland (Recommended)
If you have AnyIsland configured, installing Ship is completely seamless:
```bash
anyisland install nathfavour/ship
```
This command automatically pulls the repository, compiles the toolchain, and links the `ship` executable globally.

### Method 2: Build from Source
Ensure you have Go 1.25.5 or later installed.
```bash
git clone https://github.com/nathfavour/ship.git
cd ship
go build -o ship ./cmd/ship
```

## Your First Ship Program

Create a file named `hello.ship`:

```rust
fn main() -> int {
    let x = 10;
    let y = 5;
    return x + y;
}
```

### Compiling and Running

You can compile the file and inspect the generated x86_64 `a.out` binary:
```bash
ship build hello.ship
./a.out
```

Alternatively, use the `run` command to compile and execute the file immediately in memory without leaving a binary behind:
```bash
ship run hello.ship
```
