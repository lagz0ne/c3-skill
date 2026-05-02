import json
import sys
import tempfile
import unittest
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))

import agent_efficiency_eval as ev


class AgentEfficiencyEvalTests(unittest.TestCase):
    def test_default_matrix_has_six_cases_and_two_agents(self):
        self.assertEqual(
            [case.id for case in ev.default_cases()],
            [
                "skill_task_session",
                "skill_content_limit_adr",
                "adr_create",
                "task_session",
                "debug_session",
                "system_design_change",
            ],
        )
        self.assertEqual([agent.id for agent in ev.default_agents()], ["claude", "codex"])

    def test_dry_run_builds_full_matrix_without_execution(self):
        plan = ev.build_plan(ev.default_cases(), ev.default_agents(), dry_run=True)

        self.assertEqual(len(plan), 12)
        self.assertTrue(all(item.dry_run for item in plan))
        self.assertTrue(all(item.command for item in plan))
        self.assertEqual(
            {(item.agent_id, item.case_id) for item in plan},
            {(agent.id, case.id) for agent in ev.default_agents() for case in ev.default_cases()},
        )

    def test_score_result_handles_missing_token_metadata(self):
        result = ev.RunResult(
            agent_id="codex",
            case_id="debug_session",
            command=["codex", "exec", "prompt"],
            dry_run=False,
            exit_code=0,
            elapsed_ms=1200,
            stdout="task complete\nturn 1\nturn 2\n",
            stderr="",
            output_dir="/tmp/eval",
            artifact_dir="/tmp/artifacts",
            accuracy_checks={"mentioned_root_cause": True, "verified": False},
            token_usage=None,
            turn_count=None,
            trace_metrics={},
        )

        scored = ev.score_result(result)

        self.assertEqual(scored["agent"], "codex")
        self.assertEqual(scored["case"], "debug_session")
        self.assertIsNone(scored["tokens_total"])
        self.assertEqual(scored["turn_count"], 2)
        self.assertEqual(scored["accuracy_score"], 0.5)

    def test_extract_token_usage_reads_codex_json_events(self):
        text = (
            '{"type":"turn.completed","usage":'
            '{"input_tokens":10,"cached_input_tokens":4,'
            '"output_tokens":3,"reasoning_output_tokens":2}}\n'
        )

        usage = ev.extract_token_usage(text)

        self.assertEqual(usage["input_tokens"], 10)
        self.assertEqual(usage["cached_input_tokens"], 4)
        self.assertEqual(usage["output_tokens"], 3)
        self.assertEqual(usage["reasoning_output_tokens"], 2)
        self.assertEqual(usage["total_tokens"], 13)

    def test_extract_turn_count_reads_codex_turn_completion(self):
        text = '{"type":"turn.completed","usage":{"input_tokens":1,"output_tokens":1}}\n'

        self.assertEqual(ev.extract_turn_count(text), 1)

    def test_extract_trace_metrics_finds_c3_commands(self):
        text = 'C3X_MODE=agent bash skills/c3/bin/c3x.sh lookup cli/cmd/list.go\n{"type":"turn.completed"}\n'

        metrics = ev.extract_trace_metrics(text)

        self.assertEqual(metrics["json_event_count"], 1)
        self.assertEqual(metrics["c3_command_count"], 1)
        self.assertIn("lookup cli/cmd/list.go", metrics["c3_command_sequence"][0])

    def test_run_writes_jsonl_records_for_dry_run(self):
        with tempfile.TemporaryDirectory() as tmp:
            out = Path(tmp) / "results.jsonl"

            code = ev.main(["--dry-run", "--output", str(out)])

            self.assertEqual(code, 0)
            records = [json.loads(line) for line in out.read_text().splitlines()]
            self.assertEqual(len(records), 12)
            self.assertTrue(all(record["dry_run"] for record in records))
            self.assertEqual({record["accuracy_score"] for record in records}, {0.0})
            self.assertEqual(
                {record["metric_basis"] for record in records},
                {"tokens_total, turn_count, accuracy_score, elapsed_ms"},
            )
            self.assertTrue(all(Path(record["artifact_dir"]).exists() for record in records))

    def test_live_run_scores_eval_result_before_cleanup(self):
        original_copy = ev._copy_controlled_workspace
        ev._copy_controlled_workspace = lambda workspace: None
        try:
            case = next(case for case in ev.default_cases() if case.id == "task_session")
            item = ev.PlanItem(
                agent_id="fake",
                case_id=case.id,
                command=[
                    sys.executable,
                    "-c",
                    (
                        "import json; "
                        "open('eval_result.json','w').write(json.dumps("
                        "{'verified': True, 'artifacts': ['owner:c3-112']}))"
                    ),
                ],
                prompt=case.prompt,
                dry_run=False,
            )

            scored = ev.score_result(ev.run_plan_item(item, case, keep_workspace=False))

            self.assertEqual(scored["accuracy_score"], 1.0)
            self.assertFalse(Path(scored["output_dir"]).exists())
            self.assertTrue(Path(scored["artifact_dir"], "eval_result.json").exists())
        finally:
            ev._copy_controlled_workspace = original_copy


if __name__ == "__main__":
    unittest.main()
