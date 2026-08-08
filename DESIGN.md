# Coffe Terminal Design System

## 0. Research Log
- Embedded refs: shortlisted Linear, Warp, and Sentry; picked operational taste plus Linear precision for hierarchy.
- Lazyweb: skipped because the v0.1 surface is a terminal application, not a browser screen.
- Imagen drafts: skipped because raster concepts cannot serve as a terminal-cell fidelity contract.
- Product adaptation: replaced Linear violet with a coffee-amber focus color and terminal-safe tonal surfaces.

## 1. Atmosphere & Identity
Coffe is a quiet operations desk: dense, exact, and calm under pressure. Its signature is an amber focus rail that marks the current project, target, and approval boundary without decorating inactive data.

Design read: keyboard-first operational dashboard for one developer, with restrained Linear-like precision. `DESIGN_VARIANCE=3`, `MOTION_INTENSITY=2`, `VISUAL_DENSITY=8`.

## 2. Color

| Role | Token | Light | Dark | Usage |
|------|-------|-------|------|-------|
| Canvas | `canvas` | `#F5F1E8` | `#0D0F10` | Terminal background |
| Panel | `panel` | `#FFFDF7` | `#151719` | Pane surface |
| Elevated | `elevated` | `#EDE7DA` | `#1D2023` | Selected rows and input |
| Text | `text` | `#241F1A` | `#ECE7DE` | Primary copy |
| Muted | `muted` | `#6F675E` | `#8E9399` | Metadata and hints |
| Border | `border` | `#CFC5B5` | `#30343A` | Pane separation |
| Focus | `focus` | `#9A5B13` | `#D99A45` | Active pane and approval |
| Success | `success` | `#287A4A` | `#55B983` | Completed work |
| Warning | `warning` | `#9A5B13` | `#D99A45` | Pending approval |
| Error | `error` | `#B1342E` | `#E06C65` | Failed work |

Rules:
- Focus color is reserved for selection and approval boundaries.
- Status color always accompanies a text label; color is never the only signal.
- No raw color appears in TUI files outside the style token module.

## 3. Typography
The terminal controls font choice. Coffe uses terminal monospace with three semantic levels: title (bold), label (bold or muted), and body (normal). Technical identifiers remain unmodified and may truncate with an ellipsis when their pane is narrow.

## 4. Spacing & Layout
The base unit is one terminal cell. Pane padding uses one horizontal cell; pane gaps use one cell. At 100 columns and above, the screen is a 40/60 split with Projects/Targets and Prompt above Jobs and Logs. Below 100 columns, panes become a single vertical stack. The root owns terminal height; log and list regions own their vertical overflow.

## 5. Components

### Pane
- Structure: title, optional status, bounded body.
- States: inactive border, focused amber border, empty message, error message.
- Accessibility: title and focus are expressed in text as well as color.

### Selectable Row
- Structure: focus rail, primary label, secondary metadata.
- States: default, selected, unavailable.
- Accessibility: selected rows start with `>`.

### Prompt Editor
- Structure: multiline input and one-line key hint.
- States: idle, focused, planning, error.
- Accessibility: real terminal cursor is exposed for IME.
- Grammar: deployment prose, `git status|pull|push|approve|execute`, `agent TOOL PROMPT`, or `sql QUERY` against the first registered database.

### Action Plan
- Structure: target summary, ordered allowlisted commands, approval status.
- States: draft, approved, running, succeeded, failed.
- Accessibility: state names are always visible; approval key is never overloaded.

### Event Log
- Structure: timestamp, event kind, message.
- States: empty, streaming, terminal status.
- Accessibility: chronological text remains readable without color.

### Operations Summary
- Structure: deployment plans followed by Git operations, agent sessions, database connections, Swarm counts, worker/plugin counts, AI provider/usage counts, policy/audit counts, and the latest query row.
- States: empty per section, running, failed, succeeded, stopped.
- Accessibility: every status is rendered as text and clipped to its pane without breaking ANSI sequences.

## 6. Motion & Interaction
No decorative animation. A spinner may indicate a real in-flight request. `Tab` changes pane, arrows change selection, `Enter` runs the prompt grammar, `a` approves, `x` executes, `c` cancels, `y` retries, `b` previews rollback, `s` inspects Swarm, and `q` exits when the prompt is not editing.

## 7. Depth & Surface
Use a borders-only strategy. Pane hierarchy comes from border luminance and the selected-row background; shadows and simulated glow are forbidden.

## 8. Accessibility Constraints & Accepted Debt

Constraints:
- Full keyboard operation and visible focus.
- CJK-safe width calculation and real cursor positioning.
- Status never relies on color alone.
- Minimum supported terminal is 80x24; narrower terminals show a resize message.

Accepted debt:

| Item | Location | Why accepted | Owner / Exit |
|------|----------|--------------|--------------|
| Screen reader behavior varies by terminal | TUI | Terminal accessibility APIs are not standardized | Reassess with desktop client |
