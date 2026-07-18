#!/usr/bin/env python3
"""Prepare or capture one generic protocol-v7 structural-retrieval candidate."""

from __future__ import annotations

import argparse
import ctypes
from dataclasses import dataclass
from datetime import datetime
import hashlib
import json
import os
from pathlib import Path
import re
import selectors
import shutil
import signal
import stat
import subprocess
import tempfile
import time
from typing import Any, Callable, Iterable


REGISTERED_VARIABLE = "direct_hit_containment_owner_substitution"
CANDIDATE_EXECUTION_AUTHORIZED = False

ACCEPTED_MAIN_SHA256 = "33b6452a0e34d825ef2b895609daa77b91eb129ebdc1e6be8f7f7755fc8b9c6e"
ACCEPTED_MAIN_TEST_SHA256 = "0830a0671aa9eb146902dc6fa56e885126b89a3312ca6c38fd8507fa10481dbf"
ACCEPTED_FIXTURE_SHA256 = "15f57120c6aa9ae07bf4fdacd6ad783afa5e70ed8ebebaff3a42dcf4249e677e"
ACCEPTED_BENCHMARK_SHA256 = "b960525cc42216e6598452946da5fb68735bbf989f311f170cedcfdbe92bf0d5"
ACCEPTED_SCORER_REGION_SHA256 = "c1669d43b13a36eb1454a01a5ea7e444fe67ce28fdae9628da083927f093cbcb"
ACCEPTED_GO_EXECUTABLE_SHA256 = "86b748c64de0175db601f56805251f3b08cd12bffb927ad5c68ef8497c50c7ba"
ACCEPTED_GIT_EXECUTABLE_SHA256 = "356db14e102d68a1a37d8a1ac577dfd678d45d46e92f468bef8b7154e7bfdc60"

ACCEPTED_PARENT_AUTHORITY_SHA256 = "5f690bf41927351d9e5f1433d15576c0948c31233f9be09d40391de05ad5f38b"
ACCEPTED_PARENT_OUTPUT_SHA256 = "e65a394eb79e2bd87bb40ee1e0247b6f5eccc56f2e3ca4563d6aad5e029b03ff"
ACCEPTED_PARENT_HISTORY_SHA256 = "f1cae98f5c3a7950575c344c12df76721b052e1168714b83c178e03ac39f3a9e"
ACCEPTED_PARENT_HISTORY_TAIL = "664c40c18c9827d4d264334fb3c90247a7a94fe96a32524ecb295c99a8bad69e"
ACCEPTED_PARENT_ORDERED_RUN_SHA256 = "d910e36123481cf0954fe5a69319f8e49bbe7fd7ae679d428febe45241221db0"
ACCEPTED_PARENT_PRIVACY_SHA256 = "292fd38d6d4f5f5193a5a4715d86514d0748dfe8c748e458205c9fae65e5c396"
ACCEPTED_PARENT_BUNDLE_SHA256 = "0a7537fdba51a5750103b6f2d89b1e4332298f9383a965b20ae971f21c68c9fa"
ACCEPTED_PARENT_RUNTIME_SHA256 = "3f408146dff5f80014d66e89c6ab41f2bf905eb8b81c42e4fe1bf34a8ce298ed"
ACCEPTED_PARENT_COMMIT = "3bb7e9a700d1fba0469468b0566683ea51a4b115"
ACCEPTED_PARENT_TREE = "1c5f8b614cab6a3f23d83c9ec14504cd159dadd6"
ACCEPTED_PARENT_VALIDATOR_RECORD_HASH = "d33bcf4a49954761b801a4ebfc4503ee439ac759a7ece080abbeaa2786e14ebe"
ACCEPTED_PARENT_VALIDATOR_PAYLOAD_SHA256 = "be17283c0c0c4844c3a321efae17c4ab264025c056658c5ad73d928569a2dbbc"
ACCEPTED_PARENT_CLEANUP_RECORD_HASH = "9d35d2ce9fca4ec8706a2c415801a3923fbfc0b75fcb5b080610d1d3ea91e8b4"
ACCEPTED_PARENT_CLEANUP_PAYLOAD_SHA256 = "a6edd456eea827268701476589a602ff5cc7800a9db8b50aaa5562ce048942d9"
ACCEPTED_PARENT_RELATIVE_MANIFEST_SHA256 = "49421b1dc1579b2a213cb57dfdbbe5885cb191299cfff683b9a0feffa9b99c4d"
ACCEPTED_PARENT_FILE_COUNT = 25
ACCEPTED_PARENT_RUN_COUNT = 6

PARENT_VALIDATOR_REF_PREFIX = "workers/validator-baseline-protocol-v7/progress.jsonl#seq="
CAPABILITY_SCHEMA = "structural-retrieval-protocol-v7-candidate-capability.v1"
CAPTURE_MANIFEST_SCHEMA = "structural-retrieval-protocol-v7-candidate-capture.v1"
AUTHORIZATION_SCHEMA = "structural-retrieval-protocol-v7-candidate-authorization.v1"
ACTIVATION_SCHEMA = "structural-retrieval-protocol-v7-candidate-activation.v1"
RESULT_SCHEMA = "structural-retrieval-protocol-v7-candidate-adapter-result.v1"

MAX_SOURCE_BYTES = 4 << 20
MAX_POLICY_BYTES = 128 << 10
MAX_POLICY_TERMS = 512
MAX_METADATA_BYTES = 4 << 20
PROCESS_OUTPUT_CAP = 4 << 20
PROCESS_TIMEOUT_SECONDS = 1200

CAPABILITY_FILES = {
    "candidate-authority.v4.json",
    "capability-manifest.json",
    "freeze/controller-runtime",
    "freeze/candidate-runtime",
    "freeze/source.bundle",
}
CAPABILITY_MANIFEST_KEYS = {
    "$schema", "status", "effect_claim", "candidate_execution_authorized", "registered_variable",
    "candidate_delta_sha256", "candidate_commit", "candidate_tree", "parent_authority_sha256",
    "parent_output_sha256", "parent_validator_record_hash", "parent_validator_payload_sha256",
    "privacy_policy_sha256", "adapter_sha256", "validator_sha256", "files", "future_output_root_sha256",
}

CAPTURE_MANIFEST_KEYS = {
    "$schema", "status", "effect_claim", "candidate_execution_authorized", "max_capture_count",
    "candidate_delta_sha256", "candidate_authority_sha256", "capability_manifest_sha256",
    "adapter_sha256", "validator_sha256", "main_sha256", "main_test_sha256",
    "scorer_region_sha256", "parent_authority_sha256", "parent_output_sha256",
    "parent_validator_record_hash", "parent_cleanup_record_hash", "privacy_policy_sha256",
    "authorization_record_sha256", "activation_proof_sha256", "controller_output_sha256",
    "bundle_sha256", "candidate_runtime_sha256", "run_count",
}


class CandidateAdapterError(ValueError):
    """A bounded, generic candidate transaction failure."""


@dataclass(frozen=True)
class CandidateSource:
    path: Path
    sha256: str
    bytes: int


@dataclass(frozen=True)
class ParentBinding:
    root: Path
    authority: dict[str, Any]
    output: dict[str, Any]
    validator_store: Path
    validator_ref: str


@dataclass(frozen=True)
class OneShotBinding:
    authorization_path: Path
    authorization_sha256: str
    authorization_record_hash: str
    activation_path: Path
    activation_sha256: str
    output_root: Path
    capability_manifest_sha256: str
    candidate_authority_sha256: str
    candidate_delta_sha256: str
    candidate_runtime_sha256: str
    bundle_sha256: str
    privacy_policy_sha256: str
    adapter_sha256: str
    validator_sha256: str


@dataclass(frozen=True)
class ConsumedActivation:
    descriptor: int
    path: Path
    binding: OneShotBinding


@dataclass(frozen=True)
class ExecutionSnapshot:
    root: Path
    capability_root: Path
    privacy_policy: Path
    parent_baseline_root: Path
    parent_validator_store: Path
    tool_root: Path
    go: Path
    git: Path
    hashes: dict[str, str]


@dataclass(frozen=True)
class CandidateConfig:
    parent_baseline_root: Path
    parent_validator_store: Path
    parent_validator_ref: str
    privacy_policy: Path
    capability_root: Path
    output_root: Path
    candidate_search_go: Path | None = None
    candidate_search_test_go: Path | None = None
    activation: Path | None = None
    authorization_record: Path | None = None
    go_executable: Path | None = None
    git_executable: Path | None = None


@dataclass(frozen=True)
class ProcessResult:
    returncode: int
    stdout: bytes
    stderr: bytes


Runner = Callable[..., ProcessResult]


def canonical_json(value: Any) -> bytes:
    return json.dumps(value, sort_keys=True, separators=(",", ":"), ensure_ascii=False).encode()


def go_struct_json(value: Any) -> bytes:
    """Match encoding/json field order for decoded protocol structs."""
    return json.dumps(value, sort_keys=False, separators=(",", ":"), ensure_ascii=False).encode()


def sha256_bytes(data: bytes) -> str:
    return hashlib.sha256(data).hexdigest()


def sha256_file(path: Path, cap: int = 256 << 20, error: str = "governed_input_invalid") -> str:
    return sha256_bytes(read_regular(path, mode=None, cap=cap, error=error))


def valid_sha256(value: Any) -> bool:
    return isinstance(value, str) and bool(re.fullmatch(r"[0-9a-f]{64}", value)) and value != "0" * 64


def valid_oid(value: Any) -> bool:
    return isinstance(value, str) and bool(re.fullmatch(r"[0-9a-f]{40}", value))


def read_regular(path: Path, *, mode: int | None, cap: int, error: str) -> bytes:
    try:
        before = path.lstat()
        resolved = path.resolve(strict=True)
    except OSError as exc:
        raise CandidateAdapterError(error) from exc
    if stat.S_ISLNK(before.st_mode) or not stat.S_ISREG(before.st_mode) or resolved != path.absolute():
        raise CandidateAdapterError(error)
    if mode is not None and stat.S_IMODE(before.st_mode) != mode:
        raise CandidateAdapterError(error)
    if before.st_size < 0 or before.st_size > cap:
        raise CandidateAdapterError(error)
    try:
        with path.open("rb") as stream:
            data = stream.read(cap + 1)
        after = path.lstat()
    except OSError as exc:
        raise CandidateAdapterError(error) from exc
    identity = lambda info: (info.st_dev, info.st_ino, info.st_mode, info.st_size, info.st_mtime_ns)
    if len(data) != before.st_size or len(data) > cap or identity(before) != identity(after):
        raise CandidateAdapterError(error)
    return data


def require_directory(path: Path, mode: int | None, error: str) -> None:
    try:
        info = path.lstat()
    except OSError as exc:
        raise CandidateAdapterError(error) from exc
    if stat.S_ISLNK(info.st_mode) or not stat.S_ISDIR(info.st_mode) or path.resolve(strict=True) != path.absolute():
        raise CandidateAdapterError(error)
    if mode is not None and stat.S_IMODE(info.st_mode) != mode:
        raise CandidateAdapterError(error)


def _snapshot_registered_files(
    root: Path,
    sources: dict[str, Path],
    *,
    executable: set[str] | None = None,
    transition_hook: Callable[[], None] | None = None,
) -> dict[str, str]:
    """Copy exact governed bytes into one private, read-only execution tree."""
    executable = executable or set()
    if root.exists() or root.is_symlink() or not sources:
        raise CandidateAdapterError("governed_path_invalid")
    before: dict[str, tuple[bytes, tuple[int, int, int, int, int]]] = {}
    for relative, source in sources.items():
        path = Path(relative)
        if path.is_absolute() or ".." in path.parts or path.as_posix() != relative:
            raise CandidateAdapterError("governed_path_invalid")
        info = source.lstat()
        data = read_regular(source, mode=None, cap=256 << 20, error="governed_input_invalid")
        before[relative] = (data, (info.st_dev, info.st_ino, info.st_mode, info.st_size, info.st_mtime_ns))
    try:
        root.mkdir(mode=0o700)
        for relative, (data, _) in before.items():
            target = root / relative
            target.parent.mkdir(parents=True, exist_ok=True, mode=0o700)
            target.write_bytes(data)
            target.chmod(0o500 if relative in executable else 0o400)
        if transition_hook is not None:
            transition_hook()
        for relative, source in sources.items():
            data, identity = before[relative]
            info = source.lstat()
            current_identity = (info.st_dev, info.st_ino, info.st_mode, info.st_size, info.st_mtime_ns)
            if current_identity != identity or read_regular(source, mode=None, cap=256 << 20, error="governed_input_changed") != data:
                raise CandidateAdapterError("governed_input_changed")
            if sha256_file(root / relative, error="governed_input_changed") != sha256_bytes(data):
                raise CandidateAdapterError("governed_input_changed")
        for directory in sorted((path for path in root.rglob("*") if path.is_dir()), reverse=True):
            directory.chmod(0o500)
        root.chmod(0o500)
        return {relative: sha256_bytes(data) for relative, (data, _) in before.items()}
    except BaseException:
        shutil.rmtree(root, ignore_errors=True)
        raise


def _tree_sources(root: Path, prefix: str) -> dict[str, Path]:
    require_directory(root, None, "governed_input_invalid")
    result: dict[str, Path] = {}
    for path in sorted(root.rglob("*")):
        info = path.lstat()
        if stat.S_ISLNK(info.st_mode):
            raise CandidateAdapterError("governed_input_invalid")
        if stat.S_ISDIR(info.st_mode):
            continue
        if not stat.S_ISREG(info.st_mode):
            raise CandidateAdapterError("governed_input_invalid")
        result[f"{prefix}/{path.relative_to(root).as_posix()}"] = path
    if not result:
        raise CandidateAdapterError("governed_input_invalid")
    return result


def create_execution_snapshot(config: CandidateConfig, temporary: Path) -> ExecutionSnapshot:
    """Freeze every byte the controller may execute or read before activation is spent."""
    go = _resolve_tool(config.go_executable, "go", ACCEPTED_GO_EXECUTABLE_SHA256)
    git = _resolve_tool(config.git_executable, "git", ACCEPTED_GIT_EXECUTABLE_SHA256)
    sources = {
        **{f"capability/{relative}": config.capability_root / relative for relative in CAPABILITY_FILES},
        "policy/privacy-policy.json": config.privacy_policy,
        **_tree_sources(config.parent_baseline_root, "parent-baseline"),
        "validator-store/workers/validator-baseline-protocol-v7/progress.jsonl": config.parent_validator_store / "workers/validator-baseline-protocol-v7/progress.jsonl",
        "validator-store/workers/writer-protocol-v7-baseline-bb05-cleanup/progress.jsonl": config.parent_validator_store / "workers/writer-protocol-v7-baseline-bb05-cleanup/progress.jsonl",
        "tools/go": go,
        "tools/git": git,
    }
    root = temporary / "execution-snapshot"
    hashes = _snapshot_registered_files(
        root,
        sources,
        executable={"capability/freeze/controller-runtime", "capability/freeze/candidate-runtime", "tools/go", "tools/git"},
    )
    return ExecutionSnapshot(
        root=root,
        capability_root=root / "capability",
        privacy_policy=root / "policy/privacy-policy.json",
        parent_baseline_root=root / "parent-baseline",
        parent_validator_store=root / "validator-store",
        tool_root=root / "tools",
        go=root / "tools/go",
        git=root / "tools/git",
        hashes=hashes,
    )


def verify_execution_snapshot(snapshot: ExecutionSnapshot) -> None:
    executables = {"capability/freeze/controller-runtime", "capability/freeze/candidate-runtime", "tools/go", "tools/git"}
    if stat.S_IMODE(snapshot.root.lstat().st_mode) != 0o500:
        raise CandidateAdapterError("execution_snapshot_changed")
    seen: set[str] = set()
    for path in snapshot.root.rglob("*"):
        relative = path.relative_to(snapshot.root).as_posix()
        info = path.lstat()
        if stat.S_ISLNK(info.st_mode):
            raise CandidateAdapterError("execution_snapshot_changed")
        if stat.S_ISDIR(info.st_mode):
            if stat.S_IMODE(info.st_mode) != 0o500:
                raise CandidateAdapterError("execution_snapshot_changed")
            continue
        if not stat.S_ISREG(info.st_mode) or stat.S_IMODE(info.st_mode) != (0o500 if relative in executables else 0o400):
            raise CandidateAdapterError("execution_snapshot_changed")
        seen.add(relative)
    if seen != set(snapshot.hashes):
        raise CandidateAdapterError("execution_snapshot_changed")
    for relative, digest in snapshot.hashes.items():
        if sha256_file(snapshot.root / relative, error="execution_snapshot_changed") != digest:
            raise CandidateAdapterError("execution_snapshot_changed")


def verify_execution_snapshot_bindings(snapshot: ExecutionSnapshot, binding: OneShotBinding, config: CandidateConfig) -> None:
    verify_execution_snapshot(snapshot)
    expected = {
        "capability/capability-manifest.json": binding.capability_manifest_sha256,
        "capability/candidate-authority.v4.json": binding.candidate_authority_sha256,
        "capability/freeze/candidate-runtime": binding.candidate_runtime_sha256,
        "capability/freeze/source.bundle": binding.bundle_sha256,
        "capability/freeze/controller-runtime": ACCEPTED_PARENT_RUNTIME_SHA256,
        "policy/privacy-policy.json": binding.privacy_policy_sha256,
        "parent-baseline/freeze/source.bundle": ACCEPTED_PARENT_BUNDLE_SHA256,
        "parent-baseline/freeze/controller-runtime": ACCEPTED_PARENT_RUNTIME_SHA256,
        "parent-baseline/parent/controller-authority.v4.json": ACCEPTED_PARENT_AUTHORITY_SHA256,
        "parent-baseline/parent/controller-output.v4.json": ACCEPTED_PARENT_OUTPUT_SHA256,
        "parent-baseline/parent/history.jsonl": ACCEPTED_PARENT_HISTORY_SHA256,
        "parent-baseline/parent/privacy-scan.json": ACCEPTED_PARENT_PRIVACY_SHA256,
        "tools/go": ACCEPTED_GO_EXECUTABLE_SHA256,
        "tools/git": ACCEPTED_GIT_EXECUTABLE_SHA256,
    }
    if any(snapshot.hashes.get(relative) != digest for relative, digest in expected.items()):
        raise CandidateAdapterError("execution_snapshot_binding_invalid")
    parent_rows: list[dict[str, Any]] = []
    for relative in sorted(key for key in snapshot.hashes if key.startswith("parent-baseline/")):
        data = read_regular(snapshot.root / relative, mode=None, cap=128 << 20, error="execution_snapshot_binding_invalid")
        parent_rows.append({"path": relative.removeprefix("parent-baseline/"), "bytes": len(data), "sha256": sha256_bytes(data)})
    # The accepted cleanup record binds the parent manifest opaquely; its private
    # serializer is intentionally not recreated here. Exact key artifacts and
    # both accepted worker prefixes are checked below, while every remaining
    # parent byte is frozen and hash-checked by verify_execution_snapshot.
    if len(parent_rows) != ACCEPTED_PARENT_FILE_COUNT:
        raise CandidateAdapterError("execution_snapshot_binding_invalid")
    validator_data = read_regular(snapshot.parent_validator_store / "workers/validator-baseline-protocol-v7/progress.jsonl", mode=None, cap=4 << 20, error="execution_snapshot_binding_invalid")
    suffix = config.parent_validator_ref.removeprefix(PARENT_VALIDATOR_REF_PREFIX)
    validate_worker_record_prefix(validator_data, int(suffix), ACCEPTED_PARENT_VALIDATOR_RECORD_HASH, ACCEPTED_PARENT_VALIDATOR_PAYLOAD_SHA256)
    cleanup_data = read_regular(snapshot.parent_validator_store / "workers/writer-protocol-v7-baseline-bb05-cleanup/progress.jsonl", mode=None, cap=4 << 20, error="execution_snapshot_binding_invalid")
    validate_worker_record_prefix(cleanup_data, 1, ACCEPTED_PARENT_CLEANUP_RECORD_HASH, ACCEPTED_PARENT_CLEANUP_PAYLOAD_SHA256)
    manifest = decode_canonical(read_regular(snapshot.capability_root / "capability-manifest.json", mode=None, cap=MAX_METADATA_BYTES, error="execution_snapshot_binding_invalid"), "execution_snapshot_binding_invalid")
    authority = decode_canonical(read_regular(snapshot.capability_root / "candidate-authority.v4.json", mode=None, cap=MAX_METADATA_BYTES, error="execution_snapshot_binding_invalid"), "execution_snapshot_binding_invalid")
    if manifest.get("adapter_sha256") != binding.adapter_sha256 or manifest.get("validator_sha256") != binding.validator_sha256:
        raise CandidateAdapterError("execution_snapshot_binding_invalid")
    validate_manifest_authority_bindings(manifest, authority)


def release_execution_snapshot(snapshot: ExecutionSnapshot) -> None:
    if not snapshot.root.exists():
        return
    for path in snapshot.root.rglob("*"):
        try:
            path.chmod(0o700 if path.is_dir() else 0o600)
        except OSError:
            pass
    try:
        snapshot.root.chmod(0o700)
    except OSError:
        pass


def decode_canonical(data: bytes, error: str) -> dict[str, Any]:
    def pairs(values: list[tuple[str, Any]]) -> dict[str, Any]:
        result: dict[str, Any] = {}
        for key, value in values:
            if key in result:
                raise CandidateAdapterError(error)
            result[key] = value
        return result
    try:
        value = json.loads(data, object_pairs_hook=pairs)
    except (UnicodeDecodeError, json.JSONDecodeError, CandidateAdapterError) as exc:
        raise CandidateAdapterError(error) from exc
    if not isinstance(value, dict) or canonical_json(value) != data:
        raise CandidateAdapterError(error)
    return value


def decode_strict_object(data: bytes, error: str) -> dict[str, Any]:
    def pairs(values: list[tuple[str, Any]]) -> dict[str, Any]:
        result: dict[str, Any] = {}
        for key, value in values:
            if key in result:
                raise CandidateAdapterError(error)
            result[key] = value
        return result
    try:
        value = json.loads(data, object_pairs_hook=pairs)
    except (UnicodeDecodeError, json.JSONDecodeError, CandidateAdapterError) as exc:
        raise CandidateAdapterError(error) from exc
    if not isinstance(value, dict):
        raise CandidateAdapterError(error)
    return value


def paths_overlap(left: Path, right: Path) -> bool:
    left = left.absolute()
    right = right.absolute()
    try:
        left.relative_to(right)
        return True
    except ValueError:
        try:
            right.relative_to(left)
            return True
        except ValueError:
            return False


def validate_transaction_paths(repo: Path, capability: Path, output: Path, governed: Iterable[Path]) -> None:
    require_directory(repo, None, "governed_path_invalid")
    for absent in (capability, output):
        try:
            absent.lstat()
        except FileNotFoundError:
            pass
        except OSError as exc:
            raise CandidateAdapterError("governed_path_invalid") from exc
        else:
            raise CandidateAdapterError("governed_path_invalid")
        if not absent.is_absolute() or absent.parent.resolve(strict=True) != absent.absolute().parent:
            raise CandidateAdapterError("governed_path_invalid")
    paths = (repo, capability, output, *governed)
    if any(paths_overlap(left, right) for index, left in enumerate(paths) for right in paths[index + 1 :]):
        raise CandidateAdapterError("governed_path_invalid")


def decode_privacy_policy(data: bytes) -> tuple[str, tuple[str, ...]]:
    value = decode_canonical(data, "privacy_policy_invalid")
    terms = value.get("deny_terms")
    if set(value) != {"$schema", "deny_terms"} or value.get("$schema") != "structural-retrieval-privacy-policy.v1" or not isinstance(terms, list) or not terms or len(terms) > MAX_POLICY_TERMS:
        raise CandidateAdapterError("privacy_policy_invalid")
    if any(not isinstance(term, str) or not term or len(term.encode()) > 64 << 10 or term != term.casefold() for term in terms):
        raise CandidateAdapterError("privacy_policy_invalid")
    if terms != sorted(set(terms)):
        raise CandidateAdapterError("privacy_policy_invalid")
    return sha256_bytes(data), tuple(terms)


def validate_privacy_policy(path: Path) -> tuple[str, tuple[str, ...]]:
    return decode_privacy_policy(read_regular(path, mode=0o600, cap=MAX_POLICY_BYTES, error="privacy_policy_invalid"))


def privacy_hit(data: bytes, deny_terms: Iterable[str]) -> bool:
    folded = data.decode("utf-8", "ignore").casefold()
    generic = (
        rb"/(?:home|users|root)/[^/\s]+",
        rb"authorization:\s*bearer",
        rb"-----begin (?:rsa |ec |openssh )?private key-----",
        rb"(?:api[_-]?key|access[_-]?token)\s*[:=]\s*['\"]?[a-z0-9_\-]{16,}",
    )
    return any(term in folded for term in deny_terms) or any(re.search(pattern, data.lower()) for pattern in generic)


def validate_candidate_source(path: Path, baseline_sha256: str, deny_terms: Iterable[str]) -> CandidateSource:
    data = read_regular(path, mode=0o600, cap=MAX_SOURCE_BYTES, error="candidate_input_invalid")
    digest = sha256_bytes(data)
    if digest == baseline_sha256:
        raise CandidateAdapterError("candidate_unchanged")
    if privacy_hit(data, deny_terms):
        raise CandidateAdapterError("privacy_violation")
    return CandidateSource(path=path, sha256=digest, bytes=len(data))


def validate_worker_record_prefix(data: bytes, seq: int, record_hash: str, payload_hash: str) -> dict[str, Any]:
    if seq <= 0 or not valid_sha256(record_hash) or not valid_sha256(payload_hash) or len(data) > 4 << 20:
        raise CandidateAdapterError("parent_validator_invalid")
    previous = "GENESIS"
    found: dict[str, Any] | None = None
    lines = data.splitlines(keepends=True)
    for expected_seq, line in enumerate(lines, start=1):
        if not line.endswith(b"\n") or line.count(b"\n") != 1:
            raise CandidateAdapterError("parent_validator_invalid")
        record = decode_canonical(line[:-1], "parent_validator_invalid")
        if set(record) != {"seq", "recorded_at", "prev_hash", "payload_sha256", "payload", "record_hash"} or record.get("seq") != expected_seq or record.get("prev_hash") != previous:
            raise CandidateAdapterError("parent_validator_invalid")
        payload = record.get("payload")
        got_payload = sha256_bytes(canonical_json(payload))
        without_hash = {key: value for key, value in record.items() if key != "record_hash"}
        got_record = sha256_bytes(canonical_json(without_hash))
        if record.get("payload_sha256") != got_payload or record.get("record_hash") != got_record:
            raise CandidateAdapterError("parent_validator_invalid")
        previous = got_record
        if expected_seq == seq:
            if got_record != record_hash or got_payload != payload_hash or not isinstance(payload, dict):
                raise CandidateAdapterError("parent_validator_invalid")
            found = payload
            break
    if found is None or len(lines) != seq:
        raise CandidateAdapterError("parent_validator_invalid")
    return found


def _relative_manifest(root: Path) -> tuple[int, str]:
    rows: list[dict[str, Any]] = []
    count = 0
    for path in sorted(root.rglob("*")):
        info = path.lstat()
        relative = path.relative_to(root).as_posix()
        if stat.S_ISLNK(info.st_mode):
            raise CandidateAdapterError("parent_layout_invalid")
        if stat.S_ISDIR(info.st_mode):
            if stat.S_IMODE(info.st_mode) != 0o700:
                raise CandidateAdapterError("parent_layout_invalid")
            continue
        if not stat.S_ISREG(info.st_mode) or stat.S_IMODE(info.st_mode) != 0o600:
            raise CandidateAdapterError("parent_layout_invalid")
        data = read_regular(path, mode=0o600, cap=128 << 20, error="parent_layout_invalid")
        rows.append({"path": relative, "bytes": len(data), "sha256": sha256_bytes(data)})
        count += 1
    return count, sha256_bytes(canonical_json(rows))


def validate_parent(config: CandidateConfig) -> ParentBinding:
    root = config.parent_baseline_root
    require_directory(root, 0o700, "parent_layout_invalid")
    required = {
        "capture-manifest.json": None,
        "freeze/source.bundle": ACCEPTED_PARENT_BUNDLE_SHA256,
        "freeze/controller-runtime": ACCEPTED_PARENT_RUNTIME_SHA256,
        "parent/controller-authority.v4.json": ACCEPTED_PARENT_AUTHORITY_SHA256,
        "parent/controller-output.v4.json": ACCEPTED_PARENT_OUTPUT_SHA256,
        "parent/history.jsonl": ACCEPTED_PARENT_HISTORY_SHA256,
        "parent/privacy-scan.json": ACCEPTED_PARENT_PRIVACY_SHA256,
    }
    for relative, expected in required.items():
        digest = sha256_file(root / relative, error="parent_layout_invalid")
        if expected is not None and digest != expected:
            raise CandidateAdapterError("parent_hash_invalid")
    count, _snapshot_hash = _relative_manifest(root)
    if count != ACCEPTED_PARENT_FILE_COUNT:
        raise CandidateAdapterError("parent_layout_invalid")
    authority = decode_strict_object(read_regular(root / "parent/controller-authority.v4.json", mode=0o600, cap=MAX_METADATA_BYTES, error="parent_hash_invalid"), "parent_hash_invalid")
    output_bytes = read_regular(root / "parent/controller-output.v4.json", mode=0o600, cap=MAX_METADATA_BYTES, error="parent_hash_invalid")
    if not output_bytes.endswith(b"\n") or output_bytes.count(b"\n") != 1:
        raise CandidateAdapterError("parent_hash_invalid")
    output = decode_strict_object(output_bytes[:-1], "parent_hash_invalid")
    if authority.get("mode") != "baseline" or authority.get("candidate_delta") is not None or output.get("runs") is None or len(output["runs"]) != ACCEPTED_PARENT_RUN_COUNT:
        raise CandidateAdapterError("parent_hash_invalid")
    if not config.parent_validator_ref.startswith(PARENT_VALIDATOR_REF_PREFIX):
        raise CandidateAdapterError("parent_validator_invalid")
    suffix = config.parent_validator_ref.removeprefix(PARENT_VALIDATOR_REF_PREFIX)
    if not suffix.isdigit() or str(int(suffix)) != suffix:
        raise CandidateAdapterError("parent_validator_invalid")
    validator_path = config.parent_validator_store / "workers/validator-baseline-protocol-v7/progress.jsonl"
    validator_data = read_regular(validator_path, mode=None, cap=4 << 20, error="parent_validator_invalid")
    payload = validate_worker_record_prefix(validator_data, int(suffix), ACCEPTED_PARENT_VALIDATOR_RECORD_HASH, ACCEPTED_PARENT_VALIDATOR_PAYLOAD_SHA256)
    acceptance = payload.get("baseline_acceptance")
    expected_acceptance = {
        "$schema": "structural-retrieval-baseline-acceptance.v1",
        "verdict": "accepted",
        "authority_sha256": ACCEPTED_PARENT_AUTHORITY_SHA256,
        "output_sha256": ACCEPTED_PARENT_OUTPUT_SHA256,
        "ordered_run_manifest_sha256": ACCEPTED_PARENT_ORDERED_RUN_SHA256,
        "run_count": ACCEPTED_PARENT_RUN_COUNT,
        "history_sha256": ACCEPTED_PARENT_HISTORY_SHA256,
        "history_tail_record_hash": ACCEPTED_PARENT_HISTORY_TAIL,
        "privacy_manifest_sha256": ACCEPTED_PARENT_PRIVACY_SHA256,
        "validated_source_main_sha256": ACCEPTED_MAIN_SHA256,
        "validated_source_test_sha256": ACCEPTED_MAIN_TEST_SHA256,
    }
    if payload.get("event") != "finish" or payload.get("worker_id") != "validator-baseline-protocol-v7" or payload.get("role") != "independent baseline validator" or payload.get("status") != "accepted" or payload.get("effect_claim") is not False or acceptance != expected_acceptance:
        raise CandidateAdapterError("parent_validator_invalid")
    cleanup_path = config.parent_validator_store / "workers/writer-protocol-v7-baseline-bb05-cleanup/progress.jsonl"
    cleanup_data = read_regular(cleanup_path, mode=None, cap=4 << 20, error="parent_cleanup_invalid")
    cleanup = validate_worker_record_prefix(cleanup_data, 1, ACCEPTED_PARENT_CLEANUP_RECORD_HASH, ACCEPTED_PARENT_CLEANUP_PAYLOAD_SHA256)
    cleanup_value = cleanup.get("cleanup")
    if (
        cleanup.get("event") != "finish"
        or cleanup.get("worker_id") != "writer-protocol-v7-baseline-bb05-cleanup"
        or cleanup.get("effect_claim") is not False
        or not isinstance(cleanup_value, dict)
        or cleanup_value.get("activation_absent") is not True
        or cleanup_value.get("authorization_record_absent") is not True
        or cleanup_value.get("consumed_receipt_absent") is not True
        or cleanup_value.get("retained_baseline_file_count") != ACCEPTED_PARENT_FILE_COUNT
        or cleanup_value.get("retained_baseline_manifest_sha256") != ACCEPTED_PARENT_RELATIVE_MANIFEST_SHA256
    ):
        raise CandidateAdapterError("parent_cleanup_invalid")
    return ParentBinding(root, authority, output, config.parent_validator_store, config.parent_validator_ref)


def candidate_delta_sha256(delta: dict[str, Any]) -> str:
    allowed = delta.get("allowed_paths")
    registered = ["cli/cmd/search.go"]
    if allowed == ["cli/cmd/search.go", "cli/cmd/search_test.go"]:
        registered = allowed
    if delta.get("variable") != REGISTERED_VARIABLE or allowed != registered or delta.get("name_status") != [f"M\t{path}" for path in registered]:
        raise CandidateAdapterError("candidate_delta_invalid")
    if set(delta.get("before_blob_sha256", {})) != set(registered) or set(delta.get("after_blob_sha256", {})) != set(registered):
        raise CandidateAdapterError("candidate_delta_invalid")
    for key in ("baseline_commit", "baseline_tree", "candidate_commit", "candidate_tree"):
        if not valid_oid(delta.get(key)):
            raise CandidateAdapterError("candidate_delta_invalid")
    for key in ("diff_sha256", "name_status_sha256", "bundle_sha256", "bundle_heads_sha256"):
        if not valid_sha256(delta.get(key)):
            raise CandidateAdapterError("candidate_delta_invalid")
    if any(not valid_sha256(value) for mapping in (delta["before_blob_sha256"], delta["after_blob_sha256"]) for value in mapping.values()):
        raise CandidateAdapterError("candidate_delta_invalid")
    return sha256_bytes(canonical_json(delta))


def validate_manifest_authority_bindings(manifest: dict[str, Any], authority: dict[str, Any]) -> None:
    delta = authority.get("candidate_delta")
    expected = authority.get("expected_provenance")
    capsule = authority.get("runtime_source_capsule")
    parent = authority.get("parent_baseline")
    files = manifest.get("files")
    if not isinstance(delta, dict) or not isinstance(expected, dict) or not isinstance(capsule, dict) or not isinstance(parent, dict) or not isinstance(files, dict):
        raise CandidateAdapterError("capability_manifest_invalid")
    try:
        delta_hash = candidate_delta_sha256(delta)
    except CandidateAdapterError as exc:
        raise CandidateAdapterError("capability_manifest_invalid") from exc
    if (
        manifest.get("candidate_delta_sha256") != delta_hash
        or manifest.get("candidate_commit") != delta.get("candidate_commit")
        or manifest.get("candidate_tree") != delta.get("candidate_tree")
        or files.get("freeze/source.bundle") != delta.get("bundle_sha256")
        or expected.get("commit") != delta.get("candidate_commit")
        or expected.get("tree") != delta.get("candidate_tree")
        or expected.get("controller_commit") != delta.get("baseline_commit")
        or expected.get("controller_tree") != delta.get("baseline_tree")
        or expected.get("candidate_delta_sha256") != delta_hash
        or expected.get("bundle_sha256") != delta.get("bundle_sha256")
        or expected.get("runtime_sha256") != files.get("freeze/candidate-runtime")
        or capsule.get("head_commit") != delta.get("candidate_commit")
        or capsule.get("head_tree") != delta.get("candidate_tree")
        or parent.get("authority_sha256") != manifest.get("parent_authority_sha256")
        or parent.get("output_sha256") != manifest.get("parent_output_sha256")
        or parent.get("validator_record_hash") != manifest.get("parent_validator_record_hash")
        or parent.get("validator_payload_sha256") != manifest.get("parent_validator_payload_sha256")
    ):
        raise CandidateAdapterError("capability_manifest_invalid")


def _git(git: Path, cwd: Path, *args: str, cap: int = PROCESS_OUTPUT_CAP) -> bytes:
    environment = {
        "PATH": str(git.parent), "LC_ALL": "C", "LANG": "C", "TZ": "UTC",
        "HOME": str(cwd), "XDG_CONFIG_HOME": str(cwd), "GIT_CONFIG_NOSYSTEM": "1",
        "GIT_CONFIG_GLOBAL": "/dev/null", "GIT_OPTIONAL_LOCKS": "0", "GIT_NO_REPLACE_OBJECTS": "1",
        "GIT_ALTERNATE_OBJECT_DIRECTORIES": "", "GIT_ATTR_NOSYSTEM": "1", "GIT_DISCOVERY_ACROSS_FILESYSTEM": "0",
        "GIT_AUTHOR_NAME": "C3 Evaluation", "GIT_AUTHOR_EMAIL": "c3-eval@invalid",
        "GIT_COMMITTER_NAME": "C3 Evaluation", "GIT_COMMITTER_EMAIL": "c3-eval@invalid",
        "GIT_AUTHOR_DATE": "2000-01-01T00:00:00Z", "GIT_COMMITTER_DATE": "2000-01-01T00:00:00Z",
    }
    command = [str(git), "-c", "core.hooksPath=/dev/null", "-c", "core.attributesFile=/dev/null", "-c", "protocol.file.allow=never", *args]
    completed = run_bounded_process(command, cwd=cwd, env=environment, timeout=120, output_cap=cap)
    if completed.returncode != 0 or completed.stderr:
        raise CandidateAdapterError("git_operation_failed")
    return completed.stdout


def _resolve_tool(path: Path | None, name: str, expected: str) -> Path:
    value = path or (Path(found) if (found := shutil.which(name)) else None)
    if value is None or not value.is_absolute():
        raise CandidateAdapterError("tool_identity_invalid")
    try:
        value = value.resolve(strict=True)
    except OSError as exc:
        raise CandidateAdapterError("tool_identity_invalid") from exc
    if sha256_file(value, error="tool_identity_invalid") != expected:
        raise CandidateAdapterError("tool_identity_invalid")
    return value


def _checkout_bundle(git: Path, bundle: Path, destination: Path, commit: str) -> None:
    destination.mkdir(mode=0o700)
    _git(git, destination, "init", "-q")
    _git(git, destination, "bundle", "unbundle", str(bundle))
    _git(git, destination, "checkout", "-q", "--detach", commit)
    if _git(git, destination, "status", "--porcelain=v1", "--untracked-files=all"):
        raise CandidateAdapterError("source_checkout_invalid")


def _build_runtime(go: Path, source: Path, output: Path) -> None:
    home = Path.home()
    env = {
        "PATH": str(go.parent), "HOME": str(home), "LC_ALL": "C", "LANG": "C", "TZ": "UTC",
        "GOENV": "off", "GOWORK": "off", "GOFLAGS": "", "GOTOOLCHAIN": "local", "CGO_ENABLED": "0",
        "GOPROXY": "off", "GOSUMDB": "off", "GONOSUMDB": "", "GOPRIVATE": "",
        "GOMODCACHE": str(home / "go/pkg/mod"), "GOCACHE": str(Path(os.environ.get("XDG_CACHE_HOME", str(home / ".cache"))) / "go-build"),
    }
    completed = run_bounded_process([str(go), "build", "-mod=readonly", "-trimpath", "-buildvcs=false", "-o", str(output), "./tools/structural-search-eval-v2"], cwd=source / "cli", env=env, timeout=600, output_cap=PROCESS_OUTPUT_CAP)
    if completed.returncode != 0 or completed.stdout or completed.stderr:
        raise CandidateAdapterError("runtime_build_failed")
    output.chmod(0o600)


def _capsule_candidate(parent: dict[str, Any], commit: str, tree: str, after: dict[str, str]) -> dict[str, Any]:
    capsule = json.loads(json.dumps(parent))
    capsule["head_commit"] = commit
    capsule["head_tree"] = tree
    seen: set[str] = set()
    for entry in capsule.get("inputs", []):
        path = entry.get("path")
        if path in after:
            entry["sha256"] = after[path]
            seen.add(path)
    if seen != set(after):
        raise CandidateAdapterError("source_capsule_invalid")
    return capsule


def _parent_binding() -> dict[str, Any]:
    return {
        "authority_sha256": ACCEPTED_PARENT_AUTHORITY_SHA256,
        "output_sha256": ACCEPTED_PARENT_OUTPUT_SHA256,
        "ordered_run_manifest_sha256": ACCEPTED_PARENT_ORDERED_RUN_SHA256,
        "run_count": ACCEPTED_PARENT_RUN_COUNT,
        "history_sha256": ACCEPTED_PARENT_HISTORY_SHA256,
        "history_tail_record_hash": ACCEPTED_PARENT_HISTORY_TAIL,
        "privacy_manifest_sha256": ACCEPTED_PARENT_PRIVACY_SHA256,
        "validator_record_ref": "workers/validator-baseline-protocol-v7/progress.jsonl#seq=1",
        "validator_record_hash": ACCEPTED_PARENT_VALIDATOR_RECORD_HASH,
        "validator_payload_sha256": ACCEPTED_PARENT_VALIDATOR_PAYLOAD_SHA256,
    }


def build_candidate_authority(parent: dict[str, Any], capsule: dict[str, Any], runtime_hash: str, delta: dict[str, Any]) -> dict[str, Any]:
    authority = json.loads(json.dumps(parent))
    authority["mode"] = "candidate"
    authority["controller_source_capsule"] = parent["controller_source_capsule"]
    authority["runtime_source_capsule"] = capsule
    authority["candidate_delta"] = delta
    authority["parent_baseline"] = _parent_binding()
    authority["source_bundle_heads_sha256"] = delta["bundle_heads_sha256"]
    expected = authority["expected_provenance"]
    expected.update({
        "experiment_id": "protocol-v7-candidate-direct-hit-containment-owner-substitution",
        "arm_id": "candidate",
        "commit": delta["candidate_commit"], "tree": delta["candidate_tree"],
        "controller_commit": delta["baseline_commit"], "controller_tree": delta["baseline_tree"],
        "source_capsule_sha256": sha256_bytes(go_struct_json(capsule)),
        "controller_source_capsule_sha256": sha256_bytes(go_struct_json(parent["controller_source_capsule"])),
        "bundle_sha256": delta["bundle_sha256"], "candidate_delta_sha256": candidate_delta_sha256(delta),
        "controller_sha256": ACCEPTED_PARENT_RUNTIME_SHA256, "runtime_sha256": runtime_hash,
        "source_bundle_heads_sha256": delta["bundle_heads_sha256"],
        "parent_baseline_authority_sha256": ACCEPTED_PARENT_AUTHORITY_SHA256,
        "parent_baseline_output_sha256": ACCEPTED_PARENT_OUTPUT_SHA256,
        "parent_baseline_ordered_run_manifest_sha256": ACCEPTED_PARENT_ORDERED_RUN_SHA256,
        "parent_baseline_run_count": ACCEPTED_PARENT_RUN_COUNT,
        "parent_baseline_history_sha256": ACCEPTED_PARENT_HISTORY_SHA256,
        "parent_baseline_history_tail_record_hash": ACCEPTED_PARENT_HISTORY_TAIL,
        "parent_baseline_privacy_manifest_sha256": ACCEPTED_PARENT_PRIVACY_SHA256,
        "parent_baseline_validator_record_hash": ACCEPTED_PARENT_VALIDATOR_RECORD_HASH,
        "parent_baseline_validator_payload_sha256": ACCEPTED_PARENT_VALIDATOR_PAYLOAD_SHA256,
    })
    authority["build_replay"].update({
        "controller_sha256": ACCEPTED_PARENT_RUNTIME_SHA256,
        "runtime_sha256": runtime_hash,
        "rebuilt_controller_sha256": ACCEPTED_PARENT_RUNTIME_SHA256,
        "rebuilt_runtime_sha256": runtime_hash,
        "controller_capsule_rebuild_verified": True,
        "runtime_capsule_rebuild_verified": True,
        "bundle_verified": True,
    })
    return authority


def _manifest_snapshot(root: Path, allowed: set[str]) -> dict[str, str]:
    seen: dict[str, str] = {}
    for path in sorted(root.rglob("*")):
        info = path.lstat()
        if stat.S_ISLNK(info.st_mode):
            raise CandidateAdapterError("capability_layout_invalid")
        if stat.S_ISDIR(info.st_mode):
            if stat.S_IMODE(info.st_mode) != 0o700:
                raise CandidateAdapterError("capability_layout_invalid")
            continue
        relative = path.relative_to(root).as_posix()
        if relative not in allowed or not stat.S_ISREG(info.st_mode) or stat.S_IMODE(info.st_mode) != 0o600:
            raise CandidateAdapterError("capability_layout_invalid")
        seen[relative] = sha256_file(path, error="capability_layout_invalid")
    if set(seen) != allowed:
        raise CandidateAdapterError("capability_layout_invalid")
    return seen


def _adapter_sha256() -> str:
    return sha256_file(Path(__file__).resolve(), cap=4 << 20, error="adapter_identity_invalid")


def _validator_sha256() -> str:
    path = Path(__file__).with_name("validate_structural_retrieval_candidate_v7.py")
    return sha256_file(path, cap=4 << 20, error="validator_identity_invalid")


def _governed_input_snapshot(config: CandidateConfig, go: Path, git: Path) -> dict[str, str]:
    paths: dict[str, Path | None] = {
        "policy": config.privacy_policy,
        "candidate_search_go": config.candidate_search_go,
        "parent_validator": config.parent_validator_store / "workers/validator-baseline-protocol-v7/progress.jsonl",
        "parent_cleanup": config.parent_validator_store / "workers/writer-protocol-v7-baseline-bb05-cleanup/progress.jsonl",
        "go": go,
        "git": git,
        "adapter": Path(__file__).resolve(),
        "validator": Path(__file__).with_name("validate_structural_retrieval_candidate_v7.py").resolve(),
    }
    if config.candidate_search_test_go is not None:
        paths["candidate_search_test_go"] = config.candidate_search_test_go
    if paths["candidate_search_go"] is None:
        raise CandidateAdapterError("candidate_input_invalid")
    return {
        name: sha256_file(path, error="governed_input_invalid")
        for name, path in paths.items()
        if path is not None
    }


def validate_preparation_inputs(config: CandidateConfig) -> dict[str, Any]:
    parent = validate_parent(config)
    policy_hash, terms = validate_privacy_policy(config.privacy_policy)
    go = _resolve_tool(config.go_executable, "go", ACCEPTED_GO_EXECUTABLE_SHA256)
    git = _resolve_tool(config.git_executable, "git", ACCEPTED_GIT_EXECUTABLE_SHA256)
    if config.candidate_search_go is None:
        raise CandidateAdapterError("candidate_input_invalid")
    baseline_inputs = {
        entry["path"]: entry["sha256"]
        for entry in parent.authority["controller_source_capsule"]["inputs"]
    }
    search = validate_candidate_source(config.candidate_search_go, baseline_inputs["cli/cmd/search.go"], terms)
    governed = [config.privacy_policy, config.candidate_search_go, config.parent_validator_store, go, git]
    if config.candidate_search_test_go is not None:
        baseline_test = baseline_inputs.get("cli/cmd/search_test.go")
        if baseline_test is None:
            raise CandidateAdapterError("candidate_input_invalid")
        validate_candidate_source(config.candidate_search_test_go, baseline_test, terms)
        governed.append(config.candidate_search_test_go)
    validate_transaction_paths(config.parent_baseline_root, config.capability_root, config.output_root, governed)
    snapshot = _governed_input_snapshot(config, go, git)
    return {
        "$schema": RESULT_SCHEMA,
        "status": "inputs_valid",
        "effect_claim": False,
        "candidate_execution_authorized": False,
        "registered_variable": REGISTERED_VARIABLE,
        "candidate_search_go_sha256": search.sha256,
        "privacy_policy_sha256": policy_hash,
        "parent_authority_sha256": ACCEPTED_PARENT_AUTHORITY_SHA256,
        "parent_validator_record_hash": ACCEPTED_PARENT_VALIDATOR_RECORD_HASH,
        "parent_cleanup_record_hash": ACCEPTED_PARENT_CLEANUP_RECORD_HASH,
        "governed_input_manifest_sha256": sha256_bytes(canonical_json(snapshot)),
    }


def prepare_candidate(config: CandidateConfig, *, preflight: Callable[..., None] | None = None) -> dict[str, Any]:
    validate_preparation_inputs(config)
    parent = validate_parent(config)
    policy_hash, terms = validate_privacy_policy(config.privacy_policy)
    go = _resolve_tool(config.go_executable, "go", ACCEPTED_GO_EXECUTABLE_SHA256)
    git = _resolve_tool(config.git_executable, "git", ACCEPTED_GIT_EXECUTABLE_SHA256)
    if config.candidate_search_go is None:
        raise CandidateAdapterError("candidate_input_invalid")
    baseline_search_hash = next(entry["sha256"] for entry in parent.authority["controller_source_capsule"]["inputs"] if entry["path"] == "cli/cmd/search.go")
    search = validate_candidate_source(config.candidate_search_go, baseline_search_hash, terms)
    governed = [config.privacy_policy, config.candidate_search_go, config.parent_validator_store, go, git]
    if config.candidate_search_test_go is not None:
        governed.append(config.candidate_search_test_go)
    validate_transaction_paths(config.parent_baseline_root, config.capability_root, config.output_root, governed)
    parent_before = _relative_manifest(config.parent_baseline_root)
    governed_before = _governed_input_snapshot(config, go, git)
    with tempfile.TemporaryDirectory(prefix="c3-v7-candidate-prepare-") as raw:
        temporary = Path(raw)
        temporary.chmod(0o700)
        bundle_parent = config.parent_baseline_root / "freeze/source.bundle"
        b_root, c_root = temporary / "B", temporary / "C"
        _checkout_bundle(git, bundle_parent, b_root, ACCEPTED_PARENT_COMMIT)
        _checkout_bundle(git, bundle_parent, c_root, ACCEPTED_PARENT_COMMIT)
        replacements = {"cli/cmd/search.go": search}
        if config.candidate_search_test_go is not None:
            base_test = c_root / "cli/cmd/search_test.go"
            if not base_test.is_file():
                raise CandidateAdapterError("candidate_input_invalid")
            replacements["cli/cmd/search_test.go"] = validate_candidate_source(config.candidate_search_test_go, sha256_file(base_test), terms)
        before: dict[str, str] = {}
        after: dict[str, str] = {}
        for relative, replacement in replacements.items():
            destination = c_root / relative
            before[relative] = sha256_file(destination, error="candidate_input_invalid")
            destination.write_bytes(read_regular(replacement.path, mode=0o600, cap=MAX_SOURCE_BYTES, error="candidate_input_invalid"))
            destination.chmod(0o644)
            after[relative] = sha256_file(destination, error="candidate_input_invalid")
        _git(git, c_root, "add", "--", *replacements)
        _git(git, c_root, "commit", "-q", "--no-gpg-sign", "-m", "protocol-v7 candidate direct-hit containment owner substitution")
        candidate_commit = _git(git, c_root, "rev-parse", "HEAD").decode().strip()
        candidate_tree = _git(git, c_root, "rev-parse", "HEAD^{tree}").decode().strip()
        if _git(git, c_root, "rev-list", "--parents", "-n", "1", candidate_commit).decode().strip() != f"{candidate_commit} {ACCEPTED_PARENT_COMMIT}" or _git(git, b_root, "status", "--porcelain=v1", "--untracked-files=all") or _git(git, c_root, "status", "--porcelain=v1", "--untracked-files=all"):
            raise CandidateAdapterError("candidate_delta_invalid")
        range_arg = f"{ACCEPTED_PARENT_COMMIT}..{candidate_commit}"
        diff = _git(git, c_root, "diff", "--binary", "--full-index", "--no-ext-diff", "--no-textconv", "--no-renames", range_arg)
        names = _git(git, c_root, "diff", "--name-status", "--no-renames", range_arg)
        allowed = list(replacements)
        expected_names = "".join(f"M\t{path}\n" for path in allowed).encode()
        if names != expected_names or _git(git, c_root, "diff", "--summary", range_arg):
            raise CandidateAdapterError("candidate_delta_invalid")
        ref = "refs/c3-eval/commit-pool/protocol-v7-candidate-direct-hit-containment-owner-substitution"
        _git(git, c_root, "update-ref", ref, candidate_commit)
        bundle = temporary / "source.bundle"
        _git(git, c_root, "bundle", "create", str(bundle), ref)
        heads = _git(git, c_root, "bundle", "list-heads", str(bundle))
        runtime = temporary / "candidate-runtime"
        _build_runtime(go, c_root, runtime)
        runtime_hash = sha256_file(runtime)
        delta = {
            "variable": REGISTERED_VARIABLE,
            "baseline_commit": ACCEPTED_PARENT_COMMIT, "baseline_tree": ACCEPTED_PARENT_TREE,
            "candidate_commit": candidate_commit, "candidate_tree": candidate_tree,
            "diff_sha256": sha256_bytes(diff), "name_status_sha256": sha256_bytes(names),
            "name_status": [line for line in names.decode().splitlines() if line], "allowed_paths": allowed,
            "before_blob_sha256": before, "after_blob_sha256": after,
            "bundle_sha256": sha256_file(bundle), "bundle_heads_sha256": sha256_bytes(heads),
        }
        delta_hash = candidate_delta_sha256(delta)
        capsule = _capsule_candidate(parent.authority["controller_source_capsule"], candidate_commit, candidate_tree, after)
        authority = build_candidate_authority(parent.authority, capsule, runtime_hash, delta)
        authority_bytes = canonical_json(authority)
        if privacy_hit(authority_bytes, terms):
            raise CandidateAdapterError("privacy_violation")
        authority_path = temporary / "candidate-authority.v4.json"
        authority_path.write_bytes(authority_bytes)
        authority_path.chmod(0o600)
        controller = config.parent_baseline_root / "freeze/controller-runtime"
        (preflight or protocol_v7_preflight)(
            go=go, git=git, authority=authority_path, controller=controller, runtime=runtime,
            b_root=b_root, c_root=c_root, bundle=bundle, policy=config.privacy_policy,
            parent=parent, temporary=temporary,
        )
        stage = config.capability_root.parent / (".c3-v7-candidate-stage-" + sha256_bytes(str(config.capability_root).encode())[:16])
        if stage.exists() or stage.is_symlink():
            raise CandidateAdapterError("stage_residue")
        try:
            (stage / "freeze").mkdir(parents=True, mode=0o700)
            stage.chmod(0o700)
            for source, relative in ((authority_path, "candidate-authority.v4.json"), (controller, "freeze/controller-runtime"), (runtime, "freeze/candidate-runtime"), (bundle, "freeze/source.bundle")):
                destination = stage / relative
                shutil.copyfile(source, destination)
                destination.chmod(0o600)
            hashes = _manifest_snapshot(stage, CAPABILITY_FILES - {"capability-manifest.json"})
            manifest = {
                "$schema": CAPABILITY_SCHEMA, "status": "accepted_unexecuted", "effect_claim": False,
                "candidate_execution_authorized": False, "registered_variable": REGISTERED_VARIABLE,
                "candidate_delta_sha256": delta_hash, "candidate_commit": candidate_commit, "candidate_tree": candidate_tree,
                "parent_authority_sha256": ACCEPTED_PARENT_AUTHORITY_SHA256,
                "parent_output_sha256": ACCEPTED_PARENT_OUTPUT_SHA256,
                "parent_validator_record_hash": ACCEPTED_PARENT_VALIDATOR_RECORD_HASH,
                "parent_validator_payload_sha256": ACCEPTED_PARENT_VALIDATOR_PAYLOAD_SHA256,
                "privacy_policy_sha256": policy_hash, "adapter_sha256": _adapter_sha256(),
                "validator_sha256": _validator_sha256(), "files": hashes,
                "future_output_root_sha256": sha256_bytes(str(config.output_root.absolute()).encode()),
            }
            manifest_path = stage / "capability-manifest.json"
            manifest_path.write_bytes(canonical_json(manifest))
            manifest_path.chmod(0o600)
            _manifest_snapshot(stage, CAPABILITY_FILES)
            os.replace(stage, config.capability_root)
            directory_fd = os.open(config.capability_root.parent, os.O_RDONLY | os.O_DIRECTORY)
            try:
                os.fsync(directory_fd)
            finally:
                os.close(directory_fd)
        except BaseException:
            shutil.rmtree(stage, ignore_errors=True)
            shutil.rmtree(config.capability_root, ignore_errors=True)
            raise
    if _relative_manifest(config.parent_baseline_root) != parent_before or _governed_input_snapshot(config, go, git) != governed_before:
        shutil.rmtree(config.capability_root, ignore_errors=True)
        raise CandidateAdapterError("parent_changed")
    result = {
        "$schema": RESULT_SCHEMA, "status": "accepted_unexecuted", "effect_claim": False,
        "candidate_execution_authorized": False, "candidate_delta_sha256": delta_hash,
        "candidate_authority_sha256": sha256_bytes(authority_bytes), "candidate_runtime_sha256": runtime_hash,
        "bundle_sha256": delta["bundle_sha256"], "capability_manifest_sha256": sha256_file(config.capability_root / "capability-manifest.json"),
    }
    if not is_generic_result(result):
        shutil.rmtree(config.capability_root, ignore_errors=True)
        raise CandidateAdapterError("result_invalid")
    return result


PREFLIGHT_HARNESS = r'''package main

import (
    "bytes"
    "encoding/json"
    "os"
    "testing"
)

type candidatePreflightConfig struct {
    Authority string `json:"authority"`; Controller string `json:"controller"`; Runtime string `json:"runtime"`
    BRoot string `json:"b_root"`; CRoot string `json:"c_root"`; Bundle string `json:"bundle"`; Policy string `json:"policy"`
    ParentRoot string `json:"parent_root"`; ParentAuthority string `json:"parent_authority"`; ParentOutput string `json:"parent_output"`; ParentStore string `json:"parent_store"`
}

func TestProtocolV7CandidatePreflight(t *testing.T) {
    data, err := os.ReadFile(os.Getenv("C3_PROTOCOL_V7_CANDIDATE_PREFLIGHT_CONFIG")); if err != nil { t.Fatal("config_invalid") }
    var cfg candidatePreflightConfig; if decodeStrictBytes(data, &cfg) != nil { t.Fatal("config_invalid") }
    authorityBytes, err := os.ReadFile(cfg.Authority); if err != nil { t.Fatal("authority_invalid") }
    var authority controllerAuthorityV4; if decodeStrictBytes(authorityBytes, &authority) != nil { t.Fatal("authority_invalid") }
    fixtures := cfg.BRoot + "/research/eval/structural-retrieval/fixtures.dev.v2.jsonl"
    benchmark := cfg.BRoot + "/research/eval/structural-retrieval/benchmark.v2.json"
    scorer := cfg.BRoot + "/cli/tools/structural-search-eval-v2/main.go"
    fixtureRows, fixtureHash, err := loadFixtures(fixtures); if err != nil { t.Fatal("fixture_invalid") }
    bench, err := loadBenchmark(benchmark); if err != nil || bench.FixtureSHA256 != fixtureHash || bench.FixtureCount != len(fixtureRows) { t.Fatal("benchmark_invalid") }
    policyBytes, err := os.ReadFile(cfg.Policy); if err != nil { t.Fatal("privacy_policy_invalid") }
    policy, err := decodePrivacyPolicy(bytes.NewReader(policyBytes)); if err != nil { t.Fatal("privacy_policy_invalid") }
    scanner, err := newPrivacyScanner(policy); if err != nil { t.Fatal("privacy_policy_invalid") }
    parent := parentBaselineFiles{Root: cfg.ParentRoot, Authority: cfg.ParentAuthority, Output: cfg.ParentOutput, ValidatorStore: cfg.ParentStore}
    if err := verifyControllerAuthorityV4(authority, authorityBytes, cfg.Runtime, cfg.Controller, cfg.BRoot, cfg.CRoot, cfg.Bundle, cfg.Policy, fixtures, benchmark, scorer, bench, policy, scanner, parent); err != nil { t.Fatal(err) }
    result, _ := json.Marshal(map[string]any{"status":"accepted", "$schema":"structural-retrieval-protocol-v7-candidate-preflight.v1"})
    if err := os.WriteFile(os.Getenv("C3_PROTOCOL_V7_CANDIDATE_PREFLIGHT_RESULT"), result, 0o600); err != nil { t.Fatal("result_invalid") }
}
'''


def _copy_module_tree(source: Path, target: Path) -> None:
    for path in sorted(source.rglob("*")):
        relative = path.relative_to(source)
        destination = target / relative
        info = path.lstat()
        if stat.S_ISLNK(info.st_mode):
            raise CandidateAdapterError("harness_source_invalid")
        if stat.S_ISDIR(info.st_mode):
            destination.mkdir(parents=True, exist_ok=True)
        elif stat.S_ISREG(info.st_mode):
            destination.parent.mkdir(parents=True, exist_ok=True)
            shutil.copyfile(path, destination)
            destination.chmod(0o644)
        else:
            raise CandidateAdapterError("harness_source_invalid")


def protocol_v7_preflight(*, go: Path, git: Path, authority: Path, controller: Path, runtime: Path, b_root: Path, c_root: Path, bundle: Path, policy: Path, parent: ParentBinding, temporary: Path) -> None:
    harness_root = temporary / "harness"
    # The verifier compiles only accepted B. Candidate C is inspected as data.
    _copy_module_tree(b_root / "cli", harness_root / "cli")
    harness = harness_root / "cli/tools/structural-search-eval-v2/protocol_v7_candidate_preflight_test.go"
    harness.write_text(PREFLIGHT_HARNESS, encoding="utf-8")
    config = {
        "authority": str(authority), "controller": str(controller), "runtime": str(runtime),
        "b_root": str(b_root), "c_root": str(c_root), "bundle": str(bundle), "policy": str(policy),
        "parent_root": str(parent.root / "parent"),
        "parent_authority": str(parent.root / "parent/controller-authority.v4.json"),
        "parent_output": str(parent.root / "parent/controller-output.v4.json"),
        "parent_store": str(parent.validator_store),
    }
    config_path = temporary / "preflight.json"
    config_path.write_bytes(canonical_json(config)); config_path.chmod(0o600)
    result_path = temporary / "preflight-result.json"
    tool_root = temporary / "tools"
    tool_root.mkdir(mode=0o700)
    for source, name, digest in (
        (go, "go", ACCEPTED_GO_EXECUTABLE_SHA256),
        (git, "git", ACCEPTED_GIT_EXECUTABLE_SHA256),
    ):
        target = tool_root / name
        shutil.copyfile(source, target)
        target.chmod(0o700)
        if sha256_file(target, error="tool_identity_invalid") != digest:
            raise CandidateAdapterError("tool_identity_invalid")
    home = Path.home()
    env = {
        "PATH": str(tool_root), "HOME": str(home), "TMPDIR": str(temporary),
        "GOROOT": str(go.parent.parent),
        "LC_ALL": "C", "LANG": "C", "TZ": "UTC",
        "GOENV": "off", "GOWORK": "off", "GOFLAGS": "", "GOTOOLCHAIN": "local", "CGO_ENABLED": "0",
        "GOPROXY": "off", "GOSUMDB": "off", "GONOSUMDB": "", "GOPRIVATE": "",
        "GOMODCACHE": str(home / "go/pkg/mod"),
        "GOCACHE": str(Path(os.environ.get("XDG_CACHE_HOME", str(home / ".cache"))) / "go-build"),
        "GIT_CONFIG_NOSYSTEM": "1", "GIT_CONFIG_GLOBAL": "/dev/null", "GIT_OPTIONAL_LOCKS": "0",
        "GIT_NO_REPLACE_OBJECTS": "1", "GIT_ALTERNATE_OBJECT_DIRECTORIES": "", "GIT_ATTR_NOSYSTEM": "1",
        "C3_PROTOCOL_V7_CANDIDATE_PREFLIGHT_CONFIG": str(config_path),
        "C3_PROTOCOL_V7_CANDIDATE_PREFLIGHT_RESULT": str(result_path),
    }
    completed = run_bounded_process(
        [str(tool_root / "go"), "test", "-count=1", "-run", "^TestProtocolV7CandidatePreflight$", "./tools/structural-search-eval-v2"],
        cwd=harness_root / "cli", env=env, timeout=600, output_cap=PROCESS_OUTPUT_CAP,
    )
    if completed.returncode != 0 or completed.stderr:
        raise CandidateAdapterError("candidate_preflight_failed")
    result = decode_canonical(read_regular(result_path, mode=0o600, cap=MAX_METADATA_BYTES, error="candidate_preflight_failed"), "candidate_preflight_failed")
    if result != {"$schema": "structural-retrieval-protocol-v7-candidate-preflight.v1", "status": "accepted"}:
        raise CandidateAdapterError("candidate_preflight_failed")


def validate_capability(config: CandidateConfig, *, preflight: Callable[..., None] | None = None) -> dict[str, Any]:
    parent = validate_parent(config)
    policy_hash, _ = validate_privacy_policy(config.privacy_policy)
    go = _resolve_tool(config.go_executable, "go", ACCEPTED_GO_EXECUTABLE_SHA256)
    git = _resolve_tool(config.git_executable, "git", ACCEPTED_GIT_EXECUTABLE_SHA256)
    parent_before = _relative_manifest(config.parent_baseline_root)
    external_before = {
        "policy": sha256_file(config.privacy_policy, error="governed_input_invalid"),
        "go": sha256_file(go, error="tool_identity_invalid"),
        "git": sha256_file(git, error="tool_identity_invalid"),
        "adapter": _adapter_sha256(),
        "validator": _validator_sha256(),
    }
    require_directory(config.capability_root, 0o700, "capability_layout_invalid")
    files = _manifest_snapshot(config.capability_root, CAPABILITY_FILES)
    manifest = decode_canonical(read_regular(config.capability_root / "capability-manifest.json", mode=0o600, cap=MAX_METADATA_BYTES, error="capability_manifest_invalid"), "capability_manifest_invalid")
    embedded = {key: value for key, value in files.items() if key != "capability-manifest.json"}
    if set(manifest) != CAPABILITY_MANIFEST_KEYS or manifest.get("$schema") != CAPABILITY_SCHEMA or manifest.get("status") != "accepted_unexecuted" or manifest.get("effect_claim") is not False or manifest.get("candidate_execution_authorized") is not False or manifest.get("registered_variable") != REGISTERED_VARIABLE or manifest.get("files") != embedded or manifest.get("privacy_policy_sha256") != policy_hash or manifest.get("parent_authority_sha256") != ACCEPTED_PARENT_AUTHORITY_SHA256 or manifest.get("parent_output_sha256") != ACCEPTED_PARENT_OUTPUT_SHA256 or manifest.get("parent_validator_record_hash") != ACCEPTED_PARENT_VALIDATOR_RECORD_HASH or manifest.get("parent_validator_payload_sha256") != ACCEPTED_PARENT_VALIDATOR_PAYLOAD_SHA256 or manifest.get("adapter_sha256") != _adapter_sha256() or manifest.get("validator_sha256") != _validator_sha256() or manifest.get("future_output_root_sha256") != sha256_bytes(str(config.output_root.absolute()).encode()) or any(not valid_sha256(value) for value in embedded.values()):
        raise CandidateAdapterError("capability_manifest_invalid")
    if config.output_root.exists() or config.output_root.is_symlink():
        raise CandidateAdapterError("governed_path_invalid")
    with tempfile.TemporaryDirectory(prefix="c3-v7-candidate-validate-") as raw:
        temporary = Path(raw); temporary.chmod(0o700)
        b_root, c_root = temporary / "B", temporary / "C"
        bundle = config.capability_root / "freeze/source.bundle"
        authority_path = config.capability_root / "candidate-authority.v4.json"
        authority = decode_canonical(read_regular(authority_path, mode=0o600, cap=MAX_METADATA_BYTES, error="candidate_authority_invalid"), "candidate_authority_invalid")
        delta = authority.get("candidate_delta")
        if not isinstance(delta, dict) or candidate_delta_sha256(delta) != manifest.get("candidate_delta_sha256"):
            raise CandidateAdapterError("candidate_authority_invalid")
        validate_manifest_authority_bindings(manifest, authority)
        _checkout_bundle(git, bundle, b_root, ACCEPTED_PARENT_COMMIT)
        _checkout_bundle(git, bundle, c_root, delta["candidate_commit"])
        (preflight or protocol_v7_preflight)(go=go, git=git, authority=authority_path, controller=config.capability_root / "freeze/controller-runtime", runtime=config.capability_root / "freeze/candidate-runtime", b_root=b_root, c_root=c_root, bundle=bundle, policy=config.privacy_policy, parent=parent, temporary=temporary)
    external_after = {
        "policy": sha256_file(config.privacy_policy, error="governed_input_invalid"),
        "go": sha256_file(go, error="tool_identity_invalid"),
        "git": sha256_file(git, error="tool_identity_invalid"),
        "adapter": _adapter_sha256(),
        "validator": _validator_sha256(),
    }
    if external_after != external_before or _relative_manifest(config.parent_baseline_root) != parent_before or _manifest_snapshot(config.capability_root, CAPABILITY_FILES) != files:
        raise CandidateAdapterError("governed_input_changed")
    return {"$schema": RESULT_SCHEMA, "status": "accepted_unexecuted", "effect_claim": False, "candidate_execution_authorized": False, "capability_manifest_sha256": files["capability-manifest.json"], "candidate_delta_sha256": manifest["candidate_delta_sha256"]}


def _path_sha256(path: Path) -> str:
    return sha256_bytes(str(path.absolute()).encode())


def validate_one_shot(
    authorization_path: Path,
    activation_path: Path,
    binding: OneShotBinding,
    *,
    activation_bytes: bytes | None = None,
) -> None:
    auth_bytes = read_regular(authorization_path, mode=0o600, cap=64 << 10, error="authorization_invalid")
    if not auth_bytes.endswith(b"\n") or auth_bytes.count(b"\n") != 1 or sha256_bytes(auth_bytes) != binding.authorization_sha256:
        raise CandidateAdapterError("authorization_invalid")
    record = decode_canonical(auth_bytes[:-1], "authorization_invalid")
    expected_payload = {
        "$schema": AUTHORIZATION_SCHEMA, "candidate_execution_authorized": True, "effect_claim": False,
        "max_capture_count": 1, "registered_variable": REGISTERED_VARIABLE,
        "activation_path_sha256": _path_sha256(activation_path), "output_root_sha256": _path_sha256(binding.output_root),
        "capability_manifest_sha256": binding.capability_manifest_sha256,
        "candidate_authority_sha256": binding.candidate_authority_sha256,
        "candidate_delta_sha256": binding.candidate_delta_sha256, "candidate_runtime_sha256": binding.candidate_runtime_sha256,
        "bundle_sha256": binding.bundle_sha256, "privacy_policy_sha256": binding.privacy_policy_sha256,
        "adapter_sha256": binding.adapter_sha256, "validator_sha256": binding.validator_sha256, "verdict": "authorized",
    }
    if record.get("seq") != 1 or record.get("prev_hash") != "GENESIS" or record.get("payload") != expected_payload or record.get("payload_sha256") != sha256_bytes(canonical_json(expected_payload)):
        raise CandidateAdapterError("authorization_invalid")
    without = {key: value for key, value in record.items() if key != "record_hash"}
    if record.get("record_hash") != sha256_bytes(canonical_json(without)) or record.get("record_hash") != binding.authorization_record_hash:
        raise CandidateAdapterError("authorization_invalid")
    if activation_bytes is None:
        activation_bytes = read_regular(activation_path, mode=0o600, cap=64 << 10, error="activation_invalid")
    expected_activation = {
        "$schema": ACTIVATION_SCHEMA, "activation_path_sha256": _path_sha256(activation_path),
        "authorization_record_hash": binding.authorization_record_hash,
        "authorization_record_path_sha256": _path_sha256(authorization_path),
        "candidate_execution_authorized": True, "effect_claim": False, "max_capture_count": 1,
        "registered_variable": REGISTERED_VARIABLE, "output_root_sha256": _path_sha256(binding.output_root),
        "capability_manifest_sha256": binding.capability_manifest_sha256,
        "candidate_authority_sha256": binding.candidate_authority_sha256,
        "candidate_delta_sha256": binding.candidate_delta_sha256, "candidate_runtime_sha256": binding.candidate_runtime_sha256,
        "bundle_sha256": binding.bundle_sha256, "privacy_policy_sha256": binding.privacy_policy_sha256,
        "adapter_sha256": binding.adapter_sha256, "validator_sha256": binding.validator_sha256, "verdict": "authorized",
    }
    if activation_bytes != canonical_json(expected_activation) or sha256_bytes(activation_bytes) != binding.activation_sha256:
        raise CandidateAdapterError("activation_invalid")


def _read_pinned_activation(descriptor: int) -> bytes:
    info = os.fstat(descriptor)
    if not stat.S_ISREG(info.st_mode) or stat.S_IMODE(info.st_mode) != 0o600 or info.st_size > 64 << 10:
        raise CandidateAdapterError("activation_invalid")
    os.lseek(descriptor, 0, os.SEEK_SET)
    data = bytearray()
    while len(data) <= 64 << 10:
        chunk = os.read(descriptor, min(64 << 10, (64 << 10) + 1 - len(data)))
        if not chunk:
            break
        data.extend(chunk)
    after = os.fstat(descriptor)
    if len(data) != info.st_size or len(data) > 64 << 10 or (info.st_dev, info.st_ino, info.st_mode, info.st_size, info.st_mtime_ns) != (after.st_dev, after.st_ino, after.st_mode, after.st_size, after.st_mtime_ns):
        raise CandidateAdapterError("activation_invalid")
    return bytes(data)


def _link_pinned_descriptor(descriptor: int, receipt: Path) -> None:
    libc = ctypes.CDLL(None, use_errno=True)
    linkat = libc.linkat
    linkat.argtypes = [ctypes.c_int, ctypes.c_char_p, ctypes.c_int, ctypes.c_char_p, ctypes.c_int]
    linkat.restype = ctypes.c_int
    if linkat(descriptor, b"", -100, os.fsencode(receipt), 0x1000) != 0:  # AT_FDCWD, AT_EMPTY_PATH
        raise OSError(ctypes.get_errno(), "linkat")


def consume_candidate_activation(
    activation: Path,
    authorization: Path,
    binding: OneShotBinding,
    *,
    transition_hook: Callable[[], None] | None = None,
) -> ConsumedActivation:
    receipt = Path(str(activation) + ".consumed")
    receipt_created = False
    try:
        descriptor = os.open(activation, os.O_RDONLY | os.O_NOFOLLOW)
        info = os.fstat(descriptor)
        pinned_bytes = _read_pinned_activation(descriptor)
        validate_one_shot(authorization, activation, binding, activation_bytes=pinned_bytes)
        if transition_hook is not None:
            transition_hook()
        if receipt.exists() or receipt.is_symlink():
            raise CandidateAdapterError("activation_invalid")
        current = activation.lstat()
        if stat.S_ISLNK(current.st_mode) or (current.st_dev, current.st_ino) != (info.st_dev, info.st_ino):
            raise CandidateAdapterError("activation_invalid")
        _link_pinned_descriptor(descriptor, receipt)
        receipt_created = True
        receipt_info = receipt.lstat()
        if stat.S_ISLNK(receipt_info.st_mode) or stat.S_IMODE(receipt_info.st_mode) != 0o600 or (receipt_info.st_dev, receipt_info.st_ino) != (info.st_dev, info.st_ino) or _read_pinned_activation(descriptor) != pinned_bytes:
            raise CandidateAdapterError("activation_invalid")
        os.fsync(descriptor)
        os.unlink(activation)
        parent_fd = os.open(activation.parent, os.O_RDONLY | os.O_DIRECTORY)
        try:
            os.fsync(parent_fd)
        finally:
            os.close(parent_fd)
        return ConsumedActivation(descriptor, receipt, binding)
    except BaseException as exc:
        if receipt_created:
            try:
                os.unlink(receipt)
                parent_fd = os.open(activation.parent, os.O_RDONLY | os.O_DIRECTORY)
                try:
                    os.fsync(parent_fd)
                finally:
                    os.close(parent_fd)
            except OSError:
                pass
        if "descriptor" in locals():
            os.close(descriptor)
        if isinstance(exc, CandidateAdapterError):
            raise
        raise CandidateAdapterError("activation_invalid") from exc


def make_test_one_shot(authorization: Path, activation: Path, output: Path) -> OneShotBinding:
    zero = "1" * 64
    payload = {
        "$schema": AUTHORIZATION_SCHEMA, "candidate_execution_authorized": True, "effect_claim": False,
        "max_capture_count": 1, "registered_variable": REGISTERED_VARIABLE,
        "activation_path_sha256": _path_sha256(activation), "output_root_sha256": _path_sha256(output),
        "capability_manifest_sha256": zero, "candidate_authority_sha256": zero,
        "candidate_delta_sha256": zero, "candidate_runtime_sha256": zero, "bundle_sha256": zero,
        "privacy_policy_sha256": zero, "adapter_sha256": zero, "validator_sha256": zero, "verdict": "authorized",
    }
    record = {"seq": 1, "recorded_at": "2026-01-01T00:00:00Z", "prev_hash": "GENESIS", "payload_sha256": sha256_bytes(canonical_json(payload)), "payload": payload}
    record["record_hash"] = sha256_bytes(canonical_json(record))
    auth_bytes = canonical_json(record) + b"\n"
    authorization.write_bytes(auth_bytes); authorization.chmod(0o600)
    activation_value = {
        "$schema": ACTIVATION_SCHEMA, "activation_path_sha256": _path_sha256(activation),
        "authorization_record_hash": record["record_hash"], "authorization_record_path_sha256": _path_sha256(authorization),
        "candidate_execution_authorized": True, "effect_claim": False, "max_capture_count": 1,
        "registered_variable": REGISTERED_VARIABLE, "output_root_sha256": _path_sha256(output),
        "capability_manifest_sha256": zero, "candidate_authority_sha256": zero,
        "candidate_delta_sha256": zero, "candidate_runtime_sha256": zero, "bundle_sha256": zero,
        "privacy_policy_sha256": zero, "adapter_sha256": zero, "validator_sha256": zero, "verdict": "authorized",
    }
    activation_bytes = canonical_json(activation_value)
    activation.write_bytes(activation_bytes); activation.chmod(0o600)
    return OneShotBinding(authorization, sha256_bytes(auth_bytes), record["record_hash"], activation, sha256_bytes(activation_bytes), output, zero, zero, zero, zero, zero, zero, zero, zero)


def binding_from_capability(config: CandidateConfig) -> OneShotBinding:
    if config.activation is None or config.authorization_record is None:
        raise CandidateAdapterError("activation_invalid")
    validate_capability(config)
    manifest_path = config.capability_root / "capability-manifest.json"
    manifest = decode_canonical(read_regular(manifest_path, mode=0o600, cap=MAX_METADATA_BYTES, error="capability_manifest_invalid"), "capability_manifest_invalid")
    authorization_bytes = read_regular(config.authorization_record, mode=0o600, cap=64 << 10, error="authorization_invalid")
    if not authorization_bytes.endswith(b"\n"):
        raise CandidateAdapterError("authorization_invalid")
    record = decode_canonical(authorization_bytes[:-1], "authorization_invalid")
    activation_bytes = read_regular(config.activation, mode=0o600, cap=64 << 10, error="activation_invalid")
    return OneShotBinding(
        config.authorization_record, sha256_bytes(authorization_bytes), record.get("record_hash", ""), config.activation,
        sha256_bytes(activation_bytes), config.output_root, sha256_file(manifest_path),
        sha256_file(config.capability_root / "candidate-authority.v4.json"), manifest["candidate_delta_sha256"],
        sha256_file(config.capability_root / "freeze/candidate-runtime"), sha256_file(config.capability_root / "freeze/source.bundle"),
        manifest["privacy_policy_sha256"], manifest["adapter_sha256"], manifest["validator_sha256"],
    )


def _kill_group(process: subprocess.Popen[bytes]) -> None:
    try:
        os.killpg(process.pid, signal.SIGKILL)
    except ProcessLookupError:
        pass
    try:
        process.wait(timeout=5)
    except (subprocess.TimeoutExpired, ChildProcessError):
        pass


def run_bounded_process(command: list[str], *, cwd: Path, env: dict[str, str], timeout: int = PROCESS_TIMEOUT_SECONDS, output_cap: int = PROCESS_OUTPUT_CAP) -> ProcessResult:
    if not command or not Path(command[0]).is_absolute() or output_cap <= 0 or timeout <= 0:
        raise CandidateAdapterError("candidate_command_invalid")
    try:
        process = subprocess.Popen(command, cwd=cwd, env=env, stdin=subprocess.DEVNULL, stdout=subprocess.PIPE, stderr=subprocess.PIPE, start_new_session=True)
    except OSError as exc:
        raise CandidateAdapterError("candidate_execution_failed") from exc
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
                raise CandidateAdapterError("candidate_timeout")
            events = selector.select(min(remaining, 0.25))
            if not events and process.poll() is not None:
                events = [(key, selectors.EVENT_READ) for key in tuple(selector.get_map().values())]
            for key, _ in events:
                chunk = os.read(key.fileobj.fileno(), min(64 << 10, output_cap + 1))
                if not chunk:
                    selector.unregister(key.fileobj)
                    continue
                output[key.data].extend(chunk)
                if len(output["stdout"]) + len(output["stderr"]) > output_cap:
                    raise CandidateAdapterError("candidate_output_cap")
        remaining = max(1, int(timeout - (time.monotonic() - started)))
        returncode = process.wait(timeout=remaining)
        return ProcessResult(returncode, bytes(output["stdout"]), bytes(output["stderr"]))
    except BaseException:
        _kill_group(process)
        raise
    finally:
        selector.close()
        process.stdout.close()
        process.stderr.close()


def _capture_environment(tool_root: Path, go: Path, temporary: Path) -> dict[str, str]:
    home = Path.home()
    return {
        "PATH": str(tool_root), "HOME": str(home), "TMPDIR": str(temporary), "GOROOT": str(go.parent.parent),
        "LC_ALL": "C", "LANG": "C", "TZ": "UTC",
        "GOENV": "off", "GOWORK": "off", "GOFLAGS": "", "GOTOOLCHAIN": "local", "CGO_ENABLED": "0",
        "GOPROXY": "off", "GOSUMDB": "off", "GONOSUMDB": "", "GOPRIVATE": "",
        "GOMODCACHE": str(home / "go/pkg/mod"),
        "GOCACHE": str(Path(os.environ.get("XDG_CACHE_HOME", str(home / ".cache"))) / "go-build"),
        "GIT_CONFIG_NOSYSTEM": "1", "GIT_CONFIG_GLOBAL": "/dev/null", "GIT_OPTIONAL_LOCKS": "0",
        "GIT_NO_REPLACE_OBJECTS": "1", "GIT_ALTERNATE_OBJECT_DIRECTORIES": "", "GIT_ATTR_NOSYSTEM": "1",
    }


def _execute_candidate_controller(command: list[str], *, cwd: Path, tool_root: Path, go: Path, temporary: Path, runner: Runner) -> ProcessResult:
    return runner(
        command,
        cwd=cwd,
        env=_capture_environment(tool_root, go, temporary),
        timeout=PROCESS_TIMEOUT_SECONDS,
        output_cap=PROCESS_OUTPUT_CAP,
    )


def capture_candidate(config: CandidateConfig, *, runner: Runner = run_bounded_process) -> dict[str, Any]:
    binding = binding_from_capability(config)
    stage = config.output_root.parent / (".c3-v7-candidate-capture-" + sha256_bytes(str(config.output_root).encode())[:16])
    try:
        if stage.exists() or config.output_root.exists():
            raise CandidateAdapterError("stage_residue")
        with tempfile.TemporaryDirectory(prefix="c3-v7-candidate-run-") as raw:
            temporary = Path(raw); temporary.chmod(0o700)
            snapshot = create_execution_snapshot(config, temporary)
            try:
                verify_execution_snapshot_bindings(snapshot, binding, config)
                consumed = consume_candidate_activation(config.activation, config.authorization_record, binding)  # type: ignore[arg-type]
                try:
                    host_go = _resolve_tool(config.go_executable, "go", ACCEPTED_GO_EXECUTABLE_SHA256)
                    authority = decode_canonical(read_regular(snapshot.capability_root / "candidate-authority.v4.json", mode=None, cap=MAX_METADATA_BYTES, error="candidate_authority_invalid"), "candidate_authority_invalid")
                    manifest = decode_canonical(read_regular(snapshot.capability_root / "capability-manifest.json", mode=None, cap=MAX_METADATA_BYTES, error="capability_manifest_invalid"), "capability_manifest_invalid")
                    validate_manifest_authority_bindings(manifest, authority)
                    delta = authority["candidate_delta"]
                    b_root, c_root = temporary / "B", temporary / "C"
                    bundle = snapshot.capability_root / "freeze/source.bundle"
                    _checkout_bundle(snapshot.git, bundle, b_root, ACCEPTED_PARENT_COMMIT)
                    _checkout_bundle(snapshot.git, bundle, c_root, delta["candidate_commit"])
                    work = temporary / "work"
                    output = stage / "candidate"
                    stage.mkdir(mode=0o700)
                    command = [str(snapshot.capability_root / "freeze/controller-runtime"), "--controller", "--runtime", str(snapshot.capability_root / "freeze/candidate-runtime"), "--fixtures", str(b_root / "research/eval/structural-retrieval/fixtures.dev.v2.jsonl"), "--benchmark", str(b_root / "research/eval/structural-retrieval/benchmark.v2.json"), "--work-root", str(work), "--authority", str(snapshot.capability_root / "candidate-authority.v4.json"), "--controller-source-root", str(b_root), "--runtime-source-root", str(c_root), "--bundle", str(bundle), "--privacy-policy", str(snapshot.privacy_policy), "--scorer-source", str(b_root / "cli/tools/structural-search-eval-v2/main.go"), "--output-dir", str(output), "--parent-baseline-root", str(snapshot.parent_baseline_root / "parent"), "--parent-baseline-authority", str(snapshot.parent_baseline_root / "parent/controller-authority.v4.json"), "--parent-baseline-output", str(snapshot.parent_baseline_root / "parent/controller-output.v4.json"), "--parent-baseline-validator-store", str(snapshot.parent_validator_store)]
                    completed = _execute_candidate_controller(command, cwd=temporary, tool_root=snapshot.tool_root, go=host_go, temporary=temporary, runner=runner)
                    if completed.returncode != 0 or completed.stderr or not completed.stdout.endswith(b"\n") or completed.stdout.count(b"\n") != 1:
                        raise CandidateAdapterError("candidate_execution_failed")
                    controller_result = decode_strict_object(completed.stdout[:-1], "candidate_execution_failed")
                    if controller_result.get("$schema") != "structural-retrieval-controller-output.v4":
                        raise CandidateAdapterError("candidate_execution_failed")
                    verify_execution_snapshot(snapshot)
                    freeze = stage / "freeze"; freeze.mkdir(mode=0o700)
                    for source, name in ((snapshot.capability_root / "capability-manifest.json", "capability-manifest.json"), (snapshot.capability_root / "candidate-authority.v4.json", "candidate-authority.v4.json"), (snapshot.capability_root / "freeze/controller-runtime", "controller-runtime"), (snapshot.capability_root / "freeze/candidate-runtime", "candidate-runtime"), (bundle, "source.bundle")):
                        target = freeze / name; target.write_bytes(read_regular(source, mode=None, cap=256 << 20, error="execution_snapshot_changed")); target.chmod(0o600)
                    capture_manifest = {
                        "$schema": CAPTURE_MANIFEST_SCHEMA, "status": "captured_unvalidated", "effect_claim": False,
                        "candidate_execution_authorized": True, "max_capture_count": 1,
                        "candidate_delta_sha256": binding.candidate_delta_sha256,
                        "candidate_authority_sha256": binding.candidate_authority_sha256,
                        "capability_manifest_sha256": binding.capability_manifest_sha256,
                        "adapter_sha256": binding.adapter_sha256, "validator_sha256": binding.validator_sha256,
                        "main_sha256": ACCEPTED_MAIN_SHA256, "main_test_sha256": ACCEPTED_MAIN_TEST_SHA256,
                        "scorer_region_sha256": ACCEPTED_SCORER_REGION_SHA256,
                        "parent_authority_sha256": ACCEPTED_PARENT_AUTHORITY_SHA256,
                        "parent_output_sha256": ACCEPTED_PARENT_OUTPUT_SHA256,
                        "parent_validator_record_hash": ACCEPTED_PARENT_VALIDATOR_RECORD_HASH,
                        "parent_cleanup_record_hash": ACCEPTED_PARENT_CLEANUP_RECORD_HASH,
                        "privacy_policy_sha256": binding.privacy_policy_sha256,
                        "authorization_record_sha256": binding.authorization_sha256,
                        "activation_proof_sha256": binding.activation_sha256,
                        "controller_output_sha256": sha256_file(output / "controller-output.v4.json"),
                        "bundle_sha256": binding.bundle_sha256, "candidate_runtime_sha256": binding.candidate_runtime_sha256,
                        "run_count": len(decode_canonical(read_regular(output / "controller-output.v4.json", mode=0o600, cap=MAX_METADATA_BYTES, error="candidate_output_invalid"), "candidate_output_invalid")["runs"]),
                    }
                    manifest_bytes = canonical_json(capture_manifest)
                    _, terms = decode_privacy_policy(read_regular(snapshot.privacy_policy, mode=None, cap=MAX_POLICY_BYTES, error="privacy_policy_invalid"))
                    if set(capture_manifest) != CAPTURE_MANIFEST_KEYS or privacy_hit(manifest_bytes, terms) or not is_generic_result(capture_manifest):
                        raise CandidateAdapterError("result_invalid")
                    for retained in stage.rglob("*"):
                        if retained.is_file() and retained.suffix in {".json", ".jsonl"} and privacy_hit(read_regular(retained, mode=0o600, cap=MAX_METADATA_BYTES, error="result_invalid"), terms):
                            raise CandidateAdapterError("privacy_violation")
                    manifest_path = stage / "candidate-capture-manifest.json"; manifest_path.write_bytes(manifest_bytes); manifest_path.chmod(0o600)
                    verify_execution_snapshot(snapshot)
                    os.replace(stage, config.output_root)
                finally:
                    os.close(consumed.descriptor)
            finally:
                release_execution_snapshot(snapshot)
        result = {"$schema": RESULT_SCHEMA, "status": "captured_unvalidated", "effect_claim": False, "candidate_execution_authorized": True, "max_capture_count": 1, "candidate_delta_sha256": binding.candidate_delta_sha256, "candidate_output_manifest_sha256": sha256_file(config.output_root / "candidate-capture-manifest.json")}
        if not is_generic_result(result):
            raise CandidateAdapterError("result_invalid")
        return result
    except BaseException:
        shutil.rmtree(stage, ignore_errors=True)
        shutil.rmtree(config.output_root, ignore_errors=True)
        raise


def is_generic_result(value: Any) -> bool:
    encoded = canonical_json(value).lower()
    forbidden = (b"/home/", b"/users/", b"/root/", b"/private/", b"\\users\\", b"raw child output", b"private key", b"authorization: bearer", b'"stdout"', b'"stderr"')
    return not any(item in encoded for item in forbidden)


def _parse(argv: list[str] | None = None) -> argparse.Namespace:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("command", choices=("validate-inputs", "prepare-candidate", "capture-candidate"))
    parser.add_argument("--parent-baseline-root", type=Path, required=True)
    parser.add_argument("--parent-validator-store", type=Path, required=True)
    parser.add_argument("--parent-validator-ref", required=True)
    parser.add_argument("--privacy-policy", type=Path, required=True)
    parser.add_argument("--capability-root", type=Path, required=True)
    parser.add_argument("--output-root", type=Path, required=True)
    parser.add_argument("--candidate-search-go", type=Path)
    parser.add_argument("--candidate-search-test-go", type=Path)
    parser.add_argument("--activation", type=Path)
    parser.add_argument("--authorization-record", type=Path)
    parser.add_argument("--go-executable", type=Path)
    parser.add_argument("--git-executable", type=Path)
    return parser.parse_args(argv)


def main(argv: list[str] | None = None) -> int:
    args = _parse(argv)
    config = CandidateConfig(**{key: value for key, value in vars(args).items() if key != "command"})
    try:
        if args.command == "validate-inputs":
            result = validate_preparation_inputs(config)
        elif args.command == "prepare-candidate":
            result = prepare_candidate(config)
        else:
            result = capture_candidate(config)
    except CandidateAdapterError as exc:
        result = {"$schema": RESULT_SCHEMA, "status": "rejected", "failure_class": str(exc), "effect_claim": False, "candidate_execution_authorized": False}
        print(canonical_json(result).decode())
        return 1
    print(canonical_json(result).decode())
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
