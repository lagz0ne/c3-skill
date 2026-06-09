---
id: ref-audit-timeline
c3-seal: 94398286a226822823a5c6ffda47e1a847617644dbd1c40a07ba5925edf8c66d
title: Audit Timeline Panel
type: ref
goal: Provide a reusable, lazy-loaded timeline view of entity change history. Embedded as a tab in detail views. Designed for **business users** — shows human-readable summaries, not raw database diffs.
---

# Audit Timeline Panel

## Goal

Provide a reusable, lazy-loaded timeline view of entity change history. Embedded as a tab in detail views. Designed for **business users** — shows human-readable summaries, not raw database diffs.

## Choice

- Single `AuditLogPanel` component lazy-loaded on first tab visit via a counter-based key pattern
- Timeline layout with natural-language summaries, inline change display (top 3 visible, expand for more), and field label mapping from DB columns to user-friendly names
- Hidden technical fields, formatted values (status badges, VND currency), and priority-ordered changes (status first, then amount)

## Why

- Lazy loading avoids unnecessary API calls since audit history is an occasional reference, not a primary workflow
- Human-readable summaries let business users understand what happened without interpreting raw database diffs
- Inline change display with overflow keeps the timeline scannable while still exposing full detail on demand

## Convention

| Rule | Why |
| --- | --- |
| Use AuditLogPanel component for all entity audit views | Single implementation, consistent UX |
| Lazy-load on first tab visit | Avoid unnecessary API calls |
| Timeline layout with vertical connector line | Visual continuity of history |
| Natural-language summaries instead of raw field diffs | Users want to know what happened, not debug SQL |
| Inline change display (top 3 visible, expand for more) | Most changes are 1-3 fields; avoid unnecessary clicks |

## Component API

```tsx
<AuditLogPanel tableName="invoices" recordId={invoice.id} />
```

| Prop | Type | Purpose |
| --- | --- | --- |
| tableName | string | Database table name for audit query |
| recordId | number | Entity ID to fetch history for |

## Lazy Loading Pattern

Every screen using audit tabs follows this pattern — a counter-based key forces re-fetch on each tab visit:

```tsx
const [auditKey, setAuditKey] = useState(0)

<Tabs onValueChange={(v) => { if (v === 'audit') setAuditKey(k => k + 1) }}>
  <TabsContent value="audit" className="pt-3">
    {auditKey > 0 && <AuditLogPanel key={auditKey} tableName="..." recordId={id} />}
  </TabsContent>
</Tabs>
```

## Timeline Visual Structure

```
[vertical line running full height]
├── [colored dot] "Status changed to Approved + 2 more"
│   user@email.com · Jan 5, 2025
│   Status: [Draft badge] → [Approved badge]
│   Amount: 1,000,000 ₫ → 1,500,000 ₫
│   Note: old note → new note
│   +2 more changes
├── [green dot] "Invoice created"
│   user@email.com · Jan 3, 2025
```

## Field Label Mapping

Raw DB column names are mapped to user-friendly labels per table:

| Table | Field | Display Label |
| --- | --- | --- |
| invoices | invoice_no | Invoice Number |
| invoices | supplier_name | Supplier |
| invoices | supplier_addr | Supplier Address |
| invoices | amount | Amount |
| invoice_services | name | Service Name |
| invoice_services | quantity | Quantity |
| invoice_services | price | Unit Price |
| invoice_services | amount | Amount |
| invoice_services | tax_rate | Tax Rate |
| pr | advanced_bank_name | Bank |
| pr | attachments | Attachments |

Unmapped fields fall back to snake_case → Title Case conversion.

## Hidden Fields

Technical/internal fields are filtered out: `id`, `hash`, `maker`, `checksum`, `metadata`, `invoice_id`.

## Value Formatting

| Field Type | Formatting | Font |
| --- | --- | --- |
| Status | Human label (draft → "Draft") + colored badge | Badge variant |
| Amount / Tax | VND currency (Intl.NumberFormat) | font-mono |
| Payment Type | Human label (advanced → "Advanced") | Normal |
| Arrays | Count summary ("3 items") | Normal |

## Change Display

Changes are shown inline (no expandable table):

- **Status changes**: Old badge → arrow → new badge (color-coded per `ref-ui-patterns` status mapping)
- **Currency changes**: Monospace formatted values with arrow
- **Other changes**: Label: old value → new value
- **Priority ordering**: Status first, then amount, then name, then rest
- **Overflow**: Top 3 changes visible, "+N more changes" expander for rest

## Action Summary (Title Line)

| Action | Summary Pattern | Example |
| --- | --- | --- |
| insert | "{Entity} created" | "Invoice created" |
| delete | "{Entity} deleted" | "Payment request deleted" |
| update (status only) | "Status changed to {new}" | "Status changed to Approved" |
| update (status + more) | "Status changed to {new} + N more" | "Status changed to Approved + 2 more" |
| update (single field) | "{Label} updated" | "Amount updated" |
| update (multiple) | "{N} fields updated" | "5 fields updated" |

## Timeline Dot Colors

| Action | Dot Background | Icon Color |
| --- | --- | --- |
| insert | success/20 border success/40 | text-success |
| delete | destructive/20 border destructive/40 | text-destructive |
| update | bg-muted border border-background | text-muted-foreground |

## States

| State | Rendering |
| --- | --- |
| Loading | Skeleton shimmer (3 placeholder items) |
| Empty | Centered icon + "No audit history" / "Changes to this record will appear here." |
| Error | Centered icon + error message |
| Data | Timeline entries with inline changes |

## Applies To

- InvoiceScreen (Audit tab, `tableName="invoices"`)
- PaymentRequestsScreen (Audit tab, `tableName="pr"`)

## Cited By

- `ref-ui-patterns` (tabs, status badges, formatting)
- `ref-detail-content-strategy` (tab lazy-loading strategy)
