# Reviewer Run Prompt

Review a C3 blindbox growth output. Be strict and evidence-oriented.

Use the original topic prompt, rubric, topic notes, and candidate output. Score
whether the candidate demonstrates the value of the C3 skill for grow-as-you-go
architecture documentation.

Return:

- Overall verdict: pass, borderline, or fail.
- Score table with rubric dimensions.
- Concrete gaps, with quoted/paraphrased evidence from the candidate output.
- Whether the run is useful evidence for improving the C3 skill.

Do not reward term stuffing. Passing requires concrete C3 commands, actual
containers/components/docs, and a coherent growth path through migrations and
document gardening.
