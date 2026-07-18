"""Temporary dependency links for isolated C3 eval commands."""

from __future__ import annotations

import argparse
import os
from pathlib import Path
from typing import Callable, TypeVar


T = TypeVar("T")


def dependency_link(value: str) -> tuple[Path, Path]:
    relative, separator, raw_source = value.partition("=")
    destination = Path(relative)
    source = Path(raw_source).expanduser()
    if (
        not separator
        or not relative
        or not raw_source
        or destination.is_absolute()
        or ".." in destination.parts
        or destination == Path(".")
    ):
        raise argparse.ArgumentTypeError("expected safe relative-path=existing-absolute-path")
    source = source.resolve()
    if not source.is_absolute() or not source.exists():
        raise argparse.ArgumentTypeError("dependency source must exist")
    return destination, source


def with_dependency_links(project: Path, links: list[tuple[Path, Path]], callback: Callable[[], T]) -> T:
    created: list[Path] = []
    try:
        for relative, source in links:
            destination = project / relative
            if not destination.parent.is_dir():
                raise ValueError("dependency link parent must already exist")
            if os.path.lexists(destination):
                raise ValueError("dependency link destination already exists")
            destination.symlink_to(source, target_is_directory=source.is_dir())
            created.append(destination)
        return callback()
    finally:
        for destination in reversed(created):
            destination.unlink(missing_ok=True)
