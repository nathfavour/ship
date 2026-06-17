package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/nathfavour/ship/internal/emitter/elf"
	"github.com/nathfavour/ship/internal/ir"
	"github.com/nathfavour/ship/internal/lexer"
	"github.com/nathfavour/ship/internal/parser"
	"github.com/nathfavour/ship/internal/types"
)

func main() {
	agentFlag := flag.Bool("agent", false, "Enable machine-first diagnostic stream (JSON)")
	outputFlag := flag.String("o", "a.out", "Output executable file")
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		fatalError("No input file provided", *agentFlag)
	}

	inputFile := args[0]
	
	bytes, err := os.ReadFile(inputFile)
	if err != nil {
		fatalError(fmt.Sprintf("Could not read file: %s", err), *agentFlag)
	}

	input := string(bytes)

	// 1. Lex
	l := lexer.New(inputFile, input)

	// 2. Parse
	p := parser.New(l)
	astFile := p.ParseFile()
	
	if len(p.Errors()) > 0 {
		if *agentFlag {
			reportAgentErrors("PARSER", p.Errors())
		} else {
			for _, msg := range p.Errors() {
				fmt.Fprintf(os.Stderr, "Parser error: %s\n", msg)
			}
		}
		os.Exit(1)
	}

	// 3. Type Check
	checker := types.NewChecker()
	checker.CheckFile(astFile)
	
	if len(checker.Errors()) > 0 {
		if *agentFlag {
			reportCheckerAgentErrors("TYPE_CHECKER", checker.Errors())
		} else {
			for _, e := range checker.Errors() {
				fmt.Fprintf(os.Stderr, "Type error: %s\n", e.Error())
			}
		}
		os.Exit(1)
	}

	// 4. Lower to IR
	lowerer := ir.NewLowerer()
	program := lowerer.LowerFile(astFile)

	// 5. Emit ELF
	emitter := elf.New(program)
	binaryBytes, err := emitter.Emit()
	if err != nil {
		fatalError(fmt.Sprintf("Failed to emit ELF: %s", err), *agentFlag)
	}

	err = os.WriteFile(*outputFlag, binaryBytes, 0755)
	if err != nil {
		fatalError(fmt.Sprintf("Failed to write output binary: %s", err), *agentFlag)
	}

	if *agentFlag {
		// Emit mock .shipmap for agentic system
		emitShipmap(inputFile)
	}
}

func fatalError(msg string, agent bool) {
	if agent {
		out, _ := json.Marshal(map[string]interface{}{
			"status": "error",
			"phase":  "SYSTEM",
			"error":  msg,
		})
		fmt.Fprintln(os.Stderr, string(out))
	} else {
		fmt.Fprintf(os.Stderr, "Fatal: %s\n", msg)
	}
	os.Exit(1)
}

func reportAgentErrors(phase string, errors []string) {
	for _, msg := range errors {
		out, _ := json.Marshal(map[string]interface{}{
			"status": "error",
			"phase":  phase,
			"error":  msg,
		})
		fmt.Fprintln(os.Stderr, string(out))
	}
}

func reportCheckerAgentErrors(phase string, errors []*types.CheckerError) {
	for _, err := range errors {
		out, _ := json.Marshal(map[string]interface{}{
			"status":     "error",
			"phase":      phase,
			"error_code": err.ErrorCode,
			"target": map[string]interface{}{
				"file":     err.TargetFile,
				"function": err.Function,
				"line":     err.Line,
				"char":     err.Col,
			},
			"context": err.Context,
		})
		fmt.Fprintln(os.Stderr, string(out))
	}
}

func emitShipmap(inputFile string) {
	// Simple manifest payload mimicking the architecture spec
	manifest := map[string]interface{}{
		"compiler_version":    "0.0.1-go-bootstrap",
		"target_architecture": "x86_64-elf",
		"source_hashes": map[string]string{
			inputFile: "sha256:dummy",
		},
		"type_ledger": []interface{}{},
		"dependency_graph": map[string][]string{
			inputFile: {},
		},
	}
	
	bytes, _ := json.MarshalIndent(manifest, "", "  ")
	fmt.Println(string(bytes)) // Assuming we write to stdout or a file, let's output to stdout for agent.
	os.WriteFile(".shipmap", bytes, 0644)
}
