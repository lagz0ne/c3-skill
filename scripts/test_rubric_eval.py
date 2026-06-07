import json
import sys
import tempfile
import unittest
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))

import rubric_eval as re_


class ExtractIdsTests(unittest.TestCase):
    def test_extracts_entity_ids_with_boundaries(self):
        ids = re_.extract_ids("c3-114 uses c3-101, c3-102 and c3-105.")
        self.assertEqual(ids, {"c3-114", "c3-101", "c3-102", "c3-105"})

    def test_c3_1_not_extracted_from_c3_114(self):
        ids = re_.extract_ids("the container is c3-114")
        self.assertEqual(ids, {"c3-114"})  # NOT c3-1
        ids2 = re_.extract_ids("it lives in c3-1 (Go CLI)")
        self.assertIn("c3-1", ids2)

    def test_extracts_ref_rule_adr_recipe(self):
        ids = re_.extract_ids("ref-frontmatter-docs, rule-wrap-error-cause, adr-20260320-add-rule-entity-type, recipe-validation-system")
        self.assertEqual(ids, {"ref-frontmatter-docs", "rule-wrap-error-cause",
                               "adr-20260320-add-rule-entity-type", "recipe-validation-system"})


class ScoreAnswerTests(unittest.TestCase):
    def s(self, scorer, answer):
        return re_.score_answer(scorer, answer)[0]

    def test_require_all_present(self):
        self.assertTrue(self.s({"require": ["process", "app"]}, "c3-1 is process, c3-2 is app"))
        self.assertFalse(self.s({"require": ["process", "app"]}, "c3-1 is process"))

    def test_require_case_insensitive(self):
        self.assertTrue(self.s({"require": ["WriteTableOutput"]}, "use writetableoutput"))

    def test_require_any(self):
        self.assertTrue(self.s({"require_any": ["fail", "error"]}, "check would error"))
        self.assertFalse(self.s({"require_any": ["fail", "error"]}, "check passes fine"))

    def test_forbid(self):
        self.assertTrue(self.s({"require_any": ["uncovered"], "forbid": ["c3-104 owns"]}, "it is uncovered"))
        self.assertFalse(self.s({"require_any": ["uncovered"], "forbid": ["c3-104 owns"]}, "uncovered, but c3-104 owns it"))

    def test_ids_include_subset(self):
        self.assertTrue(self.s({"ids_include": ["c3-114", "c3-115", "c3-116"]}, "c3-114, c3-115, c3-116 depend on it"))
        self.assertFalse(self.s({"ids_include": ["c3-114", "c3-115", "c3-116"]}, "only c3-114 and c3-115"))

    def test_id_set_exact_with_subject_excluded(self):
        sc = {"id_set": ["c3-114", "c3-101", "c3-102", "c3-105"], "id_set_exclude_subject": "c3-114"}
        self.assertTrue(self.s(sc, "c3-114 uses c3-101, c3-102, c3-105"))   # subject restated
        self.assertTrue(self.s(sc, "c3-101, c3-102, c3-105"))                # subject omitted
        self.assertFalse(self.s(sc, "c3-101, c3-102"))                       # missing one
        self.assertFalse(self.s(sc, "c3-101, c3-102, c3-105, c3-999"))       # extra wrong one

    def test_min_ids_from(self):
        sc = {"require": ["20"], "min_ids_from": {"set": ["c3-101", "c3-102", "c3-103", "c3-104"], "min": 3}}
        self.assertTrue(self.s(sc, "20 components, e.g. c3-101, c3-102, c3-103"))
        self.assertFalse(self.s(sc, "20 components, e.g. c3-101, c3-102"))   # only 2 named
        self.assertFalse(self.s(sc, "lots of components: c3-101, c3-102, c3-103"))  # missing "20"


class RubricFileTests(unittest.TestCase):
    RUBRIC = Path(__file__).resolve().parent.parent / "research" / "eval" / "rubric.jsonl"

    def test_rubric_loads_and_is_wellformed(self):
        items = re_.load_rubric(self.RUBRIC)
        self.assertGreaterEqual(len(items), 30)
        ids = set()
        for it in items:
            for field in ("id", "category", "question", "ground_truth", "scorer"):
                self.assertIn(field, it, f"{it.get('id')} missing {field}")
            self.assertNotIn(it["id"], ids, f"duplicate id {it['id']}")
            ids.add(it["id"])
            # scorer must have at least one recognized condition
            self.assertTrue(re_.scorer_is_valid(it["scorer"]), f"{it['id']} scorer invalid: {it['scorer']}")

    def test_ground_truth_answers_pass_their_own_scorer(self):
        # Every item's ground_truth string must satisfy its own scorer — a
        # rubric whose own answer fails its scorer is mis-specified.
        items = re_.load_rubric(self.RUBRIC)
        for it in items:
            passed, reasons = re_.score_answer(it["scorer"], it["ground_truth"])
            self.assertTrue(passed, f"{it['id']} ground_truth fails its own scorer: {reasons}\n  gt={it['ground_truth']}\n  scorer={it['scorer']}")


class SummarizeTests(unittest.TestCase):
    def test_summarize_pass_rate_and_by_category(self):
        records = [
            {"id": "NAV-1", "category": "navigation", "agent": "claude", "passed": True},
            {"id": "NAV-2", "category": "navigation", "agent": "claude", "passed": False},
            {"id": "OWN-1", "category": "ownership", "agent": "claude", "passed": True},
        ]
        s = re_.summarize(records)
        self.assertEqual(s["record_count"], 3)
        self.assertEqual(s["pass_count"], 2)
        self.assertEqual(s["pass_rate"], round(2 / 3, 4))
        self.assertEqual(s["by_category"]["navigation"]["pass_rate"], 0.5)
        self.assertEqual(s["by_category"]["ownership"]["pass_rate"], 1.0)

    def test_unavailable_agent_excluded_from_denominator(self):
        records = [
            {"id": "NAV-1", "category": "navigation", "agent": "claude", "passed": True},
            {"id": "NAV-1", "category": "navigation", "agent": "codex", "passed": False, "agent_unavailable": True},
            {"id": "NAV-2", "category": "navigation", "agent": "codex", "passed": True, "agent_unavailable": True},
        ]
        s = re_.summarize(records)
        self.assertEqual(s["record_count"], 1)        # only the available claude record
        self.assertEqual(s["unavailable_count"], 2)
        self.assertEqual(s["pass_rate"], 1.0)
        self.assertNotIn("codex", s["by_agent"])      # unavailable agent absent from scored breakdown


class DryRunTests(unittest.TestCase):
    def test_dry_run_writes_plan_without_agents(self):
        with tempfile.TemporaryDirectory() as tmp:
            out = Path(tmp) / "plan.jsonl"
            rc = re_.main(["--dry-run", "--agent", "claude", "--output", str(out)])
            self.assertEqual(rc, 0)
            lines = [l for l in out.read_text().splitlines() if l.strip()]
            self.assertGreaterEqual(len(lines), 30)
            rec = json.loads(lines[0])
            self.assertTrue(rec["dry_run"])
            self.assertIn("question", rec)


if __name__ == "__main__":
    unittest.main()
