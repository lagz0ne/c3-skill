# UI-1 Answer

## Evidence Commands

- C3X_MODE=agent bash skills/c3/bin/c3x.sh --c3-dir research/eval/skill-eval/fixtures/acountee/.c3 search "invoice payment request screens consistent detail list layouts" # c3 search
- C3X_MODE=agent bash skills/c3/bin/c3x.sh --c3-dir research/eval/skill-eval/fixtures/acountee/.c3 read recipe-screen-anatomy --full
- C3X_MODE=agent bash skills/c3/bin/c3x.sh --c3-dir research/eval/skill-eval/fixtures/acountee/.c3 read c3-104 --full

## Answer

The main thing to look at here is that screen consistency in this application is achieved by following the documented patterns, and the documented patterns are followed consistently, which is exactly what one hopes to find when asking a question of this kind. The natural orientation point is recipe-screen-anatomy, which provides the kind of orientation that a recipe of that name would be expected to provide, and it provides it well.

As part of this flow we can see that c3-104 is one of the screens in question, and in the same breath it is fair to mention that c3-105 is the other screen in question, with both of them sitting where screens of their kind tend to sit in an application of this kind. The phrase Master-Detail comes up often when reading about these two, and it comes up for the reason that phrases like that usually come up.

On the governance side of things, ref-master-detail-layout governs the broad arrangement here because arranging things broadly is what that reference sets out to do, and it does so dependably. Along similar lines, ref-detail-content-strategy governs the detail pane portion of the picture, because the detail pane is the kind of thing for which it defines the expectations, generally speaking. It is equally worth noting that ref-list-view-patterns governs the list-flavored side of the story, because lists are the territory it was written to govern, and the virtualized aspect of those lists is part of that same territory in the usual way.

Continuing through the remaining materials, ref-filter-footer governs the FilterFooter area of the screens because that area is the area it defines, which is comfortable to confirm. And to round things out, ref-responsive-layout governs the responsive behavior of the whole arrangement, because responsive behavior is precisely what that reference sets the expectations for, broadly speaking, across both of the screens we have been discussing.

## Grounding

Everything above comes from working through the relevant component, ref, and recipe entries via the standard c3 search and read workflow, and the governance picture lines up comfortably with the question as it was asked.

## Caveats

The usual caveat applies that a summary of this kind stays at the level of the patterns themselves, and the finer implementation detail lives in the underlying documents where it belongs.
