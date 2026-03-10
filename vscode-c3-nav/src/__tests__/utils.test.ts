import { describe, it, expect, beforeEach, afterEach } from "vitest";
import { extractIdFromFilename, stripGlobSuffix, parseFrontmatter } from "../utils";
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
