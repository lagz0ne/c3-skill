---
id: adr-20260623-c3x-skill-packaging-split
c3-seal: 11f2863ae5eaabcd979492ca972dfeea97811f5887bc3da0936bde32144ba411
title: c3x-skill-packaging-split
type: adr
goal: Ship C3 skill artifacts in both no-binary installer-friendly and fat self-contained forms, with Claude and Codex plugin metadata present in each archive and `.gitattributes` included so bundled binaries stay binary-safe when unpacked into Git-backed environments.
status: done
date: "2026-06-23"
---

# Skill Packaging Split

## Goal

Ship C3 skill artifacts in both no-binary installer-friendly and fat self-contained forms, with Claude and Codex plugin metadata present in each archive and `.gitattributes` included so bundled binaries stay binary-safe when unpacked into Git-backed environments.

## Context

The existing release workflow already produces per-platform fat skill ZIPs for sandboxed or offline use, and the npm runtime manager now owns verified runtime downloads for thin installs. The missing surface is a platform-neutral skill/plugin artifact that can be installed by plugin or skills CLIs without carrying a binary. That artifact still needs a working `bin/c3x.sh`, so the wrapper must fall back to the pinned npm runtime manager when no bundled binary and no local Go source are available. The same release assembly must also include `.gitattributes` and both Claude and Codex plugin manifests so Git servers and both agent plugin systems see the artifact correctly.

## Decision

Keep the existing per-platform fat skill ZIPs and add a platform-neutral no-binary skill ZIP named `c3-skill-v{VERSION}.zip`. Assemble all skill archives through one shared script that copies `.gitattributes`, `.claude-plugin/`, `.codex-plugin/`, and `skills/`; removes any stale `skills/c3/bin/c3x-*` before packaging; and adds exactly one platform binary only for each fat ZIP. Add `.codex-plugin/plugin.json` as a first-class release manifest. Teach `skills/c3/bin/c3x.sh` to preserve its current bundled-binary and source-build paths, then delegate to `npm exec --yes --package @c3x/cli@${VERSION} -- c3x` when installed from a no-binary skill artifact.

## Affected Topology

| Entity | Type | Why affected | Evidence | Governance review |
| --- | --- | --- | --- | --- |
| c3-0 | system | The top-level distribution wording must include an agent skill surface that now has both Claude and Codex plugin metadata plus fat and no-binary skill artifacts. | c3-0#n3@v1:sha256:4295e84171aebab432093423315f2571c121a774d6338bf7330dd42644c6dfc2 "Build and distribute C3" | Parent system goal and skill container row need wording updates. |
| c3-2 | container | The skill distribution now covers Claude and Codex plugin metadata and can run through either a bundled binary or the pinned npm runtime manager. | c3-2#n540@v1:sha256:f0177f46a4bad8f33630a5c2228d6ca7e14117c9787ab7a9b59d45846ffb5866 "Teach an agent to operate C3 through shared skill instructions" | Parent Delta: update goal, c3-203 member contribution, responsibilities, and complexity framing. |
| c3-203 | component | The wrapper gains a no-binary artifact fallback through the pinned npm manager after bundled-binary and source-build attempts fail. | c3-203#n603@v1:sha256:54edea1eb796d0b265904816dad70fc6cf7ee1e15430a72c1d27662a9ad038ce "Detect the host platform, select a version-pinned full or Linux portable packaged binary," | Update goal, purpose, contract, and derived-material evidence. |
| ref-fat-thin-distribution | N.A - governing ref update | The distribution standard changes from two artifact shapes to three: fat skill, no-binary skill/plugin, and thin npm runtime manager. | ref-fat-thin-distribution#n748@v1:sha256:cf8e08bcc48d161ef4120914a8505499068ac512709adeae244258fd1618b031 "Distribute C3 as full-fat skill ZIPs (semantic binaries bundled), Linux portable fat skill" | Update the governing distribution reference before relying on it. |

## Compliance Refs

| Ref | Why required | Evidence | Action |
| --- | --- | --- | --- |
| ref-cross-compiled-binary | Fat skill artifacts and npm runtime downloads still consume the same standalone per-platform binary names and supported platform matrix. | ref-cross-compiled-binary#n730@v1:sha256:cea27fa9abdd975d6298f23e899ceb48f2945fde1e933915fdfada258a190136 "Ship c3x as named per-platform release binaries, with standard runtime-manager binaries," | Comply: do not change asset naming or supported platform gates. |
| ref-fat-thin-distribution | This decision changes the artifact split the reference governs. | ref-fat-thin-distribution#n748@v1:sha256:cf8e08bcc48d161ef4120914a8505499068ac512709adeae244258fd1618b031 "Distribute C3 as full-fat skill ZIPs (semantic binaries bundled), Linux portable fat skill" | Update-ref: describe fat skill ZIPs, platform-neutral no-binary skill/plugin ZIP, and thin npm runtime manager. |
| ref-frontmatter-docs | Frozen fact updates must stay in C3 frontmatter plus canvas-shaped markdown and land through a change-unit. | ref-frontmatter-docs#n757@v1:sha256:d4f7719668519e2f2a93de15969bc53c8f0105e7e073231a2f36d7c2626cb361 | Comply: stage fact patches and verify with c3x check --include-adr. |
| ref-eval-determinism | Inherited through the broader system, but this work changes packaging and wrapper dispatch only, not eval verdict computation. | ref-eval-determinism#n739@v1:sha256:d914f393b17de0202b7ae4cdde4df7d173c51fd820b2695487a29efb06f514d7 | N.A - no eval pipeline or selector change. |

## Compliance Rules

| Rule | Why required | Evidence | Action |
| --- | --- | --- | --- |
| rule-dispatcher-error-hint | The wrapper has a user-facing not-found path when no binary/source/npm fallback is available. | rule-dispatcher-error-hint#n766@v1:sha256:bd662000c1bc5b93d0b1cc4cf532cc1dc6e4766e5bda6b544f8aab14d21f7dc4 | Comply: keep the fallback failure hint actionable. |
| rule-output-via-helpers | Inherited from Go CLI structured output, but this change only affects shell packaging and wrapper delegation. | rule-output-via-helpers#n779@v1:sha256:b5ac8121ffc54be6c8f87ec133e69658fea023e7e73da3859fb85a33869afa29 | N.A - no Go command output path changes. |
| rule-wrap-error-cause | Inherited from Go command boundaries, but this change does not add Go error wrapping. | rule-wrap-error-cause#n791@v1:sha256:b9e4edb84b11060973de3fe6e5c0ab7b5605aa690e00e886335b054bdaab710f | N.A - shell fallback errors stay direct and include the recovery hint. |

## Work Breakdown

| Area | Detail | Evidence |
| --- | --- | --- |
| Release assembly | Add a shared release assembly script that emits c3-skill-v{VERSION}.zip plus existing per-platform fat skill ZIPs, all with .gitattributes, .claude-plugin/, .codex-plugin/, and skills/. | scripts/assemble_release_assets.sh; workflow calls |
| Plugin metadata | Add .codex-plugin/plugin.json and include it in skill archives; keep existing .claude-plugin/ metadata in the same artifacts. | .codex-plugin/plugin.json; packaging test zip assertions |
| Wrapper fallback | Let skills/c3/bin/c3x.sh delegate no-binary installs to pinned @c3x/cli@${VERSION} while retaining bundled-binary and source-build paths. | scripts/test_skill_release_packaging.py wrapper test |
| Docs | Explain no-binary plugin installs, fat ZIPs, and npm runtime-manager responsibilities. | README.md; package README; npm CLI spec/plan |

## Enforcement Surfaces

| Surface | Behavior | Evidence |
| --- | --- | --- |
| scripts/test_skill_release_packaging.py | Verifies fat and no-binary ZIP contents, .gitattributes, Claude/Codex manifests, no binary leakage in the no-binary archive, SHA256SUMS, and wrapper npm fallback. | python3 scripts/test_skill_release_packaging.py |
| release workflows | Run the packaging test before release assembly and call the shared assembly helper. | .github/workflows/release.yml; .github/workflows/distribute.yml |
| Codex plugin validator | Validates .codex-plugin/plugin.json against the local Codex plugin schema. | python3 /home/lagz0ne/.codex/skills/.system/plugin-creator/scripts/validate_plugin.py . |

## Alternatives Considered

| Alternative | Rejected because |
| --- | --- |
| Replace fat ZIPs with only no-binary plugin installs | It would break sandboxed and offline environments where npm/network access is intentionally unavailable. |
| Keep only fat ZIPs | It would force plugin/skills CLI installs to carry platform binaries even when the npm runtime manager can fetch verified assets on demand. |
| Let no-binary c3x.sh search PATH for a bare c3x | That would reintroduce global binary ambiguity and violate the repo's local-wrapper discipline; the fallback is pinned to @c3x/cli@${VERSION} instead. |

## Risks

| Risk | Mitigation | Verification |
| --- | --- | --- |
| No-binary skill installs execute an unintended runtime | Pin the npm package version from skills/c3/bin/VERSION and export C3X_VERSION before delegation. | Wrapper fallback test captures npm arguments and env. |
| Fat ZIP binary is corrupted or treated as text after unpacking into a Git repo | Include root .gitattributes in every skill ZIP with skills/c3/bin/c3x-* binary. | Packaging test asserts .gitattributes in both archive types. |
| Release workflows drift between main release and tag distribution | Use one shared assembly script from both workflows. | Packaging test plus workflow diff. |

## Verification

| Check | Result |
| --- | --- |
| python3 scripts/test_skill_release_packaging.py | Required before done. |
| python3 /home/lagz0ne/.codex/skills/.system/plugin-creator/scripts/validate_plugin.py . | Required before done. |
| claude plugin validate . --strict | Required before done. |
| cd packages/cli && npm test | Required before done. |
| C3X_MODE=agent bash skills/c3/bin/c3x.sh check --include-adr | Required before done. |
