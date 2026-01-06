---
id: c3-${N}${NN}
c3-version: 3
title: ${COMPONENT_NAME}
type: component
category: feature
parent: c3-${N}
summary: ${SUMMARY}
---
<!-- USE: Domain features (ProductCard, CheckoutFlow, UserProfile, OrderHistory) -->

# ${COMPONENT_NAME}

<!-- what this does for users -->

## Dependencies

```mermaid
graph LR
    ${COMPONENT_NAME} --> Foundation[foundation component]
    ${COMPONENT_NAME} --> Auxiliary[auxiliary pattern]
```

## Behavior

```mermaid
sequenceDiagram
    User->>+${COMPONENT_NAME}: action
    ${COMPONENT_NAME}-->>-User: result
```

## References

<!-- symbols first, then patterns, then paths -->

## Testing (if warranted)

<!-- SKIP IF: display-only, no user interactions, static content -->

<!--
trivial: skip section
simple: "Test: happy path, error states"
moderate: user flows + integration points prose
complex: + sequence for multi-step flows, state for entity lifecycle
-->
