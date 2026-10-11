# Contributing

Thanks for helping. This page covers setup, the conventions the repository
enforces, how work is released, and how to add a component to the catalog.

## Setup

You need Go (see `go.work` for the version), Node 24 with npm 11, and Git.
Freeze is optional (only for trying picture export).

```sh
git clone https://github.com/meta-ui-labs/cuppa
cd cuppa
npm ci
npx nx run cuppa-tui:start      # run the app
```

Everyday commands:

```sh
npx nx run-many -t test lint build      # everything, as CI does
npm run cuppa-tui:qa                    # lint + test one project
cd libs/render && go test ./...         # one module, plain Go
```

Each project has its own Go module and a root `go.work` ties them together, so
`go test ./...` works inside any `libs/*` or `apps/*` directory. Lint is
`golangci-lint`; install it with `go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest`.

## Conventions

**Vertical slices.** One Go package per outcome, files named
`<what>_<role>.go`, tests beside the code. The roles in use: `_model`,
`_contract`, `_use_case`, `_algorithm`, `_store`, `_content`. No `util`,
`common` or `shared`. Full rules are in [`CLAUDE.md`](../CLAUDE.md).

**Keep libraries headless.** Nothing under `libs/` imports Bubble Tea, Bubbles
or an app, and only `libs/render` imports Lip Gloss. `libs/archcheck` fails the
tests if you break this. Behaviour belongs in a library; the terminal app only
maps input and draws. See [architecture](architecture.md).

**Tests.** New behaviour comes with tests. UI slices are tested by sending
simulated `pointer.Event`s and key presses and reading the rendered lines;
look at `apps/cuppa-tui/shell/shell_model_test.go` for the pattern.

**Commits.** [Conventional Commits](https://www.conventionalcommits.org/):
`feat:`, `fix:`, `docs:`, `chore:`. The header is at most 100 characters, and
a footer needs a blank line before it. Reference issues with `#n`.

## Workflow and releases

1. Every piece of work has a GitHub issue, labelled and in a milestone.
2. Work on a branch; open a pull request whose description says `Closes #n`.
3. CI (`npx mnci ci`) must pass. Merge with squash, using a conventional title.
4. Merging to `main` runs `nx release`: it picks the version from the commits,
   tags `cuppa-tui@x.y.z` and creates a GitHub Release with a zip per platform.

The CI workflow has one hand-added step to install Go and golangci-lint,
working around [MoNecromanCI#405](https://github.com/MoNecromanCI/MoNecromanCi/issues/405).
Remove it when that is fixed upstream.

## Adding a component to the catalog

A component is a **definition** (what it is and what you can edit) plus a
**painter** (how its preview looks). Say you are adding `Badge` to Lip Gloss.

### 1. Describe it

Add an entry in the family's file under `libs/catalog/standard/`
(`lipgloss_entries_content.go`, `bubbles_entries_content.go`, …):

```go
{
	ID: "lipgloss.badge", Name: "Badge", Family: definition.FamilyLipgloss,
	Description: "A short coloured tag.",
	DefaultSize: size(10, 1), MinSize: size(3, 1),
	Import: lipglossImport, Status: definition.StatusSupported,
	Props: []definition.PropSpec{
		textProp("text", "Text", "new"),
		colorProp("color", "Background", "212"),
		boolProp("bold", "Bold", true),
	},
},
```

- **ID**: `<family>.<name>`, unique, lowercase. Saved files refer to it, so never change it later.
- **Status**: `supported` if the preview is faithful; `placeholder` if it is an approximation (the palette marks these with `~`); `planned` for known but not yet designable.
- **Props**: each has a key, a label, a kind and a default. The helpers `textProp`, `colorProp`, `boolProp`, `intProp(…, min, max)` and `choiceProp(…, choices…)` are in `prop_builders_content.go`. The inspector builds its editor from this list, so there is no UI code to write. Values are stored as strings.
- **Import**: the Go import path of the real component, for the future code generator.

### 2. Paint it

Add a painter in `libs/render/scene/` (in the family's `*_painters_algorithm.go`)
and register it in the `painters` map in `scene_render_use_case.go`:

```go
func paintBadge(g *grid.Grid, p Props) {
	style := grid.Style{Bg: p.Str("color"), Fg: "0", Bold: p.Bool("bold")}
	g.Fill(full(g), ' ', style)
	g.Text(1, 0, p.Str("text"), style, g.W-2)
}
```

A painter gets a grid the size of the node and the node's properties with
defaults already applied. It must draw inside the grid at **any** size down to
`MinSize` without panicking, and must not read anything outside its props.
Helpers such as `full`, `fg` and `dim` (`lipgloss_painters_algorithm.go`) and
`wrap` and `fillColumn` (`painting_helpers_algorithm.go`) save repeating yourself.

### 3. Test it

- `TestEveryShippedComponentRendersAtDefaultAndMinimumSize` already runs your component at both sizes and fails if a `supported` entry has no painter.
- Add one test that renders a specific set of props and compares the lines, like `TestBoxWithTitle` in `scene_render_use_case_test.go`.

Run `cd libs/catalog && go test ./... && cd ../render && go test ./...`.
The component then appears in the palette and inspector with no further work.

### A bundled pack instead

If what you want to add can be built out of components that already exist, make it a pack and
send the `.cupp` file: no Go code needed. See
[Component packs](packs.md#6-get-your-pack-bundled-with-cuppa-pull-request).

### Community components

Unofficial components need research before they are listed: confirm the module
exists, read its license and which Bubble Tea major version it targets from its
`go.mod`, and add a row to
[`docs/catalog/community-components.md`](catalog/community-components.md).
Use `StatusPlaceholder` unless the preview is faithful.

## Changing the file format

See [the format spec](spec/cuppa-format.md#changing-the-format): bump the
version, add a migration, keep a fixture of the old version.
