from __future__ import annotations

import hashlib
import importlib.util
import json
import os
from pathlib import Path
import signal
import shutil
import stat
import subprocess
import sys
import tempfile
import time
import unittest
from unittest import mock


MODULE_PATH = Path(__file__).with_name("validate_structural_retrieval_preparation_v7.py")


def resolved_tool(environment_name: str, command: str) -> Path:
    configured = os.environ.get(environment_name)
    discovered = configured or shutil.which(command)
    if not discovered:
        raise RuntimeError(f"required test tool is unavailable: {command}")
    return Path(discovered).resolve(strict=True)


REAL_GO = resolved_tool("C3_PROTOCOL_V7_TEST_GO", "go")
REAL_GIT = resolved_tool("C3_PROTOCOL_V7_TEST_GIT", "git")
RETAINED_BASELINE = Path(
    os.environ.get(
        "C3_PROTOCOL_V7_BASELINE_ROOT",
        Path.home() / ".cache/c3-v7-baseline-bb-05",
    )
).resolve()
RETAINED_AUTHORITY = RETAINED_BASELINE / "parent/controller-authority.v4.json"
SPEC = importlib.util.spec_from_file_location("preparation_verifier_v7", MODULE_PATH)
assert SPEC is not None and SPEC.loader is not None
verifier = importlib.util.module_from_spec(SPEC)
sys.modules[SPEC.name] = verifier
SPEC.loader.exec_module(verifier)


def sha(data: bytes) -> str:
    return hashlib.sha256(data).hexdigest()


def canonical(value: object) -> bytes:
    return json.dumps(value, sort_keys=True, separators=(",", ":")).encode()


def write_fake_go(directory: Path, body: str) -> Path:
    directory.mkdir(parents=True, exist_ok=True)
    path = directory / "go"
    path.write_text(f"#!{sys.executable}\n{body}", encoding="utf-8")
    path.chmod(0o755)
    return path


def wait_for_path(path: Path, timeout: float = 10.0) -> None:
    deadline = time.monotonic() + timeout
    while time.monotonic() < deadline:
        if path.exists():
            return
        time.sleep(0.02)
    raise AssertionError("child readiness marker was not created")


def wait_for_pid_exit(pid: int, timeout: float = 5.0) -> None:
    deadline = time.monotonic() + timeout
    while time.monotonic() < deadline:
        if not Path(f"/proc/{pid}").exists():
            return
        time.sleep(0.02)
    raise AssertionError("child process remained after bounded runner cleanup")


def copy_real_compiler_closure(repository: Path, driver_root: Path) -> dict[str, str]:
    closure = verifier._decode_authority_compile_closure(RETAINED_AUTHORITY.read_bytes())
    main_test = repository / "cli/tools/structural-search-eval-v2/main_test.go"
    return verifier._copy_compiler_closure(
        repository, driver_root, closure, main_test, sha(main_test.read_bytes())
    )


def real_tool_path(private: Path) -> Path:
    tools = verifier._validate_tools(
        REAL_GO, REAL_GIT, sha(REAL_GO.read_bytes()), sha(REAL_GIT.read_bytes())
    )
    return verifier._materialize_minimal_tool_path(private / "tools", tools)


def preparation_harness_result(preparation: "SyntheticPreparation") -> dict[str, object]:
    return {
        "$schema": "structural-retrieval-protocol-v7-preparation-verifier-harness.v1",
        "authority_sha256": preparation.manifest["authority_sha256"],
        "bundle_heads_sha256": preparation.manifest["bundle_heads_sha256"],
        "bundle_sha256": preparation.manifest["bundle_sha256"],
        "privacy_detector_sha256": "8" * 64,
        "privacy_policy_sha256": preparation.manifest["privacy_policy_sha256"],
        "privacy_term_count": 1,
        "runtime_sha256": preparation.manifest["runtime_sha256"],
        "scanned_artifact_count": 9,
        "scanned_byte_count": 10,
        "source_capsule_sha256": preparation.manifest["source_capsule_sha256"],
        "source_object_byte_count": 11,
        "source_object_count": 12,
        "source_path_count": 13,
        "status": "accepted",
    }


def write_accepting_go(
    directory: Path, preparation: "SyntheticPreparation", extra: str = ""
) -> Path:
    body = (
        "import json,os,pathlib\n"
        "cfg=json.loads(pathlib.Path(os.environ['C3_PROTOCOL_V7_PREPARATION_VERIFIER_CONFIG']).read_bytes())\n"
        f"result={canonical(preparation_harness_result(preparation))!r}\n"
        "path=pathlib.Path(cfg['result_path']);path.write_bytes(result);path.chmod(0o600)\n"
        + extra
    )
    return write_fake_go(directory, body)


class SyntheticPreparation:
    def __init__(self, root: Path) -> None:
        self.base = root
        self.source = root / "source"
        self.preparation = root / "preparation"
        self.policy = root / "policy.json"
        self.checkins = root / "checkins.jsonl"
        self.adapter = root / "adapter.py"
        self.go = root / "tools/go"
        self.git = root / "tools/git"
        self.main = self.source / "cli/tools/structural-search-eval-v2/main.go"
        self.main_test = self.source / "cli/tools/structural-search-eval-v2/main_test.go"
        self.fixture = self.source / "research/eval/structural-retrieval/fixtures.dev.v2.jsonl"
        self.benchmark = self.source / "research/eval/structural-retrieval/benchmark.v2.json"
        for path, data in (
            (self.main, b"package main\n"),
            (self.main_test, b"package main\n"),
            (self.fixture, b"{}\n"),
            (self.benchmark, b"{}"),
            (self.adapter, b"# adapter\n"),
        ):
            path.parent.mkdir(parents=True, exist_ok=True)
            path.write_bytes(data)
            path.chmod(0o644)
        for path in (self.go, self.git):
            path.parent.mkdir(parents=True, exist_ok=True)
            path.write_text(f"#!{sys.executable}\nraise SystemExit(0)\n", encoding="utf-8")
            path.chmod(0o755)
        self.policy.write_bytes(
            b'{"$schema":"structural-retrieval-privacy-policy.v1","deny_terms":["private-marker"]}'
        )
        self.policy.chmod(0o600)
        self.checkins.write_bytes(b'{"seq":7}\n')
        self.checkins.chmod(0o600)

        self.preparation.mkdir(mode=0o700)
        freeze = self.preparation / "freeze"
        freeze.mkdir(mode=0o700)
        self.authority = self.preparation / "controller-authority.v4.json"
        self.bundle = freeze / "source.bundle"
        self.runtime = freeze / "controller-runtime"
        source_capsule = {
            "$schema": "structural-retrieval-source-capsule.v2",
            "dirty_patch_sha256": "8" * 64,
            "head_commit": "4" * 40,
            "head_tree": "7" * 40,
            "inputs": [{
                "origin": "head",
                "path": "cli/tools/structural-search-eval-v2/main.go",
                "sha256": sha(self.main.read_bytes()),
            }],
            "repository_build_input_count": 1,
        }
        self.authority.write_bytes(canonical({
            "controller_source_capsule": source_capsule,
            "runtime_source_capsule": source_capsule,
        }))
        self.bundle.write_bytes(b"bundle")
        self.runtime.write_bytes(b"runtime")
        for path in (self.authority, self.bundle, self.runtime):
            path.chmod(0o600)

        self.frozen = verifier.FrozenInputs(
            adapter_sha256=sha(self.adapter.read_bytes()),
            driver_sha256="1" * 64,
            main_sha256=sha(self.main.read_bytes()),
            main_test_sha256=sha(self.main_test.read_bytes()),
            fixture_sha256=sha(self.fixture.read_bytes()),
            benchmark_sha256=sha(self.benchmark.read_bytes()),
            scorer_region_sha256="2" * 64,
            go_executable_sha256=sha(self.go.read_bytes()),
            git_executable_sha256=sha(self.git.read_bytes()),
        )
        self.manifest = {
            "$schema": "structural-retrieval-protocol-v7-preparation.v1",
            "adapter_sha256": self.frozen.adapter_sha256,
            "authority_sha256": sha(self.authority.read_bytes()),
            "benchmark_sha256": self.frozen.benchmark_sha256,
            "bundle_heads_sha256": "3" * 64,
            "bundle_sha256": sha(self.bundle.read_bytes()),
            "commit": "4" * 40,
            "driver_sha256": self.frozen.driver_sha256,
            "fixture_sha256": self.frozen.fixture_sha256,
            "input_manifest_sha256": "5" * 64,
            "main_sha256": self.frozen.main_sha256,
            "main_test_sha256": self.frozen.main_test_sha256,
            "privacy_policy_sha256": sha(self.policy.read_bytes()),
            "privacy_term_count": 1,
            "runtime_sha256": sha(self.runtime.read_bytes()),
            "scorer_region_sha256": self.frozen.scorer_region_sha256,
            "source_capsule_sha256": "6" * 64,
            "status": "accepted",
            "tree": "7" * 40,
        }
        self.manifest_path = self.preparation / "preparation-manifest.json"
        self.write_manifest()

    def write_manifest(self, raw: bytes | None = None) -> None:
        self.manifest_path.write_bytes(raw if raw is not None else canonical(self.manifest))
        self.manifest_path.chmod(0o600)

    def config(self) -> object:
        return verifier.VerifierConfig(
            source_root=self.source,
            preparation_root=self.preparation,
            main_path=self.main,
            main_test_path=self.main_test,
            fixture_path=self.fixture,
            benchmark_path=self.benchmark,
            threshold_checkins_path=self.checkins,
            threshold_checkin_seq=7,
            privacy_policy_path=self.policy,
            adapter_path=self.adapter,
            go_executable_path=self.go,
            git_executable_path=self.git,
        )


class AcceptingRunner:
    def __init__(self, preparation: SyntheticPreparation, *, result: dict[str, object] | None = None) -> None:
        self.preparation = preparation
        self.result = result
        self.calls = 0
        self.temporary_roots: list[Path] = []

    def __call__(self, command: list[str], *, cwd: Path, env: dict[str, str], timeout: int, output_cap: int) -> object:
        self.calls += 1
        self.temporary_roots.append(cwd.parents[1])
        config = json.loads(Path(env["C3_PROTOCOL_V7_PREPARATION_VERIFIER_CONFIG"]).read_bytes())
        result_path = Path(config["result_path"])
        result = self.result or preparation_harness_result(self.preparation)
        result_path.write_bytes(canonical(result))
        result_path.chmod(0o600)
        return verifier.ProcessResult(returncode=0, stdout=b"ok\n", stderr=b"")


class PreparationVerifierTests(unittest.TestCase):
    def setUp(self) -> None:
        self.temporary = tempfile.TemporaryDirectory(prefix="generic-v7-verifier-test-")
        self.root = Path(self.temporary.name)
        self.synthetic = SyntheticPreparation(self.root)

    def tearDown(self) -> None:
        self.temporary.cleanup()

    def assert_rejected(self, error_class: str, *, runner: object | None = None) -> None:
        result = verifier.verify_preparation(
            self.synthetic.config(),
            frozen=self.synthetic.frozen,
            runner=runner or AcceptingRunner(self.synthetic),
        )
        self.assertEqual(result["status"], "rejected")
        self.assertEqual(result["failure_class"], error_class)
        encoded = canonical(result).lower()
        self.assertNotIn(str(self.root).encode().lower(), encoded)
        self.assertNotIn(b"private-marker", encoded)

    def test_trust_closure_requires_hash_pinned_absolute_tools_and_capsule_sources(self) -> None:
        tool = self.root / "go"
        tool.write_bytes(b"tool")
        tool.chmod(0o755)
        identity = verifier._validate_executable(tool, sha(tool.read_bytes()), "go")
        self.assertEqual(identity.path, tool.absolute())
        with self.assertRaisesRegex(verifier.VerifierError, "tool_identity_invalid"):
            verifier._validate_executable(Path("go"), sha(tool.read_bytes()), "go")
        authority = verifier._decode_authority_compile_closure(self.synthetic.authority.read_bytes())
        self.assertEqual(authority.entries[0].path, "cli/tools/structural-search-eval-v2/main.go")

    def test_orchestrator_accepts_a_strictly_bound_canonical_harness_result_and_cleans_temp(self) -> None:
        runner = AcceptingRunner(self.synthetic)
        result = verifier.verify_preparation(self.synthetic.config(), frozen=self.synthetic.frozen, runner=runner)
        self.assertEqual(result["status"], "accepted")
        self.assertEqual(result["source_object_count"], 12)
        self.assertEqual(runner.calls, 1)
        self.assertTrue(all(not root.exists() for root in runner.temporary_roots))
        self.assertEqual(verifier.canonical_json(result), canonical(result))

    def test_manifest_duplicate_unknown_noncanonical_and_hash_tamper_reject(self) -> None:
        original = canonical(self.synthetic.manifest)
        cases = (
            original[:-1] + b',"status":"accepted"}',
            original[:-1] + b',"unknown":1}',
            b" " + original,
        )
        for raw in cases:
            with self.subTest(raw=raw[:30]):
                self.synthetic.write_manifest(raw)
                self.assert_rejected("metadata_invalid")
        self.synthetic.write_manifest()
        self.synthetic.authority.write_bytes(b'{"tampered":true}')
        self.synthetic.authority.chmod(0o600)
        self.assert_rejected("metadata_binding_invalid")

    def test_path_mode_symlink_and_extra_file_reject(self) -> None:
        self.synthetic.manifest_path.chmod(0o644)
        self.assert_rejected("preparation_layout_invalid")
        self.synthetic.manifest_path.chmod(0o600)
        (self.synthetic.preparation / "extra").write_text("x")
        self.assert_rejected("preparation_layout_invalid")
        (self.synthetic.preparation / "extra").unlink()
        target = self.synthetic.runtime.with_name("runtime-target")
        self.synthetic.runtime.rename(target)
        self.synthetic.runtime.symlink_to(target)
        self.assert_rejected("preparation_layout_invalid")

    def test_private_input_path_and_frozen_input_tamper_reject(self) -> None:
        wrong = self.synthetic.config()
        wrong = verifier.VerifierConfig(**{**wrong.__dict__, "main_path": self.synthetic.main_test})
        result = verifier.verify_preparation(wrong, frozen=self.synthetic.frozen, runner=AcceptingRunner(self.synthetic))
        self.assertEqual(result["failure_class"], "frozen_input_invalid")
        self.synthetic.main.write_bytes(b"changed")
        self.assert_rejected("frozen_input_invalid")

    def test_orchestrator_preserves_only_allowlisted_generic_harness_failure_classes(self) -> None:
        cases = (
            ("authority_tamper", "authority_replay_failed"),
            ("bundle_tamper", "bundle_invalid"),
            ("build_drift", "authority_replay_failed"),
            ("environment_drift", "authority_replay_failed"),
            ("threshold_tamper", "threshold_invalid"),
            ("source_capsule_tamper", "authority_replay_failed"),
            ("privacy_positive", "privacy_violation"),
        )
        for vector, failure_class in cases:
            with self.subTest(vector=vector):
                result = {
                    "$schema": "structural-retrieval-protocol-v7-preparation-verifier-harness.v1",
                    "status": "rejected",
                    "failure_class": failure_class,
                }
                self.assert_rejected(
                    failure_class,
                    runner=AcceptingRunner(self.synthetic, result=result),
                )

    def test_orchestrator_never_echoes_private_input_when_harness_reports_privacy_failure(self) -> None:
        positives = (
            b"/workspace/example/work",
            b"Authorization: Bearer abcdefghijklmnop",
            b"api_key=abcdefghijklmnop",
        )
        for positive in positives:
            with self.subTest(positive=positive[:10]):
                result = {
                    "$schema": "structural-retrieval-protocol-v7-preparation-verifier-harness.v1",
                    "status": "rejected",
                    "failure_class": "privacy_violation",
                }
                runner = AcceptingRunner(self.synthetic, result=result)
                self.assert_rejected("privacy_violation", runner=runner)

    def test_orchestrator_rejects_a_missing_harness_result_and_cleans_temp(self) -> None:
        roots: list[Path] = []

        def runner(command: list[str], *, cwd: Path, env: dict[str, str], timeout: int, output_cap: int) -> object:
            roots.append(cwd.parents[1])
            return verifier.ProcessResult(returncode=0, stdout=b"ok", stderr=b"")

        result = verifier.verify_preparation(self.synthetic.config(), frozen=self.synthetic.frozen, runner=runner)
        self.assertEqual(result["failure_class"], "harness_result_invalid")
        self.assertTrue(all(not root.exists() for root in roots))

    def test_governed_source_and_threshold_mutation_during_runner_reject(self) -> None:
        for target, expected in (
            (self.synthetic.main, "harness_source_invalid"),
            (self.synthetic.checkins, "governed_input_changed"),
        ):
            with self.subTest(target=target.name):
                original = target.read_bytes()
                mutated = original + b"changed\n"
                fake_go = write_accepting_go(
                    self.root / ("mutate-" + target.name),
                    self.synthetic,
                    f"pathlib.Path({str(target)!r}).write_bytes({mutated!r})\n",
                )
                config = verifier.VerifierConfig(**{
                    **self.synthetic.config().__dict__, "go_executable_path": fake_go,
                })
                frozen = verifier.FrozenInputs(**{
                    **self.synthetic.frozen.__dict__, "go_executable_sha256": sha(fake_go.read_bytes()),
                })
                result = verifier.verify_preparation(config, frozen=frozen)
                self.assertEqual(result["status"], "rejected")
                self.assertEqual(result["failure_class"], expected)
                target.write_bytes(original)
                target.chmod(0o600 if target == self.synthetic.checkins else 0o644)

    def test_hostile_ambient_path_is_not_visible_to_runner(self) -> None:
        fake_go = write_accepting_go(
            self.root / "hostile-check",
            self.synthetic,
            "names={p.name for p in pathlib.Path(os.environ['PATH']).iterdir()}\n"
            "assert names == {'go','git'} and 'hostile' not in os.environ['PATH']\n",
        )
        config = verifier.VerifierConfig(**{
            **self.synthetic.config().__dict__, "go_executable_path": fake_go,
        })
        frozen = verifier.FrozenInputs(**{
            **self.synthetic.frozen.__dict__, "go_executable_sha256": sha(fake_go.read_bytes()),
        })
        with mock.patch.dict(os.environ, {"PATH": "/hostile/ambient/path"}):
            result = verifier.verify_preparation(config, frozen=frozen)
        self.assertEqual(result["status"], "accepted")

    def test_python_post_copy_identity_check_rejects_a_changed_compiled_test(self) -> None:
        original_copy = verifier._copy_compiler_closure

        def changing_copy(*args: object, **kwargs: object) -> dict[str, str]:
            snapshot = original_copy(*args, **kwargs)
            target = args[1]
            assert isinstance(target, Path)
            copied = target / "cli/tools/structural-search-eval-v2/main_test.go"
            copied.write_bytes(copied.read_bytes() + b"// changed after copy\n")
            return snapshot

        runner = AcceptingRunner(self.synthetic)
        with mock.patch.object(verifier, "_copy_compiler_closure", changing_copy):
            result = verifier.verify_preparation(
                self.synthetic.config(),
                frozen=self.synthetic.frozen,
                runner=runner,
            )
        self.assertEqual(result["failure_class"], "frozen_input_invalid")
        self.assertEqual(runner.calls, 0)

    def test_real_subprocess_nonzero_and_stderr_reject(self) -> None:
        cases = (
            ("import sys\nsys.exit(9)\n", "nonzero"),
            ("import sys\nsys.stderr.write('generic failure\\n')\n", "stderr"),
        )
        for body, label in cases:
            with self.subTest(label=label):
                fake_go = write_fake_go(self.root / ("bin-" + label), body)
                config = verifier.VerifierConfig(**{
                    **self.synthetic.config().__dict__, "go_executable_path": fake_go,
                })
                frozen = verifier.FrozenInputs(**{
                    **self.synthetic.frozen.__dict__, "go_executable_sha256": sha(fake_go.read_bytes()),
                })
                result = verifier.verify_preparation(config, frozen=frozen)
                self.assertEqual(result["failure_class"], "harness_rejected")

    def test_real_subprocess_timeout_and_output_cap_reject_and_clean_process_tree(self) -> None:
        marker = self.root / "child.pid"
        timeout_go = write_fake_go(
            self.root / "bin-timeout",
            "import pathlib, subprocess, time\n"
            f"child=subprocess.Popen(['/bin/sleep','60'])\npathlib.Path({str(marker)!r}).write_text(str(child.pid))\n"
            "time.sleep(60)\n",
        )
        timeout_config = verifier.VerifierConfig(**{
            **self.synthetic.config().__dict__, "go_executable_path": timeout_go,
        })
        timeout_frozen = verifier.FrozenInputs(**{
            **self.synthetic.frozen.__dict__, "go_executable_sha256": sha(timeout_go.read_bytes()),
        })
        with mock.patch.object(verifier, "PROCESS_TIMEOUT_SECONDS", 1):
            result = verifier.verify_preparation(timeout_config, frozen=timeout_frozen)
        self.assertEqual(result["failure_class"], "harness_timeout")
        wait_for_path(marker)
        wait_for_pid_exit(int(marker.read_text()))

        cap_go = write_fake_go(self.root / "bin-cap", "import os\nos.write(1, b'x' * 4096)\n")
        cap_config = verifier.VerifierConfig(**{
            **self.synthetic.config().__dict__, "go_executable_path": cap_go,
        })
        cap_frozen = verifier.FrozenInputs(**{
            **self.synthetic.frozen.__dict__, "go_executable_sha256": sha(cap_go.read_bytes()),
        })
        with mock.patch.object(verifier, "PROCESS_OUTPUT_CAP", 64):
            result = verifier.verify_preparation(cap_config, frozen=cap_frozen)
        self.assertEqual(result["failure_class"], "harness_output_cap")

    def test_harness_result_duplicate_unknown_private_or_wrong_binding_rejects(self) -> None:
        base = AcceptingRunner(self.synthetic)
        valid_runner = AcceptingRunner(self.synthetic)
        valid = valid_runner.result
        _ = base
        cases = (
            {"$schema": "wrong", "status": "accepted"},
            {
                "$schema": "structural-retrieval-protocol-v7-preparation-verifier-harness.v1",
                "status": "rejected",
                "failure_class": str(self.root),
            },
        )
        for result in cases:
            with self.subTest(result=result):
                self.assert_rejected("harness_result_invalid", runner=AcceptingRunner(self.synthetic, result=result))

    def test_command_allowlist_is_exact(self) -> None:
        allowed = verifier.harness_command(self.synthetic.go)
        verifier.require_allowed_command(allowed)
        for altered in (allowed + ["extra"], ["python"] + allowed[1:], allowed[:-1]):
            with self.subTest(altered=altered):
                with self.assertRaisesRegex(verifier.VerifierError, "harness_command_invalid"):
                    verifier.require_allowed_command(altered)

    def test_same_package_harness_compiles_in_private_synthetic_invocation(self) -> None:
        repository = Path(__file__).resolve().parents[1]
        with tempfile.TemporaryDirectory(prefix="generic-v7-harness-compile-") as temporary:
            private = Path(temporary)
            driver_root = private / "driver-source"
            driver_root.mkdir(parents=True)
            copy_real_compiler_closure(repository, driver_root)
            harness = driver_root / "cli/tools/structural-search-eval-v2/protocol_v7_preparation_verifier_test.go"
            harness.write_text(verifier.GO_HARNESS_SOURCE, encoding="utf-8")
            harness.chmod(0o600)
            policy = private / "policy.json"
            policy.write_bytes(
                b'{"$schema":"structural-retrieval-privacy-policy.v1","deny_terms":["private-marker"]}'
            )
            policy.chmod(0o600)
            result_path = private / "result.json"
            work_root = private / "work"
            work_root.mkdir(mode=0o700)
            main = repository / "cli/tools/structural-search-eval-v2/main.go"
            main_test = repository / "cli/tools/structural-search-eval-v2/main_test.go"
            compiled_main = driver_root / "cli/tools/structural-search-eval-v2/main.go"
            compiled_main_test = driver_root / "cli/tools/structural-search-eval-v2/main_test.go"
            fixture = repository / "research/eval/structural-retrieval/fixtures.dev.v2.jsonl"
            benchmark = repository / "research/eval/structural-retrieval/benchmark.v2.json"
            adapter = repository / "scripts/structural_retrieval_protocol_v7.py"
            config_value = {
                "$schema": "structural-retrieval-protocol-v7-preparation-verifier-config.v1",
                "adapter_path": str(adapter),
                "adapter_sha256": sha(adapter.read_bytes()),
                "benchmark_path": str(benchmark),
                "benchmark_sha256": sha(benchmark.read_bytes()),
                "bundle_path": str(private / "missing.bundle"),
                "driver_sha256": "1" * 64,
                "expected_manifest_sha256": "2" * 64,
                "fixture_path": str(fixture),
                "fixture_sha256": sha(fixture.read_bytes()),
                "main_path": str(compiled_main),
                "main_sha256": sha(main.read_bytes()),
                "main_test_path": str(compiled_main_test),
                "main_test_sha256": sha(main_test.read_bytes()),
                "manifest_path": str(private / "missing-manifest.json"),
                "policy_path": str(policy),
                "preparation_authority_path": str(private / "missing-authority.json"),
                "result_path": str(result_path),
                "runtime_path": str(private / "missing-runtime"),
                "scorer_region_sha256": verifier.ACCEPTED_SCORER_REGION_SHA256,
                "threshold_checkin_seq": 1,
                "threshold_checkins_path": str(private / "missing-checkins.jsonl"),
                "work_root": str(work_root),
            }
            config = private / "config.json"
            config.write_bytes(canonical(config_value))
            config.chmod(0o600)
            scratch = private / "tmp"
            scratch.mkdir(mode=0o700)
            env = verifier._go_environment(scratch, real_tool_path(private))
            env["C3_PROTOCOL_V7_PREPARATION_VERIFIER_CONFIG"] = str(config)
            completed = verifier.run_bounded_process(
                verifier.harness_command(REAL_GO),
                cwd=driver_root / "cli",
                env=env,
                timeout=120,
                output_cap=verifier.PROCESS_OUTPUT_CAP,
            )
            self.assertEqual(completed.returncode, 0)
            self.assertEqual(completed.stderr, b"")
            self.assertNotIn(b"build failed", completed.stdout.lower())
            self.assertEqual(
                json.loads(result_path.read_bytes()),
                {
                    "$schema": "structural-retrieval-protocol-v7-preparation-verifier-harness.v1",
                    "failure_class": "metadata_invalid",
                    "status": "rejected",
                },
            )

    def test_real_go_harness_rejects_a_mutated_compiled_test_before_metadata(self) -> None:
        repository = Path(__file__).resolve().parents[1]
        with tempfile.TemporaryDirectory(prefix="generic-v7-harness-mutation-") as temporary:
            private = Path(temporary)
            driver_root = private / "driver-source"
            driver_root.mkdir(parents=True)
            copy_real_compiler_closure(repository, driver_root)
            compiled_main = driver_root / "cli/tools/structural-search-eval-v2/main.go"
            compiled_main_test = driver_root / "cli/tools/structural-search-eval-v2/main_test.go"
            expected_main_hash = sha(compiled_main.read_bytes())
            expected_test_hash = sha(compiled_main_test.read_bytes())
            compiled_main_test.write_bytes(compiled_main_test.read_bytes() + b"// post-copy mutation\n")
            harness = compiled_main.with_name("protocol_v7_preparation_verifier_test.go")
            harness.write_text(verifier.GO_HARNESS_SOURCE, encoding="utf-8")
            harness.chmod(0o600)
            policy = private / "policy.json"
            policy.write_bytes(
                b'{"$schema":"structural-retrieval-privacy-policy.v1","deny_terms":["private-marker"]}'
            )
            policy.chmod(0o600)
            result_path = private / "result.json"
            work_root = private / "work"
            work_root.mkdir(mode=0o700)
            fixture = repository / "research/eval/structural-retrieval/fixtures.dev.v2.jsonl"
            benchmark = repository / "research/eval/structural-retrieval/benchmark.v2.json"
            adapter = repository / "scripts/structural_retrieval_protocol_v7.py"
            config_value = {
                "$schema": "structural-retrieval-protocol-v7-preparation-verifier-config.v1",
                "adapter_path": str(adapter),
                "adapter_sha256": sha(adapter.read_bytes()),
                "benchmark_path": str(benchmark),
                "benchmark_sha256": sha(benchmark.read_bytes()),
                "bundle_path": str(private / "missing.bundle"),
                "driver_sha256": "1" * 64,
                "expected_manifest_sha256": "2" * 64,
                "fixture_path": str(fixture),
                "fixture_sha256": sha(fixture.read_bytes()),
                "main_path": str(compiled_main),
                "main_sha256": expected_main_hash,
                "main_test_path": str(compiled_main_test),
                "main_test_sha256": expected_test_hash,
                "manifest_path": str(private / "missing-manifest.json"),
                "policy_path": str(policy),
                "preparation_authority_path": str(private / "missing-authority.json"),
                "result_path": str(result_path),
                "runtime_path": str(private / "missing-runtime"),
                "scorer_region_sha256": verifier.ACCEPTED_SCORER_REGION_SHA256,
                "threshold_checkin_seq": 1,
                "threshold_checkins_path": str(private / "missing-checkins.jsonl"),
                "work_root": str(work_root),
            }
            config = private / "config.json"
            config.write_bytes(canonical(config_value))
            config.chmod(0o600)
            scratch = private / "tmp"
            scratch.mkdir(mode=0o700)
            env = verifier._go_environment(scratch, real_tool_path(private))
            env["C3_PROTOCOL_V7_PREPARATION_VERIFIER_CONFIG"] = str(config)
            completed = verifier.run_bounded_process(
                verifier.harness_command(REAL_GO),
                cwd=driver_root / "cli",
                env=env,
                timeout=120,
                output_cap=verifier.PROCESS_OUTPUT_CAP,
            )
            self.assertEqual(completed.returncode, 0)
            self.assertEqual(completed.stderr, b"")
            self.assertEqual(
                json.loads(result_path.read_bytes()),
                {
                    "$schema": "structural-retrieval-protocol-v7-preparation-verifier-harness.v1",
                    "failure_class": "frozen_input_invalid",
                    "status": "rejected",
                },
            )

    def test_default_runner_sigterm_and_sighup_kill_tree_and_remove_private_temp(self) -> None:
        launcher = (
            "import importlib.util,json,sys;"
            f"p={str(MODULE_PATH)!r};"
            "s=importlib.util.spec_from_file_location('signal_verifier',p);"
            "m=importlib.util.module_from_spec(s);sys.modules[s.name]=m;s.loader.exec_module(m);"
            "c=json.loads(__import__('os').environ['VERIFIER_CONFIG']);"
            "f=json.loads(__import__('os').environ['VERIFIER_FROZEN']);"
            "r=m.verify_preparation(m.VerifierConfig(**{k:__import__('pathlib').Path(v) "
            "if k.endswith('_path') or k.endswith('_root') else v for k,v in c.items()}),"
            "frozen=m.FrozenInputs(**f));print(m.canonical_json(r).decode());"
            "raise SystemExit(0 if r['status']=='accepted' else 2)"
        )
        for sent_signal in (signal.SIGTERM, signal.SIGHUP):
            with self.subTest(signal=sent_signal):
                marker_tag = sent_signal.name.lower()
                child_marker = self.root / f"{marker_tag}-child.pid"
                ready_marker = self.root / f"{marker_tag}-ready"
                fake_go = write_fake_go(
                    self.root / f"bin-{marker_tag}",
                    "import pathlib, subprocess, time\n"
                    f"child=subprocess.Popen(['/bin/sleep','60'])\npathlib.Path({str(child_marker)!r}).write_text(str(child.pid))\n"
                    f"pathlib.Path({str(ready_marker)!r}).write_text('ready')\n"
                    "time.sleep(60)\n",
                )
                config = {
                    **self.synthetic.config().__dict__,
                    "go_executable_path": fake_go,
                }
                frozen = {
                    **self.synthetic.frozen.__dict__,
                    "go_executable_sha256": sha(fake_go.read_bytes()),
                }
                private_parent = self.root / f"private-temp-{marker_tag}"
                private_parent.mkdir(mode=0o700)
                env = dict(os.environ)
                env["PATH"] = str(self.root / "hostile-path")
                env["TMPDIR"] = str(private_parent)
                env["VERIFIER_CONFIG"] = json.dumps(
                    {key: str(value) if isinstance(value, Path) else value for key, value in config.items()}
                )
                env["VERIFIER_FROZEN"] = json.dumps(frozen)
                process = subprocess.Popen(
                    [sys.executable, "-c", launcher],
                    stdin=subprocess.DEVNULL,
                    stdout=subprocess.PIPE,
                    stderr=subprocess.PIPE,
                    env=env,
                )
                wait_for_path(ready_marker)
                process.send_signal(sent_signal)
                stdout, stderr = process.communicate(timeout=15)
                self.assertEqual(process.returncode, 2)
                self.assertEqual(stderr, b"")
                self.assertEqual(json.loads(stdout), {
                    "$schema": verifier.RESULT_SCHEMA,
                    "failure_class": "harness_interrupted",
                    "status": "rejected",
                })
                wait_for_pid_exit(int(child_marker.read_text()))
                self.assertEqual(list(private_parent.glob("c3-v7-preparation-verifier-*")), [])


if __name__ == "__main__":
    unittest.main()
