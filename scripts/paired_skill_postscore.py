#!/usr/bin/env python3
"""Score a private paired-answer manifest without retaining project content.

The score command is an injected private Sol-high reviewer adapter. It must
return the score object accepted by paired_skill_eval.validate_score: one
condition-blind review and one deterministic evidence receipt are mandatory.
"""

from __future__ import annotations

import argparse
import hashlib
import json
import shlex
import subprocess
import sys
from pathlib import Path

# Support both `python -m scripts...` and direct execution from any cwd.
sys.path.insert(0, str(Path(__file__).resolve().parents[1]))
from scripts import paired_skill_eval as paired


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--manifest", required=True, type=Path)
    parser.add_argument("--cases", required=True, type=Path)
    parser.add_argument("--score-command", required=True)
    parser.add_argument("--output", required=True, type=Path)
    parser.add_argument("--study-id", required=True)
    return parser.parse_args()


def score_command(template: str, *, answer: Path, cases: Path, case_id: str) -> dict:
    command = [
        part.replace("{answer}", str(answer))
        .replace("{cases}", str(cases))
        .replace("{case_id}", case_id)
        for part in shlex.split(template)
    ]
    result = subprocess.run(command, text=True, capture_output=True, check=False)
    if result.returncode != 0:
        raise RuntimeError(f"private scorer failed for {case_id}")
    try:
        score = json.loads(result.stdout)
    except json.JSONDecodeError as exc:
        raise RuntimeError(f"private scorer returned invalid JSON for {case_id}") from exc
    if not isinstance(score, dict):
        raise RuntimeError(f"private scorer returned a non-object for {case_id}")
    paired.validate_score(score)
    return score


def generic_row(manifest: dict, score: dict, study_id: str) -> dict:
    base = manifest.get("base_result")
    if not isinstance(base, dict):
        raise ValueError("private manifest needs a base_result generic record")
    row = dict(base)
    if row.get("study_id") != study_id:
        raise ValueError("base_result study id does not match the requested study")
    if row.get("case_id") != manifest.get("case_id") or row.get("trial") != int(manifest["trial"]):
        raise ValueError("base_result identity does not match the private manifest")
    if row.get("condition") != manifest.get("condition"):
        raise ValueError("base_result condition does not match the private manifest")
    fields = (
        "quality_score", "correctness_score", "trace_completeness_score",
        "reasoning_depth_score", "grounding_score", "no_hallucination_score",
        "change_usefulness_score", "passed", "independent_review_count",
        "deterministic_evidence_count", "review_receipt_sha256",
        "deterministic_evidence_sha256", "scoring_cost_usd",
    )
    row.update({field: score[field] for field in fields})
    row["status"] = "scored"
    row["total_cost_usd"] = round(float(row["cost_usd"]) + float(score["scoring_cost_usd"]), 8)
    paired.validate_generic_result(row)
    return row


def main() -> int:
    args = parse_args()
    if args.output.exists() and args.output.stat().st_size:
        print("output must not already contain records", file=sys.stderr)
        return 2
    if not args.manifest.is_file() or not args.cases.is_file():
        print("manifest and cases must be readable private files", file=sys.stderr)
        return 2
    partial = args.output.with_name(args.output.name + ".partial")
    try:
        manifests = []
        for line_number, line in enumerate(args.manifest.read_text(encoding="utf-8").splitlines(), 1):
            if not line.strip():
                continue
            value = json.loads(line)
            if not isinstance(value, dict):
                raise ValueError(f"manifest line {line_number} must be an object")
            manifests.append(value)
        if not manifests:
            raise ValueError("private answer manifest is empty")
        args.output.parent.mkdir(parents=True, exist_ok=True)
        if partial.exists():
            raise ValueError("partial output already exists; remove it before retrying")
        with partial.open("x", encoding="utf-8") as handle:
            for manifest in manifests:
                answer = Path(str(manifest.get("answer_path") or ""))
                expected = str(manifest.get("answer_sha256") or "")
                if not answer.is_file():
                    raise ValueError(f"private answer is missing for {manifest.get('case_id')}")
                actual = hashlib.sha256(answer.read_bytes()).hexdigest()
                if actual != expected:
                    raise ValueError(f"answer hash mismatch for {manifest.get('case_id')}")
                score = score_command(
                    args.score_command,
                    answer=answer,
                    cases=args.cases,
                    case_id=str(manifest["case_id"]),
                )
                handle.write(json.dumps(generic_row(manifest, score, args.study_id), sort_keys=True) + "\n")
        partial.replace(args.output)
    except (OSError, KeyError, TypeError, ValueError, RuntimeError, json.JSONDecodeError) as exc:
        if partial.exists():
            partial.unlink()
        print(str(exc), file=sys.stderr)
        return 2
    print(json.dumps({"study_id": args.study_id, "scored_rows": len(manifests), "output": str(args.output)}))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
