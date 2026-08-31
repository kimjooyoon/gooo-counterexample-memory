package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/kimjooyoon/gooo-counterexample-memory/internal/memory"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		usage(stderr)
		return 2
	}
	switch args[0] {
	case "compile":
		return compile(args[1:], stdout, stderr)
	case "evaluate":
		return evaluate(args[1:], stdout, stderr)
	case "version":
		fmt.Fprintln(stdout, "gooo-counterexample-memory/v1")
		return 0
	default:
		usage(stderr)
		return 2
	}
}

func compile(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("compile", flag.ContinueOnError)
	flags.SetOutput(stderr)
	sourcePath := flags.String("source", "examples/counterexample-memory/main.gooo", "Gooo source")
	contractPath := flags.String("contract", "contracts/counterexample-memory-denominator-v1.json", "fixed denominator")
	outputPath := flags.String("output", "semantic-ir.json", "semantic IR output")
	if err := flags.Parse(args); err != nil || flags.NArg() != 0 {
		return 2
	}
	source, err := os.ReadFile(*sourcePath)
	if err != nil {
		fmt.Fprintf(stderr, "read source: %v\n", err)
		return 1
	}
	denominator, err := memory.LoadDenominator(*contractPath)
	if err != nil {
		fmt.Fprintf(stderr, "read denominator: %v\n", err)
		return 1
	}
	ir, err := memory.CompileSource(*sourcePath, source, denominator)
	if err != nil {
		fmt.Fprintf(stderr, "compile source: %v\n", err)
		return 1
	}
	if writeJSON(*outputPath, ir, stderr) != 0 {
		fmt.Fprintln(stderr, "write IR failed")
		return 1
	}
	return writeJSON(stdout, ir, stderr)
}

func evaluate(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("evaluate", flag.ContinueOnError)
	flags.SetOutput(stderr)
	sourcePath := flags.String("source", "examples/counterexample-memory/main.gooo", "Gooo source")
	contractPath := flags.String("contract", "contracts/counterexample-memory-denominator-v1.json", "fixed denominator")
	irPath := flags.String("ir", "semantic-ir.json", "semantic IR")
	corpusPath := flags.String("corpus", "fixtures/corpus/evidence.ndjson", "append-only corpus")
	casesPath := flags.String("cases", "fixtures/cases", "controlled cases")
	outputDir := flags.String("output-dir", "", "caller-owned output directory")
	if err := flags.Parse(args); err != nil || flags.NArg() != 0 || *outputDir == "" {
		fmt.Fprintln(stderr, "evaluate requires --output-dir and accepts no positional arguments")
		return 2
	}
	meta, err := memory.LoadMeta(*sourcePath, *contractPath, *irPath)
	if err != nil {
		fmt.Fprintf(stderr, "load semantic bindings: %v\n", err)
		return 1
	}
	records, corpusDigest, err := memory.LoadCorpus(*corpusPath)
	if err != nil {
		fmt.Fprintf(stderr, "load corpus: %v\n", err)
		return 1
	}
	cases, err := memory.LoadCases(*casesPath)
	if err != nil {
		fmt.Fprintf(stderr, "load cases: %v\n", err)
		return 1
	}
	report := memory.Evaluate(meta, *corpusPath, records, corpusDigest, cases)
	if err := os.MkdirAll(*outputDir, 0o755); err != nil {
		fmt.Fprintf(stderr, "create output directory: %v\n", err)
		return 1
	}
	if writeJSON(filepath.Join(*outputDir, "evaluation.json"), report, stderr) != 0 {
		fmt.Fprintln(stderr, "write evaluation failed")
		return 1
	}
	return writeJSON(stdout, report, stderr)
}

func writeJSON(target any, value any, extras ...io.Writer) int {
	var writer io.Writer
	switch typed := target.(type) {
	case string:
		if err := os.MkdirAll(filepath.Dir(typed), 0o755); err != nil {
			return 1
		}
		file, err := os.Create(typed)
		if err != nil {
			return 1
		}
		defer file.Close()
		writer = file
	case io.Writer:
		writer = typed
	default:
		return 1
	}
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(value); err != nil {
		if len(extras) > 0 {
			fmt.Fprintf(extras[0], "emit JSON: %v\n", err)
		}
		return 1
	}
	return 0
}

func usage(stderr io.Writer) {
	fmt.Fprintln(stderr, "usage: gooo-counterexample-memory compile --source PATH --contract PATH --output PATH")
	fmt.Fprintln(stderr, "       gooo-counterexample-memory evaluate --source PATH --contract PATH --ir PATH --corpus PATH --cases PATH --output-dir PATH")
}
