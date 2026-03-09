import * as vscode from "vscode";
import { DocMap } from "./docMap";
import { getIdAtPosition } from "./utils";

export class C3DefinitionProvider implements vscode.DefinitionProvider {
  constructor(private docMap: DocMap) {}

  provideDefinition(
    document: vscode.TextDocument,
    position: vscode.Position
  ): vscode.Definition | undefined {
    const line = document.lineAt(position.line).text;
    const match = getIdAtPosition(line, position.character);
    if (!match) {
      return undefined;
    }

    const entry = this.docMap.get(match.id);
    if (!entry) {
      return undefined;
    }

    return new vscode.Location(vscode.Uri.file(entry.path), new vscode.Position(0, 0));
  }
}
