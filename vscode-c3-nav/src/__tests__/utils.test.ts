import { describe, it, expect, beforeEach, afterEach } from "vitest";
import { extractIdFromFilename, stripGlobSuffix, parseFrontmatter, isInFrontmatter, isMarkdownTableRow, getBacktickPathAtPosition } from "../utils";
import { writeFileSync, mkdirSync, rmSync } from "fs";
import { join } from "path";

describe("extractIdFromFilename", () => {
  it("extracts c3 ID from component filename", () => {
    expect(extractIdFromFilename("c3-113-check-cmd.md")).toBe("c3-113");
  });

  it("extracts ref ID from ref filename", () => {
    expect(extractIdFromFilename("ref-frontmatter-docs.md")).toBe("ref-frontmatter-docs");
  });

  it("returns undefined for README.md", () => {
    expect(extractIdFromFilename("README.md")).toBeUndefined();
  });

  it("extracts ADR ID from filename", () => {
    expect(extractIdFromFilename("adr-00000000-c3-adoption.md")).toBe("adr-00000000-c3-adoption");
  });

  it("extracts ADR ID with date prefix", () => {
    expect(extractIdFromFilename("adr-20260309-add-diff-cmd.md")).toBe("adr-20260309-add-diff-cmd");
  });
});

describe("stripGlobSuffix", () => {
  it("strips /** suffix", () => {
    expect(stripGlobSuffix("backend/app/**")).toBe("backend/app");
  });

  it("keeps non-glob paths", () => {
    expect(stripGlobSuffix("cli/main.go")).toBe("cli/main.go");
  });
});

describe("parseFrontmatter", () => {
  const tmpDir = join(__dirname, "__tmp__");
  beforeEach(() => mkdirSync(tmpDir, { recursive: true }));
  afterEach(() => rmSync(tmpDir, { recursive: true, force: true }));

  it("parses component frontmatter with all fields", () => {
    const file = join(tmpDir, "c3-113-check-cmd.md");
    writeFileSync(file, `---
id: c3-113
title: check-cmd
type: component
category: feature
parent: c3-1
goal: Validate docs
status: active
uses: [c3-101, c3-102, c3-104]
---
# check-cmd`);

    const result = parseFrontmatter(file);
    expect(result.title).toBe("check-cmd");
    expect(result.type).toBe("component");
    expect(result.category).toBe("feature");
    expect(result.parent).toBe("c3-1");
    expect(result.status).toBe("active");
    expect(result.uses).toEqual(["c3-101", "c3-102", "c3-104"]);
  });

  it("parses ref frontmatter with via field", () => {
    const file = join(tmpDir, "ref-test.md");
    writeFileSync(file, `---
id: ref-test
title: Test Ref
type: ref
goal: A test ref
via: [c3-101, c3-103]
---
# Test`);

    const result = parseFrontmatter(file);
    expect(result.type).toBe("ref");
    expect(result.via).toEqual(["c3-101", "c3-103"]);
    expect(result.uses).toBeUndefined();
  });

  it("parses ADR frontmatter", () => {
    const file = join(tmpDir, "adr-test.md");
    writeFileSync(file, `---
id: adr-00000000-c3-adoption
title: C3 Adoption
type: adr
status: in-progress
---
# ADR`);

    const result = parseFrontmatter(file);
    expect(result.type).toBe("adr");
    expect(result.status).toBe("in-progress");
  });

  it("defaults status to active when not specified", () => {
    const file = join(tmpDir, "c3-101.md");
    writeFileSync(file, `---
id: c3-101
title: Frontmatter
type: component
category: foundation
parent: c3-1
goal: Parse frontmatter
---`);

    const result = parseFrontmatter(file);
    expect(result.status).toBe("active");
  });
});

describe("isInFrontmatter", () => {
  const lines = [
    "---",           // 0
    "id: c3-113",    // 1
    "parent: c3-1",  // 2
    "uses: [c3-101]",// 3
    "---",           // 4
    "",              // 5
    "# Title",       // 6
    "Body text c3-101", // 7
  ];

  it("returns true for lines inside frontmatter", () => {
    expect(isInFrontmatter(lines, 1)).toBe(true);
    expect(isInFrontmatter(lines, 2)).toBe(true);
    expect(isInFrontmatter(lines, 3)).toBe(true);
  });

  it("returns false for frontmatter delimiters", () => {
    expect(isInFrontmatter(lines, 0)).toBe(false);
    expect(isInFrontmatter(lines, 4)).toBe(false);
  });

  it("returns false for lines outside frontmatter", () => {
    expect(isInFrontmatter(lines, 5)).toBe(false);
    expect(isInFrontmatter(lines, 6)).toBe(false);
    expect(isInFrontmatter(lines, 7)).toBe(false);
  });
});

describe("isMarkdownTableRow", () => {
  it("matches table data rows", () => {
    expect(isMarkdownTableRow("| IN (uses) | Entity graph | c3-102 |")).toBe(true);
  });

  it("does not match separator rows", () => {
    expect(isMarkdownTableRow("|-----------|------|---------|")).toBe(false);
  });

  it("does not match non-table lines", () => {
    expect(isMarkdownTableRow("Some text mentioning c3-101")).toBe(false);
  });
});

describe("getBacktickPathAtPosition", () => {
  it("extracts path from backtick-wrapped text", () => {
    const line = "| `cli/internal/frontmatter/parse.go` | Parses YAML |";
    const result = getBacktickPathAtPosition(line, 10);
    expect(result).toBeDefined();
    expect(result!.rawPath).toBe("cli/internal/frontmatter/parse.go");
    expect(result!.folderPath).toBe("cli/internal/frontmatter/parse.go");
  });

  it("returns undefined when position is outside backticks", () => {
    const line = "| `cli/main.go` | Main entry |";
    const result = getBacktickPathAtPosition(line, 25);
    expect(result).toBeUndefined();
  });

  it("strips glob suffix from backtick paths", () => {
    const line = "Maps to `backend-core/app/**` for coverage";
    const result = getBacktickPathAtPosition(line, 15);
    expect(result!.rawPath).toBe("backend-core/app/**");
    expect(result!.folderPath).toBe("backend-core/app");
  });
});
