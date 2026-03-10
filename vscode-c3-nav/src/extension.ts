import * as vscode from "vscode";
import * as path from "path";
import * as fs from "fs";
import { DocMap } from "./docMap";
import { C3CodeLensProvider } from "./codeLensProvider";
import { C3DefinitionProvider } from "./definitionProvider";
import { C3HoverProvider } from "./hoverProvider";
import { C3TreeViewProvider } from "./treeViewProvider";
import { registerTreeCommands } from "./treeCommands";

const YAML_IN_C3: vscode.DocumentSelector = {
  scheme: "file",
  pattern: "**/.c3/**/*.yaml",
};

const MD_IN_C3: vscode.DocumentSelector = {
  scheme: "file",
  pattern: "**/.c3/**/*.md",
};

export async function activate(context: vscode.ExtensionContext): Promise<void> {
  const workspaceFolder = vscode.workspace.workspaceFolders?.[0];
  if (!workspaceFolder) {
    return;
  }

  // Set context for conditional view visibility
  vscode.commands.executeCommand("setContext", "c3Nav.hasC3Folder", true);

  const docMap = new DocMap();
  await docMap.build(workspaceFolder);

  // CodeLens
  const codeLensProvider = new C3CodeLensProvider(docMap);
  docMap.onDidRebuild(() => codeLensProvider.refresh());

  // Tree View
  const treeViewProvider = new C3TreeViewProvider(docMap);
  docMap.onDidRebuild(() => treeViewProvider.refresh());

  context.subscriptions.push(
    // File watching
    docMap.startWatching(workspaceFolder),

    // CodeLens for YAML and Markdown
    vscode.languages.registerCodeLensProvider(YAML_IN_C3, codeLensProvider),
    vscode.languages.registerCodeLensProvider(MD_IN_C3, codeLensProvider),

    // Definition for YAML and Markdown
    vscode.languages.registerDefinitionProvider(YAML_IN_C3, new C3DefinitionProvider(docMap)),
    vscode.languages.registerDefinitionProvider(MD_IN_C3, new C3DefinitionProvider(docMap)),

    // Hover for YAML and Markdown
    vscode.languages.registerHoverProvider(YAML_IN_C3, new C3HoverProvider(docMap)),
    vscode.languages.registerHoverProvider(MD_IN_C3, new C3HoverProvider(docMap)),

    // Tree view
    vscode.window.createTreeView("c3Navigator", {
      treeDataProvider: treeViewProvider,
      showCollapseAll: true,
    }),

    // Commands
    vscode.commands.registerCommand("c3Nav.openDocument", (filePath: string) => {
      const uri = vscode.Uri.file(filePath);
      vscode.window.showTextDocument(uri, { preview: true });
    }),
    vscode.commands.registerCommand("c3Nav.revealPath", async (relativePath: string) => {
      const absPath = path.join(workspaceFolder.uri.fsPath, relativePath);
      const uri = vscode.Uri.file(absPath);

      if (fs.existsSync(absPath) && fs.statSync(absPath).isDirectory()) {
        await vscode.commands.executeCommand("revealInExplorer", uri);
        await vscode.commands.executeCommand("list.expand");
      } else if (fs.existsSync(absPath)) {
        vscode.window.showTextDocument(uri, { preview: true });
      } else {
        vscode.window.showWarningMessage(`Path not found: ${relativePath}`);
      }
    }),
    vscode.commands.registerCommand("c3Nav.toggleViewMode", () => {
      treeViewProvider.toggleViewMode();
    })
  );

  // Register tree context menu commands
  registerTreeCommands(context, docMap, workspaceFolder);

}

export function deactivate(): void {
  // cleanup handled by disposables
}
