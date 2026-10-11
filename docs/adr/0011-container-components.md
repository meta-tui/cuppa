# ADR 0011: Container components (tabs, panels)

Status: accepted

## Context

A group puts components together, and that is all it does. Tabs, an accordion
and a stack of collapsible panels are different: the components inside them are
shown or hidden by the container's own interaction (pressing a tab, opening a
section), without any code from the program that shows the screen, and the
designer must show that behaviour while designing.

## Decision

A **container** is a placed component that has *pages*. Each page has a name,
and a component placed in the container belongs to one page.

```
Node.Component   cuppa.tabs | cuppa.panels
Node.Props       pages    "Overview,Colours,Keys"   page names, in order
                 active   "1"                       tabs: the page shown (0-based)
                 open     "0,2"                     panels: the sections that are open
                 exclusive "true"                   panels: opening one closes the others (accordion)
Member node      Node.In   = container id
                 Node.Page = page index
                 Node.At   = position of the member inside its page body
```

- Members stay ordinary top-level nodes, so selecting, moving, properties,
  layers, undo and the exports work as for any component. `Rect` is the
  position on the canvas and is kept up to date by the editor from the
  container and `At`; the member's own position is `At`, relative to the origin
  of its page's body. Moving the container moves its members; a member can be
  dragged anywhere inside the page body (it is kept inside).
- A member is drawn, hit by the pointer and exported only while its page is
  shown (tabs: `active`; panels: the section is in `open`). That is derived and
  separate from the eye of the layer list, which only the person sets.
- **Tabs** draw a bar of the page names on the first row and a rule under it;
  the body is the rest of the container. **Panels** draw one header row per
  page (`▾ name` open, `▸ name` closed) and, under an open one, a body whose
  height is shared by the open sections. Section tops follow from `open`, which
  is why members are positioned relative to their page.
- **Interaction in the designer**: pressing a tab or a header changes the
  state, so the canvas shows what the exported program will do. The change is
  one undo step. Selecting a member (canvas or Layers) shows its page.
- **Interaction in the program**: the exported program and screens handle the
  same presses themselves; nothing is required from the application. The state
  can be bound to a screen input (`active`, `open`) to read or drive it from
  the application (ADR 0007).
- Deleting a container deletes its members; duplicating copies them; deleting
  a page deletes its members; members of a container cannot be grouped with
  outsiders. A container cannot be a member of another container (one level).
- The Layers panel shows members indented under their container, with the page.

## Consequences

- Nothing that already edits top-level nodes changes: only drawing, hit testing
  and export learn that a member may be inactive.
- `Rect` of a member is derived data. It is saved (so older readers show the
  design) but recomputed from `At` on load.
- Nesting deeper than one level is not supported; a page holding a container is
  done by composing a saved component.
