# A2UI Sandboxed WebFrameSrcdoc Specification

## Sandboxed HTML Structure

A2UI payloads render inside an iframe using `srcdoc`. Content must be completely self-contained (bundled JS/CSS or CDNs) and communicate with the host via standard `window.postMessage`:

```html
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <script src="https://cdn.tailwindcss.com"></script>
  <script src="https://unpkg.com/react@18/umd/react.production.min.js"></script>
  <script src="https://unpkg.com/react-dom@18/umd/react-dom.production.min.js"></script>
</head>
<body class="bg-slate-950 text-white p-6">
  <div id="root"></div>
  <script>
    // Embedded React/JS rendering code
  </script>
</body>
</html>
```

## Security & CSP Rules

- **CWE-1021 (Improper Restriction of Rendered UI Layers)**: Script execution is strictly confined to the sandboxed iframe (`sandbox="allow-scripts"`). Parent frame DOM access is prohibited.
- **CWE-346 (Origin Validation Error)**: Host interaction occurs strictly through structured JSON messages over `window.postMessage` with verified origin checks.
- **CWE-79 (Cross-Site Scripting)**: Unsanitized HTML string injection is blocked; all component values must use structured declarative JSON bindings.

## HTTP 429 Rate Limit & Toast Invariants

Client-side frames must listen for rate limit events and render active countdown banners for 429 quota exhaustion states, temporarily disabling action buttons until the backoff window expires.

## Gemini Enterprise v0.8 Schema Guidelines

- Wrap all Column, Row, and Table child arrays within `explicitList`.
- Single-child constraint on `Card` components.
- Image components require `BoundValue` with `literalString`.
- Action columns must be right-aligned using flex `justify-end`.
