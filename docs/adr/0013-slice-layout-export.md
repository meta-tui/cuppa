# ADR 0013: The slice layout of an exported application

Status: proposed (plan; nothing is built yet)

## Context

ADR 0012 exports an application with pages as one flat package: a screen per
page, a router and the shared runtime side by side. The person who asked for
it works in feature slices: a page is a feature, so it should have its own
folder and a file named after it, and the components it uses should be in a
components folder. The flat layout stays the default; this is an option.

## Go constraints that shape it

- A folder is a package. A page in its own folder is a package of its own, and
  what it uses from another package must be exported.
- Packages may not import each other in a cycle. Pages must not import pages,
  so a click that goes to another page cannot call it; it raises an event and
  the router, which imports every page, changes the page. That is how ADR 0012
  already works.
- Today's runtime has about forty unexported helpers (`placed`, `drawScreen`,
  `fitEnv`, `cells`, `joinList`, the container arithmetic...) and the
  components register themselves in an unexported map. Everything a page or a
  component package touches has to become part of an exported API.
- `main.go` must be in a `package main` directory; `go run .` and `go install`
  want it at the module root.
- The generated module needs to know its own import path for the imports
  between its packages, so the export takes a module path.

## Proposal

```
myapp/
  go.mod                       module <path given with -m>
  main.go                      package main: imports app, runs the Bubble Tea program
  app/
    app.go                     the router: Page, Props, Default, View, Handle, Press (ADR 0012)
  pages/
    home/
      home.go                  the page: Props, Default, View, Layout, Handle, Press, events, Keys, Size
      <your own files>.go      untouched by the export
    settings/
      settings.go
  components/
    lipgloss/                  one package per family of the kinds the pages use
    bubbles/
    community/
    colourpicker/              (the colour picker, tabs and panels are components too)
  internal/
    cuppa/                     the shared runtime: grid, layout expressions, drawing, containers
```

Inside a page package the names lose their prefix, as Go packages do:
`home.Props`, `home.Default()`, `home.View(p, w, h)`, `home.Handle(f, msg)`,
`home.Pick{X, Y}` for the event `Pick`. The flat layout keeps `HomeProps` and
`DefaultHomeProps`, so a program that already uses it is not affected.

**`main.go` at the module root.** It imports `app`, builds `app.Default()`,
and runs the program. The router is in `app/` and not in `main` so a test, or
another binary under `cmd/`, can use it. `--layout slices` puts `main.go` at
the root; a different place is a flag (`--main cmd/shop`), not a different
layout.

**Components.** The runtime is generated into `internal/cuppa`, the components
into `components/<family>` (not one folder per kind: the kinds of a family
share their drawing helpers, such as the grid the community widgets draw
into, and splitting them would only duplicate those). A component package
exports `Build` for its kinds and registers them in an `init`; a page package
imports the families it uses with a blank import, so only the code of the
kinds a design uses is in the application. Both are generated into the
application: it has no dependency on Cuppa and the code is the application's
to change. The alternative, importing a published `cuppa/run` module, couples
the application's version to Cuppa's and to the name of its repository, and is
not offered.

**What stays hand-written.** The export replaces only files that start with
the generated header, as `cuppa screens` does today, and removes the generated
files of pages that are gone. A page folder may hold the program's own files
(`home/update.go`, tests); they are never read or deleted.

**Opening it again.** `main.go` is the entry: it imports `app`, `app` imports
the pages, and each page folder has the file named after the page. The design
is embedded once, in `app/app.go` (`//cuppa:design`), as in ADR 0012, and each
page file records the digest of the page it was made from
(`//cuppa:page name="Home" sha="…"`), so a hand edit of generated code is
noticed by `cuppa screens --check` and a page can be regenerated on its own
with diffs that touch one folder.

## How it is built

The flat export is not rewritten. The slice layout is a second writer over the
same pieces, in four steps, each its own pull request and each leaving the
flat layout byte-for-byte as it is.

1. **An exported name table.** One table says which unexported runtime names
   are part of the API (`placed` becomes `Placed`, `drawScreen` becomes
   `Draw`...). A transform over the Go tokens (`go/scanner`, identifiers only,
   not strings and not selectors) renames them in the runtime text and in the
   generated page code; tests build both layouts of the same designs and
   compare the screens' drawings cell by cell.
2. **Pages as packages and the router in `app/`, `main.go` at the root**, with
   the runtime still one package, `internal/cuppa`. `-m <module>` and
   `--layout slices` on `cuppa screens` and on *Export Go source*.
3. **Components as packages by family.** The extension files are split by
   family; the registration map becomes `cuppa.Register`.
4. **Round trip and tooling.** `--check` for a tree, `cuppa main.go` for the
   slice layout, the export dialog's layout choice, the user guide and the
   agents' migration guide.

Compatibility is kept by a test that exports every example in the repository
in both layouts and compares what they draw.

## Decisions asked of the owner

1. `main.go` at the module root (proposed) or under `cmd/<name>/`.
2. Components by family (proposed) or one package per kind.
3. Generated runtime in the application (proposed) or imported.
