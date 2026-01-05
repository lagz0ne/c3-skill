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

### Auxiliary
> Conventions for using external tools HERE. "This is how we use X in this project."

| Component | When Needed |
|-----------|-------------|
| **API Client Pattern** | How we call external services |
| **State Management** | "We use prop drilling / context / zustand like this" |
| **Styling Conventions** | "We use Tailwind like this" |
| **Type Patterns** | "Prefer type over interface" |
| **Error Handling** | "This is how we handle errors" |
| **Validation** | "This is how we validate inputs" |

**Categorization test:** "Is this documenting HOW we use an external tool/pattern?" → Yes = Auxiliary

---

### Feature
> Domain-specific. Uses Foundation + Auxiliary. Not reusable outside this context.

| Component | When Needed |
|-----------|-------------|
| **Screen/Page** | CheckoutScreen, ProductListPage |
| **Domain Component** | ProductCard, CartItem, OrderSummary |
| **Domain Hook** | useCart, useCheckout, useProductSearch |
| **Domain Service** | PaymentService, OrderService |
| **Workflow** | CheckoutFlow, OnboardingFlow |

**Categorization test:** "Is this specific to what this product DOES?" → Yes = Feature

---

## Decision Flow

```
Is this a reusable building block that others depend on?
  → Yes: Foundation
  → No: Continue

Is this documenting how we use an external tool/pattern?
  → Yes: Auxiliary
  → No: Continue

Is this specific to this product's domain/features?
  → Yes: Feature
```

---

## Examples by Container Type

### Backend Container
| Foundation | Auxiliary | Feature |
|------------|-----------|---------|
| Router, Middleware | DB access patterns, Error handling | OrderService, PaymentHandler |

### Frontend Container
| Foundation | Auxiliary | Feature |
|------------|-----------|---------|
| Layout, Button, Card | Tailwind usage, State patterns | ProductCard, CheckoutScreen |

### Worker Container
| Foundation | Auxiliary | Feature |
|------------|-----------|---------|
| Job runner, Queue consumer | Retry patterns, Logging | EmailSender, ReportGenerator |
