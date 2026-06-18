# Topic Prompt: Grow Warehouse Inventory System

Start with a small C3 project for a regional retailer's warehouse inventory
system, then grow it as the system becomes more complex.

The system must support:

- Receiving, putaway, inventory lookup, reservations, picking, packing,
  shipping, cycle counts, adjustments, returns, and inventory reporting.
- Users including warehouse associates, supervisors, inventory control staff,
  customer support, and integration operators.
- Barcode scanning with manual correction and audit trails.
- Stock states including available, reserved, transferred, damaged,
  quarantined, returned, and adjusted.
- External integrations with order management, carrier systems, procurement,
  and finance reporting.

Your task:

1. Initialize or grow C3 docs in the isolated project.
2. Keep the first rung lean and complete.
3. When the system complexity requires richer structure, grow the docs by
   raising the documentation contract and migrating affected facts completely.
4. End with a project that clearly has multiple containers: frontend, backend,
   integration, and database.
5. Add components and cross-container work for meaningful feature growth.
6. Include migration and document-gardening work needed to keep the C3 docs
   coherent as the system grows.

Constraints:

- Use local C3 only: `C3X_MODE=agent bash /opt/c3/skills/c3/bin/c3x.sh`.
- Avoid codemap work unless it is absolutely required; it is not the focus of
  this eval.
- Prefer compact, concrete C3 artifacts over long prose.
- Run verification and report the exact result.
