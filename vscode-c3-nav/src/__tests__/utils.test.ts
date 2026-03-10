import { describe, it, expect } from "vitest";
import { extractIdFromFilename, stripGlobSuffix } from "../utils";

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
