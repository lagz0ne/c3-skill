#!/usr/bin/env python3

from __future__ import annotations

import os
import re
import shutil
import subprocess
import tempfile
import unittest
from pathlib import Path


HERE = Path(__file__).resolve().parent
BUILD_PROMPT = HERE / "bin" / "build-agent-prompt.sh"
RUN_BLINDBOX = HERE / "bin" / "run-blindbox.sh"
USAGE_PROXY = HERE / "bin" / "c3-usage-proxy.sh"
COMMAND_GATE = HERE / "bin" / "c3-command-gate.sh"
IMPACT_BOOTSTRAP = HERE / "bin" / "c3-impact-bootstrap.sh"
TMUX_BACKEND = HERE / "bin" / "run-in-tmux.sh"
TOPIC = "grow-warehouse-system"


class BlindboxHarnessTest(unittest.TestCase):
    def test_treatment_rebuilds_disposable_c3_cache_before_agent_start(self) -> None:
        source = RUN_BLINDBOX.read_text(encoding="utf-8")

        self.assertIn('c3_cache_prepare_status=ready', source)
        self.assertIn('C3X_MODE=agent C3X_LOCAL_BINARY="$C3_LOCAL_BINARY"', source)
        self.assertIn('bash "$repo_root/skills/c3/bin/c3x.sh" check', source)

    def test_treatment_can_mount_an_explicit_frozen_c3_binary(self) -> None:
        source = RUN_BLINDBOX.read_text(encoding="utf-8")

        self.assertIn('C3_FROZEN_BINARY', source)
        self.assertIn('--ro-bind "$C3_LOCAL_BINARY" /opt/c3/bin/frozen-c3x', source)
        self.assertIn('--setenv C3X_LOCAL_BINARY /opt/c3/bin/frozen-c3x', source)

    def test_treatment_mounts_a_writable_c3_cache_inside_the_sandbox(self) -> None:
        source = RUN_BLINDBOX.read_text(encoding="utf-8")

        self.assertIn('c3_sandbox_cache_dir="$c3_usage_dir/cache"', source)
        self.assertIn('mkdir -p "$c3_sandbox_cache_dir"', source)
        self.assertIn('--setenv XDG_CACHE_HOME /work/project/.c3-eval/cache', source)
        self.assertNotIn('--bind "$c3_sandbox_cache_dir" /work/home/.cache', source)

    def test_impact_bootstrap_falls_back_then_runs_graph_and_evidence(self) -> None:
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            calls = root / "calls.txt"
            wrapper = root / "wrapper.sh"
            wrapper.write_text(
                """#!/usr/bin/env bash
printf '%s\\n' "$*" >> "$BOOTSTRAP_CALLS"
case "$1" in
  search) exit 1 ;;
  list) printf 'entities[1]{id,type,title}:\\n  c3-123,component,Example\\n' ;;
  graph) printf 'graph: ok\\n' ;;
  read) printf 'evidence: ok\\n' ;;
  *) exit 2 ;;
esac
""",
                encoding="utf-8",
            )
            wrapper.chmod(0o755)
            env = os.environ.copy()
            env["BOOTSTRAP_CALLS"] = str(calls)
            bootstrap = root / "bootstrap.sh"
            bootstrap.write_text(
                IMPACT_BOOTSTRAP.read_text(encoding="utf-8").replace(
                    'wrapper="/opt/c3/skills/c3/bin/c3x.sh"',
                    f'wrapper="{wrapper}"',
                ).replace('skill="/opt/c3/skills/c3/SKILL.md"', f'skill="{wrapper}"').replace(
                    'sweep="/opt/c3/skills/c3/references/sweep.md"',
                    f'sweep="{wrapper}"',
                ),
                encoding="utf-8",
            )

            result = subprocess.run(
                ["bash", str(bootstrap), "approval", "notification"],
                text=True,
                stdout=subprocess.PIPE,
                stderr=subprocess.PIPE,
                check=False,
                env=env,
            )
            call_lines = calls.read_text(encoding="utf-8").splitlines()

        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertEqual(
            call_lines,
            [
                "search approval notification --pack --limit 3",
                "list --type component",
                "graph c3-123 --depth 1 --format mermaid",
                "graph c3-123 --direction reverse --depth 1 --format mermaid",
                "read c3-123",
            ],
        )
        self.assertIn("selected_id=c3-123", result.stderr)

    def test_impact_bootstrap_uses_absolute_parser_tools_without_an_agent_bypass(self) -> None:
        source = IMPACT_BOOTSTRAP.read_text(encoding="utf-8")

        self.assertNotIn("C3_GATE_BYPASS", source)
        self.assertNotIn('/bin/cat "$skill"', source)
        self.assertNotIn('/bin/cat "$sweep"', source)
        self.assertIn('[[ ! -s "$skill" || ! -s "$sweep" ]]', source)
        self.assertIn("instruction_sources=skills/c3/SKILL.md,skills/c3/references/sweep.md", source)
        self.assertIn("/usr/bin/awk", source)
        self.assertIn("/usr/bin/head", source)
        self.assertIn("/usr/bin/head -c 1800", source)
        self.assertIn("/usr/bin/head -c 240", source)
        self.assertIn("/usr/bin/head -c 350", source)
        self.assertNotIn('/bin/cat "$route_output"', source)
        self.assertIn('graph "$selected_id" --direction reverse --depth 1 --format mermaid', source)
        self.assertIn('read "$selected_id"', source)
        self.assertNotIn('read "$selected_id" --full', source)

    def test_treatment_spends_remaining_tool_budget_on_cross_cutting_source_proof(self) -> None:
        source = (HERE / "prompts" / "c3-impact-treatment.md").read_text(encoding="utf-8")

        self.assertIn("forward and", source)
        self.assertIn("reverse relationships", source)
        self.assertIn("remaining calls", source)
        self.assertIn("source-close every requested lane", source)
        self.assertIn("zero matching record claims", source)
        self.assertIn("no more than ten additional tool calls", source)
        self.assertIn("affected, unaffected, or unknown", source)
        self.assertNotIn("at most two narrow source checks", source)

    def test_impact_bootstrap_falls_back_when_search_succeeds_empty(self) -> None:
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            calls = root / "calls.txt"
            wrapper = root / "wrapper.sh"
            wrapper.write_text(
                """#!/usr/bin/env bash
printf '%s\\n' "$*" >> "$BOOTSTRAP_CALLS"
case "$1" in
  search) exit 0 ;;
  list) printf 'entities[1]{id,type,title}:\\n  c3-456,component,Example\\n' ;;
  graph) printf 'graph: ok\\n' ;;
  read) printf 'evidence: ok\\n' ;;
  *) exit 2 ;;
esac
""",
                encoding="utf-8",
            )
            wrapper.chmod(0o755)
            env = os.environ.copy()
            env["BOOTSTRAP_CALLS"] = str(calls)
            bootstrap = root / "bootstrap.sh"
            bootstrap.write_text(
                IMPACT_BOOTSTRAP.read_text(encoding="utf-8").replace(
                    'wrapper="/opt/c3/skills/c3/bin/c3x.sh"',
                    f'wrapper="{wrapper}"',
                ).replace('skill="/opt/c3/skills/c3/SKILL.md"', f'skill="{wrapper}"').replace(
                    'sweep="/opt/c3/skills/c3/references/sweep.md"',
                    f'sweep="{wrapper}"',
                ),
                encoding="utf-8",
            )

            result = subprocess.run(
                ["bash", str(bootstrap), "archival"],
                text=True,
                stdout=subprocess.PIPE,
                stderr=subprocess.PIPE,
                check=False,
                env=env,
            )
            call_lines = calls.read_text(encoding="utf-8").splitlines()

        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertIn("closure: owner/mutation", result.stdout)
        self.assertIn("persistence/event/retry", result.stdout)
        self.assertIn("unsupported impact stays unknown", result.stdout)
        self.assertEqual(
            call_lines,
            [
                "search archival --pack --limit 3",
                "list --type component",
                "graph c3-456 --depth 1 --format mermaid",
                "graph c3-456 --direction reverse --depth 1 --format mermaid",
                "read c3-456",
            ],
        )

    @unittest.skipUnless(shutil.which("tmux"), "tmux is not installed")
    def test_tmux_backend_returns_child_output_and_status(self) -> None:
        session = f"c3-harness-test-{os.getpid()}"
        subprocess.run(
            ["tmux", "kill-session", "-t", session],
            stdout=subprocess.DEVNULL,
            stderr=subprocess.DEVNULL,
            check=False,
        )

        result = subprocess.run(
            [
                str(TMUX_BACKEND),
                "--session",
                session,
                "--",
                "bash",
                "-c",
                "printf 'tmux-out\\n'; printf 'tmux-error\\n' >&2; exit 7",
            ],
            text=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE,
            check=False,
        )

        self.assertEqual(result.returncode, 7)
        self.assertIn("tmux-out", result.stdout)
        self.assertIn("tmux-error", result.stdout)
        self.assertFalse(
            subprocess.run(
                ["tmux", "has-session", "-t", session],
                stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL,
                check=False,
            ).returncode == 0
        )

    def test_c3_usage_proxy_records_only_category_and_status(self) -> None:
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            real = root / "real.sh"
            real.write_text("#!/usr/bin/env bash\necho routed\n", encoding="utf-8")
            real.chmod(0o755)
            log = root / "usage.tsv"
            env = os.environ.copy()
            env["C3_REAL_WRAPPER"] = str(real)
            env["C3_USAGE_LOG"] = str(log)

            result = subprocess.run(
                ["bash", str(USAGE_PROXY), "search", "private concept"],
                text=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE,
                check=False, env=env,
            )

            self.assertEqual(result.returncode, 0, result.stderr)
            self.assertEqual(log.read_text(encoding="utf-8"), "route\t0\n")
            self.assertNotIn("private concept", log.read_text(encoding="utf-8"))

    def test_c3_usage_proxy_counts_list_as_a_route_fallback(self) -> None:
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            real = root / "real.sh"
            real.write_text("#!/usr/bin/env bash\necho listed\n", encoding="utf-8")
            real.chmod(0o755)
            log = root / "usage.tsv"
            env = os.environ.copy()
            env["C3_REAL_WRAPPER"] = str(real)
            env["C3_USAGE_LOG"] = str(log)

            result = subprocess.run(
                ["bash", str(USAGE_PROXY), "list", "--type", "component"],
                text=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE,
                check=False, env=env,
            )

            self.assertEqual(result.returncode, 0, result.stderr)
            self.assertEqual(log.read_text(encoding="utf-8"), "route\t0\n")

    def test_c3_usage_proxy_uses_a_clean_system_path_for_c3_internal_commands(self) -> None:
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            real = root / "real.sh"
            real.write_text(
                "#!/usr/bin/env bash\nprintf '%s\\n' \"$PATH\"\nprintf '%s\\n' \"${C3_GATE_BYPASS:-missing}\"\n",
                encoding="utf-8",
            )
            real.chmod(0o755)
            log = root / "usage.tsv"
            env = os.environ.copy()
            env["C3_REAL_WRAPPER"] = str(real)
            env["C3_USAGE_LOG"] = str(log)

            result = subprocess.run(
                ["bash", str(USAGE_PROXY), "search", "private concept"],
                text=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE,
                check=False, env=env,
            )

            self.assertEqual(result.returncode, 0, result.stderr)
            lines = result.stdout.splitlines()
            self.assertEqual(lines[0], "/usr/bin:/bin")
            self.assertEqual(lines[1], "missing")
            self.assertEqual(log.read_text(encoding="utf-8"), "route\t0\n")

    def run_command_gate(self, usage_rows: str) -> tuple[subprocess.CompletedProcess[str], bool]:
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            usage = root / "usage.tsv"
            usage.write_text(usage_rows, encoding="utf-8")
            marker = root / "real-ran"
            real = root / "real-tool.sh"
            real.write_text(
                f"#!/usr/bin/env bash\nprintf 'released:%s\\n' \"$1\"\n: > {marker}\n",
                encoding="utf-8",
            )
            real.chmod(0o755)
            env = os.environ.copy()
            env["C3_USAGE_LOG"] = str(usage)
            env["C3_GATE_REAL_TOOL"] = str(real)
            env["C3_SKILL_READ_ROOTS"] = str(root / ".agents" / "skills" / "c3")
            result = subprocess.run(
                ["bash", str(COMMAND_GATE), "probe"],
                text=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE,
                check=False, env=env,
            )
            return result, marker.exists()

    def test_command_gate_allows_only_mounted_c3_instruction_reads_before_route(self) -> None:
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            usage = root / "usage.tsv"
            usage.write_text("", encoding="utf-8")
            marker = root / "real-ran"
            real = root / "real-tool.sh"
            real.write_text(f"#!/usr/bin/env bash\n: > {marker}\n", encoding="utf-8")
            real.chmod(0o755)
            gate_as_sed = root / "sed"
            gate_as_sed.symlink_to(COMMAND_GATE)
            env = os.environ.copy()
            env["C3_USAGE_LOG"] = str(usage)
            env["C3_GATE_REAL_TOOL"] = str(real)
            env["C3_SKILL_READ_ROOTS"] = str(root / ".agents" / "skills" / "c3")

            allowed = subprocess.run(
                [str(gate_as_sed), "-n", "1,20p", ".agents/skills/c3/SKILL.md"],
                text=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE,
                check=False, env=env, cwd=root,
            )
            self.assertEqual(allowed.returncode, 0, allowed.stderr)
            self.assertTrue(marker.exists())

            marker.unlink()
            traversal = subprocess.run(
                [str(gate_as_sed), "-n", "1,20p", ".agents/skills/c3/../../../../secret"],
                text=True,
                stdout=subprocess.PIPE,
                stderr=subprocess.PIPE,
                check=False,
                env=env,
                cwd=root,
            )

            self.assertEqual(traversal.returncode, 78)
            self.assertFalse(marker.exists())

            blocked = subprocess.run(
                [str(gate_as_sed), "-n", "1,20p", "apps/start/src/server.ts"],
                text=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE,
                check=False, env=env,
            )
            self.assertEqual(blocked.returncode, 78)
            self.assertFalse(marker.exists())

    def test_command_gate_does_not_trust_an_agent_controlled_bypass(self) -> None:
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            usage = root / "usage.tsv"
            usage.write_text("", encoding="utf-8")
            marker = root / "real-ran"
            real = root / "real-tool.sh"
            real.write_text(f"#!/usr/bin/env bash\n: > {marker}\n", encoding="utf-8")
            real.chmod(0o755)
            env = os.environ.copy()
            env["C3_USAGE_LOG"] = str(usage)
            env["C3_GATE_REAL_TOOL"] = str(real)
            env["C3_GATE_BYPASS"] = "1"

            result = subprocess.run(
                ["bash", str(COMMAND_GATE), "probe"],
                text=True,
                stdout=subprocess.PIPE,
                stderr=subprocess.PIPE,
                check=False,
                env=env,
            )

            self.assertEqual(result.returncode, 78)
            self.assertFalse(marker.exists())

    def test_claude_arms_expose_the_same_native_tools(self) -> None:
        source = RUN_BLINDBOX.read_text(encoding="utf-8")

        self.assertIn('claude_allowed_tools="Bash,Read,Write,Edit,MultiEdit,Glob,Grep,LS"', source)
        self.assertIn("--system-prompt", source)
        self.assertIn("isolated repository analysis agent", source)
        self.assertIn("claude_base_system_prompt_sha256", source)
        self.assertNotIn('claude_allowed_tools="Bash"\n', source)
        self.assertIn('--allowedTools "$claude_allowed_tools"', source)

    def test_codex_arms_replace_default_model_instructions_identically(self) -> None:
        source = RUN_BLINDBOX.read_text(encoding="utf-8")

        self.assertIn("minimal-codex-model-instructions.md", source)
        self.assertIn("codex_model_instructions_sha256", source)
        self.assertIn('model_instructions_file="/opt/codex-model-instructions.md"', source)
        self.assertIn("tool_output_token_limit=400", source)
        self.assertIn("runtime_guard_tool_result_bytes", source)

    def test_command_gate_requires_successful_c3_categories_in_order(self) -> None:
        cases = (
            ("", "route"),
            ("route\t1\n", "route"),
            ("route\t0\n", "impact"),
            ("route\t0\nimpact\t7\n", "impact"),
            ("route\t0\nimpact\t0\n", "evidence"),
            ("route\t0\nimpact\t0\nevidence\t2\n", "evidence"),
        )
        for usage_rows, required_category in cases:
            with self.subTest(usage_rows=usage_rows):
                result, real_ran = self.run_command_gate(usage_rows)
                self.assertEqual(result.returncode, 78)
                self.assertFalse(real_ran)
                self.assertIn(f"required next: {required_category}", result.stderr)

    def test_command_gate_releases_the_original_tool_after_successful_uptake(self) -> None:
        result, real_ran = self.run_command_gate(
            "route\t0\nimpact\t0\nevidence\t0\n"
        )

        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertTrue(real_ran)
        self.assertEqual(result.stdout, "released:probe\n")

    def make_seed_repo(self, root: Path) -> Path:
        repo = root / "seed"
        repo.mkdir()
        subprocess.run(["git", "init", "-q", str(repo)], check=True)
        subprocess.run(["git", "-C", str(repo), "config", "user.email", "eval@example.invalid"], check=True)
        subprocess.run(["git", "-C", str(repo), "config", "user.name", "Eval"], check=True)
        (repo / "AGENTS.md").write_text("ambient c3 instruction\n", encoding="utf-8")
        (repo / "CLAUDE.md").write_text("ambient claude instruction\n", encoding="utf-8")
        nested = repo / "packages" / "feature"
        nested.mkdir(parents=True)
        (nested / "AGENTS.md").write_text("nested ambient agent instruction\n", encoding="utf-8")
        (nested / "CLAUDE.md").write_text("nested ambient claude instruction\n", encoding="utf-8")
        for dirname in (".agents", ".claude", ".codex"):
            directory = nested / dirname
            directory.mkdir()
            (directory / "ambient.txt").write_text("ambient configuration\n", encoding="utf-8")
        (repo / "README.md").write_text("seed\n", encoding="utf-8")
        (repo / ".c3").mkdir()
        (repo / ".c3" / "README.md").write_text("c3 facts\n", encoding="utf-8")
        subprocess.run(["git", "-C", str(repo), "add", "."], check=True)
        subprocess.run(["git", "-C", str(repo), "commit", "-qm", "seed"], check=True)
        return repo

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
        self.assertIn("Use the neutral repository quickstart", result.stdout)
        self.assertNotIn("Do not rely\non ambient repository files", result.stdout)

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
        self.assertIn(
            "c3_uptake_gate=supervisor_transcript_exact_first_command",
            result.stdout,
        )
        self.assertIn("path_has_usr_local_bin=no", result.stdout)
        self.assertIn("instruction_policy=sanitized_then_neutral_baseline", result.stdout)
        self.assertIn("ambient_instruction_file_count=0", result.stdout)
        self.assertIn("ambient_agent_dir_count=0", result.stdout)
        self.assertNotIn("/opt/c3/guard/bin", result.stdout)
        self.assertIn(
            "--setenv C3_USAGE_LOG /work/project/.c3-eval/unit-dryrun.c3-usage.tsv",
            result.stdout,
        )
        self.assertNotIn("--setenv C3_USAGE_LOG /runs/", result.stdout)
        self.assertNotIn("/usr/local/bin", result.stdout)

        sha_lines = [
            line for line in result.stdout.splitlines()
            if line.split("=", 1)[0].endswith("_sha256") or line.startswith("source_sha256[")
        ]
        self.assertGreaterEqual(len(sha_lines), 5)
        for line in sha_lines:
            self.assertRegex(line, r"=[0-9a-f]{64}$")

    @unittest.skipUnless(shutil.which("tmux"), "tmux is not installed")
    def test_run_blindbox_tmux_backend_preserves_treatment_metadata(self) -> None:
        result = subprocess.run(
            [
                str(RUN_BLINDBOX),
                "--agent",
                "codex",
                "--topic",
                TOPIC,
                "--label",
                f"unit-tmux-{os.getpid()}",
                "--backend",
                "tmux",
                "--dry-run",
            ],
            text=True,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            check=False,
        )

        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertIn("execution_backend=tmux", result.stdout)
        self.assertIn("c3_uptake_gate=supervisor_transcript_exact_first_comm", result.stdout)

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

    def test_run_blindbox_rejects_wrong_skill_tree_hash(self) -> None:
        env = os.environ.copy()
        env["C3_EXPECT_SKILL_TREE_SHA256"] = "0" * 64
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
        self.assertIn("Hash mismatch for skills/c3 tree", result.stderr)

    def test_run_blindbox_rejects_wrong_seed_head_hash(self) -> None:
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            repo = self.make_seed_repo(root)
            prompt = root / "prompt.md"
            prompt.write_text("Assess the change.\n", encoding="utf-8")
            env = os.environ.copy()
            env["C3_EXPECT_SEED_HEAD_SHA256"] = "0" * 64
            result = subprocess.run(
                [
                    str(RUN_BLINDBOX),
                    "--agent",
                    "codex",
                    "--prompt-file",
                    str(prompt),
                    "--seed-repo",
                    str(repo),
                    "--condition",
                    "without_c3",
                    "--run-dir",
                    str(root / "runs"),
                    "--label",
                    "control",
                    "--dry-run",
                ],
                text=True,
                stdout=subprocess.PIPE,
                stderr=subprocess.PIPE,
                check=False,
                env=env,
            )

        self.assertNotEqual(result.returncode, 0)
        self.assertIn("Hash mismatch for seed_repo_head", result.stderr)

    def assert_neutral_baseline(self, workspace: Path, output: str) -> str:
        agents = workspace / "AGENTS.md"
        claude = workspace / "CLAUDE.md"
        self.assertTrue(agents.exists())
        self.assertTrue(claude.exists())
        self.assertEqual(agents.read_bytes(), claude.read_bytes())
        baseline = agents.read_text(encoding="utf-8")
        self.assertIn("# Repository Agent Quickstart", baseline)
        self.assertIn("Explain the behavior and data flow", baseline)
        self.assertIn("Keep each command output below 40 lines", baseline)
        self.assertIn("Never print a whole large source file", baseline)
        self.assertIn("Use at most five discovery tool calls", baseline)
        self.assertNotIn("C3", baseline)
        self.assertIn("instruction_policy=sanitized_then_neutral_baseline", output)
        self.assertIn("baseline_instruction_file_count=2", output)
        self.assertIn("unexpected_instruction_file_count=0", output)
        self.assertIn("baseline_instruction_hash_match=1", output)
        self.assertRegex(output, r"baseline_instruction_sha256=[0-9a-f]{64}")
        return baseline

    def test_generic_control_dry_run_injects_neutral_baseline_after_sanitizing(self) -> None:
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            repo = self.make_seed_repo(root)
            prompt = root / "prompt.md"
            prompt.write_text("Assess the change.\n", encoding="utf-8")
            run_dir = root / "runs"
            result = subprocess.run(
                [
                    str(RUN_BLINDBOX),
                    "--agent",
                    "codex",
                    "--prompt-file",
                    str(prompt),
                    "--seed-repo",
                    str(repo),
                    "--condition",
                    "without_c3",
                    "--run-dir",
                    str(run_dir),
                    "--label",
                    "control",
                    "--dry-run",
                ],
                text=True,
                stdout=subprocess.PIPE,
                stderr=subprocess.PIPE,
                check=False,
            )

            workspace = run_dir / "control.workspace"
            self.assertEqual(result.returncode, 0, result.stderr)
            self.assertIn("condition=without_c3", result.stdout)
            self.assertIn("local_c3=unmounted", result.stdout)
            self.assertFalse((workspace / ".c3").exists())
            baseline = self.assert_neutral_baseline(workspace, result.stdout)
            self.assertNotIn("## C3 Impact Treatment", baseline)
            self.assertIn("treatment_instruction_policy=none", result.stdout)
            self.assertIn("treatment_instruction_sha256=unmounted", result.stdout)
            self.assertFalse((workspace / "packages" / "feature" / "AGENTS.md").exists())
            self.assertFalse((workspace / "packages" / "feature" / "CLAUDE.md").exists())
            self.assertFalse((workspace / "packages" / "feature" / ".agents").exists())
            self.assertFalse((workspace / "packages" / "feature" / ".claude").exists())
            self.assertFalse((workspace / "packages" / "feature" / ".codex").exists())
            self.assertTrue((workspace / "README.md").exists())
            self.assertNotIn("/opt/c3/skills/c3 --", result.stdout)
            self.assertNotIn("/opt/c3/guard/bin", result.stdout)
            self.assertIn("c3_uptake_gate=unmounted", result.stdout)
            self.assertEqual(list((run_dir / ".agent-output-control").iterdir()), [])

    def test_generic_treatment_dry_run_keeps_c3_and_mounts_local_skill(self) -> None:
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            repo = self.make_seed_repo(root)
            prompt = root / "prompt.md"
            prompt.write_text("Use the mounted local C3 skill, then assess the change.\n", encoding="utf-8")
            run_dir = root / "runs"
            result = subprocess.run(
                [
                    str(RUN_BLINDBOX),
                    "--agent",
                    "codex",
                    "--prompt-file",
                    str(prompt),
                    "--seed-repo",
                    str(repo),
                    "--condition",
                    "with_c3",
                    "--run-dir",
                    str(run_dir),
                    "--label",
                    "treatment",
                    "--dry-run",
                ],
                text=True,
                stdout=subprocess.PIPE,
                stderr=subprocess.PIPE,
                check=False,
            )

            workspace = run_dir / "treatment.workspace"
            self.assertEqual(result.returncode, 0, result.stderr)
            self.assertIn("condition=with_c3", result.stdout)
            self.assertIn("local_c3=/opt/c3/skills/c3/bin/c3x.sh", result.stdout)
            self.assertTrue((workspace / ".c3" / "README.md").exists())
            treatment_instructions = (workspace / "AGENTS.md").read_text(encoding="utf-8")
            self.assertEqual(
                (workspace / "AGENTS.md").read_bytes(),
                (workspace / "CLAUDE.md").read_bytes(),
            )
            self.assertIn("# Repository Agent Quickstart", treatment_instructions)
            self.assertNotIn("## C3 Impact Treatment", treatment_instructions)
            self.assertIn("treatment_instruction_policy=forced_c3_impact", result.stdout)
            self.assertRegex(result.stdout, r"treatment_instruction_sha256=[0-9a-f]{64}")
            self.assertIn("treatment_runtime_layer=codex_developer_instructions", result.stdout)
            self.assertIn("developer_instructions=", result.stdout)
            self.assertFalse((workspace / "packages" / "feature" / "AGENTS.md").exists())
            self.assertFalse((workspace / "packages" / "feature" / "CLAUDE.md").exists())
            self.assertFalse((workspace / "packages" / "feature" / ".agents").exists())
            self.assertFalse((workspace / "packages" / "feature" / ".claude").exists())
            self.assertFalse((workspace / "packages" / "feature" / ".codex").exists())
            self.assertIn("/opt/c3/skills/c3", result.stdout)
            self.assertIn("/work/project/.agents/skills/c3", result.stdout)
            self.assertIn("/work/project/.claude/skills/c3", result.stdout)
            self.assertIn("c3-usage-proxy.sh", result.stdout)
            self.assertIn(
                "c3_uptake_gate=supervisor_transcript_exact_first_command",
                result.stdout,
            )
            self.assertIn("--json", result.stdout)

    def test_claude_treatment_appends_the_forced_block_to_its_system_prompt(self) -> None:
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            repo = self.make_seed_repo(root)
            prompt = root / "prompt.md"
            prompt.write_text("Assess the change.\n", encoding="utf-8")
            result = subprocess.run(
                [
                    str(RUN_BLINDBOX), "--agent", "claude", "--prompt-file", str(prompt),
                    "--seed-repo", str(repo), "--condition", "with_c3",
                    "--run-dir", str(root / "runs"), "--label", "claude-treatment",
                    "--dry-run",
                ],
                text=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE, check=False,
            )

        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertIn("treatment_runtime_layer=claude_append_system_prompt", result.stdout)
        self.assertIn("--append-system-prompt", result.stdout)
        self.assertIn("## C3 Impact Treatment", result.stdout)

    def test_reasoning_effort_is_passed_to_codex_and_claude(self) -> None:
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            repo = self.make_seed_repo(root)
            prompt = root / "prompt.md"
            prompt.write_text("Assess the change.\n", encoding="utf-8")

            codex = subprocess.run(
                [
                    str(RUN_BLINDBOX), "--agent", "codex", "--prompt-file", str(prompt),
                    "--seed-repo", str(repo), "--condition", "without_c3",
                    "--run-dir", str(root / "codex"), "--label", "codex-effort",
                    "--model", "gpt-test", "--effort", "low", "--dry-run",
                ],
                text=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE, check=False,
            )
            claude = subprocess.run(
                [
                    str(RUN_BLINDBOX), "--agent", "claude", "--prompt-file", str(prompt),
                    "--seed-repo", str(repo), "--condition", "without_c3",
                    "--run-dir", str(root / "claude"), "--label", "claude-effort",
                    "--model", "opus", "--effort", "medium", "--dry-run",
                    "--max-budget-usd", "0.75",
                ],
                text=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE, check=False,
            )

        self.assertEqual(codex.returncode, 0, codex.stderr)
        self.assertIn("reasoning_effort=low", codex.stdout)
        self.assertRegex(codex.stdout, r"model_reasoning_effort[^\n]*low")
        self.assertIn("runtime_guard=process_supervisor", codex.stdout)
        self.assertRegex(codex.stdout, r"supervise-agent\.py[^\n]*--max-tool-calls 6")
        self.assertRegex(codex.stdout, r"--max-output-bytes 524288")
        self.assertEqual(claude.returncode, 0, claude.stderr)
        self.assertIn("reasoning_effort=medium", claude.stdout)
        self.assertRegex(claude.stdout, r"--effort medium")
        self.assertRegex(claude.stdout, r"--output-format stream-json")
        self.assertIn("runtime_guard=process_supervisor", claude.stdout)
        self.assertRegex(claude.stdout, r"supervise-agent\.py[^\n]*--max-tool-calls 6")
        self.assertRegex(claude.stdout, r"--max-budget-usd 0.75")

    def test_rejects_unknown_reasoning_effort(self) -> None:
        result = subprocess.run(
            [str(RUN_BLINDBOX), "--agent", "codex", "--topic", TOPIC, "--effort", "huge", "--dry-run"],
            text=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE, check=False,
        )

        self.assertNotEqual(result.returncode, 0)
        self.assertIn("Unsupported effort", result.stderr)


if __name__ == "__main__":
    unittest.main()
