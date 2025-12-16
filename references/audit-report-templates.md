# Audit Report Templates

## Structure Audit Output

```xml
<structure_audit>
  <id_pattern_violations>
    <violation doc="path">
      <expected>pattern</expected>
      <actual>found</actual>
      <fix>action</fix>
    </violation>
  </id_pattern_violations>
  <path_violations>...</path_violations>
  <frontmatter_violations>...</frontmatter_violations>
</structure_audit>
```

## Layer Content Audit Output

```xml
<layer_content_audit>
  <context doc=".c3/README.md">
    <required_present>
      <item check="name">PASS|FAIL</item>
    </required_present>
    <violations>
      <violation type="content_too_detailed" severity="high">
        <location>Line N</location>
        <content>description</content>
        <rule>violated rule</rule>
        <fix>action</fix>
      </violation>
    </violations>
  </context>
  <container doc="path" id="c3-N">...</container>
  <component doc="path" id="c3-NNN">...</component>
</layer_content_audit>
```

## Drift Findings Output

```xml
<drift_findings>
  <finding type="phantom|technology_mismatch|undocumented" severity="critical|high|medium">
    <doc_id>c3-N</doc_id>
    <documented>what doc says</documented>
    <actual>what code shows</actual>
    <action>fix action</action>
  </finding>
</drift_findings>
```

## ADR Audit Output

```xml
<adr_audit_result adr="adr-YYYYMMDD-slug">
  <doc_changes_complete>yes|no</doc_changes_complete>
  <code_verification_pass>yes|no</code_verification_pass>
  <recommendation>READY TO MARK IMPLEMENTED | GAPS REMAIN</recommendation>
  <gaps_if_any>
    <gap type="doc|code">description</gap>
  </gaps_if_any>
</adr_audit_result>
```

## Methodology Audit Report

```markdown
# C3 Methodology Audit Report

**Date:** YYYY-MM-DD
**Target:** [path]

## Summary

| Category | Pass | Warn | Fail |
|----------|------|------|------|
| Structure | N | N | N |
| Layer Content | N | N | N |
| Diagrams | N | N | N |
| Contract Chain | N | N | N |

**Overall:** COMPLIANT / NEEDS_FIXES / NON_COMPLIANT

## Violations

[Tables of violations by type]

## Priority Actions

1. [action]
```

## Full Audit Report

```markdown
# C3 Audit Report

**Date:** YYYY-MM-DD
**Mode:** [Full | ADR | Container | Quick]

## Statistics

| Metric | Documented | In Code | Match |
|--------|------------|---------|-------|
| Containers | N | N | check |
| Components | N | N | check |

## Findings by Severity

### Critical/High/Medium

| ID | Type | Issue | Action |
|----|------|-------|--------|

## Recommended Actions

1. [action]
```
