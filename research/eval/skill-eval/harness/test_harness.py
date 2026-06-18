#!/usr/bin/env python3

from __future__ import annotations

import subprocess
import unittest
from pathlib import Path


HERE = Path(__file__).resolve().parent
BUILD_PROMPT = HERE / "bin" / "build-agent-prompt.sh"
RUN_BLINDBOX = HERE / "bin" / "run-blindbox.sh"
TOPIC = "grow-warehouse-system"


class BlindboxHarnessTest(unittest.TestCase):
    def test_build_agent_prompt_includes_local_c3_and_topic(self) -> None:
        result = subprocess.run(
            [str(BUILD_PROMPT), TOPIC],
            text=True,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            check=True,
        )

        self.assertIn("===== BEGIN FILE: skills/c3/SKILL.md =====", result.stdout)
        self.assertIn("===== BEGIN FILE: skills/c3/references/change.md =====", result.stdout)
        self.assertIn("===== BEGIN FILE: harness/prompts/agent-run.md =====", result.stdout)
        self.assertIn("===== BEGIN FILE: harness/topics/grow-warehouse-system/prompt.md =====", result.stdout)
        self.assertIn("frontend, backend", result.stdout)
        self.assertIn("integration, and database", result.stdout)
        self.assertIn("/opt/c3/skills/c3/bin/c3x.sh", result.stdout)

    def test_run_blindbox_dry_run_reports_isolation_metadata(self) -> None:
        result = subprocess.run(
            [
                str(RUN_BLINDBOX),
                "--agent",
                "codex",
                "--topic",
                TOPIC,
                "--auth",
                "env",
                "--label",
                "unit-dryrun",
                "--dry-run",
            ],
            text=True,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            check=True,
        )

        self.assertIn("mounted_repo_root=no", result.stdout)
        self.assertIn("mounted_global_home=no", result.stdout)
        self.assertIn("local_c3=/opt/c3/skills/c3/bin/c3x.sh", result.stdout)
        self.assertIn("workspace=", result.stdout)
        self.assertIn("--ro-bind", result.stdout)
        self.assertIn("/opt/c3/skills/c3", result.stdout)

    def test_run_blindbox_rejects_unknown_topic(self) -> None:
        result = subprocess.run(
            [
                str(RUN_BLINDBOX),
                "--agent",
                "codex",
                "--topic",
                "missing-topic",
                "--dry-run",
            ],
            text=True,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            check=False,
        )

        self.assertNotEqual(result.returncode, 0)
        self.assertIn("Unknown topic", result.stderr)


if __name__ == "__main__":
    unittest.main()
