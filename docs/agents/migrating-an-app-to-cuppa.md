# Moving an app's screens to Cuppa (guide for agents)

For an agent (or a person) with a Bubble Tea v2 app that draws its screens by hand and wants them drawn from
Cuppa designs. First worked through on MVD (https://github.com/meta-ui-labs/cuppa, ADRs 0007 to 0009). Written
as a working checklist; the reasons are in the ADRs.

## The idea in one paragraph

Each screen of the app becomes one **design** (a `.cuppa` file). `cuppa screens` turns the designs into a Go
package of **data-agnostic screens**: for a design called *Colours* you get `Colours(props, w, h) Frame`, a
`ColoursProps` struct (everything the screen shows), `ColoursEvent` types (what a click or a key means) and
`ColoursHandle(frame, msg)`. The app keeps all its state and logic; it fills the props, prints
`frame.View`, and turns messages into events. Generated files are never edited by hand.

## Commands

```sh
# Build cuppa (or use a release): from a clone of github.com/meta-ui-labs/cuppa
cd apps/cuppa-tui && go build -o cuppa.exe .

cuppa design json   screens/colours.cuppa          # a design as JSON (read it, diff it, review it)
cuppa design build  screens/colours.json           # JSON -> .cuppa, validated (unknown fields/properties refused)
cuppa design check  screens/colours.cuppa          # problems and export notes
cuppa screens designs/ -o libs/app/screens -p screens   # every design -> Go package (replaces its own files)
cuppa pack catalog [text]                          # every component with its property keys, kinds, defaults
cuppa component options design.cuppa "Node name"   # components a node can be changed for without losing variables
cuppa component change  design.cuppa "Node name" lipgloss.colourpicker
cuppa component check   github.com/owner/repo[@ref]   # a module's component description
```

## A design, in JSON

```json
{"document": {
  "name": "Colours", "width": 80, "height": 24, "background": "",
  "theme": {"palette": [{"name": "Accent", "color": "#ff007f"}]},
  "keys": [{"key": "esc", "event": "Back", "label": "back"}],
  "nodes": [
    {"id": "n1", "component": "lipgloss.box", "name": "Frame",
     "rect": {"X": 0, "Y": 0, "W": 80, "H": 24}, "layout": {"w": "100%", "h": "100%"},
     "props": {"title": "Colours", "color": "@Accent"}},
    {"id": "n2", "component": "lipgloss.label", "name": "Status",
     "rect": {"X": 2, "Y": 22, "W": 76, "H": 1}, "layout": {"w": "100% - 4", "y": "100% - 2"},
     "props": {"text": "saved"}, "bind": {"text": "Status"}, "showIf": "Show status", "event": "Dismiss"}
  ]}}
```

- Coordinates are cells, origin top-left. `rect` is the size at the design's width and height; `layout`
  expressions (`"100%"`, `"100% - 4"`, `"50%"`) follow the area the screen is drawn in.
- `props` values are strings. A colour is `"212"` (palette), `"#rrggbb"` or `"@Name"` (a palette colour of the
  design). Lists are comma separated, rows semicolons then commas, outlines `|` with two spaces per level.
  A comma or semicolon inside an item is written `\,` and `\;` (a backslash is `\\`).
- **Bind** (`bind`: property key → input name) makes that property a field of the screen's props, typed from
  the property (number → `int`, yes/no → `bool`, list → `[]string`, rows → `[][]string`, text → `string`).
  What the property holds in the design is the default.
- `showIf` names a yes/no input: the component is drawn only while it is true.
- `event` names what a click on the component raises; `keys` raise events for keys.

## What to do, in order

1. **Inventory the screens.** For each: the file and function that draws it, the sizes it must work at, its
   states (editing, empty, error), the keys and mouse behaviours, and the colours it uses (name each one).
2. **Name the colours once.** Every colour the app takes from its configuration becomes a palette colour of
   the design (`Accent`, `Dim`, ...) and a field of the `Palette` struct; the app hands its own values in.
3. **One design per screen.** Build it with `cuppa design build` from JSON you generate, or with Cuppa's editor
   API from a Go test (`libs/export/gosource/mvd_colours_rows_example_test.go` is a worked example). Keep the
   designs in the app's repository.
4. **Export** with `cuppa screens` into the package the app imports. Commit the generated files: a removed or
   renamed input or event is then a compile error in the app, which is what you want.
5. **Switch the screen's view** to `Screen(props, w, h).View` and its mouse/keys to `ScreenHandle`. The
   app's state and update logic stay where they are.
6. **Prove it is the same screen** (next section) before deleting the old drawing code.

## Proving a screen is the same

Render the old and the new view for a list of states and sizes, strip nothing, and compare cell by cell:

- text: `ansi.Strip` both, split into lines, compare;
- colours and bold: parse both with a cell grid (`github.com/charmbracelet/x/ansi` + `uv`, or compare the
  runs of SGR codes per line) and compare foreground, background and bold per cell.

Keep the old function as `legacyView` while the test exists; delete both when it passes. Write the
differences down as gaps (below) instead of working around them in the app.

## Components you will use most

| Need | Component | Notes |
|---|---|---|
| Frame with a title | `lipgloss.box` (`title`, `border`, `color`) | the title sits in the top edge |
| Text | `lipgloss.label` (`text`, `color`, `bold`, `align`) | one line; does not wrap |
| A list the app windows itself, with a selected row, colour cells, a hit per row | `lipgloss.rows` (ADR 0008) | typed `[]<Screen><Input>Row`; click raises the event with `Y` = row; no scrolling built in |
| A colour choice | `lipgloss.colourpicker` | typed `ColourPicker` the app owns: `p.Picker = p.Picker.SetOrigin(x, y)`, `p.Picker, _ = p.Picker.Update(msg)`, `.Hex()` |
| Table | `lipgloss.table`, `bubbles.table` | |
| Tabs / buttons | `lipgloss.tabs`, `community.dialog` | `items` + `selected` ports |
| Progress, spinner | `bubbles.progress`, `bubbles.spinner` | |

Placing something on top of the screen (a modal, a toast): draw the screen, then composite the other view
over `frame.View` with Lip Gloss layers at the region of a named component (`frame.RegionNamed("Name")`).

## Changing your mind about a component

`cuppa component options` lists components that carry the same kind of data (ports, ADR 0009); `change`
keeps the name, place, bindings, show-if and event. Do this instead of deleting and re-adding.

## An input the app owns (text area, text input, viewport)

Do not try to make the design draw a cursor. Keep your own Bubbles model, size it from the design and put its
view in a *Model view* (`lipgloss.slot`) component bound to a string input:

```go
r, _ := screens.EditorLayout(props, w, h).RegionNamed("Paste")   // where the slot is, without drawing
m.textarea.SetWidth(r.W); m.textarea.SetHeight(r.H)
props.PasteView = m.textarea.View()                             // real cursor, scroll, placeholder
frame := screens.Editor(props, w, h)
```

A one-line text input inside a *Rows* cell works the same way: its `View()` goes into the cell text (SGR
colours are kept and the cell is cut to the column width).

## Keys, events and the key bar

`keys` in the design (`key=Event:label`) raise events and say what the keys are called. A key bar that changes
with the mode is a *Key bar* component with its hints bound: the program fills `[]<Name>Hint{Key, Label}`
from the same table it uses for the keys, and a click on the bar arrives as the component's event with `X`.

## What to commit

Commit the designs (`designs/*.cuppa`) and the generated package (`cuppa_*.go`, `*_screen_*.go`), and say in
the commit which `cuppa-tui` version wrote them, so CI and a contributor without Cuppa can build. Regenerate
with `cuppa screens` and review the diff; never edit a generated file. In CI run
`cuppa screens designs -o <package folder> -p <package> --check`: it writes nothing and exits 1, listing the
files that are missing, differ from what the designs export, or are left over.

## Limits to know now

- A bound text input is a string with a static cursor; real cursor/selection/scroll offset are gaps to report.
- Lists and rows are one line per row; no wrapping.
- Group and pack components cannot be bound (set bindings on their parts).
- The runtime is copied into the package (`cuppa_*.go`, regenerated on each export).

## Reporting something Cuppa cannot do

One message per gap, first line `GAP <screen>: <what>`, then the exact old output with row/column (plain text
and colours as 256 numbers or hex), what Cuppa draws, and how to reproduce (state, size). A gap becomes a
property or a component of Cuppa; a new component is its own library with a `cuppa.component.json`
(see `docs/spec/cuppa-component.md` and github.com/meta-ui-labs/bubble-colourpicker as the example).
