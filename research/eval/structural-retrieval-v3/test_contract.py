"""Generic contract checks for the frozen v3 fixture and benchmark files."""

import hashlib
import json
import pathlib
import unittest


ROOT = pathlib.Path(__file__).parent
FIXTURE = ROOT / "fixtures.v3.json"
BENCHMARK = ROOT / "benchmark.v3.json"


class V3ArtifactContractTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.fixtures = json.loads(FIXTURE.read_text())
        cls.benchmark = json.loads(BENCHMARK.read_text())

    def test_exact_case_set_and_translation_metadata(self):
        ids = [f["case_id"] for f in self.fixtures]
        want = [
            "v2-owner-1",
            "v2-owner-2",
            "v2-owner-3",
            "v2-route",
            "counter-multi-peer-owner",
            "counter-no-target",
            "counter-bound-route",
        ]
        self.assertEqual(ids, want)
        self.assertEqual(self.benchmark["fixture_count"], 7)
        self.assertEqual(self.benchmark["case_ids"], want)
        self.assertEqual(self.benchmark["baseline_file"], "B-v3-baseline.json")
        self.assertEqual(self.benchmark["baseline_status"], "captured_unchanged_controller")
        self.assertEqual(self.benchmark["baseline_capture_script"], "cli/tools/structural-search-eval-v3/cli.go")
        self.assertTrue((ROOT.parents[2] / self.benchmark["baseline_capture_script"]).exists())
        translated = self.benchmark["translation_metadata"]["translated_cases"]
        self.assertEqual(set(translated), set(want[:4]))
        for value in translated.values():
            self.assertIn(value["v2_shape"], {"wrong_layer_structural_owner", "behavioral_route_regression"})
            self.assertIn(value["v2_case_ordinal"], {1, 2, 3, 4})
        for key in (
            "source_v2_fixture_sha256",
            "source_v2_benchmark_sha256",
            "source_v2_evaluator_sha256",
            "source_v2_evaluator_tests_sha256",
        ):
            self.assertRegex(self.benchmark["translation_metadata"][key], r"^[0-9a-f]{64}$")

    def test_every_entity_and_fact_has_one_role_binding(self):
        roles = {"owner", "neutral", "forbidden", "unsupported", "unknown"}
        for fixture in self.fixtures:
            entities = {e["id"]: e for e in fixture["entities"]}
            facts = {fact["id"]: fact for fact in fixture["facts"]}
            self.assertEqual(len(entities), len(fixture["entities"]))
            self.assertEqual(len(facts), len(fixture["facts"]))
            for entity in entities.values():
                self.assertIn(entity["role"], roles)
            for fact in facts.values():
                self.assertIn(fact["entity_id"], entities)
                self.assertEqual(fact["role"], entities[fact["entity_id"]]["role"])
            oracle = fixture["oracle"]
            partitions = {
                "owner": set(oracle.get("required_owner_fact_ids", [])),
                "neutral": set(oracle.get("neutral_fact_ids", [])),
                "forbidden": set(oracle.get("forbidden_fact_ids", [])),
                "unsupported": set(oracle.get("unsupported_fact_ids", [])),
                "unknown": set(oracle.get("unknown_fact_ids", [])),
            }
            flat = [item for values in partitions.values() for item in values]
            self.assertEqual(len(flat), len(set(flat)))
            for role, ids in partitions.items():
                for fact_id in ids:
                    self.assertIn(fact_id, facts)
                    self.assertEqual(facts[fact_id]["role"], role)
            context = set(oracle.get("required_context_entity_ids", []))
            self.assertEqual(
                context,
                {facts[fact_id]["entity_id"] for fact_id in partitions["neutral"]},
            )
            bound = set(entities)
            bound_by_facts = {fact["entity_id"] for fact in facts.values()}
            self.assertEqual(bound, bound_by_facts)

    def test_countercase_shapes_and_no_target_policy(self):
        by_id = {f["case_id"]: f for f in self.fixtures}
        multi = by_id["counter-multi-peer-owner"]
        self.assertEqual(multi["oracle"]["required_owner_fact_ids"], ["mp-owner-fact"])
        self.assertEqual(multi["oracle"]["neutral_fact_ids"], ["mp-peer-a-fact", "mp-peer-b-fact"])
        self.assertEqual(multi["oracle"]["required_context_entity_ids"], ["mp-peer-a", "mp-peer-b"])
        self.assertEqual(multi["oracle"]["forbidden_fact_ids"], ["mp-child-fact"])

        no_target = by_id["counter-no-target"]["oracle"]
        self.assertEqual(no_target.get("required_owner_fact_ids", []), [])
        self.assertEqual(no_target["neutral_fact_ids"], ["nt-context-a-fact", "nt-context-b-fact"])
        self.assertEqual(no_target["forbidden_fact_ids"], ["nt-forbidden-fact"])
        self.assertEqual(no_target["no_target_policy"], {"owner_witness_count": 0, "preserve_neutral_context": True, "v2_action": "omit", "v3_actions": ["omit", "flagged"]})

        route = by_id["counter-bound-route"]
        self.assertEqual(route["oracle"]["unsupported_entity_ids"], ["br-unbound", "br-unbound-anchor"])
        self.assertEqual(route["oracle"]["unsupported_fact_ids"], ["br-unbound-fact", "br-unbound-anchor-fact"])
        witness = route["oracle"]["bound_route_witnesses"][0]
        self.assertEqual(witness["entity_content_id"], "content:br-owner")
        self.assertEqual(witness["match_source"], "graph:uses:br-anchor")
        self.assertEqual(witness["graph_from_id"], "br-owner")
        self.assertEqual(witness["graph_to_id"], "br-anchor")
        self.assertEqual(witness["direct_fts_entity_miss_id"], "br-owner")
        self.assertEqual(witness["direct_fts_content_miss_id"], "content:br-owner")
        self.assertEqual(witness["expected_route_field_values"], {"facts": ["br-owner-fact"], "graph": ["graph:uses:br-anchor"], "lanes": ["behavioral-route"], "hash": "route:br-owner:br-anchor"})

    def test_threshold_scope_privacy_and_frozen_hashes(self):
        self.assertEqual(self.benchmark["thresholds"], {"owner_recall_at_5_delta": 0.2, "structural_owner_precision": 0.8, "canonical_row_bytes_ratio": 1.05})
        self.assertEqual(self.benchmark["scope"]["top5"], "rows[0:5] only; duplicate entity IDs in this scope reject before scoring")
        self.assertIn("same-case fresh B-v3", self.benchmark["scope"]["fresh_b_v3_wall"])
        privacy = self.benchmark["privacy"]
        self.assertEqual(len(privacy["scan_scope"]), 7)
        self.assertFalse(privacy["raw_terms_retained"])
        self.assertEqual(hashlib.sha256((ROOT / "fixtures.v3.json").read_bytes()).hexdigest(), "b5afc99c59f9ff9909bfd56c0810840002c6a3b427d04bfe0bcccb3a554cc77f")
        self.assertEqual(hashlib.sha256((ROOT / "benchmark.v3.json").read_bytes()).hexdigest(), "76724b21735f0fb0f564e9fe4af71f6f46a012e52cc1ca0b59df8bc7579a04fe")


if __name__ == "__main__":
    unittest.main()
