---
id: design-act
c3-seal: aba5629dda6e5d35392e265f95f1ec7a6c9e85bc65cb89ee66f44c839ed4412f
type: canvas
description: One act of the C3 design story — its thesis, the move it asks of the user, the integrity the tool guarantees, and why the shape is what it is.
---

domain: meta
sections:
    - name: Thesis
      content_type: text
      required: true
      purpose: The act in one sentence — the spine beat
    - name: Move
      content_type: text
      required: true
      purpose: What the user does in this act
    - name: Tool Guarantee
      content_type: table
      required: true
      purpose: What the TOOL keeps integral in this act, and how
      columns:
        - name: Guarantee
          type: text
        - name: Mechanism
          type: text
        - name: Source
          type: text
    - name: Why This Shape
      content_type: text
      required: true
      purpose: Why this act is shaped this way and not the obvious alternative
    - name: Surfaces
      content_type: table
      required: true
      purpose: Which doc surface teaches this act
      columns:
        - name: Surface
          type: text
        - name: Owns
          type: text
reject_if:
    - Thesis is longer than one sentence
    - A Tool Guarantee row names no Source
    - Why This Shape restates the Thesis instead of giving the rationale
workorder: ""
