#!/usr/bin/env bun
/**
 * Checklist runner - validates C3 output against structural requirements
 * Usage: bun check.ts <checklist.yaml> <target-dir>
 */

import { readFile, readdir, stat } from "fs/promises";
import { existsSync } from "fs";
import { join, relative } from "path";
import { parse as parseYaml } from "yaml";
import { Glob } from "bun";

interface Check {
  id: string;
  description?: string;
  type: string;
  path?: string;
  pattern?: string;
  regex?: string;
  required_fields?: string[];
  required?: string[];
  section?: string;
  expect?: Record<string, unknown>;
}

interface Checklist {
  name: string;
  description?: string;
  checks: Check[];
}

interface CheckResult {
  id: string;
  status: "pass" | "fail";
  reason?: string;
}

interface Results {
  passed: number;
  failed: number;
  total: number;
  checks: CheckResult[];
}

const checklistPath = Bun.argv[2];
const targetDir = Bun.argv[3];

if (!checklistPath || !targetDir) {
  console.error("Usage: bun check.ts <checklist.yaml> <target-dir>");
  process.exit(1);
}

// --- Check implementations ---

async function checkPathExists(check: Check, targetDir: string): Promise<CheckResult> {
  const fullPath = join(targetDir, check.path!);
  const exists = existsSync(fullPath);
  return {
    id: check.id,
    status: exists ? "pass" : "fail",
    reason: exists ? undefined : `Path not found: ${check.path}`,
  };
}

async function checkGlobExists(check: Check, targetDir: string): Promise<CheckResult> {
  const glob = new Glob(check.pattern!);
  const matches: string[] = [];

  for await (const file of glob.scan({ cwd: targetDir, absolute: false })) {
    matches.push(file);
  }

  const exists = matches.length > 0;
  return {
    id: check.id,
    status: exists ? "pass" : "fail",
    reason: exists ? undefined : `No files match: ${check.pattern}`,
  };
}

async function checkGlobNotExists(check: Check, targetDir: string): Promise<CheckResult> {
  const glob = new Glob(check.pattern!);
  const matches: string[] = [];

  for await (const file of glob.scan({ cwd: targetDir, absolute: false })) {
    matches.push(file);
  }

  const notExists = matches.length === 0;
  return {
    id: check.id,
    status: notExists ? "pass" : "fail",
    reason: notExists ? undefined : `Unexpected files found: ${matches.join(", ")}`,
  };
}

function extractFrontmatter(content: string): Record<string, unknown> | null {
  const match = content.match(/^---\n([\s\S]*?)\n---/);
  if (!match) return null;
  try {
    return parseYaml(match[1]);
  } catch {
    return null;
  }
}

async function checkFrontmatter(check: Check, targetDir: string): Promise<CheckResult> {
  const fullPath = join(targetDir, check.path!);

  if (!existsSync(fullPath)) {
    return { id: check.id, status: "fail", reason: `File not found: ${check.path}` };
  }

  const content = await readFile(fullPath, "utf-8");
  const frontmatter = extractFrontmatter(content);

  if (!frontmatter) {
    return { id: check.id, status: "fail", reason: "No valid YAML frontmatter found" };
  }

  // Check required fields
  const missing: string[] = [];
  for (const field of check.required_fields ?? []) {
    if (!(field in frontmatter)) {
      missing.push(field);
    }
  }

  if (missing.length > 0) {
    return { id: check.id, status: "fail", reason: `Missing fields: ${missing.join(", ")}` };
  }

  // Check expected values
  if (check.expect) {
    for (const [key, expected] of Object.entries(check.expect)) {
      if (frontmatter[key] !== expected) {
        return {
          id: check.id,
          status: "fail",
          reason: `Expected ${key}=${expected}, got ${frontmatter[key]}`,
        };
      }
    }
  }

  return { id: check.id, status: "pass" };
}

async function checkFrontmatterGlob(check: Check, targetDir: string): Promise<CheckResult> {
  const glob = new Glob(check.pattern!);
  const files: string[] = [];

  for await (const file of glob.scan({ cwd: targetDir, absolute: false })) {
    files.push(file);
  }

  if (files.length === 0) {
    return { id: check.id, status: "fail", reason: `No files match: ${check.pattern}` };
  }

  for (const file of files) {
    const result = await checkFrontmatter(
      { ...check, path: file },
      targetDir
    );
    if (result.status === "fail") {
      return { ...result, reason: `${file}: ${result.reason}` };
    }
  }

  return { id: check.id, status: "pass" };
}

async function checkSections(check: Check, targetDir: string): Promise<CheckResult> {
  const fullPath = join(targetDir, check.path!);

  if (!existsSync(fullPath)) {
    return { id: check.id, status: "fail", reason: `File not found: ${check.path}` };
  }

  const content = await readFile(fullPath, "utf-8");
  const missing: string[] = [];

  for (const section of check.required ?? []) {
    if (!content.includes(section)) {
      missing.push(section);
    }
  }

  if (missing.length > 0) {
    return { id: check.id, status: "fail", reason: `Missing sections: ${missing.join(", ")}` };
  }

  return { id: check.id, status: "pass" };
}

async function checkSectionsGlob(check: Check, targetDir: string): Promise<CheckResult> {
  const glob = new Glob(check.pattern!);
  const files: string[] = [];

  for await (const file of glob.scan({ cwd: targetDir, absolute: false })) {
    files.push(file);
  }

  if (files.length === 0) {
    return { id: check.id, status: "fail", reason: `No files match: ${check.pattern}` };
  }

  for (const file of files) {
    const result = await checkSections({ ...check, path: file }, targetDir);
    if (result.status === "fail") {
      return { ...result, reason: `${file}: ${result.reason}` };
    }
  }

  return { id: check.id, status: "pass" };
}

async function checkContains(check: Check, targetDir: string): Promise<CheckResult> {
  const fullPath = join(targetDir, check.path!);

  if (!existsSync(fullPath)) {
    return { id: check.id, status: "fail", reason: `File not found: ${check.path}` };
  }

  const content = await readFile(fullPath, "utf-8");
  const contains = content.includes(check.pattern!);

  return {
    id: check.id,
    status: contains ? "pass" : "fail",
    reason: contains ? undefined : `Pattern not found: ${check.pattern}`,
  };
}

async function checkRegex(check: Check, targetDir: string): Promise<CheckResult> {
  const fullPath = join(targetDir, check.path!);

  if (!existsSync(fullPath)) {
    return { id: check.id, status: "fail", reason: `File not found: ${check.path}` };
  }

  const content = await readFile(fullPath, "utf-8");
  const regex = new RegExp(check.pattern ?? check.regex!, "m");
  const matches = regex.test(content);

  return {
    id: check.id,
    status: matches ? "pass" : "fail",
    reason: matches ? undefined : `Regex not matched: ${check.pattern ?? check.regex}`,
  };
}

async function checkRegexGlob(check: Check, targetDir: string): Promise<CheckResult> {
  const glob = new Glob(check.pattern!);
  const files: string[] = [];

  for await (const file of glob.scan({ cwd: targetDir, absolute: false })) {
    files.push(file);
  }

  if (files.length === 0) {
    return { id: check.id, status: "fail", reason: `No files match: ${check.pattern}` };
  }

  const regex = new RegExp(check.regex!, "m");

  for (const file of files) {
    const fullPath = join(targetDir, file);
    const content = await readFile(fullPath, "utf-8");
    if (!regex.test(content)) {
      return { id: check.id, status: "fail", reason: `${file}: Regex not matched` };
    }
  }

  return { id: check.id, status: "pass" };
}

async function checkSectionGlob(check: Check, targetDir: string): Promise<CheckResult> {
  const glob = new Glob(check.pattern!);
  const files: string[] = [];

  for await (const file of glob.scan({ cwd: targetDir, absolute: false })) {
    files.push(file);
  }

  if (files.length === 0) {
    return { id: check.id, status: "fail", reason: `No files match: ${check.pattern}` };
  }

  for (const file of files) {
    const fullPath = join(targetDir, file);
    const content = await readFile(fullPath, "utf-8");
    if (!content.includes(check.section!)) {
      return { id: check.id, status: "fail", reason: `${file}: Missing section ${check.section}` };
    }
  }

  return { id: check.id, status: "pass" };
}

// --- Main ---

async function runCheck(check: Check, targetDir: string): Promise<CheckResult> {
  switch (check.type) {
    case "path_exists":
      return checkPathExists(check, targetDir);
    case "glob_exists":
      return checkGlobExists(check, targetDir);
    case "glob_not_exists":
      return checkGlobNotExists(check, targetDir);
    case "frontmatter":
      return checkFrontmatter(check, targetDir);
    case "frontmatter_glob":
      return checkFrontmatterGlob(check, targetDir);
    case "sections":
      return checkSections(check, targetDir);
    case "sections_glob":
      return checkSectionsGlob(check, targetDir);
    case "contains":
      return checkContains(check, targetDir);
    case "regex":
      return checkRegex(check, targetDir);
    case "regex_glob":
      return checkRegexGlob(check, targetDir);
    case "section_glob":
      return checkSectionGlob(check, targetDir);
    default:
      return { id: check.id, status: "fail", reason: `Unknown check type: ${check.type}` };
  }
}

async function main() {
  const checklistContent = await readFile(checklistPath, "utf-8");
  const checklist: Checklist = parseYaml(checklistContent);

  const results: Results = {
    passed: 0,
    failed: 0,
    total: checklist.checks.length,
    checks: [],
  };

  for (const check of checklist.checks) {
    const result = await runCheck(check, targetDir);
    results.checks.push(result);

    if (result.status === "pass") {
      results.passed++;
    } else {
      results.failed++;
    }
  }

  console.log(JSON.stringify(results, null, 2));
}

main().catch((e) => {
  console.error(`Checklist error: ${e}`);
  process.exit(1);
});
