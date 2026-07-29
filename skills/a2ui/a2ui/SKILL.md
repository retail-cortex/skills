---
name: a2ui
description: A2UI v0.8 stable UI rendering & Gemini Enterprise integration SDLC. Covers sandboxed WebFrameSrcdoc rendering, iframe CSP policies (CWE-1021), postMessage origin validation (CWE-346), XSS defense (CWE-79), HTTP 429 UI toast handling, and component TDD.
license: Apache-2.0
metadata:
  author: Ryan McGuinness
  version: "1.0"
authors:
  - name: Retail Cortex Engineering
    url: https://github.com/retail-cortex/skills
category: a2ui
tags:
  - a2ui
  - ui
  - iframe
  - gemini
trigger_phrases:
  - "render agent UI"
  - "A2UI iframe security"
  - "postMessage origin checks"
execution_hints:
  preferred_model: "gemini-3.1-pro"
  requires_human_approval: false
  environment_variables: []
  timeout_seconds: 180
---
# A2UI Sandboxing & Gemini Enterprise Integration SDLC Skill

This skill prescribes best practices, design guidelines, and schema constraints for converting model card responses into **A2UI v0.8 stable** message payloads and rendering dynamic, branded, interactive UI components inside sandboxed iframes in Gemini Enterprise.

## Prerequisites & Pre-Flight Checklist

1. Python 3.13 and `uv` package manager installed.
2. `a2ui-agent-sdk==0.2.1` pinned in `pyproject.toml`.

## HTTP 429 Rate Limit & UI Toast Invariants

- Streaming UI renderers and client viewport containers must handle HTTP 429 quota exhaustion events and display disabled action states with active countdown timers (`Retry-After`).

## Security Checkpoints & CWE Invariants

- **CWE-1021 (Improper Restriction of Rendered UI Layers)**: UI dashboards and interactive components MUST render strictly inside sandboxed `srcdoc` iframes with restricted permissions (`sandbox="allow-scripts"`).
- **CWE-346 (Origin Validation Error)**: Restrict iframe communication strictly to `window.postMessage` with verified origin checks.
- **CWE-79 (Cross-Site Scripting)**: Disallow arbitrary script evaluation inside dynamically composed HTML templates; sanitize all user and model-generated strings.

## Gemini Enterprise A2UI Schema Constraints (v0.8 Stable ONLY)

Gemini Enterprise only supports **A2UI v0.8 stable**. All returned layout elements and components must strictly conform to v0.8 specifications.

> [!IMPORTANT]
> Never output v0.9 layout patterns. Doing so will trigger critical validation errors and render failure tracebacks in the chat window.

### 1. Layout Validation Mappings
- **Wrapper Required**: Column, Row, and Table children list containers MUST be wrapped inside the `explicitList` property:
  ```json
  "children": {
    "explicitList": ["child_id_1", "child_id_2"]
  }
  ```
- **Single-Child Cards**: The `Card` component only accepts a singular `child` string pointing to another component ID:
  ```json
  "Card": {
    "child": "column_wrapper_id"
  }
  ```
  If you have multiple children, you must wrap them in a `Column` or `Row` first, and point the card's `child` field to the container.

### 2. Right-Aligned Action Columns
In data tables and card layouts, ensure all action columns and interactive button groups are right-aligned using `justify-end` within a flex container.

### 3. Interactive Dropdown Collapsing
To prevent long vertical stacks of buttons in the chat viewport (e.g. site/store selectors):
- If the card children contain **two or more** buttons with matching actions (such as `SET_STORE`), collapse them into a single selection:
  - Use a `MultipleChoice` component with `"maxAllowedSelections": 1`.
  - Save selections to a specific path variable (e.g. `{"path": "/store_switcher/selected"}`).
  - Add a single submit `Button` below that reads the value path in its action context:
    ```json
    "Button": {
      "action": {
        "name": "SET_STORE",
        "context": [{ "key": "siteID", "value": { "path": "/store_switcher/selected" } }]
      }
    }
    ```

### 4. Markdown Code Block Interception
Model outputs containing UI Card JSON must be cleanly intercepted and parsed by the event converter.
- **Regex Robustness**: Use case-insensitive matching for code blocks (`re.IGNORECASE` | `re.DOTALL`).
- **Pattern**: `r'```(?:json)?\s*(.*?)\s*```'` to successfully parse uppercase ````JSON`, lowercase ````json`, or language-omitted ```` blocks.

### 5. Image Component Schema
Image components must bind the source URL using a `BoundValue` structure. Raw strings will fail validation.
```json
{
  "id": "image_component_id",
  "component": {
    "Image": {
      "url": {
        "literalString": "https://service-url.run.app/image.png"
      }
    }
  }
}
```

### 6. Spatial Digital Twin Blueprints
Since client-side SVGs, coordinate mapping, and custom canvas overlays are unsupported in chat frames:
- **Server-Side Render**: Draw plans and dynamic beacons as a server-side SVG image.
- **Service Route**: Serve SVGs with MIME type `image/svg+xml` from the FastAPI app (e.g., `GET /api/v1/blueprint?layout=linear&x=55&y=112`).
- **Mapping**: Transpile any `"canvas"` definitions in the model output to a standard `Image` pointing to this dynamic endpoint.

## 3-Phase Execution Protocol

### Phase 1: Payload Construction & Right-Aligned Actions
Compose declarative components, explicitList containers, and ensure all data table action columns are right-aligned using `justify-end`.

### Phase 2: Implement A2UI TDD Payload Verification
Write unit test fixtures asserting that generated HTML/CSS/JS strings and JSON schemas conform to A2UI v0.8 specifications.

### Phase 3: Emit Sandboxed WebFrameSrcdoc
Validate iframe sandboxing and execute pytest test suites:
```bash
uv run pytest tests/test_a2ui_payloads.py
```

## Progressive Disclosure References

- **A2UI Schema Guide**: Read [`references/a2ui_spec.md`](file:///Users/rmcguinness/Projects/skill-builder/skills/a2ui/references/a2ui_spec.md).
- **Reference Dashboard Composer**: View [`examples/dashboard_composer.py`](file:///Users/rmcguinness/Projects/skill-builder/skills/a2ui/examples/dashboard_composer.py).
