// Command cuppa-desktop is the Cuppa designer in a desktop window. The window
// is a web view (Wails) that shows the very same terminal app as cuppa-tui, in
// a real terminal emulator, through TReactUI.
package main

import (
	"embed"
	"fmt"
	"os"
	"strings"

	"github.com/meta-ui-labs/cuppa/apps/cuppa-desktop/install"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

// version is stamped at build time via -ldflags "-X main.version=...".
var version = "dev"

// fileArgument is the design to open: the first argument that is not a flag.
func fileArgument(args []string) string {
	for _, a := range args {
		if a != "" && a[0] != '-' {
			return a
		}
	}
	return ""
}

// installFlags are the arguments that steer installing: --moved-from=<path> is
// set by a move on the copy it starts, --install asks to move whatever was
// answered before and --uninstall is what Settings > Apps runs.
func installFlags(args []string) (movedFrom string, force, uninstall bool) {
	for _, a := range args {
		switch {
		case a == "--install":
			force = true
		case a == "--uninstall":
			uninstall = true
		case strings.HasPrefix(a, "--moved-from="):
			movedFrom = strings.TrimPrefix(a, "--moved-from=")
		}
	}
	return movedFrom, force, uninstall
}

func main() {
	if len(os.Args) > 1 && (os.Args[1] == "--version" || os.Args[1] == "-v") {
		fmt.Println("cuppa-desktop", version)
		return
	}
	movedFrom, force, uninstall := installFlags(os.Args[1:])
	if uninstall {
		install.Uninstall()
		return
	}
	if install.Startup(version, movedFrom, force) {
		return
	}
	app := newApp(fileArgument(os.Args[1:]))
	err := wails.Run(&options.App{
		Title:            "Cuppa",
		Width:            1280,
		Height:           800,
		MinWidth:         800,
		MinHeight:        500,
		BackgroundColour: &options.RGBA{R: 0, G: 0, B: 0, A: 1},
		AssetServer:      &assetserver.Options{Assets: assets},
		OnStartup:        app.startup,
		OnBeforeClose:    app.beforeClose,
		OnShutdown:       app.shutdown,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "cuppa-desktop:", err)
		os.Exit(1)
	}
}
