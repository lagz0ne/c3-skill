import * as fs from "node:fs";
import * as path from "node:path";
import type { CliOptions } from "../context";
import type { Lite } from "@pumped-fn/lite";
import { showHelp } from "../help";

export async function initCommand(options: CliOptions, c3Dir: string, scope: Lite.Scope): Promise<void> {
  if (options.help) {
    showHelp("init");
    return;
  }

  // c3Dir here is actually the project root (process.cwd()), not the .c3/ dir
  const projectRoot = c3Dir;
  const dotC3 = path.join(projectRoot, ".c3");

  if (fs.existsSync(dotC3)) {
    console.error("error: .c3/ directory already exists");
    process.exitCode = 1;
    return;
  }

  const today = new Date().toISOString().slice(0, 10).replace(/-/g, "");

  // Create directory structure
  fs.mkdirSync(dotC3, { recursive: true });
  fs.mkdirSync(path.join(dotC3, "refs"), { recursive: true });
  fs.mkdirSync(path.join(dotC3, "adr"), { recursive: true });

  // config.yaml
  fs.writeFileSync(
    path.join(dotC3, "config.yaml"),
    "# C3 configuration\n",
    "utf-8"
  );

  // README.md from context template
  const templatesDir = new URL("../../../templates/", import.meta.url);
  const contextTemplate = fs.readFileSync(
    new URL("context.md", templatesDir),
    "utf-8"
  );
  fs.writeFileSync(path.join(dotC3, "README.md"), contextTemplate, "utf-8");

  // adr-00000000-c3-adoption.md from adr template
  const adrTemplate = fs.readFileSync(
    new URL("adr-000.md", templatesDir),
    "utf-8"
  );
  const adrContent = adrTemplate
    .replace(/\$\{DATE\}/g, today)
    .replace(/\$\{PROJECT\}/g, path.basename(projectRoot));
  fs.writeFileSync(
    path.join(dotC3, "adr", "adr-00000000-c3-adoption.md"),
    adrContent,
    "utf-8"
  );

  console.log("Created .c3/");
  console.log("  \u251C\u2500\u2500 config.yaml");
  console.log("  \u251C\u2500\u2500 README.md");
  console.log("  \u251C\u2500\u2500 refs/");
  console.log("  \u2514\u2500\u2500 adr/");
  console.log("      \u2514\u2500\u2500 adr-00000000-c3-adoption.md");
}
