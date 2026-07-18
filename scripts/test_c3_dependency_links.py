import argparse
import tempfile
import unittest
from pathlib import Path

from scripts.c3_dependency_links import dependency_link, with_dependency_links


class C3DependencyLinksTests(unittest.TestCase):
    def test_link_exists_only_during_callback(self) -> None:
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            project = root / "project"
            source = root / "shared-deps"
            project.mkdir()
            source.mkdir()
            link = dependency_link(f"node_modules={source}")

            def inspect() -> str:
                destination = project / "node_modules"
                self.assertTrue(destination.is_symlink())
                return destination.resolve().name

            self.assertEqual(with_dependency_links(project, [link], inspect), "shared-deps")
            self.assertFalse((project / "node_modules").exists())

    def test_existing_destination_is_rejected_without_removal(self) -> None:
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            project = root / "project"
            source = root / "shared-deps"
            project.mkdir()
            source.mkdir()
            existing = project / "node_modules"
            existing.mkdir()
            with self.assertRaisesRegex(ValueError, "already exists"):
                with_dependency_links(project, [(Path("node_modules"), source)], lambda: None)
            self.assertTrue(existing.is_dir())

    def test_path_traversal_is_rejected(self) -> None:
        with tempfile.TemporaryDirectory() as tmp:
            source = Path(tmp)
            with self.assertRaises(argparse.ArgumentTypeError):
                dependency_link(f"../node_modules={source}")


if __name__ == "__main__":
    unittest.main()
