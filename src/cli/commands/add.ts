import * as fs from "node:fs";
import * as path from "node:path";
import type { CliOptions } from "../context";
import { graphAtom } from "../context";
import type { Lite } from "@pumped-fn/lite";
import { showHelp } from "../help";
import { nextContainerId, nextComponentId, nextAdrId } from "../../core/numbering";
import { addComponentToContainerTable } from "../../core/wiring";

export async function addCommand(options: CliOptions, c3Dir: string, scope: Lite.Scope): Promise<void> {
  if (options.help) {
    showHelp("add");
    return;
  }

  const [entityType, slug] = options.args;
  if (!entityType || !slug) {
    console.error("error: usage: c3 add <type> <slug>");
    console.error("hint: types: container, component, ref, adr");
    process.exitCode = 1;
    return;
  }

  const templatesDir = new URL("../../../templates/", import.meta.url);

  switch (entityType) {
    case "container":
      await addContainer(slug, c3Dir, scope, templatesDir);
      break;
    case "component":
      await addComponent(slug, c3Dir, scope, options, templatesDir);
      break;
    case "ref":
      await addRef(slug, c3Dir, templatesDir);
      break;
    case "adr":
      addAdr(slug, c3Dir, templatesDir);
      break;
    default:
      console.error(`error: unknown entity type '${entityType}'`);
      console.error("hint: types: container, component, ref, adr");
      process.exitCode = 1;
  }
}

async function addContainer(
  slug: string,
  c3Dir: string,
  scope: Lite.Scope,
  templatesDir: URL,
): Promise<void> {
  const graph = await scope.resolve(graphAtom);
  const n = nextContainerId(graph);
  const dirName = `c3-${n}-${slug}`;
  const dirPath = path.join(c3Dir, dirName);

  fs.mkdirSync(dirPath, { recursive: true });

  const template = fs.readFileSync(new URL("container.md", templatesDir), "utf-8");
  const content = template
    .replace(/\$\{N\}/g, String(n))
    .replace(/\$\{CONTAINER_NAME\}/g, slug)
    .replace(/\$\{BOUNDARY\}/g, "service")
    .replace(/\$\{GOAL\}/g, "")
    .replace(/\$\{SUMMARY\}/g, "");

  fs.writeFileSync(path.join(dirPath, "README.md"), content, "utf-8");
  console.log(`Created: ${path.relative(path.dirname(c3Dir), path.join(dirPath, "README.md"))} (id: c3-${n})`);
}

async function addComponent(
  slug: string,
  c3Dir: string,
  scope: Lite.Scope,
  options: CliOptions,
  templatesDir: URL,
): Promise<void> {
  const containerArg = options.container;
  if (!containerArg) {
    console.error("error: --container <id> is required for component");
    console.error("hint: c3 add component auth-provider --container c3-1");
    process.exitCode = 1;
    return;
  }

  const graph = await scope.resolve(graphAtom);

  // Parse container number from id like "c3-1" or "c3-3"
  const containerMatch = containerArg.match(/^c3-(\d+)$/);
  if (!containerMatch) {
    console.error(`error: invalid container id '${containerArg}'`);
    console.error("hint: use format c3-N, e.g. c3-1, c3-3");
    process.exitCode = 1;
    return;
  }
  const containerNum = parseInt(containerMatch[1], 10);

  // Find the container entity and its directory
  const containerEntity = graph.entities.get(containerArg);
  if (!containerEntity) {
    console.error(`error: container '${containerArg}' not found`);
    process.exitCode = 1;
    return;
  }

  const componentId = nextComponentId(graph, containerNum, !!options.feature);
  const category = options.feature ? "feature" : "foundation";

  // Component number part (e.g., "01" from "c3-101")
  const nn = componentId.replace(`c3-${containerNum}`, "");
  const fileName = `${componentId}-${slug}.md`;
  const containerDir = path.join(c3Dir, path.dirname(containerEntity.path));
  const filePath = path.join(containerDir, fileName);

  const template = fs.readFileSync(new URL("component.md", templatesDir), "utf-8");
  const content = template
    .replace(/\$\{N\}\$\{NN\}/g, componentId.replace("c3-", ""))
    .replace(/\$\{N\}/g, String(containerNum))
    .replace(/\$\{NN\}/g, nn)
    .replace(/\$\{COMPONENT_NAME\}/g, slug)
    .replace(/\$\{CATEGORY\}/g, category)
    .replace(/\$\{GOAL\}/g, "")
    .replace(/\$\{SUMMARY\}/g, "");

  fs.writeFileSync(filePath, content, "utf-8");

  const relPath = path.relative(path.dirname(c3Dir), filePath);
  console.log(`Created: ${relPath} (id: ${componentId})`);

  // Update container table
  const containerReadme = path.join(containerDir, "README.md");
  if (fs.existsSync(containerReadme)) {
    addComponentToContainerTable(containerReadme, componentId, slug, category, "");
    console.log(`Updated: ${path.relative(path.dirname(c3Dir), containerReadme)} (component list)`);
  }

}

async function addRef(
  slug: string,
  c3Dir: string,
  templatesDir: URL,
): Promise<void> {
  const refsDir = path.join(c3Dir, "refs");
  if (!fs.existsSync(refsDir)) {
    fs.mkdirSync(refsDir, { recursive: true });
  }

  const fileName = `ref-${slug}.md`;
  const filePath = path.join(refsDir, fileName);

  if (fs.existsSync(filePath)) {
    console.error(`error: ref-${slug} already exists`);
    process.exitCode = 1;
    return;
  }

  const template = fs.readFileSync(new URL("ref.md", templatesDir), "utf-8");
  const content = template
    .replace(/\$\{SLUG\}/g, slug)
    .replace(/\$\{TITLE\}/g, slug)
    .replace(/\$\{GOAL\}/g, "");

  fs.writeFileSync(filePath, content, "utf-8");
  console.log(`Created: ${path.relative(path.dirname(c3Dir), filePath)} (id: ref-${slug})`);
}

function addAdr(
  slug: string,
  c3Dir: string,
  templatesDir: URL,
): void {
  const adrDir = path.join(c3Dir, "adr");
  if (!fs.existsSync(adrDir)) {
    fs.mkdirSync(adrDir, { recursive: true });
  }

  const adrId = nextAdrId(slug);
  const fileName = `${adrId}.md`;
  const filePath = path.join(adrDir, fileName);

  if (fs.existsSync(filePath)) {
    console.error(`error: ${adrId} already exists`);
    process.exitCode = 1;
    return;
  }

  const template = fs.readFileSync(new URL("adr-000.md", templatesDir), "utf-8");
  const today = new Date().toISOString().slice(0, 10).replace(/-/g, "");
  const content = template
    .replace(/adr-00000000-c3-adoption/g, adrId)
    .replace(/C3 Architecture Documentation Adoption/g, slug)
    .replace(/Adopt C3 methodology for \$\{PROJECT\}\./g, "")
    .replace(/\$\{DATE\}/g, today)
    .replace(/\$\{PROJECT\}/g, "");

  fs.writeFileSync(filePath, content, "utf-8");
  console.log(`Created: ${path.relative(path.dirname(c3Dir), filePath)} (id: ${adrId})`);
}
