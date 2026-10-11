# ADR 0012: Projects, pages and shared components in the export

Status: proposed

## Context

An export today writes one package per design set: a `<screen>_screen_view.go`
and `<screen>_screen_contract.go` per screen, and the runtime and the painters
of every component the designs use as a few `cuppa_*.go` files. Inside a
screen, each placed component is a literal (`placed{Kind: "lipgloss.box", ...}`),
so a screen is data over a shared runtime, not repeated code. Still:

1. The runtime files are one flat package. An application with many screens
   cannot import "the Cuppa box" or "the Cuppa rows" as a package of its own,
   and every export of a different folder carries its own copy.
2. A design is one screen. An application is several screens with moves
   between them, and the page that is shown is state the program should be able
   to read and set, like any other input.
3. The only way back from the Go to the design is the `.cuppa` file kept
   beside it. Given only the generated folder there is no way to open it.

## Decision

### Components are packages, screens import them

```
<out>/
  cuppa.project.json            the index (below)
  cuppa/                        the runtime, once: fitEnv, grid, styles, events
  cuppa/box, cuppa/rows, ...    one package per component kind used
  screens/<screen>/             <screen>_view.go, <screen>_contract.go
  app.go                        the router: pages, Goto, the page input
```

A component package exports `Paint(g *Grid, props Props)` (and its own
interaction when it has one: the colour picker, tabs, panels). Screens import
the packages of the kinds they use and nothing else. Components a pack or
another repo ships (the colour picker, `bubble-panels`) are imported by their
module path instead of being copied.

### Pages

A project has pages; a page is one design. A component can carry a **Goto**
action (`Node.Event = "goto:Settings"`), which the router performs without
code from the application, and the router also exposes the page as a property
of the contract:

```go
type AppProps struct {
	Page string // the page shown; set it to navigate, read it to know where the user is
	...
}
```

Page selection is therefore part of the contract: the application may bind it,
read it after a Goto, or drive it. Tabs and panels (ADR 0011) work the same
way inside one page: their `active` and `open` state is an input.

### Opening an export: `main.go` is the project

The program's own `main.go` is the project file. It imports the screens, the
screens import the component packages, and the pages are registered where the
router is built:

```go
package main

import (
	"github.com/acme/app/screens/colours"
	"github.com/acme/app/screens/keys"
	"github.com/meta-tui/cuppa/run"
)

func main() {
	app := run.New(
		run.Page("Colours", colours.Screen),
		run.Page("Keys", keys.Screen),
	)
	app.Start()
}
```

`cuppa open main.go` reads it with `go/parser`: the imports and the `run.Page`
calls are the list of pages, each screen's package is loaded the same way, and
its imports are the component kinds it uses. Nothing else has to be kept in
step, and nobody maintains a second file. What the imports cannot say (which
component sits inside which, the page of a member of a container, the layer
order, positions) is in a directive comment above each placed component:

```go
//cuppa:component id="n7" kind="lipgloss.rows" parent="n3" page="0" z="4"
```

`parent` and `page` carry ADR 0011's relationships and `z` the layer. The
`.cuppa` files stay the source of truth and are listed next to each page
(`//cuppa:source designs/colours.cuppa sha="…"`), so a hand edit of generated
code is noticed by `cuppa screens --check` and the design is reopened from
the source when it is there and from the directives when it is not.

Alternatives considered: a `cuppa.project.json` index (one more file to keep
current; kept as an optional export for tools that do not read Go) and
comments only (no way to find the pages without scanning every file).

## Consequences

- The generated code stops carrying a private copy of each component; two apps
  exported separately share the package versions.
- Exports are no longer one flat package: import paths and module layout
  become part of the contract, so this needs `go.mod` handling in the export.
- Round trip from Go alone is possible but only as far as the directives go;
  code written by hand around the generated files is not read.
- Existing single-package exports keep working: the multi-file layout is an
  option of `cuppa screens` until it replaces the flat one.
