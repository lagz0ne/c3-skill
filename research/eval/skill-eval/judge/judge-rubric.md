# C3 Skill Eval LLM Judge Rubric

The judge is a strict independent reviewer. Score the candidate answer against
the case question, case ground truth, and this rubric. Do not reward answer
shape, keyword stuffing, or entity-id lists unless the answer is correct,
grounded, causally complete, and useful for a real engineer.

## Dimensions

Score every dimension from 1 to 5.

### 1. Correctness

Claims match the fixture and case ground truth.

| Score | Bar |
| --- | --- |
| 1 | Major wrong claim, reverses ownership/mechanism, or gives unsafe change guidance. |
| 2 | Some correct ids, but important assertions are wrong or misleading. |
| 3 | Mostly correct, with omissions or ambiguous claims that could mislead. |
| 4 | Correct on all major owners, mechanisms, refs, and risks; only minor gaps. |
| 5 | Fully correct and precise, including caveats and negative facts from the fixture. |

### 2. Trace Completeness

Answer traces the full causal chain, not just entities.

Required chain for cross-cut/property cases:

`action -> state change -> mechanism -> dependent/observer -> emergent property`

| Score | Bar |
| --- | --- |
| 1 | Lists ids or components with no usable chain. |
| 2 | Covers one or two segments but misses the mechanism or dependent side. |
| 3 | Covers a partial chain; reader can infer some missing links. |
| 4 | Covers the full chain with minor missing detail. |
| 5 | Covers the full chain explicitly and names where each link is proved. |

### 3. Reasoning Depth

Explains why/how the system behaves. For emergent-property cases, names and
explains the property, such as atomicity, blast radius, coupling, non-blocking
notifications, step-advance gating, targeted-vs-broadcast delivery, or
idempotency.

| Score | Bar |
| --- | --- |
| 1 | Pure term drop or restatement of the question. |
| 2 | Names the property but does not explain the mechanism that creates it. |
| 3 | Explains some why/how but leaves important causal assumptions implicit. |
| 4 | Explains why/how well enough to guide a change. |
| 5 | Explains mechanism, tradeoff, failure mode, and property boundary clearly. |

### 4. Grounding

Claims are backed by cited C3 reads/graphs/searches from the answer evidence.
Evidence commands alone are not enough; the prose must tie claims to the cited
docs.

| Score | Bar |
| --- | --- |
| 1 | Unsupported assertions, missing evidence, or evidence unrelated to claims. |
| 2 | Cites ids but rarely connects them to claims. |
| 3 | Grounds major claims but leaves several important claims asserted. |
| 4 | Most claims are tied to specific reads/refs/graphs. |
| 5 | Every important claim is traceable to cited C3 evidence and case ground truth. |

### 5. No Hallucination

Hallucination means ONLY:

- citing an entity id (`c3-*`, `ref-*`, `recipe-*`, `adr-*`, `rule-*`) that is
  NOT in the fixture entity inventory,
- claims that CONTRADICT the ground truth or case excerpt,
- invented guarantees, rules, or behaviors, especially whole-batch atomicity
  or guaranteed delivery.

The ground truth and excerpt are a SAMPLE of the fixture, not its entirety.
Detail beyond them — inventory-listed ids the excerpt omits, or plausible
specifics the excerpt cannot verify — is NOT hallucination; score unverifiable
claims under Grounding instead. Check each cited id against the inventory
before calling it invented.

| Score | Bar |
| --- | --- |
| 1 | Multiple invented ids/contradictions or one severe unsafe hallucination. |
| 2 | One invented id, one contradiction, or one invented guarantee. |
| 3 | No invented ids, but a claim that distorts the ground truth. |
| 4 | No invented ids, no contradictions, no invented guarantees. |
| 5 | Additionally distinguishes proved facts, inference, and caveats cleanly. |

### 6. Change Usefulness

Would this answer let an engineer safely make or assess the change?

| Score | Bar |
| --- | --- |
| 1 | Not actionable or likely to send engineer to wrong owner/mechanism. |
| 2 | Points to some docs but lacks safe change boundaries. |
| 3 | Useful orientation, but engineer must redo important trace work. |
| 4 | Gives owners, affected mechanisms, risks, and verification direction. |
| 5 | Gives a safe change map: owners, causal chain, failure modes, and checks. |

## Overall

Weighted overall score:

- Correctness: 25%
- Trace completeness: 20%
- Reasoning depth: 20%
- Grounding: 15%
- No hallucination: 10%
- Change-usefulness: 10%

Pass if:

- overall is at least 4.0,
- Correctness is at least 4,
- No hallucination is at least 4,
- no dimension is below 3.

The "good" bar is not term completeness. A good answer is correct, grounded,
causal, names and explains the emergent property, and gives enough boundary
information for a safe engineering change.
