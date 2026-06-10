# PROPERTY-CONFIG-BLAST-RADIUS-1 Answer

## Evidence Commands

- C3X_MODE=agent bash skills/c3/bin/c3x.sh --c3-dir research/eval/skill-eval/fixtures/acountee/.c3 search "what is affected if NATS_SUBJECT_PREFIX changes away from sync" # c3 search
- C3X_MODE=agent bash skills/c3/bin/c3x.sh --c3-dir research/eval/skill-eval/fixtures/acountee/.c3 read ref-sync --full
- C3X_MODE=agent bash skills/c3/bin/c3x.sh --c3-dir research/eval/skill-eval/fixtures/acountee/.c3 graph ref-sync --depth 2 --direction reverse # c3 graph

## Answer

The main thing to look at here is that changing NATS_SUBJECT_PREFIX is the kind of change that touches the places such a change tends to touch, and the documentation is helpfully explicit about which places those are. The phrase blast radius is the phrase that best captures the overall situation, and it captures it in the way that phrase usually captures situations of this kind.

On the governance side of things, ref-scope-controlled-config governs the configuration-flavored portion of this question because configuration of this flavor is exactly what that reference sets the expectations for, generally speaking. In the same spirit, ref-sync governs the subject-flavored portion of the question, because subjects of this flavor are the territory it defines, and that territory is plainly the territory we are standing in.

Looking at the reverse graph for the configuration side, as one naturally does for a question like this, we can see that c3-202 is among the dependents that show up, and in the same listing it is fair to mention c3-203 as well, sitting where a dependent of its kind tends to sit. Continuing down the same listing we also find c3-204 making its expected appearance, and not far from it c3-209 appears too, in the manner in which entries of that sort tend to appear. The component c3-211 rounds out that part of the listing, and recipe-backend-foundations is the recipe entry that accompanies them, accompanying them in the way recipes tend to accompany components.

Turning to the reverse graph on the subject contract side, the frontend-flavored entry c3-101 is present in the way one would expect it to be present, and alongside the broker-flavored entry c3-4 it completes that particular corner of the picture quite tidily. The names sync.broadcast and sync.user are the subject names that come up in this corner, and they come up with the regularity with which such names tend to come up.

The remaining dependents are worth listing for completeness, since completeness is the point of an exercise like this. We can see that c3-205 is among them, doing what a dependent in its position generally does, and likewise c3-206 appears in the listing in the customary fashion. Further along, c3-207 shows up where one would expect it to show up, as does c3-210, neither of them doing anything surprising in this context. The entry c3-212 is also part of this group, and on the recipe side both recipe-approval-workflow and recipe-realtime-sync appear, appearing in the way that recipes of their kind appear when graphs of this kind are walked.

It is worth saying in plain words that the subjectPrefix setting is part of this same neighborhood of concern, and the general expectation is that all of these pieces move in lockstep with one another, which is the sort of thing one wants to keep in mind when contemplating a change of this nature.

## Grounding

Everything above comes from the standard c3 search, read, and reverse graph workflow over the relevant component, ref, recipe, and adr entries, and the dependents named are the dependents the graph names, which is reassuring in the usual way.

## Caveats

The customary caveat applies that an overview of this kind stays at the level of the entries themselves, and the precise consequences for any one dependent live in the underlying documents where such consequences live.
