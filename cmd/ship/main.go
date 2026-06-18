package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/nathfavour/ship/internal/emitter/elf"
	"github.com/nathfavour/ship/internal/ir"
	"github.com/nathfavour/ship/internal/lexer"
	"github.com/nathfavour/ship/internal/parser"
	"github.com/nathfavour/ship/internal/types"
)

const version = "0.0.1-go-bootstrap"

func printUsage() {
	fmt.Println(`Ship: The Agentic Era Systems Compiler

Usage:
  ship <command> [arguments]

The commands are:

  build       Compile ship source code into an executable
  run         Compile and run ship source code immediately
  version     Print ship compiler version
  help        Print this help message

Use "ship <command> -h" for more information about a command.`)
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "build":
		buildCmd()
	case "run":
		runCmd()
	case "version":
		fmt.Printf("ship version %s\n", version)
	case "help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "ship: unknown command %q\n", command)
		printUsage()
		os.Exit(1)
	}
}

func buildCmd() {
	buildFlags := flag.NewFlagSet("build", flag.ExitOnError)
	agentFlag := buildFlags.Bool("agent", false, "Enable machine-first diagnostic stream (JSON)")
	outputFlag := buildFlags.String("o", "a.out", "Output executable file")

	buildFlags.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: ship build [options] <file.ship>")
		buildFlags.PrintDefaults()
	}

	buildFlags.Parse(os.Args[2:])

	if buildFlags.NArg() == 0 {
		buildFlags.Usage()
		os.Exit(1)
	}

	inputFile := buildFlags.Arg(0)
	compile(inputFile, *outputFlag, *agentFlag, false)
}

func runCmd() {
	runFlags := flag.NewFlagSet("run", flag.ExitOnError)
	agentFlag := runFlags.Bool("agent", false, "Enable machine-first diagnostic stream (JSON)")

	runFlags.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: ship run [options] <file.ship>")
		runFlags.PrintDefaults()
	}

	runFlags.Parse(os.Args[2:])

	if runFlags.NArg() == 0 {
		runFlags.Usage()
		os.Exit(1)
	}

	inputFile := runFlags.Arg(0)
	outputFile := filepath.Join(os.TempDir(), "ship_run_"+filepath.Base(inputFile)+".out")

	// Compile silently unless agent flag is used (but run has its own output)
	compile(inputFile, outputFile, *agentFlag, true)

	// Execute the compiled binary
	cmd := exec.Command(outputFile)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	err := cmd.Run()

	// Clean up the temporary binary
	os.Remove(outputFile)

	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			os.Exit(exitError.ExitCode())
		}
		fmt.Fprintf(os.Stderr, "Execution failed: %s\n", err)
		os.Exit(1)
	}
}

func compile(inputFile, outputFile string, agent bool, isRunCommand bool) {
	bytes, err := os.ReadFile(inputFile)
	if err != nil {
		fatalError(fmt.Sprintf("Could not read file: %s", err), agent)
	}

	input := string(bytes)

	// 1. Lex
	l := lexer.New(inputFile, input)

	// 2. Parse
	p := parser.New(l)
	astFile := p.ParseFile()

	if len(p.Errors()) > 0 {
		if agent {
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
		if agent {
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
		fatalError(fmt.Sprintf("Failed to emit ELF: %s", err), agent)
	}

	err = os.WriteFile(outputFile, binaryBytes, 0755)
	if err != nil {
		fatalError(fmt.Sprintf("Failed to write output binary: %s", err), agent)
	}

	if agent {
		// Emit mock .shipmap for agentic system
		emitShipmap(inputFile)
	} else if !isRunCommand {
		// Only print success msg if we are explicitly running `build` (not `run`) and not in agent mode
		fmt.Printf("Successfully compiled %s -> %s\n", inputFile, outputFile)
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
	manifest := map[string]interface{}{
		"compiler_version":    version,
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
	fmt.Println(string(bytes)) // Standard stdout streaming for --agent
	os.WriteFile(".shipmap", bytes, 0644)
}
