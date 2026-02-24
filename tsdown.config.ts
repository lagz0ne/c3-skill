import { defineConfig } from 'tsdown'

export default defineConfig({
  entry: { cli: 'src/cli/index.ts' },
  format: 'cjs',
  platform: 'node',
  target: 'node18',
  clean: false,
  inlineOnly: false,
  outputOptions: {
    banner: '#!/usr/bin/env node',
  },
})
