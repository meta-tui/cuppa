# Component packs

A **pack** is a file with the extension `.cupp` that holds components you made out of other
components, such as a "Card" built from a frame and two labels. Packs show in the palette next to
Lip Gloss and Bubbles, behave like any other component, and can be shared.

This page covers everything: making a pack, editing it by hand, installing it, sharing it, and
getting it **bundled into Cuppa** so everybody has it. The file format itself is in
[`spec/cupp-format.md`](spec/cupp-format.md).

> **The `cuppa` command.** The commands below are `cuppa pack …`. `cuppa` is the Cuppa program
> (the `cuppa-tui` binary from the releases, renamed or on your PATH); from a clone of the repository
> use `go run ./apps/cuppa-tui pack …` inside `apps/cuppa-tui` instead.

- [How packs fit together](#how-packs-fit-together)
- [1. Make a component in the app](#1-make-a-component-in-the-app)
- [2. Turn it into your own pack](#2-turn-it-into-your-own-pack)
- [3. Edit a pack by hand](#3-edit-a-pack-by-hand)
- [4. Install, switch off, remove](#4-install-switch-off-remove)
- [5. Share a pack or a design](#5-share-a-pack-or-a-design)
- [6. Get your pack bundled with Cuppa (pull request)](#6-get-your-pack-bundled-with-cuppa-pull-request)
- [Rules for bundled packs](#rules-for-bundled-packs)
- [Troubleshooting](#troubleshooting)

## How packs fit together

| Kind | Where it comes from | Can be switched off | Can be removed |
|---|---|---|---|
| Built-in (Lip Gloss, Bubbles, Huh forms, Glamour, ntcharts, Community) | Compiled into Cuppa | yes | no |
| Bundled (for example Starter) | `.cupp` files shipped inside Cuppa | yes | no |
| Installed | A `.cupp` file in your `packs` folder | yes | yes |
| From this design | Copies carried inside the `.cuppa` file you opened | yes | n/a |

Every component has an id of the form `<pack id>.<component id>`, for example `starter.card`.
That id is what a saved design refers to, so **never rename a pack id or a component id once
people use it**.

A pack component is drawn by painting its *parts* (ordinary components with their own settings)
scaled to the size you place it at. Only the properties the author *exposed* show in the details
bar; each one sets a property of one part.

## 1. Make a component in the app

1. Put the parts on the canvas and set them up: a Box, a Label for the heading, a Label for the text.
   Give the parts clear names in the details bar (**Heading**, **Body**), because those names become
   the labels of the component's properties.
2. Select all the parts (Shift-click, or draw a box around them).
3. Choose **Edit → Save as component…** and type a name, for example `Tea card`.

The component is saved in the pack **My components** and is in the palette straight away. Drag it
out, resize it (the parts scale), and look at the details bar: **every text property of every part**
(up to 12) is exposed as a property of the component, so each copy can say something different.

Tips for a good component:

- Draw it at the size people will use most. That becomes its default size.
- Leave room: parts scale proportionally, so text parts get wider when the component does.
- Set the parts' text to a sensible default; it becomes the default of the property.
- Group first if you want to check how it moves and scales as one (**Ctrl+G**), then save.

Saving a second component with the same name gives it the id `name-2`, so nothing is overwritten.

## 2. Turn it into your own pack

*My components* is a personal pack stored as `my-components.cupp` in your `packs` folder (see
[section 4](#4-install-switch-off-remove) for where). To publish components under **your own pack
name and id**, take a copy and edit it as JSON:

```sh
cuppa pack json "<packs folder>/my-components.cupp" > tea-shop.json
```

Open `tea-shop.json` in an editor and change the **pack** fields at the top:

```json
{
  "pack": {
    "id": "tea-shop",
    "name": "Tea shop",
    "version": "1.0.0",
    "description": "Cards and banners for a tea shop",
    "components": [ ... ]
  }
}
```

Delete the components you do not want to share, then build the file:

```sh
cuppa pack build tea-shop.json          # writes tea-shop.cupp next to it
cuppa pack check tea-shop.cupp          # lists what is inside
```

`build` refuses a pack that could not be read back and says what is wrong first, so a file you built
is always a valid one.

## 3. Edit a pack by hand

You can also write a pack from scratch, or fine-tune one the app produced. The JSON has this shape
(this is the real Starter pack that ships with Cuppa, shortened):

```json
{
  "pack": {
    "id": "starter",
    "name": "Starter",
    "version": "1.0.0",
    "description": "Small building blocks: a titled card and an alert banner",
    "components": [
      {
        "id": "card",
        "name": "Card",
        "description": "A framed card with a heading and a line of text",
        "w": 28, "h": 6,
        "nodes": [
          { "id": "frame",   "component": "lipgloss.box",   "name": "Frame",
            "rect": { "x": 0, "y": 0, "w": 28, "h": 6 } },
          { "id": "heading", "component": "lipgloss.label", "name": "Heading",
            "rect": { "x": 2, "y": 1, "w": 24, "h": 1 },
            "props": { "bold": "true", "text": "Card title" } }
        ],
        "props": [
          { "key": "heading", "label": "Heading", "kind": "text", "default": "Card title",
            "target": "heading", "targetProp": "text" },
          { "key": "color", "label": "Border color", "kind": "color", "default": "212",
            "target": "frame", "targetProp": "color" }
        ]
      }
    ]
  }
}
```

**Pack fields**

| Field | Rule |
|---|---|
| `id` | Lowercase letters, digits and single dashes (`tea-shop`). Must not be the id of a built-in or bundled pack, or of another pack you have installed. |
| `name` | Required. What the palette and the Packs dialog show. |
| `version` | Free text, for example `1.0.0`. Bundled packs must have one. Raise it when you change a published pack. |
| `description` | One line for the Packs dialog. Bundled packs must have one. |

**Component fields**

| Field | Rule |
|---|---|
| `id` | Same characters as the pack id, unique inside the pack. |
| `name`, `description` | Shown in the palette and its search. |
| `w`, `h` | Default size in cells, at least 1. The `rect`s of the parts are relative to a box of this size. |
| `nodes` | The parts, **back to front** (last is on top). Each has an `id` (unique inside the component), `component` (any component id), `name`, `rect` (`x`, `y`, `w`, `h`, with `w` and `h` at least 1) and optional `props`, `hidden`, `locked`. |
| `props` | The properties the component offers (see below). Optional. |

**Property fields**: `key` (unique in the component), `label` (shown in the details bar), `kind`
(`text`, `int`, `bool`, `color` or `choice`; `choice` also needs a `choices` list), `default`, and
the pair `target` (the `id` of a part) and `targetProp` (the property key of that part it sets).

**Finding component ids and property keys.** The command

```sh
cuppa pack catalog label
```

lists every built-in component whose name or id contains `label`, with its default and minimum size
and each property's key, kind, default and choices. Run it with no argument for everything. Values
are always strings in the file: `"true"`, `"28"`, `"212"` (an ANSI colour 0 to 255) or `"#ff5fd7"`.

A part can itself be a component from another pack (for example your own Card inside a bigger
Dashboard). Loops are stopped at a fixed depth, so a component containing itself just draws a
placeholder.

After every edit run `cuppa pack check yourfile.json` (it accepts the JSON directly) and fix what it
reports, then `cuppa pack build`.

## 4. Install, switch off, remove

**Install** in either way:

- **In the app:** *Edit → Component packs… → Add pack…*, pick the `.cupp` file. Cuppa copies it into
  your packs folder.
- **By hand:** copy the file into the folder below and restart Cuppa.

| System | Packs folder |
|---|---|
| Windows | `%AppData%\cuppa\packs` |
| macOS | `~/Library/Application Support/cuppa/packs` |
| Linux | `~/.config/cuppa/packs` |

(That is the `cuppa\packs` folder inside your user config directory.) A file whose pack id is
already taken, or that is damaged, is skipped, and a notice at start says why.

**Switch off:** *Edit → Component packs…* and click the pack's checkbox. The palette hides it at
once; designs that use it keep drawing. The choice is remembered.

**Remove:** the **Remove** button on an installed pack in the same dialog asks first and deletes the
file. Designs that use it still open (see below). Built-in and bundled packs have no Remove button.

## 5. Share a pack or a design

- **Share a pack:** send the `.cupp` file. The other person uses *Add pack…*.
- **Share a design:** just send the `.cuppa` file. When you save, Cuppa stores a copy of every pack
  component the design uses inside it, so it opens and draws correctly on a computer that does not
  have your pack. Those components are listed under **From this design**. If the pack *is* installed
  there, the installed version wins.
- A design that uses a component of a pack that is neither installed nor embedded (for example a file
  from an older Cuppa) still opens; the part is drawn as a labelled placeholder.

## 6. Get your pack bundled with Cuppa (pull request)

Bundled packs are `.cupp` files in the Cuppa repository. They are compiled into the app, so every
user has them after the next release, listed with the built-in packs. Anyone can propose one.

### Before you start

- Read [Rules for bundled packs](#rules-for-bundled-packs). The most important ones: the pack must
  be your own work (or under a licence that lets Cuppa ship it under MIT), use only built-in
  components, and have a unique id.
- Have Go (see [`contributing.md`](contributing.md#setup)) and the GitHub CLI `gh` (or use the web
  interface), and a build of Cuppa to try it in.

### Steps

1. **Fork and clone** the repository, then branch:

   ```sh
   gh repo fork meta-ui-labs/cuppa --clone
   cd cuppa
   git checkout -b feat/pack-tea-shop
   ```

   Open an issue first if the pack is big or you want feedback on the idea; mention the issue number
   in your commit and PR.

2. **Build your pack** as in sections 1 to 3 and put the finished file in
   `libs/catalog/bundled/packs/<pack id>.cupp`. The file name must be the pack id plus `.cupp`:

   ```sh
   cuppa pack build tea-shop.json libs/catalog/bundled/packs/tea-shop.cupp
   ```

   Make sure `version` and `description` are filled in.

3. **Run the checks.** The bundled-pack test fails with a clear message if the pack is damaged, has no
   description or version, repeats an id, uses a component that does not ship, or has no components:

   ```sh
   cd libs/catalog && go test ./...
   cd ../../apps/cuppa-tui && go test ./...
   ```

4. **Try it in the app.** Build and run Cuppa (`cd apps/cuppa-tui && go run .`). Open *Edit →
   Component packs…*: your pack is listed, without a Remove button. Drag each component out, resize
   it to its smallest and a large size, edit its properties, undo and redo, save the design, and
   reopen it.

5. **Take a screenshot** of your components placed on a canvas (a terminal screenshot, or export a
   PNG with *Export → Image*). You will put it in the pull request.

6. **Write the docs line.** Add a row for your pack to the table in
   [`docs/catalog/bundled-packs.md`](catalog/bundled-packs.md) (create the file if it does not exist
   yet): pack id, name, what it is for, its components, and who made it.

7. **Commit** with a conventional message. The header is at most 100 characters and the subject is
   lowercase:

   ```sh
   git add libs/catalog/bundled/packs/tea-shop.cupp docs/catalog/bundled-packs.md
   git commit -m "feat(catalog): add the tea shop pack (#123)"
   git push -u origin feat/pack-tea-shop
   ```

8. **Open the pull request** against `main`:

   ```sh
   gh pr create --title "feat(catalog): add the tea shop pack" --body-file pr.md
   ```

   Use this as the body of `pr.md`:

   ```markdown
   Closes #123

   ## Pack
   - id: `tea-shop`, version `1.0.0`
   - components: Card, Banner, Price tag
   - what it is for: ...

   ## Checklist
   - [ ] `cuppa pack check` is clean and `go test` passes in `libs/catalog`
   - [ ] only built-in components are used
   - [ ] every component looks right at its smallest and a large size
   - [ ] the pack is my own work and I am happy for it to ship under the MIT licence
   - [ ] docs row added

   (screenshot)
   ```

9. **Review.** CI runs the tests and lint. A maintainer will look at the pack with
   `cuppa pack json libs/catalog/bundled/packs/tea-shop.cupp`, which prints the readable form of the
   binary file, and may ask for changes (names, defaults, sizes). Push more commits to the same
   branch to update the pull request.

10. **Release.** After it is merged, the next release of Cuppa includes the pack: it is listed with
    the built-in packs, can be switched off but not removed, and needs nothing from the user.

### Changing a bundled pack later

Open a pull request that replaces the `.cupp` file. Keep every existing pack and component **id**
unchanged (saved designs refer to them), raise `version`, and only add components or improve defaults.
Removing or renaming a component breaks designs that use it, so it needs a good reason and a note in
the PR.

## Rules for bundled packs

- **Unique ids.** The pack id and every component id must differ from built-in and other bundled ids.
  The test checks this.
- **Only components that ship.** Parts may use built-in components (`cuppa pack catalog` lists them)
  or components of other bundled packs, never of an installed pack such as *My components*.
- **Complete metadata.** `name`, `version` and `description`, and at least one component.
- **Sensible, small and tidy.** A handful of useful components beats a hundred near-duplicates. Name
  parts clearly, give every exposed property a plain label, and choose defaults that look good.
- **Works at any size.** Check the smallest and a large size.
- **Licence.** Only submit work you made, or that you may license under the MIT licence of this
  repository, and say so in the PR.
- **No surprises.** A pack is data: parts, sizes and property values. It cannot run code.

## Troubleshooting

| Message | What it means and what to do |
|---|---|
| `not a .cupp file` | The file does not start with the pack header. Build it with `cuppa pack build`; do not rename a `.cuppa` design. |
| `file was written by a newer version of Cuppa` | Update Cuppa. |
| `file is damaged` | The file is truncated, not compressed JSON, or has no valid pack id or name. Rebuild it from the JSON. |
| `pack id "…" must be lowercase letters, digits and dashes` | Fix `id`: `Tea Shop` becomes `tea-shop`. |
| `pack "…": id already in use` | Another pack (built-in, bundled or installed) has this id. Choose another id, or remove the other pack. |
| `component "…" has a damaged part` | A part has no id, a repeated id, or a width or height under 1. |
| `property "…" drives nothing` | `target` is not the id of a part, or `targetProp` is empty. |
| `property "…" has no choices` | A `choice` property needs a `choices` list. |
| A component draws as a labelled box | Its pack is not installed or embedded, or a part uses a component that no longer exists. |
| The component is not in the palette | The pack is switched off (*Edit → Component packs…*), or the file was skipped; look for the notice at start. |
