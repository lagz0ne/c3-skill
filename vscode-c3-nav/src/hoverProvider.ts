import * as vscode from "vscode";
import * as path from "path";
import * as fs from "fs";
import { DocMap } from "./docMap";
import { getIdAtPosition, getPathAtPosition, getBacktickPathAtPosition } from "./utils";

export class C3HoverProvider implements vscode.HoverProvider {
  constructor(private docMap: DocMap) {}

  provideHover(
    document: vscode.TextDocument,
    position: vscode.Position
  ): vscode.Hover | undefined {
    const line = document.lineAt(position.line).text;

    // Try C3 ID first
    const idMatch = getIdAtPosition(line, position.character);
    if (idMatch) {
      const entry = this.docMap.get(idMatch.id);
      if (entry) {
        return this.buildDocHover(position.line, idMatch, entry);
      }
    }

    // Try quoted path
    const pathMatch = getPathAtPosition(line, position.character);
    if (pathMatch) {
      return this.buildPathHover(position.line, pathMatch);
    }

    // Try backtick path (markdown files)
    const backtickMatch = getBacktickPathAtPosition(line, position.character);
    if (backtickMatch) {
      return this.buildPathHover(position.line, backtickMatch);
    }

    return undefined;
  }

  private buildDocHover(
    lineNumber: number,
    match: { id: string; start: number; end: number },
    entry: { path: string; title?: string; goal?: string; summary?: string; status?: string }
  ): vscode.Hover {
    const md = new vscode.MarkdownString("", true);
    md.isTrusted = true;

    if (entry.title) {
      md.appendMarkdown(`**${match.id}** — ${entry.title}\n\n`);
    } else {
      md.appendMarkdown(`**${match.id}**\n\n`);
    }

    if (entry.goal) {
      md.appendMarkdown(`*Goal:* ${entry.goal}\n\n`);
    }

    if (entry.status && entry.status !== "active") {
      md.appendMarkdown(`*Status:* \`${entry.status}\`\n\n`);
    }

    if (entry.summary) {
      md.appendMarkdown(`*Summary:* ${entry.summary}\n\n`);
    }

    const fileUri = vscode.Uri.file(entry.path);
    md.appendMarkdown(`[Open document](${fileUri})`);

    const range = new vscode.Range(lineNumber, match.start, lineNumber, match.end);
    return new vscode.Hover(md, range);
  }

  private buildPathHover(
    lineNumber: number,
    match: { rawPath: string; folderPath: string; start: number; end: number }
  ): vscode.Hover | undefined {
    const workspaceFolder = vscode.workspace.workspaceFolders?.[0];
    if (!workspaceFolder) {
      return undefined;
    }

    const absPath = path.join(workspaceFolder.uri.fsPath, match.folderPath);
    const exists = fs.existsSync(absPath);
    const isDir = exists && fs.statSync(absPath).isDirectory();

    const md = new vscode.MarkdownString("", true);
    md.isTrusted = true;
    md.appendMarkdown(`**${match.folderPath}**\n\n`);

    if (!exists) {
      md.appendMarkdown(`*Path not found*\n\n`);
    } else if (isDir) {
      md.appendMarkdown(`*Folder* — [Reveal in Explorer](command:c3Nav.revealPath?${encodeURIComponent(JSON.stringify(match.folderPath))})\n\n`);
    } else {
      md.appendMarkdown(`*File* — [Open](command:c3Nav.revealPath?${encodeURIComponent(JSON.stringify(match.folderPath))})\n\n`);
    }

    const range = new vscode.Range(lineNumber, match.start, lineNumber, match.end);
    return new vscode.Hover(md, range);
  }
}
