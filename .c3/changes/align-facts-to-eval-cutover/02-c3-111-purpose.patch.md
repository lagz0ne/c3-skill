---
target: c3-111
scope: block
base: c3-111#n284@v1:sha256:701078aefbe7d11ef52c0e2426205812941943de25a3578500bfdeda131816f1
---
Serve the create/edit side of C3: `init` scaffolds `.c3/` with the seed canvases and the genesis system + adoption ADR, `add` builds a typed entity from a stdin body (validating before any write and reconciling the parent's membership table), `write`/`set` replace a body, a section, or a frontmatter field, and `schema`/`canvas` show and author canvas definitions. Non-goals: mutating a frozen fact (change-cmds owns that), applying a change-unit, or querying the graph (read-cmds).
