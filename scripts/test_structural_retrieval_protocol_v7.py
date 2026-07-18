#!/usr/bin/env python3

from __future__ import annotations

import hashlib
import json
from dataclasses import replace
from pathlib import Path
import subprocess
import sys
import tempfile
import time
import unittest
from unittest import mock

from scripts.structural_retrieval_protocol_v7 import (
    ACCEPTED_BENCHMARK_SHA256,
    ACCEPTED_FIXTURE_SHA256,
    ACCEPTED_MAIN_SHA256,
    ACCEPTED_MAIN_TEST_SHA256,
    ACCEPTED_SCORER_REGION_SHA256,
    ACTIVATION_SCHEMA,
    DRIVER_SHA256,
    DRIVER_SOURCE,
    CaptureConfig,
    ProtocolV7AdapterError,
    activation_self_test,
    capture_baseline,
    prepare_baseline,
    run_driver,
    self_test,
    validate_inputs,
    validate_privacy_policy,
)


class ProtocolV7AdapterTest(unittest.TestCase):
    def setUp(self) -> None:
        self.temp = tempfile.TemporaryDirectory(prefix="c3-v7-adapter-test-")
        self.root = Path(self.temp.name)

    def tearDown(self) -> None:
        self.temp.cleanup()

    def write(self, relative: str, data: bytes, mode: int = 0o644) -> Path:
        path = self.root / relative
        path.parent.mkdir(parents=True, exist_ok=True)
        path.write_bytes(data)
        path.chmod(mode)
        return path

    def config(self, policy: Path) -> CaptureConfig:
        main = self.write("source/cli/tools/structural-search-eval-v2/main.go", b"main")
        main_test = self.write("source/cli/tools/structural-search-eval-v2/main_test.go", b"test")
        fixture = self.write("source/research/eval/structural-retrieval/fixtures.dev.v2.jsonl", b"fixture")
        benchmark = self.write("source/research/eval/structural-retrieval/benchmark.v2.json", b"benchmark")
        checkins = self.write("checkins.jsonl", b"{}\n")
        return CaptureConfig(
            source_root=self.root / "source",
            main_path=main,
            main_test_path=main_test,
            fixture_path=fixture,
            benchmark_path=benchmark,
            threshold_checkins_path=checkins,
            threshold_checkin_seq=1,
            privacy_policy_path=policy,
            output_root=self.root / "evidence",
            expected_main_sha256=hashlib.sha256(b"main").hexdigest(),
            expected_main_test_sha256=hashlib.sha256(b"test").hexdigest(),
            expected_fixture_sha256=hashlib.sha256(b"fixture").hexdigest(),
            expected_benchmark_sha256=hashlib.sha256(b"benchmark").hexdigest(),
        )

    def canonical_policy(self) -> Path:
        return self.write(
            "policy.json",
            b'{"$schema":"structural-retrieval-privacy-policy.v1","deny_terms":["generic sentinel"]}',
            0o600,
        )

    def authorized_capture(self, config: CaptureConfig, activation: Path) -> tuple[CaptureConfig, dict[str, object]]:
        manifest = validate_inputs(config)
        authorization_path = self.root / "authorization.jsonl"
        payload = {
            "$schema": "structural-retrieval-protocol-v7-baseline-authorization.v1",
            "activation_path_sha256": hashlib.sha256(str(activation.absolute()).encode()).hexdigest(),
            "adapter_sha256": manifest["adapter_sha256"],
            "baseline_capture_authorized": True,
            "benchmark_sha256": config.expected_benchmark_sha256,
            "candidate_execution_authorized": False,
            "driver_sha256": DRIVER_SHA256,
            "effect_claim": False,
            "fixture_sha256": config.expected_fixture_sha256,
            "main_sha256": config.expected_main_sha256,
            "main_test_sha256": config.expected_main_test_sha256,
            "max_capture_count": 1,
            "output_root_sha256": hashlib.sha256(str(config.output_root.absolute()).encode()).hexdigest(),
            "privacy_policy_sha256": manifest["privacy_policy_sha256"],
            "scorer_region_sha256": ACCEPTED_SCORER_REGION_SHA256,
            "verdict": "authorized",
        }
        payload_sha256 = hashlib.sha256(
            json.dumps(payload, sort_keys=True, separators=(",", ":")).encode()
        ).hexdigest()
        record_without_hash = {
            "payload": payload,
            "payload_sha256": payload_sha256,
            "prev_hash": "GENESIS",
            "recorded_at": "2026-07-17T20:00:00Z",
            "seq": 1,
        }
        record_hash = hashlib.sha256(
            json.dumps(record_without_hash, sort_keys=True, separators=(",", ":")).encode()
        ).hexdigest()
        record = dict(record_without_hash)
        record["record_hash"] = record_hash
        authorization_path.write_bytes(
            json.dumps(record, sort_keys=True, separators=(",", ":")).encode() + b"\n"
        )
        authorization_path.chmod(0o600)
        config = replace(config, authorization_record_path=authorization_path)
        activation_payload = {
            "$schema": ACTIVATION_SCHEMA,
            "activation_path_sha256": hashlib.sha256(str(activation.absolute()).encode()).hexdigest(),
            "activation_record_hash": record_hash,
            "authorization_record_path_sha256": hashlib.sha256(
                str(authorization_path.absolute()).encode()
            ).hexdigest(),
            "verdict": "authorized",
            "adapter_sha256": manifest["adapter_sha256"],
            "driver_sha256": DRIVER_SHA256,
            "main_sha256": config.expected_main_sha256,
            "main_test_sha256": config.expected_main_test_sha256,
            "fixture_sha256": config.expected_fixture_sha256,
            "benchmark_sha256": config.expected_benchmark_sha256,
            "scorer_region_sha256": ACCEPTED_SCORER_REGION_SHA256,
            "privacy_policy_sha256": manifest["privacy_policy_sha256"],
            "max_capture_count": 1,
            "output_root_sha256": hashlib.sha256(str(config.output_root.absolute()).encode()).hexdigest(),
        }
        activation.write_bytes(json.dumps(activation_payload, sort_keys=True, separators=(",", ":")).encode())
        activation.chmod(0o600)
        return config, activation_payload

    def test_frozen_protocol_hashes_and_driver_are_bound(self) -> None:
        self.assertEqual(ACCEPTED_MAIN_SHA256, "33b6452a0e34d825ef2b895609daa77b91eb129ebdc1e6be8f7f7755fc8b9c6e")
        self.assertEqual(ACCEPTED_MAIN_TEST_SHA256, "0830a0671aa9eb146902dc6fa56e885126b89a3312ca6c38fd8507fa10481dbf")
        self.assertEqual(ACCEPTED_FIXTURE_SHA256, "15f57120c6aa9ae07bf4fdacd6ad783afa5e70ed8ebebaff3a42dcf4249e677e")
        self.assertEqual(ACCEPTED_BENCHMARK_SHA256, "b960525cc42216e6598452946da5fb68735bbf989f311f170cedcfdbe92bf0d5")
        self.assertEqual(ACCEPTED_SCORER_REGION_SHA256, "c1669d43b13a36eb1454a01a5ea7e444fe67ce28fdae9628da083927f093cbcb")
        self.assertEqual(DRIVER_SHA256, hashlib.sha256(DRIVER_SOURCE.encode()).hexdigest())

    def test_privacy_policy_requires_exact_canonical_mode_0600_regular_file(self) -> None:
        policy = self.canonical_policy()
        binding = validate_privacy_policy(policy)
        self.assertEqual(binding.term_count, 1)
        self.assertEqual(binding.sha256, hashlib.sha256(policy.read_bytes()).hexdigest())

        policy.chmod(0o644)
        with self.assertRaisesRegex(ProtocolV7AdapterError, "privacy_policy_invalid"):
            validate_privacy_policy(policy)

        policy.chmod(0o600)
        policy.write_bytes(b'{"deny_terms":["generic sentinel"],"$schema":"structural-retrieval-privacy-policy.v1"}')
        with self.assertRaisesRegex(ProtocolV7AdapterError, "privacy_policy_invalid"):
            validate_privacy_policy(policy)

        policy.unlink()
        target = self.write("policy-target.json", b"{}", 0o600)
        policy.symlink_to(target)
        with self.assertRaisesRegex(ProtocolV7AdapterError, "privacy_policy_invalid"):
            validate_privacy_policy(policy)

    def test_validate_inputs_emits_only_generic_hashes_counts_and_roles(self) -> None:
        policy = self.canonical_policy()
        config = self.config(policy)
        manifest = validate_inputs(config)
        encoded = json.dumps(manifest, sort_keys=True)
        self.assertEqual(manifest["status"], "accepted")
        self.assertEqual(manifest["privacy_term_count"], 1)
        self.assertNotIn(str(self.root), encoded)
        self.assertNotIn("generic sentinel", encoded)
        self.assertEqual(set(manifest["input_sha256"]), {"benchmark", "fixture", "main", "main_test"})

    def test_capture_activation_is_checked_before_executor_or_output_effect(self) -> None:
        policy = self.canonical_policy()
        config = self.config(policy)
        activation = self.write("activation.json", b"{}", 0o600)
        executor = mock.Mock()

        with self.assertRaisesRegex(ProtocolV7AdapterError, "authorization_invalid"):
            capture_baseline(config, activation, executor=executor)

        executor.assert_not_called()
        self.assertFalse(config.output_root.exists())

    def test_capture_activation_binds_every_frozen_identity_before_executor(self) -> None:
        policy = self.canonical_policy()
        config = self.config(policy)
        activation = self.root / "activation.json"
        config, activation_payload = self.authorized_capture(config, activation)
        input_manifest = validate_inputs(config)
        executor = mock.Mock(return_value={"status": "captured"})

        result = capture_baseline(config, activation, executor=executor)

        executor.assert_called_once_with("capture-baseline", config, input_manifest)
        self.assertEqual(result, {"status": "captured"})
        self.assertFalse(activation.exists())
        self.assertTrue(Path(str(activation) + ".consumed").is_file())

        activation.write_bytes(json.dumps(activation_payload, sort_keys=True, separators=(",", ":")).encode())
        activation.chmod(0o600)
        executor.reset_mock()
        with self.assertRaisesRegex(ProtocolV7AdapterError, "activation_used"):
            capture_baseline(config, activation, executor=executor)
        executor.assert_not_called()

        second_activation = self.root / "activation-two.json"
        activation_payload["activation_path_sha256"] = hashlib.sha256(str(second_activation.absolute()).encode()).hexdigest()
        second_activation.write_bytes(json.dumps(activation_payload, sort_keys=True, separators=(",", ":")).encode())
        second_activation.chmod(0o600)
        executor.reset_mock()
        with self.assertRaisesRegex(ProtocolV7AdapterError, "authorization_invalid"):
            capture_baseline(config, second_activation, executor=executor)
        executor.assert_not_called()

    def test_fabricated_consumed_receipt_never_reaches_driver_process(self) -> None:
        policy = self.canonical_policy()
        config = self.config(policy)
        activation = self.root / "activation.json"
        config, _ = self.authorized_capture(config, activation)
        manifest = validate_inputs(config)
        consumed = Path(str(activation) + ".consumed")
        consumed.write_bytes(activation.read_bytes())
        consumed.chmod(0o600)
        activation.unlink()

        with mock.patch("scripts.structural_retrieval_protocol_v7._run_driver_process") as process:
            with self.assertRaisesRegex(ProtocolV7AdapterError, "activation_invalid"):
                run_driver("capture-baseline", config, manifest, consumed_activation=consumed)
        process.assert_not_called()
        self.assertFalse(config.output_root.exists())

    def test_direct_native_capture_driver_requires_consumed_activation_proof(self) -> None:
        policy = self.canonical_policy()
        config = self.config(policy)
        manifest = validate_inputs(config)

        with self.assertRaisesRegex(ProtocolV7AdapterError, "activation_invalid"):
            run_driver("capture-baseline", config, manifest)

        self.assertFalse(config.output_root.exists())

    def test_failed_disposable_executor_removes_its_entire_output(self) -> None:
        policy = self.canonical_policy()
        config = self.config(policy)

        def failing_executor(_command: str, cfg: CaptureConfig, _manifest: dict[str, object]) -> dict[str, object]:
            cfg.output_root.mkdir()
            (cfg.output_root / "partial").write_text("generic")
            raise ProtocolV7AdapterError("synthetic_failure")

        with self.assertRaisesRegex(ProtocolV7AdapterError, "synthetic_failure"):
            self_test(config, executor=failing_executor)
        self.assertFalse(config.output_root.exists())

    def test_native_activation_self_test_replays_receipt_and_authorization_without_capture(self) -> None:
        repository = Path(__file__).resolve().parents[1]
        policy = self.canonical_policy()
        config = CaptureConfig(
            source_root=repository,
            main_path=repository / "cli/tools/structural-search-eval-v2/main.go",
            main_test_path=repository / "cli/tools/structural-search-eval-v2/main_test.go",
            fixture_path=repository / "research/eval/structural-retrieval/fixtures.dev.v2.jsonl",
            benchmark_path=repository / "research/eval/structural-retrieval/benchmark.v2.json",
            threshold_checkins_path=repository / ".okra/runs/c3-rag-autoresearch-v1/checkins.jsonl",
            threshold_checkin_seq=5,
            privacy_policy_path=policy,
            output_root=self.root / "native-activation-self-test-output-must-remain-absent",
        )
        activation = self.root / "native-activation-self-test.json"
        config, _ = self.authorized_capture(config, activation)

        result = activation_self_test(config, activation)

        self.assertEqual(result["status"], "accepted_disposable_non_study")
        self.assertIn("activation_proof_sha256", result)
        self.assertIn("authorization_record_sha256", result)
        self.assertFalse(activation.exists())
        self.assertFalse(Path(str(activation) + ".consumed").exists())
        self.assertFalse(config.output_root.exists())

    def test_real_preparation_is_repeatable_and_self_test_leaves_no_transaction(self) -> None:
        repository = Path(__file__).resolve().parents[1]
        policy = self.canonical_policy()

        def real_config(output: Path) -> CaptureConfig:
            return CaptureConfig(
                source_root=repository,
                main_path=repository / "cli/tools/structural-search-eval-v2/main.go",
                main_test_path=repository / "cli/tools/structural-search-eval-v2/main_test.go",
                fixture_path=repository / "research/eval/structural-retrieval/fixtures.dev.v2.jsonl",
                benchmark_path=repository / "research/eval/structural-retrieval/benchmark.v2.json",
                threshold_checkins_path=repository / ".okra/runs/c3-rag-autoresearch-v1/checkins.jsonl",
                threshold_checkin_seq=5,
                privacy_policy_path=policy,
                output_root=output,
            )

        first_root = self.root / "prepare-one"
        second_root = self.root / "prepare-two"
        first = prepare_baseline(real_config(first_root))
        second = prepare_baseline(real_config(second_root))

        self.assertEqual(first, second)

        def tree_hashes(root: Path) -> dict[str, str]:
            return {
                path.relative_to(root).as_posix(): hashlib.sha256(path.read_bytes()).hexdigest()
                for path in sorted(root.rglob("*"))
                if path.is_file()
            }

        self.assertEqual(tree_hashes(first_root), tree_hashes(second_root))
        disposable_root = self.root / "self-test-must-remain-absent"
        result = self_test(real_config(disposable_root))
        self.assertEqual(result["status"], "accepted_disposable_non_study")
        self.assertEqual(result["run_count"], 6)
        self.assertFalse(disposable_root.exists())

    def test_real_sigterm_kills_driver_session_and_removes_transaction_roots(self) -> None:
        repository = Path(__file__).resolve().parents[1]
        policy = self.canonical_policy()
        output = self.root / "interrupted-output-must-remain-absent"
        temporary_parent = Path(tempfile.gettempdir())
        before_driver_roots = {path.name for path in temporary_parent.glob("c3-v7-adapter-driver-*")}
        command = [
            sys.executable,
            str(repository / "scripts/structural_retrieval_protocol_v7.py"),
            "self-test",
            "--source-root",
            str(repository),
            "--main",
            str(repository / "cli/tools/structural-search-eval-v2/main.go"),
            "--main-test",
            str(repository / "cli/tools/structural-search-eval-v2/main_test.go"),
            "--fixtures",
            str(repository / "research/eval/structural-retrieval/fixtures.dev.v2.jsonl"),
            "--benchmark",
            str(repository / "research/eval/structural-retrieval/benchmark.v2.json"),
            "--threshold-checkins",
            str(repository / ".okra/runs/c3-rag-autoresearch-v1/checkins.jsonl"),
            "--threshold-checkin-seq",
            "5",
            "--privacy-policy",
            str(policy),
            "--output-root",
            str(output),
        ]
        process = subprocess.Popen(
            command,
            cwd=repository,
            stdin=subprocess.DEVNULL,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            start_new_session=True,
        )
        deadline = time.monotonic() + 30
        stage_seen = False
        while time.monotonic() < deadline and process.poll() is None:
            if any(self.root.glob(".c3-v7-adapter-*")):
                stage_seen = True
                break
            time.sleep(0.05)
        self.assertTrue(stage_seen, "adapter did not enter its governed transaction stage")
        time.sleep(0.5)
        process.terminate()
        stdout, stderr = process.communicate(timeout=30)

        self.assertEqual(process.returncode, 2)
        self.assertEqual(stderr, b"")
        result = json.loads(stdout)
        self.assertEqual(result["status"], "rejected")
        self.assertEqual(result["error_class"], "driver_interrupted")
        self.assertNotIn(str(self.root), stdout.decode())
        self.assertFalse(output.exists())
        self.assertEqual(list(self.root.glob(".c3-v7-adapter-*")), [])

        deadline = time.monotonic() + 10
        while time.monotonic() < deadline:
            after = {path.name for path in temporary_parent.glob("c3-v7-adapter-driver-*")}
            if after <= before_driver_roots:
                break
            time.sleep(0.05)
        self.assertLessEqual(
            {path.name for path in temporary_parent.glob("c3-v7-adapter-driver-*")},
            before_driver_roots,
        )
        for entry in Path("/proc").glob("[0-9]*/cmdline"):
            try:
                command_line = entry.read_bytes().replace(b"\0", b" ")
            except OSError:
                continue
            self.assertNotIn(str(self.root).encode(), command_line)


if __name__ == "__main__":
    unittest.main()
