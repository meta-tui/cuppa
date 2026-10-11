# 4. The desktop app is the terminal app in a Wails window

Status: accepted (issue #46)

## Context

ADR 0001 made the engine headless so more front ends could follow the terminal
one. A desktop front end could have been a second UI (HTML components calling
the engine through bindings). That would mean designing and maintaining the
same palette, canvas and inspector twice.

[TReactUI](https://github.com/meta-ui-labs/treactui) (`@treactui/tty`, `tty-go`)
shows a Bubble Tea program in a web page as a real terminal (xterm.js) over a
WebSocket, with an accessibility layer beside it. It is by the same author and
is built to be hosted by exactly this kind of window.

[Wails](https://wails.io) is the Go desktop framework: a Go backend and a web
view. Wails v3 is still in beta (`v3.0.0-beta.28` when this was written); v2 is
the stable line.

## Decision

`apps/cuppa-desktop` is a **Wails v2** app whose window contains one
`<TTY>` component showing **the same Cuppa terminal app** as `cuppa-tui`
(`apps/cuppa-tui/shell`), served in-process through `tty-go`.

- One program, one UI. Nothing in the designer is duplicated for the desktop.
- The program is one *shared* `ttygo.SharedProgram` (a reload of the window
  finds the design as it was), bound to the window with `ttygo.BindShared` over
  Wails events; the page uses `createWailsSocket()`. There is no server, port,
  secret or origin check. (The first version ran a loopback WebSocket with a
  random secret in its path; tty-go v0.1.3 made that unnecessary and it was
  removed.)
- Closing the window sends Ctrl+Q into the program, so the app's own
  unsaved-changes prompt runs. When the user quits from inside the app the
  window closes (`OnBeforeClose`, `OnQuit`).
- The first-run Freeze notice and everything else in the shell work unchanged.

Because `tty-go` serves Bubble Tea v2, it was migrated from v1 first
(`meta-ui-labs/treactui` #27). Two things surfaced while running Cuppa in the page
and were fixed there: a v2 program asks for mouse mode and the alternate
screen from its `View` (so late-joining browsers are put in the same terminal
modes), and v2's bare line feeds must not be turned into carriage returns by
the page (`ESC D` is sent instead, #31).

## Consequences

- The desktop app needs no per-component work: a new catalog entry appears in
  both front ends.
- It inherits the terminal UI's limits (cell grid, mouse and keyboard through a
  terminal emulator) and gains TReactUI's accessibility layer when the app
  describes its screen (`ttygo.Accessible`, not done yet).
- Wails v2 needs a webview and cgo on Linux (GTK, WebKit) and cgo on macOS, so
  it cannot ride mnci's static cross-compiled release. The app is not tagged
  `release:go`; desktop releases need a per-OS build job (future work).
- Moving to Wails v3 later should only touch `main.go` and `app_binding.go`.
