---
id: adr-20260415-mutation-preverify-repair-bypass
c3-seal: 2be7ffbcfef3fc061f3f5510f6878ce4d22541fc2cc6d2f4265c05e1c84ffae8
title: mutation-preverify-repair-bypass
type: adr
goal: Allow mutating C3 commands such as add, write, set, wire, delete, codemap, and migrate to reach their own payload validation and canonical export paths even when existing canonical docs fail preverification, while keeping read-only commands gated by verify.
status: implemented
date: "2026-04-15"
---

# mutation-preverify-repair-bypass

## Goal

Allow mutating C3 commands such as add, write, set, wire, delete, codemap, and migrate to reach their own payload validation and canonical export paths even when existing canonical docs fail preverification, while keeping read-only commands gated by verify.
