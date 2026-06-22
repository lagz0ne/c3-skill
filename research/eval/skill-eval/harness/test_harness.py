#!/usr/bin/env python3

from __future__ import annotations

import os
import re
import subprocess
import tempfile
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

    def test_build_agent_prompt_records_source_list(self) -> None:
        with tempfile.TemporaryDirectory() as td:
            source_list = Path(td) / "sources.tsv"
            env = os.environ.copy()
            env["C3_PROMPT_SOURCE_LIST"] = str(source_list)
            subprocess.run(
                [str(BUILD_PROMPT), TOPIC],
                text=True,
                stdout=subprocess.PIPE,
                stderr=subprocess.PIPE,
                check=True,
                env=env,
            )

            rows = source_list.read_text(encoding="utf-8").splitlines()
        self.assertTrue(any(row.startswith("skills/c3/SKILL.md\t") for row in rows))
        self.assertTrue(any(row.startswith("skills/c3/references/change.md\t") for row in rows))
        self.assertTrue(any(row.startswith("harness/topics/grow-warehouse-system/prompt.md\t") for row in rows))

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
        self.assertIn("prompt_archive=", result.stdout)
        self.assertIn("provenance=", result.stdout)
        self.assertIn("selected_binary_sha256=", result.stdout)
        self.assertIn("prompt_sha256=", result.stdout)
        self.assertIn("global_c3_guard=/opt/c3/guard/bin/c3x", result.stdout)
        self.assertIn("path_has_usr_local_bin=no", result.stdout)
        self.assertRegex(result.stdout, r"--setenv PATH [^\n]*/opt/c3/guard/bin")
        self.assertNotIn("/usr/local/bin", result.stdout)

        sha_lines = [
            line for line in result.stdout.splitlines()
            if line.split("=", 1)[0].endswith("_sha256") or line.startswith("source_sha256[")
        ]
        self.assertGreaterEqual(len(sha_lines), 5)
        for line in sha_lines:
            self.assertRegex(line, r"=[0-9a-f]{64}$")

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

    def test_run_blindbox_rejects_wrong_binary_hash(self) -> None:
        env = os.environ.copy()
        env["C3_EXPECT_SELECTED_BINARY_SHA256"] = "0" * 64
        result = subprocess.run(
            [
                str(RUN_BLINDBOX),
                "--agent",
                "codex",
                "--topic",
                TOPIC,
                "--dry-run",
            ],
            text=True,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            check=False,
            env=env,
        )

        self.assertNotEqual(result.returncode, 0)
        self.assertIn("Hash mismatch for selected_binary", result.stderr)

    def test_run_blindbox_rejects_wrong_skill_source_hash(self) -> None:
        env = os.environ.copy()
        env["C3_EXPECT_SKILL_MD_SHA256"] = "0" * 64
        result = subprocess.run(
            [
                str(RUN_BLINDBOX),
                "--agent",
                "codex",
                "--topic",
                TOPIC,
                "--dry-run",
            ],
            text=True,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            check=False,
            env=env,
        )

        self.assertNotEqual(result.returncode, 0)
        self.assertIn("Hash mismatch for skills/c3/SKILL.md", result.stderr)


if __name__ == "__main__":
    unittest.main()
