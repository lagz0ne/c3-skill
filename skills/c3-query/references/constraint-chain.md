# Constraint Chain Query

When user asks about rules, constraints, or boundaries for a layer.

## Flow

```
Identify Layer → Read Upward Chain → Collect Cited Refs → Synthesize Chain
```

## Steps

1. **Identify target layer** from query (c3-NNN, c3-N, or c3-0)

2. **Read upward through hierarchy:**
   - If component: read component → container → context
   - If container: read container → context
   - If context: read context only

3. **At each level, extract:**
   - Explicit constraints ("components MUST...", "MUST NOT...")
   - Implicit boundaries (what the layer is responsible for)
   - Layer-specific rules

4. **Collect cited refs:**
   - From component's `Related Refs` section
   - From container's pattern references
   - From context's system-wide patterns

5. **Read each cited ref** and extract key rules

## Response Format

```
**Constraint Chain for c3-NNN (Component Name)**

**Context Constraints (c3-0):**
- [System-wide rule 1]
- [System-wide rule 2]

**Container Constraints (c3-N):**
- [Container boundary 1]
- [Coordination rule]

**Cited Pattern Constraints:**
- **ref-X:** [Key rules from this pattern]
- **ref-Y:** [Key rules from this pattern]

**Layer Boundaries:**
This component MAY:
- [Permitted responsibilities]

This component MUST NOT:
- [Prohibited actions - from higher layers]

**Visualization:**
[Mermaid diagram showing constraint inheritance]
```

## Example

```
User: "What constraints apply to c3-201?"

1. Read c3-201-auth-middleware.md → cites ref-error-handling, ref-auth
2. Read c3-2-api/README.md → "Middleware must be stateless"
3. Read .c3/README.md → "All auth uses RS256"
4. Read ref-error-handling.md → "Use correlation IDs"
5. Read ref-auth.md → "Tokens expire in 24h"

Response:
**Constraint Chain for c3-201 (Auth Middleware)**

**Context Constraints (c3-0):**
- All authentication must use RS256 signing

**Container Constraints (c3-2-api):**
- Middleware must be stateless
- Request context via headers, not shared state

**Cited Pattern Constraints:**
- **ref-error-handling:** Use correlation IDs in all error responses
- **ref-auth:** Token expiry 24h, refresh via /auth/refresh

**Layer Boundaries:**
This component MAY:
- Validate JWT tokens
- Reject unauthorized requests
- Set request context for downstream handlers

This component MUST NOT:
- Store session state (container: stateless middleware)
- Define new auth schemes (context: RS256 only)
- Handle business logic errors (use ref-error-handling pattern)
```
