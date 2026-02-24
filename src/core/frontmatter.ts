// src/core/frontmatter.ts
import { parse as parseYaml } from "yaml";
import { z } from "zod";

export const frontmatterSchema = z.object({
  id: z.string(),
  "c3-version": z.number().optional(),
  title: z.string().optional(),
  type: z.enum(["container", "component", "adr"]).optional(),
  category: z.string().optional(),
  parent: z.string().optional(),
  goal: z.string().optional(),
  summary: z.string().optional(),
  boundary: z.string().optional(),
  affects: z.array(z.string()).optional(),
  status: z.string().optional(),
  date: z.string().optional(),
  scope: z.array(z.string()).optional(),
  refs: z.array(z.string()).optional(),
}).passthrough();

export type Frontmatter = z.infer<typeof frontmatterSchema>;

export type DocType = "context" | "container" | "component" | "ref" | "adr";

export interface ParsedDoc {
  frontmatter: Frontmatter;
  body: string;
  path: string;
}

export function parseFrontmatter(content: string): { frontmatter: Frontmatter | null; body: string } {
  if (!content.startsWith("---\n")) return { frontmatter: null, body: content };

  const end = content.indexOf("\n---\n", 4);
  if (end === -1) return { frontmatter: null, body: content };

  const yamlStr = content.slice(4, end);
  const body = content.slice(end + 5);

  try {
    const parsed = parseYaml(yamlStr);
    // YAML parses empty values (e.g. "goal: ") as null — strip them for zod
    if (parsed && typeof parsed === "object") {
      for (const key of Object.keys(parsed)) {
        if ((parsed as Record<string, unknown>)[key] === null) {
          delete (parsed as Record<string, unknown>)[key];
        }
      }
    }
    const result = frontmatterSchema.safeParse(parsed);
    return { frontmatter: result.success ? result.data : null, body };
  } catch {
    return { frontmatter: null, body: content };
  }
}

export function classifyDoc(fm: Frontmatter): DocType | null {
  if (fm.id === "c3-0") return "context";
  if (fm.type === "container") return "container";
  if (fm.type === "component") return "component";
  if (fm.type === "adr" || fm.id.startsWith("adr-")) return "adr";
  if (fm.id.startsWith("ref-")) return "ref";
  return null;
}

export function deriveRelationships(fm: Frontmatter): string[] {
  const rels: string[] = [];
  if (fm.parent) rels.push(fm.parent);
  if (fm.affects) rels.push(...fm.affects);
  if (fm.refs) rels.push(...fm.refs);
  if (fm.scope) rels.push(...fm.scope);
  return rels;
}
