from __future__ import annotations

import hashlib
import json
import os
from pathlib import Path
import stat
import shutil
import sys
import tempfile
import time
import unittest
from unittest import mock

from scripts import structural_retrieval_candidate_protocol_v7 as candidate


def canonical(value: object) -> bytes:
    return json.dumps(value, sort_keys=True, separators=(",", ":")).encode()


class CandidateAdapterUnitTests(unittest.TestCase):
    def test_frozen_parent_and_source_hashes_are_exact(self) -> None:
        self.assertEqual(candidate.ACCEPTED_PARENT_AUTHORITY_SHA256, "5f690bf41927351d9e5f1433d15576c0948c31233f9be09d40391de05ad5f38b")
        self.assertEqual(candidate.ACCEPTED_PARENT_OUTPUT_SHA256, "e65a394eb79e2bd87bb40ee1e0247b6f5eccc56f2e3ca4563d6aad5e029b03ff")
        self.assertEqual(candidate.ACCEPTED_PARENT_COMMIT, "3bb7e9a700d1fba0469468b0566683ea51a4b115")
        self.assertEqual(candidate.ACCEPTED_PARENT_TREE, "1c5f8b614cab6a3f23d83c9ec14504cd159dadd6")
        self.assertEqual(candidate.ACCEPTED_MAIN_SHA256, "33b6452a0e34d825ef2b895609daa77b91eb129ebdc1e6be8f7f7755fc8b9c6e")
        self.assertFalse(candidate.CANDIDATE_EXECUTION_AUTHORIZED)

    def test_candidate_input_requires_private_changed_search_source(self) -> None:
        with tempfile.TemporaryDirectory() as raw:
            root = Path(raw)
            source = root / "search.go"
            source.write_bytes(b"package cmd\n")
            source.chmod(0o600)
            accepted = candidate.validate_candidate_source(source, "0" * 64, ("needle",))
            self.assertEqual(accepted.sha256, hashlib.sha256(b"package cmd\n").hexdigest())
            source.chmod(0o644)
            with self.assertRaisesRegex(candidate.CandidateAdapterError, "candidate_input_invalid"):
                candidate.validate_candidate_source(source, "0" * 64, ())

    def test_candidate_input_rejects_equal_private_or_matching_bytes(self) -> None:
        with tempfile.TemporaryDirectory() as raw:
            source = Path(raw) / "search.go"
            source.write_bytes(b"package cmd\n")
            source.chmod(0o600)
            digest = hashlib.sha256(source.read_bytes()).hexdigest()
            with self.assertRaisesRegex(candidate.CandidateAdapterError, "candidate_unchanged"):
                candidate.validate_candidate_source(source, digest, ())
            with self.assertRaisesRegex(candidate.CandidateAdapterError, "privacy_violation"):
                candidate.validate_candidate_source(source, "1" * 64, ("package",))

    def test_paths_must_be_absent_disjoint_and_outside_repo(self) -> None:
        with tempfile.TemporaryDirectory() as raw:
            root = Path(raw)
            repo = root / "repo"
            repo.mkdir()
            output = root / "out"
            capability = root / "cap"
            candidate.validate_transaction_paths(repo, capability, output, ())
            output.mkdir()
            with self.assertRaisesRegex(candidate.CandidateAdapterError, "governed_path_invalid"):
                candidate.validate_transaction_paths(repo, capability, output, ())

    def test_parent_validator_prefix_is_hash_chained_and_exact(self) -> None:
        payload = {
            "event": "finish",
            "worker_id": "validator-baseline-protocol-v7",
            "role": "independent baseline validator",
            "status": "accepted",
            "effect_claim": False,
            "baseline_acceptance": {"$schema": "structural-retrieval-baseline-acceptance.v1"},
        }
        envelope = {"seq": 1, "recorded_at": "2026-01-01T00:00:00Z", "prev_hash": "GENESIS", "payload_sha256": hashlib.sha256(canonical(payload)).hexdigest(), "payload": payload}
        envelope["record_hash"] = hashlib.sha256(canonical(envelope)).hexdigest()
        line = canonical(envelope) + b"\n"
        got = candidate.validate_worker_record_prefix(line, 1, envelope["record_hash"], envelope["payload_sha256"])
        self.assertEqual(got, payload)
        with self.assertRaisesRegex(candidate.CandidateAdapterError, "parent_validator_invalid"):
            candidate.validate_worker_record_prefix(line + line, 1, envelope["record_hash"], envelope["payload_sha256"])

    def test_candidate_delta_is_canonical_and_one_variable(self) -> None:
        delta = {
            "variable": candidate.REGISTERED_VARIABLE,
            "baseline_commit": "a" * 40,
            "baseline_tree": "b" * 40,
            "candidate_commit": "c" * 40,
            "candidate_tree": "d" * 40,
            "diff_sha256": "1" * 64,
            "name_status_sha256": "2" * 64,
            "name_status": ["M\tcli/cmd/search.go"],
            "allowed_paths": ["cli/cmd/search.go"],
            "before_blob_sha256": {"cli/cmd/search.go": "3" * 64},
            "after_blob_sha256": {"cli/cmd/search.go": "4" * 64},
            "bundle_sha256": "5" * 64,
            "bundle_heads_sha256": "6" * 64,
        }
        self.assertEqual(candidate.candidate_delta_sha256(delta), hashlib.sha256(canonical(delta)).hexdigest())
        delta["allowed_paths"] = ["cli/cmd/search.go", "extra"]
        with self.assertRaisesRegex(candidate.CandidateAdapterError, "candidate_delta_invalid"):
            candidate.candidate_delta_sha256(delta)

    def test_authorization_and_activation_are_canonical_and_bound(self) -> None:
        with tempfile.TemporaryDirectory() as raw:
            root = Path(raw)
            authorization = root / "authorization.jsonl"
            activation = root / "activation.json"
            output = root / "output"
            binding = candidate.make_test_one_shot(authorization, activation, output)
            consumed = candidate.consume_candidate_activation(activation, authorization, binding)
            self.assertFalse(activation.exists())
            self.assertEqual(consumed.path, Path(str(activation) + ".consumed"))
            self.assertEqual(stat.S_IMODE(consumed.path.stat().st_mode), 0o600)
            with self.assertRaisesRegex(candidate.CandidateAdapterError, "activation_invalid"):
                candidate.consume_candidate_activation(activation, authorization, binding)
            os.close(consumed.descriptor)

    def test_activation_swap_cannot_consume_a_decoy_or_leave_a_receipt(self) -> None:
        with tempfile.TemporaryDirectory() as raw:
            root = Path(raw)
            authorization = root / "authorization.jsonl"
            activation = root / "activation.json"
            output = root / "output"
            binding = candidate.make_test_one_shot(authorization, activation, output)
            genuine = activation.read_bytes()

            def swap() -> None:
                activation.rename(root / "genuine-away")
                activation.write_bytes(b"decoy")
                activation.chmod(0o600)

            with self.assertRaisesRegex(candidate.CandidateAdapterError, "activation_invalid"):
                candidate.consume_candidate_activation(activation, authorization, binding, transition_hook=swap)
            self.assertEqual((root / "genuine-away").read_bytes(), genuine)
            self.assertEqual(activation.read_bytes(), b"decoy")
            self.assertFalse(Path(str(activation) + ".consumed").exists())

    def test_failed_activation_unlink_removes_new_receipt(self) -> None:
        with tempfile.TemporaryDirectory() as raw:
            root = Path(raw)
            authorization = root / "authorization.jsonl"
            activation = root / "activation.json"
            binding = candidate.make_test_one_shot(authorization, activation, root / "output")
            real_unlink = os.unlink

            def fail_activation(path: object, *args: object, **kwargs: object) -> None:
                if Path(path) == activation:
                    raise OSError("forced")
                real_unlink(path, *args, **kwargs)

            with mock.patch("os.unlink", side_effect=fail_activation):
                with self.assertRaisesRegex(candidate.CandidateAdapterError, "activation_invalid"):
                    candidate.consume_candidate_activation(activation, authorization, binding)
            self.assertTrue(activation.exists())
            self.assertFalse(Path(str(activation) + ".consumed").exists())

    def test_execution_snapshot_detects_governed_source_swap(self) -> None:
        with tempfile.TemporaryDirectory() as raw:
            root = Path(raw)
            source = root / "source"
            source.mkdir(mode=0o700)
            governed = source / "runtime"
            governed.write_bytes(b"authorized")
            governed.chmod(0o600)
            snapshot_root = root / "snapshot"

            def swap() -> None:
                governed.write_bytes(b"swapped")

            with self.assertRaisesRegex(candidate.CandidateAdapterError, "governed_input_changed"):
                candidate._snapshot_registered_files(
                    snapshot_root,
                    {"capability/freeze/runtime": governed},
                    executable={"capability/freeze/runtime"},
                    transition_hook=swap,
                )
            self.assertFalse(snapshot_root.exists())

    def test_public_result_rejects_machine_paths_and_raw_output(self) -> None:
        self.assertTrue(candidate.is_generic_result({"status": "accepted", "candidate_execution_authorized": False}))
        self.assertFalse(candidate.is_generic_result({"path": "/private/raw-output"}))
        self.assertFalse(candidate.is_generic_result({"stderr": "raw child output"}))

    def test_git_commit_identity_is_repeatable(self) -> None:
        git = candidate._resolve_tool(None, "git", candidate.ACCEPTED_GIT_EXECUTABLE_SHA256)
        commits: list[str] = []
        with tempfile.TemporaryDirectory() as raw:
            for name in ("one", "two"):
                root = Path(raw) / name
                root.mkdir(mode=0o700)
                candidate._git(git, root, "init", "-q")
                source = root / "search.go"
                source.write_bytes(b"package cmd\n")
                source.chmod(0o644)
                candidate._git(git, root, "add", "search.go")
                candidate._git(git, root, "commit", "-q", "--no-gpg-sign", "-m", "fixed")
                commits.append(candidate._git(git, root, "rev-parse", "HEAD").decode().strip())
        self.assertEqual(commits[0], commits[1])

    def test_validate_inputs_replays_every_read_only_boundary(self) -> None:
        config = candidate.CandidateConfig(
            Path("/parent"), Path("/store"), candidate.PARENT_VALIDATOR_REF_PREFIX + "1",
            Path("/policy"), Path("/capability"), Path("/output"), Path("/search.go"),
        )
        parent = candidate.ParentBinding(
            Path("/parent"),
            {"controller_source_capsule": {"inputs": [{"path": "cli/cmd/search.go", "sha256": "1" * 64}]}},
            {}, Path("/store"), config.parent_validator_ref,
        )
        with (
            mock.patch.object(candidate, "validate_parent", return_value=parent) as parent_check,
            mock.patch.object(candidate, "validate_privacy_policy", return_value=("2" * 64, ())) as policy_check,
            mock.patch.object(candidate, "_resolve_tool", side_effect=[Path("/go"), Path("/git")]) as tool_check,
            mock.patch.object(candidate, "validate_candidate_source", return_value=candidate.CandidateSource(Path("/search.go"), "3" * 64, 10)) as source_check,
            mock.patch.object(candidate, "validate_transaction_paths") as path_check,
            mock.patch.object(candidate, "_governed_input_snapshot", return_value={"all": "4" * 64}) as snapshot_check,
        ):
            result = candidate.validate_preparation_inputs(config)
        self.assertEqual(result["status"], "inputs_valid")
        self.assertEqual(parent_check.call_count, 1)
        self.assertEqual(policy_check.call_count, 1)
        self.assertEqual(tool_check.call_count, 2)
        self.assertEqual(source_check.call_count, 1)
        self.assertEqual(path_check.call_count, 1)
        self.assertEqual(snapshot_check.call_count, 1)

    def test_bounded_process_enforces_output_cap_before_completion(self) -> None:
        with self.assertRaisesRegex(candidate.CandidateAdapterError, "candidate_output_cap"):
            candidate.run_bounded_process(
                [str(Path(sys.executable).resolve()), "-c", "import os,time; os.write(1,b'x'*65536); time.sleep(30)"],
                cwd=Path.cwd(), env={"PATH": "", "LC_ALL": "C"}, timeout=10, output_cap=1024,
            )

    def test_bounded_process_timeout_kills_descendant_group(self) -> None:
        with tempfile.TemporaryDirectory() as raw:
            marker = Path(raw) / "descendant-survived"
            program = (
                "import os,subprocess,sys,time; "
                "subprocess.Popen([sys.executable,'-c',"
                + repr("import pathlib,time; time.sleep(2); pathlib.Path(" + repr(str(marker)) + ").write_text('bad')")
                + "]); time.sleep(30)"
            )
            with self.assertRaisesRegex(candidate.CandidateAdapterError, "candidate_timeout"):
                candidate.run_bounded_process(
                    [str(Path(sys.executable).resolve()), "-c", program],
                    cwd=Path.cwd(), env={"PATH": "", "LC_ALL": "C"}, timeout=1, output_cap=1024,
                )
            time.sleep(2.2)
            self.assertFalse(marker.exists())

    def test_capture_environment_does_not_inherit_ambient_secret(self) -> None:
        go = candidate._resolve_tool(None, "go", candidate.ACCEPTED_GO_EXECUTABLE_SHA256)
        observed: dict[str, str] = {}
        def runner(_command: list[str], **kwargs: object) -> candidate.ProcessResult:
            observed.update(kwargs["env"])  # type: ignore[arg-type]
            return candidate.ProcessResult(0, b"", b"")
        with mock.patch.dict(os.environ, {"C3_TEST_SENTINEL_SECRET": "must-not-cross"}):
            candidate._execute_candidate_controller(
                ["/frozen-controller"], cwd=Path.cwd(), tool_root=Path("/private-tools"),
                go=go, temporary=Path("/private-tmp"), runner=runner,
            )
        self.assertNotIn("C3_TEST_SENTINEL_SECRET", observed)
        self.assertEqual(set(observed) & {"AWS_SECRET_ACCESS_KEY", "GITHUB_TOKEN", "ANTHROPIC_API_KEY"}, set())


if __name__ == "__main__":
    unittest.main()
