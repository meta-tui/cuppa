# The `.cuppa` file format

Version **1**. Code: `libs/cuppafile/format`.

A `.cuppa` file is a short binary header followed by gzip-compressed JSON.
The header lets a reader reject foreign or too-new files before it touches
any JSON; the JSON stays easy to inspect and diff after `gunzip`.

```
offset  size  field
0       6     magic      the bytes "CUPPA\n"  (43 55 50 50 41 0A)
6       2     version    unsigned, big endian (currently 0x0001)
8       …     body       gzip stream; inflates to one JSON object
```

## Body (version 1)

```json
{
  "document": {
    "name": "Login screen",
    "width": 120,
    "height": 40,
    "seq": 3,
    "nodes": [
      {
        "id": "n1",
        "component": "huh.input",
        "name": "Input",
        "rect": { "x": 4, "y": 2, "w": 30, "h": 3 },
        "props": { "title": "Name", "placeholder": "Ada" }
      }
    ]
  }
}
```

| Field | Meaning |
|---|---|
| `document.name` | Display name of the design. |
| `document.width`, `height` | Canvas size in terminal cells. At least 1 and at most 400 × 200; larger values are clamped on load. |
| `document.background` | Optional canvas colour: `0`–`255` or `#rrggbb`. Absent means the terminal's own. An invalid value is dropped on load. |
| `document.theme` | Optional `{ "text", "muted", "border", "secondary" }`, each a colour like `document.background`. With the background, these are the design's theme: a component colour that the node does not store follows the matching theme colour (text for labels, lists and trees; border for boxes, tabs, tables and frames; secondary for accents such as prompts, bars and highlights; muted for quiet text such as placeholders and hints). A colour a node does store is its own and stays whatever the theme says. An invalid colour is dropped on load. `theme.palette` is an optional list of `{ "name", "color" }`, the design's named colours: a colour property whose value is `"@Name"` uses the swatch of that name. Entries with an invalid name or colour, and repeated names, are dropped on load. See [ADR 0007](adr/0007-screen-contracts.md). |
| `document.profile` | Optional colour profile the design targets: `256`, `16` or `none`. Absent means true colour. Colours are reduced to it in the editor and in every export. An unknown value is dropped on load. |
| `document.light` | Optional, `true` to preview on a light terminal in the editor. Never part of an export. |
| `document.hideGrid` | Optional, `true` hides the editor's dotted grid. The grid is never exported. |
| `document.effects` | Optional `{ "shadow", "scanlines", "vignette" }`, each `true` when on. Applied in the editor and in every export. |
| `document.seq` | Counter behind generated node ids; only ever grows, so ids are never reused. |
| `nodes` | Back to front: the **last** node is drawn on top. |
| `node.id` | Unique within the document. |
| `node.component` | Catalog id (`bubbles.list`, `huh.select`, `community.flexbox`…). Unknown ids load and render as a labelled placeholder, so a file from a newer catalog still opens. |
| `node.rect` | Position and size in cells. `w` and `h` are at least 1. |
| `node.props` | String values the designer changed. Anything not listed uses the component's default. |
| `node.hidden` | Optional, `true` for a hidden layer: not drawn, exported or hit by the pointer. Absent means visible. |
| `node.children`, `node.baseW`, `node.baseH` | Optional, on a group (`component` is `cuppa.group`): the grouped nodes, in the same shape, with rectangles relative to the group's top-left, laid out for a box of `baseW` × `baseH`. A placed group scales them with its own size. A group with no valid children or no base size is dropped on load; groups nest up to 16 deep. Ignored on any other node. |
| `document.components` | Optional list of `{ "id", "component" }`: a copy of every pack component the design uses (`id` is the catalog id such as `tea-shop.card`, `component` has the shape of a component in a [`.cupp` file](cupp-format.md)), written on save so the file draws without the pack. A reader ignores a copy that is invalid or repeated, keeps at most 256, and prefers an installed pack's component over the copy. |
| `node.locked` | Optional, `true` for a locked layer: drawn, but it cannot be moved, resized, deleted or edited. Absent means unlocked. |
| `node.layout` | Optional object of size expressions `x`, `y`, `w`, `h` (strings such as `"50%"`, `"100% - 10"`, `"min(50%, 40)"`; grammar in `libs/layout/expr`). An absent or empty axis is fixed at its `rect` value. `rect` is always the result at the canvas size stored in the file. Absent means every axis is fixed. |

## Reading rules

A reader:

1. rejects a file that does not start with the magic (`not a .cuppa file`);
2. rejects a version greater than it knows (`written by a newer version`);
3. rejects version 0, a missing version or a bad gzip stream (`damaged`);
4. refuses bodies that inflate beyond 64 MiB;
5. migrates older versions one step at a time to the current one;
6. drops nodes with an empty or repeated id or a rect under 1×1, and raises a
   canvas under 1×1 to 1×1, so a damaged or hand-edited file never gives the
   editor an impossible document.

## Writing rules

A writer always writes the current version, and saves atomically: it writes a
temporary file next to the target and renames it into place, so a crash never
leaves half a file where the old one was.

## Fields that no longer exist

Files written by cuppa-tui 0.0.88 and before may carry fields of the screen-contract
features that were removed (ADR 0017): `node.bind`, `node.showIf`, `node.event`,
`node.draggable`, `node.resizable`, `document.keys`, `document.samples`, `document.defaults`,
`document.perPage`, `document.export`, `document.run`, and on a page `samples`, `defaults`,
`keys`, `values` and `folder`. A reader ignores them, and the next save drops them.

## Changing the format

Adding an optional field needs no version bump (`hidden` and `locked` were added this way: files without them load as visible and unlocked). Anything that changes meaning
or removes a field:

1. raise `CurrentVersion` in `file_contract.go`;
2. add `migrations[oldVersion]` in `migrations_algorithm.go` that turns the old
   JSON body into the new one;
3. keep a fixture of the old version in the tests.

`FuzzDecode` checks that no input crashes the reader and that anything it does
accept survives an encode/decode round trip unchanged.

## The drawing layer

What the drawing tools paint is one node whose component is `draw.layer`. Its
only property, `cells`, lists the painted cells in reading order, joined by
semicolons: `x,y,codepoint,fg,bg`, with the codepoint in hexadecimal and the
colours as any colour of the format (a palette number, `#rrggbb`, or empty).
A cell at no position is not painted, so nothing under it is covered. There is
at most one such node; it covers the canvas and is locked, hidden or moved like
any other.
