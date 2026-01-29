# Self-Contained C3 Skills Build System

**Date:** 2026-01-29
**Status:** Proposed

## Problem

C3 skills currently use glob patterns like `**/references/skill-harness.md` that depend on the entire repository being cloned. When skills are installed directly (via marketplace or npm), these paths break because the parent directory structure isn't preserved.

## Goals

1. Skills and agents become self-contained packages
2. Build process transforms dev structure → distributable structure
3. CI triggers on merge to `main`, pushes built output to `dist` branch
4. Multiple output targets: Claude Code plugin, OpenCode plugin
5. `dist` becomes the default branch so direct installs work

## Non-Goals

- Inlining reference content directly into skill files (bloats context)
- Changing how skills work in development (maintain current workflow)

---

## Design

### Branch Structure

```
main (dev)          →  CI build  →  dist (default)
├── skills/                        ├── claude-code/
│   ├── c3/                        │   └── .claude-plugin/
│   ├── c3-query/                  │       ├── skills/c3/
│   └── c3-alter/                  │       │   ├── SKILL.md
├── agents/                        │       │   └── references/
│   ├── c3-navigator.md            │       ├── agents/
│   └── c3-orchestrator.md         │       └── plugin.json
├── references/                    │
│   ├── skill-harness.md           └── opencode/
│   └── layer-navigation.md            ├── skill/c3/
└── .claude-plugin/                    │   ├── SKILL.md
    └── plugin.json                    │   └── references/
                                       └── package.json
```

**Key decisions:**
- `main` remains for development (current structure preserved)
- `dist` branch contains built outputs, organized by target
- GitHub default branch set to `dist` for installs

### Build Transformation Rules

#### Skills

Each skill gets transformed to be self-contained with bundled references.

**Before (dev):**
```markdown
# In skills/c3-query/SKILL.md
## REQUIRED: Load References
Before proceeding, use Glob to find and Read these files:
1. `**/references/skill-harness.md`
2. `**/references/layer-navigation.md`
```

**After (built):**
```markdown
# In dist/claude-code/skills/c3-query/SKILL.md
## REQUIRED: Load References
Before proceeding, Read these files from this skill's directory:
1. `references/skill-harness.md`
2. `references/layer-navigation.md`
```

**Physical structure:**
```
dist/claude-code/skills/c3-query/
├── SKILL.md                    ← rewritten paths
└── references/
    ├── skill-harness.md        ← copied from root references/
    └── layer-navigation.md     ← copied from root references/
```

#### Agents

Agents don't directly reference the `references/` folder - they dispatch to other agents or skills. Therefore:

- Agents remain as single `.md` files
- Only need namespace rewriting for cross-agent dispatch
- No bundled references needed

**Namespace rewriting example:**

| Target | Namespace |
|--------|-----------|
| Claude Code | `c3-skill:c3-summarizer` (unchanged) |
| OpenCode | `opencode-c3:c3-summarizer` |

### Reference Resolution Logic

The build script scans skill content for reference patterns and bundles them.

**Patterns to match:**
```
`**/references/skill-harness.md`     → references/skill-harness.md
**/references/layer-navigation.md    → references/layer-navigation.md
```

**Algorithm:**
1. Parse skill/agent content for `**/references/*.md` patterns
2. Extract unique referenced filenames
3. Copy referenced files from root `references/` into skill's local `references/` folder
4. Rewrite paths in content from glob pattern to relative path

---

## Implementation

### Directory Structure

**Development (`main` branch):**
```
c3-design/
├── .claude-plugin/plugin.json
├── skills/
│   ├── c3/SKILL.md
│   ├── c3-query/SKILL.md
│   ├── c3-alter/SKILL.md
│   ├── c3-ref/SKILL.md
│   └── onboard/SKILL.md
├── agents/
│   ├── c3-navigator.md
│   ├── c3-orchestrator.md
│   ├── c3-summarizer.md
│   └── ...
├── references/                    ← shared source
│   ├── skill-harness.md
│   ├── layer-navigation.md
│   ├── adr-template.md
│   └── ...
├── scripts/
│   └── build.ts
└── .github/workflows/build-dist.yml
```

**Distribution (`dist` branch - default):**
```
claude-code/
├── .claude-plugin/plugin.json
├── skills/
│   ├── c3/
│   │   ├── SKILL.md
│   │   └── references/
│   ├── c3-query/
│   │   ├── SKILL.md
│   │   └── references/
│   └── ...
└── agents/
    ├── c3-navigator.md
    └── ...

opencode/
├── package.json
├── plugin.js
├── skill/
│   ├── c3/
│   │   ├── SKILL.md
│   │   └── references/
│   └── ...
└── agent/
    └── ...
```

### Build Script

**Location:** `scripts/build.ts`

```typescript
import { readdir, readFile, writeFile, mkdir, cp } from 'fs/promises';
import { join, dirname } from 'path';

const ROOT = process.cwd();
const DIST = join(ROOT, 'dist');
const TARGETS = ['claude-code', 'opencode'] as const;

// Reference patterns to match and rewrite
const REFERENCE_PATTERNS = [
  /`\*\*\/references\/([^`]+\.md)`/g,
  /\*\*\/references\/([^\s\)]+\.md)/g,
];

interface BuildConfig {
  target: typeof TARGETS[number];
  skillsDir: string;
  agentsDir: string;
  namespace: string;
}

const CONFIGS: Record<typeof TARGETS[number], BuildConfig> = {
  'claude-code': {
    target: 'claude-code',
    skillsDir: 'skills',
    agentsDir: 'agents',
    namespace: 'c3-skill',
  },
  'opencode': {
    target: 'opencode',
    skillsDir: 'skill',
    agentsDir: 'agent',
    namespace: 'opencode-c3',
  },
};

function extractReferences(content: string): string[] {
  const refs = new Set<string>();
  for (const pattern of REFERENCE_PATTERNS) {
    const regex = new RegExp(pattern.source, pattern.flags);
    for (const match of content.matchAll(regex)) {
      refs.add(match[1]);
    }
  }
  return [...refs];
}

function rewriteReferences(content: string): string {
  let result = content;
  for (const pattern of REFERENCE_PATTERNS) {
    result = result.replace(new RegExp(pattern.source, pattern.flags),
      (match, filename) => match.replace(`**/references/${filename}`, `references/${filename}`)
    );
  }
  return result;
}

function rewriteNamespace(content: string, fromNs: string, toNs: string): string {
  return content.replace(new RegExp(fromNs, 'g'), toNs);
}

async function bundleSkill(skillName: string, config: BuildConfig): Promise<void> {
  const srcDir = join(ROOT, 'skills', skillName);
  const destDir = join(DIST, config.target, config.skillsDir, skillName);

  // Read SKILL.md
  const skillPath = join(srcDir, 'SKILL.md');
  let content = await readFile(skillPath, 'utf-8');

  // Extract and bundle references
  const refs = extractReferences(content);
  if (refs.length > 0) {
    const refsDir = join(destDir, 'references');
    await mkdir(refsDir, { recursive: true });

    for (const ref of refs) {
      const srcRef = join(ROOT, 'references', ref);
      const destRef = join(refsDir, ref);
      await cp(srcRef, destRef);
    }
  }

  // Rewrite content
  content = rewriteReferences(content);
  if (config.namespace !== 'c3-skill') {
    content = rewriteNamespace(content, 'c3-skill', config.namespace);
  }

  // Write skill
  await mkdir(destDir, { recursive: true });
  await writeFile(join(destDir, 'SKILL.md'), content);
}

async function bundleAgent(agentFile: string, config: BuildConfig): Promise<void> {
  const srcPath = join(ROOT, 'agents', agentFile);
  const destDir = join(DIST, config.target, config.agentsDir);

  let content = await readFile(srcPath, 'utf-8');

  // Rewrite namespace for sub-agent dispatch
  if (config.namespace !== 'c3-skill') {
    content = rewriteNamespace(content, 'c3-skill', config.namespace);
  }

  await mkdir(destDir, { recursive: true });
  await writeFile(join(destDir, agentFile), content);
}

async function generateManifest(config: BuildConfig): Promise<void> {
  const destDir = join(DIST, config.target);

  if (config.target === 'claude-code') {
    // Copy and adapt plugin.json
    const manifest = JSON.parse(await readFile(join(ROOT, '.claude-plugin', 'plugin.json'), 'utf-8'));
    await mkdir(join(destDir, '.claude-plugin'), { recursive: true });
    await writeFile(join(destDir, '.claude-plugin', 'plugin.json'), JSON.stringify(manifest, null, 2));
  } else if (config.target === 'opencode') {
    // Generate OpenCode package.json
    const pkg = {
      name: 'opencode-c3',
      version: '1.0.0', // TODO: sync with source
      description: 'C3 architecture methodology for OpenCode',
      main: './plugin.js',
      type: 'module',
    };
    await writeFile(join(destDir, 'package.json'), JSON.stringify(pkg, null, 2));

    // Copy plugin.js if exists
    const pluginJs = join(ROOT, 'dist', 'opencode-c3', 'plugin.js');
    try {
      await cp(pluginJs, join(destDir, 'plugin.js'));
    } catch {
      // Generate minimal plugin.js
      await writeFile(join(destDir, 'plugin.js'), 'export default {};');
    }
  }
}

async function build(): Promise<void> {
  // Clean dist
  await mkdir(DIST, { recursive: true });

  // Get skills and agents
  const skills = await readdir(join(ROOT, 'skills'));
  const agents = (await readdir(join(ROOT, 'agents'))).filter(f => f.endsWith('.md'));

  for (const target of TARGETS) {
    const config = CONFIGS[target];
    console.log(`Building ${target}...`);

    // Bundle skills
    for (const skill of skills) {
      await bundleSkill(skill, config);
    }

    // Bundle agents
    for (const agent of agents) {
      await bundleAgent(agent, config);
    }

    // Generate manifest
    await generateManifest(config);
  }

  console.log('Build complete!');
}

build().catch(console.error);
```

### CI Pipeline

**Location:** `.github/workflows/build-dist.yml`

```yaml
name: Build Distribution

on:
  push:
    branches: [main]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Setup Bun
        uses: oven-sh/setup-bun@v2

      - name: Install dependencies
        run: bun install

      - name: Build all targets
        run: bun run scripts/build.ts

      - name: Configure git
        run: |
          git config user.name "github-actions[bot]"
          git config user.email "github-actions[bot]@users.noreply.github.com"

      - name: Push to dist branch
        run: |
          # Create orphan branch with just dist contents
          git checkout --orphan dist-temp
          git rm -rf .

          # Move built outputs to root
          mv dist/claude-code/* . 2>/dev/null || true
          mv dist/opencode opencode 2>/dev/null || true

          # Commit and push
          git add -A
          git commit -m "Build from ${GITHUB_SHA::7}"
          git push origin dist-temp:dist --force
```

### GitHub Configuration

After first CI run:
1. Go to repository Settings → General → Default branch
2. Change default branch from `main` to `dist`
3. Update any documentation that references installation

---

## Migration Plan

### Phase 1: Add Build Infrastructure
1. Create `scripts/build.ts`
2. Create `.github/workflows/build-dist.yml`
3. Test build locally with `bun run scripts/build.ts`
4. Verify output structure in `dist/`

### Phase 2: CI Setup
1. Push changes to `main`
2. Verify CI creates `dist` branch
3. Verify built skills have bundled references
4. Verify paths are rewritten correctly

### Phase 3: Switch Default Branch
1. Change GitHub default branch to `dist`
2. Update installation documentation
3. Test fresh install from `dist` branch
4. Verify skills work without full repo clone

### Phase 4: Cleanup
1. Remove old `dist/opencode-c3` folder from `main` (now generated)
2. Update CLAUDE.md with new workflow
3. Document build process in README

---

## Verification

After implementation, verify:

- [ ] `bun run scripts/build.ts` completes without errors
- [ ] `dist/claude-code/skills/*/references/` contains bundled files
- [ ] `dist/claude-code/skills/*/SKILL.md` has rewritten paths
- [ ] `dist/opencode/skill/*/SKILL.md` has `opencode-c3` namespace
- [ ] CI pushes to `dist` branch on merge to `main`
- [ ] Fresh clone of `dist` branch has working skills
- [ ] `claude plugin install` works from `dist` branch
