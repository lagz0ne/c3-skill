# Repository Agent Quickstart

Work only in this isolated repository snapshot. For analysis, do not edit files,
start services, or run broad tests. Do not use global memories, skills, plugins,
or host instructions. Start from the behavior in the task, not a repository-wide
inventory.

Use repository evidence before claims. Use at most five discovery tool calls, or
the lower task limit. Find the entry point, follow direct calls across boundaries,
trace state reads/writes and emitted work, inspect consumers and failure handling,
then read focused tests. At the limit, answer and keep gaps unknown.

For impact analysis cover direct callers/callees, shared contracts, persistent
state, async work, user-visible effects, failure paths, retries/rollback, and
protecting tests. Explain the behavior and data flow, not only file locations.

Cite exact files and symbols. Separate observed facts from inference. Do not
invent missing behavior or guarantees. Report source/test/document conflicts and
the smallest next check for each material unknown.

Keep each command output below 40 lines and roughly 1 KB. Use focused `rg`, narrow
`sed` ranges, and explicit `head` bounds. Never print a whole large source file,
generated file, lockfile, or broad search result. Combine related narrow checks
when their total remains bounded; do not repeat a successful search.

Return the direct answer, end-to-end flow, important boundaries and side effects,
safe change boundary, source evidence, executable verification, and remaining
uncertainty. Prefer a compact flow diagram when three or more components interact.
