#!/usr/bin/env python3

from __future__ import annotations

import importlib.util
import unittest
from pathlib import Path


MODULE_PATH = Path(__file__).with_name("condition_blind_score.py")
SPEC = importlib.util.spec_from_file_location("condition_blind_score", MODULE_PATH)
assert SPEC and SPEC.loader
score = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(score)


def mappings(*, swap_j2: bool = False) -> list[dict[str, object]]:
    return [
        {
            "judge_id": "J1",
            "case_id": "CASE-1",
            "aliases": {"A": "with_c3", "B": "without_c3"},
        },
        {
            "judge_id": "J2",
            "case_id": "CASE-1",
            "aliases": (
                {"A": "without_c3", "B": "with_c3"}
                if swap_j2
                else {"A": "with_c3", "B": "without_c3"}
            ),
        },
    ]


def judgments(*, swap_j2: bool = False) -> list[dict[str, object]]:
    j2 = (
        {
            "judge_id": "J2",
            "case_id": "CASE-1",
            "answers": {
                "A": {"gold": "NS", "forbidden": "S"},
                "B": {"gold": "SS", "forbidden": "N"},
            },
        }
        if swap_j2
        else {
            "judge_id": "J2",
            "case_id": "CASE-1",
            "answers": {
                "A": {"gold": "SS", "forbidden": "N"},
                "B": {"gold": "NS", "forbidden": "S"},
            },
        }
    )
    return [
        {
            "judge_id": "J1",
            "case_id": "CASE-1",
            "answers": {
                "A": {"gold": "SS", "forbidden": "N"},
                "B": {"gold": "NS", "forbidden": "S"},
            },
        },
        j2,
    ]


class ConditionBlindScoreTest(unittest.TestCase):
    counts = {"CASE-1": {"gold": 2, "forbidden": 1}}

    def test_independent_alias_swap_is_metric_invariant(self) -> None:
        expected = score.score_condition_blind(
            judgments(), mappings(), self.counts
        )
        actual = score.score_condition_blind(
            judgments(swap_j2=True), mappings(swap_j2=True), self.counts
        )

        self.assertEqual(actual, expected)
        self.assertEqual(actual["claim_label_agreement"], 1.0)
        self.assertEqual(actual["required_surface_lift"], 0.5)
        self.assertEqual(actual["blocking_false_claim_lift"], -1.0)

    def test_input_order_and_display_aliases_do_not_change_metrics(self) -> None:
        expected = score.score_condition_blind(
            judgments(), mappings(), self.counts
        )
        actual = score.score_condition_blind(
            list(reversed(judgments())), list(reversed(mappings())), self.counts
        )

        self.assertEqual(actual, expected)

    def test_alias_mapping_must_be_bijective(self) -> None:
        invalid = mappings()
        invalid[0]["aliases"] = {"A": "with_c3", "B": "with_c3"}

        with self.assertRaisesRegex(score.ScoreInputError, "bijection"):
            score.score_condition_blind(judgments(), invalid, self.counts)

    def test_missing_or_duplicate_judge_case_fails_closed(self) -> None:
        with self.assertRaisesRegex(score.ScoreInputError, "missing judgment"):
            score.score_condition_blind(judgments()[:1], mappings(), self.counts)

        duplicate = judgments() + [judgments()[0]]
        with self.assertRaisesRegex(score.ScoreInputError, "duplicate judgment"):
            score.score_condition_blind(duplicate, mappings(), self.counts)

    def test_stable_case_ids_and_proposition_counts_are_required(self) -> None:
        bad_case = judgments()
        bad_case[0]["case_id"] = "DISPLAY-1"
        with self.assertRaisesRegex(score.ScoreInputError, "mapping"):
            score.score_condition_blind(bad_case, mappings(), self.counts)

        bad_count = {"CASE-1": {"gold": 3, "forbidden": 1}}
        with self.assertRaisesRegex(score.ScoreInputError, "gold count"):
            score.score_condition_blind(judgments(), mappings(), bad_count)

    def test_missing_alias_cell_and_invalid_label_fail_closed(self) -> None:
        missing = judgments()
        del missing[0]["answers"]["B"]
        with self.assertRaisesRegex(score.ScoreInputError, "aliases"):
            score.score_condition_blind(missing, mappings(), self.counts)

        invalid = judgments()
        invalid[0]["answers"]["A"]["gold"] = "SX"
        with self.assertRaisesRegex(score.ScoreInputError, "invalid gold label"):
            score.score_condition_blind(invalid, mappings(), self.counts)


if __name__ == "__main__":
    unittest.main()
