---
id: c3-204
c3-version: 3
title: Logger
type: component
category: foundation
parent: c3-2
summary: Pino logger configuration with structured logging and request correlation
---

# Logger

Provides structured logging using Pino with child loggers for different components, consistent log levels, and JSON output for production.

## Contract

| Provides | Expects |
|----------|---------|
| createLogger(name) | Logger name for child |
| logger.info/warn/error | Structured log data |
| entryLogger | Root logger for startup |
| Request correlation | executionId in context |

## Edge Cases

| Scenario | Behavior |
|----------|----------|
| Missing logger name | Uses 'default' |
| Circular object logging | Pino handles safely |
| High log volume | Async writes, no blocking |
| Production mode | JSON format, no pretty-print |

## Testing

| Scenario | Verifies |
|----------|----------|
| Child logger | Name appears in log output |
| Log levels | Only configured level+ logged |
| Structured data | Object logged as JSON fields |
| Error logging | Stack trace included |

## References

- `apps/start/src/server/logger.ts` - Pino configuration
