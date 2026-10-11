# ADR 0014: A bar for pages under the canvas

Status: proposed (plan; nothing is built yet)

## Context

Pages (ADR 0012) are managed in a panel of the right bar. Pages are part of
the document as a whole, like the artboards of an illustration program, so they
belong with the canvas: a bar along the bottom, as the artboard navigator of
Illustrator is, and not among the properties of the selection. The right bar is
also crowded; its Undo and Redo buttons are in the Edit menu and on Ctrl+Z and
Ctrl+Y already.

## Proposal

- **The pages bar** is one row between the canvas and the status bar, the width
  of the canvas: `◂  Home ▾  ▸  [+]`. `◂` and `▸` show the previous and the next
  page, the name opens a list of the pages (a dropdown over the canvas, the same
  list as the right bar's dropdowns) to jump to one, and `[+]` adds a page. In
  the list, each page has `rename`, `move up`, `move down` and `delete`, and the
  list ends with `Add page`.
- Keyboard: `Alt+PageUp` and `Alt+PageDown` go to the previous and next page,
  `Ctrl+Shift+N` adds a page. The bar is a stop for Tab.
- While previewing, the bar shows which page the preview is on and `◂ ▸` move
  it; pages are also reachable by the components that go to them.
- The **Pages panel leaves the right bar**; *Click goes to* stays in Behaviour.
- **Undo and Redo leave the right bar**: the row is removed, its place goes to
  the Editor and Preview buttons alone. They stay in the Edit menu and on their
  keys.
- Layout: the stage loses one row (`layout.pages`), the status bar is below.
  The dropdown is drawn by the shell over the screen (as the menu bar's is), so
  it can extend over the canvas.

## How it is built

1. `layout.pages` and the bar's drawing and hit testing in a new `pagesbar`
   slice of the front end, driven by the editor's page operations (they exist).
2. The dropdown over the screen, moved from the inspector's own into a
   component the shell and the inspector share, so both draw lists the same.
3. Remove the right bar's Pages panel and the Undo and Redo row; update the
   tests that click them and the user guide.

## Consequences

- The canvas is one row shorter. Nothing in a design changes.
- A design with one page shows `Page 1 ▾ [+]`, so adding the second page is
  where the person looks for it.
