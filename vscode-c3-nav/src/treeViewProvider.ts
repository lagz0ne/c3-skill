import * as vscode from "vscode";
import { DocMap } from "./docMap";

type ViewMode = "hierarchy" | "flat";

export class C3TreeViewProvider implements vscode.TreeDataProvider<string> {
  private _onDidChangeTreeData = new vscode.EventEmitter<string | undefined>();
  readonly onDidChangeTreeData = this._onDidChangeTreeData.event;

  private viewMode: ViewMode = "hierarchy";

  constructor(private docMap: DocMap) {}

  refresh(): void {
    this._onDidChangeTreeData.fire(undefined);
  }

  toggleViewMode(): void {
    this.viewMode = this.viewMode === "hierarchy" ? "flat" : "hierarchy";
    this.refresh();
  }

  getViewMode(): ViewMode {
    return this.viewMode;
  }

  getTreeItem(element: string): vscode.TreeItem {
    // Group headers (non-ID nodes)
    if (element.startsWith("__group:")) {
      const label = element.replace("__group:", "");
      const item = new vscode.TreeItem(label, vscode.TreeItemCollapsibleState.Expanded);
      item.contextValue = "group";
      return item;
    }

    // Category sub-headers
    if (element.startsWith("__category:")) {
      const parts = element.replace("__category:", "").split(":");
      const label = parts[1].charAt(0).toUpperCase() + parts[1].slice(1);
      const item = new vscode.TreeItem(label, vscode.TreeItemCollapsibleState.Expanded);
      item.contextValue = "category";
      return item;
    }

    const entry = this.docMap.get(element);
    if (!entry) {
      return new vscode.TreeItem(element);
    }

    const hasChildren = this.getChildIds(element).length > 0;
    const collapsible = hasChildren
      ? vscode.TreeItemCollapsibleState.Expanded
      : vscode.TreeItemCollapsibleState.None;

    const label = entry.title ? `${element} · ${entry.title}` : element;
    const item = new vscode.TreeItem(label, collapsible);

    // Status badge
    if (entry.status && entry.status !== "active") {
      item.description = `[${entry.status}]`;
    }

    // Goal as tooltip
    if (entry.goal) {
      item.tooltip = entry.goal;
    }

    // Click opens the document
    item.command = {
      command: "c3Nav.openDocument",
      title: "Open Document",
      arguments: [entry.path],
    };

    item.contextValue = entry.type || "document";

    return item;
  }

  getChildren(element?: string): string[] {
    if (!element) {
      return this.viewMode === "hierarchy" ? this.getHierarchyRoots() : this.getFlatGroups();
    }

    if (this.viewMode === "flat") {
      return this.getFlatGroupChildren(element);
    }

    return this.getHierarchyChildren(element);
  }

  // --- Hierarchy mode ---

  private getHierarchyRoots(): string[] {
    const roots: string[] = [];

    // Find root container (c3-0) or containers with no parent
    for (const [id, entry] of this.docMap.entries()) {
      if (entry.type === "container" && !entry.parent) {
        roots.push(id);
      }
    }

    // Add References and ADRs groups
    if (this.hasEntriesOfType("ref")) {
      roots.push("__group:References");
    }
    if (this.hasEntriesOfType("adr")) {
      roots.push("__group:ADRs");
    }

    return roots;
  }

  private getHierarchyChildren(element: string): string[] {
    if (element.startsWith("__group:")) {
      const group = element.replace("__group:", "");
      if (group === "References") {
        return this.getEntriesOfType("ref");
      }
      if (group === "ADRs") {
        return this.getEntriesOfType("adr");
      }
      return [];
    }

    if (element.startsWith("__category:")) {
      const parts = element.replace("__category:", "").split(":");
      const parentId = parts[0];
      const category = parts[1];
      return this.getChildIds(parentId).filter((childId) => {
        const entry = this.docMap.get(childId);
        return entry?.category === category;
      });
    }

    // For containers: group children by category if they have categories
    const children = this.getChildIds(element);
    const categories = new Set<string>();
    for (const childId of children) {
      const entry = this.docMap.get(childId);
      if (entry?.category) {
        categories.add(entry.category);
      }
    }

    if (categories.size > 1) {
      // Group by category
      return Array.from(categories)
        .sort()
        .map((cat) => `__category:${element}:${cat}`);
    }

    // No categories or single category — list children directly
    return children;
  }

  private getChildIds(parentId: string): string[] {
    const children: string[] = [];
    for (const [id, entry] of this.docMap.entries()) {
      if (entry.parent === parentId && entry.type !== "ref" && entry.type !== "adr") {
        children.push(id);
      }
    }
    return children.sort();
  }

  // --- Flat mode ---

  private getFlatGroups(): string[] {
    const groups: string[] = [];
    if (this.hasEntriesOfType("container")) {
      groups.push("__group:Containers");
    }
    if (this.hasEntriesOfType("component")) {
      groups.push("__group:Components");
    }
    if (this.hasEntriesOfType("ref")) {
      groups.push("__group:References");
    }
    if (this.hasEntriesOfType("adr")) {
      groups.push("__group:ADRs");
    }
    return groups;
  }

  private getFlatGroupChildren(element: string): string[] {
    const group = element.replace("__group:", "");
    const typeMap: Record<string, string> = {
      Containers: "container",
      Components: "component",
      References: "ref",
      ADRs: "adr",
    };
    return this.getEntriesOfType(typeMap[group] || "");
  }

  // --- Helpers ---

  private hasEntriesOfType(type: string): boolean {
    for (const [, entry] of this.docMap.entries()) {
      if (entry.type === type) {
        return true;
      }
    }
    return false;
  }

  private getEntriesOfType(type: string): string[] {
    const ids: string[] = [];
    for (const [id, entry] of this.docMap.entries()) {
      if (entry.type === type) {
        ids.push(id);
      }
    }
    return ids.sort();
  }
}
