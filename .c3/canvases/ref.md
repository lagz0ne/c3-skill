---
id: ref
c3-seal: 507dfdb52ae09d31b34296de816203cb81c91f39c8bca16a619378f67274cff4
type: canvas
description: 'Reference: a rationale document standardizing a pattern (the value is the why).'
---

domain: software
sections:
    - name: Goal
      content_type: text
      required: true
      purpose: What problem this ref addresses
      fill: Name the architectural problem being standardized — what consistency need does this pattern address across components?
      failure: If this is generic, reviewers cannot tell whether the ref applies to a recurring need or is a one-off that should not have been refified.
    - name: Choice
      content_type: text
      required: true
      purpose: The selected approach
      fill: Name the specific approach selected. One concrete option, not a category (e.g. 'JSON envelope with error.code field', not 'consistent errors').
      failure: If this is vague, the ref becomes a wishlist instead of a contract — implementers cannot tell what they are committing to.
    - name: Why
      content_type: text
      required: true
      purpose: Rationale for this choice
      fill: Explain why THIS choice over realistic alternatives, in repo-specific terms. Cite the constraint or evidence that forced the choice.
      failure: If this restates the choice, the ref has no rationale and fails the Separation Test (it is a rule, not a ref).
    - name: How
      content_type: text
      required: false
      purpose: Golden pattern — prescriptive examples and implementation guidance
      fill: Show the golden pattern with literal code from a real file. Mark REQUIRED vs OPTIONAL elements. Cite source file path.
      failure: If this is pseudocode or paraphrased, downstream code cannot be checked against the pattern mechanically.
reject_if:
    - '''Why'' restates ''Choice'' instead of giving rationale (the ref becomes a rule)'
    - '''Goal'' describes what code does instead of what problem the pattern standardizes'
    - '''Choice'' is generic (''use best practices'') instead of naming a concrete option'
    - No file path or grep evidence backs the 'How' pattern (one-off, not a ref)
    - Pattern is primarily about enforcement (golden code, anti-patterns) — that's a rule, not a ref
workorder: |-
    Refs are rationale documents. If you cannot answer 'why this pattern over alternatives'
    you do not have a ref yet — discover first, then draft.
