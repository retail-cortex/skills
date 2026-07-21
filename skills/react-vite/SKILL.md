---
name: react-vite
description: Elite frontend SDLC for React 19 and Vite 6 with TypeScript. Enforces Google OAuth2 authentication (PKCE/GIS), HTTP 429 UI toast handling, strictNullChecks, paired Vitest TDD, and 85% coverage.
---

# React 19 & Vite 6 Frontend Architecture & SDLC Skill

This skill prescribes comprehensive standards for frontend single-page applications and micro-frontends with **React 19**, **Vite 6**, **TypeScript**, and **Tailwind CSS v4**, emphasizing **Google OAuth2 Integration**, **HTTP 429 Resilience**, **TDD**, **Null Safety**, and **Secure Coding**.

## Prerequisites & Pre-Flight Checklist

1. Node.js 22+ and `pnpm` installed locally.
2. Google OAuth 2.0 Web Client ID generated in Google Cloud Console.
3. Bazelisk installed for monorepo bundle orchestration (`aspect_rules_js`).

## Google OAuth2 Provider Architecture for UI Platforms

1. **Google Identity Services (GIS) & PKCE**:
   - Use `@react-oauth/google` or official Google Identity Services SDK for OAuth2 authentication.
   - Enforce Authorization Code Flow with PKCE (Proof Key for Code Exchange) for secure token exchange.
2. **Secure Token Storage Invariants**:
   - **XSS Prevention (CWE-79)**: Never store unencrypted access tokens or ID tokens in `localStorage` or `sessionStorage`.
   - Store session state in secure, encrypted, `HttpOnly`, `SameSite=Lax` cookies issued by the backend upon OAuth2 callback exchange.
3. **Authenticated API Client Interceptor**:
   - Axios or Fetch interceptors MUST attach credentials or the Bearer token to outbound requests.
   - On `401 Unauthorized` responses from backend APIs, interceptors must trigger silent OAuth2 token refresh or redirect to the Google login screen.

## HTTP 429 UI Handling & Rate Limit Resilience Invariants

- **429 Interceptor & Toast Notifications**: Axios / Fetch interceptors catch `429 Too Many Requests`, parse the `Retry-After` header, and render a toast alert.
- **Countdown & Disabled Retry Buttons**: Disable the trigger button and display a live countdown timer until the rate limit window resets.

## Defensive Error Handling & Strict Null Safety Invariants

- **TypeScript Strict Null Checks**: `tsconfig.json` MUST enforce `"strict": true` and `"strictNullChecks": true`.
- **Defensive Render Trees**: Guard state using optional chaining (`user?.profile?.name`) and nullish coalescing (`user?.displayName ?? "Anonymous"`).
- **Paired Positive, Negative & Empty TDD**: Vitest suites must assert valid data rendering, empty states (`items.length === 0`), and 401/429/500 error banners.

## Critical Coding Standards & Invariants

1. **Arrow Function React Components**: Always declare React components using arrow function syntax:
   ```typescript
   const ComponentName = ({ prop1, prop2 }: PropsType) => {
     return <div className="p-4">{prop1}</div>;
   };
   ```
2. **Next.js Async API Route Params**: In Next.js async API routes, `params` is a Promise and MUST be awaited:
   ```typescript
   export async function GET(request: Request, { params }: { params: Promise<{ id: string }> }) {
     const { id } = await params;
     return Response.json({ id });
   }
   ```
3. **Data Table Right-Aligned Actions**: For all data tables, the 'actions' column MUST be right-aligned using a flex container with `justify-end`:
   ```tsx
   <td className="p-3 text-right">
     <div className="flex items-center justify-end gap-2">
       <button className="btn-secondary">Edit</button>
     </div>
   </td>
   ```
4. **Dev Server Restart Notification**: When API routes or configuration changes require a server restart, always notify the user explicitly rather than restarting automatically.

## Security Checkpoints & CWE Invariants

- **CWE-79 (Cross-Site Scripting - XSS)**: Strictly prohibit unescaped `dangerouslySetInnerHTML`. Sanitize all user inputs before rendering.
- **CWE-1021 (Improper Restriction of Rendered UI Layers)**: Render AI agent dashboards strictly inside sandboxed `srcdoc` iframes with tight Content Security Policies (`default-src 'self'`).

## 3-Phase Execution Protocol

### Phase 1: Scaffold & Configure Google OAuth2 Provider
Wrap React root with `<GoogleOAuthProvider clientId="...">` and configure Tailwind CSS v4 design tokens.

### Phase 2: Implement Paired Component TDD Suites (85% Coverage)
Write unit tests using **Vitest** and **React Testing Library** for login flows, authenticated views, empty states, and 401/429 API simulations.

### Phase 3: Build, Verify & Publish on GitHub Pages
```bash
pnpm test --coverage
pnpm build
bazel build //apps/frontend:bundle
```

## Progressive Disclosure References

- **Vite & Tailwind Guide**: Read [`references/vite_tailwind_setup.md`](file:///Users/rmcguinness/Projects/skill-builder/skills/react-vite/references/vite_tailwind_setup.md).
- **Reference Vite Config**: View [`examples/vite.config.ts`](file:///Users/rmcguinness/Projects/skill-builder/skills/react-vite/examples/vite.config.ts).
- **Reference Component**: View [`examples/App.tsx`](file:///Users/rmcguinness/Projects/skill-builder/skills/react-vite/examples/App.tsx).
