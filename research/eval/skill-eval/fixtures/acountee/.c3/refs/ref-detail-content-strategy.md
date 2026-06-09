---
id: ref-detail-content-strategy
c3-seal: 135b98a97914fdbef0820249836f85d119f709cdacfa15b365bb59765161ae40
title: Detail View Content Strategy
type: ref
goal: Consistent, scannable detail views across all entity screens. Content aligns to a 2-column grid, values take natural width, vertical rhythm from section borders. Finance-ledger feel.
---

# Detail View Content Strategy

## Goal

Consistent, scannable detail views across all entity screens. Content aligns to a 2-column grid, values take natural width, vertical rhythm from section borders. Finance-ledger feel.

## Choice

Standardize detail views on a dual-primitives approach: facet-style meta grids for summary fields and BIG grids for inline key-value pairs.

## Why

This preserves visual consistency across product and admin domains while keeping dense financial data readable on both desktop and mobile layouts.

## Two Class Families, Same Pattern

| Context | Prefix | Label font-size |
| --- | --- | --- |
| Feature screens (Invoice, PR) | detail-* | 0.625rem |
| Admin screens (Users, Teams, Flows) | admin-detail-* | 0.6875rem |

Both are structurally identical: 2-column grid at 640px+, stacked on mobile, label-above-value cells, even items get `padding-left: 1.5rem`. Pick based on screen context, never mix in the same view.

## CSS Class API

### Layout Classes

| Class | What it does |
| --- | --- |
| detail-section | Vertical section with padding: 0.75rem 0. Adjacent sections get a border-top. |
| detail-section-header | Section title: 0.6875rem, 600 weight, uppercase, muted. margin-bottom: 0.5rem. |
| detail-meta-grid | 2-column grid (repeat(2, 1fr)). Even children get padding-left: 1.5rem at 640px+. |
| detail-meta-item | Flex column cell: label above, value below. padding: 0.25rem 0. |
| detail-meta-label | 0.625rem, 500 weight, uppercase, muted. |
| detail-meta-value | 0.8125rem, 500 weight. |
| detail-columns | 2-column split at 640px+. Second column gets padding-left: 1.5rem. Stacks on mobile with border between. |
| detail-column | Single column within a detail-columns split. |
| detail-list-item | Row for collection items. Stacks on mobile, horizontal at 480px+. Adjacent items get border-top. |

### BIG Classes (Key-Value Pairs)

Use BIG when you have **label-value pairs in a horizontal row** (label left, value right on the same line). The facet grid puts label above value; BIG puts them side by side.

| Class | Layout |
| --- | --- |
| big-grid big-grid-2col | auto 1fr -- label left, value right (most common) |
| big-grid big-grid-3col | 1fr 1fr auto -- label, value, action |
| big-grid big-grid-4col | 1fr 1fr 1fr auto -- multi-value rows |
| big-grid big-grid-vertical | Single column stack |
| big-grid big-grid-responsive | Stacks < 640px, 4-col on desktop |

| Cell class | Purpose |
| --- | --- |
| big-cell | Base cell: padding: 0.125rem 0, flex, align-center |
| big-cell-label | 0.625rem, uppercase, muted, min-width: 4rem, padding-right: 0.75rem |
| big-cell-value | 0.8125rem, 500 weight |
| big-cell-mono | Adds font-mono, tabular-nums |
| big-cell-status | Flex with gap: 0.5rem for badge + icon combos |
| big-cell-actions | Right-aligned action buttons |

### BIG List Classes (Bordered Collection)

| Class | Purpose |
| --- | --- |
| big-list | Grid container with 1px gap, border, radius. Border-color background creates cell separators. |
| big-list-item | 1-col mobile, 4-col at 640px |
| big-list-cell | Cell with background, padding 0.625rem 0.75rem. First child bold. |
| big-list-cell-secondary | Muted, smaller text |
| big-list-cell-mono | Monospace + tabular-nums |
| big-list-cell-actions | Right-aligned |
| big-empty | Centered muted text for empty lists |

### Status Icons (Approval Progress)

| Class | Color | Use |
| --- | --- | --- |
| big-status-complete | success | Approved step |
| big-status-pending | warning | Awaiting action |
| big-status-waiting | muted | Future step |

Apply to a `big-status-icon` wrapper (20px circle with centered 12px SVG inside).

## Decision: Which Primitive When

| You have... | Use |
| --- | --- |
| 2-4 top-level summary fields (date, amount, status) | detail-meta-grid + detail-meta-item (facet strip) |
| 3+ label-value pairs describing one thing (bank details) | big-grid big-grid-2col (inline key-value) |
| Two parallel entities (From/To, Payment/PR) | detail-columns > detail-column |
| List of related entities with name + amount | detail-list-item rows |
| Bordered collection with structured cells | big-list > big-list-item |
| Free-form text (notes, addresses) | Plain elements, full width |
| Section with a title | detail-section + detail-section-header |

**Never** use facet grid for inline label-value pairs. **Never** use BIG grid for top-level summary fields. Facet = label above value; BIG = label beside value.

## Section Order

| # | Section | Layout | When |
| --- | --- | --- | --- |
| 1 | Facet strip | detail-meta-grid | Always |
| 2 | Workflow state | Full width, custom | Entity has multi-step workflow |
| 3 | Parties / relationships | detail-columns if 2 parties | Entity involves counterparties |
| 4 | Structured details | big-grid big-grid-2col | Entity has payment/config data |
| 5 | Collections | detail-list-item rows | Entity links to other entities |
| 6 | Files | detail-list-item + upload | Entity supports attachments |

Omit sections that don't apply. Never render an empty section.

## Alignment Rules

| Data type | Alignment | Font |
| --- | --- | --- |
| Labels (all contexts) | Left | Uppercase, muted |
| Text values | Left | Normal |
| Currency (facet/key-value) | Left | font-mono |
| Currency (table column) | Right | font-mono |
| Identifiers | Left | font-mono |
| Status | Left | Badge component |
| Action buttons | Right | -- |

**Values take natural width only.** Never `width: 100%` or `flex: 1` on a value.

## Empty States

| Context | Treatment |
| --- | --- |
| Optional field empty | text-xs text-muted-foreground/60, plain text |
| Collection empty | Same |
| Never | Bordered/dashed empty boxes |

List-level empty states (entire list panel empty) use `EmptyState` component (which applies `empty-page` CSS class), vertically centered. Children use `empty-page-icon`, `empty-page-title`, `empty-page-desc`, and `empty-page-action` classes. Two variants: "no data" (icon + title + desc + optional action) and "no results" (desc + action only, no icon/title).

## Tab Strategy

| Tab | Label | Contains |
| --- | --- | --- |
| Main | "General" or "Details" | Sections 1-6 above |
| Line Items | "Services" + count badge | Breakdown table (if applicable) |
| Audit | "Audit" | Change history timeline |

Count badges: rounded pill, `bg-muted`. No header identity content -- the facet grid inside the main tab serves that role.

## Golden Example: Complete Detail View

A payment request detail view demonstrating all patterns composed together:

```tsx
<Tabs defaultValue="details" className="flex flex-col h-full">
  <DetailHeader>
    <TabsList className="w-full border-b-0">
      <TabsTrigger value="details">Details</TabsTrigger>
      <TabsTrigger value="audit">Audit</TabsTrigger>
    </TabsList>
  </DetailHeader>

  <DetailContent className="pt-0 px-3 pb-3 sm:px-4 sm:pb-4">
    <TabsContent value="details">

      {/* Section 1: Facet Strip -- label above value, 2x2 grid */}
      <div className="detail-section">
        <div className="detail-meta-grid">
          <div className="detail-meta-item">
            <span className="detail-meta-label">Created by</span>
            <span className="detail-meta-value">{pr.maker}</span>
          </div>
          <div className="detail-meta-item">
            <span className="detail-meta-label">Type</span>
            <span className="detail-meta-value capitalize">{pr.type}</span>
          </div>
          <div className="detail-meta-item">
            <span className="detail-meta-label">Amount</span>
            <span className="detail-meta-value font-mono">{formatCurrency(pr.amount)}</span>
          </div>
          <div className="detail-meta-item">
            <span className="detail-meta-label">Status</span>
            <span className={badge({ variant: getStatusVariant(pr.status), size: 'sm' })}>
              {pr.status}
            </span>
          </div>
        </div>
      </div>

      {/* Section 2: Workflow -- custom layout, full width */}
      <div className="detail-section">
        <h3 className="detail-section-header">Approval Progress</h3>
        <div className="flex items-stretch gap-2 flex-wrap text-sm">
          <div className="flex items-center gap-1.5 px-2.5 py-1.5 rounded bg-success/10 border border-success/20">
            <IconCheck className="w-3 h-3 text-success" />
            <span className="text-success/80 text-xs">{pr.maker}</span>
          </div>
          <span className="flex items-center text-muted-foreground/30">&rarr;</span>
          <div className="flex items-center gap-1.5 px-2.5 py-1.5 rounded bg-warning/10 border border-warning/20">
            <span className="w-2 h-2 rounded-full bg-warning flex-shrink-0" />
            <span className="text-xs font-medium">{nextApprover}</span>
          </div>
        </div>
      </div>

      {/* Section 4: Structured Details -- BIG grid for bank details */}
      <div className="detail-section">
        <h3 className="detail-section-header">Payment</h3>
        <div className="big-grid big-grid-2col">
          <div className="big-cell big-cell-label">Bank</div>
          <div className="big-cell big-cell-value">{pr.bank_name}</div>
          <div className="big-cell big-cell-label">Account</div>
          <div className="big-cell big-cell-value big-cell-mono">{pr.bank_account}</div>
          <div className="big-cell big-cell-label">Name</div>
          <div className="big-cell big-cell-value">{pr.account_name}</div>
        </div>
        {pr.note && (
          <div className="mt-3 pt-3 border-t border-border">
            <div className="detail-meta-label mb-1">Note</div>
            <p className="text-sm">{pr.note}</p>
          </div>
        )}
      </div>

      {/* Section 5: Collection -- linked invoices */}
      <div className="detail-section">
        <h3 className="detail-section-header mb-0">
          Invoices
          <span className="ml-2 font-normal text-muted-foreground/60">
            {invoices.length} &middot; {formatCurrency(total)}
          </span>
        </h3>
        <div>
          {invoices.map(invoice => (
            <div key={invoice.id} className="detail-list-item">
              <div className="flex-1 min-w-0">
                <div className="flex items-center gap-2 flex-wrap">
                  <span className="font-medium text-sm">{invoice.supplier_name}</span>
                  <span className="text-xs text-muted-foreground/70 font-mono">#{invoice.invoice_no}</span>
                </div>
              </div>
              <div className="flex items-center gap-3">
                <span className="font-mono text-sm">{formatCurrency(invoice.amount)}</span>
                <button className={button({ variant: 'ghost', size: 'xs' })}>Unlink</button>
              </div>
            </div>
          ))}
        </div>
      </div>

      {/* Section 6: Files */}
      <div className="detail-section" style={{ borderBottom: 'none' }}>
        <h3 className="detail-section-header mb-0">Attachments</h3>
        {attachments.map((file, i) => (
          <div key={i} className="detail-list-item">
            <span className="text-sm truncate flex-1">{file}</span>
            <div className="flex items-center gap-1">
              <a className={button({ variant: 'ghost', size: 'xs' })} href={`/files/${file}`}>Download</a>
            </div>
          </div>
        ))}
      </div>

    </TabsContent>

    <TabsContent value="audit" className="pt-3">
      <AuditLogPanel tableName="pr" recordId={pr.id} />
    </TabsContent>
  </DetailContent>

  <DetailFooter>
    {/* Action buttons */}
  </DetailFooter>
</Tabs>
```

### Two-Column Parties Example (Invoice Screen)

When an entity has two parallel parties, use `detail-columns`:

```tsx
<div className="detail-section">
  <h3 className="detail-section-header">Transaction Parties</h3>
  <div className="detail-columns">
    <div className="detail-column">
      <div className="detail-meta-label mb-1.5">From (Supplier)</div>
      <div className="text-sm font-medium">{invoice.supplier_name}</div>
      <div className="text-xs text-muted-foreground/70 mt-0.5 font-mono">{invoice.supplier_tax_code}</div>
      <div className="text-xs text-muted-foreground/60 mt-0.5 leading-relaxed">{invoice.supplier_addr}</div>
    </div>
    <div className="detail-column">
      <div className="detail-meta-label mb-1.5">To (Buyer)</div>
      <div className="text-sm font-medium">{invoice.buyer_name}</div>
      <div className="text-xs text-muted-foreground/70 mt-0.5 font-mono">{invoice.buyer_tax_code}</div>
      <div className="text-xs text-muted-foreground/60 mt-0.5 leading-relaxed">{invoice.buyer_addr}</div>
    </div>
  </div>
</div>
```

Each entity block follows progressive de-emphasis: name (sm, medium) > identifier (xs, mono, muted/70) > supplementary (xs, muted/60).

### Admin Variant

Same structure, different prefix. Replace `detail-` with `admin-detail-`:

```tsx
<div className="admin-detail-section">
  <h3 className="admin-detail-section-title">Team Information</h3>
  <div className="admin-detail-grid">
    <div className="admin-detail-item">
      <span className="admin-detail-label">Name</span>
      <span className="admin-detail-value">{team.name}</span>
    </div>
    <div className="admin-detail-item">
      <span className="admin-detail-label">Created</span>
      <span className="admin-detail-value text-muted-foreground">{formatDate(team.created_at)}</span>
    </div>
  </div>
</div>
```

## Cited By

- `c3-104-invoice-screen.md`
- `c3-105-payment-requests-screen.md`
- `c3-107-admin-screens.md`
