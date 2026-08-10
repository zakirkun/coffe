# Coffe Desktop Design System

## 0. Research Log
- Existing system: migrated the terminal dashboard's amber focus rail, strict status language, and four-region information hierarchy into a native desktop shell.
- Embedded refs: shortlisted Linear, Raycast, and Sentry; retained Linear's precise hierarchy with Sentry-like operational density.
- Product adaptation: replaced violet/neon defaults with coffee amber, parchment, graphite, and semantic green/red ramps.
- Lazyweb and Imagen: skipped because the existing TUI and shipped Coffe workflows are the concrete product reference; no marketing or image-led surface is being created.

## 1. Atmosphere & Identity
Coffe is a quiet local operations desk. It should feel precise, trustworthy, and fast under pressure rather than futuristic or decorative. The signature is an amber focus rail connecting the active project, target, command, and resulting activity.

Design read: keyboard-friendly desktop control plane for one developer, with restrained operational density. `DESIGN_VARIANCE=3`, `MOTION_INTENSITY=2`, `VISUAL_DENSITY=8`.

## 2. Color

| Role | Token | Light | Dark | Usage |
|------|-------|-------|------|-------|
| Canvas | `--canvas` | `#F3EFE6` | `#0E1011` | Window background |
| Sidebar | `--sidebar` | `#E9E2D5` | `#131617` | Navigation and project rail |
| Panel | `--panel` | `#FCFAF5` | `#181B1D` | Primary work surfaces |
| Elevated | `--elevated` | `#FFFFFF` | `#202427` | Inputs, selected rows, dialogs |
| Text | `--text` | `#211D19` | `#EEE9E0` | Primary copy |
| Muted | `--muted` | `#70675D` | `#979DA1` | Metadata and hints |
| Border | `--border` | `#D3C8B7` | `#32373A` | Structural separators |
| Focus | `--focus` | `#985B16` | `#D99A45` | Active navigation and approvals |
| Focus soft | `--focus-soft` | `#F1DFC4` | `#382A1B` | Selected row surface |
| Success | `--success` | `#247548` | `#55B983` | Healthy and completed |
| Warning | `--warning` | `#9A5B13` | `#D99A45` | Draft and pending approval |
| Error | `--error` | `#B23B32` | `#E4746C` | Failed and destructive |

Rules:
- Focus amber is reserved for active selection, primary actions, and approval boundaries.
- Status color always accompanies text or an icon-independent label.
- Theme follows `prefers-color-scheme`; semantic contrast remains WCAG 2.2 AA in both modes.
- Raw colors are declared only in `src/styles/tokens.css`.

## 3. Typography

| Role | Family | Size | Weight | Usage |
|------|--------|------|--------|-------|
| Window title | Manrope Variable | 20px | 700 | Active workspace |
| Section title | Manrope Variable | 14px | 700 | Panel headings |
| Body | Manrope Variable | 13px | 450 | Default UI copy |
| Label | Manrope Variable | 11px | 650 | Navigation and metadata |
| Technical | JetBrains Mono Variable | 11px | 450 | IDs, branches, commands, logs |

Typography is bundled through `@fontsource-variable`; the desktop app has no network font dependency. Use sentence case. IDs and command output are never transformed.

## 4. Spacing & Layout

The base unit is 4px. The desktop shell uses three bounded regions:

- Navigation rail: 208px fixed, owns global section navigation.
- Project rail: 248px fixed on wide windows, owns project and target selection.
- Main work area: `minmax(0, 1fr)`, owns vertical scroll.

At widths below 960px the project rail becomes a top selector. At widths below 720px navigation becomes a horizontal scrollable bar and all content becomes one column. The minimum supported Wails window is 900x640.

Spacing scale: 4, 8, 12, 16, 20, 24, 32px. Panels use 16px internal padding, dense rows use 8px vertical padding, and major work sections use 20px gaps.

## 5. Components

### App Shell
- Structure: navigation rail, project rail, title bar, scroll-owned content area.
- States: loading, ready, global error, empty workspace.
- Accessibility: landmarks, skip-to-content link, visible keyboard focus.

### Navigation Item
- Structure: text label, optional count, amber active rail.
- States: default, hover, active, focus.
- Motion: 120ms color/background transition only.

### Panel
- Structure: heading row, optional action cluster, bounded body.
- Variants: standard, recessed log surface, approval emphasis.
- States: loading, empty, error, populated.
- Depth: tonal shift plus 1px border, no drop shadow.

### Status Badge
- Structure: status text inside a compact rounded rectangle.
- Variants: draft, approved, running, succeeded, failed, canceled, stopped.
- Accessibility: color is secondary to the visible status word.

### Action Button
- Structure: short verb label; primary, secondary, and destructive variants.
- States: default, hover, active, focus, disabled, pending.
- Rule: destructive or remote mutation actions never bypass the preview/approval state.

### Command Composer
- Structure: mode selector, multiline prompt/query input, target context, submit action.
- Modes: deploy, agent, SQL, Git.
- States: idle, focused, submitting, validation error.
- Keyboard: `Ctrl+Enter` submits; plain Enter creates a new line.

### Data Table
- Structure: sticky header, bounded horizontal scroll, monospace technical cells.
- States: loading, empty, truncated, error.
- Rules: row caps and truncation are always visible.

### Activity Feed
- Structure: timestamp, category, action/status, message.
- States: empty, live refresh, error.
- Accessibility: chronological ordering and plain-language outcomes.

### Form Field
- Structure: visible label, control, optional helper text, inline validation message.
- Variants: text, password, number, select, checkbox, multiline.
- States: idle, focused, disabled, invalid, submitting.
- Rule: placeholders never replace labels; secret values are never redisplayed after submission.

### Native Path Field
- Structure: read-only selected path, browse action, clear action when optional.
- Variants: directory, input file, exclusive output file.
- States: empty, selected, invalid, dialog canceled.
- Rule: filesystem paths come from native Wails dialogs and remain monospace.

### Confirmation Panel
- Structure: consequence statement, explicit confirmation checkbox, destructive action.
- States: unconfirmed, confirmed, pending, failed, complete.
- Rule: backup import and equivalent state replacement cannot run until explicitly confirmed.

## 6. Motion & Interaction

Motion only communicates state. Hover/focus transitions use 120ms ease-out. Panel changes use an opacity transition no longer than 160ms. Running operations use a static text status or minimal spinner; there are no perpetual decorative animations. Respect `prefers-reduced-motion` by disabling non-essential transitions.

The dashboard polls local state every two seconds and refreshes immediately after an action. Polling is canceled when the component unmounts.

## 7. Depth & Surface

Use tonal shifts and borders. The canvas, sidebar, panel, and elevated surfaces form four luminance levels. Shadows and glow are forbidden. Rounded corners use one compact scale: 4px for controls and 8px for panels.

## 8. Accessibility Constraints & Accepted Debt

Constraints:
- WCAG 2.2 AA color contrast.
- Full keyboard reachability and visible `:focus-visible` rings.
- Minimum 36px interactive targets.
- Tables and logs retain readable order under zoom and narrow windows.
- Reduced motion and system color scheme are respected.
- Error messages are announced with `role="alert"`; background status uses `aria-live="polite"`.

Accepted debt:

| Item | Location | Why accepted | Owner / Exit |
|------|----------|--------------|--------------|
| Native screen-reader behavior varies by WebView2/WebKit version | Wails shell | Platform WebView accessibility differences | Verify on each native release runner |
