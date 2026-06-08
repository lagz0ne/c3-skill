import { defineConfig } from 'tsdown'

export default defineConfig({
  entry: ['src/cli.ts', 'src/manager.ts', 'src/version.ts'],
  format: 'esm',
  platform: 'node',
  target: 'node18',
  outDir: 'dist',
  clean: true,
  banner: { js: '#!/usr/bin/env node' },
})
