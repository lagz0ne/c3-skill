#!/usr/bin/env python3

from __future__ import annotations

import unittest

from scripts.paired_skill_cost_admission import estimate


class PairedSkillCostAdmissionTest(unittest.TestCase):
    def test_admits_v28_bounded_sol_arm(self) -> None:
        result = estimate(
            base_input_bytes=15_000,
            max_tool_result_bytes=8_000,
            tool_calls=3,
            max_output_tokens=700,
            bytes_per_token=3,
            input_per_million_usd=2,
            cached_input_per_million_usd=0.5,
            output_per_million_usd=8,
            ceiling_usd=0.034,
        )

        self.assertTrue(result.admitted)
        self.assertEqual(result.base_input_tokens, 5_000)
        self.assertEqual(result.tool_result_tokens, 2_667)
        self.assertEqual(result.estimated_cost_usd, 0.031101)
        self.assertGreater(result.headroom_usd, 0.0028)

    def test_rejects_when_tool_transport_is_not_compact(self) -> None:
        result = estimate(
            base_input_bytes=15_000,
            max_tool_result_bytes=18_000,
            tool_calls=3,
            max_output_tokens=700,
            bytes_per_token=3,
            input_per_million_usd=2,
            cached_input_per_million_usd=0.5,
            output_per_million_usd=8,
            ceiling_usd=0.034,
        )

        self.assertFalse(result.admitted)
        self.assertLess(result.headroom_usd, 0)


if __name__ == "__main__":
    unittest.main()
