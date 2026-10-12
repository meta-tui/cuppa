# ADR 0017: Cuppa is a design tool, not a development tool

Status: accepted (#271)

## Context

Over a few releases Cuppa grew from a place to draw terminal screens into something that also
tries to be the front half of an application: screens had typed inputs and events, a contract, a
generated Go source, a live connection to a running program (ADR 0016) and a pane that built and
ran that program next to the editor. Each step had a reason, and together they made the tool hard
to understand: the person who uses it could no longer tell what it was for, or how to start.

The one application that adopted it (MVD) needed per-screen glue in the end, so the promise that
"nothing is written in the application" was not met either.

## Decision

Cuppa is a design tool. It draws screens and exports the drawing.

- **Kept**: all components and component packs, pages and the master page, responsive size
  expressions (ADR 0006), themes, embedded components, undo and redo, the preview of Bubbles
  components, and exporting a drawing as an image, ANSI or text.
- **Removed**: the live connector and the run pane (ADR 0016, `libs/connector`, `libs/runhost`,
  `--connect`, `cuppa run`), the Go export and the contract (ADR 0007, 0013, `cuppa screens`,
  `cuppa design contract`), and the variables of a screen: bindings, show-if, events, screen keys,
  samples, app defaults, per-page values of the master and layout from inputs (ADR 0010).
- Files made with earlier versions still open. What no longer exists is ignored, and dropped the
  next time the file is saved.

## Consequences

- Cuppa's job fits in a sentence and its menus shrink to what a drawing needs.
- An application that wants a screen written in code writes it by hand, with the drawing as the
  reference.
- The code that was removed stays in git history (releases cuppa-tui@0.0.84 to 0.0.88) for anyone
  who needs it.
