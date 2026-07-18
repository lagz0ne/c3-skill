#!/usr/bin/env python3
"""Read-only, fail-closed validation of a protocol-v7 preparation freeze."""

from __future__ import annotations

import argparse
from contextlib import contextmanager
from dataclasses import dataclass
import hashlib
import json
import os
from pathlib import Path
import selectors
import shutil
import signal
import stat
import subprocess
import tempfile
import time
from typing import Any, Callable


RESULT_SCHEMA = "structural-retrieval-protocol-v7-preparation-verifier.v1"
HARNESS_SCHEMA = "structural-retrieval-protocol-v7-preparation-verifier-harness.v1"
PREPARATION_SCHEMA = "structural-retrieval-protocol-v7-preparation.v1"
INPUT_MANIFEST_SCHEMA = "structural-retrieval-protocol-v7-adapter-inputs.v1"

ACCEPTED_ADAPTER_SHA256 = "1c0ca65efd476211668fcd16750ead158c7f2b993efcf4845ba9a76403d98d65"
ACCEPTED_DRIVER_SHA256 = "d0d8205d20e3407048377a3e770066f0fb0013c41431f994ae6b0473532b6c45"
ACCEPTED_MAIN_SHA256 = "33b6452a0e34d825ef2b895609daa77b91eb129ebdc1e6be8f7f7755fc8b9c6e"
ACCEPTED_MAIN_TEST_SHA256 = "0830a0671aa9eb146902dc6fa56e885126b89a3312ca6c38fd8507fa10481dbf"
ACCEPTED_FIXTURE_SHA256 = "15f57120c6aa9ae07bf4fdacd6ad783afa5e70ed8ebebaff3a42dcf4249e677e"
ACCEPTED_BENCHMARK_SHA256 = "b960525cc42216e6598452946da5fb68735bbf989f311f170cedcfdbe92bf0d5"
ACCEPTED_SCORER_REGION_SHA256 = "c1669d43b13a36eb1454a01a5ea7e444fe67ce28fdae9628da083927f093cbcb"
ACCEPTED_GO_EXECUTABLE_SHA256 = "86b748c64de0175db601f56805251f3b08cd12bffb927ad5c68ef8497c50c7ba"
ACCEPTED_GIT_EXECUTABLE_SHA256 = "356db14e102d68a1a37d8a1ac577dfd678d45d46e92f468bef8b7154e7bfdc60"

MAX_METADATA_BYTES = 4 << 20
MAX_POLICY_BYTES = 128 << 10
MAX_CHECKINS_BYTES = 16 << 20
MAX_SOURCE_FILE_BYTES = 64 << 20
MAX_SOURCE_FILE_COUNT = 20_000
MAX_SOURCE_BYTES = 512 << 20
MAX_HARNESS_RESULT_BYTES = 128 << 10
PROCESS_OUTPUT_CAP = 4 << 20
PROCESS_TIMEOUT_SECONDS = 300

PREPARATION_FILES = {
    "controller-authority.v4.json": (0o600, MAX_METADATA_BYTES),
    "preparation-manifest.json": (0o600, MAX_METADATA_BYTES),
    "freeze/source.bundle": (0o600, 67_108_864),
    "freeze/controller-runtime": (0o600, 128 << 20),
}

PREPARATION_MANIFEST_KEYS = {
    "$schema",
    "adapter_sha256",
    "authority_sha256",
    "benchmark_sha256",
    "bundle_heads_sha256",
    "bundle_sha256",
    "commit",
    "driver_sha256",
    "fixture_sha256",
    "input_manifest_sha256",
    "main_sha256",
    "main_test_sha256",
    "privacy_policy_sha256",
    "privacy_term_count",
    "runtime_sha256",
    "scorer_region_sha256",
    "source_capsule_sha256",
    "status",
    "tree",
}

HARNESS_ACCEPTED_KEYS = {
    "$schema",
    "authority_sha256",
    "bundle_heads_sha256",
    "bundle_sha256",
    "privacy_detector_sha256",
    "privacy_policy_sha256",
    "privacy_term_count",
    "runtime_sha256",
    "scanned_artifact_count",
    "scanned_byte_count",
    "source_capsule_sha256",
    "source_object_byte_count",
    "source_object_count",
    "source_path_count",
    "status",
}

SAFE_FAILURE_CLASSES = {
    "authority_replay_failed",
    "bundle_invalid",
    "frozen_input_invalid",
    "harness_config_invalid",
    "metadata_binding_invalid",
    "metadata_invalid",
    "privacy_policy_invalid",
    "privacy_violation",
    "source_checkout_invalid",
    "threshold_invalid",
}


class VerifierError(ValueError):
    """A durable, generic verifier failure."""


class _VerifierInterrupted(BaseException):
    pass


@dataclass(frozen=True)
class FrozenInputs:
    adapter_sha256: str = ACCEPTED_ADAPTER_SHA256
    driver_sha256: str = ACCEPTED_DRIVER_SHA256
    main_sha256: str = ACCEPTED_MAIN_SHA256
    main_test_sha256: str = ACCEPTED_MAIN_TEST_SHA256
    fixture_sha256: str = ACCEPTED_FIXTURE_SHA256
    benchmark_sha256: str = ACCEPTED_BENCHMARK_SHA256
    scorer_region_sha256: str = ACCEPTED_SCORER_REGION_SHA256
    go_executable_sha256: str = ACCEPTED_GO_EXECUTABLE_SHA256
    git_executable_sha256: str = ACCEPTED_GIT_EXECUTABLE_SHA256


@dataclass(frozen=True)
class VerifierConfig:
    source_root: Path
    preparation_root: Path
    main_path: Path
    main_test_path: Path
    fixture_path: Path
    benchmark_path: Path
    threshold_checkins_path: Path
    threshold_checkin_seq: int
    privacy_policy_path: Path
    adapter_path: Path
    go_executable_path: Path
    git_executable_path: Path


@dataclass(frozen=True)
class ExecutableIdentity:
    role: str
    path: Path
    sha256: str


@dataclass(frozen=True)
class ToolchainIdentity:
    go: ExecutableIdentity
    git: ExecutableIdentity


@dataclass(frozen=True)
class SourceClosureEntry:
    path: str
    sha256: str


@dataclass(frozen=True)
class CompileClosure:
    entries: tuple[SourceClosureEntry, ...]


@dataclass(frozen=True)
class ProcessResult:
    returncode: int
    stdout: bytes
    stderr: bytes


Runner = Callable[..., ProcessResult]


def canonical_json(value: Any) -> bytes:
    return json.dumps(value, sort_keys=True, separators=(",", ":"), ensure_ascii=False).encode()


def _sha256_bytes(data: bytes) -> str:
    return hashlib.sha256(data).hexdigest()


def _valid_sha256(value: Any) -> bool:
    return (
        isinstance(value, str)
        and len(value) == 64
        and not set(value) - set("0123456789abcdef")
        and bool(value.strip("0"))
    )


def _valid_git_oid(value: Any) -> bool:
    return isinstance(value, str) and len(value) == 40 and not set(value) - set("0123456789abcdef")


def _reject_duplicates(pairs: list[tuple[str, Any]]) -> dict[str, Any]:
    result: dict[str, Any] = {}
    for key, value in pairs:
        if key in result:
            raise VerifierError("metadata_invalid")
        result[key] = value
    return result


def _decode_canonical_object(data: bytes, keys: set[str], error_class: str) -> dict[str, Any]:
    try:
        value = json.loads(data, object_pairs_hook=_reject_duplicates)
    except (UnicodeDecodeError, json.JSONDecodeError, VerifierError) as exc:
        raise VerifierError(error_class) from exc
    if not isinstance(value, dict) or set(value) != keys or canonical_json(value) != data:
        raise VerifierError(error_class)
    return value


def _regular_file(path: Path, *, mode: int | None, cap: int, error_class: str) -> os.stat_result:
    try:
        before = path.lstat()
    except OSError as exc:
        raise VerifierError(error_class) from exc
    if stat.S_ISLNK(before.st_mode) or not stat.S_ISREG(before.st_mode):
        raise VerifierError(error_class)
    if mode is not None and stat.S_IMODE(before.st_mode) != mode:
        raise VerifierError(error_class)
    if before.st_size < 0 or before.st_size > cap:
        raise VerifierError(error_class)
    try:
        if path.resolve(strict=True) != path.absolute():
            raise VerifierError(error_class)
    except OSError as exc:
        raise VerifierError(error_class) from exc
    return before


def _read_regular(path: Path, *, mode: int | None, cap: int, error_class: str) -> bytes:
    before = _regular_file(path, mode=mode, cap=cap, error_class=error_class)
    try:
        with path.open("rb") as stream:
            data = stream.read(cap + 1)
    except OSError as exc:
        raise VerifierError(error_class) from exc
    after = _regular_file(path, mode=mode, cap=cap, error_class=error_class)
    identity = lambda info: (info.st_dev, info.st_ino, info.st_mtime_ns, info.st_size, info.st_mode)
    if len(data) != before.st_size or len(data) > cap or identity(before) != identity(after):
        raise VerifierError(error_class)
    return data


def _directory(path: Path, mode: int | None, error_class: str) -> None:
    try:
        info = path.lstat()
    except OSError as exc:
        raise VerifierError(error_class) from exc
    if stat.S_ISLNK(info.st_mode) or not stat.S_ISDIR(info.st_mode):
        raise VerifierError(error_class)
    if mode is not None and stat.S_IMODE(info.st_mode) != mode:
        raise VerifierError(error_class)
    try:
        if path.resolve(strict=True) != path.absolute():
            raise VerifierError(error_class)
    except OSError as exc:
        raise VerifierError(error_class) from exc


def _inside(path: Path, root: Path) -> bool:
    try:
        path.absolute().relative_to(root.absolute())
        return True
    except ValueError:
        return False


def _paths_overlap(left: Path, right: Path) -> bool:
    return _inside(left, right) or _inside(right, left)


def _sha256_file(path: Path, *, cap: int, error_class: str) -> str:
    return _sha256_bytes(_read_regular(path, mode=None, cap=cap, error_class=error_class))


def _validate_preparation_layout(root: Path) -> dict[str, bytes]:
    _directory(root, 0o700, "preparation_layout_invalid")
    freeze = root / "freeze"
    _directory(freeze, 0o700, "preparation_layout_invalid")
    try:
        observed = set()
        for path in root.rglob("*"):
            info = path.lstat()
            if stat.S_ISLNK(info.st_mode):
                raise VerifierError("preparation_layout_invalid")
            relative = path.relative_to(root).as_posix()
            if stat.S_ISDIR(info.st_mode):
                if relative != "freeze":
                    raise VerifierError("preparation_layout_invalid")
            elif stat.S_ISREG(info.st_mode):
                observed.add(relative)
            else:
                raise VerifierError("preparation_layout_invalid")
    except OSError as exc:
        raise VerifierError("preparation_layout_invalid") from exc
    if observed != set(PREPARATION_FILES):
        raise VerifierError("preparation_layout_invalid")
    return {
        relative: _read_regular(
            root / relative,
            mode=mode,
            cap=cap,
            error_class="preparation_layout_invalid",
        )
        for relative, (mode, cap) in PREPARATION_FILES.items()
    }


def _validate_manifest(data: bytes, frozen: FrozenInputs, files: dict[str, bytes]) -> dict[str, Any]:
    value = _decode_canonical_object(data, PREPARATION_MANIFEST_KEYS, "metadata_invalid")
    if (
        value.get("$schema") != PREPARATION_SCHEMA
        or value.get("status") != "accepted"
        or value.get("adapter_sha256") != frozen.adapter_sha256
        or value.get("driver_sha256") != frozen.driver_sha256
        or value.get("main_sha256") != frozen.main_sha256
        or value.get("main_test_sha256") != frozen.main_test_sha256
        or value.get("fixture_sha256") != frozen.fixture_sha256
        or value.get("benchmark_sha256") != frozen.benchmark_sha256
        or value.get("scorer_region_sha256") != frozen.scorer_region_sha256
        or not isinstance(value.get("privacy_term_count"), int)
        or isinstance(value.get("privacy_term_count"), bool)
        or not 0 < value["privacy_term_count"] <= 512
        or not _valid_git_oid(value.get("commit"))
        or not _valid_git_oid(value.get("tree"))
    ):
        raise VerifierError("metadata_invalid")
    for key in (
        "authority_sha256",
        "bundle_heads_sha256",
        "bundle_sha256",
        "input_manifest_sha256",
        "privacy_policy_sha256",
        "runtime_sha256",
        "source_capsule_sha256",
    ):
        if not _valid_sha256(value.get(key)):
            raise VerifierError("metadata_invalid")
    bindings = {
        "authority_sha256": _sha256_bytes(files["controller-authority.v4.json"]),
        "bundle_sha256": _sha256_bytes(files["freeze/source.bundle"]),
        "runtime_sha256": _sha256_bytes(files["freeze/controller-runtime"]),
    }
    if any(value[key] != observed for key, observed in bindings.items()):
        raise VerifierError("metadata_binding_invalid")
    return value


def _validate_frozen_inputs(config: VerifierConfig, frozen: FrozenInputs) -> None:
    _directory(config.source_root, None, "frozen_input_invalid")
    exact = (
        (config.main_path, "cli/tools/structural-search-eval-v2/main.go", frozen.main_sha256),
        (config.main_test_path, "cli/tools/structural-search-eval-v2/main_test.go", frozen.main_test_sha256),
        (config.fixture_path, "research/eval/structural-retrieval/fixtures.dev.v2.jsonl", frozen.fixture_sha256),
        (config.benchmark_path, "research/eval/structural-retrieval/benchmark.v2.json", frozen.benchmark_sha256),
    )
    for path, relative, expected in exact:
        if path.absolute() != config.source_root.absolute() / relative:
            raise VerifierError("frozen_input_invalid")
        if _sha256_file(path, cap=MAX_SOURCE_FILE_BYTES, error_class="frozen_input_invalid") != expected:
            raise VerifierError("frozen_input_invalid")
    if _sha256_file(config.adapter_path, cap=4 << 20, error_class="frozen_input_invalid") != frozen.adapter_sha256:
        raise VerifierError("frozen_input_invalid")
    _regular_file(config.threshold_checkins_path, mode=None, cap=MAX_CHECKINS_BYTES, error_class="threshold_invalid")
    _regular_file(config.privacy_policy_path, mode=0o600, cap=MAX_POLICY_BYTES, error_class="privacy_policy_invalid")
    if config.threshold_checkin_seq <= 0:
        raise VerifierError("threshold_invalid")
    governed_pairs = (
        (config.preparation_root, config.source_root),
        (config.preparation_root, config.threshold_checkins_path),
        (config.preparation_root, config.privacy_policy_path),
        (config.preparation_root, config.adapter_path),
        (config.source_root, config.privacy_policy_path),
        (config.threshold_checkins_path, config.privacy_policy_path),
        (config.adapter_path, config.privacy_policy_path),
    )
    if any(_paths_overlap(left, right) for left, right in governed_pairs):
        raise VerifierError("governed_path_overlap")


def _validate_executable(path: Path, expected_sha256: str, role: str) -> ExecutableIdentity:
    if role not in {"go", "git"} or not path.is_absolute() or path != path.absolute():
        raise VerifierError("tool_identity_invalid")
    try:
        info = path.lstat()
        resolved = path.resolve(strict=True)
    except OSError as exc:
        raise VerifierError("tool_identity_invalid") from exc
    if (
        stat.S_ISLNK(info.st_mode)
        or not stat.S_ISREG(info.st_mode)
        or info.st_size <= 0
        or info.st_size > 256 << 20
        or info.st_mode & 0o111 == 0
        or resolved != path
    ):
        raise VerifierError("tool_identity_invalid")
    observed = _sha256_file(path, cap=256 << 20, error_class="tool_identity_invalid")
    if not _valid_sha256(expected_sha256) or observed != expected_sha256:
        raise VerifierError("tool_identity_invalid")
    return ExecutableIdentity(role=role, path=path, sha256=observed)


def _validate_tools(go_path: Path, git_path: Path, go_sha256: str, git_sha256: str) -> ToolchainIdentity:
    go = _validate_executable(go_path, go_sha256, "go")
    git = _validate_executable(git_path, git_sha256, "git")
    if go.path == git.path:
        raise VerifierError("tool_identity_invalid")
    return ToolchainIdentity(go=go, git=git)


def _decode_authority_compile_closure(authority_bytes: bytes) -> CompileClosure:
    try:
        authority = json.loads(authority_bytes, object_pairs_hook=_reject_duplicates)
    except (UnicodeDecodeError, json.JSONDecodeError, VerifierError) as exc:
        raise VerifierError("harness_source_invalid") from exc
    if not isinstance(authority, dict):
        raise VerifierError("harness_source_invalid")
    controller = authority.get("controller_source_capsule")
    runtime = authority.get("runtime_source_capsule")
    if not isinstance(controller, dict) or controller != runtime or set(controller) != {
        "$schema", "head_commit", "head_tree", "dirty_patch_sha256", "repository_build_input_count", "inputs"
    }:
        raise VerifierError("harness_source_invalid")
    inputs = controller.get("inputs")
    if (
        controller.get("$schema") != "structural-retrieval-source-capsule.v2"
        or not _valid_git_oid(controller.get("head_commit"))
        or not _valid_git_oid(controller.get("head_tree"))
        or not _valid_sha256(controller.get("dirty_patch_sha256"))
        or not isinstance(inputs, list)
        or controller.get("repository_build_input_count") != len(inputs)
        or not inputs
    ):
        raise VerifierError("harness_source_invalid")
    entries: list[SourceClosureEntry] = []
    previous = ""
    for item in inputs:
        if not isinstance(item, dict) or set(item) != {"origin", "path", "sha256"}:
            raise VerifierError("harness_source_invalid")
        relative = item.get("path")
        if (
            not isinstance(relative, str)
            or not relative.startswith("cli/")
            or relative <= previous
            or "\\" in relative
            or Path(relative).is_absolute()
            or Path(relative).as_posix() != relative
            or ".." in Path(relative).parts
            or item.get("origin") != "head"
            or not _valid_sha256(item.get("sha256"))
        ):
            raise VerifierError("harness_source_invalid")
        entries.append(SourceClosureEntry(path=relative, sha256=item["sha256"]))
        previous = relative
    main = "cli/tools/structural-search-eval-v2/main.go"
    if sum(entry.path == main for entry in entries) != 1:
        raise VerifierError("harness_source_invalid")
    return CompileClosure(entries=tuple(entries))


def _source_closure_snapshot(
    source_root: Path,
    closure: CompileClosure,
    protocol_test_path: Path,
    protocol_test_sha256: str,
) -> dict[str, str]:
    expected_test = source_root / "cli/tools/structural-search-eval-v2/main_test.go"
    if protocol_test_path.absolute() != expected_test.absolute():
        raise VerifierError("harness_source_invalid")
    snapshot: dict[str, str] = {}
    for entry in closure.entries:
        source = source_root / entry.path
        observed = _sha256_file(source, cap=MAX_SOURCE_FILE_BYTES, error_class="harness_source_invalid")
        if observed != entry.sha256:
            raise VerifierError("harness_source_invalid")
        snapshot[entry.path] = observed
    if "cli/tools/structural-search-eval-v2/main_test.go" in snapshot:
        raise VerifierError("harness_source_invalid")
    observed_test = _sha256_file(
        protocol_test_path, cap=MAX_SOURCE_FILE_BYTES, error_class="harness_source_invalid"
    )
    if observed_test != protocol_test_sha256:
        raise VerifierError("harness_source_invalid")
    snapshot["cli/tools/structural-search-eval-v2/main_test.go"] = observed_test
    return snapshot


def _copy_compiler_closure(
    source_root: Path,
    target_root: Path,
    closure: CompileClosure,
    protocol_test_path: Path,
    protocol_test_sha256: str,
) -> dict[str, str]:
    snapshot = _source_closure_snapshot(source_root, closure, protocol_test_path, protocol_test_sha256)
    for relative, expected in snapshot.items():
        source = source_root / relative
        data = _read_regular(source, mode=None, cap=MAX_SOURCE_FILE_BYTES, error_class="harness_source_invalid")
        if _sha256_bytes(data) != expected:
            raise VerifierError("harness_source_invalid")
        destination = target_root / relative
        destination.parent.mkdir(parents=True, exist_ok=True)
        destination.write_bytes(data)
        destination.chmod(0o644)
    _validate_compiler_tree(target_root, snapshot)
    return snapshot


def _validate_compiler_tree(target_root: Path, expected: dict[str, str]) -> dict[str, str]:
    observed: dict[str, str] = {}
    try:
        for path in sorted(target_root.rglob("*")):
            info = path.lstat()
            if stat.S_ISLNK(info.st_mode):
                raise VerifierError("harness_source_invalid")
            if stat.S_ISDIR(info.st_mode):
                continue
            if not stat.S_ISREG(info.st_mode):
                raise VerifierError("harness_source_invalid")
            relative = path.relative_to(target_root).as_posix()
            if relative not in expected:
                raise VerifierError("harness_source_invalid")
            observed[relative] = _sha256_file(
                path, cap=MAX_SOURCE_FILE_BYTES, error_class="harness_source_invalid"
            )
    except OSError as exc:
        raise VerifierError("harness_source_invalid") from exc
    if observed != expected:
        raise VerifierError("harness_source_invalid")
    return observed


def _materialize_minimal_tool_path(root: Path, tools: ToolchainIdentity) -> Path:
    root.mkdir(mode=0o700)
    for identity in (tools.go, tools.git):
        destination = root / identity.role
        with identity.path.open("rb") as source, destination.open("xb") as target:
            shutil.copyfileobj(source, target)
        destination.chmod(0o700)
        if _sha256_file(destination, cap=256 << 20, error_class="tool_identity_invalid") != identity.sha256:
            raise VerifierError("tool_identity_invalid")
    if {path.name for path in root.iterdir()} != {"go", "git"}:
        raise VerifierError("tool_identity_invalid")
    return root


def harness_command(go_executable: Path) -> list[str]:
    return [
        str(go_executable),
        "test",
        "-count=1",
        "-run",
        "^TestProtocolV7PreparationVerifier$",
        "./tools/structural-search-eval-v2",
    ]


def require_allowed_command(command: list[str]) -> None:
    if not command or not Path(command[0]).is_absolute() or command != harness_command(Path(command[0])):
        raise VerifierError("harness_command_invalid")


def _kill_process(process: subprocess.Popen[bytes]) -> None:
    try:
        os.killpg(process.pid, signal.SIGKILL)
    except (ProcessLookupError, PermissionError):
        try:
            process.kill()
        except ProcessLookupError:
            pass
    try:
        process.wait(timeout=5)
    except (subprocess.TimeoutExpired, ChildProcessError):
        pass


@contextmanager
def _scoped_termination_handler() -> Any:
    previous: dict[int, Any] = {}

    def interrupt(signum: int, _frame: Any) -> None:
        raise _VerifierInterrupted(signum)

    for signum in (signal.SIGTERM, signal.SIGHUP):
        try:
            previous[signum] = signal.getsignal(signum)
            signal.signal(signum, interrupt)
        except ValueError:
            for installed, handler in previous.items():
                signal.signal(installed, handler)
            previous.clear()
            break
    try:
        yield
    finally:
        for signum, handler in previous.items():
            signal.signal(signum, handler)


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
        raise VerifierError("harness_execution_failed") from exc
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
                raise VerifierError("harness_timeout")
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
                    raise VerifierError("harness_output_cap")
        returncode = process.wait(timeout=max(1, int(timeout - (time.monotonic() - started))))
        return ProcessResult(returncode, bytes(output["stdout"]), bytes(output["stderr"]))
    except BaseException:
        _kill_process(process)
        raise
    finally:
        selector.close()
        process.stdout.close()
        process.stderr.close()


def _go_environment(private_tmp: Path, tool_path: Path) -> dict[str, str]:
    home = Path.home()
    cache = Path(os.environ.get("XDG_CACHE_HOME", str(home / ".cache")))
    if not cache.is_absolute():
        cache = home / ".cache"
    return {
        "PATH": str(tool_path),
        "HOME": str(home),
        "LC_ALL": "C",
        "LANG": "C",
        "TZ": "UTC",
        "GOENV": "off",
        "GOWORK": "off",
        "GOFLAGS": "",
        "GOTOOLCHAIN": "local",
        "CGO_ENABLED": "0",
        "GOPROXY": "off",
        "GOSUMDB": "off",
        "GONOSUMDB": "",
        "GOPRIVATE": "",
        "GOMODCACHE": str(home / "go" / "pkg" / "mod"),
        "GOCACHE": str(cache / "go-build"),
        "TMPDIR": str(private_tmp),
    }


def _harness_config(
    config: VerifierConfig,
    frozen: FrozenInputs,
    manifest: dict[str, Any],
    compiled_main_path: Path,
    compiled_main_test_path: Path,
    result_path: Path,
    work_root: Path,
) -> dict[str, Any]:
    return {
        "$schema": "structural-retrieval-protocol-v7-preparation-verifier-config.v1",
        "adapter_path": str(config.adapter_path.absolute()),
        "adapter_sha256": frozen.adapter_sha256,
        "benchmark_path": str(config.benchmark_path.absolute()),
        "benchmark_sha256": frozen.benchmark_sha256,
        "bundle_path": str((config.preparation_root / "freeze/source.bundle").absolute()),
        "driver_sha256": frozen.driver_sha256,
        "fixture_path": str(config.fixture_path.absolute()),
        "fixture_sha256": frozen.fixture_sha256,
        "main_path": str(compiled_main_path.absolute()),
        "main_sha256": frozen.main_sha256,
        "main_test_path": str(compiled_main_test_path.absolute()),
        "main_test_sha256": frozen.main_test_sha256,
        "manifest_path": str((config.preparation_root / "preparation-manifest.json").absolute()),
        "policy_path": str(config.privacy_policy_path.absolute()),
        "preparation_authority_path": str((config.preparation_root / "controller-authority.v4.json").absolute()),
        "result_path": str(result_path),
        "runtime_path": str((config.preparation_root / "freeze/controller-runtime").absolute()),
        "scorer_region_sha256": frozen.scorer_region_sha256,
        "threshold_checkin_seq": config.threshold_checkin_seq,
        "threshold_checkins_path": str(config.threshold_checkins_path.absolute()),
        "work_root": str(work_root),
        "expected_manifest_sha256": _sha256_bytes(canonical_json(manifest)),
    }


def _validate_harness_result(data: bytes, manifest: dict[str, Any]) -> dict[str, Any]:
    try:
        preliminary = json.loads(data, object_pairs_hook=_reject_duplicates)
    except (UnicodeDecodeError, json.JSONDecodeError, VerifierError) as exc:
        raise VerifierError("harness_result_invalid") from exc
    if not isinstance(preliminary, dict) or canonical_json(preliminary) != data:
        raise VerifierError("harness_result_invalid")
    if preliminary.get("$schema") != HARNESS_SCHEMA:
        raise VerifierError("harness_result_invalid")
    if preliminary.get("status") == "rejected":
        if set(preliminary) != {"$schema", "failure_class", "status"} or preliminary.get("failure_class") not in SAFE_FAILURE_CLASSES:
            raise VerifierError("harness_result_invalid")
        raise VerifierError(preliminary["failure_class"])
    if preliminary.get("status") != "accepted" or set(preliminary) != HARNESS_ACCEPTED_KEYS:
        raise VerifierError("harness_result_invalid")
    for key in (
        "authority_sha256",
        "bundle_heads_sha256",
        "bundle_sha256",
        "privacy_detector_sha256",
        "privacy_policy_sha256",
        "runtime_sha256",
        "source_capsule_sha256",
    ):
        if not _valid_sha256(preliminary.get(key)):
            raise VerifierError("harness_result_invalid")
    bindings = (
        "authority_sha256",
        "bundle_heads_sha256",
        "bundle_sha256",
        "privacy_policy_sha256",
        "runtime_sha256",
        "source_capsule_sha256",
    )
    if any(preliminary[key] != manifest[key] for key in bindings):
        raise VerifierError("harness_result_invalid")
    for key in (
        "privacy_term_count",
        "scanned_artifact_count",
        "scanned_byte_count",
        "source_object_byte_count",
        "source_object_count",
        "source_path_count",
    ):
        value = preliminary.get(key)
        if not isinstance(value, int) or isinstance(value, bool) or value < 0:
            raise VerifierError("harness_result_invalid")
    if preliminary["privacy_term_count"] != manifest["privacy_term_count"]:
        raise VerifierError("harness_result_invalid")
    return preliminary


def _accepted_result(harness: dict[str, Any]) -> dict[str, Any]:
    return {
        "$schema": RESULT_SCHEMA,
        "authority_sha256": harness["authority_sha256"],
        "bundle_heads_sha256": harness["bundle_heads_sha256"],
        "bundle_sha256": harness["bundle_sha256"],
        "privacy_detector_sha256": harness["privacy_detector_sha256"],
        "privacy_policy_sha256": harness["privacy_policy_sha256"],
        "privacy_term_count": harness["privacy_term_count"],
        "runtime_sha256": harness["runtime_sha256"],
        "scanned_artifact_count": harness["scanned_artifact_count"],
        "scanned_byte_count": harness["scanned_byte_count"],
        "source_capsule_sha256": harness["source_capsule_sha256"],
        "source_object_byte_count": harness["source_object_byte_count"],
        "source_object_count": harness["source_object_count"],
        "source_path_count": harness["source_path_count"],
        "status": "accepted",
    }


def _rejected_result(error_class: str) -> dict[str, Any]:
    safe = error_class if error_class in {
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
        "preparation_layout_invalid",
        "privacy_policy_invalid",
        "privacy_violation",
        "threshold_invalid",
        "tool_identity_invalid",
        "governed_input_changed",
        *SAFE_FAILURE_CLASSES,
    } else "verification_internal_error"
    return {"$schema": RESULT_SCHEMA, "failure_class": safe, "status": "rejected"}


def _governed_snapshot(
    config: VerifierConfig,
    frozen: FrozenInputs,
    files: dict[str, bytes],
) -> dict[str, str]:
    return {
        "adapter": _sha256_file(config.adapter_path, cap=4 << 20, error_class="frozen_input_invalid"),
        "benchmark": _sha256_file(config.benchmark_path, cap=MAX_SOURCE_FILE_BYTES, error_class="frozen_input_invalid"),
        "fixture": _sha256_file(config.fixture_path, cap=MAX_SOURCE_FILE_BYTES, error_class="frozen_input_invalid"),
        "git": _sha256_file(config.git_executable_path, cap=256 << 20, error_class="tool_identity_invalid"),
        "go": _sha256_file(config.go_executable_path, cap=256 << 20, error_class="tool_identity_invalid"),
        "main": _sha256_file(config.main_path, cap=MAX_SOURCE_FILE_BYTES, error_class="frozen_input_invalid"),
        "main_test": _sha256_file(config.main_test_path, cap=MAX_SOURCE_FILE_BYTES, error_class="frozen_input_invalid"),
        "policy": _sha256_file(config.privacy_policy_path, cap=MAX_POLICY_BYTES, error_class="privacy_policy_invalid"),
        "threshold": _sha256_file(config.threshold_checkins_path, cap=MAX_CHECKINS_BYTES, error_class="threshold_invalid"),
        **{f"preparation:{relative}": _sha256_bytes(data) for relative, data in files.items()},
    }


def verify_preparation(
    config: VerifierConfig,
    *,
    frozen: FrozenInputs = FrozenInputs(),
    runner: Runner = run_bounded_process,
) -> dict[str, Any]:
    try:
        files = _validate_preparation_layout(config.preparation_root)
        _validate_frozen_inputs(config, frozen)
        tools = _validate_tools(
            config.go_executable_path,
            config.git_executable_path,
            frozen.go_executable_sha256,
            frozen.git_executable_sha256,
        )
        manifest = _validate_manifest(files["preparation-manifest.json"], frozen, files)
        closure = _decode_authority_compile_closure(files["controller-authority.v4.json"])
        before_source = _source_closure_snapshot(
            config.source_root, closure, config.main_test_path, frozen.main_test_sha256
        )
        before_governed = _governed_snapshot(config, frozen, files)
        with _scoped_termination_handler():
            with tempfile.TemporaryDirectory(prefix="c3-v7-preparation-verifier-") as temporary:
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
                    _sha256_file(compiled_main, cap=MAX_SOURCE_FILE_BYTES, error_class="frozen_input_invalid")
                    != frozen.main_sha256
                    or _sha256_file(
                        compiled_main_test,
                        cap=MAX_SOURCE_FILE_BYTES,
                        error_class="frozen_input_invalid",
                    )
                    != frozen.main_test_sha256
                ):
                    raise VerifierError("frozen_input_invalid")
                harness_path = compiled_main.with_name("protocol_v7_preparation_verifier_test.go")
                harness_path.write_text(GO_HARNESS_SOURCE, encoding="utf-8")
                harness_path.chmod(0o600)
                compiler_snapshot[
                    "cli/tools/structural-search-eval-v2/protocol_v7_preparation_verifier_test.go"
                ] = _sha256_bytes(GO_HARNESS_SOURCE.encode())
                _validate_compiler_tree(driver_root, compiler_snapshot)
                result_path = temporary_root / "harness-result.json"
                work_root = temporary_root / "work"
                work_root.mkdir(mode=0o700)
                harness_config = _harness_config(
                    config,
                    frozen,
                    manifest,
                    compiled_main,
                    compiled_main_test,
                    result_path,
                    work_root,
                )
                config_path = temporary_root / "harness-config.json"
                config_path.write_bytes(canonical_json(harness_config))
                config_path.chmod(0o600)
                environment = _go_environment(private_tmp, minimal_tools)
                environment["C3_PROTOCOL_V7_PREPARATION_VERIFIER_CONFIG"] = str(config_path)
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
                    raise VerifierError("harness_interrupted") from exc
                if completed.returncode != 0 or completed.stderr:
                    raise VerifierError("harness_rejected")
                _validate_compiler_tree(driver_root, compiler_snapshot)
                for identity in (tools.go, tools.git):
                    if _sha256_file(
                        minimal_tools / identity.role,
                        cap=256 << 20,
                        error_class="tool_identity_invalid",
                    ) != identity.sha256:
                        raise VerifierError("tool_identity_invalid")
                result_bytes = _read_regular(
                    result_path,
                    mode=0o600,
                    cap=MAX_HARNESS_RESULT_BYTES,
                    error_class="harness_result_invalid",
                )
                harness = _validate_harness_result(result_bytes, manifest)
        after_files = _validate_preparation_layout(config.preparation_root)
        after_source = _source_closure_snapshot(
            config.source_root, closure, config.main_test_path, frozen.main_test_sha256
        )
        after_governed = _governed_snapshot(config, frozen, after_files)
        _validate_tools(
            config.go_executable_path,
            config.git_executable_path,
            frozen.go_executable_sha256,
            frozen.git_executable_sha256,
        )
        if before_source != after_source or before_governed != after_governed:
            raise VerifierError("governed_input_changed")
        return _accepted_result(harness)
    except _VerifierInterrupted:
        return _rejected_result("harness_interrupted")
    except VerifierError as exc:
        return _rejected_result(str(exc))
    except (OSError, subprocess.SubprocessError):
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
)

const protocolV7PreparationVerifierHarnessSchema = "structural-retrieval-protocol-v7-preparation-verifier-harness.v1"

type protocolV7PreparationVerifierConfig struct {
    Schema string `json:"$schema"`
    AdapterPath string `json:"adapter_path"`
    AdapterSHA256 string `json:"adapter_sha256"`
    BenchmarkPath string `json:"benchmark_path"`
    BenchmarkSHA256 string `json:"benchmark_sha256"`
    BundlePath string `json:"bundle_path"`
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
    PreparationAuthorityPath string `json:"preparation_authority_path"`
    ResultPath string `json:"result_path"`
    RuntimePath string `json:"runtime_path"`
    ScorerRegionSHA256 string `json:"scorer_region_sha256"`
    ThresholdCheckinSeq int `json:"threshold_checkin_seq"`
    ThresholdCheckinsPath string `json:"threshold_checkins_path"`
    WorkRoot string `json:"work_root"`
}

type protocolV7PreparationManifest struct {
    Schema string `json:"$schema"`
    AdapterSHA256 string `json:"adapter_sha256"`
    AuthoritySHA256 string `json:"authority_sha256"`
    BenchmarkSHA256 string `json:"benchmark_sha256"`
    BundleHeadsSHA256 string `json:"bundle_heads_sha256"`
    BundleSHA256 string `json:"bundle_sha256"`
    Commit string `json:"commit"`
    DriverSHA256 string `json:"driver_sha256"`
    FixtureSHA256 string `json:"fixture_sha256"`
    InputManifestSHA256 string `json:"input_manifest_sha256"`
    MainSHA256 string `json:"main_sha256"`
    MainTestSHA256 string `json:"main_test_sha256"`
    PrivacyPolicySHA256 string `json:"privacy_policy_sha256"`
    PrivacyTermCount int `json:"privacy_term_count"`
    RuntimeSHA256 string `json:"runtime_sha256"`
    ScorerRegionSHA256 string `json:"scorer_region_sha256"`
    SourceCapsuleSHA256 string `json:"source_capsule_sha256"`
    Status string `json:"status"`
    Tree string `json:"tree"`
}

func protocolV7PreparationVerifierResult(status, failure string, values map[string]any) map[string]any {
    out := map[string]any{"$schema": protocolV7PreparationVerifierHarnessSchema, "status": status}
    if failure != "" { out["failure_class"] = failure }
    for key, value := range values { out[key] = value }
    return out
}

func protocolV7PreparationVerifierWriteResult(path string, result map[string]any) error {
    data, err := json.Marshal(result)
    if err != nil { return err }
    return os.WriteFile(path, data, 0o600)
}

func TestProtocolV7PreparationVerifier(t *testing.T) {
    configPath := os.Getenv("C3_PROTOCOL_V7_PREPARATION_VERIFIER_CONFIG")
    configBytes, err := readBoundedStandaloneRegularFile(configPath, 128<<10)
    if err != nil { t.Fatal("harness_config_invalid") }
    var cfg protocolV7PreparationVerifierConfig
    if decodeStrictBytes(configBytes, &cfg) != nil || cfg.Schema != "structural-retrieval-protocol-v7-preparation-verifier-config.v1" {
        t.Fatal("harness_config_invalid")
    }
    result := protocolV7PreparationVerifierRun(cfg)
    if err := protocolV7PreparationVerifierWriteResult(cfg.ResultPath, result); err != nil { t.Fatal("harness_result_invalid") }
}

func protocolV7PreparationVerifierRun(cfg protocolV7PreparationVerifierConfig) map[string]any {
    reject := func(class string) map[string]any { return protocolV7PreparationVerifierResult("rejected", class, nil) }
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
    policyBytes, err := readBoundedStandaloneRegularFile(cfg.PolicyPath, privacyPolicyBytesMax)
    if err != nil { return reject("privacy_policy_invalid") }
    policy, err := decodePrivacyPolicy(bytes.NewReader(policyBytes))
    if err != nil { return reject("privacy_policy_invalid") }
    scanner, err := newPrivacyScanner(policy)
    if err != nil || protocolV7PreparationVerifierPrivacySelfTest(policy) != nil { return reject("privacy_policy_invalid") }

    manifestBytes, err := readBoundedStandaloneRegularFile(cfg.ManifestPath, 4<<20)
    if err != nil || shaString(string(manifestBytes)) != cfg.ExpectedManifestSHA256 { return reject("metadata_invalid") }
    var manifest protocolV7PreparationManifest
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
    if manifest.AdapterSHA256 != cfg.AdapterSHA256 || manifest.DriverSHA256 != cfg.DriverSHA256 ||
        manifest.MainSHA256 != cfg.MainSHA256 || manifest.MainTestSHA256 != cfg.MainTestSHA256 ||
        manifest.FixtureSHA256 != cfg.FixtureSHA256 || manifest.BenchmarkSHA256 != cfg.BenchmarkSHA256 ||
        manifest.ScorerRegionSHA256 != cfg.ScorerRegionSHA256 || manifest.PrivacyPolicySHA256 != policy.SHA256 ||
        manifest.PrivacyTermCount != len(policy.DenyTerms) || manifest.InputManifestSHA256 != canonicalSHA256(inputManifest) {
        return reject("metadata_binding_invalid")
    }
    authorityBytes, err := readBoundedStandaloneRegularFile(cfg.PreparationAuthorityPath, 4<<20)
    if err != nil || shaString(string(authorityBytes)) != manifest.AuthoritySHA256 { return reject("metadata_binding_invalid") }
    var authority controllerAuthorityV4
    if decodeStrictBytes(authorityBytes, &authority) != nil { return reject("metadata_invalid") }
    canonicalAuthority, _ := json.Marshal(authority)
    if !bytes.Equal(canonicalAuthority, authorityBytes) || authority.Schema != controllerAuthorityV4Schema || authority.Mode != "baseline" {
        return reject("metadata_invalid")
    }
    if authority.Expected.Commit != manifest.Commit || authority.Expected.Tree != manifest.Tree ||
        authority.Expected.SourceCapsuleSHA256 != manifest.SourceCapsuleSHA256 || authority.Expected.RuntimeSHA256 != manifest.RuntimeSHA256 ||
        authority.Expected.BundleSHA256 != manifest.BundleSHA256 || authority.SourceBundleHeadsSHA256 != manifest.BundleHeadsSHA256 ||
        authority.PrivacyPolicySHA256 != manifest.PrivacyPolicySHA256 || authority.PrivacyTermCount != manifest.PrivacyTermCount {
        return reject("metadata_binding_invalid")
    }
    line, err := protocolV7PreparationVerifierCheckinLine(cfg.ThresholdCheckinsPath, cfg.ThresholdCheckinSeq)
    if err != nil || string(line) != authority.ContextThresholdAuthorityRecord { return reject("threshold_invalid") }

    sourceRoot := filepath.Join(cfg.WorkRoot, "B")
    if err := os.Mkdir(sourceRoot, 0o700); err != nil { return reject("source_checkout_invalid") }
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
    if err := scanner.Scan("preparation_manifest", "preparation-manifest.json", manifestBytes); err != nil { return reject("privacy_violation") }
    sourcePathCount := 0
    scannedBytes := 0
    for _, entry := range scanner.entries {
        scannedBytes += entry.Bytes
        if entry.Role == "source_tree_path" { sourcePathCount++ }
    }
    values := map[string]any{
        "authority_sha256": manifest.AuthoritySHA256, "bundle_heads_sha256": manifest.BundleHeadsSHA256,
        "bundle_sha256": manifest.BundleSHA256, "privacy_detector_sha256": scanner.detector.DefinitionSHA256,
        "privacy_policy_sha256": policy.SHA256, "privacy_term_count": len(policy.DenyTerms),
        "runtime_sha256": manifest.RuntimeSHA256, "scanned_artifact_count": len(scanner.entries),
        "scanned_byte_count": scannedBytes, "source_capsule_sha256": manifest.SourceCapsuleSHA256,
        "source_object_byte_count": scanner.sourceObjectBytes, "source_object_count": scanner.sourceObjectCount,
        "source_path_count": sourcePathCount,
    }
    return protocolV7PreparationVerifierResult("accepted", "", values)
}

func protocolV7PreparationVerifierPrivacySelfTest(policy privacyPolicy) error {
    detector, err := newGenericPrivacyDetector()
    if err != nil { return err }
    _, positives, negatives := privacyDetectorDefinition()
    for _, positive := range positives {
        if detector.Match([]byte(positive)) == "" { return errors.New("privacy detector positive missed") }
    }
    for _, negative := range negatives {
        if detector.Match([]byte(negative)) != "" { return errors.New("privacy detector negative matched") }
    }
    for _, term := range policy.DenyTerms {
        termScanner, err := newPrivacyScanner(policy)
        if err != nil || termScanner.Scan("self_test", "term", []byte(term)) == nil { return errors.New("privacy term positive missed") }
    }
    return nil
}

func protocolV7PreparationVerifierCheckinLine(path string, seq int) ([]byte, error) {
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
'''


def _parse_args(argv: list[str] | None = None) -> argparse.Namespace:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--source-root", type=Path, required=True)
    parser.add_argument("--preparation-root", type=Path, required=True)
    parser.add_argument("--main", dest="main_path", type=Path, required=True)
    parser.add_argument("--main-test", dest="main_test_path", type=Path, required=True)
    parser.add_argument("--fixtures", dest="fixture_path", type=Path, required=True)
    parser.add_argument("--benchmark", dest="benchmark_path", type=Path, required=True)
    parser.add_argument("--threshold-checkins", dest="threshold_checkins_path", type=Path, required=True)
    parser.add_argument("--threshold-checkin-seq", type=int, required=True)
    parser.add_argument("--privacy-policy", dest="privacy_policy_path", type=Path, required=True)
    parser.add_argument("--adapter", dest="adapter_path", type=Path, required=True)
    parser.add_argument("--go-executable", dest="go_executable_path", type=Path, required=True)
    parser.add_argument("--git-executable", dest="git_executable_path", type=Path, required=True)
    return parser.parse_args(argv)


def main(argv: list[str] | None = None) -> int:
    args = _parse_args(argv)
    result = verify_preparation(
        VerifierConfig(
            source_root=args.source_root,
            preparation_root=args.preparation_root,
            main_path=args.main_path,
            main_test_path=args.main_test_path,
            fixture_path=args.fixture_path,
            benchmark_path=args.benchmark_path,
            threshold_checkins_path=args.threshold_checkins_path,
            threshold_checkin_seq=args.threshold_checkin_seq,
            privacy_policy_path=args.privacy_policy_path,
            adapter_path=args.adapter_path,
            go_executable_path=args.go_executable_path,
            git_executable_path=args.git_executable_path,
        )
    )
    print(canonical_json(result).decode())
    return 0 if result["status"] == "accepted" else 2


if __name__ == "__main__":
    raise SystemExit(main())
