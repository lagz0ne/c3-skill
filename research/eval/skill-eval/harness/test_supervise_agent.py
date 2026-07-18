#!/usr/bin/env python3

from __future__ import annotations

import json
import subprocess
import tempfile
import unittest
from pathlib import Path


HERE = Path(__file__).resolve().parent
SUPERVISOR = HERE / "bin" / "supervise-agent.py"


class SuperviseAgentTest(unittest.TestCase):
    def run_supervisor(
        self,
        root: Path,
        child: Path,
        *,
        max_tool_calls: int = 8,
        max_tool_result_bytes: int = 100_000,
        max_output_bytes: int = 100_000,
        max_seconds: float = 5,
    ) -> tuple[subprocess.CompletedProcess[str], dict[str, object]]:
        stdin = root / "stdin.txt"
        stdout = root / "stdout.txt"
        stderr = root / "stderr.txt"
        status = root / "status.json"
        stdin.write_text("task\n", encoding="utf-8")
        result = subprocess.run(
            [
                "python3",
                str(SUPERVISOR),
                "--stdin",
                str(stdin),
                "--stdout",
                str(stdout),
                "--stderr",
                str(stderr),
                "--status",
                str(status),
                "--max-seconds",
                str(max_seconds),
                "--max-tool-calls",
                str(max_tool_calls),
                "--max-tool-result-bytes",
                str(max_tool_result_bytes),
                "--max-output-bytes",
                str(max_output_bytes),
                "--",
                "python3",
                str(child),
            ],
            text=True,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            check=False,
        )
        return result, json.loads(status.read_text(encoding="utf-8"))

    def test_kills_stream_after_tool_call_limit(self) -> None:
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            child = root / "child.py"
            child.write_text(
                """import json, time
for index in range(5):
    print(json.dumps({"type": "item.started", "item": {"id": f"tool-{index}", "type": "command_execution"}}), flush=True)
    time.sleep(0.2)
print("SHOULD_NOT_FINISH", flush=True)
""",
                encoding="utf-8",
            )

            result, status = self.run_supervisor(root, child, max_tool_calls=2)

            self.assertEqual(result.returncode, 86, result.stderr)
            self.assertEqual(status["status"], "budget_killed")
            self.assertEqual(status["reason"], "max_tool_calls")
            self.assertEqual(status["tool_calls_observed"], 3)
            self.assertNotIn("SHOULD_NOT_FINISH", (root / "stdout.txt").read_text(encoding="utf-8"))

    def test_counts_claude_tool_use_once_by_id(self) -> None:
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            child = root / "child.py"
            child.write_text(
                """import json
event = {"type": "assistant", "message": {"content": [{"type": "tool_use", "id": "tool-1", "name": "Read"}]}}
print(json.dumps(event), flush=True)
print(json.dumps(event), flush=True)
""",
                encoding="utf-8",
            )

            result, status = self.run_supervisor(root, child)

            self.assertEqual(result.returncode, 0, result.stderr)
            self.assertEqual(status["status"], "completed")
            self.assertEqual(status["tool_calls_observed"], 1)

    def test_kills_stream_after_output_byte_limit(self) -> None:
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            child = root / "child.py"
            child.write_text(
                """import time
for _ in range(20):
    print("x" * 200, flush=True)
    time.sleep(0.05)
""",
                encoding="utf-8",
            )

            result, status = self.run_supervisor(root, child, max_output_bytes=500)

            self.assertEqual(result.returncode, 86, result.stderr)
            self.assertEqual(status["reason"], "max_output_bytes")
            self.assertGreaterEqual(status["output_bytes_observed"], 500)

    def test_kills_stream_after_cumulative_tool_result_limit(self) -> None:
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            child = root / "child.py"
            child.write_text(
                """import json, time
for index in range(3):
    item = {"id": f"tool-{index}", "type": "command_execution", "aggregated_output": "x" * 250}
    print(json.dumps({"type": "item.completed", "item": item}), flush=True)
    time.sleep(0.2)
print("SHOULD_NOT_FINISH", flush=True)
""",
                encoding="utf-8",
            )

            result, status = self.run_supervisor(root, child, max_tool_result_bytes=400)

            self.assertEqual(result.returncode, 86, result.stderr)
            self.assertEqual(status["reason"], "max_tool_result_bytes")
            self.assertGreater(status["tool_result_bytes_observed"], 400)
            self.assertNotIn("SHOULD_NOT_FINISH", (root / "stdout.txt").read_text(encoding="utf-8"))

    def test_kills_child_after_wall_time(self) -> None:
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            child = root / "child.py"
            child.write_text("import time\ntime.sleep(10)\n", encoding="utf-8")

            result, status = self.run_supervisor(root, child, max_seconds=0.2)

            self.assertEqual(result.returncode, 86, result.stderr)
            self.assertEqual(status["reason"], "max_seconds")


if __name__ == "__main__":
    unittest.main()
