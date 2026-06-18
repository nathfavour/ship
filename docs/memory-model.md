# Memory Model

Ship enforces a strict memory model based entirely on **Zero Implicit Behavior**. 

In modern higher-level systems languages, keywords like `new` or `make` implicitly hide OS interactions and heap allocations from the developer. Ship fundamentally rejects this approach.

## 1. No Implicit Heap Allocations
The keyword `new` does not exist as an automatic heap-escaping mechanism in the execution layer. A standard Ship struct instantiates linearly based on context.

## 2. Explicit Dependency Injection
If a standard library block or dynamic data structure (like a HashMap or growable Array) requires dynamic memory, the architecture forces the developer to instantiate and explicitly pass an `Allocator` reference.

```rust
// A hypothetical standard library function requiring memory
fn compute_hash(allocator Allocator, input string) -> Result {
    // Execution uses the explicitly provided allocator...
}
```

## 3. Deterministic Cleanup
Every allocation cycle must declare a corresponding tracking step via the `defer` keyword (or explicit teardown instructions). This ensures continuous stack or frame cleanup precisely at execution limits without invoking a garbage collector cycle.

## Supported Allocators (Planned Phase 2)
The standard library (`std/memory`) will provide reference implementations for:

* **Arena Allocator:** Rapid bump-pointer allocations intended to be freed all at once.
* **Page Allocator:** Direct raw OS page mappings.
* **Heap Boundary:** Highly restrictive, strictly tracked explicit heap mapping.
