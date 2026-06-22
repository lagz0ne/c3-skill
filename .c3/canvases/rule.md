---
id: rule
c3-seal: 026581aa03cc013ade480ff5d48074e1a6240737e80338fb0c15d8e92ec6bc9f
type: canvas
description: 'Rule: an enforceable coding standard with a literal golden example.'
---

domain: software
sections:
    - name: Goal
      content_type: text
      required: true
      purpose: What standard this rule enforces
      fill: State the standard being enforced — what must hold across all uses, not what one component does.
      failure: If this describes a single component instead of a project-wide standard, the rule has no breadth and should be inline guidance instead.
    - name: Rule
      content_type: text
      required: true
      purpose: One-line statement of what must be true
      fill: 'One-line, present-tense, enforceable. Pattern: ''All <X> must <Y>.'' or ''<X> never <Y>.'''
      failure: If this is aspirational, multi-clause, or derivable only by reading Golden Example, the rule cannot be checked at compliance time.
    - name: Golden Example
      content_type: text
      required: true
      purpose: Canonical code showing the correct pattern
      fill: Literal code copied from a real file in this codebase. Annotate `// REQUIRED` vs `// OPTIONAL` for each structural element. Include file path.
      failure: If this is paraphrased, pseudocode, or invented to fit, compliance becomes interpretive and the rule loses enforcement power.
    - name: Not This
      content_type: table
      required: false
      purpose: Anti-patterns with why they're wrong here
      columns:
        - name: Anti-Pattern
          type: text
        - name: Correct
          type: text
        - name: Why Wrong Here
          type: text
    - name: Scope
      content_type: text
      required: false
      purpose: Where this rule applies and doesn't
    - name: Override
      content_type: text
      required: false
      purpose: How to deviate from this rule when justified
reject_if:
    - '''Golden Example'' is paraphrased instead of literal code copied from a real file'
    - '''Rule'' is multi-clause or aspirational (''should generally'') instead of one-line enforceable'
    - No 1-3 YES/NO compliance question can be derived from 'Rule' + 'Golden Example'
    - Rule is primarily about rationale (why this approach) — that's a ref, not a rule
    - '''Goal'' describes a single component instead of a project-wide standard'
workorder: |-
    Rules are enforceable standards. Find the canonical code in the codebase FIRST.
    If no real example exists, the rule is premature — author the first instance, then extract the rule.
