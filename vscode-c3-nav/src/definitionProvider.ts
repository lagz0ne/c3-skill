import * as vscode from "vscode";
import * as path from "path";
import { DocMap } from "./docMap";
import { getIdAtPosition, getPathAtPosition } from "./utils";

export class C3DefinitionProvider implements vscode.DefinitionProvider {
  constructor(private docMap: DocMap) {}

  provideDefinition(
    document: vscode.TextDocument,
    position: vscode.Position
  ): vscode.Definition | undefined {
    const line = document.lineAt(position.line).text;

    // Try C3 ID first
    const idMatch = getIdAtPosition(line, position.character);
    if (idMatch) {
      const entry = this.docMap.get(idMatch.id);
      if (entry) {
        return new vscode.Location(vscode.Uri.file(entry.path), new vscode.Position(0, 0));
      }
    }

    // Try quoted path
    const pathMatch = getPathAtPosition(line, position.character);
    if (pathMatch) {
      const workspaceFolder = vscode.workspace.workspaceFolders?.[0];
      if (workspaceFolder) {
        const absPath = path.join(workspaceFolder.uri.fsPath, pathMatch.folderPath);
        return new vscode.Location(vscode.Uri.file(absPath), new vscode.Position(0, 0));
      }
    }

    return undefined;
  }
}
