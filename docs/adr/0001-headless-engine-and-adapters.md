# ADR 0001: Headless engine with thin front-end adapters

Status: accepted

## Context

Cuppa starts as a Bubble Tea terminal app, but is expected to grow a Wails
desktop app, an HTTP/WebSocket server with a tray launcher, a web app and an
npm package. If design behaviour lived in the TUI, each of those would have to
re-implement it.

## Decision

- All design behaviour lives in UI-agnostic Go libraries under `libs/`:
  `document` (model), `catalog` (component definitions), `canvas` (editing
  engine), `render` (document to styled cell grid), `cuppafile` (`.cuppa`
  codec), `export` (Freeze and text export).
- Libraries never import Bubble Tea, Bubbles or an app. Only `render` imports
  Lip Gloss, and only to turn a cell grid into an ANSI string. This is enforced
  by `libs/archcheck`, which runs with the normal `nx test`.
- The editing API is intent based ("move node N to x,y") and works on plain
  structs, so any front end, or a network transport, can drive it.
- Front ends live under `apps/` (and `packages/` for npm) and are adapters:
  translate input into engine intents, draw engine state.
- Code is organised in vertical slices (see `CLAUDE.md`): `libs/<capability>/<slice>/`
  is one Go package per outcome, with files named `<what>_<role>.go`.

## Module layout

mnci generates one Go module per project plus a root `go.work` (confirmed in
issue #5, contrary to mnci's `mnci-details.md` which describes a single root
`go.mod`). Module paths are `github.com/meta-ui-labs/cuppa/<dir>`.

## Consequences

- A new front end only needs input mapping and drawing.
- Rendering used for export and for the live canvas is the same code, so
  exported images match what the designer shows.
- Some duplication of "state to view" logic per front end is accepted.
