#!/usr/bin/env python3
"""Check the route-enrichment OKRA artifacts.

This is intentionally stricter than a markdown-shape check. It verifies:
- each expected pilot exists
- each pilot has impact-analysis and code/fix routing text
- each pilot cites expected C3 facts that resolve through the local C3 wrapper
- each pilot's stable code anchors exist in the current source tree
- the spec does not turn route enrichment into proof, correction, or an apply gate
- hash basis text excludes full-file, line-number, and whole-body inputs
- negative controls fail the checker
"""

from __future__ import annotations

import json
import re
import subprocess
from dataclasses import dataclass
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
SPEC = ROOT / "docs/specs/2026-07-09-architecture-trace-graph.md"
OKRA = ROOT / "docs/specs/2026-07-09-architecture-trace-graph-okra.md"


@dataclass(frozen=True)
class Anchor:
    symbol: str
    globs: tuple[str, ...]


@dataclass(frozen=True)
class Pilot:
    title: str
    question: str
    facts: tuple[str, ...]
    anchors: tuple[Anchor, ...]
    path_anchors: tuple[str, ...]
    impact_terms: tuple[str, ...]
    code_fix_terms: tuple[str, ...]
    query_terms: tuple[str, ...] = ()


EXPECTED_PILOTS: tuple[Pilot, ...] = (
    Pilot(
        title="Lookup Ownership",
        question="Which path handles mapping a file or glob",
        facts=("c3-110", "c3-106", "c3-102", "c3-109", "rule-output-via-helpers"),
        anchors=(
            Anchor("RunLookup", ("cli/cmd/lookup.go",)),
            Anchor("GlobFiles", ("cli/internal/codemap/*.go",)),
            Anchor("IsGlobPattern", ("cli/internal/codemap/*.go",)),
            Anchor("WriteTableOutput", ("cli/cmd/*.go",)),
            Anchor("WriteObjectOutput", ("cli/cmd/*.go",)),
            Anchor("writeJSON", ("cli/cmd/*.go",)),
        ),
        path_anchors=("cli/cmd/lookup.go", "cli/internal/store/**"),
        impact_terms=("c3-110", "c3-106", "c3-102"),
        code_fix_terms=("RunLookup", "GlobFiles", "IsGlobPattern"),
    ),
    Pilot(
        title="Eval Conformance",
        question="Before changing eval gather semantics",
        facts=("c3-108", "ref-eval-determinism", "c3-106"),
        anchors=(
            Anchor("EvalBindings", ("cli/cmd/eval.go",)),
            Anchor("gather", ("cli/internal/eval/*.go",)),
            Anchor("ExternalState", ("cli/internal/eval/*.go", "cli/cmd/eval.go")),
            Anchor("GlobFiles", ("cli/internal/codemap/*.go",)),
            Anchor("assert", ("cli/internal/eval/*.go",)),
        ),
        path_anchors=("cli/internal/eval",),
        impact_terms=("c3-108", "ref-eval-determinism", "c3-106"),
        code_fix_terms=("cli/internal/eval", "gather", "assert"),
    ),
    Pilot(
        title="Agent Output Format",
        question="What should I inspect before changing structured agent output",
        facts=("rule-output-via-helpers", "c3-109", "c3-110"),
        anchors=(
            Anchor("WriteTableOutput", ("cli/cmd/*.go",)),
            Anchor("WriteObjectOutput", ("cli/cmd/*.go",)),
            Anchor("writeJSON", ("cli/cmd/*.go",)),
            Anchor("writeHints", ("cli/cmd/*.go",)),
            Anchor("RunLookup", ("cli/cmd/lookup.go",)),
            Anchor("RunSearch", ("cli/cmd/search.go",)),
            Anchor("RunCheckV2", ("cli/cmd/check_enhanced.go",)),
        ),
        path_anchors=(),
        impact_terms=("rule-output-via-helpers", "c3-109", "c3-110"),
        code_fix_terms=("WriteTableOutput", "WriteObjectOutput", "writeJSON", "writeHints"),
    ),
    Pilot(
        title="Runtime Resolution",
        question="Where does version/platform runtime selection live",
        facts=("c3-203", "c3-301", "ref-cross-compiled-binary", "ref-fat-thin-distribution"),
        anchors=(
            Anchor("asset_name", ("skills/c3/bin/c3x.sh",)),
            Anchor("VERSION", ("skills/c3/bin/c3x.sh", "skills/c3/bin/VERSION")),
            Anchor("assetNames", ("packages/cli/src/manager.ts",)),
            Anchor("resolvePlatform", ("packages/cli/src/manager.ts",)),
            Anchor("ensureCachedAsset", ("packages/cli/src/manager.ts",)),
            Anchor("runCli", ("packages/cli/src/manager.ts",)),
        ),
        path_anchors=("skills/c3/bin/c3x.sh", "skills/c3/bin/VERSION", "packages/cli/src/manager.ts"),
        impact_terms=("c3-203", "c3-301", "skill-wrapper", "npm thin-client"),
        code_fix_terms=("skills/c3/bin/c3x.sh", "packages/cli/src/manager.ts"),
    ),
    Pilot(
        title="Canvas Validation",
        question="Before changing fact/canvas validation",
        facts=("c3-103", "c3-101", "c3-102", "c3-110"),
        anchors=(
            Anchor("ParseCanvasDocument", ("cli/internal/schema/*.go",)),
            Anchor("ResolveCanvas", ("cli/internal/schema/*.go",)),
            Anchor("DefinitionForDir", ("cli/internal/schema/*.go",)),
            Anchor("ParseFrontmatter", ("cli/internal/frontmatter/*.go",)),
            Anchor("ParseSections", ("cli/internal/markdown/*.go",)),
            Anchor("ParseTable", ("cli/internal/markdown/*.go",)),
            Anchor("WriteEntity", ("cli/internal/content/*.go",)),
            Anchor("ReadEntity", ("cli/internal/content/*.go",)),
            Anchor("RenderMarkdown", ("cli/internal/content/*.go",)),
            Anchor("RunCheckV2", ("cli/cmd/check_enhanced.go",)),
        ),
        path_anchors=(),
        impact_terms=("canvas shape", "markdown parsing", "stored", "command reporting"),
        code_fix_terms=("schema resolution", "doc-model parsing", "RunCheckV2"),
    ),
    Pilot(
        title="Query Reasoning / RAG",
        question="How should C3 retrieval guide an answer",
        facts=("c3-110", "c3-102", "c3-401", "c3-402"),
        anchors=(
            Anchor("RunSearch", ("cli/cmd/search.go",)),
            Anchor("collectSearchRows", ("cli/cmd/search.go",)),
            Anchor("expandHybridRows", ("cli/cmd/search.go",)),
            Anchor("fuseSemanticRows", ("cli/cmd/search.go",)),
            Anchor("enrichSearchRow", ("cli/cmd/search.go",)),
            Anchor("SearchContent", ("cli/internal/store/search.go",)),
            Anchor("SearchWithLimit", ("cli/internal/store/search.go",)),
            Anchor("EnsureSemanticIndexWithOptions", ("cli/internal/store/semantic.go",)),
            Anchor("SearchSemanticWithOptions", ("cli/internal/store/semantic.go",)),
            Anchor("SearchResultRow", ("cli/cmd/search.go",)),
            Anchor("SearchContext", ("cli/cmd/search.go",)),
            Anchor("MRR", ("cli/tools/search-eval/*.go",)),
        ),
        path_anchors=(
            "cli/cmd/search.go",
            "cli/internal/store/semantic.go",
            "cli/tools/search-eval/**",
            "cli/tools/semantic-assets/**",
            "docs/specs/2026-06-08-local-onnx-semantic-search.md",
        ),
        impact_terms=("c3-110", "c3-102", "c3-401", "c3-402"),
        code_fix_terms=("RunSearch", "collectSearchRows", "fuseSemanticRows", "SearchSemanticWithOptions"),
        query_terms=("candidate ids", "match_sources", "graph context", "refs/rules", "stale/missing-anchor"),
    ),
)


class CheckError(Exception):
    pass


def fail(message: str) -> None:
    print(json.dumps({"ok": False, "error": message}, indent=2))
    raise SystemExit(1)


def read(path: Path) -> str:
    if not path.exists():
        raise CheckError(f"missing required artifact: {path.relative_to(ROOT)}")
    return path.read_text(encoding="utf-8")


def run(cmd: list[str]) -> subprocess.CompletedProcess[str]:
    return subprocess.run(
        cmd,
        cwd=ROOT,
        text=True,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        check=False,
    )


def changed_paths() -> list[str]:
    proc = run(["git", "status", "--short", "--untracked-files=all"])
    if proc.returncode != 0:
        raise CheckError(f"git status failed: {proc.stderr.strip()}")

    paths: list[str] = []
    for line in proc.stdout.splitlines():
        if not line.strip():
            continue
        path = line[3:].strip()
        if " -> " in path:
            path = path.split(" -> ", 1)[1].strip()
        if path:
            paths.append(path)
    return paths


def tracked_pilot_paths() -> list[str]:
    proc = run(["git", "diff", "--name-only", "--", "cli", ".c3", "docs/specs", "scripts"])
    if proc.returncode != 0:
        raise CheckError(f"git diff failed: {proc.stderr.strip()}")
    status_paths = changed_paths()
    return sorted(set(proc.stdout.splitlines()) | set(status_paths))


def section_for(okra: str, title: str) -> str:
    pattern = re.compile(
        rf"^## Pilot Trace \d+: {re.escape(title)}\n(?P<body>.*?)(?=^## Pilot Trace \d+:|^## Five-Question Scorecard|\Z)",
        flags=re.MULTILINE | re.DOTALL,
    )
    match = pattern.search(okra)
    if not match:
        raise CheckError(f"missing pilot section: {title}")
    return match.group("body")


def files_for_glob(pattern: str) -> list[Path]:
    candidate = ROOT / pattern
    if candidate.exists() and candidate.is_dir():
        return [p for p in candidate.rglob("*") if p.is_file()]
    return [p for p in ROOT.glob(pattern) if p.is_file()]


def symbol_exists(anchor: Anchor) -> bool:
    needle = anchor.symbol
    for pattern in anchor.globs:
        for path in files_for_glob(pattern):
            try:
                if needle in path.read_text(encoding="utf-8", errors="ignore"):
                    return True
            except UnicodeDecodeError:
                continue
    return False


def c3_fact_exists(fact: str) -> bool:
    proc = run(["bash", "skills/c3/bin/c3x.sh", "read", fact])
    return proc.returncode == 0 and f"id: {fact}" in proc.stdout


def require_two_phase_text(section: str, pilot: Pilot) -> None:
    if pilot.question not in section:
        raise CheckError(f"{pilot.title}: missing representative question text")

    impact = re.search(
        r"Impact-analysis signal:\n\n```text\n(?P<body>.*?)\n```",
        section,
        flags=re.DOTALL,
    )
    code_fix = re.search(
        r"Code/fix signal:\n\n```text\n(?P<body>.*?)\n```",
        section,
        flags=re.DOTALL,
    )
    if not impact or len(impact.group("body").split()) < 8:
        raise CheckError(f"{pilot.title}: missing substantive impact-analysis signal")
    if not code_fix or len(code_fix.group("body").split()) < 8:
        raise CheckError(f"{pilot.title}: missing substantive code/fix signal")

    impact_body = impact.group("body")
    code_fix_body = code_fix.group("body")
    missing_impact_terms = [term for term in pilot.impact_terms if term not in impact_body]
    missing_code_terms = [term for term in pilot.code_fix_terms if term not in code_fix_body]
    if missing_impact_terms:
        raise CheckError(f"{pilot.title}: impact signal missing expected term(s): {missing_impact_terms}")
    if missing_code_terms:
        raise CheckError(f"{pilot.title}: code/fix signal missing expected term(s): {missing_code_terms}")

    if pilot.query_terms:
        query = re.search(
            r"Query/RAG signal:\n\n```text\n(?P<body>.*?)\n```",
            section,
            flags=re.DOTALL,
        )
        if not query or len(query.group("body").split()) < 8:
            raise CheckError(f"{pilot.title}: missing substantive Query/RAG signal")
        query_body = query.group("body")
        missing_query_terms = [term for term in pilot.query_terms if term not in query_body]
        if missing_query_terms:
            raise CheckError(f"{pilot.title}: Query/RAG signal missing expected term(s): {missing_query_terms}")


def require_pilot_authority(okra: str, section: str, pilot: Pilot) -> None:
    for fact in pilot.facts:
        if fact not in section:
            raise CheckError(f"{pilot.title}: missing expected fact {fact}")
        if fact not in okra:
            raise CheckError(f"{pilot.title}: expected fact {fact} absent from OKRA document")
        if not c3_fact_exists(fact):
            raise CheckError(f"{pilot.title}: C3 fact does not resolve through local wrapper: {fact}")

    def has_documented_anchor(symbol: str) -> bool:
        return bool(re.search(rf"(?<![A-Za-z0-9_]){re.escape(symbol)}(?![A-Za-z0-9_])", section))

    missing_doc_anchors = [anchor.symbol for anchor in pilot.anchors if not has_documented_anchor(anchor.symbol)]
    if missing_doc_anchors:
        raise CheckError(f"{pilot.title}: missing documented anchor(s): {missing_doc_anchors}")

    missing_source_anchors = [anchor.symbol for anchor in pilot.anchors if not symbol_exists(anchor)]
    if missing_source_anchors:
        raise CheckError(f"{pilot.title}: missing source anchor(s): {missing_source_anchors}")

    missing_doc_paths = [path for path in pilot.path_anchors if path not in section]
    if missing_doc_paths:
        raise CheckError(f"{pilot.title}: missing documented path anchor(s): {missing_doc_paths}")
    missing_source_paths = [path for path in pilot.path_anchors if not files_for_glob(path)]
    if missing_source_paths:
        raise CheckError(f"{pilot.title}: documented path anchor(s) resolve to no files: {missing_source_paths}")


def require_hash_basis(spec: str, okra: str) -> None:
    forbidden = ("full file", "line number", "function bod", "whole function", "formatting")

    hash_these = re.search(
        r"## Route Hash Signal\n(?P<body>.*?)(?=^## Status Signal)",
        spec,
        flags=re.MULTILINE | re.DOTALL,
    )
    if not hash_these:
        raise CheckError("missing Route Hash Signal section")

    positive_hash_body = re.search(
        r"Hash these:\n\n(?P<body>.*?)(?=Do not hash these:)",
        hash_these.group("body"),
        flags=re.DOTALL,
    )
    if not positive_hash_body:
        raise CheckError("missing positive hash basis table")
    lowered_positive = positive_hash_body.group("body").lower().replace("-", " ")
    bad_positive = [term for term in forbidden if term in lowered_positive]
    if bad_positive:
        raise CheckError(f"forbidden positive hash basis term(s): {bad_positive}")

    for line in (spec + "\n" + okra).splitlines():
        lowered = line.lower().replace("-", " ")
        if "hash_basis" not in lowered and not (line.lstrip().startswith("|") and "hash basis" in lowered):
            continue
        bad = [term for term in forbidden if term in lowered]
        if bad:
            raise CheckError(f"forbidden hash basis line: {line}")

    for line in okra.splitlines():
        lowered = line.lower().replace("-", " ")
        if not line.lstrip().startswith("|"):
            continue
        if "exclude" in lowered or "do not" in lowered:
            continue
        bad = [term for term in forbidden if term in lowered]
        if bad:
            raise CheckError(f"forbidden OKRA table hash-noise term: {line}")


def require_signal_not_gate(spec: str, okra: str) -> None:
    required_spec_phrases = [
        "Brainstorming / impact analysis",
        "Code / fix execution",
        "Query Reasoning / RAG",
        "Route enrichment is allowed to guide a code fix.",
        "It is not allowed to certify that the fix is correct.",
        "Adding primitives",
        "Do not hash full file content or line numbers.",
    ]
    missing = [phrase for phrase in required_spec_phrases if phrase not in spec]
    if missing:
        raise CheckError(f"missing spec phrase(s): {missing}")

    combined = (spec + "\n" + okra).lower().replace("-", " ")
    forbidden_patterns = [
        r"trace graph\s+(is|becomes|acts as|serves as)\s+(the|an|a)?\s*apply gate",
        r"route enrichment\s+(is|becomes|acts as|serves as)\s+(the|an|a)?\s*apply gate",
        r"trace graph\s+(certifies|proves|guarantees)\s+(the\s+)?(fix|correctness|conformance)",
        r"route enrichment\s+(certifies|proves|guarantees)\s+(the\s+)?(fix|correctness|conformance)",
        r"changed trace hash\s+(blocks|must block|shall block)",
        r"changed route hash\s+(blocks|must block|shall block)",
        r"trace graph\s+is\s+required\s+before\s+change apply\s+can\s+proceed",
        r"route enrichment\s+is\s+required\s+before\s+change apply\s+can\s+proceed",
        r"trace graph.*(required|mandatory|must run).*change apply",
        r"route enrichment.*(required|mandatory|must run).*change apply",
        r"trace graph.*must\s+be\s+completed\s+before\s+change apply\s+may\s+proceed",
        r"route enrichment.*must\s+be\s+completed\s+before\s+change apply\s+may\s+proceed",
        r"trace graph.*(must|shall|required|mandatory|complete|completed).*before.*change apply.*(proceed|run|continue|apply)",
        r"route enrichment.*(must|shall|required|mandatory|complete|completed).*before.*change apply.*(proceed|run|continue|apply)",
        r"change apply.*(requires|is blocked by|waits for).*trace graph",
        r"change apply.*(requires|is blocked by|waits for).*route enrichment",
        r"trace graph\s+(rewrites|corrects)\s+code",
        r"route enrichment\s+(rewrites|corrects)\s+code",
    ]
    hits = [pat for pat in forbidden_patterns if re.search(pat, combined)]
    if hits:
        raise CheckError(f"forbidden authority pattern(s): {hits}")


def require_no_new_primitive(paths: list[str]) -> None:
    forbidden_changes = [
        path
        for path in paths
        if path.startswith(".c3/canvases/")
        or path.startswith(".c3/c3-")
        or path.startswith(".c3/refs/")
        or path.startswith(".c3/rules/")
    ]
    if forbidden_changes:
        raise CheckError(f"route enrichment touched forbidden fact/canvas surfaces: {forbidden_changes}")
    help_go = (ROOT / "cli/cmd/help.go").read_text(encoding="utf-8")
    for name in ("trace", "route", "spine", "impact"):
        if re.search(rf'Name:\s+"{name}"', help_go):
            raise CheckError(f"new user-facing primitive command registered: {name}")


def validate(spec: str, okra: str, paths: list[str]) -> dict[str, int | float]:
    for pilot in EXPECTED_PILOTS:
        section = section_for(okra, pilot.title)
        require_two_phase_text(section, pilot)
        require_pilot_authority(okra, section, pilot)

    require_signal_not_gate(spec, okra)
    require_hash_basis(spec, okra)
    require_no_new_primitive(paths)

    pilot_count = len(EXPECTED_PILOTS)
    return {
        "pilot_trace_count": pilot_count,
        "representative_question_count": pilot_count,
        "useful_first_signal_count": pilot_count,
        "impact_analysis_signal_count": pilot_count,
        "code_fix_signal_count": pilot_count,
        "query_reasoning_signal_count": sum(1 for pilot in EXPECTED_PILOTS if pilot.query_terms),
        "trace_graph_decision_readiness": 1.0,
        "false_authority_claim_count": 0,
        "node_hash_noise_failures": 0,
        "new_user_facing_primitive_count": 0,
    }


def require_negative_controls_fail(spec: str, okra: str, paths: list[str]) -> list[str]:
    useless_impact = re.sub(
        r"Impact-analysis signal:\n\n```text\n.*?\n```",
        "Impact-analysis signal:\n\n```text\nThis block has enough words but no expected facts or useful routing terms for the phase.\n```",
        okra,
        count=1,
        flags=re.DOTALL,
    )
    useless_code_fix = re.sub(
        r"Code/fix signal:\n\n```text\n.*?\n```",
        "Code/fix signal:\n\n```text\nThis block has enough words but no concrete code anchors or useful fix routing terms.\n```",
        okra,
        count=1,
        flags=re.DOTALL,
    )
    useless_query_rag = re.sub(
        r"Query/RAG signal:\n\n```text\n.*?\n```",
        "Query/RAG signal:\n\n```text\nThis block has enough words but no retrieval provenance or context-pack routing terms.\n```",
        okra,
        count=1,
        flags=re.DOTALL,
    )

    cases: list[tuple[str, str, str]] = [
        (
            "apply-gate-positive",
            spec,
            okra + "\nRoute enrichment is the apply gate for C3 changes.\n",
        ),
        (
            "change-apply-required",
            spec,
            okra + "\nRoute enrichment is required before change apply can proceed.\n",
        ),
        (
            "change-apply-completed-before",
            spec,
            okra + "\nRoute enrichment must be completed before change apply may proceed.\n",
        ),
        (
            "certifies-fix-positive",
            spec + "\nRoute enrichment certifies the fix correctness.\n",
            okra,
        ),
        (
            "analysis-only",
            spec,
            okra.replace("Code/fix signal:", "Code/fix route removed:", 1),
        ),
        (
            "useless-impact-signal",
            spec,
            useless_impact,
        ),
        (
            "useless-code-fix-signal",
            spec,
            useless_code_fix,
        ),
        (
            "useless-query-rag-signal",
            spec,
            useless_query_rag,
        ),
        (
            "noisy-hash-basis",
            spec.replace("| entity ids |", "| full file content |", 1),
            okra,
        ),
        (
            "pilot-row-noisy-hash-basis",
            spec,
            okra.replace("kind, clue, owners, symbols, edges", "full file content, line numbers", 1),
        ),
        (
            "pilot-row-hyphenated-noisy-hash-basis",
            spec,
            okra.replace("kind, clue, owners, symbols, edges", "full-file content, line-number anchors", 1),
        ),
        (
            "missing-path-anchor",
            spec,
            okra.replace("cli/internal/store/**", "missing/nope/**"),
        ),
        (
            "missing-anchor",
            spec,
            okra.replace("RunLookup", "RunLookupMissingSentinel"),
        ),
    ]

    passed_when_should_fail: list[str] = []
    for name, case_spec, case_okra in cases:
        try:
            validate(case_spec, case_okra, paths)
        except CheckError:
            continue
        passed_when_should_fail.append(name)
    if passed_when_should_fail:
        raise CheckError(f"negative control(s) passed unexpectedly: {passed_when_should_fail}")
    return [name for name, _, _ in cases]


def main() -> None:
    try:
        spec = read(SPEC)
        okra = read(OKRA)
        paths = tracked_pilot_paths()
        metrics = validate(spec, okra, paths)
        negative_controls = require_negative_controls_fail(spec, okra, paths)
    except CheckError as exc:
        fail(str(exc))

    result = {
        "ok": True,
        "metrics": metrics,
        "negative_controls": negative_controls,
        "changed_paths_checked": paths,
    }
    print(json.dumps(result, indent=2, sort_keys=True))


if __name__ == "__main__":
    main()
