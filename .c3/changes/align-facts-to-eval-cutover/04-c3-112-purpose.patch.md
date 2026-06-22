---
target: c3-112
scope: block
base: c3-112#n311@v1:sha256:978a582b8e710cea4df539763f6b3e2e10ba311ca4939c52e6ef52266740dcb8
---
Serve the change-unit command surface: `change new`/`scaffold` stage a unit and its climb patches, `change view`/`status` project each patch's drift and apply state, and `change apply` runs the gates (drift, canvas, morph, retire) before committing atomically. Around them: the freeze guard refuses a direct write/set/delete of a frozen fact, `supersede` flips a terminal decision under a successor, the overlay previews the graph as a unit would leave it, and `conflict`/`materialize` support rebase and seed materialization. Non-goals: the apply transaction's internals (changeset), pre-freeze authoring (author-cmds), or read-only queries (read-cmds).
