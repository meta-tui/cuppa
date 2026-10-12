# Concept and architecture

## The idea

Designing a terminal UI today means writing Go, running it, squinting at the
result, and repeating. Cuppa closes that loop: lay the interface out visually
with real [Charm](https://charm.sh) components in mind, then save it, share it
or export a picture. It is itself a Bubble Tea app, so it uses the same stack
it designs, and when something feels wrong in Cuppa we are feeling what our
users' terminals feel.

Scope today: **design and export**. Cuppa draws faithful previews of each
component from its properties; it does not run the real component models, and
it does not yet generate Go code (both are on the roadmap).

## One engine, thin front ends

Everything that decides what a design *is* or how it can change lives in Go
libraries that know nothing about any user interface. The terminal app is the
first front end; the same libraries are meant to be driven later by a desktop
app (Wails), a local server with a tray icon, a web app and an npm package.
The rationale is in [ADR 0001](adr/0001-headless-engine-and-adapters.md).

```
                         ┌────────────────────────────┐
  apps/cuppa-tui  ─────▶ │  libs/canvas   editor      │ select, move, resize,
  (Bubble Tea)           │                 hittest    │ z-order, undo/redo,
                         │                 snap       │ properties, guides
  (future: Wails,        ├────────────────────────────┤
   server, web, npm)     │  libs/render   scene, grid │ Document ▶ styled cells
                         │  libs/export   text, image │ ANSI, plain, Freeze
                         │  libs/cuppafile format,disk│ .cuppa files
                         ├────────────────────────────┤
                         │  libs/catalog  definition, │ what components exist,
                         │                registry,   │ their properties
                         │                standard    │
                         │  libs/document design      │ Document, Node, Rect
                         └────────────────────────────┘
```

| Capability | Slice packages | Responsibility |
|---|---|---|
| `document` | `design` | The model: a canvas and nodes in z-order. No behaviour beyond its own invariants. |
| `catalog` | `definition`, `registry`, `standard` | What a component is (id, family, default size, property schema, import path, status), a lookup, and the shipped entries. |
| `layout` | `expr` | Size expressions (`50%`, `100% - 10`, `min(50%, 40)`) resolved to whole cells for a parent size. See [ADR 0006](adr/0006-responsive-layout.md). |
| `canvas` | `editor`, `hittest`, `snap` | Intent-based editing with snapshot undo/redo; hit testing and resize handles; alignment snapping. |
| `render` | `grid`, `scene` | Turns a document into a grid of styled cells. The one place that uses Lip Gloss. |
| `cuppafile` | `format`, `disk` | The binary + JSON codec with migrations; atomic file read and write. |
| `export` | `text`, `image` | ANSI and plain text; pictures through the Freeze program. |
| `archcheck` | | A test that fails the build if a library imports Bubble Tea, Bubbles, an app, or (outside `render`) Lip Gloss. |

Why this shape:

- **Exports match the screen.** The canvas and every export call the same `scene.Render`.
- **Intents, not events.** The editor exposes commands such as "move the selection by (dx, dy)". They are plain data, so a web socket could carry them as easily as a mouse handler calls them.
- **The catalog is data.** A component is a `Definition` value plus a painter function; the palette, inspector and renderer all read the same entry.

## The terminal app

`apps/cuppa-tui` is split by what the user does, each slice a Go package:

| Slice | Role |
|---|---|
| `shell` | Layout, routing mouse and keys, palette-to-canvas drag, status line. The only slice that knows the others. |
| `palette`, `stage`, `inspector` | The three panes. |
| `menubar` | The bar and its dropdowns; reports which action was chosen. |
| `modal`, `confirm`, `filedialog` | The dialog contract and the two dialogs. |
| `fileflow` | The file journeys (New, Open, Save, exports) and the current path. |
| `pointer` | Terminal-independent mouse events, so panes never see Bubble Tea types. |
| `theme` | Colours and small text styles every pane shares. |

Mouse events arrive in the shell as Bubble Tea messages, become `pointer.Event`s,
and go to whichever dialog, menu or pane should receive them. A pane that
started a drag keeps receiving events until the button is released, even if the
pointer leaves it.

## The desktop app

`apps/cuppa-desktop` is not a second UI. It is a Wails window whose one
component, TReactUI's `<TTY>`, shows this same terminal app over Wails events
from the same process ([ADR 0004](adr/0004-desktop-app-wails-and-treactui.md)).
Anything added to the catalog or the shell appears in both front ends.

## Code layout rules

Code is organised in vertical slices rather than technical layers:
`libs/<capability>/<slice>/` is one Go package per outcome, and file names say
their role (`move_use_case.go`, `rect_model.go`, `props_contract.go`). There
are no `util`, `common` or `shared` packages. See [`CLAUDE.md`](../CLAUDE.md).

## Build, release, CI

The repository is an Nx workspace generated by `mnci`: one Go module per
project and a root `go.work`. CI runs through `npx mnci ci`, `nx release`
versions from conventional commits and creates a GitHub Release with zips for
Windows, macOS and Linux. Findings from setting this up are in
[ADR 0002](adr/0002-mnci-go-findings.md).

## Roadmap shape

Because behaviour lives in libraries, most roadmap items add a front end or a
library, not a rewrite: `codegen` (design to Go source) beside `export`; a
server wrapping `canvas`; a Wails app or web app as another adapter.
