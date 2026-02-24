import type { CliOptions } from "../context";
import { graphAtom } from "../context";
import type { Lite } from "@pumped-fn/lite";
import { showHelp } from "../help";
import { renderTopology, renderFlatList, renderJson } from "../output";

export async function listCommand(options: CliOptions, c3Dir: string, scope: Lite.Scope): Promise<void> {
  if (options.help) {
    showHelp("list");
    return;
  }

  const graph = await scope.resolve(graphAtom);

  if (options.json) {
    const data = [...graph.entities.values()].map(e => ({
      id: e.id,
      type: e.type,
      title: e.title,
      path: e.path,
      relationships: e.relationships,
      frontmatter: e.frontmatter,
    }));
    console.log(renderJson(data));
    return;
  }

  if (options.flat) {
    console.log(renderFlatList(graph));
    return;
  }

  console.log(renderTopology(graph));
}
