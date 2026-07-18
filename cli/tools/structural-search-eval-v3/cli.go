package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// The CLI intentionally exposes only the read-only fresh baseline capture.
// Candidate, capability, activation, authorization, and effect commands do not
// exist in this package.
func main() {
	if err := runV3CLI(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runV3CLI(args []string) error {
	if len(args) == 0 || args[0] != "capture" {
		return errors.New("usage: structural-search-eval-v3 capture --fixtures <path> --benchmark <path> --out <path>")
	}
	fixturesPath, err := requiredV3Arg(args[1:], "--fixtures")
	if err != nil {
		return err
	}
	benchmarkPath, err := requiredV3Arg(args[1:], "--benchmark")
	if err != nil {
		return err
	}
	outPath, err := requiredV3Arg(args[1:], "--out")
	if err != nil {
		return err
	}
	fixtures, fixtureHash, err := LoadV3FixtureFile(fixturesPath)
	if err != nil {
		return err
	}
	bench, benchmarkHash, err := LoadV3BenchmarkFile(benchmarkPath)
	if err != nil {
		return err
	}
	if bench.FixtureCount != len(fixtures) {
		return fmt.Errorf("benchmark fixture_count=%d, fixtures=%d", bench.FixtureCount, len(fixtures))
	}
	artifact, err := CaptureFreshBaseline(fixtures)
	if err != nil {
		return err
	}
	artifact.FixtureSHA256 = fixtureHash
	artifact.BenchmarkSHA256 = benchmarkHash
	artifact.ScorerSHA256 = sha256File(filepathFromRoot("cli/tools/structural-search-eval-v3/main.go"))
	// Bind the retained artifact to the requested output name. This keeps the
	// v3 default byte-compatible while making the same generic capture command
	// safe for separately versioned benchmark arms.
	artifact.BenchmarkBaselineFile = filepath.Base(outPath)
	artifact.Privacy = bench.Privacy
	if err := WriteV3Baseline(outPath, artifact); err != nil {
		return err
	}
	return nil
}

func requiredV3Arg(args []string, name string) (string, error) {
	for i := 0; i+1 < len(args); i++ {
		if args[i] == name && args[i+1] != "" {
			return args[i+1], nil
		}
	}
	return "", fmt.Errorf("%s is required", name)
}
