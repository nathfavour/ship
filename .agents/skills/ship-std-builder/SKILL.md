# Ship Standard Library Builder Skill

<description>
Guides AI agents in writing `.ship` source files for the Ship Standard Library (`std/`). Use this skill when implementing fundamental data structures, memory allocators, or system calls in the Ship language.
</description>

<instructions>
When building standard library modules in `.ship`, adhere to these constraints:

1.  **Manual Memory:** You must not use implicit allocations. Dynamic memory structures must accept an `Allocator` as an injected dependency.
    ```text
    // Example
    fn create_buffer(alloc Allocator, size int) -> Buffer { ... }
    ```
2.  **Contracts:** Utilize `contract` blocks extensively to ensure data integrity and array bounds.
    ```text
    contract {
        require: size > 0
    }
    ```
3.  **Strict Typing:** There is no implicit type coercion. Explicitly cast or handle types.
4.  **Minimal Dependencies:** Standard library modules must rely only on other standard library primitives.
5.  **Deterministic Cleanup:** Track allocations and ensure `defer` is utilized appropriately for cleanup when that feature is fully supported.
</instructions>
