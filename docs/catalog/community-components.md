# Component research

Three rounds: #26 (first catalog), #103 (2026-10-09) and #181 (2026-10-10, section 5). Every Bubble Tea / Lip Gloss
version below was read from the module's `go.mod` on the Go proxy (latest release); stars,
licence and last push come from the GitHub API; features come from each project's README. "README
says" marks a claim I did not check in code.

Cuppa targets Bubble Tea, Lip Gloss and Bubbles **v2** (`charm.land/...`). A v1 library can still
be *designed* in Cuppa (the preview is a drawing), but the generated program (#44) is one module
and cannot mix v1 and v2, so v1-only components stay placeholders there.

## What the catalog has today (63, plus 14 in the bundled packs)

| Family | Components |
|---|---|
| Bubbles (13) | textinput, textarea, list, table, tree, viewport, paginator, filepicker, spinner, progress, timer, stopwatch, help |
| Huh (9) | input, text, select, multiselect, confirm, note, spinner, filepicker, form |
| Lip Gloss (12) | box, label, list, tabs, table, tree, joinh, joinv, place, swatch, rows, colour picker |
| Glamour (1) | markdown |
| ntcharts (8) | bar, line, sparkline, streamline, time series, heatmap, canvas, 3D chart |
| Community (19) | bubbletable, flexbox, boxer, datepicker, overlay, statusbar, filetree, frame, dialog, statusmessage, toast, bigtext, qrcode, image, dropdown, promptinput, promptselect, datatree, pdfview |
| Bundled packs (14) | Card, Alert; Divider, Badge, Key hint, Stat card, Breadcrumb, Sidebar menu, Command palette; Checkbox, Radio, Toggle, Accordion (open and closed) |

## 1. Official libraries: what is missing

| Candidate | Source | Finding | Decision |
|---|---|---|---|
| **Tree** | `charm.land/bubbles/v2/tree` (new in Bubbles v2.2.0, by dlvhdr) | A navigable tree with expand and collapse; the catalog only has the static Lip Gloss tree | **Added** (#105): `bubbles.tree` |
| **Spinner (Huh)** | `charm.land/huh/v2/spinner` | A spinner with a title, shown while something runs | **Added** (#105): `huh.spinner` |
| **Layers and canvas** | Lip Gloss v2 (`Layer`, `Canvas`, compositing with x, y, z) | How modals, toasts and popups are drawn in v2 | Not a component; it is how `overlay`, `modal` and `toast` should be generated |
| **Gradients and blending** | Lip Gloss v2 (`Blend1D`, `Blend2D`) | Gradient fills and borders | **Added** (#105): the `gradient` property of `lipgloss.box` blends the border, in the designer and in the generated program |
| Cursor | `bubbles/v2/cursor` | The blinking cursor inside text inputs | Skip: part of textinput and textarea |
| Key | `bubbles/v2/key` | Key bindings, no visual | Skip: non-visual |
| Textarea options | Bubbles v2.1 and v2.2 | Dynamic height, selection | Properties of `bubbles.textarea` |
| Viewport options | Bubbles v2 | Gutter with line numbers, regex highlight, horizontal scroll, soft wrap | Properties of `bubbles.viewport` |
| Progress colours | Bubbles v2 | Multi-stop colours (`WithColors`) | Property of `bubbles.progress` |

Huh, Glamour and Log have also moved to `charm.land/.../v2` (huh v2.0.3, glamour v2.0.1, log v2.0.1).

## 2. Community libraries, checked on the Go proxy

Bubble Tea column: the major version in the module's `go.mod`.

| Library | Provides | Version (date) | Stars | Licence | Tea | Decision |
|---|---|---|---|---|---|---|
| Evertras/bubble-table | Interactive, paginated, filterable table | v0.23.0 (2026-09) | 579 | MIT | **v2** | In catalog (`community.bubbletable`) |
| NimbleMarkets/ntcharts/v2 | Seven chart types | v2.7.2 (2026-10) | 804 | MIT | **v2** | In catalog. Use the `/v2` module, not the root one (that is v1) |
| lrstanley/bubblezone/v2 | Mouse regions | v2.0.0 (2026-02) | 918 | MIT | **v2** | Non-visual. README says it may not work with the Lip Gloss v2 compositor, which has its own mouse support |
| clambin/bubbles (codeberg) | **frame** (titled container), **dialog** (buttons), **statusbar**, FilterTable, ticker, msglogger | v0.14.1 (2026-08) | n/a | MIT | **v2** | **Added** (#107): `community.frame`, `community.dialog`, `community.statusmessage`. Its status bar is one message line with a level and a spinner, not teacup's four segments, so teacup's `community.statusbar` stays |
| DaltonSW/BubbleUp/v2 | Toast notifications: info, error, success, custom; six positions; NerdFont, Unicode or ASCII symbols | v2.0.0 (2026-08) | 52 | MIT | **v2** | **Added** (#107): `community.toast`. The import path is `go.dalton.dog/bubbleup/v2` (the github.com path on the proxy is the v1 module) |
| erikgeiser/promptkit | Selection, text input, confirmation prompts | v0.12.0 (2026-07) | 311 | MIT | **v2** | **Added** (#181): `community.promptinput` and `community.promptselect`; Huh covers the same inside forms, promptkit is the stand-alone prompt |
| CameronJHall/bubble-datepicker/v2 | Date picker (a fork of EthanEFung's) | v2.0.0 pre (2026-03) | 0 | MIT | **v2** | **Done** (#107): `community.datepicker` now points at this; the original is v1 and unmaintained since January |
| sraaaaaaay/bubbletea-modal/v2 | Modal, dialog and toast overlays using `lipgloss.NewLayer()` | pre-release (2026-04) | 0 | MIT | **v2** (but `go.mod` still lists Lip Gloss v1) | Backlog: watch; replaces `community.overlay` if it matures |
| rmhubbert/bubbletea-overlay | Overlay compositing | v0.6.9 (2026-08) | 126 | MIT | v1 | Keep as placeholder; superseded by Lip Gloss v2 layers |
| 76creates/stickers | FlexBox, responsive table | v1.5.0 (2025-09) | 402 | MIT | v1 | Keep `community.flexbox` as placeholder |
| treilik/bubbleboxer | Layout tree | v0.2.0 (2023) | 87 | MIT | v1 (old) | Keep as placeholder; stale |
| mistakenelf/teacup (statusbar, filetree) | Status bar, file tree | v0.4.1 (2023) | 272 | MIT | v1 (old) | Filetree is superseded by `bubbles.tree`; statusbar stays (different shape from clambin's) |
| lrstanley/bubbletint/v2 | 340+ colour schemes, `color.Color` based | v2.0.2 (2026-05) | 150 | MIT | none | **Done** (#111): the Themes tab of the colour dialog, through `libs/color/scheme` |
| Digital-Shane/treeview/v2 | Tree with search and icon providers | v2.0.1 (2026-05) | 86 | **GPL-3.0** | v2 | **Skip**: a generated program that imports it would have to be GPL |
| pgavlin/tea-grid | Data grid with sorting, filters, pinning | pre (2026-06) | 3 | **none** | v2 | **Skip**: no licence means no right to use it |
| madicen/bubble-color-picker | Colour picker | v0.1.1 (2026-05) | 0 | MIT | v1 | Backlog: v1, 0 stars |
| KevM/bubbleo | Navigation stack, breadcrumbs, menus | v0.1.5 (2024) | 76 | MIT | v1 (old) | Backlog: stale. Breadcrumb is easy to build as a pack |
| Genekkion/bubblegum | Full-screen selection window | v1.1.0 (2024) | n/a | n/a | v1 (old) | Skip |
| marcelblijleven/bubbles-hlist | Horizontal list | v1.0.5 (2025) | 6 | unclear | v1 | Skip: `lipgloss.joinh` of items covers it |
| junhinhow/charm-nav | Tabs, breadcrumb, sidebar | pre (2026-06) | 0 | MIT | Lip Gloss v2 only | Backlog: very new; the three are easy packs |
| blacktop/go-termimg | Images (Kitty, Sixel, iTerm2, half blocks) with a Bubble Tea widget | v0.1.26 (2026-02) | 68 | MIT | no tea dependency listed | Backlog: image placeholder worth having; the `image` component needs a decision on protocol support |
| truffle-dev/glyph | 23 copy-in components: chat bubble, chat input, command palette, diff view, log stream, toast, modal, stat card, kbd, key hints | v0.51.0 (2026-06) | 3 | MIT | v1 | Not a dependency (copy-paste, v1). **Use as a checklist of widgets modern TUIs want** (section 3) |
| muhamm-ad/bubble-ssh | SSH terminal component | v1.1.3 (2026-09) | 0 | MIT | v2 | Skip: niche |
| ella.to/flex, 0xdeafcafe/photon, Hayao0819/reactea | Layout, utilities, React-like framework | pre | 0 | various | v2 | Skip: frameworks and utilities, not widgets |
| ras0q/bubbletree, bntrtm/structly, allisonhere/ripple, tideui, tui-kit | Tree, struct menus, text editor, themes | pre | ≤1 | mixed | v1 or v2 | Skip: too new, 0–1 stars |
| ortizalec/bubbles, calyptia/go-bubble-table, listExtensions, mritd/bubbles, bubblelister, bubble-carousel, bubble-plot | Key-value metrics, tables, lists, prompts, carousel, plots | 2021–2026 | ≤105 | mixed | v1 (old) | Skip: stale or v1; ntcharts and bubble-table cover tables and plots |

## 3. Widgets people expect that no library covers

Real TUIs (gh-dash, k9s, gum, AI chat tools) use these. Most are small compositions of what Cuppa
already has, so they fit as **packs** (`.cupp`, see [packs](../packs.md)): they stay v2-clean,
need no extra dependency, and the generator already expands packs into their parts.

| Widget | Built from | Where |
|---|---|---|
| Divider / rule | Label of box-drawing dashes | pack (done, #108) |
| Badge, tag, chip | Label with background | pack (done, #108) |
| Key hint (`kbd`) | Box + label | pack (done, #108) |
| Stat card | Box + two labels | pack (done, #108) |
| Breadcrumb | One label with arrow separators | pack (done, #108) |
| Stepper / wizard header | joinh of labels + progress | pack |
| Checkbox, radio, toggle | Label with glyphs (`[x]`, `(•)`, `●━`) | pack |
| Sidebar menu | Box + list | pack (done, #108) |
| Command palette | Box + textinput + list | pack (done, #108) |
| Chat bubble, chat input, thread | Box + label + viewport + textarea | pack |
| Log viewer | Viewport with line numbers | `bubbles.viewport` properties |
| Diff view | Viewport with coloured lines | pack |
| Gauge / meter | `bubbles.progress` or ntcharts | existing |
| Slider | Progress with a handle glyph | pack |
| Big text / banner | go-figure FIGlet fonts | `community.bigtext` (done, #113) |
| Image | go-termimg (preview is a placeholder wash) | `community.image` (done, #113) |
| QR code | skip2/go-qrcode, half blocks | `community.qrcode` (done, #113) |

## 4. Proposed order

1. **Official and clearly v2** (small, safe): `bubbles.tree`, `huh.spinner`, gradient property on `lipgloss.box`.
2. **v2 community, MIT, active:** `community.frame`, `community.dialog`, `community.statusbar` (clambin), `community.toast` (BubbleUp/v2). Move `community.datepicker` to the v2 fork.
3. **Packs** for the section 3 list, starting with divider, badge, key hint, stat card, breadcrumb, sidebar menu, command palette.
4. **Theme palettes** from bubbletint (a separate feature, not a component).
5. **Backlog:** image, big text, QR, bubbletea-modal/v2 when it has a release.

Clean-up that follows from this page: replace the teacup statusbar and filetree entries, point
`community.datepicker` at the v2 module, and note in each community entry's package which module
path (`/v2`) the generator must import.

## How an entry is chosen

1. It draws something visible (so it can be placed on a canvas). Non-visual helpers such as
   bubblezone, key and cursor are documented, not catalogued.
2. The licence is permissive. GPL and unlicensed code is left out because generated programs
   would import it.
3. It is not a duplicate of an official component.
4. Prefer a Bubble Tea v2 module with a release in the last year. A v1-only or stale library
   becomes a placeholder.
5. Anything that is only a composition of existing components is a pack, not a new dependency.

## Follow-up

- Re-run the `go.mod` check (Go proxy `@latest`) when libraries release; the v1 rows are the ones to watch.
- Not verified: I read features from READMEs and did not build any of these libraries against Bubble Tea v2.1; each addition's issue should do that first.

## 5. Round 3 (#181): a list of components brought by the user

Each row was checked on 2026-10-10 against the repository, its `go.mod` and its README.

| Asked for | Source | Finding | Decision |
|---|---|---|---|
| Dropdown | madicen/bubble-dropdown | MIT, 1 star, pushed 2026-07. The v2 API is the sub-module `github.com/madicen/bubble-dropdown/v2` (package `dropdownv2`). Needs bubblezone and bubble-overlay | **Added**: `community.dropdown` |
| TextInput, Selection (Promptkit) | erikgeiser/promptkit | MIT, v2 (`charm.land/bubbletea/v2`), 311 stars. Packages `.../textinput` and `.../selection` | **Added**: `community.promptinput`, `community.promptselect` |
| Data-Tree | Evertras/bubble-data-tree | MIT, but Bubble Tea v0.19 and last pushed 2022 | **Added** as a placeholder (like teacup): `community.datatree`. Its drawing is an approximation: the README shows one screenshot only |
| NTCharts 3D | NimbleMarkets/ntcharts3d | MIT, v0.1.0, Go 1.26, GPU-based rendering with a software fallback, pins an unreleased ntcharts commit | **Added** as a placeholder: `ntcharts.chart3d`. The generated program draws the sketch; use the real module by hand |
| NTCharts PDF viewer | NimbleMarkets/ntcharts-pdf | MIT, 2 stars; pdfium through WASM; its `go.mod` replaces Bubble Tea with a fork for the browser build | **Added** as a placeholder: `community.pdfview` |
| Viewport, Table, List, TextInput, TextArea, FilePicker, Spinner, Progress, Paginator, Help, Timer, Stopwatch; Huh form, input, text, select, multiselect, confirm, filepicker; Evertras bubble-table; ntcharts canvas, time series, bar chart | charm and others | Already in the catalog | None |
| Key | charmbracelet/bubbles/key | Not visual. Keys are the design's *Screen keys* (`key=Event:label`) and `bubbles.help` draws them | Skip |
| Marinelli Panes | john-marinelli/panes | **No licence file** (GitHub reports none) and Bubble Tea v1, 19 stars | **Skip**: no licence means no right to use it. `community.boxer` and `lipgloss.joinh` cover pane layouts. Worth asking the author for a licence |
| Funkit Components | github.com/Funkit/bubbletea-components | The URL returns 404; the user Funkit has no such repository | **Skip**: needs the right address |
| Rust bubbletea-widgets | crates.io | A Rust crate. Cuppa exports Go | Skip |
| bubblyui | newbpydev/bubblyui | MIT, 17 stars. A framework, not a widget set: Vue-style reactivity (`ctx.Ref`, `.Setup().Template()`) that replaces the Bubble Tea architecture and starts the program itself (`bubbly.Run()`). Bubble Tea **v1** (v1.3.10), pulls in Sentry and Prometheus. Its built-ins depend on the bubblyui runtime and are not standalone | **Not adopted**: an export that imports it would stop being a plain Bubble Tea program, and v1 cannot mix with v2. Its component list is useful as a checklist (section 6) |

### How the new components are drawn

The designer paints them with `libs/render/scene/community_widgets_painters_algorithm.go`, from what each library's README shows.

## 6. bubblyui's component list against the catalog

| bubblyui | Cuppa today |
|---|---|
| Button, Text, Icon, Spacer | `lipgloss.box`, `lipgloss.label`; no separate button |
| Badge | Widgets pack: Badge |
| Spinner | `bubbles.spinner`, `huh.spinner` |
| Input, TextArea | `bubbles.textinput`, `bubbles.textarea`, `huh.input`, `huh.text`, `community.promptinput` |
| Select | `bubbles.list`, `huh.select`, `community.dropdown`, `community.promptselect` |
| Checkbox, Radio, Toggle | Controls pack (#183) |
| Form | `huh.form` |
| Table, List | `bubbles.table`, `lipgloss.table`, `community.bubbletable`, `bubbles.list` |
| Modal | `community.overlay`, `community.dialog` |
| Card | Starter pack: Card |
| Menu | Widgets pack: Sidebar menu |
| Tabs | `lipgloss.tabs` |
| Accordion | Controls pack: Accordion (open) and (closed) (#183) |
| AppLayout, PageLayout, PanelLayout, GridLayout | Responsive layout expressions on every component, `community.flexbox`, `community.boxer` |

## 7. How a component is added (a checklist for the MCP and the agent instructions)

1. **Check it** against "How an entry is chosen" below: visible, permissive licence, not a duplicate, Bubble Tea v2 preferred. Read its `go.mod` on the proxy and its README. Record what you found in a table above.
2. **Entry**: add a `definition.Definition` in `libs/catalog/standard/` (`community_widgets_entries_content.go` for community ones): `ID` (`family.name`), `Name`, `Family`, one-line `Description`, `DefaultSize`, `MinSize`, `Import` (the real module path, with `/v2` when the module has one), `Status` and the props from the README's options (`textProp`, `intProp`, `boolProp`, `choiceProp`, `colorProp`). Lists are comma separated; multi-line values use `|`.
3. **Designer painter**: a `paintX(g *grid.Grid, p Props)` in `libs/render/scene/`, registered in `scene_render_use_case.go`. It must survive any size from 0x0 and any prop value.
4. **Ports**: say which of its properties carry a string, items, the chosen item, rows and so on in `libs/catalog/standard/ports_content.go`, so it can take the place of (and be replaced by) compatible components with its variables kept ([ADR 0009](../adr/0009-component-ports.md)).
5. **Tests**: a case in the painter test; the default-registry consistency test runs by itself.
6. **Docs**: the counts here and in `README.md`, and the user guide if the component has something to learn.
8. **Check**: `go test ./...` and `golangci-lint run ./...` in `libs/catalog`, `libs/render`, `libs/export`.
