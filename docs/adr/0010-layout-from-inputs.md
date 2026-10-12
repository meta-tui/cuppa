# 10. Layout that reads inputs and other components

Status: superseded by ADR 0017 (removed in #271)

## Context

A layout expression (ADR 0006) is a calculation over the size of the area: `100% - 4`, `min(50%, 40)`. A real screen needs more: a box that is as high as the list it shows (within a third of the window), a box that starts where the one above it ends, a box that exists only in a wide window.

## Decision

### Expressions

Besides the window, an expression can read:

| Written | Is |
|---|---|
| `$Name` or `$"A name"` | the screen input of that name: a number as it is, a yes/no input as 1 or 0 |
| `below("Name")` | the bottom edge (y + height) of the component called Name, placed before this one |
| `right("Name")` | its right edge (x + width) |
| `top("Name")`, `left("Name")`, `height("Name")`, `width("Name")` | its y, x, height and width |
| `w`, `h` | the width and height of the window (in conditions) |

`min`, `max`, `+ - * /` and `%` work as before. Example: the Playlists box is `h = min(max($PlaylistCount + 2, 5), (100% - 4) / 3)`, the Entries box under it is `y = below("Playlists")` and `h = 100% - below("Playlists") - 1`.

Components are placed in layer order (back to front); an expression sees the components before its own.

### Conditions

*Show if* takes a condition as well as the name of a yes/no input: `w >= 100`, `$Count > 0 && !$Busy`, `h < 30 || $Compact`, with `== != >= <= > <`, `&& || !` and parentheses. A condition that cannot read the place of a component. A name that is only read by a condition or an expression, and bound to no property, becomes a number input (0 by default) and the export says so.

### In the export

- A component whose layout reads inputs or places gets `FitEnv` instead of `Fit`; the program passes the inputs through the screen's props, nothing else changes.
- `<Screen>Layout(props, w, h)` gives the same rectangles without drawing.
- Whole-program export (`cuppa export go`) has no inputs: such an axis is fixed at the size it has in the design.

### In the designer

An input has the value of the property bound to it in the design, and a name's place is where the component is now; they are re-read whenever the design changes. Dragging a component whose axis reads inputs does not rewrite that axis. The Variables screen lists the inputs read by conditions and expressions, and renaming one rewrites `$Name` in them.

## Limits

- Only top-level components can be read by `below()` and the others; a group's parts are not addressable by name.
- A cycle (an axis reading a component placed after it) reads 0 for what is not placed yet.
- Conditions do not read text inputs.
