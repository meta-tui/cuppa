# cuppa-desktop

Cuppa in a desktop window: a [Wails](https://wails.io) v2 shell around the very
same terminal app as `cuppa-tui`, shown in a real terminal emulator by
[TReactUI](https://github.com/meta-ui-labs/treactui). See
[ADR 0004](../../docs/adr/0004-desktop-app-wails-and-treactui.md).

```
window (web view)                       Go process
  <TTY createSocket={createWailsSocket()}>  ◀── Wails events ──▶  ttygo.BindShared  ──▶  shell.New(...)
  React page                                  treactui:up / :down   (no server, no port)
```

## Run

You need Go, Node and the Wails CLI (`go install github.com/wailsapp/wails/v2/cmd/wails@latest`;
`wails doctor` checks the rest).

```sh
npx nx run cuppa-desktop:start      # wails dev: live window, also at http://localhost:34115
npx nx run cuppa-desktop:bundle     # wails build: build/bin/cuppa-desktop(.exe)
npx nx run cuppa-desktop:test       # Go tests
```

On Linux install GTK and WebKit first (`libgtk-3-dev`, and `libwebkit2gtk-4.1-dev`
with `-tags webkit2_41` on Ubuntu 24.04). CI does this for you.

## Releases

The `Desktop release` workflow builds the app on Windows, macOS (universal) and Linux (amd64) after a
cuppa-tui release and attaches `cuppa-desktop-<os>-<arch>.zip` to that GitHub Release. It starts when the CI
workflow finishes, because a release made with `GITHUB_TOKEN` does not trigger other workflows. The desktop app
releases together with cuppa-tui: they share the engine, and a change to either rebuilds both.
You can try it without releasing from Actions > Desktop release > Run workflow (leave "upload" off).

The builds are not signed: Windows may show SmartScreen, and macOS needs right-click > Open the first time.
Installers and signing are tracked in #68.

## Layout

| | |
|---|---|
| `main.go` | Starts Wails: window options and the embedded page |
| `app_binding.go` | The window lifecycle (start, close, shutdown) and the shared program bound to the window |
| `wails_events_adapter.go` | Gives tty-go Wails' `EventsOn` / `EventsEmit`, so tty-go needs no Wails dependency |
| `terminal/` | `OnQuit`, which closes the window when the app quits |
| `frontend/` | The React page: one `<TTY>`. Built by `wails build`, not by Nx |

## Behaviour worth knowing

- The page and the program talk over Wails events, so there is no server, no
  open port and nothing for another program on the computer to connect to.
- Closing the window asks the app to quit first, so unsaved changes prompt.
- Reloading the window (`wails dev`) keeps the design: the program is shared.
- `frontend/dist` only holds a `.gitkeep` in git so the Go code compiles; real
  assets come from `wails build`.
