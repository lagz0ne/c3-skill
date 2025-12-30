import type { Plugin } from "@opencode-ai/plugin"
import { existsSync } from "fs"

export const C3Plugin: Plugin = async (ctx) => {
  const c3Path = `${ctx.worktree}/.c3`

  return {
    // ─────────────────────────────────────────────
    // EVENT OBSERVER: Track file and session events
    // ─────────────────────────────────────────────
    event: async ({ event }) => {
      // Session created - auto-detect C3 project
      if (event.type === "session.created") {
        const hasC3 = existsSync(`${c3Path}/README.md`)
        if (hasC3) {
          console.log("🏗️  C3 architecture detected")
        }
        return
      }

      // File changes - track ADR and container modifications
      if (event.type === "file.edited" && "path" in event.properties) {
        const path = event.properties.path as string

        // Track ADR changes
        if (path.includes("/adr-") && path.endsWith(".md")) {
          console.log(`📋 ADR modified: ${path}`)
        }

        // Track container changes
        if (/\.c3\/c3-\d+-/.test(path)) {
          console.log(`📦 Container doc modified: ${path}`)
        }
      }
    },

    // ─────────────────────────────────────────────
    // TOOL GUARDS: Warn/block on sensitive C3 edits
    // ─────────────────────────────────────────────
    "tool.execute.before": async (input, output) => {
      const { tool } = input
      const { args } = output

      // Warn on Context doc edits (high-impact)
      if (tool === "edit" && args?.file_path === `${c3Path}/README.md`) {
        console.warn("⚠️  Editing Context document - system-wide impact")
      }

      // Block deletion of C3 docs
      if (tool === "bash" && typeof args?.command === "string") {
        if (/rm\s+(-rf?\s+)?.*\.c3/.test(args.command)) {
          throw new Error("🛑 Cannot delete C3 architecture documents")
        }
      }
    },

    // ─────────────────────────────────────────────
    // TOOL OBSERVERS: React after tool completion
    // ─────────────────────────────────────────────
    "tool.execute.after": async (input, output) => {
      const { tool } = input
      const { title, metadata } = output

      // Log C3 doc modifications based on title/metadata
      // Note: args are not available in after hook, so we check output
      if (tool === "write" && title?.includes(".c3/")) {
        console.log(`📝 C3 doc written: ${title}`)
      }

      // Also check metadata if available
      if (tool === "write" && metadata?.file_path?.includes(".c3/")) {
        console.log(`📝 C3 doc written: ${metadata.file_path}`)
      }
    },

    // ─────────────────────────────────────────────
    // PERMISSION GATE: Protect critical operations
    // ─────────────────────────────────────────────
    "permission.ask": async (input, output) => {
      // Auto-allow reads on C3 docs
      if (input.permission === "Read" && input.path?.includes(".c3/")) {
        output.status = "allow"
        return
      }

      // Default: let user decide (don't mutate output)
    },
  }
}

export default C3Plugin
