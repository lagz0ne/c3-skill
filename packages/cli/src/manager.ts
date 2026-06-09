import { spawnSync } from 'node:child_process'
import { createHash } from 'node:crypto'
import { chmodSync, existsSync, mkdirSync, readdirSync, readFileSync, rmSync, renameSync, writeFileSync } from 'node:fs'
import { get as httpGet } from 'node:http'
import { get as httpsGet } from 'node:https'
import { homedir, platform as nodePlatform, arch as nodeArch } from 'node:os'
import { dirname, join } from 'node:path'
import { C3X_VERSION, SEMANTIC_MODEL_REVISION } from './version.js'

const RELEASE_REPO = 'https://github.com/lagz0ne/c3-skill/releases/download'

export interface PlatformTarget {
  os: string
  arch: string
}

export interface ManagerEnv {
  HOME?: string
  XDG_CACHE_HOME?: string
  C3X_VERSION?: string
  C3X_RELEASE_BASE_URL?: string
  C3X_SKIP_MODEL_DOWNLOAD?: string
  [key: string]: string | undefined
}

export interface DownloadClient {
  download(url: string): Promise<Uint8Array>
}

export interface ManagerOptions {
  env?: ManagerEnv
  cwd?: string
  platform?: string
  arch?: string
  downloader?: DownloadClient
  exec?: (file: string, args: string[], env: NodeJS.ProcessEnv, cwd: string) => number
}

export interface PreparedRuntime {
  version: string
  cacheDir: string
  binaryPath: string
  modelPath: string
  vocabPath: string
}

export function resolvePlatform(platform = nodePlatform(), arch = nodeArch()): PlatformTarget {
  const os = platform === 'darwin' || platform === 'linux' ? platform : ''
  const mappedArch = arch === 'x64' ? 'amd64' : arch === 'arm64' ? 'arm64' : ''
  if (!os || !mappedArch) {
    throw new Error(`error: unsupported platform ${platform}/${arch}\nhint: @c3x/cli supports linux/darwin on x64/arm64`)
  }
  return { os, arch: mappedArch }
}

export function resolveVersion(env: ManagerEnv = process.env): string {
  return (env.C3X_VERSION || C3X_VERSION).trim()
}

export function cacheBase(env: ManagerEnv = process.env): string {
  if (env.XDG_CACHE_HOME && env.XDG_CACHE_HOME.trim() !== '') {
    return env.XDG_CACHE_HOME
  }
  return join(env.HOME || homedir(), '.cache')
}

export function cacheDirForVersion(version: string, env: ManagerEnv = process.env): string {
  return join(cacheBase(env), 'c3x', version)
}

export function releaseBaseURL(version: string, env: ManagerEnv = process.env): string {
  if (env.C3X_RELEASE_BASE_URL && env.C3X_RELEASE_BASE_URL.trim() !== '') {
    return env.C3X_RELEASE_BASE_URL.replace(/\/+$/, '')
  }
  return `${RELEASE_REPO}/v${version}`
}

export function assetNames(version: string, target: PlatformTarget): { binary: string; model: string; vocab: string } {
  return {
    binary: `c3x-${version}-${target.os}-${target.arch}`,
    model: `c3x-semantic-model-all-MiniLM-L6-v2-${SEMANTIC_MODEL_REVISION}.onnx`,
    vocab: `c3x-semantic-vocab-all-MiniLM-L6-v2-${SEMANTIC_MODEL_REVISION}.txt`,
  }
}

export async function prepareRuntime(options: ManagerOptions = {}): Promise<PreparedRuntime> {
  const env = options.env || process.env
  const version = resolveVersion(env)
  const target = resolvePlatform(options.platform, options.arch)
  const cacheDir = cacheDirForVersion(version, env)
  const names = assetNames(version, target)
  const baseURL = releaseBaseURL(version, env)
  const downloader = options.downloader || new HttpDownloadClient()

  mkdirSync(cacheDir, { recursive: true })
  gcOldVersions(dirname(cacheDir), version)

  const binaryPath = join(cacheDir, names.binary)
  await ensureCachedAsset({ assetName: names.binary, targetPath: binaryPath, baseURL, executable: true, downloader })

  const modelPath = join(cacheDir, 'semantic', 'models', `all-MiniLM-L6-v2-${SEMANTIC_MODEL_REVISION}`, 'model.onnx')
  const vocabPath = join(cacheDir, 'semantic', 'models', `all-MiniLM-L6-v2-${SEMANTIC_MODEL_REVISION}`, 'vocab.txt')
  if (!env.C3X_SKIP_MODEL_DOWNLOAD) {
    await ensureCachedAsset({ assetName: names.model, targetPath: modelPath, baseURL, downloader })
    await ensureCachedAsset({ assetName: names.vocab, targetPath: vocabPath, baseURL, downloader })
  }

  return { version, cacheDir, binaryPath, modelPath, vocabPath }
}

interface EnsureAssetOptions {
  assetName: string
  targetPath: string
  baseURL: string
  executable?: boolean
  downloader: DownloadClient
}

export async function ensureCachedAsset(opts: EnsureAssetOptions): Promise<void> {
  const checksumURL = `${opts.baseURL}/${opts.assetName}.sha256`
  const expected = parseSHA256(Buffer.from(await opts.downloader.download(checksumURL)).toString('utf8'), opts.assetName)
  if (existsSync(opts.targetPath) && sha256File(opts.targetPath) === expected) {
    if (opts.executable) chmodSync(opts.targetPath, 0o755)
    return
  }

  mkdirSync(dirname(opts.targetPath), { recursive: true })
  const tmp = `${opts.targetPath}.tmp`
  const data = Buffer.from(await opts.downloader.download(`${opts.baseURL}/${opts.assetName}`))
  const got = sha256Buffer(data)
  if (got !== expected) {
    throw new Error(`error: checksum mismatch for ${opts.assetName}\nhint: clear ${dirname(opts.targetPath)} and retry`)
  }
  writeFileSync(tmp, data)
  if (opts.executable) chmodSync(tmp, 0o755)
  renameSync(tmp, opts.targetPath)
}

export function parseSHA256(text: string, assetName: string): string {
  for (const field of text.trim().split(/\s+/)) {
    if (/^[a-fA-F0-9]{64}$/.test(field)) {
      return field.toLowerCase()
    }
  }
  throw new Error(`error: checksum for ${assetName} did not contain a sha256 digest`)
}

export function gcOldVersions(root: string, keepVersion: string): string[] {
  if (!existsSync(root)) return []
  const removed: string[] = []
  for (const entry of readdirSync(root, { withFileTypes: true })) {
    if (!entry.isDirectory() || entry.name === keepVersion) continue
    const path = join(root, entry.name)
    rmSync(path, { recursive: true, force: true })
    removed.push(path)
  }
  return removed
}

export async function runCli(argv: string[], options: ManagerOptions = {}): Promise<number> {
  const cwd = options.cwd || process.cwd()
  const runtime = await prepareRuntime(options)
  const env = {
    ...process.env,
    ...(options.env || {}),
    C3X_VERSION: runtime.version,
    C3_SEMANTIC_CACHE_DIR: join(runtime.cacheDir, 'semantic'),
  }
  const exec = options.exec || defaultExec
  return exec(runtime.binaryPath, argv, env, cwd)
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
  async download(url: string): Promise<Uint8Array> {
    return this.fetch(url, 0)
  }

  private fetch(url: string, redirects: number): Promise<Uint8Array> {
    return new Promise((resolve, reject) => {
      const get = url.startsWith('http://') ? httpGet : httpsGet
      const req = get(url, (res) => {
        const status = res.statusCode ?? 0
        // GitHub release assets 302-redirect to a CDN — follow redirects.
        if (status >= 300 && status < 400 && res.headers.location) {
          res.resume()
          if (redirects >= 5) {
            reject(new Error(`download ${url}: too many redirects`))
            return
          }
          const next = new URL(res.headers.location, url).toString()
          this.fetch(next, redirects + 1).then(resolve, reject)
          return
        }
        if (status < 200 || status >= 300) {
          reject(new Error(`download ${url}: status ${status}`))
          res.resume()
          return
        }
        const chunks: Buffer[] = []
        res.on('data', (chunk) => chunks.push(Buffer.from(chunk)))
        res.on('end', () => resolve(Buffer.concat(chunks)))
      })
      req.on('error', (err) => reject(new Error(`download ${url}: ${err.message}\nhint: connect to GitHub Releases, or prefill the @c3x/cli cache`)))
      req.setTimeout(15 * 60 * 1000, () => {
        req.destroy(new Error(`download ${url}: timed out`))
      })
    })
  }
}
