---
id: c3-120
c3-seal: 0b85ede33eeb2eda7a668f85b3be6d797b4172325694bdf7b59d2ead26149b2b
title: history-marketplace-cmds
type: component
category: feature
parent: c3-1
goal: Handle versions, hash, nodes, prune, and marketplace command families.
---

# history-marketplace-cmds
## Goal

Handle versions, hash, nodes, prune, and marketplace command families.

## Dependencies

| Direction | What | From/To |
| --- | --- | --- |
| IN | Persistence APIs | c3-107 |
| IN | Marketplace backend |  |
| OUT | History inspection and marketplace flows |  |
## Container Connection

Covers advanced command families outside core topology and mutation flows.
