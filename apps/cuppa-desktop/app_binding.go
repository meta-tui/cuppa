package main

import (
	"context"
	goruntime "runtime"
	"sync/atomic"

	tea "charm.land/bubbletea/v2"
	"github.com/meta-ui-labs/cuppa/apps/cuppa-desktop/terminal"
	"github.com/meta-ui-labs/cuppa/apps/cuppa-tui/shell"
	"github.com/meta-ui-labs/cuppa/libs/catalog/standard"
	ttygo "github.com/meta-ui-labs/treactui/packages/tty-go"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App ties the window to the terminal app: one shared Bubble Tea program that
// the page reaches over Wails events.
type App struct {
	ctx context.Context
	// openPath is a design to open at start, from the command line.
	openPath string
	shared   *ttygo.SharedProgram
	unbind   func()
	// closing is set once the app itself has decided to quit, so closing the
	// window then goes ahead instead of asking the app again.
	closing atomic.Bool
}

// newApp takes the design to open at start, or "".
func newApp(openPath string) *App { return &App{openPath: openPath} }

// startup binds the program to the window's events. The program is shared:
// reloading the window finds it as it was.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.shared = ttygo.NewSharedProgram(a.newModel)
	a.unbind = ttygo.BindShared(ctx, a.shared, wailsEvents{ctx}, ttygo.BindOptions{})
}

// newModel builds the Cuppa terminal app. The window closes when the user
// quits from inside it.
func (a *App) newModel() tea.Model {
	app := shell.New(standard.Default())
	app.UseIconFont()
	if goruntime.GOOS == "darwin" {
		app.UseCommandKey() // the page turns Cmd into the Ctrl the app listens for
	}
	app.RestoreLayout()
	app.LoadUserPacks()
	if a.openPath != "" {
		app.OpenFileOrNotify(a.openPath)
	}
	app.Welcome()
	app.ReportPackProblems()
	quit := terminal.OnQuit(app, func() {
		a.closing.Store(true)
		runtime.Quit(a.ctx)
	})
	return terminal.Describe(quit, app)
}

// beforeClose turns "the user closed the window" into Ctrl+Q inside the app,
// so its own unsaved-changes prompt runs. It returns true to keep the window
// open until the app quits itself.
func (a *App) beforeClose(context.Context) bool {
	if a.closing.Load() || a.shared == nil {
		return false
	}
	return a.shared.Send(tea.KeyPressMsg{Code: 'q', Mod: tea.ModCtrl})
}

func (a *App) shutdown(context.Context) {
	if a.unbind != nil {
		a.unbind()
	}
}
