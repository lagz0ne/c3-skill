// src/cli/index.ts
import { createScope, type Lite } from "@pumped-fn/lite";
import { findC3Dir } from "../core/config";
import { c3DirTag, optionsTag, type CliOptions } from "./context";
import { showHelp } from "./help";

// Command imports
import { listCommand } from "./commands/list";
import { checkCommand } from "./commands/check";
import { initCommand } from "./commands/init";
import { addCommand } from "./commands/add";
import { syncCommand } from "./commands/sync";

const VERSION = "5.0.0";

type CommandHandler = (options: CliOptions, c3Dir: string, scope: Lite.Scope) => Promise<void>;

export class C3Error extends Error {
  constructor(message: string, public hint?: string) {
    super(message);
    this.name = "C3Error";
  }
}

function parseArgs(argv: string[]): CliOptions {
  const args: string[] = [];
  let json = false, flat = false, help = false, version = false;
  let container: string | undefined, c3Dir: string | undefined;
  let feature = false;

  for (let i = 0; i < argv.length; i++) {
    const arg = argv[i];
    if (arg === "--json") json = true;
    else if (arg === "--flat") flat = true;
    else if (arg === "--feature") feature = true;
    else if (arg === "-h" || arg === "--help") help = true;
    else if (arg === "-v" || arg === "--version") version = true;
    else if (arg === "--container" && argv[i + 1]) container = argv[++i];
    else if (arg === "--c3-dir" && argv[i + 1]) c3Dir = argv[++i];
    else args.push(arg);
  }

  return {
    command: args[0] || "",
    args: args.slice(1),
    json, flat, container, c3Dir,
    feature, help, version,
  };
}

const commandMap: Record<string, CommandHandler> = {
  list: listCommand,
  check: checkCommand,
  add: addCommand,
  sync: syncCommand,
};

async function main() {
  const options = parseArgs(process.argv.slice(2));

  if (options.version) {
    console.log(VERSION);
    return;
  }

  if (options.help || !options.command) {
    showHelp(options.command || undefined);
    return;
  }

  // init is special — creates .c3/
  if (options.command === "init") {
    await initCommand(options, process.cwd(), undefined as unknown as Lite.Scope);
    return;
  }

  const handler = commandMap[options.command];
  if (!handler) {
    console.error(`error: unknown command '${options.command}'`);
    console.error(`hint: run 'c3x --help' to see available commands`);
    process.exitCode = 1;
    return;
  }

  const c3Dir = options.c3Dir ?? findC3Dir(process.cwd());
  if (!c3Dir) {
    console.error("error: No .c3/ directory found");
    console.error("hint: run 'c3x init' to create one, or use --c3-dir <path>");
    process.exitCode = 1;
    return;
  }

  const scope = createScope({
    tags: [c3DirTag(c3Dir), optionsTag(options)],
  });

  try {
    await handler(options, c3Dir, scope);
  } catch (err) {
    if (err instanceof C3Error) {
      console.error(`error: ${err.message}`);
      if (err.hint) console.error(`hint: ${err.hint}`);
      process.exitCode = 1;
    } else {
      console.error(`unexpected error: ${err}`);
      console.error(`Report: https://github.com/lagz0ne/c3-skill/issues`);
      process.exitCode = 2;
    }
  } finally {
    await scope.dispose();
  }
}

main();
