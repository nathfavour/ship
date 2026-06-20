# Backend Architecture & Multi-Target Strategy

This document details the architectural strategies for expanding the Ship compiler backend to multiple architectures and operating systems (such as RISC-V and macOS/Mach-O) without compromising compiler simplicity or introducing target contamination.

---

## 1. Decoupling the Intermediate Representation (IR)

To prevent target contamination, the Ship compiler enforces a strict boundary between the frontend/middle-end and the code generation backends:

* **Target-Agnostic IR:** The type checker and AST parser have no knowledge of host registers, page sizes, or executable formats. They emit a flat, linear IR representing 3-address instructions.
* **Modular Code Generation:** Each target (e.g., `compiler/emitter/elf`, `compiler/emitter/macho`) operates solely on the lowered IR. Adding a new architecture only requires consuming this IR.

---

## 2. Register Allocation and Calling Conventions

Currently, the compiler uses a naive stack-allocated register model. To scale performance and support multiple ABIs:

* **Linear Scan Register Allocation (Phase 3):** We will transition from stack-only allocations to a register allocator that maps active IR variables to a target's general-purpose registers (GPRs).
* **Decoupled ABIs:**
  * **System V AMD64 ABI (Linux/macOS x86_64):** Arguments passed via `RDI`, `RSI`, `RDX`, `RCX`, `R8`, `R9`.
  * **Windows x64 ABI (PE Target):** Arguments passed via `RCX`, `RDX`, `R8`, `R9`.
  * **ARM64 AAPCS (Apple Silicon / Linux ARM64):** Arguments passed via `x0` through `x7`.

---

## 3. RISC-V Targeting Strategy

Adding a RISC-V (`RV64GC`) backend is a key milestone for validating our backend decoupling:

* **Simpler ISA:** Unlike x86_64's variable-length instruction prefix complexity, RISC-V uses a clean load/store architecture with uniform 32-bit (or compressed 16-bit) instruction widths.
* **Uniform GPRs:** Direct mapping of 32 orthogonal registers (`x0` through `x31`, where `x0` is hardwired to zero) simplifies register allocator state machines.

---

## 4. macOS Mach-O & Linux Cross-Compilation Logistics

Ship enables native cross-compilation from Linux to macOS (`macho`) without external LLVM or macOS dependencies:

### Dynamic Linking over Direct Syscalls
Apple does not guarantee syscall ABI stability across minor OS updates. Therefore, the Mach-O emitter will:
* Dynamically link against `/usr/lib/libSystem.dylib` (which wraps stable system calls like `exit`, `write`, `read`).
* Emit Mach-O dynamic import tables (`LC_LOAD_DYLIB`, `LC_DYLD_INFO_ONLY`) so that `dyld` resolves these symbols at runtime.

### Ad-hoc Code Signing
Apple Silicon machines immediately `SIGKILL` unsigned Mach-O binaries. We will append ad-hoc signature blocks using the `LC_CODE_SIGNATURE` load command directly inside the Mach-O generator or recommend open-source tools like `rcodesign` inside our toolchain distribution workflow.
