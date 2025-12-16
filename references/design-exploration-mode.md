# C3 Design - Exploration Mode

Read-only navigation of existing C3 documentation.

## When to Use

- "What's the architecture?"
- "How does X work?"
- "Why did we choose X?"

## Process

1. **Load TOC:** `cat .c3/TOC.md`

2. **Parse for overview:**
   - Document counts
   - Container summaries
   - ADR list

3. **Present based on intent:**

   **General orientation:**
   ```
   ## {System} Architecture
   **Purpose:** {from Context}
   ### At a Glance
   - Containers: N
   - Components: N
   - Decisions: N ADRs
   ### Container Overview
   | ID | Container | Purpose |
   ```

   **Focused:** Load relevant docs for user's area.

   **Decisions:** Present ADR list, load specific ones.

4. **Navigate on demand:**
   - "Tell me about c3-1" → Load container
   - "Why X?" → Search ADRs
   - "I need to change Y" → Switch to Design Mode
