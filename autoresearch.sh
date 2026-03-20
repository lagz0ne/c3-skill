#!/usr/bin/env bash
set -euo pipefail
# Autoresearch benchmark: Go test coverage
cd "$(dirname "$0")/cli"

# Run tests with coverage
go test ./... -coverprofile=coverage.out -count=1 2>&1 >/dev/null

# Get total coverage
TOTAL=$(go tool cover -func=coverage.out | tail -1 | awk '{print $NF}' | tr -d '%')

# Count per-package coverage
PKGS_AT_100=$(go tool cover -func=coverage.out | grep -E '^\S+\s+\S+\s+100\.0%' | wc -l)

# Count total functions
TOTAL_FUNCS=$(go tool cover -func=coverage.out | grep -v '^total:' | wc -l)
UNCOVERED=$(go tool cover -func=coverage.out | grep -v '^total:' | grep -v '100.0%' | wc -l)

echo "METRIC coverage=$TOTAL"
echo "METRIC funcs_at_100=$((TOTAL_FUNCS - UNCOVERED))"
echo "METRIC total_funcs=$TOTAL_FUNCS"
echo "METRIC uncovered_funcs=$UNCOVERED"
