# Explainer Artifact Guide

Use this reference only when the user asks for a visual or shareable explanation of the reverse tornado loop.

## Constraint

The artifact has its own anti-goal:

- single HTML file
- no external runtime or network dependency
- no decorative element that does not carry meaning
- readable without the surrounding chat

## Structure

Build a self-contained HTML file with:

- the objective and target at the top
- the measured anti-goal as a visible wall or boundary
- the widening-to-narrowing discovery/execution funnel
- DKR, CKR, and PKR/task units with labels from the user's domain
- the three anti-goal eval points: admissibility, direct read, paired goal read
- the three flags: cannot, breaking, pointless
- the human-only frame boundary

Keep copy short. The artifact should explain the loop by showing the moving parts, not by embedding the full skill text.

## Implementation Notes

Use inline CSS and minimal inline JavaScript only when it clarifies interaction. Do not import fonts, icon libraries, chart libraries, or images. Prefer semantic HTML with a small SVG or CSS layout for the funnel.

Make the artifact printable and readable on mobile. Include enough labels that a viewer can understand the objective, wall, and loop without hovering.
