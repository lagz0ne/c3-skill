#!/usr/bin/env python3
"""Strict LLM judge runner for C3 skill-eval answers.

Usage:
  python3 research/eval/skill-eval/judge/judge.py CASE_ID ANSWER.md
  python3 research/eval/skill-eval/judge/judge.py --prompt-only CASE_ID ANSWER.md
"""

from __future__ import annotations

import argparse
import json
import re
import statistics
import subprocess
import sys
import tempfile
from concurrent.futures import ThreadPoolExecutor
from pathlib import Path
from typing import Any

REVIEWERS = 3


DIMENSIONS = (
    "correctness",
    "trace_completeness",
    "reasoning_depth",
    "grounding",
    "no_hallucination",
    "change_usefulness",
)

WEIGHTS = {
    "correctness": 0.25,
    "trace_completeness": 0.20,
    "reasoning_depth": 0.20,
    "grounding": 0.15,
    "no_hallucination": 0.10,
    "change_usefulness": 0.10,
}

HERE = Path(__file__).resolve().parent
ROOT = HERE.parent
REPO = HERE.parents[3]
RUBRIC = HERE / "judge-rubric.md"
SCHEMA = HERE / "judge-schema.json"
CASE_FILES = (
    ROOT / "cases" / "acountee-round1.md",
    ROOT / "cases" / "acountee-crosscut.md",
    ROOT / "cases" / "acountee-properties.md",
)
TEXT_RUBRIC = ROOT / "rubric.jsonl"


def load_cases() -> dict[str, dict[str, Any]]:
    cases: dict[str, dict[str, Any]] = {}
    with TEXT_RUBRIC.open("r", encoding="utf-8") as f:
        for line in f:
            if line.strip():
                row = json.loads(line)
                cases[row["id"]] = row
    return cases


def extract_case_excerpt(case_id: str, case_files: tuple[Path, ...] = CASE_FILES) -> str:
    header_re = re.compile(rf"^##\s+{re.escape(case_id)}(?::|\b).*$", re.MULTILINE)
    for path in case_files:
        text = path.read_text(encoding="utf-8")
        match = header_re.search(text)
        if not match:
            continue
        next_match = re.search(r"^##\s+\S", text[match.end() :], re.MULTILINE)
        end = len(text) if not next_match else match.end() + next_match.start()
        return text[match.start() : end].strip()
    raise SystemExit(f"case excerpt not found: {case_id}")


def fixture_entity_ids() -> str:
    fixture = ROOT / "fixtures" / "acountee" / ".c3"
    ids: set[str] = set()
    for path in sorted(fixture.rglob("*.md")):
        if path.parent.name in ("adr", "refs", "recipes"):
            ids.add(path.stem)
        ids.update(re.findall(r"^id:\s*(\S+)", path.read_text(encoding="utf-8"), re.MULTILINE))
    return ", ".join(sorted(ids))


def build_prompt(case_id: str, answer_file: Path) -> str:
    cases = load_cases()
    if case_id not in cases:
        raise SystemExit(f"unknown case id: {case_id}")

    case = cases[case_id]
    answer = answer_file.read_text(encoding="utf-8")
    rubric = RUBRIC.read_text(encoding="utf-8")
    excerpt = extract_case_excerpt(case_id)

    return f"""You are a STRICT INDEPENDENT C3 skill-eval judge.

Score an UNKNOWN candidate answer. Do not defend the answer. Do not reward
shape, term stuffing, or grounded-looking ids unless the answer is correct,
deep, grounded, and useful.

Use only:
- the judge rubric,
- the case question,
- the machine ground truth,
- the case-file excerpt,
- the fixture entity inventory,
- the candidate answer.

Return JSON only. Follow this schema:
- case_id: string
- dimensions: object with correctness, trace_completeness, reasoning_depth,
  grounding, no_hallucination, change_usefulness
- each dimension has score integer 1-5 and justification string
- overall: weighted 1-5 score using rubric weights
- verdict: "pass" or "fail" using the rubric pass bar
- summary: short strict review summary
- quality_gaps: array of concrete gaps, especially text-match-looking but
  semantically shallow, unsupported, wrong, or unsafe issues

# Judge Rubric

{rubric}

# Case

case_id: {case_id}
category: {case.get("category", "")}
question: {case["question"]}

# Machine Ground Truth

{case["ground_truth"]}

# Fixture Entity Inventory

Every id below EXISTS in the fixture. Citing one is never an invented id,
even when it is absent from the ground truth or excerpt. A cited id that is a
prefix of exactly one inventory id refers to that entity (sloppy citation, not
invented); only ids matching zero inventory ids, or ambiguously matching
several, are invented.

{fixture_entity_ids()}

# Case File Excerpt

{excerpt}

# Candidate Answer

{answer}
"""


def extract_json_object(text: str) -> dict[str, Any]:
    stripped = text.strip()
    if stripped.startswith("```"):
        stripped = re.sub(r"^```(?:json)?\s*", "", stripped)
        stripped = re.sub(r"\s*```$", "", stripped)
    try:
        return json.loads(stripped)
    except json.JSONDecodeError:
        start = stripped.find("{")
        end = stripped.rfind("}")
        if start == -1 or end == -1 or end < start:
            raise
        return json.loads(stripped[start : end + 1])


def normalize_verdict(verdict: dict[str, Any]) -> dict[str, Any]:
    dimensions = verdict.get("dimensions", {})
    missing = [name for name in DIMENSIONS if name not in dimensions]
    if missing:
        raise ValueError(f"missing dimensions: {missing}")

    weighted = 0.0
    lows: list[str] = []
    for name in DIMENSIONS:
        item = dimensions[name]
        score = int(item["score"])
        if score < 1 or score > 5:
            raise ValueError(f"{name} score outside 1-5: {score}")
        item["score"] = score
        weighted += score * WEIGHTS[name]
        if score < 3:
            lows.append(name)

    computed = round(weighted, 2)
    verdict["overall"] = computed
    should_pass = (
        computed >= 4.0
        and dimensions["correctness"]["score"] >= 4
        and dimensions["no_hallucination"]["score"] >= 4
        and not lows
    )
    verdict["verdict"] = "pass" if should_pass else "fail"
    verdict.setdefault("quality_gaps", [])
    return verdict


def aggregate_reviews(reviews: list[dict[str, Any]]) -> dict[str, Any]:
    """Median per dimension across reviewers; verdict from the pass rule on the aggregate."""
    if len(reviews) == 1:
        return reviews[0]
    majority = "pass" if sum(r["verdict"] == "pass" for r in reviews) * 2 > len(reviews) else "fail"
    base = next((r for r in reviews if r["verdict"] == majority), reviews[0])
    agg = dict(base)
    agg["dimensions"] = {
        name: {
            "score": int(statistics.median(r["dimensions"][name]["score"] for r in reviews)),
            "justification": base["dimensions"][name]["justification"],
        }
        for name in DIMENSIONS
    }
    # Verdict = pass rule applied to the aggregated dimensions (normalize_verdict),
    # never reviewer majority — a 2-1 pass vote can disagree with median overall
    # (observed: votes [pass, fail, pass] around aggregated overall 3.85).
    agg = normalize_verdict(agg)
    agg["reviewer_verdicts"] = [r["verdict"] for r in reviews]
    agg["reviewer_overalls"] = [r["overall"] for r in reviews]
    return agg


def run_codex(prompt: str, model: str | None) -> dict[str, Any]:
    with tempfile.TemporaryDirectory(prefix="c3-judge-") as td:
        out = Path(td) / "verdict.json"
        cmd = [
            "codex",
            "exec",
            "--ephemeral",
            "--cd",
            str(REPO),
            "--sandbox",
            "read-only",
            "--output-schema",
            str(SCHEMA),
            "--output-last-message",
            str(out),
        ]
        if model:
            cmd.extend(["--model", model])
        cmd.append("-")
        result = subprocess.run(
            cmd,
            input=prompt,
            text=True,
            check=False,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
        )
        if result.returncode != 0:
            if result.stdout:
                print(result.stdout, file=sys.stderr)
            if result.stderr:
                print(result.stderr, file=sys.stderr)
            raise subprocess.CalledProcessError(result.returncode, cmd)
        return extract_json_object(out.read_text(encoding="utf-8"))


def parse_args(argv: list[str]) -> argparse.Namespace:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("case_id")
    parser.add_argument("answer_file", type=Path)
    parser.add_argument("--model", default=None, help="optional codex model override")
    parser.add_argument("--prompt-only", action="store_true", help="print judge prompt and exit")
    parser.add_argument("--output", type=Path, help="write verdict JSON to file")
    return parser.parse_args(argv)


def main(argv: list[str]) -> int:
    args = parse_args(argv)
    prompt = build_prompt(args.case_id, args.answer_file)
    if args.prompt_only:
        print(prompt)
        return 0

    with ThreadPoolExecutor(max_workers=REVIEWERS) as pool:
        reviews = [
            normalize_verdict(r)
            for r in pool.map(lambda _: run_codex(prompt, args.model), range(REVIEWERS))
        ]
    verdict = aggregate_reviews(reviews)
    verdict["case_id"] = args.case_id
    payload = json.dumps(verdict, indent=2, sort_keys=True)
    if args.output:
        args.output.parent.mkdir(parents=True, exist_ok=True)
        args.output.write_text(payload + "\n", encoding="utf-8")
    print(payload)
    return 0


if __name__ == "__main__":
    raise SystemExit(main(sys.argv[1:]))
