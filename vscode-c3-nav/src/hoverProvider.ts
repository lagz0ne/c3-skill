import * as vscode from "vscode";
import { DocMap } from "./docMap";
import { getIdAtPosition } from "./utils";

export class C3HoverProvider implements vscode.HoverProvider {
  constructor(private docMap: DocMap) {}

  provideHover(
    document: vscode.TextDocument,
    position: vscode.Position
  ): vscode.Hover | undefined {
    const line = document.lineAt(position.line).text;
    const match = getIdAtPosition(line, position.character);
    if (!match) {
      return undefined;
    }

    const entry = this.docMap.get(match.id);
    if (!entry) {
      return undefined;
    }

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

    if (entry.summary) {
      md.appendMarkdown(`*Summary:* ${entry.summary}\n\n`);
    }

    const fileUri = vscode.Uri.file(entry.path);
    md.appendMarkdown(`[Open document](${fileUri})`);

    const range = new vscode.Range(
      position.line,
      match.start,
      position.line,
      match.end
    );

    return new vscode.Hover(md, range);
  }
}
