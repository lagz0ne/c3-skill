import * as vscode from "vscode";
import { DocMap } from "./docMap";
import { C3CodeLensProvider } from "./codeLensProvider";
import { C3DefinitionProvider } from "./definitionProvider";
import { C3HoverProvider } from "./hoverProvider";

const YAML_IN_C3: vscode.DocumentSelector = {
  scheme: "file",
  pattern: "**/.c3/**/*.yaml",
};

export async function activate(context: vscode.ExtensionContext): Promise<void> {
  const workspaceFolder = vscode.workspace.workspaceFolders?.[0];
  if (!workspaceFolder) {
    return;
  }

  const docMap = new DocMap();
  await docMap.build(workspaceFolder);

  const codeLensProvider = new C3CodeLensProvider(docMap);

  // Refresh CodeLens after doc map rebuilds (single watcher, no race condition)
  docMap.onDidRebuild(() => codeLensProvider.refresh());

  context.subscriptions.push(
    docMap.startWatching(workspaceFolder),
    vscode.languages.registerCodeLensProvider(YAML_IN_C3, codeLensProvider),
    vscode.languages.registerDefinitionProvider(YAML_IN_C3, new C3DefinitionProvider(docMap)),
    vscode.languages.registerHoverProvider(YAML_IN_C3, new C3HoverProvider(docMap)),
    vscode.commands.registerCommand("c3Nav.openDocument", (filePath: string) => {
      const uri = vscode.Uri.file(filePath);
      vscode.window.showTextDocument(uri, { preview: true });
    })
  );

  console.log("[C3 Nav] Extension activated");
}

export function deactivate(): void {
  // cleanup handled by disposables
}
