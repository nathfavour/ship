# Ship Language Syntax

Ship uses a strict, C-like grammar with no implicit type coercion and zero hidden allocations. This document provides the complete syntax structure of the language, easily ingestable for agents and developers.

## Primitives and Types
Ship currently supports the following base types:
- `int` (64-bit integer)
- `string` (immutable UTF-8 sequence)
- `bool` (boolean `true` or `false`)

## Variables
Variables are declared using the `let` keyword. Reassignments and mutability are explicit.

```rust
let x = 5;
let message = "Hello, Agent";
let is_valid = true;
```

## Functions
Functions use the `fn` keyword. Types must be explicitly declared for both parameters and the return value (denoted by `->`).

```rust
fn add(x int, y int) -> int {
    return x + y;
}
```

## Control Flow
Ship supports standard `if` and `else` blocks. Parentheses around conditions are required.

```rust
if (x > 5) {
    return true;
} else {
    return false;
}
```

## Structs (Data Types)
Custom data structures are defined using `type ... struct`. Memory alignment is strictly linear.

```rust
type Transaction struct {
    amount int
    recipient string
}
```

## Contracts (Compile-Time Validation)
Ship uniquely supports `contract` blocks that enforce invariants at compile-time. Use `require` for preconditions and `ensure` for postconditions.

```rust
fn process(x int) -> bool {
    contract {
        require: x > 0
        ensure: x != 100
    }
    
    // Function body...
    return true;
}
```

## Operators
- **Arithmetic:** `+`, `-`, `*`, `/`
- **Comparison:** `==`, `!=`, `<`, `>`
- **Logical:** Prefix `!`
