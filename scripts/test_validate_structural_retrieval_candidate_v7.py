from __future__ import annotations

import hashlib
import json
import os
from pathlib import Path
import shutil
import tempfile
import unittest
from unittest import mock

from scripts import validate_structural_retrieval_candidate_v7 as validator


def canonical(value: object) -> bytes:
    return json.dumps(value, sort_keys=True, separators=(",", ":")).encode()


class CandidateValidatorUnitTests(unittest.TestCase):
    def test_validator_is_separate_and_execution_is_false(self) -> None:
        self.assertFalse(validator.CANDIDATE_EXECUTION_AUTHORIZED)
        self.assertEqual(validator.ACCEPTED_MAIN_SHA256, "33b6452a0e34d825ef2b895609daa77b91eb129ebdc1e6be8f7f7755fc8b9c6e")

    def test_capability_layout_is_exact_private_and_no_symlinks(self) -> None:
        with tempfile.TemporaryDirectory() as raw:
            root = Path(raw)
            validator.make_test_capability(root)
            snapshot = validator.validate_capability_layout(root)
            self.assertEqual(set(snapshot), set(validator.CAPABILITY_FILES))
            extra = root / "extra"
            extra.write_bytes(b"x")
            extra.chmod(0o600)
            with self.assertRaisesRegex(validator.CandidateValidatorError, "capability_layout_invalid"):
                validator.validate_capability_layout(root)

    def test_manifest_rejects_copied_writer_status_and_wrong_binding(self) -> None:
        manifest = validator.make_test_capability_manifest()
        authority = validator.make_test_candidate_authority(manifest)
        validator.validate_capability_manifest(canonical(manifest), validator.make_test_capability_hashes(manifest), authority)
        manifest["status"] = "keep"
        with self.assertRaisesRegex(validator.CandidateValidatorError, "capability_manifest_invalid"):
            validator.validate_capability_manifest(canonical(manifest), {}, authority)

    def test_manifest_cross_binds_commit_tree_delta_and_bundle_to_authority(self) -> None:
        manifest = validator.make_test_capability_manifest()
        authority = validator.make_test_candidate_authority(manifest)
        hashes = validator.make_test_capability_hashes(manifest)
        for field, bad in (("candidate_commit", "c" * 40), ("candidate_tree", "d" * 40), ("candidate_delta_sha256", "e" * 64)):
            changed = dict(manifest)
            changed[field] = bad
            with self.subTest(field=field), self.assertRaisesRegex(validator.CandidateValidatorError, "capability_manifest_invalid"):
                validator.validate_capability_manifest(canonical(changed), hashes, authority)
        changed_authority = json.loads(json.dumps(authority))
        changed_authority["candidate_delta"]["bundle_sha256"] = "f" * 64
        with self.assertRaisesRegex(validator.CandidateValidatorError, "capability_manifest_invalid"):
            validator.validate_capability_manifest(canonical(manifest), hashes, changed_authority)

    def test_capture_manifest_schema_is_exact_and_privacy_scanned(self) -> None:
        manifest = {key: "1" * 64 for key in validator.adapter.CAPTURE_MANIFEST_KEYS}
        manifest.update({
            "$schema": validator.adapter.CAPTURE_MANIFEST_SCHEMA,
            "status": "captured_unvalidated",
            "effect_claim": False,
            "candidate_execution_authorized": True,
            "max_capture_count": 1,
            "run_count": validator.adapter.ACCEPTED_PARENT_RUN_COUNT,
        })
        encoded = canonical(manifest)
        validator.validate_capture_manifest_bytes(encoded, ("private-token",))
        extra = dict(manifest, private_note="private-token")
        with self.assertRaisesRegex(validator.CandidateValidatorError, "candidate_output_invalid"):
            validator.validate_capture_manifest_bytes(canonical(extra), ("private-token",))
        tainted = dict(manifest, status="private-token")
        with self.assertRaisesRegex(validator.CandidateValidatorError, "candidate_output_invalid"):
            validator.validate_capture_manifest_bytes(canonical(tainted), ("private-token",))

    def test_result_layout_is_derived_from_controller_output(self) -> None:
        output = {
            "$schema": "structural-retrieval-controller-output.v4",
            "runs": [
                {"result_path": "results/01.json", "report_path": "reports/01.json"},
                {"result_path": "results/02.json", "report_path": "reports/02.json"},
            ],
            "history_path": "history.jsonl",
            "privacy_manifest_path": "privacy-scan.json",
        }
        expected = validator.expected_result_artifacts(output)
        self.assertIn("candidate/results/01.json", expected)
        self.assertIn("candidate/reports/02.json", expected)
        self.assertIn("candidate/runtime/01.stderr", expected)
        self.assertIn("candidate-capture-manifest.json", expected)

    def test_result_layout_rejects_non_string_controller_paths(self) -> None:
        output = {
            "$schema": "structural-retrieval-controller-output.v4",
            "runs": [{"result_path": 7, "report_path": "reports/01.json"}],
            "history_path": "history.jsonl",
            "privacy_manifest_path": "privacy-scan.json",
        }
        with self.assertRaisesRegex(validator.CandidateValidatorError, "candidate_output_invalid"):
            validator.expected_result_artifacts(output)

    def test_gate_payload_must_be_frozen_go_harness_output(self) -> None:
        budget = {"wall_time_millis": 1, "cpu_time_millis": 1, "max_rss_bytes": 1, "process_count": 1, "sqlite_row_count": 1, "logical_dump_bytes": 1, "stdout_bytes": 1, "stderr_bytes": 0, "case_count": 1}
        payload = {
            "$schema": validator.GATE_HARNESS_SCHEMA,
            "status": "accepted",
            "candidate_status": "discard",
            "microbench_structural_owner_recall_delta": 0.25,
            "blocking_false_structural_claim_count": 0,
            "strong_blast_radius_recall_regression": 0,
            "run_count": 6,
            "run_actuals": [
                {"order": index, "mode": "isolated" if index <= 4 else ("combined" if index == 5 else "scale"), "case_id": f"case-{index}" if index <= 4 else "", "actual_budget": budget}
                for index in range(1, 7)
            ],
            "validated_source_main_sha256": validator.ACCEPTED_MAIN_SHA256,
            "validated_source_test_sha256": validator.ACCEPTED_MAIN_TEST_SHA256,
        }
        self.assertEqual(validator.validate_gate_harness_result(canonical(payload))["candidate_status"], "discard")
        payload["candidate_status"] = "keep"
        payload["blocking_false_structural_claim_count"] = 1
        with self.assertRaisesRegex(validator.CandidateValidatorError, "gate_result_invalid"):
            validator.validate_gate_harness_result(canonical(payload))

    def test_rejected_result_is_generic_and_never_effect_claim(self) -> None:
        result = validator.rejected_result("raw private detail")
        self.assertEqual(result["status"], "rejected")
        self.assertFalse(result["effect_claim"])
        self.assertNotIn("raw private detail", canonical(result).decode())

    def test_embedded_gate_harness_compiles_from_b_and_never_executes_c(self) -> None:
        parent_root = Path(os.environ.get("C3_PROTOCOL_V7_PARENT_ROOT", str(Path.home() / ".cache/c3-v7-baseline-bb-05")))
        if not (parent_root / "freeze/source.bundle").is_file():
            self.skipTest("retained generic parent is unavailable")
        go = validator.adapter._resolve_tool(None, "go", validator.adapter.ACCEPTED_GO_EXECUTABLE_SHA256)
        git = validator.adapter._resolve_tool(None, "git", validator.adapter.ACCEPTED_GIT_EXECUTABLE_SHA256)
        with tempfile.TemporaryDirectory() as raw:
            temporary = Path(raw)
            b_root, c_root = temporary / "B", temporary / "C"
            validator.adapter._checkout_bundle(git, parent_root / "freeze/source.bundle", b_root, validator.adapter.ACCEPTED_PARENT_COMMIT)
            validator.adapter._checkout_bundle(git, parent_root / "freeze/source.bundle", c_root, validator.adapter.ACCEPTED_PARENT_COMMIT)
            marker = temporary / "candidate-init-executed"
            poison = c_root / "cli/tools/structural-search-eval-v2/candidate_poison_test.go"
            poison.write_text(f'package main\nimport "os"\nfunc init() {{ _ = os.WriteFile({json.dumps(str(marker))}, []byte("bad"), 0600) }}\n', encoding="utf-8")
            harness_root = temporary / "harness"
            validator.adapter._copy_module_tree(b_root / "cli", harness_root / "cli")
            harness = harness_root / "cli/tools/structural-search-eval-v2/protocol_v7_candidate_gate_test.go"
            harness.write_text(validator.GATE_HARNESS, encoding="utf-8")
            preflight_harness = harness_root / "cli/tools/structural-search-eval-v2/protocol_v7_candidate_preflight_test.go"
            preflight_harness.write_text(validator.adapter.PREFLIGHT_HARNESS, encoding="utf-8")
            home = Path.home()
            env = {
                "PATH": str(go.parent), "HOME": str(home), "LC_ALL": "C", "LANG": "C", "TZ": "UTC",
                "GOENV": "off", "GOWORK": "off", "GOFLAGS": "", "GOTOOLCHAIN": "local", "CGO_ENABLED": "0",
                "GOPROXY": "off", "GOSUMDB": "off", "GONOSUMDB": "", "GOPRIVATE": "",
                "GOMODCACHE": str(home / "go/pkg/mod"), "GOCACHE": str(Path(os.environ.get("XDG_CACHE_HOME", str(home / ".cache"))) / "go-build"),
            }
            completed = validator.adapter.run_bounded_process(
                [str(go), "test", "-count=1", "-run", "^$", "./tools/structural-search-eval-v2"],
                cwd=harness_root / "cli", env=env, timeout=600, output_cap=validator.adapter.PROCESS_OUTPUT_CAP,
            )
            self.assertEqual((completed.returncode, completed.stderr), (0, b""), completed.stderr.decode(errors="replace"))
            self.assertFalse(marker.exists())

    @unittest.skipUnless(os.environ.get("C3_RUN_PROTOCOL_V7_CANDIDATE_INTEGRATION") == "1", "explicit disposable integration gate")
    def test_real_disposable_prepare_and_independent_capability_replay(self) -> None:
        parent_root = Path(os.environ.get("C3_PROTOCOL_V7_PARENT_ROOT", str(Path.home() / ".cache/c3-v7-baseline-bb-05")))
        policy = Path(os.environ.get("C3_PROTOCOL_V7_POLICY", str(Path.home() / ".cache/c3-v7-baseline-bb-02.policy.json")))
        store = Path(os.environ.get("C3_PROTOCOL_V7_STORE", str(Path.cwd() / ".okra/runs/c3-rag-autoresearch-v1"))).resolve()
        go = validator.adapter._resolve_tool(None, "go", validator.adapter.ACCEPTED_GO_EXECUTABLE_SHA256)
        git = validator.adapter._resolve_tool(None, "git", validator.adapter.ACCEPTED_GIT_EXECUTABLE_SHA256)
        parent_before = validator.adapter._relative_manifest(parent_root)
        with tempfile.TemporaryDirectory() as raw:
            temporary = Path(raw)
            b_root = temporary / "B"
            validator.adapter._checkout_bundle(git, parent_root / "freeze/source.bundle", b_root, validator.adapter.ACCEPTED_PARENT_COMMIT)
            candidate_source = temporary / "candidate-search.go"
            candidate_source.write_bytes((b_root / "cli/cmd/search.go").read_bytes() + b"\n// protocol-v7 disposable adapter composition proof\n")
            candidate_source.chmod(0o600)
            capability = temporary / "capability"
            output = temporary / "future-output"
            config = validator.adapter.CandidateConfig(
                parent_root, store, validator.adapter.PARENT_VALIDATOR_REF_PREFIX + "1", policy,
                capability, output, candidate_source, go_executable=go, git_executable=git,
            )
            self.assertEqual(validator.adapter.validate_preparation_inputs(config)["status"], "inputs_valid")
            prepared = validator.adapter.prepare_candidate(config)
            self.assertEqual(prepared["status"], "accepted_unexecuted")
            replayed = validator.validate_capability(config)
            self.assertEqual(replayed["status"], "capability_accepted")
            manifest = validator.adapter.decode_canonical((capability / "capability-manifest.json").read_bytes(), "test")
            binding = validator.adapter.OneShotBinding(
                temporary / "unused-authorization", "1" * 64, "2" * 64,
                temporary / "unused-activation", "3" * 64, output,
                validator.adapter.sha256_file(capability / "capability-manifest.json"),
                validator.adapter.sha256_file(capability / "candidate-authority.v4.json"),
                manifest["candidate_delta_sha256"], validator.adapter.sha256_file(capability / "freeze/candidate-runtime"),
                validator.adapter.sha256_file(capability / "freeze/source.bundle"), manifest["privacy_policy_sha256"],
                manifest["adapter_sha256"], manifest["validator_sha256"],
            )
            snapshot_parent = temporary / "snapshot-parent"
            snapshot_parent.mkdir(mode=0o700)
            snapshot = validator.adapter.create_execution_snapshot(config, snapshot_parent)
            validator.adapter.verify_execution_snapshot_bindings(snapshot, binding, config)
            validator.adapter.release_execution_snapshot(snapshot)
            shutil.rmtree(snapshot.root)
            self.assertFalse(output.exists())
            self.assertFalse(any(path.name.endswith(".consumed") for path in temporary.rglob("*")))
            shutil.rmtree(capability)
        self.assertEqual(validator.adapter._relative_manifest(parent_root), parent_before)


if __name__ == "__main__":
    unittest.main()
