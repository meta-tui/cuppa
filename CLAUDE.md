# Cuppa

A terminal design tool for Bubble Tea interfaces, built with Bubble Tea.
Monorepo generated and managed by `mnci` (Nx). See `docs/adr/` for decisions.

## Layout

- `libs/<capability>/<slice>/` Go packages (one per outcome). Headless: never import Bubble Tea or apps. Only `libs/render` imports Lip Gloss. `libs/archcheck` enforces this.
- `packages/components` the Bubble Tea components Cuppa designs with and anyone can import (`colourpicker`, `panels`, `menubar`), one Go package each with its `cuppa.component.json`. They use Bubble Tea and Lip Gloss, so they live outside `libs/`.
- `apps/cuppa-tui` the Bubble Tea front end; slices `shell`, `palette`, `stage`, `inspector`, `menubar`.
- Docs in `docs/`.

## Conventions

- Vertical feature slices, Go mapping: folder = short lowercase package; files are `<what>_<role>.go` (`move_use_case.go`, `rect_model.go`, `node_contract.go`); tests beside the file as `_test.go`. No `util`, `common`, `shared`, `helper`.
- A slice is reached through its package API only; no cycles between slices.
- Conventional commits (`feat:`, `fix:`, `docs:`, `chore:`); versions and releases come from them via `nx release`. Reference issues (`#n`); close with `Closes #n` in the PR.
- Work on a branch, open a PR, merge, release per milestone.

## Commands

- `npx nx run-many -t test lint build`
- Go modules are separate (`go.work`): `go test ./...` inside a lib or app dir.
