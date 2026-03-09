import * as fs from "fs";

/** Matches c3-0 through c3-999 and ref-xxx-yyy patterns */
export const C3_ID_PATTERN = /\b(c3-\d{1,3}|ref-[a-z][\w-]*)\b/g;

export interface DocEntry {
  path: string;
  title?: string;
  goal?: string;
  summary?: string;
}

/**
 * Parse YAML frontmatter from a markdown file.
 * Extracts title, goal, and summary fields from the --- delimited block.
 */
export function parseFrontmatter(filePath: string): Pick<DocEntry, "title" | "goal" | "summary"> {
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
  const result: Pick<DocEntry, "title" | "goal" | "summary"> = {};

  const titleMatch = fm.match(/^title:\s*(.+)$/m);
  if (titleMatch) {
    result.title = titleMatch[1].trim();
  }

  const goalMatch = fm.match(/^goal:\s*(.+)$/m);
  if (goalMatch) {
    result.goal = goalMatch[1].trim();
  }

  const summaryMatch = fm.match(/^summary:\s*(.+)$/m);
  if (summaryMatch) {
    result.summary = summaryMatch[1].trim();
  }

  return result;
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
  const regex = /\b(c3-\d{1,3}|ref-[a-z][\w-]*)\b/g;
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
