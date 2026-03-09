import * as vscode from "vscode";
import * as path from "path";
import { DocEntry, extractIdFromFilename, parseFrontmatter } from "./utils";

export class DocMap {
  private map = new Map<string, DocEntry>();
  private watcher: vscode.FileSystemWatcher | undefined;

  async build(workspaceFolder: vscode.WorkspaceFolder): Promise<void> {
    this.map.clear();

    const c3Dir = vscode.Uri.joinPath(workspaceFolder.uri, ".c3");
    const pattern = new vscode.RelativePattern(c3Dir, "**/*.md");
    const files = await vscode.workspace.findFiles(pattern);

    for (const file of files) {
      const filename = path.basename(file.fsPath);
      const id = extractIdFromFilename(filename);
      if (!id) {
        continue;
      }

      if (this.map.has(id)) {
        console.warn(`[C3 Nav] Duplicate ID "${id}" — keeping first match, skipping ${file.fsPath}`);
        continue;
      }

      const frontmatter = parseFrontmatter(file.fsPath);
      this.map.set(id, {
        path: file.fsPath,
        ...frontmatter,
      });
    }

    console.log(`[C3 Nav] Built document map with ${this.map.size} entries`);
  }

  get(id: string): DocEntry | undefined {
    return this.map.get(id);
  }

  entries(): IterableIterator<[string, DocEntry]> {
    return this.map.entries();
  }

  startWatching(workspaceFolder: vscode.WorkspaceFolder): vscode.Disposable {
    const c3Dir = vscode.Uri.joinPath(workspaceFolder.uri, ".c3");
    const pattern = new vscode.RelativePattern(c3Dir, "**/*.md");
    this.watcher = vscode.workspace.createFileSystemWatcher(pattern);

    const rebuild = () => this.build(workspaceFolder);
    this.watcher.onDidCreate(rebuild);
    this.watcher.onDidDelete(rebuild);
    this.watcher.onDidChange(rebuild);

    return this.watcher;
  }

  dispose(): void {
    this.watcher?.dispose();
  }
}
