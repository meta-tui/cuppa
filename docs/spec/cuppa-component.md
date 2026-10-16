# The Cuppa component description

Version 1. A Go module that offers a Bubble Tea component publishes a file named `cuppa.component.json` at
the root of its repository (or, in a monorepo, in the folder of the component). It says what the component is, how to use it from Go, which properties a designer
can set and which messages it sends. A design tool needs nothing else to list the component, show its
properties and generate the Go that uses it.

The reference implementation is `libs/catalog/component` in Cuppa; `cuppa component check <file | address |
github.com/owner/repo[/folder][@ref]>` reads a description and lists what is wrong with it. Cuppa's own
components publish one each: [colourpicker](../../packages/components/colourpicker), [panels](../../packages/components/panels)
and [menubar](../../packages/components/menubar), in the `packages/components` module of this repository.

## Example

```json
{
  "cuppa": 1,
  "name": "colourpicker",
  "title": "Colour picker",
  "description": "Tabs for the 16 and 256 terminal palettes and for RGB and HSL sliders.",
  "version": "0.1.2",
  "license": "MIT",
  "homepage": "https://github.com/meta-ui-labs/cuppa/tree/main/packages/components/colourpicker",
  "go": {
    "module": "github.com/meta-ui-labs/cuppa/packages/components",
    "import": "github.com/meta-ui-labs/cuppa/packages/components/colourpicker",
    "package": "colourpicker",
    "bubbletea": 2,
    "model": "Model", "constructor": "New",
    "init": "Init", "update": "Update", "view": "View",
    "width": "Width", "height": "Height", "origin": "SetOrigin"
  },
  "size": { "default": { "w": 46, "h": 10 }, "min": { "w": 46, "h": 4 } },
  "props": [
    { "key": "value", "label": "Colour", "kind": "color", "default": "#ff007f",
      "go": { "option": "WithValue", "setter": "SetValue", "getter": "Value" } },
    { "key": "tab", "label": "Tab shown", "kind": "choice", "default": "RGB",
      "choices": ["16", "256", "RGB", "HSL"], "go": { "option": "WithTab", "setter": "SetTab", "getter": "Tab" } }
  ],
  "events": [
    { "name": "changed", "label": "Colour changed", "go": { "message": "ChangedMsg" },
      "fields": [{ "name": "Value", "type": "string", "prop": "value" }] }
  ],
  "preview": { "kind": "frame" }
}
```

## Fields

| Field | Rule |
|---|---|
| `cuppa` | Required. The version of this format: `1`. |
| `name` | Required. Lowercase letters, digits and single dashes. With `family` it makes the catalog id (`community.colourpicker`). |
| `title`, `description` | Required. What the palette and its search show. |
| `version`, `homepage` | Free text. |
| `license` | Required. A generated program imports the code, so its licence must be known. Cuppa leaves out components that are not permissively licensed. |
| `family` | Optional provider: `lipgloss`, `bubbles`, `huh`, `glamour`, `ntcharts` or `community` (the default). Shown beside the name and in About; it also makes the pack the component belongs to. |
| `category` | Optional palette group, what the component is for: `layout`, `box`, `text`, `input`, `select`, `data`, `charts`, `feedback`, `navigation`, `files`, `colour`, `media` or `other` (the default). |

### `go`

| Field | Rule |
|---|---|
| `module`, `import`, `package` | Required: the module path, the import path of the package and its name. |
| `bubbletea` | Required: `1` or `2`, the major version the component is written for. A program can only mix components of the version it uses. |
| `model`, `constructor` | Required: the exported type and the function that makes one. |
| `init`, `update`, `view` | The methods of the Bubble Tea shape; empty if the component has none. |
| `width`, `height` | Methods that give the size of what `view` draws, if it has a fixed size. |
| `origin` | A method that tells the component where on the screen it is drawn, `(x, y int)`, so that mouse positions can be read. |

### `size`

`default` and `min` are `{ "w", "h" }` in cells, at least 1, and `min` is not larger than `default`. `note` is
free text for the designer.

### `props`

The properties a designer sets. Each has:

| Field | Rule |
|---|---|
| `key` | Required. Letters, digits and underscores; unique. |
| `label` | Required. What the details bar shows. |
| `kind` | `text`, `int`, `float`, `bool`, `color`, `choice` or `menu`. |
| `default` | The value in a new component, always a string. A choice's default is one of `choices`; an int's is within `min` and `max`; a bool's is `true` or `false`. |
| `choices`, `min`, `max` | For `choice` (two or more) and `int`. |
| `menu` | A list of menus, written as JSON text: `[{"name": "File", "entries": [{"name": "Open…", "shortcut": "Ctrl+O"}, {"separator": true}]}]`. A name is required for every menu and every entry that is not a separator; the shortcut is optional. The default must parse. The details bar edits it as a list, with entries that can be added, reordered and dragged between menus. |
| `go.option` | The constructor option that sets it (`WithValue`). |
| `go.setter`, `go.getter` | The methods that set it later and read it. One of option or setter is required. |
| `go.variadic`, `go.split` | A list written in one text property (`"16,256,RGB"`): `split` is the separator, `variadic` says the option takes the items as separate arguments. |

### `events`

The messages the component sends. `name` is lowercase with dashes; `go.message` is the Go type of the message;
`fields` list its fields, each with `name`, `type` and, if it carries a property, `prop`.

### `preview`

How a design draws the component without running it. `frame` (the default) is a labelled frame;
`static` draws `lines`. A component that is built into Cuppa has its own drawing.

## What a tool does with it

1. **Lists it**: the palette entry is `family.name`, titled `title`, with the sizes and properties above.
2. **Draws it**: as the preview says.
3. **Exports it**: a program that places the component imports `go.import` and builds it with
   `constructor`, passing each property through its `option`. Components that Cuppa knows, like the colour
   picker, are exported with their real code; for others the import and the construction are the
   generator's to write from this file.

Unknown fields are reported by the checker, so a misspelt property is not silently lost. A description that
changes keeps `name` and the property `key`s: saved designs refer to them.

## Publishing one

1. Put `cuppa.component.json` next to the code: at the root of the repository, or in the component's folder.
2. Run `cuppa component check .` (or the file) until it says the description is in order.
3. Keep it true: a test in the module that every `option`, `setter`, `getter` and `message` it names exists
   (`packages/components/panels`'s `TestComponentDescriptionMatchesTheCode` is an example).
4. Tag a release; tools read the description at the tag (`github.com/owner/repo@v1.2.3`, or `github.com/owner/repo/folder@v1.2.3`).
