# Zero-Syscall Kernel I/O via io_uring (SQPOLL)

Ship provides raw, bare-metal built-in function wrappers for the Linux `io_uring` subsystems and memory mapping system calls. This enables high-performance, asynchronous, zero-syscall user-space I/O.

---

## 1. Exposed Compiler Built-ins

The following low-level built-in primitives are now wired directly into the compiler, lowering to their respective Linux x86_64 system calls:

### `sys_io_uring_setup`
Initializes a new `io_uring` instance.
```rust
fn sys_io_uring_setup(entries int, params int) -> int
```
* **Syscall ID:** 425
* **Parameters:**
  * `entries`: Size of the submission and completion queues.
  * `params`: Memory pointer to the `io_uring_params` structural setup.
* **Returns:** File descriptor of the `io_uring` instance, or `-1` on error.

### `sys_io_uring_enter`
Initiates and completes I/O operations.
```rust
fn sys_io_uring_enter(fd int, to_submit int, min_complete int, flags int) -> int
```
* **Syscall ID:** 426
* **Parameters:**
  * `fd`: `io_uring` ring file descriptor.
  * `to_submit`: Number of entries to submit.
  * `min_complete`: Minimum completed operations to wait for.
  * `flags`: Completion modifiers (e.g. `IORING_ENTER_GETEVENTS`).

### `sys_io_uring_register`
Registers user-space buffers or files for kernel polling optimization.
```rust
fn sys_io_uring_register(fd int, opcode int, arg int, nr_args int) -> int
```
* **Syscall ID:** 427

### `sys_mmap`
Maps memory pages shared between user space and kernel space.
```rust
fn sys_mmap(addr int, len int, prot int, flags int, fd int, offset int) -> int
```
* **Syscall ID:** 9

---

## 2. Implementing Zero-Syscall I/O (SQPOLL)

To bypass the context-switching penalty of system calls entirely:

1. **Enable SQPOLL Mode:** Set the `IORING_SETUP_SQPOLL` flag inside your `io_uring_params` struct before calling `sys_io_uring_setup`.
2. **Kernel Thread Loop:** The kernel spawns a dedicated background thread that continuously polls the shared Submission Queue (SQ).
3. **Queue Updates:** In user-space, write requests (e.g., file reads, packet sends) straight to the submission ring memory mapped via `sys_mmap`, and increment the queue's tail pointer.
4. **Result Collection:** Read results from the Completion Queue (CQ) buffers.

This lock-less ring buffer flow entirely avoids `syscall` transitions, context switches, and TLB flushes, achieving maximum hardware throughput.
