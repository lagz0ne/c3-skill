# Autoresearch: 100% Test Coverage

## Config
- **Benchmark**: `bash autoresearch.sh`
- **Target metric**: `coverage` (higher is better)
- **Scope**: `cli/` — all Go packages
- **Branch**: `autoresearch/100-percent-test-coverage`
- **Started**: 2026-03-20T00:00:00Z

## Rules
1. One change per experiment
2. Run benchmark after every change
3. Keep if metric improves, discard if it regresses
4. Log every run to autoresearch.jsonl
5. Commit kept changes with `Result:` trailer

## Strategy
Priority order for maximum coverage gain per test:
1. Functions at 0% coverage (biggest bang per test)
2. Functions below 50% (high marginal gain)
3. Functions 50-80% (medium gain)
4. Functions 80-95% (fill gaps)
5. main.go entry point (may need refactoring or exclusion)
