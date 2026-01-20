# Container Component Patterns

## Component Categories

### Foundation
> Primitives others build on. High impact when changed. Reusable within this container.

| Component | When Needed |
|-----------|-------------|
| **Layout** | Page/screen structure |
| **Entry Point** | HTTP handler, CLI, event consumer |
| **Router** | Request routing |
| **Auth Provider** | Identity/authentication |
| **Design Tokens** | Colors, spacing, typography |
| **Base Components** | Button, Input, Card (UI primitives) |

**Categorization test:** "Would changing this break many other components?" → Yes = Foundation

---

### Feature
> Domain-specific. Uses Foundation + cites refs for patterns. Not reusable outside this context.

| Component | When Needed |
|-----------|-------------|
| **Screen/Page** | CheckoutScreen, ProductListPage |
| **Domain Component** | ProductCard, CartItem, OrderSummary |
| **Domain Hook** | useCart, useCheckout, useProductSearch |
| **Domain Service** | PaymentService, OrderService |
| **Workflow** | CheckoutFlow, OnboardingFlow |

**Categorization test:** "Is this specific to what this product DOES?" → Yes = Feature

---

## References (refs/)

> Conventions and patterns live in `.c3/refs/`, not as components.

| Pattern Type | Examples |
|--------------|----------|
| **API Client Pattern** | How we call external services |
| **State Management** | Context usage, store patterns |
| **Styling Conventions** | Tailwind/CSS patterns |
| **Error Handling** | Error boundary patterns |
| **Validation** | Input validation patterns |

**When to create a ref:** "Is this documenting HOW we use an external tool/pattern?" → Yes = ref

---

## Decision Flow

```
Is this a reusable building block that others depend on?
  → Yes: Foundation
  → No: Continue

Is this specific to this product's domain/features?
  → Yes: Feature
  → No: Continue

Is this documenting how we use an external tool/pattern?
  → Yes: Create a ref in .c3/refs/
```

---

## Examples by Container Type

### Backend Container
| Foundation | Feature | Refs Used |
|------------|---------|-----------|
| Router, Middleware | OrderService, PaymentHandler | ref-error-handling, ref-db-patterns |

### Frontend Container
| Foundation | Feature | Refs Used |
|------------|---------|-----------|
| Layout, Button, Card | ProductCard, CheckoutScreen | ref-tailwind, ref-state-management |

### Worker Container
| Foundation | Feature | Refs Used |
|------------|---------|-----------|
| Job runner, Queue consumer | EmailSender, ReportGenerator | ref-retry-patterns, ref-logging |
