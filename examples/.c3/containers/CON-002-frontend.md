---
id: CON-002-frontend
title: Web Frontend Container
summary: >
  Describes the React-based web frontend architecture, component organization,
  state management, and API integration patterns. Read this to understand how
  the frontend handles user interactions, manages application state, and
  communicates with the backend API.
---

# [CON-002-frontend] Web Frontend Container

::: info Context
This container is part of [CTX-001: System Overview](../CTX-001-system-overview.md).
:::

## Overview {#con-002-overview}
<!--
High-level description of container purpose and responsibilities.
-->

The Web Frontend container provides the user interface for TaskFlow. It handles all user interactions, renders task data, and communicates with the backend via REST API.

**Responsibilities:**
- User interface rendering
- Form handling and validation
- State management for user session and tasks
- API communication with backend
- Responsive design for multiple screen sizes

## Technology Stack {#con-002-technology-stack}
<!--
Lists languages, frameworks, and key libraries used. Read to understand
the technical foundation.
-->

| Category | Technology | Version | Purpose |
|----------|-----------|---------|---------|
| Language | TypeScript | 5.3 | Type-safe development |
| Framework | React | 18.x | UI component framework |
| Build Tool | Vite | 5.x | Fast dev/build tooling |
| State | Zustand | 4.x | Lightweight state management |
| Routing | React Router | 6.x | Client-side routing |
| Forms | React Hook Form | 7.x | Form handling |
| Styling | Tailwind CSS | 3.x | Utility-first CSS |
| HTTP Client | Axios | 1.x | API requests |
| Testing | Vitest + RTL | - | Unit/component tests |

## Component Organization {#con-002-components}
<!--
Shows how components are structured inside the container.
-->

```mermaid
graph TD
    subgraph "UI Layer"
        P[Pages]
        L[Layouts]
        C[Components]
    end

    subgraph "State Layer"
        S[Stores]
        H[Hooks]
    end

    subgraph "Data Layer"
        API[API Client]
        T[Types]
    end

    P --> L
    L --> C
    P --> H
    H --> S
    H --> API
    API --> T
```

**Layers:**
- **UI Layer**: Pages, layouts, reusable components
- **State Layer**: Zustand stores, custom hooks
- **Data Layer**: API client, TypeScript types

### Directory Structure {#con-002-directory-structure}

```
src/
  pages/           # Route-level components
    Dashboard.tsx
    TaskList.tsx
    TaskDetail.tsx
    Login.tsx
  layouts/         # Page layouts
    MainLayout.tsx
    AuthLayout.tsx
  components/      # Reusable UI components
    Task/
    Form/
    Common/
  hooks/           # Custom React hooks
    useAuth.ts
    useTasks.ts
  stores/          # Zustand stores
    authStore.ts
    taskStore.ts
  api/             # API client
    client.ts
    tasks.ts
    auth.ts
  types/           # TypeScript types
    task.ts
    user.ts
```

### Key Components

| Component | Location | Description |
|-----------|----------|-------------|
| [COM-004-api-client](../components/frontend/COM-004-api-client.md) | `src/api/client.ts` | Axios-based API client |
| TaskCard | `src/components/Task/` | Task display component |
| TaskForm | `src/components/Form/` | Task creation/editing |
| AuthProvider | `src/hooks/useAuth.ts` | Authentication context |

## State Management {#con-002-state}
<!--
How application state is organized and managed.
-->

### Store Organization {#con-002-stores}

```mermaid
graph LR
    subgraph "Global State"
        AS[Auth Store]
        TS[Task Store]
        US[UI Store]
    end

    subgraph "Server State"
        RQ[API Cache]
    end

    AS -->|User session| C[Components]
    TS -->|Task data| C
    US -->|UI state| C
    RQ -->|Cached data| C
```

**Store Responsibilities:**

| Store | Purpose | Persistence |
|-------|---------|-------------|
| Auth Store | User session, tokens | localStorage |
| Task Store | Task list, filters | None (fetched) |
| UI Store | Sidebar state, theme | localStorage |

### Example Store {#con-002-store-example}

```typescript
// src/stores/authStore.ts
import { create } from 'zustand';
import { persist } from 'zustand/middleware';

interface AuthState {
  user: User | null;
  isAuthenticated: boolean;
  setUser: (user: User | null) => void;
  logout: () => void;
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      user: null,
      isAuthenticated: false,
      setUser: (user) => set({ user, isAuthenticated: !!user }),
      logout: () => set({ user: null, isAuthenticated: false }),
    }),
    { name: 'auth-storage' }
  )
);
```

## Routing {#con-002-routing}
<!--
Client-side routing configuration.
-->

### Route Structure {#con-002-routes}

| Path | Component | Auth Required | Description |
|------|-----------|---------------|-------------|
| `/` | Dashboard | Yes | Main dashboard |
| `/tasks` | TaskList | Yes | All tasks view |
| `/tasks/:id` | TaskDetail | Yes | Single task view |
| `/tasks/new` | TaskForm | Yes | Create new task |
| `/login` | Login | No | Authentication |
| `/register` | Register | No | User registration |

### Protected Routes {#con-002-protected-routes}

```typescript
// src/components/ProtectedRoute.tsx
function ProtectedRoute({ children }: { children: ReactNode }) {
  const { isAuthenticated } = useAuthStore();
  const location = useLocation();

  if (!isAuthenticated) {
    return <Navigate to="/login" state={{ from: location }} replace />;
  }

  return children;
}
```

## Communication Patterns {#con-002-communication}
<!--
How this container communicates with backend.
-->

### To Backend {#con-002-to-backend}

- **Protocol**: REST API over HTTPS
- **Client**: Axios via [COM-004-api-client](../components/frontend/COM-004-api-client.md)
- **Authentication**: JWT in Authorization header

```mermaid
sequenceDiagram
    participant U as User Action
    participant C as Component
    participant H as useTask Hook
    participant A as API Client
    participant B as Backend

    U->>C: Click "Create Task"
    C->>H: createTask(data)
    H->>A: POST /api/v1/tasks
    A->>B: HTTP Request
    B-->>A: 201 Created
    A-->>H: Task object
    H->>H: Update store
    H-->>C: Re-render
    C-->>U: Show new task
```

### Error Handling {#con-002-error-handling}

API errors are caught and transformed:

```typescript
// Centralized error handling
api.interceptors.response.use(
  response => response,
  error => {
    if (error.response?.status === 401) {
      useAuthStore.getState().logout();
      window.location.href = '/login';
    }
    return Promise.reject(error);
  }
);
```

## Configuration {#con-002-configuration}
<!--
Environment-based configuration for this container.
-->

| Variable | Dev Default | Production | Description |
|----------|-------------|------------|-------------|
| `VITE_API_URL` | `http://localhost:3000` | `https://api.taskflow.app` | Backend API URL |
| `VITE_APP_NAME` | `TaskFlow Dev` | `TaskFlow` | Application name |
| `VITE_ENABLE_DEVTOOLS` | `true` | `false` | React DevTools |

## Build & Deployment {#con-002-deployment}
<!--
Container-specific deployment characteristics.
-->

**Build Output:**
- Static assets (HTML, CSS, JS)
- Deployed to CDN or static hosting
- No server runtime required

**Build Process:**
```bash
npm run build
# Outputs to dist/
```

**Deployment Options:**
- Vercel (recommended for simplicity)
- CloudFront + S3
- Nginx static serving

**Characteristics:**
- Asset hashing for cache busting
- Gzip compression
- Lazy loading for routes

## Related {#con-002-related}

- [CTX-001: System Overview](../CTX-001-system-overview.md)
- [CON-001: Backend Container](./CON-001-backend.md)
- [COM-004: API Client](../components/frontend/COM-004-api-client.md)
