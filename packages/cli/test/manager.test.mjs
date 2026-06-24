import assert from 'node:assert/strict'
import { spawnSync } from 'node:child_process'
import { createHash } from 'node:crypto'
import { existsSync, mkdirSync, mkdtempSync, readFileSync, writeFileSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { join } from 'node:path'
import { test } from 'node:test'
import { fileURLToPath } from 'node:url'
import {
  assetNames,
  cacheDirForVersion,
  createProgressReporter,
  ensureCachedAsset,
  fetchAvailableVersions,
  gcOldVersions,
  installRuntime,
  listInstalledRuntimes,
  pruneRuntimeCache,
  prepareRuntime,
  readProjectRuntimeConfig,
  resolvePlatform,
  runCli,
  runManagerCommand,
  uninstallRuntime,
  writeProjectRuntimeConfig,
} from '../dist/manager.mjs'

test('resolvePlatform maps supported Node platform names', () => {
  assert.deepEqual(resolvePlatform('linux', 'x64'), { os: 'linux', arch: 'amd64' })
  assert.deepEqual(resolvePlatform('linux', 'arm64'), { os: 'linux', arch: 'arm64' })
  assert.deepEqual(resolvePlatform('darwin', 'arm64'), { os: 'darwin', arch: 'arm64' })
  assert.throws(() => resolvePlatform('darwin', 'x64'), /unsupported platform/)
  assert.throws(() => resolvePlatform('win32', 'x64'), /unsupported platform/)
})

test('cacheDirForVersion respects XDG_CACHE_HOME', () => {
  assert.equal(
    cacheDirForVersion('9.9.1', { XDG_CACHE_HOME: '/tmp/cache', HOME: '/home/test' }),
    '/tmp/cache/c3x/9.9.1',
  )
})

test('prepareRuntime downloads binary model and vocab with checksums without pruning old versions', async () => {
  const root = testTempDir()
  mkdirSync(join(root, 'c3x', '1.0.0'), { recursive: true })
  const version = '9.9.1'
  const names = assetNames(version, { os: 'linux', arch: 'amd64' })
  const assets = new Map([
    [names.binary, Buffer.from('binary')],
    [names.model, Buffer.from('model')],
    [names.vocab, Buffer.from('vocab')],
  ])
  const downloads = stubDownloader(assets)

  const runtime = await prepareRuntime({
    env: {
      XDG_CACHE_HOME: root,
      HOME: root,
      C3X_VERSION: version,
      C3X_RELEASE_BASE_URL: 'https://example.invalid/release',
    },
    platform: 'linux',
    arch: 'x64',
    downloader: downloads,
  })

  assert.equal(runtime.cacheDir, join(root, 'c3x', version))
  assert.ok(existsSync(runtime.binaryPath))
  assert.equal(runtime.astGrepPath, undefined)
  assert.ok(existsSync(runtime.modelPath))
  assert.ok(existsSync(runtime.vocabPath))
  assert.equal(existsSync(join(root, 'c3x', '1.0.0')), true)
})

test('fetchAvailableVersions lists semver releases newest first', async () => {
  const downloader = {
    async download(url) {
      assert.equal(url, 'https://example.invalid/releases')
      return Buffer.from(JSON.stringify([
        { tag_name: 'v10.1.0', draft: false, prerelease: false },
        { tag_name: 'v11.0.0', draft: false, prerelease: false },
        { tag_name: 'not-a-version', draft: false, prerelease: false },
        { tag_name: 'v12.0.0-beta.1', draft: false, prerelease: true },
        { tag_name: 'v9.9.9', draft: true, prerelease: false },
      ]))
    },
  }

  const versions = await fetchAvailableVersions({
    env: { C3X_RELEASES_URL: 'https://example.invalid/releases' },
    downloader,
  })

  assert.deepEqual(versions, ['11.0.0', '10.1.0'])
})

test('installRuntime keeps other installed versions and records manifest with progress', async () => {
  const root = testTempDir()
  mkdirSync(join(root, 'c3x', '1.0.0'), { recursive: true })
  const version = '11.5.0'
  const names = assetNames(version, { os: 'linux', arch: 'amd64' })
  const assets = new Map([
    [names.binary, Buffer.from('binary')],
    [names.astGrep, Buffer.from('ast-grep')],
    [names.model, Buffer.from('model')],
    [names.vocab, Buffer.from('vocab')],
  ])
  const progress = []

  const runtime = await installRuntime({
    version,
    env: {
      XDG_CACHE_HOME: root,
      HOME: root,
      C3X_RELEASE_BASE_URL: 'https://example.invalid/release',
    },
    platform: 'linux',
    arch: 'x64',
    downloader: stubDownloader(assets),
    progress: (event) => progress.push(event),
  })

  assert.ok(existsSync(join(root, 'c3x', '1.0.0')))
  assert.ok(existsSync(runtime.binaryPath))
  const manifest = JSON.parse(readFileSync(join(runtime.cacheDir, 'manifest.json'), 'utf8'))
  assert.equal(manifest.version, version)
  assert.equal(manifest.assets.binary.sha256, sha256(Buffer.from('binary')))
  assert.equal(manifest.assets.astGrep.sha256, sha256(Buffer.from('ast-grep')))
  assert.deepEqual(progress.map((event) => event.kind), [
    'asset-start',
    'asset-progress',
    'asset-downloaded',
    'asset-ready',
    'asset-start',
    'asset-progress',
    'asset-downloaded',
    'asset-ready',
    'asset-start',
    'asset-progress',
    'asset-downloaded',
    'asset-ready',
    'asset-start',
    'asset-progress',
    'asset-downloaded',
    'asset-ready',
    'runtime-ready',
  ])
})

test('runCli prints root help without resolving or downloading a runtime', async () => {
  const stdout = []
  let downloaded = false
  let executed = false

  const code = await runCli(['--help'], {
    downloader: {
      async download() {
        downloaded = true
        throw new Error('should not download')
      },
    },
    stdout: (line) => stdout.push(line),
    stderr: () => {},
    exec() {
      executed = true
      return 0
    },
  })

  assert.equal(code, 0)
  assert.equal(downloaded, false)
  assert.equal(executed, false)
  assert.match(stdout.join('\n'), /Usage:/)
  assert.match(stdout.join('\n'), /c3x runtime install/)
})

test('runCli prints root help for empty argv without resolving or downloading a runtime', async () => {
  const stdout = []
  let downloaded = false

  const code = await runCli([], {
    downloader: {
      async download() {
        downloaded = true
        throw new Error('should not download')
      },
    },
    stdout: (line) => stdout.push(line),
    stderr: () => {},
  })

  assert.equal(code, 0)
  assert.equal(downloaded, false)
  assert.match(stdout.join('\n'), /Normal C3 commands are forwarded/)
})

test('runCli prints package version without resolving or downloading a runtime', async () => {
  const stdout = []
  let downloaded = false

  const code = await runCli(['--version'], {
    downloader: {
      async download() {
        downloaded = true
        throw new Error('should not download')
      },
    },
    stdout: (line) => stdout.push(line),
    stderr: () => {},
  })

  assert.equal(code, 0)
  assert.equal(downloaded, false)
  assert.match(stdout[0], /^\d+\.\d+\.\d+/)
})

test('runCli uses project runtime metadata before latest release', async () => {
  const root = testTempDir()
  const project = join(root, 'project')
  mkdirSync(join(project, '.c3'), { recursive: true })
  writeFileSync(join(project, '.c3', 'runtime.json'), JSON.stringify({ version: '8.8.8' }))
  const names = assetNames('8.8.8', { os: 'linux', arch: 'amd64' })
  const assets = new Map([
    [names.binary, Buffer.from('binary')],
    [names.model, Buffer.from('model')],
    [names.vocab, Buffer.from('vocab')],
  ])
  let executed

  await runCli(['list'], {
    cwd: project,
    env: {
      XDG_CACHE_HOME: root,
      HOME: root,
      C3X_RELEASE_BASE_URL: 'https://example.invalid/release',
      C3X_RELEASES_URL: 'https://example.invalid/releases',
    },
    platform: 'linux',
    arch: 'x64',
    downloader: stubDownloader(assets),
    stderr: () => {},
    exec(file, args, env, cwd) {
      executed = { file, args, env, cwd }
      return 0
    },
  })

  assert.equal(executed.cwd, project)
  assert.equal(executed.env.C3X_VERSION, '8.8.8')
  assert.equal(executed.env.C3_AST_GREP, undefined)
  assert.deepEqual(executed.args, ['list'])
  assert.match(executed.file, /8\.8\.8/)
})

test('runCli defaults to latest release when no project runtime is set', async () => {
  const root = testTempDir()
  const names = assetNames('11.5.0', { os: 'linux', arch: 'amd64' })
  const assets = new Map([
    [names.binary, Buffer.from('binary')],
    [names.astGrep, Buffer.from('ast-grep')],
    [names.model, Buffer.from('model')],
    [names.vocab, Buffer.from('vocab')],
  ])
  const downloader = stubDownloader(assets, {
    releases: [
      { tag_name: 'v10.0.0', draft: false, prerelease: false },
      { tag_name: 'v11.5.0', draft: false, prerelease: false },
    ],
  })
  let executed

  await runCli(['list'], {
    cwd: root,
    env: {
      XDG_CACHE_HOME: root,
      HOME: root,
      C3X_RELEASE_BASE_URL: 'https://example.invalid/release',
      C3X_RELEASES_URL: 'https://example.invalid/releases',
    },
    platform: 'linux',
    arch: 'x64',
    downloader,
    stderr: () => {},
    exec(file, args, env) {
      executed = { file, args, env }
      return 0
    },
  })

  assert.equal(executed.env.C3X_VERSION, '11.5.0')
  assert.match(executed.env.C3_AST_GREP, /ast-grep-0\.44\.0-linux-amd64/)
  assert.match(executed.file, /11\.5\.0/)
})

test('runCli reports automatic download progress for real commands', async () => {
  const root = testTempDir()
  const names = assetNames('11.5.0', { os: 'linux', arch: 'amd64' })
  const assets = new Map([
    [names.binary, Buffer.from('binary')],
    [names.astGrep, Buffer.from('ast-grep')],
    [names.model, Buffer.from('model')],
    [names.vocab, Buffer.from('vocab')],
  ])
  const stderr = []

  await runCli(['list'], {
    cwd: root,
    env: {
      XDG_CACHE_HOME: root,
      HOME: root,
      C3X_RELEASE_BASE_URL: 'https://example.invalid/release',
      C3X_RELEASES_URL: 'https://example.invalid/releases',
    },
    platform: 'linux',
    arch: 'x64',
    downloader: stubDownloader(assets, {
      releases: [{ tag_name: 'v11.5.0', draft: false, prerelease: false }],
    }),
    stderr: (line) => stderr.push(line),
    exec() {
      return 0
    },
  })

  assert.match(stderr.join('\n'), /downloading required runtime assets/)
  assert.match(stderr.join('\n'), /\[####################\] 100%/)
  assert.match(stderr.join('\n'), /runtime ready/)
})

test('createProgressReporter throttles progress output', () => {
  const lines = []
  const progress = createProgressReporter((line) => lines.push(line))

  progress({ kind: 'asset-start', version: '1.2.3', assetName: 'c3x-test' })
  progress({ kind: 'asset-progress', version: '1.2.3', assetName: 'c3x-test', downloadedBytes: 1, totalBytes: 100 })
  progress({ kind: 'asset-progress', version: '1.2.3', assetName: 'c3x-test', downloadedBytes: 2, totalBytes: 100 })
  progress({ kind: 'asset-progress', version: '1.2.3', assetName: 'c3x-test', downloadedBytes: 11, totalBytes: 100 })
  progress({ kind: 'asset-progress', version: '1.2.3', assetName: 'c3x-test', downloadedBytes: 100, totalBytes: 100 })

  assert.equal(lines.filter((line) => line.includes('c3x-test [')).length, 3)
})

test('project runtime metadata rejects paths and URLs', () => {
  const project = testTempDir()
  mkdirSync(join(project, '.c3'), { recursive: true })
  writeFileSync(join(project, '.c3', 'runtime.json'), JSON.stringify({ version: 'https://evil.invalid/c3x' }))
  assert.throws(() => readProjectRuntimeConfig(project), /invalid project runtime version/)

  writeFileSync(join(project, '.c3', 'runtime.json'), JSON.stringify({ version: '../bin/c3x' }))
  assert.throws(() => readProjectRuntimeConfig(project), /invalid project runtime version/)

  writeFileSync(join(project, '.c3', 'runtime.json'), JSON.stringify({ version: '9.9.1', updatedAt: '2026-06-23T00:00:00.000Z' }))
  assert.throws(() => readProjectRuntimeConfig(project), /invalid project runtime metadata field updatedAt/)
})

test('writeProjectRuntimeConfig stores only the operational project version', () => {
  const project = testTempDir()
  mkdirSync(join(project, '.c3'), { recursive: true })

  const config = writeProjectRuntimeConfig(project, '9.9.1')

  assert.equal(config.version, '9.9.1')
  assert.deepEqual(Object.keys(config), ['version'])
  assert.equal(readProjectRuntimeConfig(project).version, '9.9.1')
})

test('runManagerCommand lists available versions with installed marker', async () => {
  const root = testTempDir()
  mkdirSync(join(root, 'c3x', '10.0.0'), { recursive: true })
  const stdout = []

  await runManagerCommand(['versions'], {
    env: { XDG_CACHE_HOME: root, HOME: root, C3X_RELEASES_URL: 'https://example.invalid/releases' },
    downloader: stubDownloader(new Map(), {
      releases: [
        { tag_name: 'v10.0.0', draft: false, prerelease: false },
        { tag_name: 'v11.0.0', draft: false, prerelease: false },
      ],
    }),
    stdout: (line) => stdout.push(line),
  })

  assert.deepEqual(stdout, ['11.0.0', '10.0.0 installed'])
})

test('runManagerCommand install reports progress on stderr', async () => {
  const root = testTempDir()
  const version = '11.5.0'
  const names = assetNames(version, { os: 'linux', arch: 'amd64' })
  const assets = new Map([
    [names.binary, Buffer.from('binary')],
    [names.astGrep, Buffer.from('ast-grep')],
    [names.model, Buffer.from('model')],
    [names.vocab, Buffer.from('vocab')],
  ])
  const stdout = []
  const stderr = []

  await runManagerCommand(['install', version], {
    env: {
      XDG_CACHE_HOME: root,
      HOME: root,
      C3X_RELEASE_BASE_URL: 'https://example.invalid/release',
    },
    platform: 'linux',
    arch: 'x64',
    downloader: stubDownloader(assets),
    stdout: (line) => stdout.push(line),
    stderr: (line) => stderr.push(line),
  })

  assert.deepEqual(stdout, [`installed ${version}`])
  assert.match(stderr.join('\n'), /preparing c3x-11\.5\.0-linux-amd64/)
  assert.match(stderr.join('\n'), /preparing ast-grep-0\.44\.0-linux-amd64/)
  assert.match(stderr.join('\n'), /\[####################\] 100%/)
  assert.match(stderr.join('\n'), /runtime ready/)
})

test('runManagerCommand writes project runtime metadata', async () => {
  const project = testTempDir()
  mkdirSync(join(project, '.c3'), { recursive: true })
  const stdout = []

  await runManagerCommand(['use', '9.9.1'], { cwd: project, stdout: (line) => stdout.push(line) })

  assert.deepEqual(stdout, ['project runtime 9.9.1'])
  assert.deepEqual(JSON.parse(readFileSync(join(project, '.c3', 'runtime.json'), 'utf8')), { version: '9.9.1' })
})

test('cli runtime namespace is handled by the npm manager', () => {
  const root = testTempDir()
  const project = join(root, 'project')
  mkdirSync(join(project, '.c3'), { recursive: true })
  const packageDir = fileURLToPath(new URL('..', import.meta.url))

  const result = spawnSync(process.execPath, [join(packageDir, 'dist', 'cli.mjs'), 'runtime', 'use', '9.9.1'], {
    cwd: project,
    env: { ...process.env, XDG_CACHE_HOME: root, HOME: root },
    encoding: 'utf8',
  })

  assert.equal(result.status, 0, result.stderr)
  assert.deepEqual(JSON.parse(readFileSync(join(project, '.c3', 'runtime.json'), 'utf8')), { version: '9.9.1' })
})

test('cli root discovery commands do not resolve or download runtimes', () => {
  const root = testTempDir()
  const packageDir = fileURLToPath(new URL('..', import.meta.url))
  const env = {
    ...process.env,
    XDG_CACHE_HOME: root,
    HOME: root,
    C3X_RELEASES_URL: 'http://example.invalid/releases',
    C3X_RELEASE_BASE_URL: 'http://example.invalid/release',
  }

  for (const args of [[], ['--help'], ['--version']]) {
    const result = spawnSync(process.execPath, [join(packageDir, 'dist', 'cli.mjs'), ...args], {
      cwd: root,
      env,
      encoding: 'utf8',
    })

    assert.equal(result.status, 0, `${args.join(' ') || '<empty>'}\n${result.stderr}`)
  }
  assert.equal(existsSync(join(root, 'c3x')), false)
})

test('cli root discovery commands pass in bwrap cache and network isolation', { skip: !canRunBwrapNet() && 'bwrap network isolation unavailable' }, () => {
  const root = testTempDir()

  for (const args of [[], ['--help'], ['--version']]) {
    const result = runBwrapCli(root, args)
    assert.equal(result.status, 0, `${args.join(' ') || '<empty>'}\n${result.stderr}`)
    assert.equal(result.stderr, '')
    assert.equal(existsSync(join(root, 'cache', 'c3x')), false)
  }

  const realCommand = runBwrapCli(root, ['list'])
  assert.notEqual(realCommand.status, 0)
  assert.equal(existsSync(join(root, 'cache', 'c3x')), false)
})

test('uninstallRuntime refuses the project selected runtime unless forced', async () => {
  const root = testTempDir()
  const project = join(root, 'project')
  mkdirSync(join(root, 'c3x', '9.9.1'), { recursive: true })
  mkdirSync(join(project, '.c3'), { recursive: true })
  writeFileSync(join(project, '.c3', 'runtime.json'), JSON.stringify({ version: '9.9.1' }))

  assert.throws(
    () => uninstallRuntime('9.9.1', { env: { XDG_CACHE_HOME: root, HOME: root }, cwd: project }),
    /project selected runtime/,
  )
  assert.ok(existsSync(join(root, 'c3x', '9.9.1')))

  uninstallRuntime('9.9.1', { env: { XDG_CACHE_HOME: root, HOME: root }, cwd: project, force: true })
  assert.equal(existsSync(join(root, 'c3x', '9.9.1')), false)
})

test('pruneRuntimeCache removes unselected versions and keeps the project runtime', () => {
  const root = testTempDir()
  const project = join(root, 'project')
  mkdirSync(join(root, 'c3x', '1.0.0'), { recursive: true })
  mkdirSync(join(root, 'c3x', '2.0.0'), { recursive: true })
  mkdirSync(join(project, '.c3'), { recursive: true })
  writeFileSync(join(project, '.c3', 'runtime.json'), JSON.stringify({ version: '2.0.0' }))

  const removed = pruneRuntimeCache({ env: { XDG_CACHE_HOME: root, HOME: root }, cwd: project })

  assert.deepEqual(removed, [join(root, 'c3x', '1.0.0')])
  assert.equal(existsSync(join(root, 'c3x', '1.0.0')), false)
  assert.equal(existsSync(join(root, 'c3x', '2.0.0')), true)
})

test('pruneRuntimeCache keeps newest installed runtime when project selection is missing', () => {
  const root = testTempDir()
  const project = join(root, 'project')
  mkdirSync(join(root, 'c3x', '1.0.0'), { recursive: true })
  mkdirSync(join(root, 'c3x', '2.0.0'), { recursive: true })
  mkdirSync(join(project, '.c3'), { recursive: true })
  writeFileSync(join(project, '.c3', 'runtime.json'), JSON.stringify({ version: '9.9.1' }))

  const removed = pruneRuntimeCache({ env: { XDG_CACHE_HOME: root, HOME: root }, cwd: project })

  assert.deepEqual(removed, [join(root, 'c3x', '1.0.0')])
  assert.equal(existsSync(join(root, 'c3x', '1.0.0')), false)
  assert.equal(existsSync(join(root, 'c3x', '2.0.0')), true)
})

test('ensureCachedAsset rejects checksum mismatch', async () => {
  const dir = testTempDir()
  const downloader = {
    async download(url) {
      if (url.endsWith('.sha256')) {
        return Buffer.from(`${'0'.repeat(64)}  c3x-test\n`)
      }
      return Buffer.from('not matching')
    },
  }
  await assert.rejects(
    ensureCachedAsset({
      assetName: 'c3x-test',
      targetPath: join(dir, 'c3x-test'),
      baseURL: 'https://example.invalid/release',
      downloader,
    }),
    /checksum mismatch/,
  )
})

test('fetchAvailableVersions rejects non-https release indexes', async () => {
  await assert.rejects(
    fetchAvailableVersions({
      env: { C3X_RELEASES_URL: 'http://example.invalid/releases' },
      downloader: { async download() { return Buffer.from('[]') } },
    }),
    /only https URLs are supported/,
  )
})

test('gcOldVersions keeps only current version directory', () => {
  const root = testTempDir()
  mkdirSync(join(root, '8.0.0'), { recursive: true })
  mkdirSync(join(root, '9.9.1'), { recursive: true })
  const removed = gcOldVersions(root, '9.9.1')
  assert.deepEqual(removed, [join(root, '8.0.0')])
  assert.equal(existsSync(join(root, '8.0.0')), false)
  assert.equal(existsSync(join(root, '9.9.1')), true)
})

function stubDownloader(assets, options = {}) {
  return {
    async download(url, progress) {
      if (url === 'https://example.invalid/releases') {
        return Buffer.from(JSON.stringify(options.releases || []))
      }
      const name = url.split('/').pop()
      if (name.endsWith('.sha256')) {
        const assetName = name.slice(0, -'.sha256'.length)
        const data = assets.get(assetName)
        if (!data) throw new Error(`missing ${assetName}`)
        return Buffer.from(`${sha256(data)}  ${assetName}\n`)
      }
      const data = assets.get(name)
      if (!data) throw new Error(`missing ${name}`)
      progress?.(data.length, data.length)
      return data
    },
  }
}

function sha256(data) {
  return createHash('sha256').update(data).digest('hex')
}

function testTempDir() {
  return mkdtempSync(join(tmpdir(), 'c3x-cli-test-'))
}

function canRunBwrapNet() {
  const result = spawnSync('bwrap', ['--unshare-net', '--ro-bind', '/', '/', 'true'], { encoding: 'utf8' })
  return result.status === 0
}

function runBwrapCli(root, args) {
  const packageDir = fileURLToPath(new URL('..', import.meta.url))
  return spawnSync('bwrap', [
    '--clearenv',
    '--unshare-net',
    '--ro-bind', '/', '/',
    '--proc', '/proc',
    '--dev', '/dev',
    '--tmpfs', '/tmp',
    '--bind', root, '/tmp/work',
    '--dir', '/tmp/work/home',
    '--dir', '/tmp/work/cache',
    '--setenv', 'HOME', '/tmp/work/home',
    '--setenv', 'XDG_CACHE_HOME', '/tmp/work/cache',
    '--setenv', 'C3X_RELEASES_URL', 'https://api.github.com/repos/lagz0ne/c3-skill/releases?per_page=1',
    process.execPath,
    join(packageDir, 'dist', 'cli.mjs'),
    ...args,
  ], {
    cwd: packageDir,
    encoding: 'utf8',
  })
}
