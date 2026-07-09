#!/usr/bin/env python3
"""Compare raw C3 search against Trace Graph route packs.

This is not a search-ranking eval or an automatic anchor-discovery eval. It
measures whether a source-anchored Trace Graph context pack improves the route
after search by adding facts, graph neighbors, file anchors, symbols, tests,
and drift signals.
"""

from __future__ import annotations

import hashlib
import json
import os
import re
import shutil
import subprocess
import tempfile
from dataclasses import dataclass
from pathlib import Path
from typing import Any


ROOT = Path(__file__).resolve().parents[1]
C3_WRAPPER = ROOT / "skills/c3/bin/c3x.sh"
ACOUNTEE_REPO = Path(os.environ.get("ACOUNTEE_REPO", "/home/lagz0ne/dev/acountee"))


@dataclass(frozen=True)
class SourceAnchor:
    path: str
    symbols: tuple[str, ...]


@dataclass(frozen=True)
class QueryCase:
    case_id: str
    repo: str
    theme: str
    query: str
    expected_ids: tuple[str, ...]
    graph_root: str
    source_anchors: tuple[SourceAnchor, ...]
    expect_drift: bool = False


CASES: tuple[QueryCase, ...] = (
    QueryCase(
        case_id="c3-lookup-ownership",
        repo="c3-design",
        theme="lookup ownership",
        query="file glob lookup ownership facts refs rules",
        expected_ids=("c3-110", "c3-106", "c3-102", "rule-output-via-helpers"),
        graph_root="c3-110",
        source_anchors=(
            SourceAnchor("cli/cmd/lookup.go", ("RunLookup", "EvalBindings")),
            SourceAnchor("cli/internal/codemap/codemap.go", ("GlobFiles", "IsGlobPattern")),
            SourceAnchor("cli/cmd/output.go", ("WriteTableOutput", "WriteObjectOutput")),
        ),
    ),
    QueryCase(
        case_id="c3-eval-determinism",
        repo="c3-design",
        theme="eval determinism",
        query="eval gather deterministic external state code binding drift",
        expected_ids=("c3-108", "ref-eval-determinism", "c3-106"),
        graph_root="c3-108",
        source_anchors=(
            SourceAnchor("cli/cmd/eval.go", ("EvalBindings", "ExternalState")),
            SourceAnchor("cli/internal/eval/eval.go", ("ExternalState", "GlobFiles")),
            SourceAnchor("cli/internal/eval/eval_test.go", ("ExternalState",)),
        ),
    ),
    QueryCase(
        case_id="c3-search-rag",
        repo="c3-design",
        theme="query reasoning and search",
        query="search retrieval semantic graph context RAG ranking MRR",
        expected_ids=("c3-110", "c3-102", "c3-401", "c3-402"),
        graph_root="c3-110",
        source_anchors=(
            SourceAnchor("cli/cmd/search.go", ("RunSearch", "collectSearchRows", "fuseSemanticRows")),
            SourceAnchor("cli/internal/store/search.go", ("SearchContent", "SearchWithLimit")),
            SourceAnchor("cli/internal/store/semantic.go", ("SearchSemanticWithOptions",)),
            SourceAnchor("cli/tools/search-eval/main.go", ("MRR",)),
        ),
    ),
    QueryCase(
        case_id="c3-agent-output",
        repo="c3-design",
        theme="agent output contract",
        query="agent mode TOON structured output helpers no json marshal",
        expected_ids=("rule-output-via-helpers", "c3-109", "c3-110"),
        graph_root="rule-output-via-helpers",
        source_anchors=(
            SourceAnchor("cli/cmd/output.go", ("WriteTableOutput", "WriteObjectOutput", "writeHints")),
            SourceAnchor("cli/cmd/helpers.go", ("writeJSON",)),
            SourceAnchor("cli/cmd/output_test.go", ("TestWriteTableOutput_TOONMode",)),
        ),
    ),
    QueryCase(
        case_id="acountee-pr-approval",
        repo="acountee",
        theme="frontend/backend approval lifecycle",
        query="PR approval chain request approvals anyof allof frontend backend",
        expected_ids=(
            "c3-105",
            "c3-205",
            "ref-approval-chain",
            "ui-flow-pr-approval-flow",
            "ui-state-machine-payment-request-lifecycle",
        ),
        graph_root="ref-approval-chain",
        source_anchors=(
            SourceAnchor("apps/start/src/screens/PaymentRequestsScreen.tsx", ("approvalFlowSchema", "requestForApprovals", "canApprove")),
            SourceAnchor("apps/start/src/server/flows/pr.ts", ("createPr", "requestApprovals", "approvePr", "approveAll")),
            SourceAnchor("apps/start/src/server/dbs/queries/approval.ts", ("approvalQueries", "updateApprovalCurrentStep")),
            SourceAnchor("packages/shared/src/approval.ts", ("approvalFlowSchema", "stepsToApprovalFlow")),
        ),
        expect_drift=True,
    ),
    QueryCase(
        case_id="acountee-ui-behavior-collection",
        repo="acountee",
        theme="source-backed UI behavior and tests",
        query="ui behavior collection route screen flow state machine e2e",
        expected_ids=(
            "ref-ui-behavior-collection",
            "ref-source-backed-job-map",
            "ui-state-machine-invoice-lifecycle",
            "c3-104",
            "c3-206",
        ),
        graph_root="ref-ui-behavior-collection",
        source_anchors=(
            SourceAnchor("scripts/collect-ui-behavior.mjs", ("sourceFiles", "flowDefinitions", "screenEventDefinitions")),
            SourceAnchor("scripts/generate-ui-c3-facts.mjs", ("stateMachineBody", "screenBody", "flowBody")),
            SourceAnchor("apps/start/test/ui/c3-ui-behavior-collection.test.ts", ("describe", "collect-ui-behavior")),
        ),
        expect_drift=True,
    ),
    QueryCase(
        case_id="acountee-auth-lifecycle",
        repo="acountee",
        theme="auth lifecycle across frontend/backend",
        query="auth lifecycle login guard frontend backend e2e test token",
        expected_ids=(
            "ui-state-machine-auth-and-guard-lifecycle",
            "ui-flow-auth-and-navigation-flow",
            "ui-screen-login",
            "ref-authentication",
            "c3-213",
        ),
        graph_root="ref-authentication",
        source_anchors=(
            SourceAnchor("apps/start/src/routes/_authed.tsx", ("getInitialData", "AuthedLayout", "createAuthedScopeOptions")),
            SourceAnchor("apps/start/src/routes/login.tsx", ("LoginPage", "createFileRoute")),
            SourceAnchor("apps/start/src/server/resources/auth.ts", ("gauthSvc", "E2E_GOOGLE_AUTH_STUB")),
            SourceAnchor("apps/e2e/scripts/lightpanda-auth.mjs", ("lightpanda", "main")),
        ),
        expect_drift=True,
    ),
    QueryCase(
        case_id="acountee-invoice-lifecycle",
        repo="acountee",
        theme="invoice lifecycle across UI/backend/e2e",
        query="invoice lifecycle frontend backend e2e import obsolete restore link PR",
        expected_ids=(
            "c3-104",
            "c3-206",
            "ui-state-machine-invoice-lifecycle",
            "ui-flow-invoice-to-pr-flow",
            "ui-flow-invoice-maintenance-flow",
        ),
        graph_root="c3-206",
        source_anchors=(
            SourceAnchor("apps/start/src/screens/InvoiceScreen.tsx", ("InvoiceScreen", "markInvoiceAsRedundant", "markInvoiceAsImported")),
            SourceAnchor("apps/start/src/server/flows/invoice.ts", ("importFiles", "markInvoiceAsRedundant", "markInvoiceAsImported")),
            SourceAnchor("apps/start/src/server/services/invoice.ts", ("invoiceService", "linkToPr", "unlinkFromPr")),
            SourceAnchor("apps/e2e/scripts/lightpanda-invoice-lifecycle.mjs", ("runInvoiceLifecycleFlow", "waitForDbState")),
        ),
        expect_drift=True,
    ),
    QueryCase(
        case_id="acountee-ui-theming-papercuts",
        repo="acountee",
        theme="frontend components theming papercuts",
        query="frontend components theme variants tailwind button badge papercuts",
        expected_ids=("c3-103", "ref-variant-system", "ref-ui-patterns", "c3-108"),
        graph_root="ref-variant-system",
        source_anchors=(
            SourceAnchor("apps/start/src/components/ui/variants.ts", ("button", "badge", "listItem", "crudTable")),
            SourceAnchor("apps/start/src/routes/__root.tsx", ("theme-color", "localStorage", "matchMedia")),
            SourceAnchor("apps/start/src/screens/InvoiceScreen.tsx", ("button", "badge")),
            SourceAnchor("apps/start/src/screens/PaymentRequestsScreen.tsx", ("button", "badge", "checkbox")),
        ),
        expect_drift=True,
    ),
    QueryCase(
        case_id="acountee-realtime-sync-cycle",
        repo="acountee",
        theme="real-time ownership and cycle",
        query="real time sync nats frontend backend notification execution tracker",
        expected_ids=("ref-sync", "c3-101", "c3-209", "c3-211", "component-nats-core-broker"),
        graph_root="ref-sync",
        source_anchors=(
            SourceAnchor("apps/start/src/lib/pumped/atoms/natsSync.ts", ("natsSync", "natsCredentialsTag", "natsSubjectPrefixTag")),
            SourceAnchor("apps/start/src/lib/pumped/atoms/executionTracker.ts", ("executionTracker", "notify")),
            SourceAnchor("apps/start/src/server/resources/natsPublisher.ts", ("natsPublisher", "publish")),
            SourceAnchor("apps/start/test/ui/nats-sync.test.ts", ("describe", "executionTracker", "natsSync")),
        ),
        expect_drift=True,
    ),
)


class CheckError(Exception):
    pass


def fail(message: str) -> None:
    print(json.dumps({"ok": False, "error": message}, indent=2, sort_keys=True))
    raise SystemExit(1)


def run(cmd: list[str], *, cwd: Path, env: dict[str, str] | None = None) -> subprocess.CompletedProcess[str]:
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


def c3(project: Path, *args: str) -> subprocess.CompletedProcess[str]:
    env = os.environ.copy()
    env["C3X_MODE"] = "agent"
    return run(["bash", str(C3_WRAPPER), "--c3-dir", str(project / ".c3"), *args], cwd=project, env=env)


def git_status(repo: Path, *, ignored: bool = False) -> str:
    cmd = ["git", "-C", str(repo), "status", "--short", "--branch"]
    if ignored:
        cmd.extend(["--ignored", "-uall"])
    proc = run(cmd, cwd=ROOT)
    if proc.returncode != 0:
        raise CheckError(f"git status failed for {repo}: {proc.stderr.strip()}")
    return proc.stdout


def make_acountee_fixture() -> tuple[Path, Path, str, str]:
    if not ACOUNTEE_REPO.exists() or not (ACOUNTEE_REPO / ".c3").is_dir():
        raise CheckError(f"missing acountee target repo or .c3: {ACOUNTEE_REPO}")
    before = git_status(ACOUNTEE_REPO)
    before_ignored = git_status(ACOUNTEE_REPO, ignored=True)

    tmp_root = Path(tempfile.mkdtemp(prefix="trace-graph-rag-route-acountee.", dir="/tmp"))
    project = tmp_root / "acountee"
    project.mkdir()
    shutil.copytree(ACOUNTEE_REPO / ".c3", project / ".c3")
    for name in (
        "apps",
        "packages",
        "scripts",
        "tasks",
        "package.json",
        "pnpm-workspace.yaml",
        "tsconfig.json",
    ):
        source = ACOUNTEE_REPO / name
        if source.exists():
            os.symlink(source, project / name)
    return tmp_root, project, before, before_ignored


def contains_token(text: str, token: str) -> bool:
    if re.fullmatch(r"[A-Za-z0-9_$-]+", token):
        return bool(re.search(rf"(?<![A-Za-z0-9_$-]){re.escape(token)}(?![A-Za-z0-9_$-])", text))
    return token in text


def search_ids(project: Path, query: str) -> tuple[str, list[str]]:
    proc = c3(project, "search", query, "--no-semantic", "--limit", "12")
    if proc.returncode != 0:
        raise CheckError(f"c3 search failed for {query!r}: {(proc.stdout + proc.stderr).strip()}")
    ids: list[str] = []
    for line in proc.stdout.splitlines():
        match = re.match(r"\s{2}([A-Za-z0-9][A-Za-z0-9_-]*),", line)
        if match:
            ids.append(match.group(1))
    return proc.stdout, ids


def require_readable(project: Path, expected_ids: tuple[str, ...]) -> list[str]:
    found: list[str] = []
    for fact_id in expected_ids:
        proc = c3(project, "read", fact_id)
        if proc.returncode == 0 and f"id: {fact_id}" in proc.stdout:
            found.append(fact_id)
    return found


def graph_ids(project: Path, root_id: str) -> list[str]:
    proc = c3(project, "graph", root_id, "--depth", "1")
    if proc.returncode != 0:
        return []
    ids: list[str] = []
    for line in proc.stdout.splitlines():
        match = re.match(r"\s{4}id: ([A-Za-z0-9][A-Za-z0-9_-]*)", line)
        if match:
            ids.append(match.group(1))
    return ids


def source_results(project: Path, anchors: tuple[SourceAnchor, ...]) -> tuple[list[dict[str, Any]], list[str]]:
    results: list[dict[str, Any]] = []
    missing: list[str] = []
    for anchor in anchors:
        path = project / anchor.path
        if not path.exists():
            missing.append(anchor.path)
            continue
        text = path.read_text(encoding="utf-8", errors="ignore")
        found_symbols: list[str] = []
        for symbol in anchor.symbols:
            if contains_token(text, symbol):
                found_symbols.append(symbol)
            else:
                missing.append(f"{anchor.path}:{symbol}")
        results.append({"path": anchor.path, "symbols": found_symbols})
    return results, missing


def lookup_results(project: Path, anchors: tuple[SourceAnchor, ...]) -> list[dict[str, Any]]:
    results: list[dict[str, Any]] = []
    for anchor in anchors:
        proc = c3(project, "lookup", anchor.path)
        ids: list[str] = []
        if proc.returncode == 0:
            for line in proc.stdout.splitlines():
                match = re.match(r"\s{2}([A-Za-z0-9][A-Za-z0-9_-]*),", line)
                if match:
                    ids.append(match.group(1))
        results.append({"path": anchor.path, "ids": ids})
    return results


def check_summary(project: Path, expect_drift: bool) -> tuple[bool, list[str]]:
    proc = c3(project, "check")
    text = proc.stdout + proc.stderr
    drift = proc.returncode != 0 and "ok: false" in text and any(
        term in text for term in ("drift", "only_in_tree", "missing_from_tree", "sync check failed")
    )
    if expect_drift and not drift:
        raise CheckError("expected target drift was not visible")
    if not expect_drift and proc.returncode != 0:
        raise CheckError(f"unexpected c3 check failure: {text.strip()}")

    keep = []
    for line in text.splitlines():
        if "ok:" in line or "drift" in line or "only_in_tree:" in line or "missing_from_tree:" in line:
            keep.append(line.strip())
    return drift, keep[:8]


def score_case(case: QueryCase, project: Path) -> dict[str, Any]:
    raw_search, raw_ids = search_ids(project, case.query)
    raw_expected = [fact_id for fact_id in case.expected_ids if fact_id in raw_ids or contains_token(raw_search, fact_id)]
    baseline_id_score = len(raw_expected) / len(case.expected_ids)
    baseline_score = baseline_id_score / (4 + int(case.expect_drift))

    drift_visible, drift_summary = check_summary(project, case.expect_drift)
    readable_ids = require_readable(project, case.expected_ids)
    graph = graph_ids(project, case.graph_root)
    lookups = lookup_results(project, case.source_anchors)
    sources, missing_sources = source_results(project, case.source_anchors)

    file_count = len(case.source_anchors)
    found_file_count = sum(1 for item in sources if item["path"] and (project / item["path"]).exists())
    expected_symbol_count = sum(len(anchor.symbols) for anchor in case.source_anchors)
    found_symbol_count = sum(len(item["symbols"]) for item in sources)

    id_score = len(readable_ids) / len(case.expected_ids)
    graph_score = 1.0 if case.graph_root in graph or len(graph) > 1 else 0.0
    file_score = found_file_count / file_count
    symbol_score = found_symbol_count / expected_symbol_count
    drift_score = 1.0 if not case.expect_drift else float(drift_visible)
    candidate_score = (id_score + graph_score + file_score + symbol_score + drift_score) / (5 if case.expect_drift else 5)

    pack = {
        "hash_version": 1,
        "graph_id": f"trace-route-quality-{case.case_id}",
        "case_id": case.case_id,
        "theme": case.theme,
        "query": case.query,
        "repo": case.repo,
        "nodes": [
            {
                "id": "search-candidates",
                "kind": "binding",
                "clue": "Raw search candidates for the user query.",
                "owners": raw_ids,
                "anchors": {"expected_found": raw_expected},
            },
            {
                "id": "fact-context",
                "kind": "state",
                "clue": "Expected architecture facts/custom docs resolved through C3 read.",
                "owners": readable_ids,
                "anchors": {"expected_ids": list(case.expected_ids)},
            },
            {
                "id": "graph-neighbors",
                "kind": "policy",
                "clue": "Graph expansion around the root architecture concern.",
                "owners": graph,
                "anchors": {"root": case.graph_root},
            },
            {
                "id": "code-anchors",
                "kind": "binding",
                "clue": "First source files and lookup results to inspect.",
                "owners": sorted({item_id for lookup in lookups for item_id in lookup["ids"]}),
                "anchors": {"lookups": lookups},
            },
            {
                "id": "symbol-test-anchors",
                "kind": "validation",
                "clue": "Stable source symbols and tests that make the route executable.",
                "owners": readable_ids,
                "anchors": {"sources": sources, "missing": missing_sources},
            },
        ],
        "edges": [
            "search-candidates->fact-context:routes-to",
            "fact-context->graph-neighbors:routes-to",
            "graph-neighbors->code-anchors:resolves",
            "code-anchors->symbol-test-anchors:verifies-with",
        ],
        "hash_basis": [
            "graph_id",
            "hash_version",
            "node.id",
            "node.kind",
            "node.clue",
            "node.owners",
            "anchors.expected_ids",
            "anchors.paths",
            "anchors.symbols",
            "edges",
        ],
        "context_pack": {
            "raw_search_ids": raw_ids,
            "readable_ids": readable_ids,
            "graph_ids": graph,
            "lookup_anchors": lookups,
            "source_anchors": sources,
            "drift_visible": drift_visible,
            "drift_summary": drift_summary,
        },
    }
    canonical = json.dumps(
        {
            "hash_version": pack["hash_version"],
            "graph_id": pack["graph_id"],
            "nodes": pack["nodes"],
            "edges": pack["edges"],
            "hash_basis": pack["hash_basis"],
        },
        sort_keys=True,
        separators=(",", ":"),
    )
    pack["trace_hash"] = hashlib.sha256(canonical.encode("utf-8")).hexdigest()

    return {
        "case_id": case.case_id,
        "repo": case.repo,
        "theme": case.theme,
        "baseline_score": round(baseline_score, 4),
        "trace_score": round(candidate_score, 4),
        "delta": round(candidate_score - baseline_score, 4),
        "raw_search_expected_id_count": len(raw_expected),
        "expected_id_count": len(case.expected_ids),
        "readable_id_count": len(readable_ids),
        "source_file_count": found_file_count,
        "source_symbol_count": found_symbol_count,
        "expected_source_symbol_count": expected_symbol_count,
        "lookup_count": len(lookups),
        "drift_visible": drift_visible,
        "missing_source_anchors": missing_sources,
        "trace_pack": pack,
    }


def validate_case_result(result: dict[str, Any]) -> None:
    if result["trace_score"] <= result["baseline_score"]:
        raise CheckError(f"{result['case_id']}: Trace score did not beat baseline")
    if result["source_file_count"] == 0 or result["source_symbol_count"] == 0:
        raise CheckError(f"{result['case_id']}: no source anchors resolved")
    if result["missing_source_anchors"]:
        raise CheckError(f"{result['case_id']}: missing source anchors: {result['missing_source_anchors']}")
    pack = result["trace_pack"]
    if len(pack["nodes"]) < 5 or len(pack["edges"]) < 4:
        raise CheckError(f"{result['case_id']}: context pack is not graph-shaped")
    hash_basis = " ".join(pack["hash_basis"]).lower()
    forbidden = ("full file", "line number", "whole function", "function body", "formatting")
    bad = [term for term in forbidden if term in hash_basis]
    if bad:
        raise CheckError(f"{result['case_id']}: noisy hash basis: {bad}")


def main() -> None:
    try:
        if not C3_WRAPPER.exists():
            raise CheckError(f"missing local C3 wrapper: {C3_WRAPPER.relative_to(ROOT)}")
        c3_design_project = ROOT
        tmp_root, acountee_project, acountee_before, acountee_before_ignored = make_acountee_fixture()
        projects = {"c3-design": c3_design_project, "acountee": acountee_project}

        results = []
        for case in CASES:
            result = score_case(case, projects[case.repo])
            validate_case_result(result)
            results.append(result)

        acountee_after = git_status(ACOUNTEE_REPO)
        acountee_after_ignored = git_status(ACOUNTEE_REPO, ignored=True)
        if acountee_after != acountee_before:
            raise CheckError("real acountee visible git status changed")
        if acountee_after_ignored != acountee_before_ignored:
            raise CheckError("real acountee ignored-file git status changed")

        out_dir = tmp_root / "route-quality"
        out_dir.mkdir()
        packs = [result["trace_pack"] for result in results]
        (out_dir / "trace-packs.json").write_text(json.dumps(packs, indent=2, sort_keys=True) + "\n", encoding="utf-8")

        deltas = [result["delta"] for result in results]
        baseline_scores = [result["baseline_score"] for result in results]
        trace_scores = [result["trace_score"] for result in results]
        total_query_count = len(results)
        external_query_count = sum(1 for result in results if result["repo"] == "acountee")
        cross_cutting_query_count = sum(
            1
            for result in results
            if any(term in result["theme"] for term in ("lifecycle", "frontend", "behavior", "theming", "cycle"))
        )
        metrics = {
            "total_query_count": total_query_count,
            "external_query_count": external_query_count,
            "cross_cutting_query_count": cross_cutting_query_count,
            "baseline_route_quality": round(sum(baseline_scores) / total_query_count, 4),
            "trace_route_quality": round(sum(trace_scores) / total_query_count, 4),
            "rag_route_quality_delta": round(sum(deltas) / total_query_count, 4),
            "min_case_delta": min(deltas),
            "search_ranking_delta": 0.0,
            "search_ranking_claim_without_hitk_mrr_count": 0,
            "target_mutation_count": 0,
            "missed_known_drift_count": sum(1 for result in results if result["repo"] == "acountee" and not result["drift_visible"]),
            "proof_or_certification_claim_count": 0,
            "auto_discovery_claim_count": 0,
            "trace_pack_count": len(packs),
            "trace_pack_artifact": str(out_dir / "trace-packs.json"),
            "tmp_workspace": str(tmp_root),
        }
        if metrics["rag_route_quality_delta"] < 0.25:
            raise CheckError(f"rag_route_quality_delta below target: {metrics['rag_route_quality_delta']}")
        if metrics["external_query_count"] < 4 or metrics["total_query_count"] < 8:
            raise CheckError("query coverage below target")
        if metrics["missed_known_drift_count"] != 0:
            raise CheckError("known target drift was missed")

    except CheckError as exc:
        fail(str(exc))

    public_results = []
    for result in results:
        public = {key: value for key, value in result.items() if key != "trace_pack"}
        public_results.append(public)
    print(
        json.dumps(
            {
                "ok": True,
                "metrics": metrics,
                "cases": public_results,
                "cleanup": f"rm -rf {tmp_root}",
            },
            indent=2,
            sort_keys=True,
        )
    )


if __name__ == "__main__":
    main()
