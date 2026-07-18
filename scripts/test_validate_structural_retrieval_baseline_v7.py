from __future__ import annotations

import hashlib
import importlib.util
import json
import os
from pathlib import Path
import shutil
import signal
import subprocess
import sys
import tempfile
import time
import unittest
from unittest import mock


def resolved_tool(environment_name: str, command: str) -> Path:
    configured = os.environ.get(environment_name)
    discovered = configured or shutil.which(command)
    if not discovered:
        raise RuntimeError(f"required test tool is unavailable: {command}")
    return Path(discovered).resolve(strict=True)


REAL_GO = resolved_tool("C3_PROTOCOL_V7_TEST_GO", "go")
REAL_GIT = resolved_tool("C3_PROTOCOL_V7_TEST_GIT", "git")


MODULE_PATH = Path(__file__).with_name("validate_structural_retrieval_baseline_v7.py")
SPEC = importlib.util.spec_from_file_location("baseline_verifier_v7", MODULE_PATH)
assert SPEC is not None and SPEC.loader is not None
verifier = importlib.util.module_from_spec(SPEC)
sys.modules[SPEC.name] = verifier
SPEC.loader.exec_module(verifier)


def canonical(value: object) -> bytes:
    return json.dumps(value, sort_keys=True, separators=(",", ":")).encode()


def sha(data: bytes) -> str:
    return hashlib.sha256(data).hexdigest()


def write_mode(path: Path, data: bytes, mode: int = 0o600) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_bytes(data)
    path.chmod(mode)


def write_fake_tool(directory: Path, name: str, body: str) -> Path:
    path = directory / name
    write_mode(path, f"#!{sys.executable}\n{body}".encode(), 0o755)
    return path


def wait_for_pid_exit(pid: int, timeout: float = 5.0) -> None:
    deadline = time.monotonic() + timeout
    while time.monotonic() < deadline:
        if not Path(f"/proc/{pid}").exists():
            return
        time.sleep(0.02)
    raise AssertionError("child process remained after bounded runner cleanup")


def wait_for_path(path: Path, timeout: float = 10.0) -> None:
    deadline = time.monotonic() + timeout
    while time.monotonic() < deadline:
        if path.exists():
            return
        time.sleep(0.02)
    raise AssertionError("readiness marker was not created")


def with_go(
    baseline: "SyntheticBaseline", go: Path
) -> tuple[object, object]:
    config = verifier.BaselineVerifierConfig(**{
        **baseline.config().__dict__, "go_executable_path": go,
    })
    frozen = verifier.FrozenInputs(**{
        **baseline.frozen.__dict__, "go_executable_sha256": sha(go.read_bytes()),
    })
    return config, frozen


def write_accepting_go(
    directory: Path, baseline: "SyntheticBaseline", extra: str = ""
) -> Path:
    harness_result = {
        "$schema": verifier.HARNESS_SCHEMA,
        "baseline_acceptance": baseline.harness_acceptance(),
        "status": "accepted",
    }
    body = (
        "import json,os,pathlib\n"
        "cfg=json.loads(pathlib.Path(os.environ['C3_PROTOCOL_V7_BASELINE_VERIFIER_CONFIG']).read_bytes())\n"
        f"result={canonical(harness_result)!r}\n"
        "path=pathlib.Path(cfg['result_path']);path.write_bytes(result);path.chmod(0o600)\n"
        + extra
    )
    return write_fake_tool(directory, "go", body)


class SyntheticBaseline:
    def __init__(self, root: Path) -> None:
        self.base = root
        self.source = root / "source"
        self.baseline = root / "baseline"
        self.policy = root / "policy.json"
        self.checkins = root / "checkins.jsonl"
        self.adapter = root / "adapter.py"
        self.preparation_verifier = root / "preparation-verifier.py"
        self.go = root / "tools/go"
        self.git = root / "tools/git"
        self.activation = root / "capture.activation.json"
        self.receipt = Path(str(self.activation) + ".consumed")
        self.authorization = root / "capture.authorization.jsonl"
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
            (self.preparation_verifier, b"# verifier\n"),
        ):
            write_mode(path, data, 0o644)
        write_fake_tool(self.go.parent, "go", "raise SystemExit(0)\n")
        write_fake_tool(self.git.parent, "git", "raise SystemExit(0)\n")
        write_mode(
            self.policy,
            b'{"$schema":"structural-retrieval-privacy-policy.v1","deny_terms":["private-marker"]}',
        )
        write_mode(self.checkins, b'{"seq":7}\n')

        self.frozen = verifier.FrozenInputs(
            adapter_sha256=sha(self.adapter.read_bytes()),
            driver_sha256="1" * 64,
            main_sha256=sha(self.main.read_bytes()),
            main_test_sha256=sha(self.main_test.read_bytes()),
            fixture_sha256=sha(self.fixture.read_bytes()),
            benchmark_sha256=sha(self.benchmark.read_bytes()),
            scorer_region_sha256="2" * 64,
            preparation_verifier_sha256=sha(self.preparation_verifier.read_bytes()),
            go_executable_sha256=sha(self.go.read_bytes()),
            git_executable_sha256=sha(self.git.read_bytes()),
        )

        self.baseline.mkdir(mode=0o700)
        (self.baseline / "freeze").mkdir(mode=0o700)
        (self.baseline / "parent").mkdir(mode=0o700)
        for directory in ("reports", "results", "runtime"):
            (self.baseline / "parent" / directory).mkdir(mode=0o700)
        write_mode(self.baseline / "freeze/source.bundle", b"bundle")
        write_mode(self.baseline / "freeze/controller-runtime", b"runtime")
        capsule = {
            "$schema": "structural-retrieval-source-capsule.v2",
            "dirty_patch_sha256": "c" * 64,
            "head_commit": "6" * 40,
            "head_tree": "9" * 40,
            "inputs": [{
                "origin": "head",
                "path": "cli/tools/structural-search-eval-v2/main.go",
                "sha256": sha(self.main.read_bytes()),
            }],
            "repository_build_input_count": 1,
        }
        write_mode(self.baseline / "parent/controller-authority.v4.json", canonical({
            "controller_source_capsule": capsule,
            "runtime_source_capsule": capsule,
        }))
        write_mode(self.baseline / "parent/controller-output.v4.json", b"{}\n")
        write_mode(self.baseline / "parent/history.jsonl", b"{}\n")
        write_mode(self.baseline / "parent/privacy-scan.json", b"{}\n")
        for index in range(1, 7):
            write_mode(self.baseline / f"parent/results/{index:02d}.json", b"{}\n")
            write_mode(self.baseline / f"parent/reports/{index:02d}.json", b"{}\n")
            write_mode(self.baseline / f"parent/runtime/{index:02d}.stderr", b"")

        input_manifest = {
            "$schema": "structural-retrieval-protocol-v7-adapter-inputs.v1",
            "adapter_sha256": self.frozen.adapter_sha256,
            "driver_sha256": self.frozen.driver_sha256,
            "input_sha256": {
                "benchmark": self.frozen.benchmark_sha256,
                "fixture": self.frozen.fixture_sha256,
                "main": self.frozen.main_sha256,
                "main_test": self.frozen.main_test_sha256,
            },
            "privacy_policy_sha256": sha(self.policy.read_bytes()),
            "privacy_term_count": 1,
            "scorer_region_sha256": self.frozen.scorer_region_sha256,
            "status": "accepted",
            "threshold_checkin_seq": 7,
        }
        self.manifest = {
            "$schema": "structural-retrieval-protocol-v7-preparation.v1",
            "activation_proof_sha256": "3" * 64,
            "activation_record_hash": "4" * 64,
            "adapter_sha256": self.frozen.adapter_sha256,
            "authority_sha256": sha((self.baseline / "parent/controller-authority.v4.json").read_bytes()),
            "benchmark_sha256": self.frozen.benchmark_sha256,
            "bundle_heads_sha256": "5" * 64,
            "bundle_sha256": sha((self.baseline / "freeze/source.bundle").read_bytes()),
            "commit": "6" * 40,
            "controller_output_sha256": sha((self.baseline / "parent/controller-output.v4.json").read_bytes()),
            "driver_sha256": self.frozen.driver_sha256,
            "fixture_sha256": self.frozen.fixture_sha256,
            "history_sha256": sha((self.baseline / "parent/history.jsonl").read_bytes()),
            "input_manifest_sha256": sha(canonical(input_manifest)),
            "main_sha256": self.frozen.main_sha256,
            "main_test_sha256": self.frozen.main_test_sha256,
            "ordered_run_manifest_sha256": "7" * 64,
            "privacy_manifest_sha256": sha((self.baseline / "parent/privacy-scan.json").read_bytes()),
            "privacy_policy_sha256": sha(self.policy.read_bytes()),
            "privacy_term_count": 1,
            "run_count": 6,
            "runtime_sha256": sha((self.baseline / "freeze/controller-runtime").read_bytes()),
            "scorer_region_sha256": self.frozen.scorer_region_sha256,
            "source_capsule_sha256": "8" * 64,
            "status": "accepted",
            "tree": "9" * 40,
            "authorization_record_sha256": "a" * 64,
        }
        self.write_capabilities()
        self.write_manifest()

    @staticmethod
    def path_hash(path: Path) -> str:
        return sha(str(path.absolute()).encode())

    def write_capabilities(self) -> None:
        payload = {
            "$schema": "structural-retrieval-protocol-v7-baseline-authorization.v1",
            "activation_path_sha256": self.path_hash(self.activation),
            "adapter_sha256": self.frozen.adapter_sha256,
            "baseline_capture_authorized": True,
            "benchmark_sha256": self.frozen.benchmark_sha256,
            "candidate_execution_authorized": False,
            "driver_sha256": self.frozen.driver_sha256,
            "effect_claim": False,
            "fixture_sha256": self.frozen.fixture_sha256,
            "main_sha256": self.frozen.main_sha256,
            "main_test_sha256": self.frozen.main_test_sha256,
            "max_capture_count": 1,
            "output_root_sha256": self.path_hash(self.baseline),
            "privacy_policy_sha256": self.manifest["privacy_policy_sha256"],
            "scorer_region_sha256": self.frozen.scorer_region_sha256,
            "verdict": "authorized",
        }
        record = {
            "payload": payload,
            "payload_sha256": sha(canonical(payload)),
            "prev_hash": "GENESIS",
            "recorded_at": "2026-07-17T23:59:59Z",
            "seq": 1,
        }
        record["record_hash"] = sha(canonical(record))
        authorization_bytes = canonical(record) + b"\n"
        write_mode(self.authorization, authorization_bytes)
        activation = {
            "$schema": "structural-retrieval-protocol-v7-capture-activation.v1",
            "activation_path_sha256": self.path_hash(self.activation),
            "activation_record_hash": record["record_hash"],
            "adapter_sha256": self.frozen.adapter_sha256,
            "authorization_record_path_sha256": self.path_hash(self.authorization),
            "benchmark_sha256": self.frozen.benchmark_sha256,
            "driver_sha256": self.frozen.driver_sha256,
            "fixture_sha256": self.frozen.fixture_sha256,
            "main_sha256": self.frozen.main_sha256,
            "main_test_sha256": self.frozen.main_test_sha256,
            "max_capture_count": 1,
            "output_root_sha256": self.path_hash(self.baseline),
            "privacy_policy_sha256": self.manifest["privacy_policy_sha256"],
            "scorer_region_sha256": self.frozen.scorer_region_sha256,
            "verdict": "authorized",
        }
        receipt_bytes = canonical(activation)
        write_mode(self.receipt, receipt_bytes)
        self.manifest["activation_proof_sha256"] = sha(receipt_bytes)
        self.manifest["activation_record_hash"] = record["record_hash"]
        self.manifest["authorization_record_sha256"] = sha(authorization_bytes)

    def write_manifest(self, raw: bytes | None = None) -> None:
        write_mode(
            self.baseline / "capture-manifest.json",
            canonical(self.manifest) if raw is None else raw,
        )

    def config(self) -> object:
        return verifier.BaselineVerifierConfig(
            source_root=self.source,
            baseline_root=self.baseline,
            main_path=self.main,
            main_test_path=self.main_test,
            fixture_path=self.fixture,
            benchmark_path=self.benchmark,
            threshold_checkins_path=self.checkins,
            threshold_checkin_seq=7,
            privacy_policy_path=self.policy,
            adapter_path=self.adapter,
            preparation_verifier_path=self.preparation_verifier,
            activation_original_path=self.activation,
            consumed_receipt_path=self.receipt,
            authorization_record_path=self.authorization,
            go_executable_path=self.go,
            git_executable_path=self.git,
        )

    def harness_acceptance(self) -> dict[str, object]:
        return {
            "$schema": "structural-retrieval-baseline-acceptance.v1",
            "authority_sha256": self.manifest["authority_sha256"],
            "history_sha256": self.manifest["history_sha256"],
            "history_tail_record_hash": "b" * 64,
            "ordered_run_manifest_sha256": self.manifest["ordered_run_manifest_sha256"],
            "output_sha256": self.manifest["controller_output_sha256"],
            "privacy_manifest_sha256": self.manifest["privacy_manifest_sha256"],
            "run_count": 6,
            "validated_source_main_sha256": self.frozen.main_sha256,
            "validated_source_test_sha256": self.frozen.main_test_sha256,
            "verdict": "accepted",
        }


class AcceptingRunner:
    def __init__(self, baseline: SyntheticBaseline, mutate: dict[str, object] | None = None) -> None:
        self.baseline = baseline
        self.mutate = mutate or {}

    def __call__(self, _command: list[str], *, cwd: Path, env: dict[str, str], timeout: int, output_cap: int) -> object:
        del cwd, timeout, output_cap
        config = json.loads(Path(env["C3_PROTOCOL_V7_BASELINE_VERIFIER_CONFIG"]).read_bytes())
        acceptance = self.baseline.harness_acceptance()
        acceptance.update(self.mutate)
        result = {
            "$schema": "structural-retrieval-protocol-v7-baseline-verifier-harness.v1",
            "baseline_acceptance": acceptance,
            "status": "accepted",
        }
        write_mode(Path(config["result_path"]), canonical(result))
        return verifier.ProcessResult(0, b"", b"")


class BaselineValidatorTests(unittest.TestCase):
    def test_trust_closure_requires_explicit_tools_and_minimal_path(self) -> None:
        with tempfile.TemporaryDirectory() as temporary:
            root = Path(temporary)
            go = root / "go"
            git = root / "git"
            write_mode(go, b"go", 0o755)
            write_mode(git, b"git", 0o755)
            tools = verifier._validate_tools(go, git, sha(go.read_bytes()), sha(git.read_bytes()))
            private = root / "private"
            path = verifier._materialize_minimal_tool_path(private, tools)
            self.assertEqual(set(item.name for item in path.iterdir()), {"git", "go"})

    def test_parser_orchestration_emits_exact_parent_payload(self) -> None:
        with tempfile.TemporaryDirectory() as temporary:
            baseline = SyntheticBaseline(Path(temporary))
            result = verifier.verify_baseline(
                baseline.config(),
                frozen=baseline.frozen,
                runner=AcceptingRunner(baseline),
            )
            self.assertEqual(
                set(result),
                {"baseline_acceptance", "effect_claim", "event", "role", "status", "worker_id"},
            )
            self.assertEqual(result["worker_id"], "validator-baseline-protocol-v7")
            self.assertEqual(result["status"], "accepted")
            self.assertFalse(result["effect_claim"])

    def test_harness_result_cannot_change_manifest_binding(self) -> None:
        with tempfile.TemporaryDirectory() as temporary:
            baseline = SyntheticBaseline(Path(temporary))
            result = verifier.verify_baseline(
                baseline.config(),
                frozen=baseline.frozen,
                runner=AcceptingRunner(baseline, {"output_sha256": "c" * 64}),
            )
            self.assertEqual(result["status"], "rejected")
            self.assertEqual(result["failure_class"], "harness_result_invalid")

    def test_capture_manifest_is_canonical_and_exact(self) -> None:
        with tempfile.TemporaryDirectory() as temporary:
            baseline = SyntheticBaseline(Path(temporary))
            baseline.write_manifest(canonical({**baseline.manifest, "extra": True}))
            result = verifier.verify_baseline(
                baseline.config(), frozen=baseline.frozen, runner=AcceptingRunner(baseline)
            )
            self.assertEqual(result["failure_class"], "metadata_invalid")

    def test_layout_rejects_extra_and_wrong_mode(self) -> None:
        with tempfile.TemporaryDirectory() as temporary:
            baseline = SyntheticBaseline(Path(temporary))
            write_mode(baseline.baseline / "extra", b"x")
            result = verifier.verify_baseline(
                baseline.config(), frozen=baseline.frozen, runner=AcceptingRunner(baseline)
            )
            self.assertEqual(result["failure_class"], "baseline_layout_invalid")
        with tempfile.TemporaryDirectory() as temporary:
            baseline = SyntheticBaseline(Path(temporary))
            (baseline.baseline / "parent/history.jsonl").chmod(0o644)
            result = verifier.verify_baseline(
                baseline.config(), frozen=baseline.frozen, runner=AcceptingRunner(baseline)
            )
            self.assertEqual(result["failure_class"], "baseline_layout_invalid")

    def test_one_shot_original_present_or_receipt_changed_rejects(self) -> None:
        with tempfile.TemporaryDirectory() as temporary:
            baseline = SyntheticBaseline(Path(temporary))
            write_mode(baseline.activation, b"replacement")
            result = verifier.verify_baseline(
                baseline.config(), frozen=baseline.frozen, runner=AcceptingRunner(baseline)
            )
            self.assertEqual(result["failure_class"], "activation_invalid")
        with tempfile.TemporaryDirectory() as temporary:
            baseline = SyntheticBaseline(Path(temporary))
            write_mode(baseline.receipt, baseline.receipt.read_bytes() + b"\n")
            result = verifier.verify_baseline(
                baseline.config(), frozen=baseline.frozen, runner=AcceptingRunner(baseline)
            )
            self.assertEqual(result["failure_class"], "activation_invalid")

    def test_authorization_chain_and_path_binding_rejects(self) -> None:
        with tempfile.TemporaryDirectory() as temporary:
            baseline = SyntheticBaseline(Path(temporary))
            record = json.loads(baseline.authorization.read_bytes())
            record["payload"]["max_capture_count"] = 2
            write_mode(baseline.authorization, canonical(record) + b"\n")
            result = verifier.verify_baseline(
                baseline.config(), frozen=baseline.frozen, runner=AcceptingRunner(baseline)
            )
            self.assertEqual(result["failure_class"], "authorization_invalid")

    def test_stage_residue_rejects(self) -> None:
        with tempfile.TemporaryDirectory() as temporary:
            baseline = SyntheticBaseline(Path(temporary))
            (baseline.baseline.parent / ".c3-v7-adapter-residue").mkdir()
            result = verifier.verify_baseline(
                baseline.config(), frozen=baseline.frozen, runner=AcceptingRunner(baseline)
            )
            self.assertEqual(result["failure_class"], "stage_residue")

    def test_default_runner_nonzero_and_stderr_reject(self) -> None:
        for label, body in (
            ("nonzero", "raise SystemExit(9)\n"),
            ("stderr", "import sys\nsys.stderr.write('generic failure\\n')\n"),
        ):
            with self.subTest(label=label), tempfile.TemporaryDirectory() as temporary:
                baseline = SyntheticBaseline(Path(temporary))
                go = write_fake_tool(baseline.base / label, "go", body)
                config, frozen = with_go(baseline, go)
                result = verifier.verify_baseline(config, frozen=frozen)
                self.assertEqual(result["failure_class"], "harness_rejected")

    def test_default_runner_timeout_output_cap_and_process_tree_cleanup(self) -> None:
        with tempfile.TemporaryDirectory() as temporary:
            baseline = SyntheticBaseline(Path(temporary))
            marker = baseline.base / "child.pid"
            go = write_fake_tool(
                baseline.base / "timeout", "go",
                "import pathlib,subprocess,time\n"
                f"child=subprocess.Popen(['/bin/sleep','60'])\npathlib.Path({str(marker)!r}).write_text(str(child.pid))\n"
                "time.sleep(60)\n",
            )
            config, frozen = with_go(baseline, go)
            with mock.patch.object(verifier, "PROCESS_TIMEOUT_SECONDS", 1):
                result = verifier.verify_baseline(config, frozen=frozen)
            self.assertEqual(result["failure_class"], "harness_timeout")
            wait_for_pid_exit(int(marker.read_text()))

        with tempfile.TemporaryDirectory() as temporary:
            baseline = SyntheticBaseline(Path(temporary))
            go = write_fake_tool(baseline.base / "cap", "go", "import os\nos.write(1,b'x'*4096)\n")
            config, frozen = with_go(baseline, go)
            with mock.patch.object(verifier, "PROCESS_OUTPUT_CAP", 64):
                result = verifier.verify_baseline(config, frozen=frozen)
            self.assertEqual(result["failure_class"], "harness_output_cap")

    def test_source_threshold_and_baseline_mutation_during_runner_reject(self) -> None:
        for target_name, expected in (
            ("main", "harness_source_invalid"),
            ("checkins", "governed_input_changed"),
            ("history", "governed_input_changed"),
        ):
            with self.subTest(target=target_name), tempfile.TemporaryDirectory() as temporary:
                baseline = SyntheticBaseline(Path(temporary))
                target = {
                    "main": baseline.main,
                    "checkins": baseline.checkins,
                    "history": baseline.baseline / "parent/history.jsonl",
                }[target_name]
                mutated = target.read_bytes() + b"changed\n"
                fake_go = write_accepting_go(
                    baseline.base / ("mutate-" + target_name), baseline,
                    f"pathlib.Path({str(target)!r}).write_bytes({mutated!r})\n",
                )
                config, frozen = with_go(baseline, fake_go)
                result = verifier.verify_baseline(config, frozen=frozen)
                self.assertEqual(result["status"], "rejected")
                self.assertEqual(result["failure_class"], expected)

    def test_hostile_ambient_path_is_replaced_by_private_pinned_tools(self) -> None:
        with tempfile.TemporaryDirectory() as temporary:
            baseline = SyntheticBaseline(Path(temporary))
            fake_go = write_accepting_go(
                baseline.base / "hostile-check", baseline,
                "names={p.name for p in pathlib.Path(os.environ['PATH']).iterdir()}\n"
                "assert names == {'go','git'} and 'hostile' not in os.environ['PATH']\n",
            )
            config, frozen = with_go(baseline, fake_go)
            with mock.patch.dict(os.environ, {"PATH": "/hostile/ambient/path"}):
                result = verifier.verify_baseline(config, frozen=frozen)
            self.assertEqual(result["status"], "accepted")

    def test_default_runner_sigterm_and_sighup_kill_tree_and_remove_private_temp(self) -> None:
        launcher = (
            "import importlib.util,json,sys;"
            f"p={str(MODULE_PATH)!r};"
            "s=importlib.util.spec_from_file_location('signal_baseline_verifier',p);"
            "m=importlib.util.module_from_spec(s);sys.modules[s.name]=m;s.loader.exec_module(m);"
            "c=json.loads(__import__('os').environ['VERIFIER_CONFIG']);"
            "f=json.loads(__import__('os').environ['VERIFIER_FROZEN']);"
            "r=m.verify_baseline(m.BaselineVerifierConfig(**{k:__import__('pathlib').Path(v) "
            "if k.endswith('_path') or k.endswith('_root') else v for k,v in c.items()}),"
            "frozen=m.FrozenInputs(**f));print(m.canonical_json(r).decode());"
            "raise SystemExit(0 if r['status']=='accepted' else 2)"
        )
        for sent_signal in (signal.SIGTERM, signal.SIGHUP):
            with self.subTest(signal=sent_signal), tempfile.TemporaryDirectory() as temporary:
                baseline = SyntheticBaseline(Path(temporary))
                child_marker = baseline.base / f"{sent_signal.name.lower()}-child.pid"
                ready_marker = baseline.base / f"{sent_signal.name.lower()}-ready"
                fake_go = write_fake_tool(
                    baseline.base / sent_signal.name.lower(), "go",
                    "import pathlib,subprocess,time\n"
                    f"child=subprocess.Popen(['/bin/sleep','60'])\npathlib.Path({str(child_marker)!r}).write_text(str(child.pid))\n"
                    f"pathlib.Path({str(ready_marker)!r}).write_text('ready')\n"
                    "time.sleep(60)\n",
                )
                config, frozen = with_go(baseline, fake_go)
                private_parent = baseline.base / "private-temp"
                private_parent.mkdir(mode=0o700)
                env = dict(os.environ)
                env["PATH"] = "/hostile/ambient/path"
                env["TMPDIR"] = str(private_parent)
                env["VERIFIER_CONFIG"] = json.dumps({
                    key: str(value) if isinstance(value, Path) else value
                    for key, value in config.__dict__.items()
                })
                env["VERIFIER_FROZEN"] = json.dumps(frozen.__dict__)
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
                    "$schema": verifier.REJECTED_SCHEMA,
                    "failure_class": "harness_interrupted",
                    "status": "rejected",
                })
                wait_for_pid_exit(int(child_marker.read_text()))
                self.assertEqual(list(private_parent.glob("c3-v7-baseline-verifier-*")), [])

    def test_generated_clone_of_retained_bb05_runs_full_default_verifier_and_cleans(self) -> None:
        repository = Path(__file__).resolve().parents[1]
        retained = Path.home() / ".cache/c3-v7-baseline-bb-05"
        before = verifier._snapshot_baseline_tree(retained)
        temporary_path: Path | None = None
        with tempfile.TemporaryDirectory(prefix="generic-v7-generated-clone-") as temporary:
            temporary_path = Path(temporary)
            clone = temporary_path / "baseline"
            shutil.copytree(retained, clone, copy_function=shutil.copy2)
            policy = temporary_path / "privacy-policy.json"
            shutil.copy2(Path.home() / ".cache/c3-v7-baseline-bb-02.policy.json", policy)
            policy.chmod(0o600)
            activation = temporary_path / "capture.activation.json"
            receipt = Path(str(activation) + ".consumed")
            authorization = temporary_path / "capture.authorization.jsonl"
            manifest_path = clone / "capture-manifest.json"
            manifest = json.loads(manifest_path.read_bytes())

            payload = {
                "$schema": verifier.AUTHORIZATION_SCHEMA,
                "activation_path_sha256": sha(str(activation.absolute()).encode()),
                "adapter_sha256": manifest["adapter_sha256"],
                "baseline_capture_authorized": True,
                "benchmark_sha256": manifest["benchmark_sha256"],
                "candidate_execution_authorized": False,
                "driver_sha256": manifest["driver_sha256"],
                "effect_claim": False,
                "fixture_sha256": manifest["fixture_sha256"],
                "main_sha256": manifest["main_sha256"],
                "main_test_sha256": manifest["main_test_sha256"],
                "max_capture_count": 1,
                "output_root_sha256": sha(str(clone.absolute()).encode()),
                "privacy_policy_sha256": sha(policy.read_bytes()),
                "scorer_region_sha256": manifest["scorer_region_sha256"],
                "verdict": "authorized",
            }
            record = {
                "payload": payload,
                "payload_sha256": sha(canonical(payload)),
                "prev_hash": "GENESIS",
                "recorded_at": "2026-07-18T00:00:00Z",
                "seq": 1,
            }
            record["record_hash"] = sha(canonical(record))
            authorization_bytes = canonical(record) + b"\n"
            write_mode(authorization, authorization_bytes)
            activation_payload = {
                "$schema": verifier.ACTIVATION_SCHEMA,
                "activation_path_sha256": sha(str(activation.absolute()).encode()),
                "activation_record_hash": record["record_hash"],
                "adapter_sha256": manifest["adapter_sha256"],
                "authorization_record_path_sha256": sha(str(authorization.absolute()).encode()),
                "benchmark_sha256": manifest["benchmark_sha256"],
                "driver_sha256": manifest["driver_sha256"],
                "fixture_sha256": manifest["fixture_sha256"],
                "main_sha256": manifest["main_sha256"],
                "main_test_sha256": manifest["main_test_sha256"],
                "max_capture_count": 1,
                "output_root_sha256": sha(str(clone.absolute()).encode()),
                "privacy_policy_sha256": sha(policy.read_bytes()),
                "scorer_region_sha256": manifest["scorer_region_sha256"],
                "verdict": "authorized",
            }
            receipt_bytes = canonical(activation_payload)
            write_mode(receipt, receipt_bytes)
            manifest.update({
                "activation_proof_sha256": sha(receipt_bytes),
                "activation_record_hash": record["record_hash"],
                "authorization_record_sha256": sha(authorization_bytes),
            })
            write_mode(manifest_path, canonical(manifest))

            config = verifier.BaselineVerifierConfig(
                source_root=repository,
                baseline_root=clone,
                main_path=repository / "cli/tools/structural-search-eval-v2/main.go",
                main_test_path=repository / "cli/tools/structural-search-eval-v2/main_test.go",
                fixture_path=repository / "research/eval/structural-retrieval/fixtures.dev.v2.jsonl",
                benchmark_path=repository / "research/eval/structural-retrieval/benchmark.v2.json",
                threshold_checkins_path=repository / ".okra/runs/c3-rag-autoresearch-v1/checkins.jsonl",
                threshold_checkin_seq=5,
                privacy_policy_path=policy,
                adapter_path=repository / "scripts/structural_retrieval_protocol_v7.py",
                preparation_verifier_path=repository / "scripts/validate_structural_retrieval_preparation_v7.py",
                activation_original_path=activation,
                consumed_receipt_path=receipt,
                authorization_record_path=authorization,
                go_executable_path=REAL_GO,
                git_executable_path=REAL_GIT,
            )
            result = verifier.verify_baseline(config)
            self.assertEqual(result["status"], "accepted", result)
            self.assertEqual(result["worker_id"], "validator-baseline-protocol-v7")
        assert temporary_path is not None
        self.assertFalse(temporary_path.exists())
        self.assertEqual(verifier._snapshot_baseline_tree(retained), before)


if __name__ == "__main__":
    unittest.main()
