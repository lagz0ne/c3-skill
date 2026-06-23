import { runCli, runManagerCommand } from './manager.js'

const argv = process.argv.slice(2)
const run = argv[0] === 'runtime' ? runManagerCommand(argv.slice(1)) : runCli(argv)

run.then((code) => {
  if (typeof code === 'number') process.exitCode = code
}).catch((err: unknown) => {
  const message = err instanceof Error ? err.message : String(err)
  console.error(message)
  process.exit(1)
})
