# Benchmarks

Ship is designed for **raw compilation velocity**. This document serves as the foundation for tracking our performance metrics against other systems programming toolchains (like Rust/rustc, C/Clang, Go, and Zig).

*(Benchmarks are currently being gathered. The following is the planned structure for tracking).*

## Compilation Velocity

### Metric: Source to Binary (Cold Cache)
| Language / Toolchain | Project Size (LOC) | Compilation Time (ms) | Peak Memory Usage (MB) |
|----------------------|--------------------|-----------------------|------------------------|
| **Ship**             | 10,000             | TBD                   | TBD                    |
| C (Clang -O0)        | 10,000             | TBD                   | TBD                    |
| Go                   | 10,000             | TBD                   | TBD                    |
| Rust (rustc debug)   | 10,000             | TBD                   | TBD                    |

### Metric: Source to Binary (Warm Cache / Incremental)
*(Note: Ship's single-pass execution aims to make cold compilation so fast that incremental compilation is barely distinguishable).*
| Language / Toolchain | Project Size (LOC) | Compilation Time (ms) | Peak Memory Usage (MB) |
|----------------------|--------------------|-----------------------|------------------------|
| **Ship**             | 10,000             | TBD                   | TBD                    |
| C (Clang -O0)        | 10,000             | TBD                   | TBD                    |
| Go                   | 10,000             | TBD                   | TBD                    |
| Rust (rustc debug)   | 10,000             | TBD                   | TBD                    |

## Execution Performance (Generated x86_64)

### Metric: N-Body Simulation
| Language / Toolchain | Execution Time (s) |
|----------------------|--------------------|
| **Ship**             | TBD                |
| C (Clang -O3)        | TBD                |
| Rust (rustc --release)| TBD               |

## Methodology

- CPU: [Insert CPU Model]
- OS: [Insert OS/Kernel Version]
- Ram: [Insert RAM specs]
- Scripts: Benchmark scripts are available in the `scripts/benchmarks` directory (planned).
