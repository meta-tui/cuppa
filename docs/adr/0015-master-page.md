# ADR 0015: A master page for what every page shows

Status: accepted; steps 1 and 3 built, step 2 (one master package in the slice
layout) not yet (#225)

## Context

A real application repeats its chrome on every page: MVD has the same Title
label on 8 pages, the same key bar on 9 and a bottom line on 4, and two pages
share a whole frame. Each copy is a node in the design and a block of generated
code, so a change to the key bar is made 9 times. Embedded user components
(ADR 0012) are packs of Go, not compositions, so they do not help.

## Proposal

A design has one **master**: a set of nodes, drawn on every page below the
page's own components.

- **Model.** `Document.Master []Node` (the nodes are ordinary nodes, with the
  same layout expressions, bindings and conditions) and `Page.Hide []NodeID`,
  the master nodes a page does not show (MVD's Download has no Title). A master
  node cannot be moved or changed from a page; it is edited on the master.
- **Inputs stay flat.** A master node binds inputs by name like any node, and
  the inputs of the master nodes a page shows are inputs of that page, under the
  same names. A page's `Props` therefore does not change (`Title`, `Hints`,
  `Status`), and two pages can still give an input different values.
- **Designer.** The pages bar's list starts with **Master**. Selecting it edits
  the master's nodes alone on the canvas. On a page, the master's nodes are
  drawn but cannot be selected; the Layers panel has a Master section with an
  eye per node, which is the page's opt-out. Preview and export show them as
  they are.
- **Export, first step.** `Document.Screens()` puts the shown master nodes,
  minus the hidden, under each page's nodes. The flat and the slice layouts
  need no change, and every layout draws exactly what the designer shows. The
  design has each piece once, which is what the people editing it need; the
  generated code still repeats it.
- **Export, second step.** The slice layout writes the master once, as a
  package under `internal/components/master`, with its parts typed on a small
  `master.Props` of the inputs it reads; each page calls it through a generated
  adapter from its own Props. A change to the key bar is then one file in the
  generated code too.
- **Events.** The events of master nodes (a click on a key bar hint) are
  raised by each page under the same names, so the page's `Handle` and its
  `Event` types do not change.

## Rules a real application needs (MVD's review)

- **Names.** Master and page nodes share one namespace per page: a layout
  expression on a page node may name a master node (`below("Title")`) and the
  other way round. A page node with the name of a master node shown on that
  page is an error in `cuppa design check`, and nothing renames it silently;
  the page hides the master node or renames its own.
- **Conditions.** A master node's condition reads inputs by name, and those
  inputs are inputs of every page that shows the node (the same rule as a bound
  property), with the sample or default value, 0 or false. A page that does not
  want the input hides the node (MVD's Download hides the key bar and keeps its
  own bottom group). `check` lists the inputs a page gains from the master.
- **Events.** A master node raises the page's own event under its own name
  (`HintEvent`, unprefixed), part of the page's `Event` interface; `Handle`
  returns it like any other.
- **Regions.** Master nodes are in the page's `Frame` and its `Layout` with
  their names, so `RegionNamed("Key bar")` works on every page that shows it.
- **Lifting.** `design merge --master <page>` matches nodes by component,
  layout and properties except bound text, prints what it lifted and what it
  left, and renames nothing.

## Not in this ADR

- Different text per page on a master node: the page has an input for it
  (`Title`), which is what a bound property is for.
- A master node that a page moves or resizes: a page that needs it hides the
  node and draws its own.
- More than one master (a "setup" chrome and a "run" chrome): one design, one
  master; a second chrome is a second design until it is needed.

## Steps

1. Model, file format, `Screens()`, editor operations (add to the master, hide
   on a page), the designer's Master entry and the Layers section. Tests: the
   export of a page equals the export of the same page with copies of the
   master's nodes.
2. Slice layout: one master package and the adapters, with the same parity
   test.
3. `cuppa design merge --master <page>` to lift the nodes MVD repeats into the
   master.

## What was built

- The master is a page of the design flagged `master`; it is not a screen.
  `Page.Hide` holds the master components a page does not show; `Screens()`
  puts the shown ones, with ids prefixed `m_`, below each page's own.
- Designer: the pages list has *Add master page* and, on a page, *Move
  selection to master*; the master is marked ◆. The canvas draws the master
  under the page, not selectable; the Layers panel has a Master section with an
  eye per component (hide on this page). The preview runs the master too.
- `cuppa design merge ... --master [--master-min N]` lifts what is the same on N
  pages or more, prints what it lifted and left, and has the other pages hide it.
  `cuppa design check` checks every page and reports a name clash between a
  page and the master.
- Exports are unchanged in shape: each page's code still draws the master's
  components itself (step 2 would write them once). Older files open as before;
  a design with no master has no `master` or `hide` fields.
- A property of a master component that is bound to an input has one value on
  the master; a page may give the input its own text (`values` on the page,
  "Master text" in the designer's canvas panels), which the page's drawing, its
  screen and its `Default()` start from. `design merge --master` records the
  text each page had.
