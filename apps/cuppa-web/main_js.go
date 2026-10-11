//go:build js && wasm

// Command cuppa-web is the Cuppa designer compiled to WebAssembly. The page
// shows it in the same TReactUI terminal the desktop app uses; there is no
// server, the terminal app runs inside the browser tab.
package main

import (
	"context"
	"strings"
	"syscall/js"

	tea "charm.land/bubbletea/v2"
	"github.com/meta-ui-labs/cuppa/apps/cuppa-tui/shell"
	"github.com/meta-ui-labs/cuppa/libs/catalog/standard"
	ttygo "github.com/meta-tui/treactui/packages/tty-go"
)

// onAMac reports whether the browser says it runs on a Mac, from the User
// Agent, so the menus can say Cmd. The page turns Cmd into the Ctrl the app
// listens for (see frontend/src/shortcuts.ts).
func onAMac() bool {
	navigator := js.Global().Get("navigator")
	if navigator.IsUndefined() {
		return false
	}
	ua := navigator.Get("userAgent").String()
	return strings.Contains(ua, "Macintosh") || strings.Contains(ua, "Mac OS X")
}

func newModel() tea.Model {
	app := shell.New(standard.Default())
	app.UseIconFont()
	if onAMac() {
		app.UseCommandKey()
	}
	app.Welcome()
	return app
}

func main() {
	shared := ttygo.NewSharedProgram(newModel)
	ttygo.BindShared(context.Background(), shared, newPageEvents(), ttygo.BindOptions{})
	select {} // keep the module alive; everything happens in callbacks
}
