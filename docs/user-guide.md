# User guide

Cuppa has three panes and a menu bar:

```
 ☕ Cuppa │ File  Edit  Export  Help                 design.cuppa •
 COMPONENTS      │   canvas                     │  DETAILS
 search          │                              │  X Y W H, layer order
 ▼ Lip Gloss     │      (your design)           │  properties
   Box           │                              │  Layers
 ▼ Bubbles       │                              │
 …               │                              │
 status line: what is selected, or what just happened
```

The dot after the name means there are unsaved changes.

## Building a design

**Add a component.** Press on a name in the left bar, drag onto the canvas and release. A ghost outline shows where it will land; releasing anywhere else cancels, and so does Esc. Click a family heading (Lip Gloss, Bubbles…) to fold or unfold it. Scroll with the wheel, and type in the search box at the top to filter.

**Select.**

| Do | Result |
|---|---|
| Click a component | Select it |
| Shift+click | Add it to, or remove it from, the selection (every selected component shows hollow corner marks and the group gets a dashed frame; a single selection shows solid resize handles) |
| Drag on empty canvas | Draw a box; everything it touches is selected |
| Click empty canvas, or Esc | Deselect |

**Drawing tools.** Above the components, the left bar has a **Tools** list, as in an image editor. A tool is chosen, not dragged: it stays chosen, and dragging on the canvas draws with it. The bar above the canvas shows the options of the chosen tool.

| Tool | Key | Options in the bar | Drag on the canvas |
|---|---|---|---|
| Select | V | Instructions | The normal selecting, moving and resizing |
| Rectangle | U | Fill and border colours; corners ╭ round, ┌ square, ╱ angled | Draws a box between the two corners (Shift: a square as it looks on screen) |
| Path | P | Line character (auto picks `─ │ ╲ ╱`), end (none, arrow, circle), thickness, colour | Draws a line (Shift: 0°, 45° or 90°) |
| Brush | B | Character, thickness, colour | Paints along the drag |
| Erase | E | Thickness. Erases **only the drawing**, never a component | Rubs out what was drawn |
| Text | T | Colour | Click, then type the text; it is placed from the clicked cell |

Everything you draw goes into **one layer**, "Drawing", which is created by the first stroke, sits in front and is the only one. It is transparent where nothing was drawn, so components under it show. Each stroke is one undo step; the layer can be hidden, locked (a locked drawing refuses strokes) or moved like any other from the Layers list. The drawing is saved with the design, drawn in images and text exports, and exported to Go source as the same cells, so it becomes part of the Bubble Tea program. Esc cancels a stroke in progress.

**Move and resize.** Drag a selected component to move it (a whole selection moves together). Drag a corner handle to resize. Components never go below their smallest sensible size or off the canvas. Hold **Shift** while moving to keep the move to a straight line: sideways, up and down, or on a diagonal as it looks on screen (two columns for every row, because a cell is about twice as tall as wide). Hold **Shift** while resizing to keep the proportions (a corner follows whichever side you moved more; a side handle changes the other side to match), and **Alt** (Option on a Mac) to resize from the centre, so both sides move; the two together scale around the centre. Some terminals keep Shift and Alt clicks for themselves; the desktop app and the web page pass them through. While you drag, edges and centres snap to neighbours and a guide line appears; switch this off with *Snap to guides* in the details bar.

**The slice layout.** *File → Export Go source* asks for the layout of the project. **One folder** is the flat layout. **Slices** writes a folder and a package per page (`pages/home/home.cuppa.go`, with `Props`, `Default`, `View`, `Handle`), the router in `app/`, the components the pages use in `components/<family>/`, the runtime in `internal/cuppa/` and `main.go` at the root of the module you name; nothing in it depends on Cuppa. Generated files end in `.cuppa.go` (an earlier export's `home.go`, which has the generated header, is removed after asking; a `home.go` you wrote is never touched). Add your own files beside the generated ones (a page's update logic, tests): exporting again replaces only what Cuppa wrote, removes the pages that are gone and leaves the rest, and `go.mod` is written only the first time. From the command line: `cuppa screens design.cuppa -o dir --layout slices -m github.com/you/app`, with `--check` to see whether the folder is current. `cuppa main.go` opens the design again.

**A master page.** Components that every page shows (a title, a key bar) live once on the master: *Add master page* in the pages list, then design them there. Every other page draws them below its own components, not selectable; the Layers panel lists them under *Master* with an eye to hide one on that page. *Move selection to master* moves components you drew on a page. `cuppa design merge a.cuppa b.cuppa -o app.cuppa --master` lifts what the designs share by itself. Inputs keep the names they had, so the pages' props do not change.

**The design remembers where its Go files go.** *File → Export → Go source* asks for the folder and the layout the first time (*One folder*, *Slices* or *Library*), and saves them in the `.cuppa` file (the folder relative to the design, so the design and the program move together). After that the same menu item only asks "Export, Change… or Cancel", and replaces what Cuppa wrote, keeping your own files. When a page, or the last use of a component, is gone from the design, the files Cuppa wrote for it are listed and you choose *Delete*, *Keep* or *Cancel* before anything is written (from the command line they are listed as they are removed, and `--keep-stale` leaves them). The same goes for `go.mod`: when a component needs a module it lacks, or only a removed component needed one, the designer lists the change and asks *Update go.mod*, *Update and tidy* (runs `go mod tidy`, which writes `go.sum`) or *Skip*; `cuppa screens` lists it and `--update-mod` applies it. Your own requirements are never touched. *Check Go export* tells whether the Go files are what the design exports, without writing anything. From the command line, `cuppa screens design.cuppa` does the same export (and `--check` the check) with no other arguments. The pages list is grouped by folder when the pages have one (*Folder…* in the list sets it).

**An application that exists already.** With `--library` the slices are written into a folder of your module: `cuppa screens app.cuppa -o libs/core --layout slices --library -m your/module/libs/core`. There is no go.mod, main.go or router (`--router` adds it); the runtime goes to `internal/cuppa` and the components to `internal/components`, and each page to the folder you gave it (*Folder…* in the pages list, `setup/preferences`), so the pages of several features share one runtime. Each page declares aliases of the runtime types it uses, so your code names `preferences.CellStyle` and never imports the runtime. A page keeps its own sample values and keys. To make one design from several: `cuppa design merge setup.cuppa run.cuppa -o app.cuppa --folder setup=setup/prefs`.

**Layout inputs.** An input that only a layout expression (`$LeftWidth`) or a condition reads is bound to no property, so the designer had no value for it. With nothing selected, the *Layout inputs* panel lists those names; click one and type the number (or `true`/`false`) the design should be previewed with. The canvas follows it at once, it is saved in the file, and the exported screen's `Default<Screen>Props` starts from it (the field keeps its name and its type, so a program that sets it every frame is not affected). An empty value takes the sample away.

**Pages make an application.** The bar under the canvas is the pages bar: `◂` and `▸` show the previous and the next page, the name (`Home  2/3 ▾`) opens a list to jump to a page or to add, rename, move and delete the one being edited, and `[+ Page]` adds one (the first time it splits the design in two). `Alt+PageUp` and `Alt+PageDown` walk the pages from the keyboard. Each page is a screen: its own components, over the same canvas, theme, keys and named colours. Select a component and choose, in *Behaviour*, **Click goes to** a page: in the **Preview** a click on it takes you there, and in the exported application it does, with no code. *File → Export Go source* of a design with pages writes a whole application (its `main.go`, a screen per page and a router `<name>_app.go`); `<Name>Props.Page` says which page is shown, so the program can read it or set it, and `<Name>Handle` turns clicks and keys into what happened on the shown page. `cuppa screens design.cuppa -o dir --app` does the same from the command line, and `cuppa main.go` (or *File → Open* on a `.go` file) opens the design again out of the application, with all its pages. ADR 0012 has the details.

**Tabs and panels hold components.** *Tab pages* and *Panels* (Lip Gloss, in the left bar) are containers: they have pages, and the components you place on a page belong to it. Drop a component on the body of a container and it joins the page that is shown; it then appears only while that page does. While you edit, pressing a tab or a header only selects the container, so a component is not carried away from under your pointer; to see it work press **Preview** (the two buttons, *Editor* and *Preview*, at the top of the right bar, or Ctrl+P): pressing a tab or a header then changes the page, exactly as it will in the exported program with no code of yours, and *Editor* (or Esc) goes back with the design as it was. To show another page of a container while editing, press its name in the container's *Properties*, or select one of its components in Layers. The container's *Properties* list its pages: press a name to show it, `[x]` removes a page with its components, `[+ Page]` adds one, and the *Pages* text renames them. *Panels* can be an accordion (*Only one open*). Moving or resizing a container moves its components with it, and in Layers its components are listed under it. A container and its components cannot be grouped. The state (`active` for tabs, `open` for panels) can be bound to a screen input like any property. In an exported **program** pressing tabs and headers works by itself. In exported **screens** the state is an input of the contract (`TabsActive int`, `StackOpen []int`, or the name you bound) and `<Screen>Press(&props, frame, msg)` turns a click on a tab or header into the new value, so one line in the program's update changes the page.

**Group components.** Select two or more components (Shift-click or draw a box) and press Ctrl+G, choose *Edit → Group*, or click *Group* in the details bar. They become one component that moves, resizes (its parts scale) and layers as a single row in Layers, and takes the place of its front-most member. Groups can contain groups. Select a group and press Ctrl+U (*Ungroup*) to get the parts back at the group's current place and size. Locked components cannot be grouped, and a locked group cannot be ungrouped.

**Component packs.** The palette is made of packs: Lip Gloss, Bubbles, Huh forms, Glamour, ntcharts and Community. *Edit → Component packs…* lists them with a checkbox each; click one to switch it off or on, and the palette follows at once. Switching a pack off only hides it from the palette: designs that already use its components keep drawing them. Your choice is remembered in `packs.json` in the Cuppa folder of your user config directory.

**Light and dark terminals, colour profiles.** With nothing selected, *Terminal* in the details bar has two choices. *Preview on* **Dark** or **Light** shows the canvas the way a light or dark terminal would: text and backgrounds that use the terminal's own colours turn dark-on-pale, while colours you chose stay, so light text on a light terminal looks as bad as it will be. It is a preview only; exports keep the terminal's own colours. *Colours* steps through **true colour**, **256 colours**, **16 colours** and **no colour**: every colour is reduced to the nearest one the profile has, in the editor and in every export, so you can see (and ship) how the design looks on a limited terminal. With *no colour* only bold and faint text remain.

**Preview.** *View → Preview design* (Ctrl+P) runs the design: the Bubbles components on the canvas become the real Bubbles models and respond to you. Click a component to give it the keyboard (Tab and Shift+Tab move between them), then type in a **text input** or **text area**, move through a **list** or **table** with the arrow keys, page a **paginator** with left and right, scroll a **viewport** with the wheel, nudge a **progress** bar with the arrows, and press Space on a **stopwatch** or **timer**; **spinners** just spin. Everything else (Lip Gloss, Huh, Glamour, charts, your own components) stays as drawn. Nothing you do in the preview changes the design. Esc, Ctrl+P again, or any editing command goes back to editing.

**Screen readers (desktop app).** The desktop window describes the screen to assistive technology: the menus and their items, the palette, the canvas size, the selected component with its position, size and properties, each layer with whether it is hidden or locked, and every dialog as the whole screen while it is open. Feedback such as "Saved design.cuppa" is announced. The description follows what is on screen, with the area that has the keyboard marked as focused, and the keyboard actions are announced: where focus went, the layer you selected with its size and place, where a move or resize ended up, a refusal and why ("Cannot move, Box 1 is locked"), and what a menu item, a palette row or a details control is as you reach it. Everything can be done with the keyboard; see Shortcuts.

**Packs in depth.** [Component packs](packs.md) covers making, editing, installing, sharing and contributing packs, step by step.

**Make your own component.** Select a group (or two or more components) and choose *Edit → Save as component…*, then type a name. The group is saved in the pack *My components* (`my-components.cupp`) and appears in the palette at once. Placed copies scale like a group, and the text properties of its parts (titles, labels, items…) become the component's own properties, so each copy can say something different. Saving again with the same name makes `name-2`.

**Add and remove packs.** *Edit → Component packs…* also has *Add pack…*, which copies a `.cupp` file you picked into the `packs` folder of Cuppa's config directory (see the [format](spec/cupp-format.md)), and a *Remove* button on each installed pack, which asks first and deletes the file. Built-in packs can only be switched off. A pack that cannot be read is skipped, and a notice at start says why.

**Sharing designs.** A `.cuppa` file carries a copy of every pack component it uses, so it opens and draws correctly on a computer that does not have your packs; those components are listed under *From this design*. If the pack is installed, the installed one is used.

**Document options.** With nothing selected the details bar shows the canvas: its width and height (click a number to type one; up to 400 × 200), the background colour (click it to open the colour picker; *None* keeps the terminal's own) and effects applied over the whole design, in the editor and in every export: *Shadows* (a drop shadow under each component), *Scanlines* (every other row dimmed) and *Vignette* (dimmed edges). *Grid dots* only hides the editor's dotted grid; it is never exported. Each change is one undo step and is saved in the `.cuppa` file. Image exports use the background colour as the picture's backdrop.

**Resize the side bars.** Drag the thin line between the left bar and the canvas, or between the canvas and the
details bar, to make a bar wider or narrower. The line lights up when you hover it. The canvas keeps at least 20
columns, and Cuppa remembers the widths for next time.

**The right bar is made of panels**, like the side bar of an image editor. *Properties*, *Behaviour* and *Layers* (and, with nothing selected, *Canvas*, *Theme*, *Named colours*, *Screen keys* and *Display*) each have a header you click to close or open (Enter on the keyboard). Drag a header up or down to move the edge between it and the panel above; the last open panel takes what is left. Each panel scrolls inside its own height with the wheel, with a thin bar at its right edge, so choosing another component never moves the panels, and a component with many properties only scrolls the Properties panel. The sizes and which panels are closed are kept while Cuppa is open. Choices, the *Change* row, a colour's variable name and an input binding open a **dropdown list** under the control: click an option, or use Up, Down and Enter (Esc closes it). Left and Right on the keyboard still step through the options without opening it.

**The contextual bar.** With the Select tool chosen, the two lines above the canvas are the controls of the selection, as the options bar of an image editor is: its name (click to type), its **X, Y, W, H** (click the number and type, `◂` `▸` nudge by one cell, `%` turns the axis into a share of the canvas and `#` back to cells), and the buttons **[Front] [Fwd] [Bwd] [Back]**, **[Duplicate] [Delete]**, **[Group] [Ungroup]** and **Snap**. The drawing tools show their own options there instead.

**Edit details.** The right bar shows the selected component:

- **Layers**: the list at the bottom shows every component, front first; click one to select it, drag to reorder.
- **Theme**: with nothing selected, the details bar shows the design's theme: **Background**, **Text**, **Muted**, **Border** and **Secondary**. Click one to choose its colour (the colour dialog's Themes tab lists terminal colour schemes to pick from). **[Colour scheme…]** fills all five at once from one of 342 terminal colour schemes (background and text as the scheme has them, muted from its bright black, border from its blue, secondary from its purple): step through the schemes or jump by letter, check the sample, and Apply. It is one undo step, and you can change any colour afterwards. Components that have no colour of their own follow the theme: text for labels, lists and trees, muted for placeholders and hints, border for boxes, tabs, tables and frames, secondary for prompts, bars and highlights. A new component comes with no colours of its own, so it follows the theme; the theme is saved in the `.cuppa` file. When you pick a colour for a component's property, that colour is its own and stays when the theme changes: the property then shows `[theme]`, and clicking it removes the override so the component follows the theme again.
- **Properties**: what you can change depends on the component. Text fields are click-then-type (Enter to accept, Esc to cancel). Numbers have `[-]`/`[+]`. Choices (border style, alignment…) are arrows. Yes/no options are checkboxes. Click a colour to open the colour dialog: the **16** and **256** tabs are swatches of the terminal palettes
(a click picks one), **RGB** and **HSL** have sliders for any colour (click or drag the bar, `[-]`/`[+]` for single
steps, true colour on terminals that support it). The **Themes** tab lists 342 terminal colour schemes (Dracula, Nord, Solarized and more, from [bubbletint](https://github.com/lrstanley/bubbletint)): step with `«` `◂` `▸` `»` or the wheel, jump by letter, and click one of its 16 colours, or its text, page, cursor or selection colour; the design stores that colour as plain hex. Type a palette number (`0` to `255`) or hex (`#ff5fd7`) in the
value field and press Enter, or choose **None** for no colour. Long text values wrap inside the bar.

**Responsive layout.** A design is an interface, so a component's position and size can follow the canvas instead of being fixed. In the details bar, click the number of **X**, **Y**, **W** or **H** and type an expression instead of a number:

| Type | Means |
|---|---|
| `30` | 30 cells (columns for X and W, rows for Y and H) |
| `50%` | half of the canvas width (X, W) or height (Y, H) |
| `100% - 10` | calculations with `+ - * /` and parentheses |
| `min(50%, 40)`, `max(…)` | the smaller or larger value |

The `[%]` button next to a row turns that axis into a percentage of the canvas and keeps it where it is; `[#]` turns it back into cells. Typing a plain number over an expression makes the axis fixed again. Dragging, resizing or nudging a component keeps its unit: a `100% - 10` wide component dragged two cells wider becomes `100% - 8`. To see the design at other terminal sizes, change the canvas width and height, or click one of the `[80×24]` `[120×40]` `[160×50]` buttons shown with nothing selected. A top bar that is as wide as the terminal and three rows tall is X `0`, Y `0`, W `100%`, H `3`; a left bar is W `30`, H `100% - 3`. Layout is saved in the `.cuppa` file and applies to components on the canvas; the children of a group keep scaling with their group. `examples/cuppa-shell.cuppa` is Cuppa's own screen built this way.

**Draggable and resizable.** Under X, Y, W and H, *Draggable* and *Resizable* are for the person who uses the exported program, not for the design: a draggable component can be moved with the mouse from anywhere on it, a resizable one by dragging its bottom-right cell. They are saved in the `.cuppa` file and written into the Go export (see *Responsive export*); a component that follows the window keeps the change the person made when the window is resized.

**Undo and redo.** The Edit menu, or Ctrl+Z / Ctrl+Y. A drag or a typed edit is one step.

A `~` after a component in the palette means its preview is an approximation of the real thing.

## Files

| Menu item | Shortcut | What it does |
|---|---|---|
| File ▸ New | Ctrl+N | Start an empty design |
| File ▸ Open… | Ctrl+O | Pick a `.cuppa` file |
| File ▸ Save | Ctrl+S | Save; asks for a name the first time |
| File ▸ Save As… | Ctrl+Shift+S | Save under a new name |
| File ▸ Quit | Ctrl+Q | Leave |

If there are unsaved changes, New, Open and Quit ask whether to save, discard or cancel. Saving over an existing file asks first.

**The file dialog.** Click a folder to go into it (`..` goes up). Click a file once to pick it and again to open it, or type a name and press Enter. Typing a folder path and pressing Enter goes there. On Windows type `D:\` to switch drive. Esc or *Cancel* closes it.

## Exporting

| Export item | Result |
|---|---|
| Image (PNG / SVG / WebP) | A picture of the canvas in a window frame, made by Freeze |
| Colour text (ANSI) | A `.ans` file you can `cat` in a terminal |
| Plain text | A `.txt` file, no colours |
| Go source (Bubble Tea) | A folder with a Go program that runs the design (see below) |
| Go screen (contract) | A Go package that an existing program calls to draw the design (see *Screens for existing programs*) |

**Go source.** *Export → Go source (Bubble Tea)…* asks for a folder (it suggests `<design name>-app` next to the design; it must not already hold a project) and writes a Bubble Tea v2 program there: `main.go` (the model), `layout.go` (your design: where each component sits and its properties), `runtime.go` (the code that runs them) and `go.mod`. Then:

```sh
cd <folder>
go mod tidy
go run .
```

The program places the **real components**: the Bubbles models (text input, text area, list, table, tree, file picker, help, viewport, paginator, spinner, progress, stopwatch, timer), the Huh fields, Glamour markdown, the ntcharts charts, the Evertras bubble table, and Lip Gloss boxes (with their border gradient), labels, lists, tables, trees and tabs. The Huh spinner, the layout sketches (join, place) and the community widgets without a Bubble Tea v2 library (frame, dialog, status message, toast, flex box, boxer, date picker, overlay, status bar, file tree) are drawn to look like the library they stand for, using Lip Gloss. Components that need a library of their own (Huh, Glamour, ntcharts, the bubble table, big text, the QR code) add it to `go.mod` only when the design uses them, and the image reads the file named in its File property from where the program runs. Tab and Shift+Tab move between components that take input, a click focuses one, Esc quits. Groups and pack components are expanded into their parts, hidden components are left out, and the canvas background colour fills behind the components. A component the generator does not know appears as an empty frame; the export notice and the project's README list it. Effects, colour profiles and the light-terminal preview are not part of the generated program.

**Responsive export.** Components with [layout expressions](#building-a-design) are written with a `Fit` function in `layout.go` that computes their rectangle from the window size, and the program places them again whenever the terminal is resized (and fills the whole window with the background colour). A component that is placed again starts from its initial state (typed text is not kept). Groups and pack components keep the size they have in the design.

**Screens for existing programs.** A program that already exists (and keeps its own state and logic) can take its views from Cuppa. Design one screen per file, mark what the program supplies, and export:

- **Bind a property.** In the details bar, under any property, the `⇄` line takes an *input name*. The property is then an input of the screen (typed from the property: number, yes/no, text, or a list of items for `items`, `options`, `rows` and the like). What the property holds in the design stays as the input's default and as what you see.
- **Show if.** A component's *Show if* names a yes/no input; the component is drawn only while it is true.
- **On click.** A component's *On click* names an event the screen raises when the component is clicked.
- **Screen keys.** With nothing selected, *Screen keys* takes `key=Event:label` pairs separated by commas (`s=Save:save, esc=Back:back`). Each key raises its event, and the label is for the program's key bar.
- **Theme.** The roles of the design's theme (text, muted, border, secondary, background) are handed to the screen at run time; components that set their own colour keep it.
- **Named colours.** A colour you set on a property is named after it (*Foreground*, then *Foreground 2*). Under a colour property, the `name ◂ ▸` row steps through the names: choose the same name on another component and both follow it. With nothing selected, *Named colours* lists them (click one to change it, `[x]` to delete one that is not used, `[+ colour]` to add one). A Go screen exposes each name as a field, `p.Palette.Accent = "#33ccff"`, and every component that uses it follows. The *Colour swatch* component shows a colour as a block with its value and a label.
- **Rows.** The *Rows* component (Lip Gloss) is a list whose rows have several parts and a style each. *Columns* names the parts (`Mark:2,Name:14,Swatch:6:colour,Value`: a width after the colon, `:colour` for a block of colour), *Sample rows* is what the design shows (cells between commas, rows between semicolons; spaces and empty cells are kept) and *Row styles* gives each sample row `normal`, `selected`, `dim` or `accent`. Bind its rows and the screen gets a typed list, `p.Slots = []screens.ColoursSlotsRow{{Mark: "▸", Name: "Accent", Swatch: "#7d56f4", Style: screens.RowSelected}}`, and a click on a row raises the event with `Y` as the row. See [ADR 0008](adr/0008-row-templates.md).
- **Layout from inputs.** An axis expression can read a number input (`$PlaylistCount + 2`) and the place of a component before it (`below("Playlists")`, `right`, `top`, `left`, `height`, `width`); *Show if* takes a condition such as `w >= 100 && !$FullLog`. See [ADR 0010](adr/0010-layout-from-inputs.md).
- **Box labels and progress line.** *Box* has a *Right label* in its top edge (`╭─ Title ───── right ─╮`), a title colour and bold flag; a title that would leave no room is cut with `…`. *Progress line* is a bar followed by a text, with the value as a number with a fraction.
- **Key bar.** The *Key bar* component draws key hints (`↑↓:move,enter:edit`): the key in bold in the key colour, the label after it, two spaces between hints, cut at the width. Bind its hints and the screen takes a typed list `[]<Screen><Input>Hint{Key, Label}` the program fills (the hints change with the mode); a click raises the event with the column, and the generated `<Screen><Input>HintAt(props, x)` says which hint that column is on.
- **Model view.** The *Model view* component is where a program draws a view of its own: bind its view and give it what a model rendered (`textarea.Model.View()`, a viewport). The design places and cuts it, colours pass through, and the model keeps its real cursor and scroll. `<Screen>Layout(props, w, h)` returns the regions without drawing, so the program can size the model first (`frame.RegionNamed("Editor")`, then `textarea.SetWidth/SetHeight`, then `view := model.View()`).
- **Colour picker.** The *Colour picker* component (Lip Gloss) is Cuppa's colour dialog as a component: tabs 16, 256, RGB and HSL, gradient sliders, a preview and the value. *Tabs* chooses which show (the 256 tab needs 23 rows). Bind its colour and the screen takes a `ColourPicker` the program owns: `p.Colour = p.Colour.SetOrigin(region.X, region.Y)` once the screen is drawn, `p.Colour, _ = p.Colour.Update(msg)` for the keys and the mouse, and `p.Colour.Value()` is `""`, a palette number or `#rrggbb`. It is also a package of its own for programs that do not use Cuppa: [bubble-colourpicker](https://github.com/meta-ui-labs/bubble-colourpicker), which carries a [component description](spec/cuppa-component.md) that `cuppa component check github.com/meta-ui-labs/bubble-colourpicker` reads.
- **Change a component.** Under *Type* the details bar has a *Change* row when other components can take the selected one's place without losing its variables: a text input and a colour picker (both carry a value), tabs and dialog buttons (both carry items and the chosen one), a list and a dropdown, one table and another. The arrows choose, pressing the name changes it. The name, place, size, show-if, event and bindings stay, and so do the values that are valid for the new component; it is one undo step. From a script (or an agent): `cuppa design json|build|check` turns a design into JSON and back and validates it, and `cuppa component options design.cuppa "Colour field"` and `cuppa component change design.cuppa "Colour field" lipgloss.colourpicker`. See [ADR 0009](adr/0009-component-ports.md).
- **Variables.** *View → Variables…* lists every named colour, screen input (bound property or *Show if*) and event, with its kind, its value and how many places use it. *Go to* selects the component that uses it; when there are several uses a list opens and you choose one. *Rename…* (or `r`) renames a variable everywhere. Typing a different name in a component's own row (the `⇄` line, *Show if*, *On click*, or the colour `name`) asks whether to *rename it everywhere* or *only this one* use another variable; a name that already exists simply links to it.

*Export → Go screen (contract)…* writes the open design into a folder that is a package of its own, next to the screens already there. For a folder of designs at once use the command line, which also removes the files of a design that is gone:

```sh
cuppa screens designs/ -o internal/screens
cuppa screens designs/ -o internal/screens --check   # exits 1 if the files are not what the designs export (for CI)
```

Every file is generated (`// Code generated by Cuppa. DO NOT EDIT.`); Cuppa never overwrites a file it did not write. A screen called *Colours* gives:

```go
p := screens.DefaultColoursProps()   // the screen as designed
p.Slots = []string{"Accent  #7d56f4", "Focus  #00afaf"}
p.ShowStatus = false
frame := screens.Colours(p, width, height)   // frame.View is what to print
// in Update: a click or a key becomes a typed event
if ev, ok := screens.ColoursHandle(frame, msg); ok {
    switch ev := ev.(type) {
    case screens.ColoursSave: // ...
    case screens.ColoursPickSlot: row := ev.Y // the cell clicked inside the component
    }
}
```

Layout expressions follow the `width` and `height` you pass. A removed or renamed input or event is a compile error in your program. See [ADR 0007](adr/0007-screen-contracts.md) for the limits (text inputs draw a static cursor, groups and pack components cannot be bound).

If Freeze is not installed, the image items are greyed out; clicking one explains how to install it. After installing, restart Cuppa, or just click the item again.

## Shortcuts

| Key | Action |
|---|---|
| Ctrl+N / Ctrl+O / Ctrl+S | New / Open / Save |
| Ctrl+Shift+S | Save As |
| Ctrl+Z / Ctrl+Y (or Ctrl+Shift+Z) | Undo / Redo |
| Ctrl+C / Ctrl+V | Copy / Paste the selection (each paste lands a little further on; works between designs) |
| Ctrl+D | Duplicate the selection |
| Del | Delete the selection |
| Ctrl+G / Ctrl+U | Group / Ungroup |
| Ctrl+Shift+] / Ctrl+] | Move the selection to the front / one layer forward |
| Ctrl+[ / Ctrl+Shift+[ | Move the selection one layer back / to the back |
| Arrow keys | Move the selection 1 cell (with Shift, 10 cells); each press is one undo step |
| Up / Down in a number field | Add or subtract 1 (with Shift, 10); Enter ends the edit |
| Ctrl+F | Go to the component search |
| F6 / Shift+F6 | Move the keyboard to the next / previous area: menu bar, palette, canvas, details bar (the focused area's title is highlighted) |
| Alt+1 to Alt+4 | Jump to the menu bar, palette, canvas or details bar |
| F10, or Alt+F / E / V / X / H | Open the first menu, or File / Edit / View / Export / Help |
| Left, Right, Up, Down, Enter, Esc in a menu | Change menu, move through items, run the item, close |
| Up / Down, Home / End, Left / Right, Enter, `/` in the palette | Move through the components, fold or unfold a pack, place the component under the cursor on the canvas at the canvas cursor (and select it), start the search |
| Arrow keys on an empty canvas | Move the canvas cursor (the `┼` mark; with Shift, 10 cells); it is where Enter in the palette places things, and it moves below the component just placed |
| Alt + arrow keys | Resize the selected component: Right / Down grow it, Left / Up shrink it, 1 cell (with Shift, 10); one undo step per press |
| Tab / Shift+Tab | Select the next / previous layer, back to front; Space adds the next layer to the selection |
| Ctrl+A | Select every visible layer |
| Enter on a selection | Move the keyboard to the details bar |
| Tab / Down, Shift+Tab / Up (details bar) | Next / previous control; the focused one is highlighted |
| Enter or Space (details bar) | Press it: edit a field, press a button, tick a checkbox, open the colour dialog, select a layer |
| Left / Right, or `-` / `+` (details bar) | Change a number or a choice by 1 (with Shift, 10) |
| `h` / `l` / Alt+Up / Alt+Down / Delete (on a layer) | Hide, lock, move forward or back, delete |
| Esc | Deselect, cancel a drag, close a menu or dialog; from the menu bar, palette or details bar it returns to the canvas first |
| Ctrl+Q | Quit (Ctrl+C is Copy) |

**On a Mac**, the desktop app and the web page show **Cmd** where this page says Ctrl (and accept it); the Cmd key does what Ctrl does. In a terminal on a Mac, the terminal program keeps Cmd for itself, so use Ctrl there, unless your terminal reports Cmd to programs (kitty, WezTerm and Ghostty can), in which case Cmd works too.

**Keys a plain terminal cannot tell apart.** Ctrl+Shift+S looks like Ctrl+S, and Ctrl+[ looks like Esc, to a terminal that only sends the old key codes. Terminals that report full key presses (those that implement the Kitty keyboard protocol or CSI u, such as kitty, WezTerm and Ghostty) send them properly, and so do the desktop app and the web page. Check yours: if Ctrl+Shift+S saves without asking for a name, it does not. Elsewhere, use the menus: *File ▸ Save As…* and the *Edit* menu's layer items do the same.

## Dialogs by keyboard

| Dialog | Keys |
|---|---|
| Message boxes | Tab / Shift+Tab or the arrows choose the button, Enter presses it, Esc cancels |
| Name prompt | Tab / Shift+Tab move between the field and the buttons |
| Open / Save | Up / Down move through the files (the name fills in), PageUp / PageDown, Home / End, Alt+Up goes to the parent folder; Enter opens a folder or accepts the name |
| Component packs | Up / Down choose, Space switches a pack on or off, Delete or `r` removes an installed pack, `a` adds one, Enter is Done |
| Colour scheme (theme) | Left / Right step (Shift: 10), Up / Down by 10, a letter jumps to it |
| Colour picker | Tab / Shift+Tab change tab; arrows move through swatches; on RGB and HSL, Up / Down choose a slider and Left / Right change it (Shift: 10); PageUp / PageDown change the scheme on Themes; type a value any time |

## Troubleshooting

- **Clicks do nothing**: your terminal needs mouse reporting. Try Windows Terminal, iTerm2, kitty or WezTerm.
- **Odd squares instead of ☕ or ⌕**: your font lacks those symbols; nothing is wrong.
- **Colours look off**: Cuppa uses ANSI 256 colours; use a terminal set to 256 colours or true colour.
- **"not a Cuppa design"** when opening: the file is not a `.cuppa` file or is damaged. See the [format](spec/cuppa-format.md).
