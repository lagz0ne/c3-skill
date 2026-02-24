// src/cli/context.ts
import { atom, tag, tags, type Lite } from "@pumped-fn/lite";
import { walkC3Docs, buildRelationshipGraph } from "../core/walker";

// Tags — ambient context per CLI invocation
export const c3DirTag = tag<string>({ label: "c3Dir" });

export interface CliOptions {
  command: string;
  args: string[];
  json: boolean;
  flat?: boolean;
  feature?: boolean;
  container?: string;
  c3Dir?: string;
  help: boolean;
  version: boolean;
}

export const optionsTag = tag<CliOptions>({ label: "options" });

// Atoms — cached per scope, built once
export const graphAtom = atom({
  deps: { c3Dir: tags.required(c3DirTag) },
  factory: async (_ctx: Lite.ResolveContext, { c3Dir }: { c3Dir: string }) => {
    const docs = await walkC3Docs(c3Dir);
    return buildRelationshipGraph(docs);
  },
});
