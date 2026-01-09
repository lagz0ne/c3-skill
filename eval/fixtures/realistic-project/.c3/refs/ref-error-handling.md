---
id: ref-error-handling
title: Error Handling
---

# Error Handling

## Goal

Establish conventions for catching, displaying, and recovering from errors in the UI using React ErrorBoundary with retry support.

## Conventions

| Rule | Why |
|------|-----|
| Wrap feature screens in ErrorBoundary | Prevents full app crash on component error |
| Use ErrorDisplay for user-facing errors | Consistent styling and messaging |
| Provide retry mechanism with maxRetries | Allows recovery from transient failures |
| Log errors to console in componentDidCatch | Debugging visibility in development |
| Pass onError callback for telemetry | Enables error tracking integration |
| Show fallback UI during error state | User understands something went wrong |
| InitErrorDisplay for SSR/hydration errors | Special handling for initial load failures |

## Testing

| Convention | How to Test |
|------------|-------------|
| Error catching | Throw in child component, verify boundary catches |
| Retry mechanism | Click retry, verify component re-mounts |
| Max retries limit | Exhaust retries, verify retry disabled |
| Error display | Trigger error, verify ErrorDisplay renders |
| onError callback | Trigger error, verify callback invoked with error |

## References

- `apps/start/src/components/Error*.tsx` - Error boundary and display components
