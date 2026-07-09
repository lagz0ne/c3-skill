#!/usr/bin/env python3
"""Replay the Trace Graph external-target pilot against acountee.

This checker treats acountee as a hostile eval fixture:
- all fixture writes happen under /tmp
- the real acountee checkout must stay unchanged
- target c3 check drift is surfaced, not repaired or hidden
- query/RAG still has to return a useful context pack for PR approvals
"""

from __future__ import annotations

import json
import os
import hashlib
import re
import shutil
import subprocess
import tempfile
from dataclasses import dataclass
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
C3_WRAPPER = ROOT / "skills/c3/bin/c3x.sh"
DEFAULT_TARGET = Path(os.environ.get("ACOUNTEE_REPO", "/home/lagz0ne/dev/acountee"))


@dataclass(frozen=True)
class LookupExpectation:
    path: str
    facts: tuple[str, ...]


@dataclass(frozen=True)
class SourceAnchor:
    path: str
    symbols: tuple[str, ...]


QUESTION = "How does PR approval/approval-chain work, and where would a fix start?"
EXPECTED_FACTS = (
    "ref-approval-chain",
    "c3-105",
    "c3-205",
    "ui-flow-pr-approval-flow",
    "ui-state-machine-payment-request-lifecycle",
)
EXPECTED_SEARCH_HITS = (
    "c3-105",
    "c3-205",
    "ref-approval-chain",
    "ui-flow-pr-approval-flow",
    "ui-state-machine-payment-request-lifecycle",
)
EXPECTED_GRAPH_NEIGHBORS = (
    "c3-105",
    "c3-205",
    "c3-210",
    "ref-approval-chain",
)
LOOKUPS = (
    LookupExpectation(
        "apps/start/src/screens/PaymentRequestsScreen.tsx",
        ("c3-105", "ref-approval-chain"),
    ),
    LookupExpectation(
        "apps/start/src/server/flows/pr.ts",
        ("c3-205", "ref-approval-chain", "ref-pumped-fn"),
    ),
    LookupExpectation(
        "apps/start/src/server/dbs/queries/approval.ts",
        ("ref-approval-chain", "ref-query-services"),
    ),
    LookupExpectation(
        "packages/shared/src/approval.ts",
        ("ref-approval-chain",),
    ),
)
SOURCE_ANCHORS = (
    SourceAnchor(
        "apps/start/src/screens/PaymentRequestsScreen.tsx",
        ("approvalFlowSchema", "requestForApprovals", "canApprove", "selectedApprovalFlow"),
    ),
    SourceAnchor(
        "apps/start/src/server/flows/pr.ts",
        ("createPr", "requestApprovals", "approvePr", "approveAll", "notifyNextApprovers"),
    ),
    SourceAnchor(
        "apps/start/src/server/dbs/queries/approval.ts",
        ("approvalQueries", "updateApprovalCurrentStep", "approvalRecords"),
    ),
    SourceAnchor(
        "packages/shared/src/approval.ts",
        ("approvalFlowSchema", "ApprovalFlow", "stepsToApprovalFlow"),
    ),
)


class CheckError(Exception):
    pass


def fail(message: str) -> None:
    print(json.dumps({"ok": False, "error": message}, indent=2, sort_keys=True))
    raise SystemExit(1)


def run(
    cmd: list[str],
    *,
    cwd: Path | None = None,
    env: dict[str, str] | None = None,
) -> subprocess.CompletedProcess[str]:
    proc = subprocess.run(
        cmd,
        cwd=cwd,
        env=env,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        check=False,
    )
    return subprocess.CompletedProcess(
        proc.args,
        proc.returncode,
        stdout=proc.stdout.decode("utf-8", errors="replace"),
        stderr=proc.stderr.decode("utf-8", errors="replace"),
    )


def git_status(repo: Path) -> str:
    proc = run(["git", "-C", str(repo), "status", "--short", "--branch"])
    if proc.returncode != 0:
        raise CheckError(f"target git status failed: {proc.stderr.strip()}")
    return proc.stdout


def git_status_with_ignored(repo: Path) -> str:
    proc = run(["git", "-C", str(repo), "status", "--short", "--branch", "--ignored", "-uall"])
    if proc.returncode != 0:
        raise CheckError(f"target git ignored-status failed: {proc.stderr.strip()}")
    return proc.stdout


def ensure_target(repo: Path) -> None:
    if not repo.exists():
        raise CheckError(f"missing acountee target repo: {repo}")
    if not (repo / ".c3").is_dir():
        raise CheckError(f"target repo has no .c3 directory: {repo}")
    if not C3_WRAPPER.exists():
        raise CheckError(f"missing local C3 wrapper: {C3_WRAPPER.relative_to(ROOT)}")


def make_fixture(repo: Path) -> tuple[Path, Path]:
    tmp_root = Path(tempfile.mkdtemp(prefix="acountee-trace-graph-eval.", dir="/tmp"))
    project = tmp_root / "acountee"
    project.mkdir()

    shutil.copytree(repo / ".c3", project / ".c3")
    for name in (
        "apps",
        "packages",
        "scripts",
        "package.json",
        "pnpm-workspace.yaml",
        "tsconfig.json",
    ):
        source = repo / name
        if source.exists():
            os.symlink(source, project / name)

    return tmp_root, project


def c3(project: Path, *args: str) -> subprocess.CompletedProcess[str]:
    env = os.environ.copy()
    env["C3X_MODE"] = "agent"
    return run(
        ["bash", str(C3_WRAPPER), "--c3-dir", str(project / ".c3"), *args],
        cwd=project,
        env=env,
    )


def require_success(label: str, proc: subprocess.CompletedProcess[str]) -> str:
    if proc.returncode != 0:
        raise CheckError(
            f"{label} failed with exit {proc.returncode}: "
            f"{(proc.stdout + proc.stderr).strip()}"
        )
    return proc.stdout


def contains_signal(text: str, needle: str) -> bool:
    if re.fullmatch(r"[A-Za-z0-9_-]+", needle):
        return bool(re.search(rf"(?<![A-Za-z0-9_-]){re.escape(needle)}(?![A-Za-z0-9_-])", text))
    return needle in text


def require_contains(label: str, text: str, needles: tuple[str, ...]) -> list[str]:
    missing = [needle for needle in needles if not contains_signal(text, needle)]
    if missing:
        raise CheckError(f"{label} missing expected signal(s): {missing}")
    return list(needles)


def contains_symbol(text: str, symbol: str) -> bool:
    return bool(re.search(rf"(?<![A-Za-z0-9_$]){re.escape(symbol)}(?![A-Za-z0-9_$])", text))


def require_source_anchors(project: Path) -> list[dict[str, object]]:
    found: list[dict[str, object]] = []
    missing: list[str] = []
    for anchor in SOURCE_ANCHORS:
        source = project / anchor.path
        if not source.exists():
            missing.append(anchor.path)
            continue
        body = source.read_text(encoding="utf-8", errors="ignore")
        symbols: list[str] = []
        for symbol in anchor.symbols:
            if contains_symbol(body, symbol):
                symbols.append(symbol)
            else:
                missing.append(f"{anchor.path}:{symbol}")
        if symbols:
            found.append({"path": anchor.path, "symbols": symbols})
    if missing:
        raise CheckError(f"missing source anchor(s): {missing}")
    return found


def summarize_check(proc: subprocess.CompletedProcess[str]) -> list[str]:
    lines = (proc.stdout + "\n" + proc.stderr).splitlines()
    keep = []
    for line in lines:
        if (
            "ok:" in line
            or "only_in_tree:" in line
            or "missing_from_tree:" in line
            or "canonical markdown drift" in line
            or "hint:" in line
        ):
            keep.append(line.strip())
    return keep[:8]


def stable_trace_hash(pack: dict[str, object]) -> str:
    canonical = {
        "hash_version": pack["hash_version"],
        "graph_id": pack["graph_id"],
        "nodes": pack["nodes"],
        "edges": pack["edges"],
        "hash_basis": pack["hash_basis"],
    }
    body = json.dumps(canonical, sort_keys=True, separators=(",", ":"))
    return hashlib.sha256(body.encode("utf-8")).hexdigest()


def build_trace_pack(
    *,
    project: Path,
    check_summary: list[str],
    search_ids: list[str],
    fact_ids: list[str],
    graph_ids: list[str],
    lookup_results: list[dict[str, object]],
    source_results: list[dict[str, object]],
) -> dict[str, object]:
    lookup_paths = [str(item["path"]) for item in lookup_results]
    lookup_facts = sorted({fact for item in lookup_results for fact in item["facts"]})
    source_paths = [str(item["path"]) for item in source_results]
    source_symbols = sorted({symbol for item in source_results for symbol in item["symbols"]})

    nodes = [
        {
            "id": "target-c3-drift",
            "kind": "validation",
            "clue": "Target C3 check is red and must be carried as drift context.",
            "owners": [],
            "anchors": {"check_status": "red", "summary": check_summary},
        },
        {
            "id": "approval-query-entry",
            "kind": "entrypoint",
            "clue": QUESTION,
            "owners": ["ref-approval-chain"],
            "anchors": {"query": "PR approval chain request approvals anyof allof"},
        },
        {
            "id": "c3-search-candidates",
            "kind": "binding",
            "clue": "Keyword plus graph search returns candidate facts and custom docs.",
            "owners": search_ids,
            "anchors": {"c3_ids": search_ids},
        },
        {
            "id": "approval-graph-neighbors",
            "kind": "policy",
            "clue": "The approval-chain ref fans out to UI, backend PR flows, and admin configuration.",
            "owners": graph_ids,
            "anchors": {"c3_ids": graph_ids},
        },
        {
            "id": "code-lookup-anchors",
            "kind": "binding",
            "clue": "Lookup maps first inspection files back to governing facts and refs.",
            "owners": lookup_facts,
            "anchors": {"paths": lookup_paths},
        },
        {
            "id": "source-symbol-anchors",
            "kind": "output",
            "clue": "Stable symbols name the first source surfaces to inspect before broad grep.",
            "owners": ["ref-approval-chain", "c3-105", "c3-205"],
            "anchors": {"paths": source_paths, "symbols": source_symbols},
        },
    ]
    edges = [
        "target-c3-drift->approval-query-entry:governed-by",
        "approval-query-entry->c3-search-candidates:routes-to",
        "c3-search-candidates->approval-graph-neighbors:routes-to",
        "approval-graph-neighbors->code-lookup-anchors:resolves",
        "code-lookup-anchors->source-symbol-anchors:resolves",
    ]
    pack: dict[str, object] = {
        "hash_version": 1,
        "graph_id": "trace-acountee-pr-approval-context-pack",
        "question": QUESTION,
        "target": "acountee",
        "status": "context_pack",
        "nodes": nodes,
        "edges": edges,
        "hash_basis": [
            "graph_id",
            "hash_version",
            "node.id",
            "node.kind",
            "node.clue",
            "node.owners",
            "anchors.c3_ids",
            "anchors.paths",
            "anchors.symbols",
            "edges",
        ],
        "context_pack": {
            "drift": check_summary,
            "search_candidate_ids": search_ids,
            "fact_ids": fact_ids,
            "graph_neighbor_ids": graph_ids,
            "lookup_anchors": lookup_results,
            "source_anchors": source_results,
        },
    }
    pack["trace_hash"] = stable_trace_hash(pack)

    out = project / "trace-pack.pr-approval.json"
    out.write_text(json.dumps(pack, indent=2, sort_keys=True) + "\n", encoding="utf-8")
    pack["artifact_path"] = str(out)
    return pack


def validate_trace_pack(pack: dict[str, object]) -> None:
    expected_node_ids = {
        "target-c3-drift",
        "approval-query-entry",
        "c3-search-candidates",
        "approval-graph-neighbors",
        "code-lookup-anchors",
        "source-symbol-anchors",
    }
    nodes = pack.get("nodes")
    if not isinstance(nodes, list):
        raise CheckError("trace pack missing nodes")
    node_ids = {node.get("id") for node in nodes if isinstance(node, dict)}
    if node_ids != expected_node_ids:
        raise CheckError(f"trace pack node set mismatch: {sorted(node_ids)}")

    edges = pack.get("edges")
    if not isinstance(edges, list) or len(edges) < 5:
        raise CheckError("trace pack missing edge chain")

    hash_basis = " ".join(str(item).lower() for item in pack.get("hash_basis", []))
    forbidden = ("full file", "line number", "whole function", "function body", "formatting")
    bad = [term for term in forbidden if term in hash_basis]
    if bad:
        raise CheckError(f"trace pack has noisy hash basis: {bad}")

    context = pack.get("context_pack")
    if not isinstance(context, dict):
        raise CheckError("trace pack missing context_pack")
    search_candidate_ids = context.get("search_candidate_ids", [])
    lookup_anchors = context.get("lookup_anchors", [])
    source_anchors = context.get("source_anchors", [])
    if not isinstance(search_candidate_ids, list) or len(search_candidate_ids) < 5:
        raise CheckError("trace pack has too few search candidates")
    if not isinstance(lookup_anchors, list) or len(lookup_anchors) < 4:
        raise CheckError("trace pack has too few lookup anchors")
    if not isinstance(source_anchors, list) or len(source_anchors) < 4:
        raise CheckError("trace pack has too few source anchors")
    if not isinstance(pack.get("trace_hash"), str) or len(str(pack["trace_hash"])) != 64:
        raise CheckError("trace pack missing stable hash")


def validate(project: Path) -> dict[str, int | float | str | list[str]]:
    check_proc = c3(project, "check")
    check_text = check_proc.stdout + check_proc.stderr
    drift_visible = (
        check_proc.returncode != 0
        and "ok: false" in check_text
        and any(term in check_text for term in ("drift", "only_in_tree", "missing_from_tree", "sync check failed"))
    )
    if not drift_visible:
        raise CheckError("expected acountee fixture to surface C3 drift without repair")
    check_summary = summarize_check(check_proc)

    search = require_success(
        "c3 search",
        c3(
            project,
            "search",
            "PR approval chain request approvals anyof allof",
            "--no-semantic",
            "--limit",
            "8",
        ),
    )
    search_ids = require_contains("search results", search, EXPECTED_SEARCH_HITS)

    fact_ids: list[str] = []
    for fact in EXPECTED_FACTS:
        out = require_success(f"c3 read {fact}", c3(project, "read", fact))
        if f"id: {fact}" not in out:
            raise CheckError(f"c3 read {fact} did not return the expected id")
        fact_ids.append(fact)

    graph = require_success("c3 graph ref-approval-chain", c3(project, "graph", "ref-approval-chain", "--depth", "1"))
    graph_ids = require_contains("graph results", graph, EXPECTED_GRAPH_NEIGHBORS)

    lookup_signals = 0
    lookup_results: list[dict[str, object]] = []
    for lookup in LOOKUPS:
        out = require_success(f"c3 lookup {lookup.path}", c3(project, "lookup", lookup.path))
        found = require_contains(f"lookup {lookup.path}", out, lookup.facts)
        lookup_signals += len(found)
        lookup_results.append({"path": lookup.path, "facts": found})

    source_results = require_source_anchors(project)
    source_anchor_signals = sum(len(item["symbols"]) for item in source_results)
    trace_pack = build_trace_pack(
        project=project,
        check_summary=check_summary,
        search_ids=search_ids,
        fact_ids=fact_ids,
        graph_ids=graph_ids,
        lookup_results=lookup_results,
        source_results=source_results,
    )
    validate_trace_pack(trace_pack)

    return {
        "external_target_count": 1,
        "target_c3_check_red_visible_count": 1,
        "target_repair_or_mutation_count": 0,
        "query_reasoning_context_pack_count": 1,
        "search_candidate_signal_count": len(search_ids),
        "c3_fact_signal_count": len(fact_ids),
        "graph_neighbor_signal_count": len(graph_ids),
        "lookup_anchor_signal_count": lookup_signals,
        "source_anchor_signal_count": source_anchor_signals,
        "trace_context_pack_node_count": len(trace_pack["nodes"]),
        "trace_context_pack_edge_count": len(trace_pack["edges"]),
        "trace_graph_acountee_decision_readiness": 1.0,
        "trace_pack_hash": str(trace_pack["trace_hash"]),
        "trace_pack_artifact": str(trace_pack["artifact_path"]),
        "target_check_summary": check_summary,
    }


def main() -> None:
    repo = DEFAULT_TARGET
    tmp_root: Path | None = None
    metrics: dict[str, int | float | str | list[str]] | None = None
    error: CheckError | None = None
    before_status: str | None = None
    before_ignored_status: str | None = None
    try:
        ensure_target(repo)
        before_status = git_status(repo)
        before_ignored_status = git_status_with_ignored(repo)
        tmp_root, project = make_fixture(repo)
        metrics = validate(project)
    except CheckError as exc:
        error = exc

    try:
        after_status = git_status(repo)
        after_ignored_status = git_status_with_ignored(repo)
        if before_status is not None and after_status != before_status:
            error = CheckError("real acountee visible git status changed during external eval")
        if before_ignored_status is not None and after_ignored_status != before_ignored_status:
            error = CheckError("real acountee ignored-file git status changed during external eval")
        if metrics is not None:
            metrics["target_visible_status_unchanged_count"] = 1
            metrics["target_ignored_status_unchanged_count"] = 1
    except CheckError as exc:
        error = exc

    if error is not None:
        fail(str(error))
    if tmp_root is None or metrics is None:
        fail("external eval did not produce metrics")

    result = {
        "ok": True,
        "question": QUESTION,
        "target_repo": str(repo),
        "tmp_workspace": str(tmp_root),
        "cleanup": f"rm -rf {tmp_root}",
        "metrics": metrics,
    }
    print(json.dumps(result, indent=2, sort_keys=True))


if __name__ == "__main__":
    main()
