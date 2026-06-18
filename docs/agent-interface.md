# Agentic Interface

Ship is a **machine-first toolchain**. It targets silicon binaries and cognitive AI agents symmetrically. When invoked with the `--agent` flag, standard text logs are suppressed, and the toolchain outputs deterministic structured data payloads instead.

## The Cognitive Blueprint (`.shipmap`)

Every successful execution of `ship build --agent` drops a `.shipmap` metadata manifest file into the target directory. 

This enables autonomous swarm systems and local LLM agents to compute compilation cascades without needing to re-parse the code.

```json
{
  "compiler_version": "0.0.1-go-bootstrap",
  "target_architecture": "x86_64-elf",
  "source_hashes": {
    "test.ship": "sha256:dummy"
  },
  "type_ledger": [],
  "dependency_graph": {
    "test.ship": []
  }
}
```

## Structured Error Payloads

When a compilation error occurs in `--agent` mode, the terminal output is suppressed, and Ship streams JSON directly to `stderr`. This allows wrapper orchestrators to programmatically patch code rather than parsing string regexes.

```json
{
  "status": "error",
  "phase": "TYPE_CHECKER",
  "error_code": "CONTRACT_VIOLATION_REQUIRE",
  "target": {
    "file": "test.ship",
    "function": "process_data",
    "line": 14,
    "char": 9
  },
  "context": {
    "ast_node": "ContractBlock",
    "violated_expression": "payload.len > 0",
    "inferred_type": "Slice[U8]"
  }
}
```
