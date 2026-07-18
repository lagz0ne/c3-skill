#!/usr/bin/env python3
"""Run an agent CLI behind provider-neutral, in-flight budget guards."""

from __future__ import annotations

import argparse
import json
import os
import signal
import subprocess
import sys
import time
from pathlib import Path
from typing import Any, Iterable


BUDGET_KILLED_EXIT = 86
TOOL_ITEM_TYPES = {
    "command_execution",
    "computer_tool_call",
    "function_call",
    "mcp_tool_call",
    "tool_call",
    "tool_use",
    "web_search",
}


def tool_events(value: dict[str, Any]) -> Iterable[str]:
    event_type = str(value.get("type") or "")
    item = value.get("item") if isinstance(value.get("item"), dict) else {}
    item_type = str(item.get("type") or "")
    if event_type in {"item.started", "tool_use"} and (
        item_type in TOOL_ITEM_TYPES or "tool" in item_type or "command" in item_type
    ):
        yield str(item.get("id") or value.get("id") or "")

    block = value.get("content_block") if isinstance(value.get("content_block"), dict) else {}
    if str(block.get("type") or "") == "tool_use":
        yield str(block.get("id") or "")

    message = value.get("message") if isinstance(value.get("message"), dict) else {}
    content = message.get("content") if isinstance(message.get("content"), list) else []
    for entry in content:
        if isinstance(entry, dict) and str(entry.get("type") or "") == "tool_use":
            yield str(entry.get("id") or "")


def tool_result(value: dict[str, Any]) -> tuple[str, int] | None:
    if str(value.get("type") or "") != "item.completed":
        return None
    item = value.get("item") if isinstance(value.get("item"), dict) else {}
    item_type = str(item.get("type") or "")
    if item_type not in TOOL_ITEM_TYPES and "tool" not in item_type and "command" not in item_type:
        return None
    output = item.get("aggregated_output")
    if not isinstance(output, str):
        output = item.get("output")
    if not isinstance(output, str):
        return None
    return str(item.get("id") or value.get("id") or ""), len(output.encode("utf-8"))


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--stdin", required=True)
    parser.add_argument("--stdout", required=True)
    parser.add_argument("--stderr", required=True)
    parser.add_argument("--status", required=True)
    parser.add_argument("--max-seconds", required=True, type=float)
    parser.add_argument("--max-tool-calls", required=True, type=int)
    parser.add_argument("--max-tool-result-bytes", required=True, type=int)
    parser.add_argument("--max-output-bytes", required=True, type=int)
    parser.add_argument("command", nargs=argparse.REMAINDER)
    args = parser.parse_args()
    if args.command and args.command[0] == "--":
        args.command = args.command[1:]
    if not args.command:
        parser.error("a command is required after --")
    if (
        args.max_seconds <= 0
        or args.max_tool_calls <= 0
        or args.max_tool_result_bytes <= 0
        or args.max_output_bytes <= 0
    ):
        parser.error("all limits must be positive")
    return args


def stop_process(process: subprocess.Popen[bytes]) -> None:
    if process.poll() is not None:
        return
    try:
        os.killpg(process.pid, signal.SIGTERM)
        process.wait(timeout=1)
    except (ProcessLookupError, subprocess.TimeoutExpired):
        if process.poll() is None:
            try:
                os.killpg(process.pid, signal.SIGKILL)
            except ProcessLookupError:
                pass
            process.wait()


def main() -> int:
    args = parse_args()
    stdin_path = Path(args.stdin)
    stdout_path = Path(args.stdout)
    stderr_path = Path(args.stderr)
    status_path = Path(args.status)
    for path in (stdout_path, stderr_path, status_path):
        path.parent.mkdir(parents=True, exist_ok=True)

    started = time.monotonic()
    seen_tools: set[str] = set()
    seen_tool_results: set[str] = set()
    anonymous_tool_count = 0
    tool_result_bytes = 0
    offset = 0
    pending = b""
    reason: str | None = None

    with (
        stdin_path.open("rb") as stdin_handle,
        stdout_path.open("wb") as stdout_handle,
        stderr_path.open("wb") as stderr_handle,
    ):
        process = subprocess.Popen(
            args.command,
            stdin=stdin_handle,
            stdout=stdout_handle,
            stderr=stderr_handle,
            start_new_session=True,
        )
        while process.poll() is None:
            stdout_handle.flush()
            stderr_handle.flush()
            with stdout_path.open("rb") as reader:
                reader.seek(offset)
                chunk = reader.read()
            if chunk:
                offset += len(chunk)
                pending += chunk
                lines = pending.split(b"\n")
                pending = lines.pop()
                for raw in lines:
                    try:
                        value = json.loads(raw)
                    except (UnicodeDecodeError, json.JSONDecodeError):
                        continue
                    if not isinstance(value, dict):
                        continue
                    for tool_id in tool_events(value):
                        if tool_id:
                            seen_tools.add(tool_id)
                        else:
                            anonymous_tool_count += 1
                    result = tool_result(value)
                    if result is not None:
                        result_id, result_bytes = result
                        if not result_id or result_id not in seen_tool_results:
                            tool_result_bytes += result_bytes
                            if result_id:
                                seen_tool_results.add(result_id)

            tool_count = len(seen_tools) + anonymous_tool_count
            output_bytes = stdout_path.stat().st_size + stderr_path.stat().st_size
            elapsed = time.monotonic() - started
            if tool_count > args.max_tool_calls:
                reason = "max_tool_calls"
            elif tool_result_bytes > args.max_tool_result_bytes:
                reason = "max_tool_result_bytes"
            elif output_bytes >= args.max_output_bytes:
                reason = "max_output_bytes"
            elif elapsed >= args.max_seconds:
                reason = "max_seconds"
            if reason:
                stop_process(process)
                break
            time.sleep(0.025)

        child_exit = process.wait()
        stdout_handle.flush()
        with stdout_path.open("rb") as reader:
            reader.seek(offset)
            pending += reader.read()
        for raw in pending.splitlines():
            try:
                value = json.loads(raw)
            except (UnicodeDecodeError, json.JSONDecodeError):
                continue
            if not isinstance(value, dict):
                continue
            for tool_id in tool_events(value):
                if tool_id:
                    seen_tools.add(tool_id)
                else:
                    anonymous_tool_count += 1
            result = tool_result(value)
            if result is not None:
                result_id, result_bytes = result
                if not result_id or result_id not in seen_tool_results:
                    tool_result_bytes += result_bytes
                    if result_id:
                        seen_tool_results.add(result_id)

    elapsed_ms = int((time.monotonic() - started) * 1000)
    output_bytes = stdout_path.stat().st_size + stderr_path.stat().st_size
    status = {
        "status": "budget_killed" if reason else "completed",
        "reason": reason,
        "tool_calls_observed": len(seen_tools) + anonymous_tool_count,
        "tool_result_bytes_observed": tool_result_bytes,
        "output_bytes_observed": output_bytes,
        "elapsed_ms": elapsed_ms,
        "child_exit_code": child_exit,
    }
    temporary = status_path.with_suffix(status_path.suffix + ".tmp")
    temporary.write_text(json.dumps(status, sort_keys=True) + "\n", encoding="utf-8")
    temporary.replace(status_path)
    return BUDGET_KILLED_EXIT if reason else child_exit


if __name__ == "__main__":
    sys.exit(main())
