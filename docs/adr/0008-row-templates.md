# 8. Row templates for bound lists

Status: superseded by ADR 0017 (removed in #271)

## Context

Drawing MVD's Colours screen from a Cuppa design (spike #170, ADR 0007) worked, but a list of rows could not say what the hand-built screen says:

- a **per-row style**: the selected row is highlighted, a row can be muted, each row has a colour swatch;
- a **row template**: a row of several parts (cursor, name, swatch, value) that the program fills per row, not one string it has to pre-join;
- **whitespace**: list items are trimmed, so leading spaces cannot line things up.

The workaround was one component per row plus a "cursor on <row>" input for each: eight swatches, eight cursor markers and sixteen inputs for one list.

## Decision

### A component for repeated rows: `lipgloss.rows`

The design says how a row looks; the program gives the rows.

| Property | Holds |
|---|---|
| `columns` | `Name[:width][:colour]` items separated by commas: the parts of a row, left to right. A column with no width takes twelve cells, except the last, which takes the rest of the line. `:colour` makes its cells colours (a block of that background) instead of text. |
| `rows` | The sample rows the design shows: cells separated by commas, rows by semicolons (the same escapes as every list: a backslash before a comma, a semicolon or a backslash is that character). **Cells are not trimmed and may be empty.** |
| `styles` | The style of each sample row: `normal`, `selected`, `dim` or `accent`. |
| `color` | The accent: the background of a selected row and the text of an accent row. |

Row styles: **normal** draws the cells as they are, **selected** inverts the whole line in the accent colour (a colour cell keeps its own colour), **dim** uses the muted colour, **accent** colours the text with the accent.

One line per row, as many rows as the component is tall; the program is expected to window a long list itself.

### The contract

Binding `rows` to an input named `Slots` in a screen called `Colours` gives a typed list instead of `[][]string`:

```go
// ColoursSlotsRow is one row of Slots. Style says how the row is drawn.
type ColoursSlotsRow struct {
	Mark   string // a cell
	Name   string // a cell
	Swatch string // a colour
	Hex    string // a cell
	Style  RowStyle
}

type ColoursProps struct {
	Slots []ColoursSlotsRow
	// ...
}
```

The field names come from the column names (`Mark`, `Name`, `Swatch`, `Hex`); a name that gives no Go identifier, or repeats, becomes `Column<N>`. `RowStyle` is shared by every screen of the package: `RowNormal`, `RowSelected`, `RowDim`, `RowAccent`; the zero value is a normal row. The design's sample rows are the default value.

- A click on the component raises the component's event with `Y` the **row index** (rows are one line each), so a program needs no per-row events.
- Binding `styles` as well is ignored with a note: the style of a row is the `Style` of its struct.
- Two rows components that share an input name must have the same columns.

### Column styles, cell styles and cells with colours (batch 1 for MVD)

`columns` takes more tokens after the name and width: `fg=<colour>`, `bg=<colour>`, `selbg=<colour>` (the background the column takes on a selected row; when any column has one, a selected row no longer inverts the whole line), `ellipsis` (a cell that is cut ends in `…`), `colour` (the cell is a block of that background) and `glyph` (the cell is `█` repeated to the width, drawn in that colour). A colour is a palette name (`@Accent`), `#rrggbb` or a palette number. Names are read through the palette the screen hands its parts (`theme.palette`), so a program that sets `p.Palette.Accent` restyles the column.

Every typed row also has a `<Column>Style CellStyle{Fg, Bg string; Bold, Dim, Reverse bool}` next to each cell: what the program says about that cell now (a green tick, a red error tag, the cell under the cursor). It wins over the column's style, and the zero value keeps it. The text of a cell may carry SGR escape sequences (colours, bold, dim, reverse): they are parsed into cells and cut by width, so a text input's view can be a cell.

Widths: a number is fixed; `fill` takes what the other columns leave (shared between fill columns; a last column with no width is one), and `auto` is as wide as the text of its own cell in that row. Columns after a `fill` column are flush with the right end, so a count of any width and a glyph can sit at the end of a row: `Mark:1,Sp:1,Title:fill:ellipsis,Gap:1,Count:auto,Sp:1,Glyph:1`.

Widths from the content: `fit=6..30%` is as wide as the longest cell of any row, at least 6 cells and at most 30% of what the fixed columns and the gaps leave (never below `floor=<n>`, 6 by default here); `fit=0..20` bounds it by cells instead. `hide-empty` gives a column no width and no gap when every cell is empty, `min=<n>` is the least a `fill` column takes, and the component's `gap` property is the number of cells between the visible columns. The playlist editor of MVD is `Review:9,Artist:fit=6..30%,Song:fit=8..50%,Video:fill:min=4,Replaced:fit=0..25%:floor=8:hide-empty,Address:31` with `gap=2`.

Further tokens: `selfg=<colour>` and `selbold` (the foreground and bold a column takes on a selected row) and `fit` (the selected background covers only the cells the text uses, not the column's width). `CellStyle.Plain` keeps the column's colours off one cell; an SGR reset (`0`, `39`, `49`) in a cell's text means the terminal's default colour.

### What does not change

- `lipgloss.list` and the Bubbles components keep trimming their items. Rows are the way to a list with alignment, styles or several parts per row.
- The file format stays version 1: the component is new, its properties are plain text.
- The drawing is the same in the designer and in the exported program (`TestWidgetsDrawTheSameInTheDesignerAndTheProgram` compares them), and the typed rows are checked by building a package and drawing it (`TestRowsScreenBuildsAndDraws`).

## Limits

- One line per row, no wrapping and no per-cell styles other than a colour cell's own colour.
- Cell text is cut at its column's width; a program that needs ellipsis or wrapping shortens the text itself.
- A row has no events of its own (use the click's `Y`) and no state: a cursor is the `Selected` style of one row, kept by the program.
