---
id: adr-20260303-pdf-inline-preview
c3-seal: 9d04c04887ab399a7b00aa243e6c78da265d2b4a7fc30d0040a90f86f754d38a
title: Inline PDF Preview in Dialog Overlay
type: adr
goal: Replace download-only PDF attachments with an inline preview dialog that renders PDFs cross-browser (including mobile), so users never leave the app to view a document.
status: implemented
date: "2026-03-03"
affects:
    - c3-1
    - c3-103
    - c3-105
approved-files:
    - apps/start/test/ui/pdf-preview.test.ts
    - apps/start/src/lib/pdfPreviewUtils.ts
    - apps/start/src/components/PDFPreviewDialog.tsx
    - apps/start/src/routes/files.$hash.tsx
    - apps/start/src/screens/PaymentRequestsScreen.tsx
base-commit: ccd84ff490cd92ef21dfdd8ead558a0ef8dc75be
---

# Inline PDF Preview in Dialog Overlay

## Goal

Replace download-only PDF attachments with an inline preview dialog that renders PDFs cross-browser (including mobile), so users never leave the app to view a document.

## Problem

PDF attachments on PR detail (and future file-bearing screens) force a download. Users must download, find the file, open it in an external app, then return to the browser. This breaks flow — especially on mobile where file management is clunky.

## Decision

Add an inline PDF preview that renders PDFs directly in a dialog overlay, with full viewer controls (page nav, zoom, download fallback). Must work on desktop and mobile browsers.

### Library: react-pdf (wraps pdfjs-dist)

**Why not native `<iframe>`/`<embed>`?** iOS Safari does not render PDFs inline — it shows a single-page preview or triggers download. Android support is inconsistent across OEMs. `pdfjs-dist` canvas rendering works everywhere.

**Why `react-pdf` over raw `pdfjs-dist`?** React-friendly API, handles worker setup, page rendering, and loading states. Less boilerplate for the same result.

### UX

- Attachment rows gain a **Preview** button (PDF files only; non-PDF files keep Download only)
- Click Preview → full-screen dialog with:
Page indicator: "Page 3 of 12"

Prev/Next page buttons

Zoom in/out (pinch-to-zoom on mobile via CSS `touch-action`)

Download button (fallback to current behavior)

Close button

- On mobile: dialog renders as full-screen sheet

### Backend Change

- `files.$hash.tsx` route: serve PDFs with `Content-Disposition: inline` (other types stay `attachment`)
- This allows the `react-pdf` component to fetch the PDF URL directly

## Work Breakdown

| # | Task | Files | Layer |
| --- | --- | --- | --- |
| 1 | Install react-pdf + configure worker | package.json, vite config | Foundation |
| 2 | Create PDFPreviewDialog component | src/components/PDFPreviewDialog.tsx | UI (c3-103) |
| 3 | Update file route to serve PDF inline | src/routes/files.$hash.tsx | Backend |
| 4 | Add Preview button to PR attachment list | src/screens/PaymentRequestsScreen.tsx | Feature (c3-105) |
| 5 | Add Preview button to PR edit drawer | src/screens/PaymentRequestsScreen.tsx | Feature (c3-105) |
| 6 | Mobile responsive testing | — | Verification |

## Risks

| Risk | Mitigation |
| --- | --- |
| pdfjs-dist bundle size (~2.5MB) | Lazy-load the dialog + worker; only downloaded when user clicks Preview |
| Worker file serving in Vite | Use pdfjs-dist/build/pdf.worker.min.mjs with Vite's worker import or CDN fallback |
| Encrypted/protected PDFs | Graceful error state in dialog with download fallback |

## Affected Components

- **c3-103** (UI Kit) — new `PDFPreviewDialog` component
- **c3-105** (PaymentRequestsScreen) — attachment section UI change
- **ref-file-handling** — inline serving mode for PDFs
- **ref-detail-content-strategy** — Section 6 (Files) gains preview behavior
- **ref-responsive-layout** — dialog must be mobile-friendly

## Context

N.A - historical ADR; original context is captured in the git log around the ADR date and in the current code that implements the decision.

## Underlay C3 Changes

| Underlay area | Exact C3 change | Verification evidence |
| --- | --- | --- |
| N.A - historical | Current .c3 entities, refs, and code-map are the post-change state. | c3x verify and c3x check |

## Enforcement Surfaces

| Surface | Behavior | Evidence |
| --- | --- | --- |
| N.A - historical | Enforcement is implicit in the currently linked components and refs. | c3x graph and cited ref ids on the relevant components |

## Alternatives Considered

| Alternative | Rejected because |
| --- | --- |
| N.A - historical | Alternatives were considered at decision time; rationale is preserved in the original commit message or branch discussion. |

## Verification

| Check | Result |
| --- | --- |
| Merged and running in production | PASS - see git log for the merge commit |
