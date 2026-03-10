import * as fs from "fs";

/** Matches c3-0 through c3-999, ref-xxx-yyy, and adr-xxx-yyy patterns */
export const C3_ID_PATTERN = /\b(c3-\d{1,3}|ref-[a-z][\w-]*|adr-[\w-]+)\b/g;

export interface DocEntry {
  path: string;
  title?: string;
  goal?: string;
  summary?: string;
  type?: "container" | "component" | "ref" | "adr";
  category?: "foundation" | "feature";
  parent?: string;
  uses?: string[];
  via?: string[];
  status?: string;
}

/**
 * Parse YAML frontmatter from a markdown file.
 * Extracts title, goal, and summary fields from the --- delimited block.
 */
export function parseFrontmatter(filePath: string): Omit<DocEntry, "path"> {
  let content: string;
  try {
    content = fs.readFileSync(filePath, "utf-8");
  } catch {
    return {};
  }

  const fmMatch = content.match(/^---\n([\s\S]*?)\n---/);
  if (!fmMatch) {
    return {};
  }

  const fm = fmMatch[1];
  const result: Omit<DocEntry, "path"> = {};

  const titleMatch = fm.match(/^title:\s*(.+)$/m);
  if (titleMatch) {
    result.title = stripYamlQuotes(titleMatch[1].trim());
  }

  const goalMatch = fm.match(/^goal:\s*(.+)$/m);
  if (goalMatch) {
    result.goal = stripYamlQuotes(goalMatch[1].trim());
  }

  const summaryMatch = fm.match(/^summary:\s*(.+)$/m);
  if (summaryMatch) {
    result.summary = stripYamlQuotes(summaryMatch[1].trim());
  }

  const typeMatch = fm.match(/^type:\s*(.+)$/m);
  if (typeMatch) {
    result.type = stripYamlQuotes(typeMatch[1].trim()) as DocEntry["type"];
  }

  const categoryMatch = fm.match(/^category:\s*(.+)$/m);
  if (categoryMatch) {
    result.category = stripYamlQuotes(categoryMatch[1].trim()) as DocEntry["category"];
  }

  const parentMatch = fm.match(/^parent:\s*(.+)$/m);
  if (parentMatch) {
    result.parent = stripYamlQuotes(parentMatch[1].trim());
  }

  const usesMatch = fm.match(/^uses:\s*\[([^\]]*)\]$/m);
  if (usesMatch) {
    result.uses = usesMatch[1].split(",").map((s) => s.trim()).filter(Boolean);
  }

  const viaMatch = fm.match(/^via:\s*\[([^\]]*)\]$/m);
  if (viaMatch) {
    result.via = viaMatch[1].split(",").map((s) => s.trim()).filter(Boolean);
  }

  const statusMatch = fm.match(/^status:\s*(.+)$/m);
  result.status = statusMatch ? stripYamlQuotes(statusMatch[1].trim()) : "active";

  return result;
}

/** Strip surrounding YAML quotes (single or double) from a value. */
function stripYamlQuotes(value: string): string {
  if ((value.startsWith('"') && value.endsWith('"')) ||
      (value.startsWith("'") && value.endsWith("'"))) {
    return value.slice(1, -1);
  }
  return value;
}

/**
 * Extract the C3/ref ID from a markdown filename.
 * Returns undefined for non-matching filenames (e.g., "README.md").
 */
export function extractIdFromFilename(filename: string): string | undefined {
  const c3Match = filename.match(/^(c3-\d{1,3})-/);
  if (c3Match) {
    return c3Match[1];
  }

  const refMatch = filename.match(/^(ref-[a-z][\w-]*)\.md$/);
  if (refMatch) {
    return refMatch[1];
  }

  const adrMatch = filename.match(/^(adr-[\w-]+)\.md$/);
  if (adrMatch) {
    return adrMatch[1];
  }

  return undefined;
}

/**
 * Get the C3 ID at a given position in a line of text.
 * Returns the matched ID and its start/end character positions, or undefined.
 */
export function getIdAtPosition(
  lineText: string,
  characterPos: number
): { id: string; start: number; end: number } | undefined {
  const regex = /\b(c3-\d{1,3}|ref-[a-z][\w-]*|adr-[\w-]+)\b/g;
  let match: RegExpExecArray | null;

  while ((match = regex.exec(lineText)) !== null) {
    const start = match.index;
    const end = start + match[0].length;
    if (characterPos >= start && characterPos <= end) {
      return { id: match[0], start, end };
    }
  }

  return undefined;
}

/** Matches a quoted glob path in a YAML list item, e.g. - "backend-core/app/sysadmin/**" */
const QUOTED_PATH_PATTERN = /["']([^"']+)["']/g;

/**
 * Get the quoted path string at a given position in a line of text.
 * Returns the raw path (with glob), the folder path (glob stripped), and positions.
 */
export function getPathAtPosition(
  lineText: string,
  characterPos: number
): { rawPath: string; folderPath: string; start: number; end: number } | undefined {
  const regex = new RegExp(QUOTED_PATH_PATTERN.source, "g");
  let match: RegExpExecArray | null;

  while ((match = regex.exec(lineText)) !== null) {
    // +1 / -1 to cover inside the quotes
    const start = match.index + 1;
    const end = start + match[1].length;
    if (characterPos >= start && characterPos <= end) {
      const rawPath = match[1];
      const folderPath = stripGlobSuffix(rawPath);
      return { rawPath, folderPath, start, end };
    }
  }

  return undefined;
}

/**
 * Strip glob suffixes from a path to get the navigable folder/file.
 * "backend-core/app/sysadmin/**" → "backend-core/app/sysadmin"
 * "ext/companion/main.go" → "ext/companion/main.go" (no glob, keep as-is)
 */
export function stripGlobSuffix(globPath: string): string {
  return globPath.replace(/\/\*\*$/, "").replace(/\/\*\.[a-z]+$/, "");
}

/**
 * Check if a line index is inside the YAML frontmatter block.
 * Frontmatter is between the first `---` (line 0) and the next `---`.
 * Returns true for content lines, false for delimiters and outside.
 */
export function isInFrontmatter(lines: string[], lineIndex: number): boolean {
  if (lines[0] !== "---") {
    return false;
  }
  for (let i = 1; i < lines.length; i++) {
    if (lines[i] === "---") {
      return lineIndex > 0 && lineIndex < i;
    }
  }
  return false;
}

/**
 * Check if a line is a markdown table data row (not a separator row).
 * Table rows start and end with | and contain non-dash content.
 */
export function isMarkdownTableRow(line: string): boolean {
  const trimmed = line.trim();
  if (!trimmed.startsWith("|") || !trimmed.endsWith("|")) {
    return false;
  }
  // Separator rows contain only |, -, :, and spaces
  return !/^\|[\s|:-]+\|$/.test(trimmed);
}

/**
 * Get a backtick-wrapped file path at a given position in a line.
 * Matches patterns like `cli/internal/frontmatter/parse.go`.
 * Returns path info with glob suffix stripped for navigation.
 */
export function getBacktickPathAtPosition(
  lineText: string,
  characterPos: number
): { rawPath: string; folderPath: string; start: number; end: number } | undefined {
  const regex = /`([^`]+\.[a-z]+[^`]*)`|`([^`]+\/[^`]+)`/g;
  let match: RegExpExecArray | null;

  while ((match = regex.exec(lineText)) !== null) {
    const pathValue = match[1] || match[2];
    const start = match.index + 1; // after opening backtick
    const end = start + pathValue.length;
    if (characterPos >= start && characterPos <= end) {
      return {
        rawPath: pathValue,
        folderPath: stripGlobSuffix(pathValue),
        start,
        end,
      };
    }
  }

  return undefined;
}
