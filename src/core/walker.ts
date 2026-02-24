// src/core/walker.ts
import * as fs from "node:fs";
import * as path from "node:path";
import { parseFrontmatter, classifyDoc, deriveRelationships, type Frontmatter, type DocType, type ParsedDoc } from "./frontmatter";

export interface C3Entity {
  id: string;
  type: DocType;
  title: string;
  slug: string;
  path: string;         // relative to .c3/
  frontmatter: Frontmatter;
  body: string;
  relationships: string[];  // IDs this entity references
}

export interface C3Graph {
  entities: Map<string, C3Entity>;
  byType: Map<DocType, C3Entity[]>;

  // Relationship queries
  children(parentId: string): C3Entity[];
  refsFor(entityId: string): C3Entity[];
  citedBy(refId: string): C3Entity[];
  forward(id: string): C3Entity[];        // what does this affect
  reverse(id: string): C3Entity[];        // what points to this
  transitive(id: string, depth: number): C3Entity[]; // blast radius
}

export async function walkC3Docs(c3Dir: string): Promise<ParsedDoc[]> {
  const docs: ParsedDoc[] = [];

  function walk(dir: string) {
    const entries = fs.readdirSync(dir, { withFileTypes: true });
    for (const entry of entries) {
      const fullPath = path.join(dir, entry.name);
      if (entry.isDirectory()) {
        walk(fullPath);
      } else if (entry.name.endsWith(".md")) {
        const content = fs.readFileSync(fullPath, "utf-8");
        const { frontmatter, body } = parseFrontmatter(content);
        if (frontmatter) {
          docs.push({
            frontmatter,
            body,
            path: path.relative(c3Dir, fullPath),
          });
        }
      }
    }
  }

  walk(c3Dir);
  return docs;
}

function slugFromPath(filePath: string): string {
  const base = path.basename(filePath, ".md");
  // Strip ID prefix: c3-1-api -> api, c3-101-auth -> auth, ref-logging -> logging
  return base.replace(/^(c3-\d+-|c3-\d+|ref-|adr-\d+-|README)/, "") || base;
}

export function buildRelationshipGraph(docs: ParsedDoc[]): C3Graph {
  const entities = new Map<string, C3Entity>();
  const byType = new Map<DocType, C3Entity[]>();

  // Phase 1: Build entity map
  for (const doc of docs) {
    const type = classifyDoc(doc.frontmatter);
    if (!type) continue;

    const entity: C3Entity = {
      id: doc.frontmatter.id,
      type,
      title: doc.frontmatter.title || doc.frontmatter.id,
      slug: slugFromPath(doc.path),
      path: doc.path,
      frontmatter: doc.frontmatter,
      body: doc.body,
      relationships: deriveRelationships(doc.frontmatter),
    };

    entities.set(entity.id, entity);
    const list = byType.get(type) || [];
    list.push(entity);
    byType.set(type, list);
  }

  // Phase 2: Build graph query methods
  function children(parentId: string): C3Entity[] {
    return [...entities.values()].filter(e => e.frontmatter.parent === parentId);
  }

  function refsFor(entityId: string): C3Entity[] {
    const entity = entities.get(entityId);
    if (!entity) return [];
    return (entity.frontmatter.refs || [])
      .map(id => entities.get(id))
      .filter((e): e is C3Entity => !!e);
  }

  function citedBy(refId: string): C3Entity[] {
    return [...entities.values()].filter(e =>
      e.frontmatter.refs?.includes(refId) ||
      e.frontmatter.scope?.includes(refId)
    );
  }

  function forward(id: string): C3Entity[] {
    const entity = entities.get(id);
    if (!entity) return [];

    const result: C3Entity[] = [];
    // Direct children
    result.push(...children(id));
    // Entities in affects list
    if (entity.frontmatter.affects) {
      for (const affectedId of entity.frontmatter.affects) {
        const affected = entities.get(affectedId);
        if (affected) result.push(affected);
      }
    }
    // If this is a ref, find citers
    if (entity.type === "ref") {
      result.push(...citedBy(id));
    }
    return result;
  }

  function reverse(id: string): C3Entity[] {
    return [...entities.values()].filter(e =>
      e.relationships.includes(id) ||
      e.frontmatter.parent === id ||
      e.frontmatter.affects?.includes(id)
    );
  }

  function transitive(id: string, depth: number): C3Entity[] {
    const visited = new Set<string>([id]);
    const result: C3Entity[] = [];
    let frontier = [id];

    for (let d = 0; d < depth && frontier.length > 0; d++) {
      const nextFrontier: string[] = [];
      for (const currentId of frontier) {
        for (const entity of forward(currentId)) {
          if (!visited.has(entity.id)) {
            visited.add(entity.id);
            result.push(entity);
            nextFrontier.push(entity.id);
          }
        }
      }
      frontier = nextFrontier;
    }
    return result;
  }

  return { entities, byType, children, refsFor, citedBy, forward, reverse, transitive };
}
