# Test: Discovery-Based Audit

## Setup

Fixture: `fixtures/07-discovery-audit/`

Pre-existing .c3/ docs with deliberate drift:
- c3-1-backend lists component c3-103-legacy (deleted in code)
- Code has new src/cache module (not in inventory)

## Query

```
Audit the C3 documentation.
```

## Expect

### PASS: Drift Detection

| Element | Check |
|---------|-------|
| missing_in_inventory | src/cache detected |
| missing_in_code | c3-103-legacy detected |

### PASS: Report Format

| Element | Check |
|---------|-------|
| Drift report generated | Following audit template |
| Recommendations included | How to fix each drift |
