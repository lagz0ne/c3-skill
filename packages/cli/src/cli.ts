import { execFileSync } from 'node:child_process'
import { existsSync, readFileSync, readdirSync } from 'node:fs'
import { dirname, join } from 'node:path'
import { homedir } from 'node:os'

// --- Arg parsing ---

interface ParsedArgs {
  agent: string | undefined
  rest: string[]
}

function parseArgs(argv: string[]): ParsedArgs {
  const rest: string[] = []
  let agent: string | undefined

  for (let i = 0; i < argv.length; i++) {
    if (argv[i] === '--agent' && i + 1 < argv.length) {
      agent = argv[++i]
    } else {
      rest.push(argv[i])
    }
  }
  return { agent, rest }
}

// --- Version comparison ---

type Semver = [number, number, number]

function parseSemver(v: string): Semver | null {
  const m = v.trim().match(/^(\d+)\.(\d+)\.(\d+)$/)
  if (!m) return null
  return [Number(m[1]), Number(m[2]), Number(m[3])]
}

function compareSemver(a: Semver, b: Semver): number {
  for (let i = 0; i < 3; i++) {
    if (a[i] !== b[i]) return a[i] - b[i]
  }
  return 0
}

// --- Discovery ---

interface Candidate {
  binDir: string
  version: Semver
  priority: number
}

function readVersion(binDir: string): Semver | null {
  try {
    const raw = readFileSync(join(binDir, 'VERSION'), 'utf-8')
    return parseSemver(raw)
  } catch {
    return null
  }
}

function addCandidate(candidates: Candidate[], binDir: string, priority: number): void {
  const version = readVersion(binDir)
  if (version) candidates.push({ binDir, version, priority })
}

function discoverProjectScope(cwd: string): string[] {
  const results: string[] = []
  let dir = cwd
  while (true) {
    const candidate = join(dir, 'skills', 'c3', 'bin')
    if (existsSync(join(candidate, 'VERSION'))) {
      results.push(candidate)
    }
    if (existsSync(join(dir, '.git'))) break
    const parent = dirname(dir)
    if (parent === dir) break
    dir = parent
  }
  return results
}

function discoverMarketplace(): string[] {
  const base = join(homedir(), '.claude', 'plugins', 'marketplaces')
  if (!existsSync(base)) return []
  const results: string[] = []
  try {
    for (const entry of readdirSync(base)) {
      const candidate = join(base, entry, 'skills', 'c3', 'bin')
      if (existsSync(join(candidate, 'VERSION'))) {
        results.push(candidate)
      }
    }
  } catch { /* skip */ }
  return results
}

function discover(cwd: string, agentFilter: string | undefined): Candidate[] {
  const candidates: Candidate[] = []
  let priority = 0

  // 1. Project scope (always included)
  for (const binDir of discoverProjectScope(cwd)) {
    addCandidate(candidates, binDir, priority++)
  }

  // 2. Claude skills
  if (!agentFilter || agentFilter === 'claude') {
    addCandidate(candidates, join(homedir(), '.claude', 'skills', 'c3', 'bin'), priority++)
  }

  // 3. Codex skills
  if (!agentFilter || agentFilter === 'codex') {
    addCandidate(candidates, join(homedir(), '.codex', 'skills', 'c3', 'bin'), priority++)
  }

  // 4. Marketplace (claude scope)
  if (!agentFilter || agentFilter === 'claude') {
    for (const binDir of discoverMarketplace()) {
      addCandidate(candidates, binDir, priority++)
    }
  }

  return candidates
}

function pickBest(candidates: Candidate[]): Candidate | null {
  if (candidates.length === 0) return null
  return candidates.sort((a, b) => {
    const versionDiff = compareSemver(b.version, a.version)
    if (versionDiff !== 0) return versionDiff
    return a.priority - b.priority
  })[0]
}

// --- Main ---

const { agent, rest } = parseArgs(process.argv.slice(2))
const candidates = discover(process.cwd(), agent)
const best = pickBest(candidates)

if (!best) {
  console.error('error: No c3x installation found.')
  console.error('')
  console.error('Install via Claude Code skill marketplace or Codex:')
  console.error('  claude skill install c3-skill')
  console.error('')
  console.error('Searched:')
  if (!agent || agent === 'claude') {
    console.error('  ~/.claude/skills/c3/bin/')
    console.error('  ~/.claude/plugins/marketplaces/*/skills/c3/bin/')
  }
  if (!agent || agent === 'codex') {
    console.error('  ~/.codex/skills/c3/bin/')
  }
  console.error('  <project>/skills/c3/bin/ (walking up from cwd)')
  process.exit(1)
}

const c3xSh = join(best.binDir, 'c3x.sh')

try {
  execFileSync('bash', [c3xSh, ...rest], {
    stdio: 'inherit',
    cwd: process.cwd(),
  })
} catch (err: any) {
  process.exit(err.status ?? 1)
}
