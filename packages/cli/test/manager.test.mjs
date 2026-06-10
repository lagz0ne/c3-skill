import assert from 'node:assert/strict'
import { createHash } from 'node:crypto'
import { existsSync, mkdirSync, mkdtempSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { join } from 'node:path'
import { test } from 'node:test'
import {
  assetNames,
  cacheDirForVersion,
  ensureCachedAsset,
  gcOldVersions,
  prepareRuntime,
  resolvePlatform,
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

test('prepareRuntime downloads binary model and vocab with checksums then prunes old versions', async () => {
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
  assert.ok(existsSync(runtime.modelPath))
  assert.ok(existsSync(runtime.vocabPath))
  assert.equal(existsSync(join(root, 'c3x', '1.0.0')), false)
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

test('gcOldVersions keeps only current version directory', () => {
  const root = testTempDir()
  mkdirSync(join(root, '8.0.0'), { recursive: true })
  mkdirSync(join(root, '9.9.1'), { recursive: true })
  const removed = gcOldVersions(root, '9.9.1')
  assert.deepEqual(removed, [join(root, '8.0.0')])
  assert.equal(existsSync(join(root, '8.0.0')), false)
  assert.equal(existsSync(join(root, '9.9.1')), true)
})

function stubDownloader(assets) {
  return {
    async download(url) {
      const name = url.split('/').pop()
      if (name.endsWith('.sha256')) {
        const assetName = name.slice(0, -'.sha256'.length)
        const data = assets.get(assetName)
        if (!data) throw new Error(`missing ${assetName}`)
        return Buffer.from(`${sha256(data)}  ${assetName}\n`)
      }
      const data = assets.get(name)
      if (!data) throw new Error(`missing ${name}`)
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
