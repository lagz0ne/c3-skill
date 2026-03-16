import * as vscode from "vscode";
import * as path from "path";
import { DocMap } from "./docMap";

export function registerTreeCommands(
  context: vscode.ExtensionContext,
  docMap: DocMap,
  workspaceFolder: vscode.WorkspaceFolder
): void {
  context.subscriptions.push(
    vscode.commands.registerCommand("c3Nav.showDependencies", async (id: string) => {
      const entry = docMap.get(id);
      if (!entry?.uses || entry.uses.length === 0) {
        vscode.window.showInformationMessage(`${id} has no dependencies.`);
        return;
      }

      const items = entry.uses
        .map((depId) => {
          const dep = docMap.get(depId);
          return {
            label: depId,
            description: dep?.title,
            detail: dep?.goal,
            depPath: dep?.path,
          };
        })
        .filter((item) => item.depPath);

      const selected = await vscode.window.showQuickPick(items, {
        placeHolder: `Dependencies of ${id}`,
      });

      if (selected?.depPath) {
        vscode.window.showTextDocument(vscode.Uri.file(selected.depPath), { preview: true });
      }
    }),

    vscode.commands.registerCommand("c3Nav.showDependents", async (id: string) => {
      const dependents: Array<{ id: string; title?: string; path: string }> = [];

      for (const [entryId, entry] of docMap.entries()) {
        if (entry.uses?.includes(id) || entry.via?.includes(id)) {
          dependents.push({ id: entryId, title: entry.title, path: entry.path });
        }
      }

      if (dependents.length === 0) {
        vscode.window.showInformationMessage(`No components depend on ${id}.`);
        return;
      }

      const items = dependents.map((dep) => ({
        label: dep.id,
        description: dep.title,
        depPath: dep.path,
      }));

      const selected = await vscode.window.showQuickPick(items, {
        placeHolder: `Dependents of ${id}`,
      });

      if (selected?.depPath) {
        vscode.window.showTextDocument(vscode.Uri.file(selected.depPath), { preview: true });
      }
    }),

    vscode.commands.registerCommand("c3Nav.openInCodeMap", async (id: string) => {
      const codeMapPath = path.join(workspaceFolder.uri.fsPath, ".c3", "code-map.yaml");
      const doc = await vscode.workspace.openTextDocument(codeMapPath);
      const editor = await vscode.window.showTextDocument(doc, { preview: true });

      for (let i = 0; i < doc.lineCount; i++) {
        if (doc.lineAt(i).text.includes(id)) {
          const range = new vscode.Range(i, 0, i, 0);
          editor.revealRange(range, vscode.TextEditorRevealType.InCenter);
          editor.selection = new vscode.Selection(range.start, range.start);
          break;
        }
      }
    }),

    vscode.commands.registerCommand("c3Nav.copyId", (id: string) => {
      vscode.env.clipboard.writeText(id);
      vscode.window.showInformationMessage(`Copied: ${id}`);
    })
  );
}
