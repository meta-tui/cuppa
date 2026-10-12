# ADR 0012: Pages, a router, and opening an application from its Go

Status: accepted; the router, the Go export and opening a design from Go were removed by ADR 0017, pages stay

## Context

An export wrote one screen per design: a `<screen>_screen_view.go` and a
`<screen>_screen_contract.go` with the runtime and painters shared in a few
`cuppa_*.go` files. Each placed component is data over that shared runtime
(`placed{Kind: "lipgloss.box", ...}`), not repeated code. What was missing for
an application:

1. A design was one screen. An application is several, with moves between
   them, and the page that is shown is state the program should read and set.
2. The only way back from the Go to the design was the `.cuppa` file kept beside
   it.

## Decision

### Pages are part of the design

A design may be split into **pages**, each a screen of the application with
its own components and the same canvas, theme, keys and named colours. The page
being edited has its components in `Document.Nodes`, so every editing operation
works on it as before; the others are kept in `Document.Pages[i].Nodes`. A
design with one page has no `pages` in its file.

A component can **go to a page** (`Node.Goto`): a click takes the person there
with no code, in the preview and in the exported application. It is not an
`Event`: the router recognises the generated event (`HomeGoToSettings`) and
changes the page itself.

### The export is a router over the screens

A design with pages exports one screen per page (named after the page) and a
file `<design>_app.go`:

```go
type MyAppPage string                 // "Home", "Settings"
type MyAppProps struct {
	Page     MyAppPage                // the page shown: set it to navigate, read it to know where the user is
	Home     HomeProps                // the props of each page
	Settings SettingsProps
}
func MyApp(p MyAppProps, w, h int) Frame                                  // draws the page shown
func MyAppHandle(p *MyAppProps, f Frame, msg tea.Msg) (MyAppEvent, bool)  // goes to pages; reports the page's own events
func MyAppPress(p *MyAppProps, f Frame, msg tea.Msg) bool                 // tabs and panels (ADR 0011)
```

Page selection is therefore part of the contract: the application may bind
`Page`, read it after a click that goes to a page, or drive it. Tabs and panels
keep their own state the same way inside one page (`TabsActive`, `StackOpen`).

`cuppa screens design.cuppa -o dir --app`, and *File → Export Go source* for a
design with pages, write the whole application: `go.mod`, `main.go`, the
screens, the router and the shared runtime, in package `main`.

### Opening an application from its Go

A design that was exported carries itself: the router (or, for a design with
one page, the screen's contract or `layout.go`) has a line

```go
//cuppa:design <base64 of the .cuppa file>
```

`cuppa main.go`, and *File → Open* on a `.go` file, read the design out of it:
the file itself, or, for a `main.go` that only runs the application, the file
beside it that carries it. The design opens as an untitled design (there is no
`.cuppa` to save over), with all its pages, so the application's `main.go` is
the project file. The `.cuppa` stays the source of truth: a hand edit of the
generated Go is not read back, and `cuppa screens --check` still reports it.

### Not done (and why)

Components are not split into packages of their own that screens import. The
runtime is already written once per package and a screen only holds data over
it, so there is no repeated component code to remove; splitting it means
exporting the whole runtime API and a module layout for the generated
application, which is a larger change for a smaller gain. It stays a possible
later step, together with importing components from other repositories by
address (see `docs/spec/cuppa-component.md`).

## Consequences

- Existing designs and exports are unchanged: one page, no router.
- A design with pages is one `.cuppa` file; pages share the canvas size.
- `Document.Pages` and `Document.Page` are part of the file format; older
  readers show the page being edited and drop the others.
