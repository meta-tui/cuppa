# menubar

A menu bar for [Bubble Tea](https://github.com/charmbracelet/bubbletea) v2 programs: headings along the top
(File, Edit, Help) and, under the one that is open, a box of entries, each with an optional shortcut, and
separators between groups. It is the menu bar of [Cuppa](https://github.com/meta-ui-labs/cuppa), as a component
you can drop into any program.

```sh
go get github.com/meta-ui-labs/cuppa/packages/components
```

## Use

```go
import menubar "github.com/meta-ui-labs/cuppa/packages/components/menubar"

bar := menubar.New(
    menubar.WithMenus(
        menubar.Menu{Name: "File", Entries: []menubar.Entry{
            {Name: "Open…", Shortcut: "Ctrl+O"},
            {Separator: true},
            {Name: "Quit", Shortcut: "Ctrl+Q"},
        }},
        menubar.Menu{Name: "Help", Entries: []menubar.Entry{{Name: "About"}}},
    ),
    menubar.WithSize(40, 8), // the bar is one row, the box of the open menu is drawn under it
    menubar.WithOrigin(0, 0), // where it is drawn, so mouse clicks line up
)

// In Update, forward the messages you want it to see:
bar, cmd = bar.Update(msg)
// A choice arrives as menubar.SelectedMsg{Menu, Entry, MenuName, Name, Shortcut}.

// In View:
bar.View() // a string, bar.Width() columns by bar.Height() lines
```

`bar.ContentHeight()` is the height the bar and its tallest menu need.

## Keys and mouse

| Input | Does |
|---|---|
| Click a heading | Opens its menu, or closes it |
| Move over another heading while open | Switches menu |
| Click an entry | Chooses it |
| F10, or Alt and the first letter of a heading | Opens a menu |
| Left, Right, Up, Down, Home, End | Move through the menus and entries (separators are skipped) |
| Enter or Space | Chooses the entry |
| Esc | Closes the menu |
| An entry's shortcut (`Ctrl+O`, `Ctrl+Shift+S`, `Del`) | Chooses it, with the menu open or not |

The arrow keys are left to your program until the bar has been opened.

## In a design file

The menus are one property, a JSON list, which Cuppa edits as a list of menus with entries you can add,
reorder and drag between menus:

```json
[
  {"name": "File", "entries": [
    {"name": "Open…", "shortcut": "Ctrl+O"},
    {"separator": true},
    {"name": "Quit", "shortcut": "Ctrl+Q"}
  ]}
]
```

`WithMenusJSON` reads it and `MenusJSON` writes it. The description for tools is `cuppa.component.json`.
