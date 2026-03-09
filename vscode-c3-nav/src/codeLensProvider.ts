import * as vscode from "vscode";
import * as path from "path";
import { DocMap } from "./docMap";
import { C3_ID_PATTERN } from "./utils";

export class C3CodeLensProvider implements vscode.CodeLensProvider {
  private _onDidChangeCodeLenses = new vscode.EventEmitter<void>();
  readonly onDidChangeCodeLenses = this._onDidChangeCodeLenses.event;

  constructor(private docMap: DocMap) {}

  refresh(): void {
    this._onDidChangeCodeLenses.fire();
  }

  provideCodeLenses(document: vscode.TextDocument): vscode.CodeLens[] {
    const lenses: vscode.CodeLens[] = [];

    for (let i = 0; i < document.lineCount; i++) {
      const line = document.lineAt(i);
      const regex = new RegExp(C3_ID_PATTERN.source, "g");
      let match: RegExpExecArray | null;

      while ((match = regex.exec(line.text)) !== null) {
        const id = match[0];
        const entry = this.docMap.get(id);
        if (!entry) {
          continue;
        }

        const range = new vscode.Range(i, match.index, i, match.index + id.length);
        const filename = path.basename(entry.path);
        const title = entry.title ? `${entry.title} (${filename})` : filename;

        lenses.push(
          new vscode.CodeLens(range, {
            title: `→ ${title}`,
            command: "c3Nav.openDocument",
            arguments: [entry.path],
          })
        );
      }
    }

    return lenses;
  }
}
