#!/usr/bin/env python3
"""Read-only, fail-closed validation of a captured protocol-v7 baseline."""

from __future__ import annotations

import argparse
from dataclasses import dataclass
from datetime import datetime, timezone
import hashlib
import importlib.util
import json
import os
from pathlib import Path
import re
import selectors
import signal
import stat
import subprocess
import sys
import tempfile
import time
from typing import Any, Callable


ACCEPTED_PREPARATION_VERIFIER_SHA256 = "38fce6d0991fafaf5f7b8b1d2bfaee9410451ab1034c4ec4d9fe7ad792b4bb42"
_PREPARATION_VERIFIER_PATH = Path(__file__).with_name("validate_structural_retrieval_preparation_v7.py")
if hashlib.sha256(_PREPARATION_VERIFIER_PATH.read_bytes()).hexdigest() != ACCEPTED_PREPARATION_VERIFIER_SHA256:
    raise RuntimeError("accepted preparation verifier source drift")
_PREPARATION_SPEC = importlib.util.spec_from_file_location(
    "_c3_protocol_v7_accepted_preparation_verifier",
    _PREPARATION_VERIFIER_PATH,
)
assert _PREPARATION_SPEC is not None and _PREPARATION_SPEC.loader is not None
preparation = importlib.util.module_from_spec(_PREPARATION_SPEC)
sys.modules[_PREPARATION_SPEC.name] = preparation
_PREPARATION_SPEC.loader.exec_module(preparation)


ACCEPTED_ADAPTER_SHA256 = "1c0ca65efd476211668fcd16750ead158c7f2b993efcf4845ba9a76403d98d65"
ACCEPTED_DRIVER_SHA256 = "d0d8205d20e3407048377a3e770066f0fb0013c41431f994ae6b0473532b6c45"
ACCEPTED_MAIN_SHA256 = "33b6452a0e34d825ef2b895609daa77b91eb129ebdc1e6be8f7f7755fc8b9c6e"
ACCEPTED_MAIN_TEST_SHA256 = "0830a0671aa9eb146902dc6fa56e885126b89a3312ca6c38fd8507fa10481dbf"
ACCEPTED_FIXTURE_SHA256 = "15f57120c6aa9ae07bf4fdacd6ad783afa5e70ed8ebebaff3a42dcf4249e677e"
ACCEPTED_BENCHMARK_SHA256 = "b960525cc42216e6598452946da5fb68735bbf989f311f170cedcfdbe92bf0d5"
ACCEPTED_SCORER_REGION_SHA256 = "c1669d43b13a36eb1454a01a5ea7e444fe67ce28fdae9628da083927f093cbcb"
ACCEPTED_GO_EXECUTABLE_SHA256 = preparation.ACCEPTED_GO_EXECUTABLE_SHA256
ACCEPTED_GIT_EXECUTABLE_SHA256 = preparation.ACCEPTED_GIT_EXECUTABLE_SHA256

HARNESS_SCHEMA = "structural-retrieval-protocol-v7-baseline-verifier-harness.v1"
REJECTED_SCHEMA = "structural-retrieval-protocol-v7-baseline-verifier-result.v1"
CAPTURE_SCHEMA = "structural-retrieval-protocol-v7-preparation.v1"
INPUT_MANIFEST_SCHEMA = "structural-retrieval-protocol-v7-adapter-inputs.v1"
AUTHORIZATION_SCHEMA = "structural-retrieval-protocol-v7-baseline-authorization.v1"
ACTIVATION_SCHEMA = "structural-retrieval-protocol-v7-capture-activation.v1"
ACCEPTANCE_SCHEMA = "structural-retrieval-baseline-acceptance.v1"

MAX_METADATA_BYTES = 4 << 20
MAX_POLICY_BYTES = 128 << 10
MAX_CHECKINS_BYTES = 16 << 20
MAX_SOURCE_FILE_BYTES = 64 << 20
MAX_HARNESS_RESULT_BYTES = 128 << 10
MAX_BASELINE_FILES = 2_048
MAX_BASELINE_BYTES = 268_435_456
PROCESS_OUTPUT_CAP = 4 << 20
PROCESS_TIMEOUT_SECONDS = 600

CAPTURE_MANIFEST_KEYS = {
    "$schema",
    "activation_proof_sha256",
    "activation_record_hash",
    "adapter_sha256",
    "authority_sha256",
    "authorization_record_sha256",
    "benchmark_sha256",
    "bundle_heads_sha256",
    "bundle_sha256",
    "commit",
    "controller_output_sha256",
    "driver_sha256",
    "fixture_sha256",
    "history_sha256",
    "input_manifest_sha256",
    "main_sha256",
    "main_test_sha256",
    "ordered_run_manifest_sha256",
    "privacy_manifest_sha256",
    "privacy_policy_sha256",
    "privacy_term_count",
    "run_count",
    "runtime_sha256",
    "scorer_region_sha256",
    "source_capsule_sha256",
    "status",
    "tree",
}

ACCEPTANCE_KEYS = {
    "$schema",
    "authority_sha256",
    "history_sha256",
    "history_tail_record_hash",
    "ordered_run_manifest_sha256",
    "output_sha256",
    "privacy_manifest_sha256",
    "run_count",
    "validated_source_main_sha256",
    "validated_source_test_sha256",
    "verdict",
}

SAFE_FAILURE_CLASSES = {
    "activation_invalid",
    "authority_replay_failed",
    "authorization_invalid",
    "baseline_layout_invalid",
    "baseline_replay_failed",
    "bundle_invalid",
    "frozen_input_invalid",
    "governed_path_overlap",
    "harness_command_invalid",
    "harness_execution_failed",
    "harness_interrupted",
    "harness_output_cap",
    "harness_rejected",
    "harness_result_invalid",
    "harness_source_invalid",
    "harness_timeout",
    "metadata_binding_invalid",
    "metadata_invalid",
    "privacy_policy_invalid",
    "privacy_violation",
    "source_checkout_invalid",
    "stage_residue",
    "threshold_invalid",
    "tool_identity_invalid",
    "governed_input_changed",
}


class BaselineVerifierError(ValueError):
    """A bounded generic validation failure."""


@dataclass(frozen=True)
class FrozenInputs:
    adapter_sha256: str = ACCEPTED_ADAPTER_SHA256
    driver_sha256: str = ACCEPTED_DRIVER_SHA256
    main_sha256: str = ACCEPTED_MAIN_SHA256
    main_test_sha256: str = ACCEPTED_MAIN_TEST_SHA256
    fixture_sha256: str = ACCEPTED_FIXTURE_SHA256
    benchmark_sha256: str = ACCEPTED_BENCHMARK_SHA256
    scorer_region_sha256: str = ACCEPTED_SCORER_REGION_SHA256
    preparation_verifier_sha256: str = ACCEPTED_PREPARATION_VERIFIER_SHA256
    go_executable_sha256: str = ACCEPTED_GO_EXECUTABLE_SHA256
    git_executable_sha256: str = ACCEPTED_GIT_EXECUTABLE_SHA256


@dataclass(frozen=True)
class BaselineVerifierConfig:
    source_root: Path
    baseline_root: Path
    main_path: Path
    main_test_path: Path
    fixture_path: Path
    benchmark_path: Path
    threshold_checkins_path: Path
    threshold_checkin_seq: int
    privacy_policy_path: Path
    adapter_path: Path
    preparation_verifier_path: Path
    activation_original_path: Path
    consumed_receipt_path: Path
    authorization_record_path: Path
    go_executable_path: Path
    git_executable_path: Path


ProcessResult = preparation.ProcessResult
Runner = Callable[..., ProcessResult]
canonical_json = preparation.canonical_json
_sha256_bytes = preparation._sha256_bytes
_valid_sha256 = preparation._valid_sha256
_valid_git_oid = preparation._valid_git_oid
_decode_canonical_object = preparation._decode_canonical_object
_regular_file = preparation._regular_file
_read_regular = preparation._read_regular
_directory = preparation._directory
_paths_overlap = preparation._paths_overlap
_sha256_file = preparation._sha256_file
_go_environment = preparation._go_environment
_validate_tools = preparation._validate_tools
_decode_authority_compile_closure = preparation._decode_authority_compile_closure
_source_closure_snapshot = preparation._source_closure_snapshot
_copy_compiler_closure = preparation._copy_compiler_closure
_validate_compiler_tree = preparation._validate_compiler_tree
_materialize_minimal_tool_path = preparation._materialize_minimal_tool_path


def _path_sha256(path: Path) -> str:
    return _sha256_bytes(str(path.absolute()).encode())


def _missing(path: Path) -> bool:
    try:
        path.lstat()
    except FileNotFoundError:
        return True
    except OSError as exc:
        raise BaselineVerifierError("activation_invalid") from exc
    return False


def _snapshot_baseline_tree(root: Path) -> dict[str, tuple[int, int, str]]:
    snapshot: dict[str, tuple[int, int, str]] = {}
    count = 0
    total = 0
    try:
        for path in sorted(root.rglob("*")):
            info = path.lstat()
            relative = path.relative_to(root).as_posix()
            if stat.S_ISLNK(info.st_mode):
                raise BaselineVerifierError("baseline_layout_invalid")
            if stat.S_ISDIR(info.st_mode):
                if stat.S_IMODE(info.st_mode) != 0o700:
                    raise BaselineVerifierError("baseline_layout_invalid")
                snapshot[relative + "/"] = (stat.S_IMODE(info.st_mode), 0, "")
                continue
            if not stat.S_ISREG(info.st_mode) or stat.S_IMODE(info.st_mode) != 0o600:
                raise BaselineVerifierError("baseline_layout_invalid")
            count += 1
            total += info.st_size
            if count > MAX_BASELINE_FILES or total > MAX_BASELINE_BYTES:
                raise BaselineVerifierError("baseline_layout_invalid")
            data = _read_regular(
                path,
                mode=0o600,
                cap=max(MAX_SOURCE_FILE_BYTES, 128 << 20),
                error_class="baseline_layout_invalid",
            )
            snapshot[relative] = (stat.S_IMODE(info.st_mode), len(data), _sha256_bytes(data))
    except OSError as exc:
        raise BaselineVerifierError("baseline_layout_invalid") from exc
    return snapshot


def _validate_baseline_layout(root: Path) -> tuple[dict[str, bytes], dict[str, tuple[int, int, str]]]:
    _directory(root, 0o700, "baseline_layout_invalid")
    freeze = root / "freeze"
    parent = root / "parent"
    _directory(freeze, 0o700, "baseline_layout_invalid")
    _directory(parent, 0o700, "baseline_layout_invalid")
    try:
        if {path.name for path in root.iterdir()} != {"capture-manifest.json", "freeze", "parent"}:
            raise BaselineVerifierError("baseline_layout_invalid")
        if {path.name for path in freeze.iterdir()} != {"controller-runtime", "source.bundle"}:
            raise BaselineVerifierError("baseline_layout_invalid")
        if {path.name for path in parent.iterdir()} != {
            "controller-authority.v4.json",
            "controller-output.v4.json",
            "history.jsonl",
            "privacy-scan.json",
            "reports",
            "results",
            "runtime",
        }:
            raise BaselineVerifierError("baseline_layout_invalid")
    except OSError as exc:
        raise BaselineVerifierError("baseline_layout_invalid") from exc
    snapshot = _snapshot_baseline_tree(root)
    fixed = {
        "capture-manifest.json": MAX_METADATA_BYTES,
        "freeze/source.bundle": 67_108_864,
        "freeze/controller-runtime": 128 << 20,
        "parent/controller-authority.v4.json": MAX_METADATA_BYTES,
        "parent/controller-output.v4.json": MAX_METADATA_BYTES,
        "parent/history.jsonl": 16 << 20,
        "parent/privacy-scan.json": 16 << 20,
    }
    return {
        relative: _read_regular(root / relative, mode=0o600, cap=cap, error_class="baseline_layout_invalid")
        for relative, cap in fixed.items()
    }, snapshot


def _validate_frozen_inputs(config: BaselineVerifierConfig, frozen: FrozenInputs) -> None:
    _directory(config.source_root, None, "frozen_input_invalid")
    exact = (
        (config.main_path, "cli/tools/structural-search-eval-v2/main.go", frozen.main_sha256),
        (config.main_test_path, "cli/tools/structural-search-eval-v2/main_test.go", frozen.main_test_sha256),
        (config.fixture_path, "research/eval/structural-retrieval/fixtures.dev.v2.jsonl", frozen.fixture_sha256),
        (config.benchmark_path, "research/eval/structural-retrieval/benchmark.v2.json", frozen.benchmark_sha256),
    )
    for path, relative, expected in exact:
        if path.absolute() != config.source_root.absolute() / relative:
            raise BaselineVerifierError("frozen_input_invalid")
        if _sha256_file(path, cap=MAX_SOURCE_FILE_BYTES, error_class="frozen_input_invalid") != expected:
            raise BaselineVerifierError("frozen_input_invalid")
    for path, expected in (
        (config.adapter_path, frozen.adapter_sha256),
        (config.preparation_verifier_path, frozen.preparation_verifier_sha256),
    ):
        if _sha256_file(path, cap=4 << 20, error_class="frozen_input_invalid") != expected:
            raise BaselineVerifierError("frozen_input_invalid")
    _regular_file(config.threshold_checkins_path, mode=None, cap=MAX_CHECKINS_BYTES, error_class="threshold_invalid")
    _regular_file(config.privacy_policy_path, mode=0o600, cap=MAX_POLICY_BYTES, error_class="privacy_policy_invalid")
    _regular_file(config.consumed_receipt_path, mode=0o600, cap=64 << 10, error_class="activation_invalid")
    _regular_file(config.authorization_record_path, mode=0o600, cap=64 << 10, error_class="authorization_invalid")
    if config.threshold_checkin_seq <= 0:
        raise BaselineVerifierError("threshold_invalid")
    if config.consumed_receipt_path.absolute() != Path(str(config.activation_original_path.absolute()) + ".consumed"):
        raise BaselineVerifierError("activation_invalid")
    if not _missing(config.activation_original_path):
        raise BaselineVerifierError("activation_invalid")
    governed_pairs = (
        (config.baseline_root, config.source_root),
        (config.baseline_root, config.privacy_policy_path),
        (config.baseline_root, config.authorization_record_path),
        (config.baseline_root, config.consumed_receipt_path),
        (config.baseline_root, config.threshold_checkins_path),
        (config.baseline_root, config.adapter_path),
        (config.baseline_root, config.preparation_verifier_path),
        (config.source_root, config.privacy_policy_path),
        (config.source_root, config.authorization_record_path),
        (config.source_root, config.consumed_receipt_path),
        (config.privacy_policy_path, config.authorization_record_path),
        (config.privacy_policy_path, config.consumed_receipt_path),
        (config.authorization_record_path, config.consumed_receipt_path),
        (config.baseline_root, config.go_executable_path),
        (config.baseline_root, config.git_executable_path),
        (config.source_root, config.go_executable_path),
        (config.source_root, config.git_executable_path),
    )
    if any(_paths_overlap(left, right) for left, right in governed_pairs):
        raise BaselineVerifierError("governed_path_overlap")
    try:
        for path in config.baseline_root.parent.iterdir():
            if path.name.startswith(".c3-v7-adapter-"):
                raise BaselineVerifierError("stage_residue")
    except OSError as exc:
        raise BaselineVerifierError("stage_residue") from exc


def _validate_capture_manifest(
    data: bytes,
    files: dict[str, bytes],
    snapshot: dict[str, tuple[int, int, str]],
    config: BaselineVerifierConfig,
    frozen: FrozenInputs,
) -> dict[str, Any]:
    value = _decode_canonical_object(data, CAPTURE_MANIFEST_KEYS, "metadata_invalid")
    if (
        value.get("$schema") != CAPTURE_SCHEMA
        or value.get("status") != "accepted"
        or value.get("adapter_sha256") != frozen.adapter_sha256
        or value.get("driver_sha256") != frozen.driver_sha256
        or value.get("main_sha256") != frozen.main_sha256
        or value.get("main_test_sha256") != frozen.main_test_sha256
        or value.get("fixture_sha256") != frozen.fixture_sha256
        or value.get("benchmark_sha256") != frozen.benchmark_sha256
        or value.get("scorer_region_sha256") != frozen.scorer_region_sha256
        or not _valid_git_oid(value.get("commit"))
        or not _valid_git_oid(value.get("tree"))
        or not isinstance(value.get("run_count"), int)
        or isinstance(value.get("run_count"), bool)
        or value.get("run_count") != 6
        or not isinstance(value.get("privacy_term_count"), int)
        or isinstance(value.get("privacy_term_count"), bool)
        or not 0 < value["privacy_term_count"] <= 512
    ):
        raise BaselineVerifierError("metadata_invalid")
    for key in CAPTURE_MANIFEST_KEYS - {
        "$schema", "status", "privacy_term_count", "run_count", "commit", "tree"
    }:
        if key.endswith("sha256") or key.endswith("hash"):
            if not _valid_sha256(value.get(key)):
                raise BaselineVerifierError("metadata_invalid")
    bindings = {
        "authority_sha256": _sha256_bytes(files["parent/controller-authority.v4.json"]),
        "bundle_sha256": _sha256_bytes(files["freeze/source.bundle"]),
        "controller_output_sha256": _sha256_bytes(files["parent/controller-output.v4.json"]),
        "history_sha256": _sha256_bytes(files["parent/history.jsonl"]),
        "privacy_manifest_sha256": _sha256_bytes(files["parent/privacy-scan.json"]),
        "runtime_sha256": _sha256_bytes(files["freeze/controller-runtime"]),
        "privacy_policy_sha256": _sha256_file(
            config.privacy_policy_path, cap=MAX_POLICY_BYTES, error_class="privacy_policy_invalid"
        ),
    }
    if any(value[key] != observed for key, observed in bindings.items()):
        raise BaselineVerifierError("metadata_binding_invalid")
    input_manifest = {
        "$schema": INPUT_MANIFEST_SCHEMA,
        "adapter_sha256": frozen.adapter_sha256,
        "driver_sha256": frozen.driver_sha256,
        "input_sha256": {
            "benchmark": frozen.benchmark_sha256,
            "fixture": frozen.fixture_sha256,
            "main": frozen.main_sha256,
            "main_test": frozen.main_test_sha256,
        },
        "privacy_policy_sha256": value["privacy_policy_sha256"],
        "privacy_term_count": value["privacy_term_count"],
        "scorer_region_sha256": frozen.scorer_region_sha256,
        "status": "accepted",
        "threshold_checkin_seq": config.threshold_checkin_seq,
    }
    if value["input_manifest_sha256"] != _sha256_bytes(canonical_json(input_manifest)):
        raise BaselineVerifierError("metadata_binding_invalid")
    file_count = sum(1 for relative in snapshot if not relative.endswith("/"))
    if file_count != 7 + 3 * value["run_count"]:
        raise BaselineVerifierError("baseline_layout_invalid")
    return value


def _validate_one_shot(config: BaselineVerifierConfig, manifest: dict[str, Any]) -> dict[str, str]:
    authorization_bytes = _read_regular(
        config.authorization_record_path,
        mode=0o600,
        cap=64 << 10,
        error_class="authorization_invalid",
    )
    if not authorization_bytes.endswith(b"\n") or authorization_bytes.count(b"\n") != 1:
        raise BaselineVerifierError("authorization_invalid")
    record = _decode_canonical_object(
        authorization_bytes[:-1],
        {"payload", "payload_sha256", "prev_hash", "record_hash", "recorded_at", "seq"},
        "authorization_invalid",
    )
    expected_payload = {
        "$schema": AUTHORIZATION_SCHEMA,
        "activation_path_sha256": _path_sha256(config.activation_original_path),
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
        "output_root_sha256": _path_sha256(config.baseline_root),
        "privacy_policy_sha256": manifest["privacy_policy_sha256"],
        "scorer_region_sha256": manifest["scorer_region_sha256"],
        "verdict": "authorized",
    }
    recorded_at = record.get("recorded_at")
    try:
        parsed = datetime.strptime(recorded_at, "%Y-%m-%dT%H:%M:%SZ").replace(tzinfo=timezone.utc)
    except (TypeError, ValueError) as exc:
        raise BaselineVerifierError("authorization_invalid") from exc
    if (
        parsed.strftime("%Y-%m-%dT%H:%M:%SZ") != recorded_at
        or record.get("seq") != 1
        or record.get("prev_hash") != "GENESIS"
        or record.get("payload") != expected_payload
        or record.get("payload_sha256") != _sha256_bytes(canonical_json(expected_payload))
    ):
        raise BaselineVerifierError("authorization_invalid")
    without_hash = {key: value for key, value in record.items() if key != "record_hash"}
    record_hash = _sha256_bytes(canonical_json(without_hash))
    if record.get("record_hash") != record_hash or manifest["activation_record_hash"] != record_hash:
        raise BaselineVerifierError("authorization_invalid")
    authorization_sha256 = _sha256_bytes(authorization_bytes)
    if manifest["authorization_record_sha256"] != authorization_sha256:
        raise BaselineVerifierError("authorization_invalid")

    receipt_bytes = _read_regular(
        config.consumed_receipt_path,
        mode=0o600,
        cap=64 << 10,
        error_class="activation_invalid",
    )
    activation = _decode_canonical_object(
        receipt_bytes,
        {
            "$schema", "activation_path_sha256", "activation_record_hash", "adapter_sha256",
            "authorization_record_path_sha256", "benchmark_sha256", "driver_sha256", "fixture_sha256",
            "main_sha256", "main_test_sha256", "max_capture_count", "output_root_sha256",
            "privacy_policy_sha256", "scorer_region_sha256", "verdict",
        },
        "activation_invalid",
    )
    expected_activation = {
        "$schema": ACTIVATION_SCHEMA,
        "activation_path_sha256": _path_sha256(config.activation_original_path),
        "activation_record_hash": record_hash,
        "adapter_sha256": manifest["adapter_sha256"],
        "authorization_record_path_sha256": _path_sha256(config.authorization_record_path),
        "benchmark_sha256": manifest["benchmark_sha256"],
        "driver_sha256": manifest["driver_sha256"],
        "fixture_sha256": manifest["fixture_sha256"],
        "main_sha256": manifest["main_sha256"],
        "main_test_sha256": manifest["main_test_sha256"],
        "max_capture_count": 1,
        "output_root_sha256": _path_sha256(config.baseline_root),
        "privacy_policy_sha256": manifest["privacy_policy_sha256"],
        "scorer_region_sha256": manifest["scorer_region_sha256"],
        "verdict": "authorized",
    }
    activation_sha256 = _sha256_bytes(receipt_bytes)
    if activation != expected_activation or manifest["activation_proof_sha256"] != activation_sha256:
        raise BaselineVerifierError("activation_invalid")
    return {
        "activation_proof_sha256": activation_sha256,
        "activation_record_hash": record_hash,
        "authorization_record_sha256": authorization_sha256,
    }


def harness_command(go_executable: Path) -> list[str]:
    return [
        str(go_executable), "test", "-count=1", "-run", "^TestProtocolV7BaselineVerifier$",
        "./tools/structural-search-eval-v2",
    ]


def require_allowed_command(command: list[str]) -> None:
    if not command or not Path(command[0]).is_absolute() or command != harness_command(Path(command[0])):
        raise BaselineVerifierError("harness_command_invalid")


def run_bounded_process(
    command: list[str],
    *,
    cwd: Path,
    env: dict[str, str],
    timeout: int,
    output_cap: int,
) -> ProcessResult:
    require_allowed_command(command)
    try:
        process = subprocess.Popen(
            command,
            cwd=cwd,
            env=env,
            stdin=subprocess.DEVNULL,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            start_new_session=True,
        )
    except OSError as exc:
        raise BaselineVerifierError("harness_execution_failed") from exc
    assert process.stdout is not None and process.stderr is not None
    selector = selectors.DefaultSelector()
    selector.register(process.stdout, selectors.EVENT_READ, "stdout")
    selector.register(process.stderr, selectors.EVENT_READ, "stderr")
    output = {"stdout": bytearray(), "stderr": bytearray()}
    started = time.monotonic()
    try:
        while selector.get_map():
            remaining = timeout - (time.monotonic() - started)
            if remaining <= 0:
                raise BaselineVerifierError("harness_timeout")
            events = selector.select(min(remaining, 0.25))
            if not events and process.poll() is not None:
                events = [(key, selectors.EVENT_READ) for key in tuple(selector.get_map().values())]
            for key, _ in events:
                chunk = os.read(key.fileobj.fileno(), 64 << 10)
                if not chunk:
                    selector.unregister(key.fileobj)
                    continue
                output[key.data].extend(chunk)
                if len(output["stdout"]) + len(output["stderr"]) > output_cap:
                    raise BaselineVerifierError("harness_output_cap")
        returncode = process.wait(timeout=max(1, int(timeout - (time.monotonic() - started))))
        return ProcessResult(returncode, bytes(output["stdout"]), bytes(output["stderr"]))
    except BaseException:
        preparation._kill_process(process)
        raise
    finally:
        selector.close()
        process.stdout.close()
        process.stderr.close()


def _harness_config(
    config: BaselineVerifierConfig,
    frozen: FrozenInputs,
    manifest: dict[str, Any],
    compiled_main: Path,
    compiled_main_test: Path,
    result_path: Path,
    work_root: Path,
) -> dict[str, Any]:
    return {
        "$schema": "structural-retrieval-protocol-v7-baseline-verifier-config.v1",
        "activation_original_path": str(config.activation_original_path.absolute()),
        "adapter_path": str(config.adapter_path.absolute()),
        "adapter_sha256": frozen.adapter_sha256,
        "authorization_record_path": str(config.authorization_record_path.absolute()),
        "baseline_root": str(config.baseline_root.absolute()),
        "benchmark_path": str(config.benchmark_path.absolute()),
        "benchmark_sha256": frozen.benchmark_sha256,
        "bundle_path": str((config.baseline_root / "freeze/source.bundle").absolute()),
        "consumed_receipt_path": str(config.consumed_receipt_path.absolute()),
        "driver_sha256": frozen.driver_sha256,
        "expected_manifest_sha256": _sha256_bytes(canonical_json(manifest)),
        "fixture_path": str(config.fixture_path.absolute()),
        "fixture_sha256": frozen.fixture_sha256,
        "main_path": str(compiled_main.absolute()),
        "main_sha256": frozen.main_sha256,
        "main_test_path": str(compiled_main_test.absolute()),
        "main_test_sha256": frozen.main_test_sha256,
        "manifest_path": str((config.baseline_root / "capture-manifest.json").absolute()),
        "policy_path": str(config.privacy_policy_path.absolute()),
        "result_path": str(result_path),
        "runtime_path": str((config.baseline_root / "freeze/controller-runtime").absolute()),
        "scorer_region_sha256": frozen.scorer_region_sha256,
        "threshold_checkin_seq": config.threshold_checkin_seq,
        "threshold_checkins_path": str(config.threshold_checkins_path.absolute()),
        "work_root": str(work_root),
    }


def _validate_harness_result(data: bytes, manifest: dict[str, Any], frozen: FrozenInputs) -> dict[str, Any]:
    preliminary = _decode_canonical_object(
        data,
        {"$schema", "baseline_acceptance", "status"} if b'"status":"accepted"' in data else {"$schema", "failure_class", "status"},
        "harness_result_invalid",
    )
    if preliminary.get("$schema") != HARNESS_SCHEMA:
        raise BaselineVerifierError("harness_result_invalid")
    if preliminary.get("status") == "rejected":
        if preliminary.get("failure_class") not in SAFE_FAILURE_CLASSES:
            raise BaselineVerifierError("harness_result_invalid")
        raise BaselineVerifierError(preliminary["failure_class"])
    if preliminary.get("status") != "accepted" or not isinstance(preliminary.get("baseline_acceptance"), dict):
        raise BaselineVerifierError("harness_result_invalid")
    acceptance = preliminary["baseline_acceptance"]
    if set(acceptance) != ACCEPTANCE_KEYS:
        raise BaselineVerifierError("harness_result_invalid")
    for key in ACCEPTANCE_KEYS - {"$schema", "verdict", "run_count"}:
        if not _valid_sha256(acceptance.get(key)):
            raise BaselineVerifierError("harness_result_invalid")
    if (
        acceptance.get("$schema") != ACCEPTANCE_SCHEMA
        or acceptance.get("verdict") != "accepted"
        or acceptance.get("run_count") != manifest["run_count"]
        or acceptance.get("authority_sha256") != manifest["authority_sha256"]
        or acceptance.get("output_sha256") != manifest["controller_output_sha256"]
        or acceptance.get("ordered_run_manifest_sha256") != manifest["ordered_run_manifest_sha256"]
        or acceptance.get("history_sha256") != manifest["history_sha256"]
        or acceptance.get("privacy_manifest_sha256") != manifest["privacy_manifest_sha256"]
        or acceptance.get("validated_source_main_sha256") != frozen.main_sha256
        or acceptance.get("validated_source_test_sha256") != frozen.main_test_sha256
    ):
        raise BaselineVerifierError("harness_result_invalid")
    return {
        "baseline_acceptance": acceptance,
        "effect_claim": False,
        "event": "finish",
        "role": "independent baseline validator",
        "status": "accepted",
        "worker_id": "validator-baseline-protocol-v7",
    }


def _rejected_result(error_class: str) -> dict[str, Any]:
    safe = error_class if error_class in SAFE_FAILURE_CLASSES else "verification_internal_error"
    return {"$schema": REJECTED_SCHEMA, "failure_class": safe, "status": "rejected"}


def _governed_external_snapshot(config: BaselineVerifierConfig) -> dict[str, str]:
    return {
        "policy": _sha256_file(config.privacy_policy_path, cap=MAX_POLICY_BYTES, error_class="privacy_policy_invalid"),
        "authorization": _sha256_file(config.authorization_record_path, cap=64 << 10, error_class="authorization_invalid"),
        "receipt": _sha256_file(config.consumed_receipt_path, cap=64 << 10, error_class="activation_invalid"),
        "adapter": _sha256_file(config.adapter_path, cap=4 << 20, error_class="frozen_input_invalid"),
        "preparation_verifier": _sha256_file(config.preparation_verifier_path, cap=4 << 20, error_class="frozen_input_invalid"),
        "threshold": _sha256_file(config.threshold_checkins_path, cap=MAX_CHECKINS_BYTES, error_class="threshold_invalid"),
        "go": _sha256_file(config.go_executable_path, cap=256 << 20, error_class="tool_identity_invalid"),
        "git": _sha256_file(config.git_executable_path, cap=256 << 20, error_class="tool_identity_invalid"),
    }


def verify_baseline(
    config: BaselineVerifierConfig,
    *,
    frozen: FrozenInputs = FrozenInputs(),
    runner: Runner = run_bounded_process,
) -> dict[str, Any]:
    try:
        files, before_tree = _validate_baseline_layout(config.baseline_root)
        _validate_frozen_inputs(config, frozen)
        tools = _validate_tools(
            config.go_executable_path,
            config.git_executable_path,
            frozen.go_executable_sha256,
            frozen.git_executable_sha256,
        )
        manifest = _validate_capture_manifest(
            files["capture-manifest.json"], files, before_tree, config, frozen
        )
        _validate_one_shot(config, manifest)
        closure = _decode_authority_compile_closure(files["parent/controller-authority.v4.json"])
        before_source = _source_closure_snapshot(
            config.source_root, closure, config.main_test_path, frozen.main_test_sha256
        )
        before_external = _governed_external_snapshot(config)
        with preparation._scoped_termination_handler():
            with tempfile.TemporaryDirectory(prefix="c3-v7-baseline-verifier-") as temporary:
                temporary_root = Path(temporary)
                temporary_root.chmod(0o700)
                private_tmp = temporary_root / "tmp"
                private_tmp.mkdir(mode=0o700)
                minimal_tools = _materialize_minimal_tool_path(temporary_root / "tools", tools)
                driver_root = temporary_root / "driver-source"
                driver_root.mkdir(mode=0o700)
                compiler_snapshot = _copy_compiler_closure(
                    config.source_root,
                    driver_root,
                    closure,
                    config.main_test_path,
                    frozen.main_test_sha256,
                )
                compiled_main = driver_root / "cli/tools/structural-search-eval-v2/main.go"
                compiled_main_test = driver_root / "cli/tools/structural-search-eval-v2/main_test.go"
                if (
                    _sha256_file(compiled_main, cap=MAX_SOURCE_FILE_BYTES, error_class="frozen_input_invalid") != frozen.main_sha256
                    or _sha256_file(compiled_main_test, cap=MAX_SOURCE_FILE_BYTES, error_class="frozen_input_invalid") != frozen.main_test_sha256
                ):
                    raise BaselineVerifierError("frozen_input_invalid")
                harness_path = compiled_main.with_name("protocol_v7_baseline_verifier_test.go")
                harness_path.write_text(GO_HARNESS_SOURCE, encoding="utf-8")
                harness_path.chmod(0o600)
                compiler_snapshot[
                    "cli/tools/structural-search-eval-v2/protocol_v7_baseline_verifier_test.go"
                ] = _sha256_bytes(GO_HARNESS_SOURCE.encode())
                _validate_compiler_tree(driver_root, compiler_snapshot)
                result_path = temporary_root / "harness-result.json"
                work_root = temporary_root / "work"
                work_root.mkdir(mode=0o700)
                harness_config = _harness_config(
                    config, frozen, manifest, compiled_main, compiled_main_test, result_path, work_root
                )
                config_path = temporary_root / "harness-config.json"
                config_path.write_bytes(canonical_json(harness_config))
                config_path.chmod(0o600)
                environment = _go_environment(private_tmp, minimal_tools)
                environment["C3_PROTOCOL_V7_BASELINE_VERIFIER_CONFIG"] = str(config_path)
                command = harness_command(tools.go.path)
                require_allowed_command(command)
                try:
                    completed = runner(
                        command,
                        cwd=driver_root / "cli",
                        env=environment,
                        timeout=PROCESS_TIMEOUT_SECONDS,
                        output_cap=PROCESS_OUTPUT_CAP,
                    )
                except KeyboardInterrupt as exc:
                    raise BaselineVerifierError("harness_interrupted") from exc
                if completed.returncode != 0 or completed.stderr:
                    raise BaselineVerifierError("harness_rejected")
                _validate_compiler_tree(driver_root, compiler_snapshot)
                for identity in (tools.go, tools.git):
                    if _sha256_file(
                        minimal_tools / identity.role,
                        cap=256 << 20,
                        error_class="tool_identity_invalid",
                    ) != identity.sha256:
                        raise BaselineVerifierError("tool_identity_invalid")
                result_bytes = _read_regular(
                    result_path,
                    mode=0o600,
                    cap=MAX_HARNESS_RESULT_BYTES,
                    error_class="harness_result_invalid",
                )
                accepted = _validate_harness_result(result_bytes, manifest, frozen)
        after_tree = _snapshot_baseline_tree(config.baseline_root)
        after_source = _source_closure_snapshot(
            config.source_root, closure, config.main_test_path, frozen.main_test_sha256
        )
        after_external = _governed_external_snapshot(config)
        _validate_tools(
            config.go_executable_path,
            config.git_executable_path,
            frozen.go_executable_sha256,
            frozen.git_executable_sha256,
        )
        if before_tree != after_tree or before_external != after_external or before_source != after_source:
            raise BaselineVerifierError("governed_input_changed")
        _validate_frozen_inputs(config, frozen)
        return accepted
    except preparation._VerifierInterrupted:
        return _rejected_result("harness_interrupted")
    except (BaselineVerifierError, preparation.VerifierError) as exc:
        return _rejected_result(str(exc))
    except (OSError, subprocess.SubprocessError, ValueError, TypeError):
        return _rejected_result("verification_internal_error")


GO_HARNESS_SOURCE = r'''package main

import (
    "bufio"
    "bytes"
    "encoding/json"
    "errors"
    "os"
    "path/filepath"
    "testing"
    "time"
)

const protocolV7BaselineVerifierHarnessSchema = "structural-retrieval-protocol-v7-baseline-verifier-harness.v1"

type protocolV7BaselineVerifierConfig struct {
    Schema string `json:"$schema"`
    ActivationOriginalPath string `json:"activation_original_path"`
    AdapterPath string `json:"adapter_path"`
    AdapterSHA256 string `json:"adapter_sha256"`
    AuthorizationRecordPath string `json:"authorization_record_path"`
    BaselineRoot string `json:"baseline_root"`
    BenchmarkPath string `json:"benchmark_path"`
    BenchmarkSHA256 string `json:"benchmark_sha256"`
    BundlePath string `json:"bundle_path"`
    ConsumedReceiptPath string `json:"consumed_receipt_path"`
    DriverSHA256 string `json:"driver_sha256"`
    ExpectedManifestSHA256 string `json:"expected_manifest_sha256"`
    FixturePath string `json:"fixture_path"`
    FixtureSHA256 string `json:"fixture_sha256"`
    MainPath string `json:"main_path"`
    MainSHA256 string `json:"main_sha256"`
    MainTestPath string `json:"main_test_path"`
    MainTestSHA256 string `json:"main_test_sha256"`
    ManifestPath string `json:"manifest_path"`
    PolicyPath string `json:"policy_path"`
    ResultPath string `json:"result_path"`
    RuntimePath string `json:"runtime_path"`
    ScorerRegionSHA256 string `json:"scorer_region_sha256"`
    ThresholdCheckinSeq int `json:"threshold_checkin_seq"`
    ThresholdCheckinsPath string `json:"threshold_checkins_path"`
    WorkRoot string `json:"work_root"`
}

type protocolV7BaselineCaptureManifest struct {
    Schema string `json:"$schema"`
    ActivationProofSHA256 string `json:"activation_proof_sha256"`
    ActivationRecordHash string `json:"activation_record_hash"`
    AdapterSHA256 string `json:"adapter_sha256"`
    AuthoritySHA256 string `json:"authority_sha256"`
    AuthorizationRecordSHA256 string `json:"authorization_record_sha256"`
    BenchmarkSHA256 string `json:"benchmark_sha256"`
    BundleHeadsSHA256 string `json:"bundle_heads_sha256"`
    BundleSHA256 string `json:"bundle_sha256"`
    Commit string `json:"commit"`
    ControllerOutputSHA256 string `json:"controller_output_sha256"`
    DriverSHA256 string `json:"driver_sha256"`
    FixtureSHA256 string `json:"fixture_sha256"`
    HistorySHA256 string `json:"history_sha256"`
    InputManifestSHA256 string `json:"input_manifest_sha256"`
    MainSHA256 string `json:"main_sha256"`
    MainTestSHA256 string `json:"main_test_sha256"`
    OrderedRunManifestSHA256 string `json:"ordered_run_manifest_sha256"`
    PrivacyManifestSHA256 string `json:"privacy_manifest_sha256"`
    PrivacyPolicySHA256 string `json:"privacy_policy_sha256"`
    PrivacyTermCount int `json:"privacy_term_count"`
    RunCount int `json:"run_count"`
    RuntimeSHA256 string `json:"runtime_sha256"`
    ScorerRegionSHA256 string `json:"scorer_region_sha256"`
    SourceCapsuleSHA256 string `json:"source_capsule_sha256"`
    Status string `json:"status"`
    Tree string `json:"tree"`
}

type protocolV7BaselineAuthorizationRecord struct {
    Seq int `json:"seq"`
    RecordedAt string `json:"recorded_at"`
    PrevHash string `json:"prev_hash"`
    PayloadSHA256 string `json:"payload_sha256"`
    Payload json.RawMessage `json:"payload"`
    RecordHash string `json:"record_hash"`
}

func TestProtocolV7BaselineVerifier(t *testing.T) {
    configPath := os.Getenv("C3_PROTOCOL_V7_BASELINE_VERIFIER_CONFIG")
    configBytes, err := readBoundedStandaloneRegularFile(configPath, 128<<10)
    if err != nil { t.Fatal("harness_config_invalid") }
    var cfg protocolV7BaselineVerifierConfig
    if decodeStrictBytes(configBytes, &cfg) != nil || cfg.Schema != "structural-retrieval-protocol-v7-baseline-verifier-config.v1" {
        t.Fatal("harness_config_invalid")
    }
    result := protocolV7BaselineVerifierRun(cfg)
    data, err := json.Marshal(result)
    if err != nil || os.WriteFile(cfg.ResultPath, data, 0o600) != nil { t.Fatal("harness_result_invalid") }
}

func protocolV7BaselineVerifierReject(class string) map[string]any {
    return map[string]any{"$schema": protocolV7BaselineVerifierHarnessSchema, "failure_class": class, "status": "rejected"}
}

func protocolV7BaselineVerifierPrivate(path string, max int64) ([]byte, error) {
    data, err := readBoundedStandaloneRegularFile(path, max)
    info, statErr := os.Lstat(path)
    if err != nil || statErr != nil || !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 || info.Mode().Perm() != 0o600 {
        return nil, errors.New("invalid private evidence")
    }
    return data, nil
}

func protocolV7BaselineVerifierCheckinLine(path string, seq int) ([]byte, error) {
    file, err := os.Open(path)
    if err != nil { return nil, err }
    defer file.Close()
    scanner := bufio.NewScanner(file)
    scanner.Buffer(make([]byte, 64<<10), 4<<20)
    matches := [][]byte{}
    for scanner.Scan() {
        line := append([]byte(nil), scanner.Bytes()...)
        var envelope struct { Seq int `json:"seq"` }
        if rejectDuplicateJSONKeys(line) != nil || json.Unmarshal(line, &envelope) != nil { return nil, errors.New("invalid checkin") }
        if envelope.Seq == seq { matches = append(matches, append(line, '\n')) }
    }
    if scanner.Err() != nil || len(matches) != 1 { return nil, errors.New("invalid checkin") }
    return matches[0], nil
}

func protocolV7BaselineVerifierPrivacySelfTest(policy privacyPolicy) error {
    detector, err := newGenericPrivacyDetector()
    if err != nil { return err }
    _, positives, negatives := privacyDetectorDefinition()
    for _, value := range positives { if detector.Match([]byte(value)) == "" { return errors.New("positive missed") } }
    for _, value := range negatives { if detector.Match([]byte(value)) != "" { return errors.New("negative matched") } }
    for _, term := range policy.DenyTerms {
        scanner, err := newPrivacyScanner(policy)
        if err != nil || scanner.Scan("self_test", "term", []byte(term)) == nil { return errors.New("term missed") }
    }
    return nil
}

func protocolV7BaselineVerifierOneShot(cfg protocolV7BaselineVerifierConfig, manifest protocolV7BaselineCaptureManifest, policy privacyPolicy, scanner *privacyScanner) error {
    if _, err := os.Lstat(cfg.ActivationOriginalPath); err == nil || !errors.Is(err, os.ErrNotExist) {
        return errors.New("activation_invalid")
    }
    authorizationBytes, err := protocolV7BaselineVerifierPrivate(cfg.AuthorizationRecordPath, 64<<10)
    if err != nil || len(authorizationBytes) < 2 || authorizationBytes[len(authorizationBytes)-1] != '\n' || bytes.Count(authorizationBytes, []byte{'\n'}) != 1 {
        return errors.New("authorization_invalid")
    }
    var record protocolV7BaselineAuthorizationRecord
    if decodeStrictBytes(authorizationBytes[:len(authorizationBytes)-1], &record) != nil { return errors.New("authorization_invalid") }
    parsed, err := time.Parse(time.RFC3339, record.RecordedAt)
    if err != nil || parsed.UTC().Format("2006-01-02T15:04:05Z") != record.RecordedAt || record.Seq != 1 || record.PrevHash != "GENESIS" {
        return errors.New("authorization_invalid")
    }
    expectedPayload := map[string]any{
        "$schema": "structural-retrieval-protocol-v7-baseline-authorization.v1",
        "activation_path_sha256": shaString(cfg.ActivationOriginalPath), "adapter_sha256": cfg.AdapterSHA256,
        "baseline_capture_authorized": true, "benchmark_sha256": cfg.BenchmarkSHA256,
        "candidate_execution_authorized": false, "driver_sha256": cfg.DriverSHA256, "effect_claim": false,
        "fixture_sha256": cfg.FixtureSHA256, "main_sha256": cfg.MainSHA256, "main_test_sha256": cfg.MainTestSHA256,
        "max_capture_count": 1, "output_root_sha256": shaString(cfg.BaselineRoot),
        "privacy_policy_sha256": policy.SHA256, "scorer_region_sha256": cfg.ScorerRegionSHA256, "verdict": "authorized",
    }
    expectedPayloadBytes, _ := json.Marshal(expectedPayload)
    if !bytes.Equal(record.Payload, expectedPayloadBytes) || record.PayloadSHA256 != shaString(string(expectedPayloadBytes)) {
        return errors.New("authorization_invalid")
    }
    withoutHash := map[string]any{"payload": json.RawMessage(expectedPayloadBytes), "payload_sha256": record.PayloadSHA256, "prev_hash": record.PrevHash, "recorded_at": record.RecordedAt, "seq": record.Seq}
    withoutHashBytes, _ := json.Marshal(withoutHash)
    if record.RecordHash != shaString(string(withoutHashBytes)) || record.RecordHash != manifest.ActivationRecordHash || shaString(string(authorizationBytes)) != manifest.AuthorizationRecordSHA256 {
        return errors.New("authorization_invalid")
    }
    expectedRecord := map[string]any{"payload": json.RawMessage(expectedPayloadBytes), "payload_sha256": record.PayloadSHA256, "prev_hash": record.PrevHash, "record_hash": record.RecordHash, "recorded_at": record.RecordedAt, "seq": record.Seq}
    expectedRecordBytes, _ := json.Marshal(expectedRecord)
    expectedRecordBytes = append(expectedRecordBytes, '\n')
    if !bytes.Equal(expectedRecordBytes, authorizationBytes) { return errors.New("authorization_invalid") }
    receiptBytes, err := protocolV7BaselineVerifierPrivate(cfg.ConsumedReceiptPath, 64<<10)
    if err != nil { return errors.New("activation_invalid") }
    expectedActivation := map[string]any{
        "$schema": "structural-retrieval-protocol-v7-capture-activation.v1", "activation_path_sha256": shaString(cfg.ActivationOriginalPath),
        "activation_record_hash": record.RecordHash, "adapter_sha256": cfg.AdapterSHA256,
        "authorization_record_path_sha256": shaString(cfg.AuthorizationRecordPath), "benchmark_sha256": cfg.BenchmarkSHA256,
        "driver_sha256": cfg.DriverSHA256, "fixture_sha256": cfg.FixtureSHA256, "main_sha256": cfg.MainSHA256,
        "main_test_sha256": cfg.MainTestSHA256, "max_capture_count": 1, "output_root_sha256": shaString(cfg.BaselineRoot),
        "privacy_policy_sha256": policy.SHA256, "scorer_region_sha256": cfg.ScorerRegionSHA256, "verdict": "authorized",
    }
    expectedActivationBytes, _ := json.Marshal(expectedActivation)
    if !bytes.Equal(expectedActivationBytes, receiptBytes) || shaString(string(receiptBytes)) != manifest.ActivationProofSHA256 {
        return errors.New("activation_invalid")
    }
    if scanner.Scan("capture_authorization", "authorization.jsonl", authorizationBytes) != nil || scanner.Scan("capture_activation", "activation.json", receiptBytes) != nil {
        return errors.New("privacy_violation")
    }
    return nil
}

func protocolV7BaselineVerifierRun(cfg protocolV7BaselineVerifierConfig) map[string]any {
    reject := func(class string) map[string]any { return protocolV7BaselineVerifierReject(class) }
    exact := []struct{ path, hash string }{
        {cfg.AdapterPath, cfg.AdapterSHA256}, {cfg.MainPath, cfg.MainSHA256}, {cfg.MainTestPath, cfg.MainTestSHA256},
        {cfg.FixturePath, cfg.FixtureSHA256}, {cfg.BenchmarkPath, cfg.BenchmarkSHA256},
    }
    for _, input := range exact {
        got, err := fileSHA256(input.path)
        if err != nil || got != input.hash { return reject("frozen_input_invalid") }
    }
    region, err := scoringRegionSHA256(cfg.MainPath)
    if err != nil || region != cfg.ScorerRegionSHA256 { return reject("frozen_input_invalid") }
    policyBytes, err := protocolV7BaselineVerifierPrivate(cfg.PolicyPath, privacyPolicyBytesMax)
    if err != nil { return reject("privacy_policy_invalid") }
    policy, err := decodePrivacyPolicy(bytes.NewReader(policyBytes))
    if err != nil || protocolV7BaselineVerifierPrivacySelfTest(policy) != nil { return reject("privacy_policy_invalid") }
    scanner, err := newPrivacyScanner(policy)
    if err != nil { return reject("privacy_policy_invalid") }

    manifestBytes, err := protocolV7BaselineVerifierPrivate(cfg.ManifestPath, 4<<20)
    if err != nil || shaString(string(manifestBytes)) != cfg.ExpectedManifestSHA256 { return reject("metadata_invalid") }
    var manifest protocolV7BaselineCaptureManifest
    if decodeStrictBytes(manifestBytes, &manifest) != nil { return reject("metadata_invalid") }
    canonicalManifest, _ := json.Marshal(manifest)
    if !bytes.Equal(canonicalManifest, manifestBytes) || manifest.Schema != "structural-retrieval-protocol-v7-preparation.v1" || manifest.Status != "accepted" {
        return reject("metadata_invalid")
    }
    inputManifest := map[string]any{
        "$schema": "structural-retrieval-protocol-v7-adapter-inputs.v1", "status": "accepted",
        "adapter_sha256": cfg.AdapterSHA256, "driver_sha256": cfg.DriverSHA256,
        "input_sha256": map[string]string{"main": cfg.MainSHA256, "main_test": cfg.MainTestSHA256, "fixture": cfg.FixtureSHA256, "benchmark": cfg.BenchmarkSHA256},
        "scorer_region_sha256": cfg.ScorerRegionSHA256, "privacy_policy_sha256": policy.SHA256,
        "privacy_term_count": len(policy.DenyTerms), "threshold_checkin_seq": cfg.ThresholdCheckinSeq,
    }
    if manifest.AdapterSHA256 != cfg.AdapterSHA256 || manifest.DriverSHA256 != cfg.DriverSHA256 || manifest.MainSHA256 != cfg.MainSHA256 ||
        manifest.MainTestSHA256 != cfg.MainTestSHA256 || manifest.FixtureSHA256 != cfg.FixtureSHA256 || manifest.BenchmarkSHA256 != cfg.BenchmarkSHA256 ||
        manifest.ScorerRegionSHA256 != cfg.ScorerRegionSHA256 || manifest.PrivacyPolicySHA256 != policy.SHA256 || manifest.PrivacyTermCount != len(policy.DenyTerms) ||
        manifest.InputManifestSHA256 != canonicalSHA256(inputManifest) || manifest.RunCount != 6 {
        return reject("metadata_binding_invalid")
    }
    if err := protocolV7BaselineVerifierOneShot(cfg, manifest, policy, scanner); err != nil {
        if err.Error() == "privacy_violation" { return reject("privacy_violation") }
        return reject(err.Error())
    }
    authorityPath := filepath.Join(cfg.BaselineRoot, "parent/controller-authority.v4.json")
    outputPath := filepath.Join(cfg.BaselineRoot, "parent/controller-output.v4.json")
    authorityBytes, err := protocolV7BaselineVerifierPrivate(authorityPath, 4<<20)
    if err != nil || shaString(string(authorityBytes)) != manifest.AuthoritySHA256 { return reject("metadata_binding_invalid") }
    outputBytes, err := protocolV7BaselineVerifierPrivate(outputPath, 4<<20)
    if err != nil || shaString(string(outputBytes)) != manifest.ControllerOutputSHA256 { return reject("metadata_binding_invalid") }
    if hash, err := fileSHA256(cfg.BundlePath); err != nil || hash != manifest.BundleSHA256 { return reject("metadata_binding_invalid") }
    if hash, err := fileSHA256(cfg.RuntimePath); err != nil || hash != manifest.RuntimeSHA256 { return reject("metadata_binding_invalid") }
    var authority controllerAuthorityV4
    if decodeStrictBytes(authorityBytes, &authority) != nil || authority.Schema != controllerAuthorityV4Schema || authority.Mode != "baseline" || authority.ParentBaseline != nil || authority.CandidateDelta != nil {
        return reject("metadata_invalid")
    }
    if authority.Expected.Commit != manifest.Commit || authority.Expected.Tree != manifest.Tree || authority.Expected.SourceCapsuleSHA256 != manifest.SourceCapsuleSHA256 ||
        authority.Expected.RuntimeSHA256 != manifest.RuntimeSHA256 || authority.Expected.BundleSHA256 != manifest.BundleSHA256 || authority.SourceBundleHeadsSHA256 != manifest.BundleHeadsSHA256 {
        return reject("metadata_binding_invalid")
    }
    line, err := protocolV7BaselineVerifierCheckinLine(cfg.ThresholdCheckinsPath, cfg.ThresholdCheckinSeq)
    if err != nil || string(line) != authority.ContextThresholdAuthorityRecord { return reject("threshold_invalid") }

    sourceRoot := filepath.Join(cfg.WorkRoot, "B")
    if os.Mkdir(sourceRoot, 0o700) != nil { return reject("source_checkout_invalid") }
    if _, err := gitCommandBytes(sourceRoot, "init", "-q"); err != nil { return reject("source_checkout_invalid") }
    if _, err := gitCommandBytes(sourceRoot, "bundle", "unbundle", cfg.BundlePath); err != nil { return reject("bundle_invalid") }
    if _, err := gitCommandBytes(sourceRoot, "checkout", "-q", "--detach", manifest.Commit); err != nil { return reject("source_checkout_invalid") }
    if err := rejectReplaceRefs(sourceRoot); err != nil { return reject("bundle_invalid") }
    fixture := filepath.Join(sourceRoot, "research/eval/structural-retrieval/fixtures.dev.v2.jsonl")
    benchmark := filepath.Join(sourceRoot, "research/eval/structural-retrieval/benchmark.v2.json")
    scorer := filepath.Join(sourceRoot, "cli/tools/structural-search-eval-v2/main.go")
    bench, err := loadBenchmark(benchmark)
    if err != nil { return reject("metadata_invalid") }
    if err := verifyControllerAuthorityV4(authority, authorityBytes, cfg.RuntimePath, cfg.RuntimePath, sourceRoot, sourceRoot, cfg.BundlePath, cfg.PolicyPath, fixture, benchmark, scorer, bench, policy, scanner, parentBaselineFiles{}); err != nil {
        if scanner.tainted { return reject("privacy_violation") }
        return reject("authority_replay_failed")
    }

    var output controllerOutput
    if decodeStrictBytes(outputBytes, &output) != nil { return reject("baseline_replay_failed") }
    canonicalOutput, _ := json.Marshal(output)
    canonicalOutput = append(canonicalOutput, '\n')
    expectedRuns := bench.FixtureCount + 2
    if !bytes.Equal(canonicalOutput, outputBytes) || output.Schema != "structural-retrieval-controller-output.v4" || output.Admitted || output.Admission != "diagnostic_unadmitted" || output.Failure != nil ||
        len(output.Runs) != expectedRuns || expectedRuns != manifest.RunCount || output.OrderedRunManifestSHA256 != canonicalSHA256(output.Runs) ||
        output.OrderedRunManifestSHA256 != manifest.OrderedRunManifestSHA256 || output.HistorySHA256 != manifest.HistorySHA256 || output.PrivacyManifestSHA256 != manifest.PrivacyManifestSHA256 {
        return reject("baseline_replay_failed")
    }
    parentRoot := filepath.Join(cfg.BaselineRoot, "parent")
    historyBytes, err := parentRelativeArtifact(parentRoot, output.HistoryPath, authority.ScanCaps.SingleDurableArtifactBytesMax)
    if err != nil || shaString(string(historyBytes)) != output.HistorySHA256 { return reject("baseline_replay_failed") }
    history, err := decodeHistoryBytes(historyBytes)
    if err != nil || len(history) != expectedRuns || verifyHistorySchema(history, historyV4Schema) != nil { return reject("baseline_replay_failed") }
    var canonicalHistory bytes.Buffer
    for _, record := range history { data, _ := json.Marshal(record); canonicalHistory.Write(data); canonicalHistory.WriteByte('\n') }
    if !bytes.Equal(canonicalHistory.Bytes(), historyBytes) { return reject("baseline_replay_failed") }
    fixtures, fixtureHash, err := loadFixtures(fixture)
    if err != nil || fixtureHash != authority.Expected.FixtureSHA256 || len(fixtures) != bench.FixtureCount { return reject("baseline_replay_failed") }
    replayScanner, err := newPrivacyScanner(policy)
    if err != nil { return reject("privacy_policy_invalid") }
    known, err := verifyParentRunArtifacts(parentRoot, output, history, authority, fixtures, bench, replayScanner)
    if err != nil { return reject("baseline_replay_failed") }
    known[output.HistoryPath] = historyBytes
    for index := range output.Runs {
        path := filepath.ToSlash(filepath.Join("runtime", twoDigit(index+1)+".stderr"))
        data, err := parentRelativeArtifact(parentRoot, path, authority.ScanCaps.SingleDurableArtifactBytesMax)
        if err != nil { return reject("baseline_replay_failed") }
        known[path] = data
    }
    if err := verifyParentPrivacyManifest(parentRoot, output, authority, authorityBytes, known, sourceRoot, fixture, benchmark, scorer, replayScanner); err != nil {
        if replayScanner.tainted { return reject("privacy_violation") }
        return reject("baseline_replay_failed")
    }
    files := parentBaselineFiles{Root: parentRoot, Authority: authorityPath, Output: outputPath}
    if verifyParentRootCoverage(files, output) != nil { return reject("baseline_replay_failed") }
    if replayScanner.Scan("capture_manifest", "capture-manifest.json", manifestBytes) != nil || replayScanner.Scan("controller_output", "controller-output.v4.json", outputBytes) != nil {
        return reject("privacy_violation")
    }
    mainHash, err := fileSHA256(cfg.MainPath)
    if err != nil { return reject("frozen_input_invalid") }
    testHash, err := fileSHA256(cfg.MainTestPath)
    if err != nil { return reject("frozen_input_invalid") }
    acceptance := map[string]any{
        "$schema": "structural-retrieval-baseline-acceptance.v1", "verdict": "accepted",
        "authority_sha256": manifest.AuthoritySHA256, "output_sha256": manifest.ControllerOutputSHA256,
        "ordered_run_manifest_sha256": manifest.OrderedRunManifestSHA256, "run_count": manifest.RunCount,
        "history_sha256": manifest.HistorySHA256, "history_tail_record_hash": history[len(history)-1].RecordHash,
        "privacy_manifest_sha256": manifest.PrivacyManifestSHA256,
        "validated_source_main_sha256": mainHash, "validated_source_test_sha256": testHash,
    }
    return map[string]any{"$schema": protocolV7BaselineVerifierHarnessSchema, "baseline_acceptance": acceptance, "status": "accepted"}
}

func twoDigit(value int) string {
    if value < 10 { return "0" + string(rune('0'+value)) }
    return string(rune('0'+value/10)) + string(rune('0'+value%10))
}

'''


def _parse_args(argv: list[str] | None = None) -> argparse.Namespace:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--source-root", type=Path, required=True)
    parser.add_argument("--baseline-root", type=Path, required=True)
    parser.add_argument("--main", dest="main_path", type=Path, required=True)
    parser.add_argument("--main-test", dest="main_test_path", type=Path, required=True)
    parser.add_argument("--fixtures", dest="fixture_path", type=Path, required=True)
    parser.add_argument("--benchmark", dest="benchmark_path", type=Path, required=True)
    parser.add_argument("--threshold-checkins", dest="threshold_checkins_path", type=Path, required=True)
    parser.add_argument("--threshold-checkin-seq", type=int, required=True)
    parser.add_argument("--privacy-policy", dest="privacy_policy_path", type=Path, required=True)
    parser.add_argument("--adapter", dest="adapter_path", type=Path, required=True)
    parser.add_argument("--preparation-verifier", dest="preparation_verifier_path", type=Path, required=True)
    parser.add_argument("--activation-original", dest="activation_original_path", type=Path, required=True)
    parser.add_argument("--consumed-receipt", dest="consumed_receipt_path", type=Path, required=True)
    parser.add_argument("--authorization-record", dest="authorization_record_path", type=Path, required=True)
    parser.add_argument("--go-executable", dest="go_executable_path", type=Path, required=True)
    parser.add_argument("--git-executable", dest="git_executable_path", type=Path, required=True)
    return parser.parse_args(argv)


def main(argv: list[str] | None = None) -> int:
    args = _parse_args(argv)
    result = verify_baseline(BaselineVerifierConfig(**vars(args)))
    print(canonical_json(result).decode())
    return 0 if result.get("status") == "accepted" else 2


if __name__ == "__main__":
    raise SystemExit(main())
