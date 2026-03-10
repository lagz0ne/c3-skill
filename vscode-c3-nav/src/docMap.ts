import * as vscode from "vscode";
import * as path from "path";
import * as fs from "fs";
import { DocEntry, extractIdFromFilename, parseFrontmatter } from "./utils";

export class DocMap {
  private map = new Map<string, DocEntry>();
  private watcher: vscode.FileSystemWatcher | undefined;
  private _onDidRebuild = new vscode.EventEmitter<void>();
  readonly onDidRebuild = this._onDidRebuild.event;

  async build(workspaceFolder: vscode.WorkspaceFolder): Promise<void> {
    this.map.clear();

    const c3Dir = vscode.Uri.joinPath(workspaceFolder.uri, ".c3");
    const pattern = new vscode.RelativePattern(c3Dir, "**/*.md");
    const files = await vscode.workspace.findFiles(pattern);

    for (const file of files) {
      const filename = path.basename(file.fsPath);
      let id = extractIdFromFilename(filename);

      // README.md files store their ID in frontmatter (containers like c3-0, c3-1, c3-2)
      if (!id && filename === "README.md") {
        const frontmatter = parseFrontmatter(file.fsPath);
        id = this.readFrontmatterId(file.fsPath);
        if (!id) {
          continue;
        }
        if (this.map.has(id)) {
          console.warn(`[C3 Nav] Duplicate ID "${id}" — keeping first match, skipping ${file.fsPath}`);
          continue;
        }
        this.map.set(id, { path: file.fsPath, ...frontmatter });
        continue;
      }

      if (!id) {
        continue;
      }

      if (this.map.has(id)) {
        console.warn(`[C3 Nav] Duplicate ID "${id}" — keeping first match, skipping ${file.fsPath}`);
        continue;
      }

      const frontmatter = parseFrontmatter(file.fsPath);
      this.map.set(id, { path: file.fsPath, ...frontmatter });
    }

    console.log(`[C3 Nav] Built document map with ${this.map.size} entries`);
    this._onDidRebuild.fire();
  }

  get(id: string): DocEntry | undefined {
    return this.map.get(id);
  }

  size(): number {
    return this.map.size;
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

  private readFrontmatterId(filePath: string): string | undefined {
    let content: string;
    try {
      content = fs.readFileSync(filePath, "utf-8");
    } catch {
      return undefined;
    }
    const fmMatch = content.match(/^---\n([\s\S]*?)\n---/);
    if (!fmMatch) {
      return undefined;
    }
    const idMatch = fmMatch[1].match(/^id:\s*(.+)$/m);
    return idMatch ? idMatch[1].trim() : undefined;
  }

  dispose(): void {
    this.watcher?.dispose();
    this._onDidRebuild.dispose();
  }
}
