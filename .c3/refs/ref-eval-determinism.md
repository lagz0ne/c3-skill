---
id: ref-eval-determinism
c3-seal: 8c2fd41462e0f43d0f0fb88135d236b4c51a968d55d5a588d8f4db22db81c898
title: eval-determinism
type: ref
goal: A conformance verdict must mean the same thing every time it is computed for an unchanged subject. Without a reproducibility rule an eval that "holds" today could read "drift" tomorrow with no change to the fact or the code it governs — turning the conformance signal into noise no one can trust.
---

# Eval Determinism

## Goal

A conformance verdict must mean the same thing every time it is computed for an unchanged subject. Without a reproducibility rule an eval that "holds" today could read "drift" tomorrow with no change to the fact or the code it governs — turning the conformance signal into noise no one can trust.

## Choice

A verdict is stamped to the exact (claim, external-state) pair it measured: the engine hashes the gathered frame into an ExternalState, and the verdict is solid only for that hash. Every gather source admitted into a pipeline must therefore be deterministic — the same unchanged subject must always produce a byte-identical frame. Membership selectors are allowed (globs, id-globs, the file tree) because the selected set IS the external being asserted; ranked retrieval is not.

## Why

The external a fact governs (code, an artifact) is not frozen, so the engine cannot assume stability — it must instead guarantee that, given an unchanged subject, it gathers an identical frame. A ranked or search-based source breaks this by construction: its result set is a function of the whole corpus and a truncation limit, so an unrelated change re-ranks the list and shifts the checked set while the governed subject sits still. The verdict would then move with nothing real having changed — exactly the false signal this ref exists to prevent. A deterministic selector does not have this failure mode: when its output shifts, it is because the asserted subject itself shifted, which is a legitimate drift.

## How

The terminal stamp is `ExternalState = sha256(join(frame))` in cli/internal/eval/eval.go. A gather source is admissible iff, for an unchanged subject, `frame` is byte-identical across runs — so file, command, glob, and id-glob selectors qualify (sorted for stability) and ranked `search` does not. Discovery may use ranked search at authoring time to find what to bind; the bind it emits into the spec must be a deterministic selector.
