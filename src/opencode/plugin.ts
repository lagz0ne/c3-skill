import type { Plugin } from "@opencode-ai/plugin"

export const C3Plugin: Plugin = async (ctx) => {
  const c3Path = `${ctx.worktree}/.c3`

  return {
    // ─────────────────────────────────────────────
    // TOOL GUARDS: Warn/block on sensitive C3 edits
    // ─────────────────────────────────────────────
    'tool.execute.before': async (input) => {
      const { tool, args } = input as { tool: string; args: Record<string, unknown> }

      // Warn on Context doc edits (high-impact)
      if (tool === 'edit' && args.file_path === `${c3Path}/README.md`) {
        console.warn("⚠️  Editing Context document - system-wide impact")
      }

      // Block deletion of C3 docs
      if (tool === 'bash' && typeof args.command === 'string') {
        if (/rm\s+(-rf?\s+)?.*\.c3/.test(args.command)) {
          throw new Error("🛑 Cannot delete C3 architecture documents")
        }
      }
    },

    // ─────────────────────────────────────────────
    // TOOL OBSERVERS: React after tool completion
    // ─────────────────────────────────────────────
    'tool.execute.after': async (input) => {
      const { tool, args } = input as { tool: string; args: Record<string, unknown> }

      // Log C3 doc modifications
      if (tool === 'write' && typeof args.file_path === 'string') {
        if (args.file_path.includes('.c3/')) {
          console.log(`📝 C3 doc written: ${args.file_path}`)
        }
      }
    },

    // ─────────────────────────────────────────────
    // FILE OBSERVER: Track architecture changes
    // ─────────────────────────────────────────────
    'file.edited': async ({ event }) => {
      const { path } = event as { path: string }

      // Track ADR changes
      if (path.includes('/adr-') && path.endsWith('.md')) {
        console.log(`📋 ADR modified: ${path}`)
      }

      // Track container changes
      if (/\.c3\/c3-\d+-/.test(path)) {
        console.log(`📦 Container doc modified: ${path}`)
      }
    },

    // ─────────────────────────────────────────────
    // SESSION INIT: Auto-detect C3 project
    // ─────────────────────────────────────────────
    'session.created': async () => {
      const file = Bun.file(`${c3Path}/README.md`)
      const hasC3 = await file.exists()
      if (hasC3) {
        console.log("🏗️  C3 architecture detected")
      }
    },

    // ─────────────────────────────────────────────
    // PERMISSION GATE: Protect critical operations
    // ─────────────────────────────────────────────
    'permission.ask': async (input) => {
      const { permission, context } = input as {
        permission: string
        context?: { path?: string }
      }

      // Auto-allow reads on C3 docs
      if (permission === 'read' && context?.path?.includes('.c3/')) {
        return { decision: 'allow' }
      }

      // Default: let user decide
      return { decision: 'ask' }
    },
  }
}

export default C3Plugin
