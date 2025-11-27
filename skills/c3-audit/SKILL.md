---
name: c3-audit
description: Audit C3 documentation against codebase reality - find drift between docs and implementation
---

# C3 Audit

## Overview

The audit skill is used **after implementation** to verify:

1. **ADR Implementation** - Are accepted ADRs actually implemented in C3 docs?
2. **Architecture Conformance** - Does the code follow what C3 docs describe?
3. **Verification Checklists** - Are ADR verification items satisfied?

<audit_position>
When to use:
- After completing implementation work
- After accepting an ADR (to verify it was implemented)
- Periodically to catch drift
- Before major releases

This is NOT for:
- Active design work (use c3-design)
- Creating new documentation (use layer skills)
</audit_position>

**Announce at start:** "I'm using the c3-audit skill to audit documentation against codebase reality."

---

## Audit Modes

The audit can run in different modes based on what you want to verify:

| Mode | Purpose | When to Use |
|------|---------|-------------|
| **Full** | Complete audit of all docs vs code | Periodic health check |
| **ADR** | Verify specific ADR was implemented | After ADR implementation |
| **ADR-Plan** | Verify ADR ↔ Plan coherence | After c3-design, before handoff |
| **Container** | Audit single container and its components | After container changes |
| **Quick** | High-level inventory check only | Fast sanity check |

---

## Mode 1: Full Audit

### Phase 1: Load Documentation State

<chain_prompt id="load_docs">
<instruction>Load all C3 documentation and build inventory</instruction>

<action>
```bash
# Read Context
cat .c3/README.md

# List all container docs
find .c3 -name "README.md" -path "*/c3-[0-9]-*/*"

# List all component docs
find .c3 -name "c3-[0-9][0-9][0-9]-*.md"

# List all ADRs
find .c3/adr -name "adr-*.md" 2>/dev/null
```
</action>

<extract>
```xml
<doc_inventory>
  <context path=".c3/README.md">
    <containers_listed>[c3-1, c3-2, ...]</containers_listed>
    <protocols_listed>[list]</protocols_listed>
  </context>

  <containers>
    <container id="c3-{N}" path=".c3/c3-{N}-slug/README.md">
      <type>code|infrastructure</type>
      <components_listed>[c3-{N}01, c3-{N}02]</components_listed>
      <technology_documented>[stack]</technology_documented>
    </container>
  </containers>

  <adrs>
    <adr id="adr-YYYYMMDD-slug" status="proposed|accepted|implemented">
      <changes_across_layers>[what should change]</changes_across_layers>
      <verification_checklist>[items to verify]</verification_checklist>
    </adr>
  </adrs>
</doc_inventory>
```
</extract>
</chain_prompt>

### Phase 2: Explore Codebase

<chain_prompt id="explore_codebase">
<instruction>Discover actual codebase structure</instruction>

<action>
Use Task tool with subagent_type="Explore" (very thorough):

```
Explore the codebase to identify:
1. All deployable units (containers) - look for package.json, go.mod, Dockerfile, etc.
2. Major code modules within each (components) - handlers, services, repositories
3. Infrastructure from docker-compose or config
4. Technology stack of each container

Return structured inventory of what actually exists in code.
```
</action>

<extract>
```xml
<code_inventory>
  <containers>
    <container path="/api" type="code">
      <technology>Node.js 20, Express 4.x</technology>
      <components>
        <component path="/api/src/handlers"/>
        <component path="/api/src/services"/>
        <component path="/api/src/db"/>
      </components>
    </container>
  </containers>
</code_inventory>
```
</extract>
</chain_prompt>

### Phase 3: Compare & Find Drift

<extended_thinking>
<goal>Systematically compare documented vs actual state</goal>

<comparison_categories>

**1. INVENTORY DRIFT**
Doc says X exists, but code doesn't (or vice versa)

| Finding | Severity | Example |
|---------|----------|---------|
| Phantom container | Critical | c3-3 documented, no /frontend code |
| Undocumented container | High | /services/notify exists, no c3 doc |
| Phantom component | High | c3-102 documented, code removed |
| Undocumented component | Medium | New handler added, not in docs |

**2. TECHNOLOGY DRIFT**
Doc says technology X, code uses Y

| Finding | Severity | Example |
|---------|----------|---------|
| Framework mismatch | High | Doc: Express, Code: Fastify |
| Version mismatch | Medium | Doc: Node 18, Code: Node 20 |
| Dependency drift | Low | New major dependency not documented |

**3. STRUCTURE DRIFT**
Code organization doesn't match documented patterns

| Finding | Severity | Example |
|---------|----------|---------|
| Pattern violation | Medium | Doc says repository pattern, code has direct DB calls |
| Component boundary blur | Medium | Service A imports Service B internals |

**4. PROTOCOL DRIFT**
Inter-container communication differs from Context

| Finding | Severity | Example |
|---------|----------|---------|
| Protocol mismatch | High | Context: REST, Code: gRPC |
| Undocumented protocol | High | New queue added between containers |

</comparison_categories>

<output>
```xml
<drift_findings>
  <finding type="phantom" severity="critical">
    <doc_id>c3-3</doc_id>
    <expected>/frontend</expected>
    <actual>NOT FOUND</actual>
    <action>Remove doc OR implement code</action>
  </finding>

  <finding type="technology_mismatch" severity="high">
    <doc_id>c3-1</doc_id>
    <documented>Express 4.x</documented>
    <actual>Fastify 4.x</actual>
    <action>Update doc technology section</action>
  </finding>
</drift_findings>
```
</output>
</extended_thinking>

---

## Mode 2: ADR Audit

When verifying a specific ADR was implemented correctly.

### Phase 1: Load the ADR

<chain_prompt id="load_adr">
<instruction>Load ADR and extract verification requirements</instruction>

<action>
```bash
cat .c3/adr/adr-YYYYMMDD-slug.md
```
</action>

<extract>
```xml
<adr_requirements>
  <id>adr-YYYYMMDD-slug</id>
  <status>[should be accepted or implemented]</status>

  <changes_across_layers>
    <context_changes>
      <change doc="c3-0">[what should change]</change>
    </context_changes>
    <container_changes>
      <change doc="c3-N-slug">[what should change]</change>
    </container_changes>
    <component_changes>
      <change doc="c3-NNN-slug">[what should change]</change>
    </component_changes>
  </changes_across_layers>

  <verification_checklist>
    <item>[verification item 1]</item>
    <item>[verification item 2]</item>
  </verification_checklist>
</adr_requirements>
```
</extract>
</chain_prompt>

### Phase 2: Verify Document Changes

<extended_thinking>
<goal>Verify each documented change was made</goal>

<verification_process>
For each item in "Changes Across Layers":

1. Read the target document
2. Check if the documented change exists
3. Mark as: IMPLEMENTED | PARTIAL | MISSING

```xml
<doc_change_verification>
  <change doc="c3-0" expected="Add new actor: Webhook System">
    <status>IMPLEMENTED</status>
    <evidence>Found in Actors table line 45</evidence>
  </change>

  <change doc="c3-1-api" expected="Add webhook handler component">
    <status>PARTIAL</status>
    <evidence>Component listed but missing details</evidence>
    <gap>Missing interface documentation</gap>
  </change>

  <change doc="c3-102-webhook" expected="Create component doc">
    <status>MISSING</status>
    <evidence>File does not exist</evidence>
  </change>
</doc_change_verification>
```
</verification_process>
</extended_thinking>

### Phase 3: Verify Against Code

<extended_thinking>
<goal>Verify code matches what ADR and updated docs describe</goal>

<code_verification_process>
For each verification checklist item in the ADR:

1. Identify what code artifact to check
2. Explore/read the code
3. Verify it matches the documented architecture

For structural checks:
- Does the component exist in code?
- Does it follow the documented patterns?
- Are the interfaces as documented?

For behavioral checks:
- Does the flow work as documented?
- Are protocols implemented correctly?
</code_verification_process>

<output>
```xml
<code_verification>
  <item check="Is webhook handler at correct abstraction level?">
    <status>PASS</status>
    <evidence>Located at /api/src/handlers/webhook.ts, follows handler pattern</evidence>
  </item>

  <item check="Does webhook handler use documented error pattern?">
    <status>FAIL</status>
    <evidence>Uses custom error class instead of documented AppError</evidence>
    <action>Refactor to use AppError from error handling pattern</action>
  </item>

  <item check="Are downstream consumers updated?">
    <status>PASS</status>
    <evidence>NotificationService updated to receive webhook events</evidence>
  </item>
</code_verification>
```
</output>
</extended_thinking>

### Phase 4: ADR Status Recommendation

<extended_thinking>
<goal>Determine if ADR can be marked as implemented</goal>

<decision_matrix>
| Doc Changes | Code Verification | Recommendation |
|-------------|-------------------|----------------|
| All IMPLEMENTED | All PASS | Mark ADR as `implemented` |
| Some PARTIAL/MISSING | Any | Do NOT mark, list gaps |
| All IMPLEMENTED | Some FAIL | Do NOT mark, fix code issues |
</decision_matrix>

<output>
```xml
<adr_audit_result adr="adr-YYYYMMDD-slug">
  <doc_changes_complete>yes|no</doc_changes_complete>
  <code_verification_pass>yes|no</code_verification_pass>

  <recommendation>
    [READY TO MARK IMPLEMENTED | GAPS REMAIN]
  </recommendation>

  <gaps_if_any>
    <gap type="doc">[missing doc change]</gap>
    <gap type="code">[code verification failure]</gap>
  </gaps_if_any>

  <next_steps>
    1. [Fix gap 1]
    2. [Fix gap 2]
    3. Re-run audit
  </next_steps>
</adr_audit_result>
```
</output>
</extended_thinking>

---

## Mode 3: ADR-Plan Coherence Audit

Verify the mutual reference between ADR and Implementation Plan is complete and correct.

**When to use:**
- After c3-design creates an ADR+Plan
- Before handoff to ensure quality
- When reviewing someone else's ADR

### Phase 1: Load ADR and Extract Sections

<chain_prompt id="load_adr_plan">
<instruction>Load ADR and extract both ADR sections and Implementation Plan</instruction>

<action>
```bash
cat .c3/adr/adr-YYYYMMDD-slug.md
```
</action>

<extract>
```xml
<adr_plan_structure>
  <id>adr-YYYYMMDD-slug</id>

  <!-- ADR sections (medium abstraction) -->
  <changes_across_layers>
    <change layer="context" doc="c3-0">[what changes]</change>
    <change layer="container" doc="c3-N-slug">[what changes]</change>
    <change layer="component" doc="c3-NNN-slug">[what changes]</change>
  </changes_across_layers>

  <verification>
    <item>[verification item 1]</item>
    <item>[verification item 2]</item>
  </verification>

  <!-- Implementation Plan sections (low abstraction) -->
  <implementation_plan>
    <code_changes>
      <change layer_ref="c3-0" location="path/to/file" action="create|modify|delete">[details]</change>
      <change layer_ref="c3-N-slug" location="path/to/file" action="create|modify|delete">[details]</change>
    </code_changes>

    <dependencies>
      <step order="1">[first thing]</step>
      <step order="2" depends="1">[second thing]</step>
    </dependencies>

    <acceptance_criteria>
      <criterion verification_ref="item 1">[testable criterion]</criterion>
      <criterion verification_ref="item 2">[testable criterion]</criterion>
    </acceptance_criteria>
  </implementation_plan>
</adr_plan_structure>
```
</extract>
</chain_prompt>

### Phase 2: Verify Mapping Completeness

<extended_thinking>
<goal>Verify every ADR item has a corresponding Plan item</goal>

<mapping_checks>

**1. Changes Across Layers → Code Changes mapping**

For each item in "Changes Across Layers":
- Does a Code Change exist that references this layer change?
- Is the code location specific (file:function, not vague)?

| Layer Change | Has Code Change? | Location Specific? | Status |
|--------------|------------------|-------------------|--------|
| c3-0: [change] | yes/no | yes/no | PASS/FAIL |
| c3-N: [change] | yes/no | yes/no | PASS/FAIL |

**2. Verification → Acceptance Criteria mapping**

For each item in "Verification":
- Does an Acceptance Criterion exist that references this?
- Is the criterion testable (command/test, not "should work")?

| Verification Item | Has Criterion? | Testable? | Status |
|-------------------|----------------|-----------|--------|
| [item 1] | yes/no | yes/no | PASS/FAIL |
| [item 2] | yes/no | yes/no | PASS/FAIL |

**3. Orphan Detection**

Code Changes without Layer Change reference = ORPHAN (suspicious)
Acceptance Criteria without Verification reference = ORPHAN (suspicious)

</mapping_checks>

<coherence_score>
Calculate coherence:
- Layer Changes mapped: X/Y
- Verifications mapped: A/B
- Orphans found: N

COHERENT if: All mapped (X=Y, A=B) AND no orphans
PARTIAL if: Most mapped but some missing
INCOHERENT if: Major gaps or many orphans
</coherence_score>

<output>
```xml
<adr_plan_coherence adr="adr-YYYYMMDD-slug">
  <layer_to_code_mapping>
    <total_layer_changes>N</total_layer_changes>
    <mapped_to_code>M</mapped_to_code>
    <unmapped>[list of unmapped layer changes]</unmapped>
  </layer_to_code_mapping>

  <verification_to_criteria_mapping>
    <total_verifications>N</total_verifications>
    <mapped_to_criteria>M</mapped_to_criteria>
    <unmapped>[list of unmapped verifications]</unmapped>
  </verification_to_criteria_mapping>

  <orphans>
    <orphan_code_changes>[list]</orphan_code_changes>
    <orphan_criteria>[list]</orphan_criteria>
  </orphans>

  <quality_issues>
    <vague_locations>[list of non-specific code locations]</vague_locations>
    <untestable_criteria>[list of non-testable criteria]</untestable_criteria>
  </quality_issues>

  <coherence_verdict>COHERENT | PARTIAL | INCOHERENT</coherence_verdict>

  <gaps_to_fix>
    1. [specific gap to address]
    2. [specific gap to address]
  </gaps_to_fix>
</adr_plan_coherence>
```
</output>
</extended_thinking>

### Phase 3: Verify Plan Feasibility (Optional)

If coherence passes, optionally verify the Plan is feasible:

<extended_thinking>
<goal>Verify code locations in Plan actually exist or can be created</goal>

<feasibility_checks>
For each Code Change:
1. Does the target file/path exist (for modify/delete)?
2. Is the location reachable (not blocked by permissions, dependencies)?
3. Does the dependency order make sense?

For each Acceptance Criterion:
1. Can the test/command actually be run?
2. Are test dependencies available?
</feasibility_checks>

<output>
```xml
<plan_feasibility>
  <code_locations>
    <location path="src/handlers/auth.ts" action="modify">
      <exists>yes|no</exists>
      <feasible>yes|no</feasible>
      <issue_if_any>[description]</issue_if_any>
    </location>
  </code_locations>

  <dependency_order>
    <valid>yes|no</valid>
    <issue_if_any>[circular dependency, missing prerequisite, etc.]</issue_if_any>
  </dependency_order>

  <test_feasibility>
    <criterion test="npm test auth">[runnable: yes|no]</criterion>
  </test_feasibility>

  <overall_feasibility>FEASIBLE | NEEDS_ADJUSTMENT | BLOCKED</overall_feasibility>
</plan_feasibility>
```
</output>
</extended_thinking>

### ADR-Plan Audit Report

```markdown
# ADR-Plan Coherence Audit

**ADR:** adr-YYYYMMDD-slug
**Date:** YYYY-MM-DD

## Coherence Summary

| Mapping | Count | Mapped | Status |
|---------|-------|--------|--------|
| Layer Changes → Code Changes | N | M | ✓/✗ |
| Verifications → Acceptance Criteria | N | M | ✓/✗ |
| Orphans | - | N | ✓/✗ |

**Coherence Verdict:** COHERENT / PARTIAL / INCOHERENT

## Gaps Found

| Type | Item | Issue |
|------|------|-------|
| Unmapped Layer Change | c3-N: [change] | No Code Change entry |
| Unmapped Verification | [item] | No Acceptance Criterion |
| Vague Location | Code Change #3 | "auth code" not specific |
| Untestable Criterion | Criterion #2 | "should work" not testable |

## Quality Issues

- [List of vague locations that need specificity]
- [List of criteria that need testable form]

## Recommendation

[ ] **PASS** - Ready for handoff
[ ] **FIX REQUIRED** - Address gaps before handoff
[ ] **REJECT** - Major coherence issues, revise ADR+Plan

## Next Steps

1. [Specific action to fix gap 1]
2. [Specific action to fix gap 2]
```

---

## Mode 4: Container Audit

Focused audit of a single container and its components.

### Process

1. **Load container doc** and its component docs
2. **Explore container code** (just that path)
3. **Compare**:
   - Technology stack matches?
   - All documented components exist in code?
   - All code modules documented?
   - Patterns followed?
4. **Report** container-specific findings

<container_audit_output>
```xml
<container_audit container="c3-1-api">
  <technology_match>yes|no - [details]</technology_match>

  <component_coverage>
    <documented>5</documented>
    <found_in_code>4</found_in_code>
    <missing_in_code>[c3-105]</missing_in_code>
    <undocumented_code>[/src/middleware/rateLimit]</undocumented_code>
  </component_coverage>

  <pattern_compliance>
    <pattern name="error_handling">COMPLIANT|VIOLATION - [details]</pattern>
    <pattern name="data_access">COMPLIANT|VIOLATION - [details]</pattern>
  </pattern_compliance>

  <findings>
    <finding severity="medium">
      [Component c3-105 documented but code removed]
    </finding>
  </findings>
</container_audit>
```
</container_audit_output>

---

## Mode 4: Quick Audit

Fast inventory-only check.

```bash
# Just verify counts match
echo "=== Quick Audit ==="
echo "Containers in Context:"
grep -c "| c3-[0-9]" .c3/README.md

echo "Container folders:"
ls -d .c3/c3-[0-9]-*/ 2>/dev/null | wc -l

echo "ADRs by status:"
grep -h "^status:" .c3/adr/*.md 2>/dev/null | sort | uniq -c
```

---

## Audit Report Template

<report_template>
```markdown
# C3 Audit Report

**Date:** YYYY-MM-DD
**Mode:** [Full | ADR | Container | Quick]
**Target:** [all | adr-YYYYMMDD-slug | c3-N-slug]

## Executive Summary

[1-2 sentence summary]

## Statistics

| Metric | Documented | In Code | Match |
|--------|------------|---------|-------|
| Containers | N | N | ✓/✗ |
| Components | N | N | ✓/✗ |

## ADR Status

| ADR | Status | Docs Complete | Code Verified |
|-----|--------|---------------|---------------|
| adr-YYYYMMDD-slug | accepted | ✓/✗ | ✓/✗ |

## Findings by Severity

### Critical
| ID | Type | Issue | Action |
|----|------|-------|--------|

### High
| ID | Type | Issue | Action |
|----|------|-------|--------|

### Medium
| ID | Type | Issue | Action |
|----|------|-------|--------|

## Verification Checklist Results

[For ADR audits - show verification item results]

| Check | Result | Evidence |
|-------|--------|----------|
| [item] | PASS/FAIL | [details] |

## Recommended Actions

1. [Priority action 1]
2. [Priority action 2]
```
</report_template>

---

## Handoff

<chain_prompt id="handoff">
<instruction>Execute handoff based on settings or user choice</instruction>

<action>
```bash
cat .c3/settings.yaml 2>/dev/null | grep -A 5 "^audit:"
```
</action>

<handoff_options>
| Setting | Action |
|---------|--------|
| `audit: manual` | Present report only |
| `audit: tasks` | Create tracking tasks via vibe_kanban |
| `audit: fix` | Dispatch agents to fix (safe issues only) |
| Not set | Ask user preference |
</handoff_options>

<safe_to_auto_fix>
- Update technology version in docs
- Add missing component stub doc
- Update stale descriptions

NOT safe to auto-fix (require user decision):
- Remove phantom docs (might need to implement code instead)
- Structural changes
- Protocol changes
</safe_to_auto_fix>

</chain_prompt>

---

## Checklist

<verification_checklist>
**For Full Audit:**
- [ ] All C3 docs loaded
- [ ] Codebase explored
- [ ] Inventory compared
- [ ] Drift categorized by severity
- [ ] ADR statuses checked
- [ ] Report generated

**For ADR Audit:**
- [ ] ADR loaded and parsed
- [ ] Each "Changes Across Layers" item verified in docs
- [ ] Each verification checklist item checked against code
- [ ] Implementation recommendation made
- [ ] Gaps documented with next steps

**For ADR-Plan Coherence Audit:**
- [ ] ADR + Implementation Plan sections extracted
- [ ] Every "Changes Across Layers" mapped to Code Change
- [ ] Every Verification item mapped to Acceptance Criterion
- [ ] No orphan Code Changes (without Layer Change reference)
- [ ] No orphan Acceptance Criteria (without Verification reference)
- [ ] Code locations are specific (file:function, not vague)
- [ ] Acceptance Criteria are testable (command/test, not "should work")
- [ ] Coherence verdict determined
- [ ] Gaps documented with fix actions

**For Container Audit:**
- [ ] Container doc loaded
- [ ] Container code explored
- [ ] Technology match verified
- [ ] Component coverage checked
- [ ] Pattern compliance verified
</verification_checklist>

---

## Related

- [adr-template.md](../../references/adr-template.md) - ADR structure
- [v3-structure.md](../../references/v3-structure.md) - Expected doc structure
- [hierarchy-model.md](../../references/hierarchy-model.md) - C3 layer inheritance
