import { spawnSync } from 'node:child_process'
import { createHash } from 'node:crypto'
import {
  chmodSync,
  existsSync,
  mkdirSync,
  readdirSync,
  readFileSync,
  rmSync,
  renameSync,
  writeFileSync,
} from 'node:fs'
import { get as httpsGet } from 'node:https'
import { homedir, platform as nodePlatform, arch as nodeArch } from 'node:os'
import { basename, dirname, join, resolve } from 'node:path'
import { AST_GREP_VERSION, C3X_VERSION, SEMANTIC_MODEL_REVISION } from './version.js'

const RELEASE_REPO = 'https://github.com/lagz0ne/c3-skill/releases/download'
const RELEASES_API = 'https://api.github.com/repos/lagz0ne/c3-skill/releases?per_page=100'
const PROJECT_RUNTIME_FILE = 'runtime.json'
const VERSION_RE = /^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-[0-9A-Za-z.-]+)?(?:\+[0-9A-Za-z.-]+)?$/
const FIRST_AST_GREP_RUNTIME_VERSION = '11.5.0'

export interface PlatformTarget {
  os: string
  arch: string
}

export interface ManagerEnv {
  HOME?: string
  XDG_CACHE_HOME?: string
  C3X_VERSION?: string
  C3X_RELEASE_BASE_URL?: string
  C3X_RELEASES_URL?: string
  C3X_SKIP_MODEL_DOWNLOAD?: string
  [key: string]: string | undefined
}

export interface DownloadClient {
  download(url: string, progress?: (downloadedBytes: number, totalBytes?: number) => void): Promise<Uint8Array>
}

export interface ManagerOptions {
  env?: ManagerEnv
  cwd?: string
  platform?: string
  arch?: string
  downloader?: DownloadClient
  progress?: (event: ProgressEvent) => void
  stdout?: (line: string) => void
  stderr?: (line: string) => void
  exec?: (file: string, args: string[], env: NodeJS.ProcessEnv, cwd: string) => number
}

export interface RuntimeCommandOptions extends ManagerOptions {
  force?: boolean
}

export interface PreparedRuntime {
  version: string
  cacheDir: string
  binaryPath: string
  astGrepPath?: string
  modelPath: string
  vocabPath: string
}

export interface CachedAsset {
  name: string
  path: string
  sha256: string
  cached: boolean
}

export interface RuntimeManifest {
  version: string
  target: PlatformTarget
  semanticModelRevision: string
  installedAt: string
  assets: {
    binary: CachedAsset
    astGrep?: CachedAsset
    model?: CachedAsset
    vocab?: CachedAsset
  }
}

export interface InstalledRuntime {
  version: string
  path: string
  manifest?: RuntimeManifest
}

export interface ProjectRuntimeConfig {
  version: string
}

export type ProgressEvent =
  | { kind: 'asset-start'; version: string; assetName: string }
  | { kind: 'asset-progress'; version: string; assetName: string; downloadedBytes: number; totalBytes?: number }
  | { kind: 'asset-downloaded'; version: string; assetName: string; bytes: number }
  | { kind: 'asset-ready'; version: string; assetName: string; path: string; cached: boolean; sha256: string }
  | { kind: 'runtime-ready'; version: string; cacheDir: string }

export function resolvePlatform(platform = nodePlatform(), arch = nodeArch()): PlatformTarget {
  const os = platform === 'darwin' || platform === 'linux' ? platform : ''
  const mappedArch = arch === 'x64' ? 'amd64' : arch === 'arm64' ? 'arm64' : ''
  if (!os || !mappedArch || (os === 'darwin' && mappedArch !== 'arm64')) {
    throw new Error(`error: unsupported platform ${platform}/${arch}\nhint: @c3x/cli supports linux x64/arm64 and darwin arm64`)
  }
  return { os, arch: mappedArch }
}

export function resolveVersion(env: ManagerEnv = process.env): string {
  return normalizeVersion(env.C3X_VERSION || C3X_VERSION)
}

export function cacheBase(env: ManagerEnv = process.env): string {
  if (env.XDG_CACHE_HOME && env.XDG_CACHE_HOME.trim() !== '') {
    return env.XDG_CACHE_HOME
  }
  return join(env.HOME || homedir(), '.cache')
}

export function cacheDirForVersion(version: string, env: ManagerEnv = process.env): string {
  return join(cacheBase(env), 'c3x', normalizeVersion(version))
}

export function releaseBaseURL(version: string, env: ManagerEnv = process.env): string {
  if (env.C3X_RELEASE_BASE_URL && env.C3X_RELEASE_BASE_URL.trim() !== '') {
    return validateHttpsURL(env.C3X_RELEASE_BASE_URL.replace(/\/+$/, ''), 'C3X_RELEASE_BASE_URL')
  }
  return `${RELEASE_REPO}/v${normalizeVersion(version)}`
}

export function releasesURL(env: ManagerEnv = process.env): string {
  if (env.C3X_RELEASES_URL && env.C3X_RELEASES_URL.trim() !== '') {
    return validateHttpsURL(env.C3X_RELEASES_URL, 'C3X_RELEASES_URL')
  }
  return RELEASES_API
}

export function assetNames(version: string, target: PlatformTarget): { binary: string; astGrep: string; model: string; vocab: string } {
  const normalized = normalizeVersion(version)
  return {
    binary: `c3x-${normalized}-${target.os}-${target.arch}`,
    astGrep: `ast-grep-${AST_GREP_VERSION}-${target.os}-${target.arch}`,
    model: `c3x-semantic-model-all-MiniLM-L6-v2-${SEMANTIC_MODEL_REVISION}.onnx`,
    vocab: `c3x-semantic-vocab-all-MiniLM-L6-v2-${SEMANTIC_MODEL_REVISION}.txt`,
  }
}

function runtimeIncludesAstGrep(version: string): boolean {
  return compareVersionsAsc(normalizeVersion(version), FIRST_AST_GREP_RUNTIME_VERSION) >= 0
}

export async function fetchAvailableVersions(options: ManagerOptions = {}): Promise<string[]> {
  const env = options.env || process.env
  const downloader = options.downloader || new HttpDownloadClient()
  const text = Buffer.from(await downloader.download(releasesURL(env))).toString('utf8')
  let releases: unknown
  try {
    releases = JSON.parse(text)
  } catch {
    throw new Error('error: release index did not contain valid JSON')
  }
  if (!Array.isArray(releases)) {
    throw new Error('error: release index did not contain a release list')
  }

  const versions = new Set<string>()
  for (const release of releases) {
    if (!release || typeof release !== 'object') continue
    const data = release as { tag_name?: unknown; name?: unknown; draft?: unknown; prerelease?: unknown }
    if (data.draft === true || data.prerelease === true) continue
    const raw = typeof data.tag_name === 'string' ? data.tag_name : typeof data.name === 'string' ? data.name : ''
    const version = tryNormalizeVersion(raw)
    if (version) versions.add(version)
  }

  return [...versions].sort(compareVersionsDesc)
}

export async function installRuntime(options: ManagerOptions & { version?: string } = {}): Promise<PreparedRuntime> {
  const version = await resolveVersionSpec(options.version || 'latest', options)
  return prepareRuntimeForVersion(version, options)
}

export async function prepareRuntime(options: ManagerOptions = {}): Promise<PreparedRuntime> {
  const version = await resolveRuntimeVersion(options)
  return prepareRuntimeForVersion(version, options)
}

export function listInstalledRuntimes(options: ManagerOptions = {}): InstalledRuntime[] {
  const root = join(cacheBase(options.env || process.env), 'c3x')
  if (!existsSync(root)) return []
  const installed: InstalledRuntime[] = []
  for (const entry of readdirSync(root, { withFileTypes: true })) {
    if (!entry.isDirectory()) continue
    const version = tryNormalizeVersion(entry.name)
    if (!version) continue
    const path = join(root, entry.name)
    installed.push({ version, path, manifest: readRuntimeManifest(path) })
  }
  return installed.sort((a, b) => compareVersionsDesc(a.version, b.version))
}

export function uninstallRuntime(
  version: string,
  options: RuntimeCommandOptions = {},
): string {
  const normalized = normalizeVersion(version)
  const selected = readProjectRuntimeConfig(options.cwd || process.cwd())
  if (!options.force && selected?.version === normalized) {
    throw new Error(`error: cannot uninstall project selected runtime ${normalized}\nhint: run 'c3x runtime use <version>' first, or pass --force`)
  }
  const dir = cacheDirForVersion(normalized, options.env || process.env)
  rmSync(dir, { recursive: true, force: true })
  return dir
}

export function pruneRuntimeCache(options: RuntimeCommandOptions = {}): string[] {
  const env = options.env || process.env
  const installed = listInstalledRuntimes(options)
  const selected = readProjectRuntimeConfig(options.cwd || process.cwd())
  const installedVersions = new Set(installed.map((runtime) => runtime.version))
  const keepVersion = selected && selected.version !== 'latest' && installedVersions.has(selected.version)
    ? selected.version
    : installed[0]?.version
  const removed: string[] = []

  for (const runtime of installed) {
    if (runtime.version === keepVersion) continue
    rmSync(runtime.path, { recursive: true, force: true })
    removed.push(runtime.path)
  }

  const root = join(cacheBase(env), 'c3x')
  return removed.map((path) => resolve(path)).sort()
}

export function findProjectC3Dir(cwd = process.cwd()): string | undefined {
  let current = resolve(cwd)
  while (true) {
    const candidate = join(current, '.c3')
    if (existsSync(candidate)) return candidate
    const next = dirname(current)
    if (next === current) return undefined
    current = next
  }
}

export function readProjectRuntimeConfig(cwd = process.cwd()): ProjectRuntimeConfig | undefined {
  const c3Dir = findProjectC3Dir(cwd)
  if (!c3Dir) return undefined
  const path = join(c3Dir, PROJECT_RUNTIME_FILE)
  if (!existsSync(path)) return undefined

  let raw: unknown
  try {
    raw = JSON.parse(readFileSync(path, 'utf8'))
  } catch {
    throw new Error(`error: invalid project runtime metadata at ${path}`)
  }
  if (!raw || typeof raw !== 'object') {
    throw new Error(`error: invalid project runtime metadata at ${path}`)
  }
  const data = raw as Record<string, unknown>
  for (const key of Object.keys(data)) {
    if (key !== 'version') {
      throw new Error(`error: invalid project runtime metadata field ${key}\nhint: project metadata may store only a runtime version`)
    }
  }
  if (typeof data.version !== 'string') {
    throw new Error(`error: invalid project runtime version in ${path}`)
  }
  const version = data.version === 'latest' ? 'latest' : normalizeVersion(data.version, 'project runtime version')
  return { version }
}

export function writeProjectRuntimeConfig(cwd: string, version: string): ProjectRuntimeConfig {
  const c3Dir = findProjectC3Dir(cwd)
  if (!c3Dir) {
    throw new Error("error: no .c3 directory found\nhint: run 'c3x runtime use' from inside a C3 project")
  }
  const normalized = version === 'latest' ? 'latest' : normalizeVersion(version)
  const config = { version: normalized }
  writeFileSync(join(c3Dir, PROJECT_RUNTIME_FILE), `${JSON.stringify(config, null, 2)}\n`)
  return config
}

export function gcOldVersions(root: string, keepVersion: string): string[] {
  if (!existsSync(root)) return []
  const normalizedKeep = normalizeVersion(keepVersion)
  const removed: string[] = []
  for (const entry of readdirSync(root, { withFileTypes: true })) {
    if (!entry.isDirectory() || entry.name === normalizedKeep) continue
    const path = join(root, entry.name)
    rmSync(path, { recursive: true, force: true })
    removed.push(path)
  }
  return removed
}

export async function ensureCachedAsset(opts: EnsureAssetOptions): Promise<CachedAsset> {
  const checksumURL = `${opts.baseURL}/${opts.assetName}.sha256`
  opts.progress?.({ kind: 'asset-start', version: opts.version, assetName: opts.assetName })
  const expected = parseSHA256(Buffer.from(await opts.downloader.download(checksumURL)).toString('utf8'), opts.assetName)
  if (existsSync(opts.targetPath) && sha256File(opts.targetPath) === expected) {
    if (opts.executable) chmodSync(opts.targetPath, 0o755)
    const asset = { name: opts.assetName, path: opts.targetPath, sha256: expected, cached: true }
    opts.progress?.({ kind: 'asset-ready', version: opts.version, assetName: opts.assetName, path: opts.targetPath, cached: true, sha256: expected })
    return asset
  }

  mkdirSync(dirname(opts.targetPath), { recursive: true })
  const tmp = `${opts.targetPath}.tmp`
  const data = Buffer.from(await opts.downloader.download(`${opts.baseURL}/${opts.assetName}`, (downloadedBytes, totalBytes) => {
    opts.progress?.({ kind: 'asset-progress', version: opts.version, assetName: opts.assetName, downloadedBytes, totalBytes })
  }))
  opts.progress?.({ kind: 'asset-downloaded', version: opts.version, assetName: opts.assetName, bytes: data.length })
  const got = sha256Buffer(data)
  if (got !== expected) {
    rmSync(tmp, { force: true })
    throw new Error(`error: checksum mismatch for ${opts.assetName}\nhint: clear ${dirname(opts.targetPath)} and retry`)
  }
  writeFileSync(tmp, data)
  if (opts.executable) chmodSync(tmp, 0o755)
  renameSync(tmp, opts.targetPath)
  const asset = { name: opts.assetName, path: opts.targetPath, sha256: expected, cached: false }
  opts.progress?.({ kind: 'asset-ready', version: opts.version, assetName: opts.assetName, path: opts.targetPath, cached: false, sha256: expected })
  return asset
}

export function parseSHA256(text: string, assetName: string): string {
  for (const field of text.trim().split(/\s+/)) {
    if (/^[a-fA-F0-9]{64}$/.test(field)) {
      return field.toLowerCase()
    }
  }
  throw new Error(`error: checksum for ${assetName} did not contain a sha256 digest`)
}

export async function runCli(argv: string[], options: ManagerOptions = {}): Promise<number> {
  const stdout = options.stdout || ((line: string) => console.log(line))
  const stderr = options.stderr || ((line: string) => console.error(line))
  if (isRootHelpCommand(argv)) {
    printRootHelp(stdout)
    return 0
  }
  if (isRootVersionCommand(argv)) {
    stdout(C3X_VERSION)
    return 0
  }

  const cwd = options.cwd || process.cwd()
  const runtime = await prepareRuntime({
    ...options,
    progress: options.progress || createProgressReporter(stderr, { mode: 'auto' }),
  })
  const env = {
    ...process.env,
    ...(options.env || {}),
    C3X_VERSION: runtime.version,
    C3_SEMANTIC_CACHE_DIR: join(runtime.cacheDir, 'semantic'),
  }
  if (runtime.astGrepPath) {
    env.C3_AST_GREP = runtime.astGrepPath
  }
  const exec = options.exec || defaultExec
  return exec(runtime.binaryPath, argv, env, cwd)
}

export async function runManagerCommand(argv: string[], options: RuntimeCommandOptions = {}): Promise<number> {
  const stdout = options.stdout || ((line: string) => console.log(line))
  const stderr = options.stderr || ((line: string) => console.error(line))
  const [command, ...rest] = argv

  switch (command) {
    case 'versions': {
      const available = await fetchAvailableVersions(options)
      const installed = new Set(listInstalledRuntimes(options).map((runtime) => runtime.version))
      for (const version of available) {
        const suffix = installed.has(version) ? ' installed' : ''
        stdout(`${version}${suffix}`)
      }
      return 0
    }
    case 'installed': {
      for (const runtime of listInstalledRuntimes(options)) {
        stdout(runtime.version)
      }
      return 0
    }
    case 'install': {
      const version = rest[0] || 'latest'
      const runtime = await installRuntime({
        ...options,
        version,
        progress: options.progress || createProgressReporter(stderr, { mode: 'install' }),
      })
      stdout(`installed ${runtime.version}`)
      return 0
    }
    case 'use': {
      const version = rest[0]
      if (!version) {
        throw new Error("error: missing runtime version\nhint: usage: c3x runtime use <version|latest>")
      }
      const config = writeProjectRuntimeConfig(options.cwd || process.cwd(), version)
      stdout(`project runtime ${config.version}`)
      return 0
    }
    case 'uninstall': {
      const version = rest.find((arg) => arg !== '--force')
      if (!version) {
        throw new Error("error: missing runtime version\nhint: usage: c3x runtime uninstall <version> [--force]")
      }
      const removed = uninstallRuntime(version, { ...options, force: options.force || rest.includes('--force') })
      stdout(`removed ${basename(removed)}`)
      return 0
    }
    case 'prune': {
      const removed = pruneRuntimeCache(options)
      for (const path of removed) stdout(`removed ${basename(path)}`)
      if (removed.length === 0) stdout('nothing to prune')
      return 0
    }
    default:
      throw new Error(`error: unknown runtime command ${command || ''}\nhint: usage: c3x runtime <versions|installed|install|use|uninstall|prune>`)
  }
}

export function isRootHelpCommand(argv: string[]): boolean {
  if (argv.length === 0) return true
  return argv.length === 1 && (argv[0] === '--help' || argv[0] === '-h' || argv[0] === 'help')
}

export function isRootVersionCommand(argv: string[]): boolean {
  return argv.length === 1 && (argv[0] === '--version' || argv[0] === 'version')
}

export function printRootHelp(stdout: (line: string) => void): void {
  stdout(`c3x ${C3X_VERSION}`)
  stdout('')
  stdout('Usage:')
  stdout('  c3x <command> [args...]')
  stdout('  c3x runtime <versions|installed|install|use|uninstall|prune>')
  stdout('')
  stdout('Runtime commands:')
  stdout('  c3x runtime versions')
  stdout('  c3x runtime installed')
  stdout('  c3x runtime install [version|latest]')
  stdout('  c3x runtime use <version|latest>')
  stdout('  c3x runtime uninstall <version> [--force]')
  stdout('  c3x runtime prune')
  stdout('')
  stdout('Normal C3 commands are forwarded to the selected runtime. The first real command may install runtime assets.')
}

export function createProgressReporter(
  writeLine: (line: string) => void,
  options: { mode?: 'install' | 'auto' } = {},
): (event: ProgressEvent) => void {
  const mode = options.mode || 'install'
  const assets = new Map<string, { lastBucket: number; lastBytes: number }>()
  let announcedAutoInstall = false
  let downloadedAnything = false

  return (event) => {
    if (event.kind === 'asset-start') {
      assets.set(event.assetName, { lastBucket: -10, lastBytes: 0 })
      if (mode === 'install') writeLine(formatProgressEvent(event))
      return
    }

    if (event.kind === 'asset-progress') {
      if (mode === 'auto' && !announcedAutoInstall) {
        writeLine(`c3x runtime ${event.version}: downloading required runtime assets`)
        announcedAutoInstall = true
      }
      downloadedAnything = true
      if (shouldRenderProgress(assets, event)) {
        writeLine(formatProgressEvent(event))
      }
      return
    }

    if (event.kind === 'asset-downloaded') {
      if (mode === 'install' || downloadedAnything) {
        writeLine(formatProgressEvent(event))
      }
      return
    }

    if (event.kind === 'asset-ready') {
      if (mode === 'install' || (!event.cached && downloadedAnything)) {
        writeLine(formatProgressEvent(event))
      }
      return
    }

    if (mode === 'install' || downloadedAnything) {
      writeLine(formatProgressEvent(event))
    }
  }
}

export function formatProgressEvent(event: ProgressEvent): string {
  switch (event.kind) {
    case 'asset-start':
      return `c3x runtime ${event.version}: preparing ${event.assetName}`
    case 'asset-progress':
      if (event.totalBytes && event.totalBytes > 0) {
        const pct = Math.floor((event.downloadedBytes / event.totalBytes) * 100)
        return `c3x runtime ${event.version}: ${event.assetName} ${progressBar(pct)} ${pct}% ${formatBytes(event.downloadedBytes)}/${formatBytes(event.totalBytes)}`
      }
      return `c3x runtime ${event.version}: ${event.assetName} ${formatBytes(event.downloadedBytes)}`
    case 'asset-downloaded':
      return `c3x runtime ${event.version}: downloaded ${event.assetName} (${formatBytes(event.bytes)})`
    case 'asset-ready':
      return `c3x runtime ${event.version}: ${event.cached ? 'cached' : 'verified'} ${event.assetName}`
    case 'runtime-ready':
      return `c3x runtime ${event.version}: runtime ready at ${event.cacheDir}`
  }
}

function shouldRenderProgress(
  assets: Map<string, { lastBucket: number; lastBytes: number }>,
  event: Extract<ProgressEvent, { kind: 'asset-progress' }>,
): boolean {
  const state = assets.get(event.assetName) || { lastBucket: -10, lastBytes: 0 }
  assets.set(event.assetName, state)

  if (event.totalBytes && event.totalBytes > 0) {
    const pct = Math.min(100, Math.floor((event.downloadedBytes / event.totalBytes) * 100))
    const bucket = pct === 100 ? 100 : Math.floor(pct / 10) * 10
    if (bucket > state.lastBucket) {
      state.lastBucket = bucket
      state.lastBytes = event.downloadedBytes
      return true
    }
    return false
  }

  const oneMiB = 1024 * 1024
  if (state.lastBytes === 0 || event.downloadedBytes - state.lastBytes >= oneMiB) {
    state.lastBytes = event.downloadedBytes
    return true
  }
  return false
}

function progressBar(percent: number): string {
  const width = 20
  const clamped = Math.max(0, Math.min(100, percent))
  const filled = Math.round((clamped / 100) * width)
  return `[${'#'.repeat(filled)}${'.'.repeat(width - filled)}]`
}

function formatBytes(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`
  const kib = bytes / 1024
  if (kib < 1024) return `${kib.toFixed(1)} KiB`
  const mib = kib / 1024
  if (mib < 1024) return `${mib.toFixed(1)} MiB`
  return `${(mib / 1024).toFixed(1)} GiB`
}

interface EnsureAssetOptions {
  version: string
  assetName: string
  targetPath: string
  baseURL: string
  executable?: boolean
  downloader: DownloadClient
  progress?: (event: ProgressEvent) => void
}

async function prepareRuntimeForVersion(version: string, options: ManagerOptions): Promise<PreparedRuntime> {
  const env = options.env || process.env
  const normalized = normalizeVersion(version)
  const target = resolvePlatform(options.platform, options.arch)
  const cacheDir = cacheDirForVersion(normalized, env)
  const names = assetNames(normalized, target)
  const baseURL = releaseBaseURL(normalized, env)
  const downloader = options.downloader || new HttpDownloadClient()

  mkdirSync(cacheDir, { recursive: true })

  const binaryPath = join(cacheDir, names.binary)
  const binary = await ensureCachedAsset({
    version: normalized,
    assetName: names.binary,
    targetPath: binaryPath,
    baseURL,
    executable: true,
    downloader,
    progress: options.progress,
  })
  let astGrepPath: string | undefined
  let astGrep: CachedAsset | undefined
  if (runtimeIncludesAstGrep(normalized)) {
    astGrepPath = join(cacheDir, names.astGrep)
    astGrep = await ensureCachedAsset({
      version: normalized,
      assetName: names.astGrep,
      targetPath: astGrepPath,
      baseURL,
      executable: true,
      downloader,
      progress: options.progress,
    })
  }

  const modelPath = join(cacheDir, 'semantic', 'models', `all-MiniLM-L6-v2-${SEMANTIC_MODEL_REVISION}`, 'model.onnx')
  const vocabPath = join(cacheDir, 'semantic', 'models', `all-MiniLM-L6-v2-${SEMANTIC_MODEL_REVISION}`, 'vocab.txt')
  let model: CachedAsset | undefined
  let vocab: CachedAsset | undefined
  if (!env.C3X_SKIP_MODEL_DOWNLOAD) {
    model = await ensureCachedAsset({ version: normalized, assetName: names.model, targetPath: modelPath, baseURL, downloader, progress: options.progress })
    vocab = await ensureCachedAsset({ version: normalized, assetName: names.vocab, targetPath: vocabPath, baseURL, downloader, progress: options.progress })
  }

  writeRuntimeManifest(cacheDir, {
    version: normalized,
    target,
    semanticModelRevision: SEMANTIC_MODEL_REVISION,
    installedAt: new Date().toISOString(),
    assets: { binary, astGrep, model, vocab },
  })
  options.progress?.({ kind: 'runtime-ready', version: normalized, cacheDir })

  return { version: normalized, cacheDir, binaryPath, astGrepPath, modelPath, vocabPath }
}

async function resolveRuntimeVersion(options: ManagerOptions): Promise<string> {
  const env = options.env || process.env
  if (env.C3X_VERSION && env.C3X_VERSION.trim() !== '') {
    return normalizeVersion(env.C3X_VERSION)
  }
  const project = readProjectRuntimeConfig(options.cwd || process.cwd())
  if (project && project.version !== 'latest') {
    return project.version
  }
  try {
    return await resolveVersionSpec('latest', options)
  } catch (err) {
    const installed = listInstalledRuntimes(options)[0]
    if (installed) return installed.version
    throw err
  }
}

async function resolveVersionSpec(version: string, options: ManagerOptions): Promise<string> {
  if (version === 'latest') {
    const versions = await fetchAvailableVersions(options)
    if (versions.length === 0) {
      throw new Error('error: no C3 releases found')
    }
    return versions[0]
  }
  return normalizeVersion(version)
}

function readRuntimeManifest(cacheDir: string): RuntimeManifest | undefined {
  const path = join(cacheDir, 'manifest.json')
  if (!existsSync(path)) return undefined
  try {
    return JSON.parse(readFileSync(path, 'utf8')) as RuntimeManifest
  } catch {
    return undefined
  }
}

function writeRuntimeManifest(cacheDir: string, manifest: RuntimeManifest): void {
  writeFileSync(join(cacheDir, 'manifest.json'), `${JSON.stringify(manifest, null, 2)}\n`)
}

function normalizeVersion(version: string, label = 'version'): string {
  const normalized = version.trim().replace(/^v/, '')
  if (!VERSION_RE.test(normalized)) {
    throw new Error(`error: invalid ${label}: ${version}`)
  }
  return normalized
}

function tryNormalizeVersion(version: string): string | undefined {
  try {
    return normalizeVersion(version)
  } catch {
    return undefined
  }
}

function compareVersionsDesc(a: string, b: string): number {
  return -compareVersionsAsc(a, b)
}

function compareVersionsAsc(a: string, b: string): number {
  const parsedA = parseVersion(a)
  const parsedB = parseVersion(b)
  for (const part of ['major', 'minor', 'patch'] as const) {
    if (parsedA[part] !== parsedB[part]) return parsedA[part] - parsedB[part]
  }
  if (parsedA.prerelease === parsedB.prerelease) return 0
  if (!parsedA.prerelease) return 1
  if (!parsedB.prerelease) return -1
  return parsedA.prerelease.localeCompare(parsedB.prerelease)
}

function parseVersion(version: string): { major: number; minor: number; patch: number; prerelease: string } {
  const [core, prerelease = ''] = version.split('-', 2)
  const [major, minor, patch] = core.split('.').map((part) => Number.parseInt(part, 10))
  return { major, minor, patch, prerelease }
}

function validateHttpsURL(url: string, envName: string): string {
  let parsed: URL
  try {
    parsed = new URL(url)
  } catch {
    throw new Error(`error: invalid ${envName}`)
  }
  if (parsed.protocol !== 'https:') {
    throw new Error(`error: invalid ${envName}: only https URLs are supported`)
  }
  return parsed.toString().replace(/\/+$/, '')
}

function defaultExec(file: string, args: string[], env: NodeJS.ProcessEnv, cwd: string): number {
  const result = spawnSync(file, args, { stdio: 'inherit', cwd, env })
  if (result.error) {
    throw result.error
  }
  const status = result.status ?? 1
  process.exit(status)
}

function sha256File(path: string): string {
  return sha256Buffer(readFileSync(path))
}

function sha256Buffer(data: Uint8Array): string {
  return createHash('sha256').update(data).digest('hex')
}

class HttpDownloadClient implements DownloadClient {
  async download(url: string, progress?: (downloadedBytes: number, totalBytes?: number) => void): Promise<Uint8Array> {
    return this.fetch(url, 0, progress)
  }

  private fetch(url: string, redirects: number, progress?: (downloadedBytes: number, totalBytes?: number) => void): Promise<Uint8Array> {
    return new Promise((resolvePromise, reject) => {
      const parsed = new URL(url)
      if (parsed.protocol !== 'https:') {
        reject(new Error(`download ${url}: only https URLs are supported`))
        return
      }
      const req = httpsGet(url, {
        headers: {
          'User-Agent': '@c3x/cli',
          Accept: 'application/vnd.github+json, application/octet-stream',
        },
      }, (res) => {
        const status = res.statusCode ?? 0
        // GitHub release assets 302-redirect to a CDN.
        if (status >= 300 && status < 400 && res.headers.location) {
          res.resume()
          if (redirects >= 5) {
            reject(new Error(`download ${url}: too many redirects`))
            return
          }
          const next = new URL(res.headers.location, url).toString()
          this.fetch(next, redirects + 1, progress).then(resolvePromise, reject)
          return
        }
        if (status < 200 || status >= 300) {
          reject(new Error(`download ${url}: status ${status}`))
          res.resume()
          return
        }
        const totalHeader = res.headers['content-length']
        const totalBytes = typeof totalHeader === 'string' ? Number.parseInt(totalHeader, 10) : undefined
        let downloadedBytes = 0
        const chunks: Buffer[] = []
        res.on('data', (chunk) => {
          const buffer = Buffer.from(chunk)
          chunks.push(buffer)
          downloadedBytes += buffer.length
          progress?.(downloadedBytes, totalBytes && Number.isFinite(totalBytes) ? totalBytes : undefined)
        })
        res.on('end', () => resolvePromise(Buffer.concat(chunks)))
      })
      req.on('error', (err) => reject(new Error(`download ${url}: ${err.message}\nhint: connect to GitHub Releases, or prefill the @c3x/cli cache`)))
      req.setTimeout(15 * 60 * 1000, () => {
        req.destroy(new Error(`download ${url}: timed out`))
      })
    })
  }
}
