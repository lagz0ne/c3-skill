---
target: recipe-change-unit-saga
scope: whole
type: recipe
title: Change-Unit Saga
affects: c3-112, c3-104, c3-103, c3-102
---
# Change-Unit Saga

## Goal

Trace the only legal mutation of a frozen fact: change-cmds (c3-112) scaffolds a unit and authors patches, changeset (c3-104) runs the apply gates (drift, canvas, morph, retire, inspection), schema (c3-103) validates each merged body, and store (c3-102) commits every patch, edge, membership row, and seal in one atomic transaction — all-or-nothing.
