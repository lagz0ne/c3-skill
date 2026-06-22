# Operating Loop and Freshness

Use this reference when the goal is being run across more than one turn, on a clock, or by an
automated/delegated loop. A Reverse Tornado run is only useful while its metrics are fresh enough
to steer the next move.

## Metric Freshness Contract

For the objective, each CKR, and every anti-goal metric, record a metric contract before committing
work:

- **source of truth**: dashboard, query, API, report, or human-owned sheet
- **owner**: the person who can answer definition or data-quality questions
- **definition**: numerator, denominator, cohort, segment, attribution rule, and measurement window
- **target or threshold**: number, unit, comparator, and baseline
- **read method**: exact report name, query id, API path, or manual procedure
- **freshness rule**: `max_age`, expected update cadence, and stale-data policy
- **lag rule**: impact window before the metric can reasonably be judged
- **missing-data rule**: whether to pause, use last-known-good, or flag human review

Store both `observed_at` (when the world value was measured) and `recorded_at` (when the loop read
it). A metric can be newly recorded and still stale if `observed_at` is too old.

## Ritual Clock

Every run has a clock. It can be turn-based, time-based, or both:

- **start-of-turn**: check whether required metrics are fresh enough before choosing a move
- **pre-dispatch**: run admissibility against the latest admissible anti-goal reading or dry-run
- **post-move**: append direct objective and anti-goal readings after the worker returns
- **end-of-turn**: write status, open flags, next move candidate, and `next_check_at`
- **idle heartbeat**: when no worker finishes, still refresh metrics on schedule and review flags

Do not dispatch committing work when required metric readings are stale unless the human explicitly
waives the stale state for that move. The waiver is a flag resolution record, not a quiet override.

## Pointless Needs a Window

`pointless` is not "the metric did not move immediately." It fires only after the relevant metric's
lag rule has expired and enough fresh reads have been taken to judge the work.

Before that, mark the branch as **waiting_for_measurement** with:

- move or CKR being observed
- expected impact window
- required next reads
- earliest `pointless` review time

## Flag Lifecycle

Flags are operating states, not just notes. Each flag record should include:

- `type`: `cannot`, `breaking`, `pointless`, or `authority_drift`
- `status`: `open`, `acknowledged`, `resolved`, or `waived`
- `opened_at`, `owner`, `severity`, and `requires_human_by`
- affected objective, CKR, anti-goal, tree node, or move
- evidence: metric readings, budget spent, worker result, or rejected proposal
- resolution: decision, approver, timestamp, and linked frame revision when relevant

Blocking behavior:

- **breaking** pauses committing moves by default until resolved or explicitly waived
- **cannot** stops the affected DKR/branch until the human changes budget, scope, or direction
- **pointless** stops the affected branch after the lag window proves flat impact
- **authority_drift** stops the proposed move and escalates unauthorized scope, metric, threshold,
  approval, or action-envelope changes

## Operating Loop Output

When producing an automated or recurring run artifact, include an **Operating Loop** section with:

- cadence: turn budget, review frequency, and idle heartbeat
- `current_round`, `last_metric_read_at`, and `next_check_at`
- metric read table with source, owner, observed_at, recorded_at, max_age, and lag rule
- stale-data policy and what is currently stale, if anything
- open flags with status, owner, deadline, and blocking effect
- next admissibility check and whether it needs a dry-run propose-cost worker

