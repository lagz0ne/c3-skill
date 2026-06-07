import json
import sys
import tempfile
import time
import unittest
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))

import eval_gate as gate


def rec(case, *, exit_code=0, accuracy=1.0, canvas_passed=True,
        adr_quality=1.0, tokens=100_000, effective=80_000, threshold="ok",
        agent="claude", trial=1, dry_run=False, unavailable=False):
    """Build one eval record shaped like agent_efficiency_eval emits.

    `threshold` mirrors the harness's threshold_status (ok/soft/upper/no_go),
    computed on effective tokens — that, not tokens_total, drives the gate's
    secondary token guardrail.
    """
    return {
        "agent": agent,
        "case": case,
        "trial": trial,
        "exit_code": exit_code,
        "accuracy_score": accuracy,
        "canvas_quality_passed": canvas_passed,
        "adr_quality_score": adr_quality,
        "tokens_total": tokens,
        "effective_tokens_total": effective,
        "threshold_status": threshold,
        "dry_run": dry_run,
        "agent_unavailable": unavailable,
    }


class QualityRuleTests(unittest.TestCase):
    def test_plain_case_passes_on_accuracy(self):
        self.assertTrue(gate.record_passes_quality(rec("task_session")))

    def test_plain_case_fails_on_low_accuracy(self):
        self.assertFalse(gate.record_passes_quality(rec("task_session", accuracy=0.5)))

    def test_nonzero_exit_fails(self):
        self.assertFalse(gate.record_passes_quality(rec("task_session", exit_code=1)))

    def test_canvas_case_requires_canvas_floor(self):
        self.assertTrue(gate.record_passes_quality(rec("canvas_prd")))
        self.assertFalse(gate.record_passes_quality(rec("canvas_prd", canvas_passed=False)))

    def test_adr_case_requires_adr_floor(self):
        self.assertTrue(gate.record_passes_quality(rec("adr_create", adr_quality=0.8)))
        self.assertFalse(gate.record_passes_quality(rec("adr_create", adr_quality=0.79)))

    def test_canvas_adr_classified_as_canvas_not_adr(self):
        # canvas_c3_adr contains "adr" but is a canvas case — canvas floor wins.
        r = rec("canvas_c3_adr", canvas_passed=True, adr_quality=0.0)
        self.assertTrue(gate.record_passes_quality(r))


class ComputeQualityTests(unittest.TestCase):
    def test_excludes_dry_run_and_unavailable(self):
        q = gate.compute_quality([
            rec("task_session"),
            rec("task_session", dry_run=True),
            rec("task_session", unavailable=True),
        ])
        self.assertEqual(q["record_count"], 1)

    def test_pass_rate_and_by_case(self):
        q = gate.compute_quality([
            rec("adr_create", adr_quality=1.0),
            rec("adr_create", adr_quality=0.5),  # fails floor
            rec("canvas_prd"),
        ])
        self.assertEqual(q["record_count"], 3)
        self.assertEqual(q["pass_count"], 2)
        self.assertEqual(q["pass_rate"], round(2 / 3, 4))
        self.assertEqual(q["by_case"]["adr_create"]["pass_rate"], 0.5)
        self.assertEqual(q["by_case"]["canvas_prd"]["pass_rate"], 1.0)

    def test_token_no_go_counted_from_threshold_status(self):
        # Inflated tokens_total alone must NOT trip the guardrail; only the
        # harness's effective-token verdict (threshold_status) does.
        q = gate.compute_quality([
            rec("task_session", tokens=500_000, threshold="ok"),
            rec("task_session", tokens=90_000, threshold="no_go"),
        ])
        self.assertEqual(q["token_no_go_count"], 1)
        self.assertEqual(q["tokens_max"], 500_000)
        self.assertEqual(q["effective_tokens_mean"], 80_000)


class DecideTests(unittest.TestCase):
    def test_establish_when_no_baseline(self):
        cand = gate.compute_quality([rec("task_session")])
        v = gate.decide(cand, {"established": False})
        self.assertEqual(v["decision"], "establish")

    def test_keep_when_held(self):
        cand = gate.compute_quality([rec("task_session"), rec("canvas_prd")])
        base = {"established": True, "pass_rate": 1.0, "by_case": {}}
        self.assertEqual(gate.decide(cand, base)["decision"], "keep")

    def test_keep_when_improved(self):
        cand = gate.compute_quality([rec("task_session"), rec("canvas_prd")])
        base = {"established": True, "pass_rate": 0.5, "by_case": {}}
        v = gate.decide(cand, base)
        self.assertEqual(v["decision"], "keep")
        self.assertEqual(v["pass_rate_delta"], 0.5)

    def test_discard_when_regressed(self):
        cand = gate.compute_quality([rec("task_session", accuracy=0.0), rec("canvas_prd")])
        base = {"established": True, "pass_rate": 1.0, "by_case": {}}
        v = gate.decide(cand, base)
        self.assertEqual(v["decision"], "discard")

    def test_inflated_total_alone_does_not_block(self):
        # A fully-cached run with a huge tokens_total but ok effective spend
        # must still establish/keep — this was the bug the live smoke caught.
        cand = gate.compute_quality([rec("task_session", tokens=423_740, threshold="ok")])
        self.assertEqual(gate.decide(cand, {"established": False})["decision"], "establish")

    def test_discard_on_token_breach_even_if_quality_held(self):
        cand = gate.compute_quality([rec("task_session", threshold="no_go")])
        base = {"established": True, "pass_rate": 1.0, "by_case": {}}
        v = gate.decide(cand, base)
        self.assertEqual(v["decision"], "discard")
        self.assertTrue(v["token_guardrail_breached"])

    def test_establish_blocked_by_token_breach(self):
        cand = gate.compute_quality([rec("task_session", threshold="no_go")])
        v = gate.decide(cand, {"established": False})
        self.assertEqual(v["decision"], "discard")

    def test_per_case_regression_surfaced_in_reasons(self):
        cand = gate.compute_quality([rec("adr_create", adr_quality=0.5), rec("canvas_prd")])
        base = {
            "established": True,
            "pass_rate": 0.0,  # overall won't regress
            "by_case": {"adr_create": {"pass_rate": 1.0}},
        }
        v = gate.decide(cand, base)
        self.assertEqual(v["decision"], "keep")
        self.assertTrue(any("adr_create dropped" in r for r in v["reasons"]))


class CliTests(unittest.TestCase):
    def _write_jsonl(self, path, records):
        path.write_text("".join(json.dumps(r) + "\n" for r in records))

    def test_establish_writes_baseline_and_exits_zero(self):
        with tempfile.TemporaryDirectory() as tmp:
            tmp = Path(tmp)
            cand = tmp / "run.jsonl"
            base = tmp / "baseline.json"
            hist = tmp / "history.jsonl"
            self._write_jsonl(cand, [rec("task_session"), rec("canvas_prd")])
            code = gate.main([
                "--candidate", str(cand), "--baseline", str(base),
                "--history", str(hist), "--update-baseline",
                "--label", "seed", "--timestamp", "1700000000",
            ])
            self.assertEqual(code, 0)
            self.assertTrue(base.exists())
            saved = json.loads(base.read_text())
            self.assertTrue(saved["established"])
            self.assertEqual(saved["pass_rate"], 1.0)
            self.assertEqual(saved["label"], "seed")
            self.assertEqual(len(hist.read_text().strip().splitlines()), 1)

    def test_discard_does_not_overwrite_baseline_and_exits_one(self):
        with tempfile.TemporaryDirectory() as tmp:
            tmp = Path(tmp)
            cand = tmp / "run.jsonl"
            base = tmp / "baseline.json"
            hist = tmp / "history.jsonl"
            base.write_text(json.dumps({"established": True, "pass_rate": 1.0, "by_case": {}}))
            self._write_jsonl(cand, [rec("task_session", accuracy=0.0)])
            code = gate.main([
                "--candidate", str(cand), "--baseline", str(base),
                "--history", str(hist), "--update-baseline",
                "--timestamp", "1700000001",
            ])
            self.assertEqual(code, 1)
            # baseline untouched on discard
            self.assertEqual(json.loads(base.read_text())["pass_rate"], 1.0)
            entry = json.loads(hist.read_text().strip())
            self.assertEqual(entry["decision"], "discard")
            self.assertFalse(entry["baseline_updated"])

    def test_no_history_flag_skips_log(self):
        with tempfile.TemporaryDirectory() as tmp:
            tmp = Path(tmp)
            cand = tmp / "run.jsonl"
            hist = tmp / "history.jsonl"
            self._write_jsonl(cand, [rec("task_session")])
            gate.main([
                "--candidate", str(cand), "--baseline", str(tmp / "b.json"),
                "--history", str(hist), "--no-history", "--timestamp", "1",
            ])
            self.assertFalse(hist.exists())


if __name__ == "__main__":
    unittest.main()
