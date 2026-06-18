# Ship Architect Skill

<description>
Guides AI agents and developers in maintaining the strict architectural integrity of the Ship compiler. Use this skill when modifying the compiler's Go source code (Lexer, Parser, IR, or ELF Emitter).
</description>

<instructions>
When activated, you must enforce the following rules for any changes to the Ship compiler:

1.  **Zero Implicit Behavior:** Ensure no implicit memory allocations or garbage collection mechanics are introduced into the Ship architecture.
2.  **Linear Processing:** The compiler is a single-pass system. Do not introduce cyclical dependencies or multi-pass parsing loops unless absolutely critical for SSA optimizations (Phase 3).
3.  **Agentic Priority:** Any new error types or parser failures MUST be formatted into the machine-readable JSON diagnostic stream if the `--agent` flag is present. Do not use standard standard output (`fmt.Println`) for errors when in agent mode.
4.  **No External Dependencies:** Do not import LLVM bindings, external assemblers, or heavy CGO libraries. The compiler must remain purely Go and eventually self-hosting.
5.  **Test Driven:** Every new AST node, Token, or IR opcode requires a corresponding unit test.
</instructions>
