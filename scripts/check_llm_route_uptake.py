#!/usr/bin/env python3
"""Prepare and score the LLM route-uptake eval.

This is intentionally not a route-quality replay. The older checker proves that
route context contains better clues. This checker measures whether agent answers
actually use those clues when the prompt does not tell them to use route fields.

Usage:
  python3 scripts/check_llm_route_uptake.py --generate /tmp/llm-route-uptake
  python3 scripts/check_llm_route_uptake.py --answers /tmp/llm-route-uptake/answers
  python3 scripts/check_llm_route_uptake.py --self-test
"""

from __future__ import annotations

import argparse
import csv
import copy
import hashlib
import io
import json
import re
import shutil
import tempfile
from dataclasses import dataclass
from pathlib import Path
from typing import Any

import check_trace_graph_rag_route_quality as rq


ROOT = Path(__file__).resolve().parents[1]
DEFAULT_CASE_LIMIT = 10
SCORE_FIELDS = (
    "first_file_precision",
    "first_fact_precision",
    "broad_grep_avoidance",
    "wrong_owner_claim_score",
    "stale_anchor_noticing",
    "fix_start_quality",
)

ANSWER_KEY_LEAK_PATTERNS = (
    r"\bexpected_ids\b",
    r"\bexpected_found\b",
    r"\bsource_anchors\b",
    r"\breadable_ids\b",
    r"First source files",
    r"Expected architecture facts",
    r'"sources"\s*:',
    r'"missing"\s*:',
)


class CheckError(Exception):
    pass


@dataclass(frozen=True)
class PromptBundle:
    case_id: str
    arm: str
    prompt: str
    expected: dict[str, Any]
    context: dict[str, Any]


def fail(message: str) -> None:
    print(json.dumps({"ok": False, "error": message}, indent=2, sort_keys=True))
    raise SystemExit(1)


def selected_cases(limit: int = DEFAULT_CASE_LIMIT) -> tuple[rq.QueryCase, ...]:
    return rq.CASES[:limit]


def strip_route_column(search_output: str) -> str:
    out: list[str] = []
    for line in search_output.splitlines():
        if "{id,title,why,ctx,route,s}" in line:
            out.append(line.replace("{id,title,why,ctx,route,s}", "{id,title,why,ctx,s}"))
            continue
        if not line.startswith("  "):
            out.append(line)
            continue
        row = next(csv.reader([line.strip()]))
        if len(row) == 6:
            row.pop(4)
            buf = io.StringIO()
            csv.writer(buf, lineterminator="").writerow(row)
            out.append("  " + buf.getvalue())
        else:
            out.append(line)
    return "\n".join(out) + ("\n" if search_output.endswith("\n") else "")


def strip_route_blocks(text: str) -> str:
    out: list[str] = []
    skipping = False
    route_indent = 0
    for line in text.splitlines():
        indent = len(line) - len(line.lstrip(" "))
        if re.match(r"^\s+route:\s*$", line):
            skipping = True
            route_indent = indent
            continue
        if skipping and line.strip() and indent > route_indent:
            continue
        skipping = False
        out.append(line)
    return "\n".join(out) + ("\n" if text.endswith("\n") else "")


def prompt_instruction_block(prompt: str) -> str:
    return prompt.split("## Context", 1)[0]


def route_instruction_leak_count(prompt: str) -> int:
    instructions = prompt_instruction_block(prompt).lower()
    patterns = (
        r"\buse\s+route\b",
        r"\broute\s+fields?\b",
        r"\broute\s+column\b",
        r"\bprefer\s+route\b",
        r"\bfollow\s+route\b",
        r"\broute-enriched\b",
    )
    return sum(1 for pattern in patterns if re.search(pattern, instructions))


def answer_key_leak_count(prompt: str) -> int:
    return sum(1 for pattern in ANSWER_KEY_LEAK_PATTERNS if re.search(pattern, prompt))


def source_paths(case: rq.QueryCase) -> list[str]:
    return [anchor.path for anchor in case.source_anchors]


def source_symbols(case: rq.QueryCase) -> list[str]:
    return [symbol for anchor in case.source_anchors for symbol in anchor.symbols]


def expected_for_case(case: rq.QueryCase, trace_pack: dict[str, Any]) -> dict[str, Any]:
    graph_ids = trace_pack["context_pack"]["graph_ids"]
    allowed_owner_ids = sorted(set(case.expected_ids) | set(graph_ids))
    return {
        "case_id": case.case_id,
        "repo": case.repo,
        "query": case.query,
        "expected_facts": list(case.expected_ids),
        "allowed_owner_ids": allowed_owner_ids,
        "expected_files": source_paths(case),
        "expected_symbols": source_symbols(case),
        "expect_stale_anchor": bool(case.expect_drift),
    }


def visible_route_context(trace_pack: dict[str, Any]) -> dict[str, Any]:
    pack = copy.deepcopy(trace_pack)
    if isinstance(pack.get("hash_basis"), list):
        pack["hash_basis"] = [item for item in pack["hash_basis"] if "expected" not in str(item)]
    context_pack = pack.get("context_pack", {})
    context_pack.pop("source_anchors", None)
    context_pack.pop("readable_ids", None)
    for node in pack.get("nodes", []):
        node_id = node.get("id")
        if node_id == "fact-context":
            node["clue"] = "Architecture facts/custom docs resolved through C3 read."
            node["owners"] = []
        elif node_id == "code-anchors":
            node["clue"] = "Source files and lookup results to inspect."
        elif node_id == "symbol-test-anchors":
            node["owners"] = []
        anchors = node.get("anchors")
        if not isinstance(anchors, dict):
            continue
        anchors.pop("expected_found", None)
        anchors.pop("expected_ids", None)
        anchors.pop("sources", None)
        anchors.pop("missing", None)
    return pack


def digest_text(text: str) -> str:
    return hashlib.sha256(text.encode("utf-8", errors="replace")).hexdigest()


def make_prompt(case: rq.QueryCase, arm: str, context: dict[str, Any], expected: dict[str, Any]) -> str:
    schema = {
        "first_files": ["path/to/file.ts"],
        "first_facts": ["c3-or-ref-id"],
        "broad_grep_needed": False,
        "owner_claims": ["c3-or-ref-id"],
        "stale_anchor_noticed": False,
        "fix_start": "one concise sentence naming the first inspection step",
    }
    return (
        "You are answering an architecture-to-code inspection question.\n"
        "Use only the provided context. Do not run commands. Do not edit files.\n"
        "Return only JSON in this schema:\n"
        f"{json.dumps(schema, indent=2)}\n\n"
        f"Question: {case.query}\n"
        f"Repository: {case.repo}\n\n"
        "## Context\n"
        f"{json.dumps(context, indent=2, sort_keys=True)}\n"
    )


def build_bundle(case: rq.QueryCase, project: Path, arm: str) -> PromptBundle:
    search_output, search_ids = rq.search_ids(project, case.query)
    graph_proc = rq.c3(project, "graph", case.graph_root, "--depth", "1", "--json")
    if graph_proc.returncode != 0:
        raise CheckError(f"graph failed for {case.case_id}: {(graph_proc.stdout + graph_proc.stderr).strip()}")
    result = rq.score_case(case, project)
    rq.validate_case_result(result)
    trace_pack = result["trace_pack"]
    expected = expected_for_case(case, trace_pack)

    if arm == "no_route":
        context = {
            "case_id": case.case_id,
            "theme": case.theme,
            "search": strip_route_column(search_output),
            "graph": strip_route_blocks(graph_proc.stdout),
            "known_search_ids": search_ids,
        }
    elif arm == "route":
        context = {
            "case_id": case.case_id,
            "theme": case.theme,
            "search": search_output,
            "context_pack": visible_route_context(trace_pack),
        }
    else:
        raise CheckError(f"unknown arm: {arm}")

    prompt = make_prompt(case, arm, context, expected)
    if route_instruction_leak_count(prompt) != 0:
        raise CheckError(f"{case.case_id}/{arm}: prompt instruction leaks route guidance")
    if answer_key_leak_count(prompt) != 0:
        raise CheckError(f"{case.case_id}/{arm}: prompt leaks eval answer-key labels")
    return PromptBundle(case.case_id, arm, prompt, expected, context)


def prepare_projects() -> tuple[dict[str, Path], Path, str, str]:
    tmp_root, acountee_project, before, before_ignored = rq.make_acountee_fixture()
    return {"c3-design": ROOT, "acountee": acountee_project}, tmp_root, before, before_ignored


def write_prompt_bundles(out_dir: Path, case_limit: int) -> dict[str, Any]:
    if out_dir.exists():
        shutil.rmtree(out_dir)
    out_dir.mkdir(parents=True)
    projects, tmp_root, before, before_ignored = prepare_projects()
    bundles: list[dict[str, Any]] = []
    try:
        for case in selected_cases(case_limit):
            for arm in ("no_route", "route"):
                bundle = build_bundle(case, projects[case.repo], arm)
                case_dir = out_dir / "prompts" / case.case_id / arm
                case_dir.mkdir(parents=True, exist_ok=True)
                (case_dir / "prompt.md").write_text(bundle.prompt, encoding="utf-8")
                (case_dir / "expected.json").write_text(json.dumps(bundle.expected, indent=2, sort_keys=True) + "\n", encoding="utf-8")
                (case_dir / "context.json").write_text(json.dumps(bundle.context, indent=2, sort_keys=True) + "\n", encoding="utf-8")
                bundles.append({"case_id": case.case_id, "repo": case.repo, "arm": arm, "prompt": str(case_dir / "prompt.md")})
        answers_dir = out_dir / "answers"
        answers_dir.mkdir()
        after = rq.git_status(rq.ACOUNTEE_REPO)
        after_ignored = rq.git_status(rq.ACOUNTEE_REPO, ignored=True)
        target_mutation_count = 0 if after == before and after_ignored == before_ignored else 1
        if target_mutation_count != 0:
            raise CheckError("real acountee git status changed while generating prompts")
        manifest = {
            "prompt_count": len(bundles),
            "case_count": case_limit,
            "arms": ["no_route", "route"],
            "answers_dir": str(answers_dir),
            "prompts": bundles,
            "anti_goals": {
                "route_instruction_leak_count": 0,
                "answer_key_leak_count": 0,
                "search_ranking_claim_count": 0,
                "single_llm_truth_acceptance_count": "enforced at scoring by min_agents_per_arm",
                "target_mutation_count": target_mutation_count,
                "auto_discovery_claim_count": 0,
            },
            "target_status_guard": {
                "repo": str(rq.ACOUNTEE_REPO),
                "visible_before_sha256": digest_text(before),
                "visible_after_sha256": digest_text(after),
                "ignored_before_sha256": digest_text(before_ignored),
                "ignored_after_sha256": digest_text(after_ignored),
                "mutation_count": target_mutation_count,
            },
            "tmp_workspace": str(tmp_root),
            "cleanup": f"rm -rf {tmp_root}",
        }
        (out_dir / "manifest.json").write_text(json.dumps(manifest, indent=2, sort_keys=True) + "\n", encoding="utf-8")
        return manifest
    except Exception:
        raise


def normalize_answer(data: Any) -> dict[str, Any]:
    if not isinstance(data, dict):
        raise CheckError("answer must be a JSON object")
    return data


def answer_payload(data: dict[str, Any]) -> dict[str, Any]:
    payload = data.get("answer", data)
    if not isinstance(payload, dict):
        raise CheckError("answer field must be a JSON object")
    return payload


def answer_meta(data: dict[str, Any]) -> dict[str, Any]:
    meta = data.get("meta")
    return meta if isinstance(meta, dict) else {}


def as_list(value: Any) -> list[str]:
    if value is None:
        return []
    if isinstance(value, str):
        return [value]
    if isinstance(value, list):
        return [str(item) for item in value if isinstance(item, (str, int, float))]
    return []


def boolish(value: Any) -> bool:
    if isinstance(value, bool):
        return value
    if isinstance(value, str):
        return value.strip().lower() in {"true", "yes", "y", "1"}
    return bool(value)


def contains_any(text: str, needles: list[str]) -> bool:
    lower = text.lower()
    return any(needle.lower() in lower for needle in needles)


def precision(got: list[str], expected: list[str]) -> float:
    if not got:
        return 0.0
    expected_set = set(expected)
    return sum(1 for item in got if item in expected_set) / len(got)


def score_answer(answer: dict[str, Any], expected: dict[str, Any]) -> dict[str, Any]:
    answer = answer_payload(answer)
    first_files = as_list(answer.get("first_files"))
    first_facts = as_list(answer.get("first_facts"))
    owner_claims = as_list(answer.get("owner_claims"))
    fix_start = str(answer.get("fix_start") or "")
    all_text = json.dumps(answer, sort_keys=True)

    wrong_owner_claims = [claim for claim in owner_claims if claim not in set(expected["allowed_owner_ids"])]
    if not owner_claims:
        wrong_owner_claims = [
            token
            for token in sorted(set(re.findall(r"\b(?:c3|ref|rule|ui|component)-[A-Za-z0-9_-]+\b", all_text)))
            if token not in set(expected["allowed_owner_ids"])
        ]

    broad_grep = boolish(answer.get("broad_grep_needed")) or bool(re.search(r"\b(grep|ripgrep|rg\s+-n|search the repo)\b", all_text, re.I))
    stale_seen = boolish(answer.get("stale_anchor_noticed")) or bool(re.search(r"\b(stale|missing[- ]anchor|drift|unmatched)\b", all_text, re.I))
    if not expected["expect_stale_anchor"]:
        stale_score = 1.0 if not stale_seen else 0.0
    else:
        stale_score = 1.0 if stale_seen else 0.0

    fix_has_file = contains_any(fix_start, expected["expected_files"])
    fix_has_symbol = contains_any(fix_start, expected["expected_symbols"])
    fix_has_fact = contains_any(fix_start, expected["expected_facts"])
    fix_quality = 1.0 if (fix_has_file or fix_has_symbol) and fix_has_fact else 0.5 if (fix_has_file or fix_has_symbol or fix_has_fact) else 0.0

    scores = {
        "first_file_precision": round(precision(first_files, expected["expected_files"]), 4),
        "first_fact_precision": round(precision(first_facts, expected["expected_facts"]), 4),
        "broad_grep_avoidance": 0.0 if broad_grep else 1.0,
        "wrong_owner_claim_score": 0.0 if wrong_owner_claims else 1.0,
        "stale_anchor_noticing": stale_score,
        "fix_start_quality": fix_quality,
    }
    scores["llm_route_uptake_score"] = round(sum(scores[field] for field in SCORE_FIELDS) / len(SCORE_FIELDS), 4)
    scores["wrong_owner_claims"] = wrong_owner_claims
    return scores


def load_expected_for_answer(answer_path: Path, prompts_dir: Path) -> tuple[str, str, dict[str, Any]]:
    rel = answer_path.relative_to(answer_path.parents[2])
    if len(rel.parts) < 3:
        raise CheckError(f"answer path must be <case>/<arm>/<agent>.json: {answer_path}")
    case_id, arm = rel.parts[0], rel.parts[1]
    expected_path = prompts_dir / case_id / arm / "expected.json"
    if not expected_path.exists():
        raise CheckError(f"missing expected file for answer: {expected_path}")
    return case_id, arm, json.loads(expected_path.read_text(encoding="utf-8"))


def score_answers(answer_dir: Path, min_agents_per_arm: int) -> dict[str, Any]:
    prompts_dir = answer_dir.parent / "prompts"
    if not prompts_dir.is_dir():
        raise CheckError(f"missing prompts dir beside answers: {prompts_dir}")
    manifest_path = answer_dir.parent / "manifest.json"
    if not manifest_path.exists():
        raise CheckError(f"missing manifest beside answers: {manifest_path}")
    manifest = json.loads(manifest_path.read_text(encoding="utf-8"))
    target_guard = manifest.get("target_status_guard")
    if not isinstance(target_guard, dict) or target_guard.get("mutation_count") != 0:
        raise CheckError("target mutation guard failed or missing from manifest")
    records: list[dict[str, Any]] = []
    for answer_path in sorted(answer_dir.glob("*/*/*.json")):
        agent_id = answer_path.stem
        case_id, arm, expected = load_expected_for_answer(answer_path, prompts_dir)
        answer = normalize_answer(json.loads(answer_path.read_text(encoding="utf-8")))
        meta = answer_meta(answer)
        if meta.get("agent_id") != agent_id:
            raise CheckError(f"{answer_path}: meta.agent_id must match filename stem")
        if not isinstance(meta.get("source_prompt"), str) or not meta["source_prompt"].endswith(f"/{case_id}/{arm}/prompt.md"):
            raise CheckError(f"{answer_path}: meta.source_prompt must name the scored prompt")
        if not isinstance(meta.get("runner_id"), str) or not meta["runner_id"].strip():
            raise CheckError(f"{answer_path}: meta.runner_id is required for independent-runner provenance")
        scores = score_answer(answer, expected)
        records.append({
            "case_id": case_id,
            "arm": arm,
            "agent": agent_id,
            "runner_id": meta["runner_id"],
            "source_prompt": meta["source_prompt"],
            **scores,
        })

    if not records:
        raise CheckError("no agent answer artifacts found")

    grouped: dict[tuple[str, str], set[str]] = {}
    for record in records:
        grouped.setdefault((record["case_id"], record["arm"]), set()).add(record["runner_id"])
    underfilled = [f"{case}/{arm}" for (case, arm), runners in grouped.items() if len(runners) < min_agents_per_arm]
    if underfilled:
        raise CheckError(f"single-LLM-truth guard failed; need {min_agents_per_arm} agents per arm: {underfilled}")

    route_records = [record for record in records if record["arm"] == "route"]
    baseline_records = [record for record in records if record["arm"] == "no_route"]
    if not route_records or not baseline_records:
        raise CheckError("both no_route and route arms are required")

    baseline_by_case_agent = {(record["case_id"], record["agent"]): record for record in baseline_records}
    paired_deltas: list[float] = []
    paired_delta_source = {"same_agent": 0, "case_mean": 0}
    for case_id in sorted({record["case_id"] for record in records}):
        route_case = [record for record in route_records if record["case_id"] == case_id]
        baseline_case = [record for record in baseline_records if record["case_id"] == case_id]
        case_pairs: list[float] = []
        for record in route_case:
            baseline = baseline_by_case_agent.get((record["case_id"], record["agent"]))
            if baseline:
                case_pairs.append(record["llm_route_uptake_score"] - baseline["llm_route_uptake_score"])
        if case_pairs:
            paired_deltas.append(sum(case_pairs) / len(case_pairs))
            paired_delta_source["same_agent"] += 1
        elif route_case and baseline_case:
            route_mean = sum(record["llm_route_uptake_score"] for record in route_case) / len(route_case)
            baseline_mean = sum(record["llm_route_uptake_score"] for record in baseline_case) / len(baseline_case)
            paired_deltas.append(route_mean - baseline_mean)
            paired_delta_source["case_mean"] += 1

    metrics = {
        "answer_artifact_count": len(records),
        "case_count": len({record["case_id"] for record in records}),
        "agent_count": len({record["agent"] for record in records}),
        "runner_count": len({record["runner_id"] for record in records}),
        "route_answer_count": len(route_records),
        "baseline_answer_count": len(baseline_records),
        "llm_route_uptake_score": round(sum(record["llm_route_uptake_score"] for record in route_records) / len(route_records), 4),
        "baseline_uptake_score": round(sum(record["llm_route_uptake_score"] for record in baseline_records) / len(baseline_records), 4),
        "paired_uptake_delta": round(sum(paired_deltas) / len(paired_deltas), 4) if paired_deltas else 0.0,
        "paired_delta_case_count": len(paired_deltas),
        "paired_delta_source": paired_delta_source,
        "route_instruction_leak_count": sum(
            route_instruction_leak_count(path.read_text(encoding="utf-8"))
            for path in answer_dir.parent.glob("prompts/*/*/prompt.md")
        ),
        "answer_key_leak_count": sum(
            answer_key_leak_count(path.read_text(encoding="utf-8"))
            for path in answer_dir.parent.glob("prompts/*/*/prompt.md")
        ),
        "search_ranking_claim_count": 0,
        "auto_discovery_claim_count": 0,
        "target_mutation_count": target_guard["mutation_count"],
        "single_llm_truth_acceptance_count": 0,
    }
    if metrics["route_instruction_leak_count"] != 0:
        raise CheckError("prompt instruction leak guard failed")
    if metrics["answer_key_leak_count"] != 0:
        raise CheckError("prompt answer-key leak guard failed")
    if metrics["llm_route_uptake_score"] < 0.7:
        raise CheckError(f"llm_route_uptake_score below target: {metrics['llm_route_uptake_score']}")
    if metrics["paired_uptake_delta"] <= 0:
        raise CheckError(f"route arm did not beat no-route arm: {metrics['paired_uptake_delta']}")
    return {"ok": True, "metrics": metrics, "records": records}


def write_synthetic_answers(work_dir: Path) -> Path:
    manifest = write_prompt_bundles(work_dir, case_limit=2)
    answers = Path(manifest["answers_dir"])
    for prompt_path in (work_dir / "prompts").glob("*/*/expected.json"):
        expected = json.loads(prompt_path.read_text(encoding="utf-8"))
        case_id = prompt_path.parents[1].name
        arm = prompt_path.parent.name
        for agent in ("agent-a", "agent-b"):
            answer_dir = answers / case_id / arm
            answer_dir.mkdir(parents=True, exist_ok=True)
            if arm == "route":
                payload = {
                    "first_files": expected["expected_files"][:2],
                    "first_facts": expected["expected_facts"][:2],
                    "broad_grep_needed": False,
                    "owner_claims": expected["expected_facts"][:2],
                    "stale_anchor_noticed": expected["expect_stale_anchor"],
                    "fix_start": f"Start at {expected['expected_files'][0]} under {expected['expected_facts'][0]}.",
                }
            else:
                payload = {
                    "first_files": [],
                    "first_facts": expected["expected_facts"][:1],
                    "broad_grep_needed": True,
                    "owner_claims": expected["expected_facts"][:1],
                    "stale_anchor_noticed": False,
                    "fix_start": f"Read {expected['expected_facts'][0]} then search the repo.",
                }
            answer = {
                "meta": {
                    "agent_id": agent,
                    "runner_id": f"synthetic-{agent}",
                    "source_prompt": str(work_dir / "prompts" / case_id / arm / "prompt.md"),
                },
                "answer": payload,
            }
            (answer_dir / f"{agent}.json").write_text(json.dumps(answer, indent=2, sort_keys=True) + "\n", encoding="utf-8")
    return answers


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--generate", type=Path, help="write prompt bundles to this directory")
    parser.add_argument("--answers", type=Path, help="score answer artifacts in this directory")
    parser.add_argument("--case-limit", type=int, default=DEFAULT_CASE_LIMIT)
    parser.add_argument("--min-agents-per-arm", type=int, default=2)
    parser.add_argument("--self-test", action="store_true", help="run scorer against synthetic answer artifacts")
    return parser.parse_args()


def main() -> None:
    args = parse_args()
    try:
        if args.self_test:
            tmp = Path(tempfile.mkdtemp(prefix="llm-route-uptake-selftest.", dir="/tmp"))
            answers = write_synthetic_answers(tmp)
            result = score_answers(answers, args.min_agents_per_arm)
            result["self_test_dir"] = str(tmp)
            result["cleanup"] = f"rm -rf {tmp}"
            print(json.dumps(result, indent=2, sort_keys=True))
            return
        if args.generate:
            manifest = write_prompt_bundles(args.generate, args.case_limit)
            print(json.dumps({"ok": True, "mode": "generate", "manifest": manifest}, indent=2, sort_keys=True))
            return
        if args.answers:
            print(json.dumps(score_answers(args.answers, args.min_agents_per_arm), indent=2, sort_keys=True))
            return
        raise CheckError("choose --generate, --answers, or --self-test")
    except CheckError as exc:
        fail(str(exc))


if __name__ == "__main__":
    main()
