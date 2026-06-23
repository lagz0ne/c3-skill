import os
import platform
import shutil
import subprocess
import tempfile
import unittest
import zipfile
from pathlib import Path


REPO_ROOT = Path(__file__).resolve().parents[1]


class SkillReleasePackagingTest(unittest.TestCase):
    def test_release_assembly_emits_fat_and_no_binary_skill_archives(self):
        with tempfile.TemporaryDirectory() as tmp:
            tmp_path = Path(tmp)
            artifacts = tmp_path / "artifacts"
            release = tmp_path / "release"

            thin_dir = artifacts / "linux-amd64" / "thin"
            fat_dir = artifacts / "linux-amd64" / "fat"
            portable_dir = artifacts / "linux-amd64" / "portable"
            semantic_dir = artifacts / "semantic-assets"
            thin_dir.mkdir(parents=True)
            fat_dir.mkdir(parents=True)
            portable_dir.mkdir(parents=True)
            semantic_dir.mkdir(parents=True)

            (thin_dir / "c3x-1.2.3-linux-amd64").write_text("thin binary\n")
            (thin_dir / "c3x-1.2.3-linux-amd64.sha256").write_text("thin checksum\n")
            (fat_dir / "c3x-1.2.3-linux-amd64-fat").write_text("fat binary\n")
            (portable_dir / "c3x-1.2.3-linux-amd64-portable").write_text("portable binary\n")
            (semantic_dir / "model.onnx").write_text("model\n")

            subprocess.run(
                [
                    "bash",
                    str(REPO_ROOT / "scripts" / "assemble_release_assets.sh"),
                    "--version",
                    "1.2.3",
                    "--artifacts-dir",
                    str(artifacts),
                    "--out-dir",
                    str(release),
                ],
                cwd=REPO_ROOT,
                check=True,
            )

            fat_zip = release / "c3-skill-linux-amd64-v1.2.3.zip"
            portable_zip = release / "c3-skill-linux-amd64-portable-v1.2.3.zip"
            no_binary_zip = release / "c3-skill-v1.2.3.zip"
            self.assertTrue(fat_zip.exists(), sorted(p.name for p in release.iterdir()))
            self.assertTrue(portable_zip.exists(), sorted(p.name for p in release.iterdir()))
            self.assertTrue(no_binary_zip.exists(), sorted(p.name for p in release.iterdir()))

            with zipfile.ZipFile(fat_zip) as archive:
                names = set(archive.namelist())
                self.assertIn(".gitattributes", names)
                self.assertIn(".claude-plugin/plugin.json", names)
                self.assertNotIn(".codex-plugin/plugin.json", names)
                self.assertIn("skills/c3/bin/c3x.sh", names)
                self.assertIn("skills/c3/bin/VERSION", names)
                self.assertIn("skills/c3/bin/c3x-1.2.3-linux-amd64", names)
                self.assertEqual(
                    zipfile.ZIP_DEFLATED,
                    archive.getinfo("skills/c3/bin/c3x.sh").compress_type,
                )

            with zipfile.ZipFile(portable_zip) as archive:
                names = set(archive.namelist())
                self.assertIn(".gitattributes", names)
                self.assertIn(".claude-plugin/plugin.json", names)
                self.assertNotIn(".codex-plugin/plugin.json", names)
                self.assertIn("skills/c3/bin/c3x.sh", names)
                self.assertIn("skills/c3/bin/VERSION", names)
                self.assertIn("skills/c3/bin/c3x-1.2.3-linux-amd64-portable", names)
                self.assertNotIn("skills/c3/bin/c3x-1.2.3-linux-amd64", names)

            with zipfile.ZipFile(no_binary_zip) as archive:
                names = set(archive.namelist())
                self.assertIn(".gitattributes", names)
                self.assertIn(".claude-plugin/plugin.json", names)
                self.assertNotIn(".codex-plugin/plugin.json", names)
                self.assertIn("skills/c3/bin/c3x.sh", names)
                binaries = [
                    name
                    for name in names
                    if name.startswith("skills/c3/bin/c3x-")
                    and not name.endswith("/")
                    and Path(name).name != "c3x.sh"
                ]
                self.assertEqual([], sorted(binaries))

            checksums = (release / "SHA256SUMS").read_text()
            self.assertIn("c3-skill-linux-amd64-v1.2.3.zip", checksums)
            self.assertIn("c3-skill-linux-amd64-portable-v1.2.3.zip", checksums)
            self.assertIn("c3-skill-v1.2.3.zip", checksums)

    def test_release_assembly_fails_when_linux_fat_lacks_matching_portable(self):
        with tempfile.TemporaryDirectory() as tmp:
            tmp_path = Path(tmp)
            artifacts = tmp_path / "artifacts"
            release = tmp_path / "release"

            (artifacts / "linux-amd64" / "fat").mkdir(parents=True)
            (artifacts / "linux-arm64" / "fat").mkdir(parents=True)
            (artifacts / "linux-amd64" / "portable").mkdir(parents=True)
            (artifacts / "linux-amd64" / "fat" / "c3x-1.2.3-linux-amd64-fat").write_text("fat\n")
            (artifacts / "linux-arm64" / "fat" / "c3x-1.2.3-linux-arm64-fat").write_text("fat\n")
            (artifacts / "linux-amd64" / "portable" / "c3x-1.2.3-linux-amd64-portable").write_text("portable\n")

            result = subprocess.run(
                [
                    "bash",
                    str(REPO_ROOT / "scripts" / "assemble_release_assets.sh"),
                    "--version",
                    "1.2.3",
                    "--artifacts-dir",
                    str(artifacts),
                    "--out-dir",
                    str(release),
                ],
                cwd=REPO_ROOT,
                text=True,
                stdout=subprocess.PIPE,
                stderr=subprocess.PIPE,
            )

            self.assertNotEqual(0, result.returncode)
            self.assertIn("Missing portable skill binary for linux-arm64-portable", result.stderr)

    def test_wrapper_prefers_linux_portable_binary_before_npm_fallback(self):
        os_name, arch = self._linux_target()
        if os_name != "linux":
            self.skipTest("portable wrapper fallback is linux-only")

        with tempfile.TemporaryDirectory() as tmp:
            tmp_path = Path(tmp)
            skill_bin = tmp_path / "skills" / "c3" / "bin"
            skill_bin.mkdir(parents=True)
            shutil.copy2(REPO_ROOT / "skills" / "c3" / "bin" / "c3x.sh", skill_bin / "c3x.sh")
            (skill_bin / "VERSION").write_text("1.2.3\n")

            portable = skill_bin / f"c3x-1.2.3-linux-{arch}-portable"
            portable.write_text(
                "#!/usr/bin/env bash\n"
                "printf 'portable:%s:%s\\n' \"$C3X_VERSION\" \"$*\"\n"
            )
            portable.chmod(0o755)

            fake_bin = tmp_path / "fake-bin"
            fake_bin.mkdir()
            npm_capture = tmp_path / "npm-capture"
            npm = fake_bin / "npm"
            npm.write_text(
                "#!/usr/bin/env bash\n"
                "printf 'npm-called\\n' > \"$NPM_CAPTURE\"\n"
                "exit 99\n"
            )
            npm.chmod(0o755)

            env = os.environ.copy()
            env["PATH"] = f"{fake_bin}{os.pathsep}{env['PATH']}"
            env["NPM_CAPTURE"] = str(npm_capture)

            result = subprocess.run(
                ["bash", str(skill_bin / "c3x.sh"), "--help"],
                cwd=tmp_path,
                env=env,
                text=True,
                stdout=subprocess.PIPE,
                stderr=subprocess.PIPE,
                check=True,
            )

            self.assertEqual("portable:1.2.3:--help\n", result.stdout)
            self.assertFalse(npm_capture.exists(), result.stderr)

    def test_no_binary_skill_wrapper_handles_passive_commands_without_npm(self):
        with tempfile.TemporaryDirectory() as tmp:
            tmp_path = Path(tmp)
            skill_bin = tmp_path / "skills" / "c3" / "bin"
            skill_bin.mkdir(parents=True)
            shutil.copy2(REPO_ROOT / "skills" / "c3" / "bin" / "c3x.sh", skill_bin / "c3x.sh")
            (skill_bin / "VERSION").write_text("1.2.3\n")

            fake_bin = tmp_path / "fake-bin"
            fake_bin.mkdir()
            capture = tmp_path / "npm-capture"
            npm = fake_bin / "npm"
            npm.write_text(
                "#!/usr/bin/env bash\n"
                "printf 'npm-called\\n' > \"$NPM_CAPTURE\"\n"
                "exit 99\n"
            )
            npm.chmod(0o755)

            env = os.environ.copy()
            env.pop("C3X_VERSION", None)
            env["PATH"] = f"{fake_bin}{os.pathsep}{env['PATH']}"
            env["NPM_CAPTURE"] = str(capture)

            for args, expected in ((["--help"], "Usage: c3x"), (["--version"], "c3x 1.2.3")):
                if capture.exists():
                    capture.unlink()
                result = subprocess.run(
                    ["bash", str(skill_bin / "c3x.sh"), *args],
                    cwd=tmp_path,
                    env=env,
                    text=True,
                    stdout=subprocess.PIPE,
                    stderr=subprocess.PIPE,
                    check=True,
                )

                self.assertIn(expected, result.stdout)
                self.assertFalse(capture.exists(), result.stderr)

    def test_no_binary_skill_wrapper_delegates_to_pinned_npm_manager(self):
        with tempfile.TemporaryDirectory() as tmp:
            tmp_path = Path(tmp)
            skill_bin = tmp_path / "skills" / "c3" / "bin"
            skill_bin.mkdir(parents=True)
            shutil.copy2(REPO_ROOT / "skills" / "c3" / "bin" / "c3x.sh", skill_bin / "c3x.sh")
            (skill_bin / "VERSION").write_text("1.2.3\n")

            fake_bin = tmp_path / "fake-bin"
            fake_bin.mkdir()
            capture = tmp_path / "npm-capture"
            npm = fake_bin / "npm"
            npm.write_text(
                "#!/usr/bin/env bash\n"
                "printf 'C3X_VERSION=%s\\n' \"$C3X_VERSION\" > \"$NPM_CAPTURE\"\n"
                "printf 'ARGS=%s\\n' \"$*\" >> \"$NPM_CAPTURE\"\n"
                "exit 17\n"
            )
            npm.chmod(0o755)

            env = os.environ.copy()
            env.pop("C3X_VERSION", None)
            env["PATH"] = f"{fake_bin}{os.pathsep}{env['PATH']}"
            env["NPM_CAPTURE"] = str(capture)

            result = subprocess.run(
                ["bash", str(skill_bin / "c3x.sh"), "list", "--flat"],
                cwd=tmp_path,
                env=env,
                text=True,
                stdout=subprocess.PIPE,
                stderr=subprocess.PIPE,
            )

            self.assertEqual(17, result.returncode)
            captured = capture.read_text()
            self.assertIn("C3X_VERSION=", captured)
            self.assertNotIn("C3X_VERSION=1.2.3", captured)
            self.assertIn(
                "ARGS=exec --yes --package @c3x/cli@1.2.3 -- c3x list --flat",
                captured,
            )

    def test_portable_build_has_no_elf_interpreter_and_runs_in_bwrap(self):
        os_name, arch = self._linux_target()
        if os_name != "linux":
            self.skipTest("portable binary check is linux-only")
        if shutil.which("go") is None:
            self.skipTest("go is required for portable build smoke test")
        if shutil.which("readelf") is None:
            self.skipTest("readelf is required for portable build smoke test")

        with tempfile.TemporaryDirectory() as tmp:
            out_dir = Path(tmp) / "dist"
            subprocess.run(
                [
                    "bash",
                    str(REPO_ROOT / "scripts" / "build.sh"),
                    "--version",
                    "1.2.3",
                    "--variant",
                    "portable",
                    "--os",
                    "linux",
                    "--arch",
                    arch,
                    "--out-dir",
                    str(out_dir),
                ],
                cwd=REPO_ROOT,
                check=True,
            )
            binary = out_dir / "portable" / f"c3x-1.2.3-linux-{arch}-portable"
            self.assertTrue(binary.exists(), binary)

            readelf = subprocess.run(
                ["readelf", "-l", str(binary)],
                text=True,
                stdout=subprocess.PIPE,
                stderr=subprocess.PIPE,
                check=True,
            )
            self.assertNotIn("Requesting program interpreter", readelf.stdout)

            if shutil.which("bwrap") is None:
                if self._running_in_ci():
                    self.fail("bwrap is required for CI portable isolation smoke test")
                self.skipTest("bwrap is required for portable isolation smoke test")

            result = subprocess.run(
                [
                    "bwrap",
                    "--unshare-net",
                    "--proc",
                    "/proc",
                    "--dev",
                    "/dev",
                    "--tmpfs",
                    "/tmp",
                    "--dir",
                    "/tmp/home",
                    "--dir",
                    "/tmp/cache",
                    "--ro-bind",
                    str(binary),
                    "/c3x",
                    "--setenv",
                    "HOME",
                    "/tmp/home",
                    "--setenv",
                    "XDG_CACHE_HOME",
                    "/tmp/cache",
                    "--setenv",
                    "C3_SEMANTIC_OFFLINE",
                    "1",
                    "/c3x",
                    "--help",
                ],
                text=True,
                stdout=subprocess.PIPE,
                stderr=subprocess.PIPE,
            )
            if result.returncode != 0 and (
                "Operation not permitted" in result.stderr
                or "No permissions to create new namespace" in result.stderr
            ):
                if self._running_in_ci():
                    self.fail(f"bwrap unavailable in CI: {result.stderr.strip()}")
                self.skipTest(f"bwrap unavailable in this environment: {result.stderr.strip()}")
            self.assertEqual(0, result.returncode, result.stderr)
            self.assertIn("Usage: c3x", result.stdout)

    def _running_in_ci(self):
        return os.environ.get("CI") == "true" or os.environ.get("GITHUB_ACTIONS") == "true"

    def _linux_target(self):
        system = platform.system().lower()
        machine = platform.machine().lower()
        if machine == "x86_64":
            arch = "amd64"
        elif machine in ("aarch64", "arm64"):
            arch = "arm64"
        else:
            arch = machine
        return system, arch


if __name__ == "__main__":
    unittest.main()
