#!/usr/bin/env python3
"""Guard the skill's documented C3 command invocation contract."""

from __future__ import annotations

import re
import unittest
from pathlib import Path


REPO_ROOT = Path(__file__).resolve().parents[1]
SKILL_ROOT = REPO_ROOT / "skills" / "c3"
DOCS = (SKILL_ROOT / "SKILL.md", *sorted((SKILL_ROOT / "references").glob("*.md")))

BARE_C3_COMMAND = re.compile(
    r"(?<![A-Za-z0-9_./-])c3\s+(?:"
    r"add|canvas|change|check|delete|eval|git|graph|init|list|lookup|migrate|"
    r"read|repair|schema|search|set|supersede|write|<cmd>|--help"
    r")(?![A-Za-z0-9_-])"
)


class SkillCLIContractTest(unittest.TestCase):
    def test_skill_docs_do_not_present_bare_c3_as_an_executable(self) -> None:
        violations: list[str] = []
        for path in DOCS:
            for line_number, line in enumerate(path.read_text(encoding="utf-8").splitlines(), 1):
                if BARE_C3_COMMAND.search(line):
                    violations.append(f"{path.relative_to(REPO_ROOT)}:{line_number}: {line}")

        self.assertEqual([], violations, "skill docs contain bare c3 command examples:\n" + "\n".join(violations))

    def test_skill_declares_the_explicit_local_wrapper(self) -> None:
        skill = (SKILL_ROOT / "SKILL.md").read_text(encoding="utf-8")
        self.assertIn('C3X_MODE=agent bash "<skill-dir>/bin/c3x.sh"', skill)
        self.assertNotIn("c3() {", skill)

    def test_sweep_requires_auditable_impact_closure(self) -> None:
        sweep = (SKILL_ROOT / "references" / "sweep.md").read_text(encoding="utf-8")
        for required in (
            "Reverse dependency route",
            "Code propagation route",
            "Contract and failure route",
            "affected / unaffected / unknown",
            "Evidence",
            "Isolation Boundaries",
            "Unknowns",
            "Code Changes Proposed",
            "C3 Fact Patches Required",
            "## Impact Classification",
            "| Route | Lane | Evidence checked | Status | Next check |",
        ):
            self.assertIn(required, sweep)
        self.assertNotIn("## Affected Entities", sweep)
        self.assertNotIn("File Changes Required (patches in .c3/changes", sweep)
        self.assertIn("N/A for a change that does not remove or retire a fact", sweep)
        self.assertIn("one row for every surfaced lane", sweep)
        self.assertIn("Unknown rows require evidence checked and a next check", sweep)


if __name__ == "__main__":
    unittest.main()
