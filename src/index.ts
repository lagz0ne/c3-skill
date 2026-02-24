// src/index.ts — library exports for programmatic use
export { loadConfig, findC3Dir, type C3Config } from "./core/config";
export { parseFrontmatter, classifyDoc, deriveRelationships, type Frontmatter, type DocType, type ParsedDoc } from "./core/frontmatter";
export { walkC3Docs, buildRelationshipGraph, type C3Entity, type C3Graph } from "./core/walker";
export { nextContainerId, nextComponentId, nextAdrId } from "./core/numbering";
