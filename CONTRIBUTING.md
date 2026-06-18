# Contributing to Ship

Thank you for your interest in contributing to **Ship**! 

Ship is designed to be an uncompromisingly deterministic, machine-readable compiler toolchain. When contributing, please ensure that your code adheres to our core philosophy of **Zero Implicit Behavior**.

## Getting Started

1. **Fork & Clone:** Fork the repository and clone it locally.
2. **Go Workspace:** The project uses Go workspaces (`go.work`). Simply run `go build ./cmd/ship` to test your changes.
3. **Tests:** Ensure you run all existing tests before submitting a PR:
   ```bash
   go test ./...
   ```

## Development Guidelines

- **No External Toolchains:** Do not introduce dependencies on external linkers, assemblers, or heavy libraries (like LLVM).
- **Agent-First Reporting:** If you modify parser or type-checker errors, ensure they properly format into the `--agent` JSON diagnostic stream.
- **AST and IR:** Follow the strict separation of phases: Lexer -> Parser -> Type Checker -> IR Lowering -> Target Emission.
- **Test-Driven:** Write tests for any new lexer tokens, AST nodes, or IR instructions.

## Submitting Pull Requests

1. Create a descriptive branch name.
2. Commit your changes with clear, concise messages.
3. Open a Pull Request against the `main` branch.
4. Ensure all CI checks pass.

Welcome aboard! 🚢
