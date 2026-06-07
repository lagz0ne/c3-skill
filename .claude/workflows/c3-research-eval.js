export const meta = {
  name: 'c3-research-eval',
  description: 'Continuous C3 improvement loop: research ideas, apply one minimal change, eval, gate on quality pass-rate, keep or revert',
  whenToUse: 'Push C3 skill quality forward autonomously. Spends tokens running the real eval harness (claude/codex CLIs). Stops after 3 consecutive discards.',
  phases: [
    { title: 'Research', detail: 'fan out: trace-failure, skill-gap, best-practices' },
    { title: 'Rank', detail: 'dedup + rank hypotheses by expected quality impact' },
    { title: 'Iterate', detail: 'per hypothesis: propose -> apply -> eval -> gate -> keep/revert' },
  ],
}

// ---- args (all optional) -------------------------------------------------
// args.cases        : restrict eval to these case ids (default: let proposer pick affected cases)
// args.maxHypotheses: hard cap on hypotheses to try this run (default 6)
// args.focus        : steer research toward a theme, e.g. "adr quality" or "token efficiency"
// args.timestamp    : unix seconds, stamped into learnings (scripts can't read the clock)
const MAX_HYP = (args && args.maxHypotheses) || 6
const FOCUS = (args && args.focus) || 'whole-matrix quality pass-rate (accuracy + adr_quality>=0.8 + canvas_quality>=0.9)'
const TS = (args && args.timestamp) || 0
const CASE_HINT = args && Array.isArray(args.cases) && args.cases.length
  ? `Restrict eval to these cases: ${args.cases.join(', ')}.`
  : 'Pick the smallest set of cases the change could affect.'
// Don't start a new hypothesis unless this many output tokens remain — one
// eval cycle (real claude/codex over a few cases) is expensive.
const EVAL_RESERVE = 250_000

const REPO_RULES = `
Repo rules you MUST follow:
- This repo is the C3 source. Use ONLY the local CLI: \`C3X_MODE=agent bash skills/c3/bin/c3x.sh <cmd>\`. Never bare c3x or a global skill.
- Run python from repo root: \`python scripts/agent_efficiency_eval.py ...\` and \`python scripts/eval_gate.py ...\`.
- Rebuild the Go CLI (\`bash scripts/build.sh\`) ONLY if you edited Go under cli/. Skill/reference markdown edits need no build.
- Treat scripts/agent_efficiency_eval.py as the frozen benchmark — do NOT edit its cases or floors inside this loop.
`

const HYPOTHESES_SCHEMA = {
  type: 'object',
  additionalProperties: false,
  required: ['hypotheses'],
  properties: {
    hypotheses: {
      type: 'array',
      items: {
        type: 'object',
        additionalProperties: false,
        required: ['title', 'rationale', 'change_sketch', 'target_cases', 'expected_impact'],
        properties: {
          title: { type: 'string' },
          rationale: { type: 'string', description: 'evidence: which cases fail and why this addresses it' },
          change_sketch: { type: 'string', description: 'the concrete minimal edit: which file(s), what changes' },
          target_cases: { type: 'array', items: { type: 'string' } },
          expected_impact: { type: 'string', description: 'which pass-rate metric should move and roughly how much' },
        },
      },
    },
  },
}

const RANK_SCHEMA = {
  type: 'object',
  additionalProperties: false,
  required: ['ranked'],
  properties: {
    ranked: {
      type: 'array',
      items: {
        type: 'object',
        additionalProperties: false,
        required: ['title', 'rationale', 'change_sketch', 'target_cases', 'score', 'why'],
        properties: {
          title: { type: 'string' },
          rationale: { type: 'string' },
          change_sketch: { type: 'string' },
          target_cases: { type: 'array', items: { type: 'string' } },
          score: { type: 'number', description: '0-1 expected quality lift per unit risk' },
          why: { type: 'string', description: 'why this rank — impact vs risk vs minimality' },
        },
      },
    },
  },
}

const PROPOSE_SCHEMA = {
  type: 'object',
  additionalProperties: false,
  required: ['applied', 'files', 'summary', 'affected_cases', 'build_needed'],
  properties: {
    applied: { type: 'boolean', description: 'true if a minimal change was written to the working tree' },
    skip_reason: { type: 'string', description: 'if applied=false, why (e.g. change too large/risky)' },
    files: { type: 'array', items: { type: 'string' }, description: 'paths edited or created' },
    summary: { type: 'string', description: 'one line: what changed' },
    affected_cases: { type: 'array', items: { type: 'string' } },
    build_needed: { type: 'boolean', description: 'true only if Go under cli/ changed' },
  },
}

const GATE_SCHEMA = {
  type: 'object',
  additionalProperties: false,
  required: ['decision', 'pass_rate', 'reasons', 'run_path'],
  properties: {
    decision: { type: 'string', enum: ['keep', 'establish', 'discard', 'error'] },
    pass_rate: { type: 'number' },
    baseline_pass_rate: { type: ['number', 'null'] },
    reasons: { type: 'array', items: { type: 'string' } },
    run_path: { type: 'string', description: 'path to the candidate run JSONL' },
    notes: { type: 'string' },
  },
}

// ===========================================================================
// Phase 1 — Research (parallel, read-only)
// ===========================================================================
phase('Research')
log(`Researching improvement hypotheses (focus: ${FOCUS})`)

const researchers = [
  {
    label: 'research:trace-failure',
    type: 'Explore',
    prompt: `Find where the C3 eval matrix is WEAKEST and why.${REPO_RULES}
Read in this order:
1. research/eval/baseline.json and research/eval/history.jsonl — current best pass-rate and the trend.
2. The most recent research/eval/runs/*.jsonl if any exist — per-record exit_code, accuracy_score, adr_quality_score/adr_quality_checks, canvas_quality_passed/canvas_quality_checks, threshold_status, and trace_metrics (broad_search_count, c3_command_count, tool_output_bytes_total).
3. If no run JSONL exists yet, run a DRY plan to see the matrix shape: \`python scripts/agent_efficiency_eval.py --dry-run --output /tmp/dry.jsonl\` (no tokens spent). Then reason from the case definitions in scripts/agent_efficiency_eval.py about likely failure modes.
Identify the 2-4 cases or quality checks most likely failing or fragile, with concrete evidence (which check, which case, what value). Propose hypotheses: each a SINGLE minimal change to the C3 skill/references/CLI that should raise that case's pass-rate. Do NOT propose editing the eval harness itself.`,
  },
  {
    label: 'research:skill-gap',
    type: 'Explore',
    prompt: `Map gaps in the C3 skill that would cause eval failures.${REPO_RULES}
Read skills/c3/SKILL.md and every file in skills/c3/references/ (onboard, query, audit, change, ref, rule, sweep). Cross-reference against the eval case prompts and quality checks in scripts/agent_efficiency_eval.py (CANVAS_EXPECTATIONS, evaluate_adr_quality, accuracy checks).
Find places where the skill's guidance is vague, missing, or contradictory relative to what a case is graded on — e.g. an ADR quality check (root_cause_specific, decision_concrete, alternatives_real, verification_executable) the skill never tells the agent to satisfy, or a canvas section/column the skill doesn't mention. Propose hypotheses: each a SINGLE minimal edit to a skill/reference file that closes one gap. Quote the relevant skill text and the relevant grading check.`,
  },
  {
    label: 'research:best-practices',
    type: 'Explore',
    prompt: `Research external best practices that could improve C3, focused on: ${FOCUS}.${REPO_RULES}
Use web search for current (2025-2026) guidance on the weak area — e.g. high-quality ADR authoring, architecture decision records, lean agent prompting for token efficiency, or structured-canvas/PRD authoring, depending on focus. Skim skills/c3/SKILL.md first so you know what C3 already does and don't propose duplicates.
Propose hypotheses: each a SINGLE minimal, concrete edit to a C3 skill/reference file that imports one external best practice, with a one-line citation of the source idea. Be skeptical — only propose practices that map to a specific eval quality check.`,
  },
]

const research = await parallel(researchers.map((r) => () =>
  agent(r.prompt, { label: r.label, phase: 'Research', agentType: r.type, schema: HYPOTHESES_SCHEMA })
))
const allHyp = research.filter(Boolean).flatMap((r) => r.hypotheses || [])
log(`Collected ${allHyp.length} raw hypotheses from ${research.filter(Boolean).length} researchers`)
if (!allHyp.length) {
  return { stopped: 'no hypotheses produced', kept: [], discarded: [] }
}

// ===========================================================================
// Phase 2 — Rank (single synthesizer, needs ALL hypotheses at once -> barrier ok)
// ===========================================================================
phase('Rank')
const ranking = await agent(
  `You are triaging improvement hypotheses for the C3 skill. Here are ${allHyp.length} raw hypotheses (JSON):

${JSON.stringify(allHyp, null, 2)}

Dedup overlapping ones, drop any that propose editing the eval harness or that aren't a single minimal change. Rank the survivors by expected quality-pass-rate lift per unit of risk — favor small, targeted, high-confidence edits over speculative rewrites. Keep at most ${MAX_HYP}. ${CASE_HINT}`,
  { label: 'rank:synthesize', phase: 'Rank', schema: RANK_SCHEMA }
)
const ranked = (ranking && ranking.ranked || []).slice(0, MAX_HYP)
log(`Ranked ${ranked.length} hypotheses to try (cap ${MAX_HYP})`)

// ===========================================================================
// Phase 3 — Iterate (STRICTLY sequential: each cycle mutates the tree + baseline)
// ===========================================================================
phase('Iterate')
const kept = []
const discarded = []
let consecutiveDiscards = 0

for (let i = 0; i < ranked.length; i++) {
  const h = ranked[i]
  if (consecutiveDiscards >= 3) {
    log(`Stopping: 3 consecutive discards — the idea well is dry, re-research next run.`)
    break
  }
  if (budget.total && budget.remaining() < EVAL_RESERVE) {
    log(`Stopping: ${Math.round(budget.remaining() / 1000)}k tokens left < ${EVAL_RESERVE / 1000}k reserve for one eval cycle.`)
    break
  }

  const tag = `${i + 1}/${ranked.length}`
  log(`[${tag}] "${h.title}" (score ${h.score})`)

  // --- Propose + apply ONE minimal change to the working tree ---
  const proposal = await agent(
    `Apply ONE minimal change to the working tree implementing this hypothesis, then report.
Hypothesis: ${h.title}
Rationale: ${h.rationale}
Change sketch: ${h.change_sketch}
Target cases: ${(h.target_cases || []).join(', ') || '(infer)'}
${REPO_RULES}
Rules: edit ONLY C3 skill/reference/CLI files needed for this one change. Keep the diff small (<50 lines if at all possible). Do NOT touch scripts/agent_efficiency_eval.py. Do NOT commit. If the change turns out to be larger/riskier than "minimal", set applied=false with a skip_reason instead of forcing it. Report the affected eval case ids and whether a Go build is needed.`,
    { label: `propose:${tag}`, phase: 'Iterate', schema: PROPOSE_SCHEMA }
  )
  if (!proposal || !proposal.applied) {
    log(`[${tag}] skipped: ${proposal && proposal.skip_reason || 'no change applied'}`)
    discarded.push({ title: h.title, decision: 'skipped', reason: proposal && proposal.skip_reason })
    consecutiveDiscards++
    continue
  }

  const evalCases = (args && Array.isArray(args.cases) && args.cases.length)
    ? args.cases
    : (proposal.affected_cases && proposal.affected_cases.length ? proposal.affected_cases : null)
  const caseFlags = evalCases ? evalCases.map((c) => `--case ${c}`).join(' ') : ''

  // --- Apply build (if needed), run the real eval, gate it ---
  const gate = await agent(
    `Run the eval + quality gate for the change just applied ("${h.title}", files: ${(proposal.files || []).join(', ')}).${REPO_RULES}
Steps, in order:
1. ${proposal.build_needed ? 'Go code changed — run: bash scripts/build.sh' : 'No build needed (markdown/skill change only).'}
2. Run the real eval on the affected cases (this spends tokens):
   mkdir -p research/eval/runs
   RUN=research/eval/runs/hyp-${tag.replace('/', '-')}-$(git rev-parse --short HEAD).jsonl
   python scripts/agent_efficiency_eval.py --run ${caseFlags} --output "$RUN"
   (If a case id is invalid, list valid ones with --dry-run and pick the closest; record what you ran.)
3. Gate it against the committed baseline:
   python scripts/eval_gate.py --candidate "$RUN" --label "${h.title.replace(/"/g, "'")}"${TS ? ` --timestamp ${TS}` : ''} --json
4. Read eval_gate's JSON output. Its "verdict.decision" is keep | establish | discard. Report that decision verbatim, the candidate pass_rate, baseline_pass_rate, the verdict reasons, and the RUN path. If any command errored, set decision="error" and explain in notes. Do NOT commit and do NOT revert here — the loop handles that.`,
    { label: `eval:${tag}`, phase: 'Iterate', schema: GATE_SCHEMA }
  )

  const decision = gate && gate.decision
  if (decision === 'keep' || decision === 'establish') {
    consecutiveDiscards = 0
    // Accept: re-write baseline, commit, record the learning.
    await agent(
      `The change "${h.title}" PASSED the gate (${decision}, pass_rate ${gate.pass_rate}). Make it durable.${REPO_RULES}
Steps:
1. Promote it to the new baseline (re-run the gate with --update-baseline, no tokens — it reads the existing run file):
   python scripts/eval_gate.py --candidate "${gate.run_path}" --label "${h.title.replace(/"/g, "'")}" --update-baseline${TS ? ` --timestamp ${TS}` : ''} --no-history
2. Write a learning note research/learnings/<date-or-slug>.md following research/learnings/TEMPLATE.md — hypothesis, the change (files ${(proposal.files || []).join(', ')}), result (pass_rate -> ${gate.pass_rate}), why it stuck, follow-ups. ${TS ? `Use timestamp ${TS} for the date.` : ''}
3. Stage and commit ONLY: the changed skill/CLI files, research/eval/baseline.json, research/eval/history.jsonl, the new learning note. Conventional-commit style, subject like: \`research(c3): ${h.title.replace(/"/g, "'").slice(0, 60)}\`. Body: pass-rate delta + the gate reasons.
Report the commit hash.`,
      { label: `keep:${tag}`, phase: 'Iterate' }
    )
    kept.push({ title: h.title, pass_rate: gate.pass_rate, files: proposal.files, run_path: gate.run_path })
    log(`[${tag}] KEPT — pass_rate ${gate.pass_rate}`)
  } else {
    consecutiveDiscards++
    // Reject: revert the working-tree change cleanly. history.jsonl already logged the discard.
    await agent(
      `The change "${h.title}" was ${decision === 'error' ? 'ERRORED' : 'DISCARDED'} by the gate (reasons: ${(gate && gate.reasons || []).join('; ') || 'n/a'}). Revert it cleanly so the tree returns to baseline.${REPO_RULES}
Steps:
1. Revert ONLY these working-tree paths: ${(proposal.files || []).join(', ')}. Use \`git restore <path>\` for tracked files; \`git rm -f --quiet\` or plain file removal for files this change newly created (check \`git status\` first).
2. If a build was run, the binaries are gitignored — leave them.
3. Do NOT touch research/eval/history.jsonl (the discard verdict stays logged) or research/eval/baseline.json (unchanged on discard).
4. Confirm \`git status\` shows the C3 source files back to their pre-change state. Report what you reverted.`,
      { label: `revert:${tag}`, phase: 'Iterate' }
    )
    discarded.push({ title: h.title, decision, reasons: gate && gate.reasons, run_path: gate && gate.run_path })
    log(`[${tag}] ${decision.toUpperCase()} — reverted`)
  }
}

return {
  hypotheses_tried: kept.length + discarded.length,
  kept,
  discarded,
  stopped_on: consecutiveDiscards >= 3 ? '3 consecutive discards' : 'hypotheses exhausted or budget',
  note: 'Baseline, history.jsonl, and learnings are committed for kept changes. Run again to continue improving.',
}
