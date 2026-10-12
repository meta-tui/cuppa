package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/meta-ui-labs/cuppa/apps/cuppa-tui/componentcmd"
	"github.com/meta-ui-labs/cuppa/apps/cuppa-tui/designcmd"
	"github.com/meta-ui-labs/cuppa/apps/cuppa-tui/packcmd"
	"github.com/meta-ui-labs/cuppa/apps/cuppa-tui/runcmd"
	"github.com/meta-ui-labs/cuppa/apps/cuppa-tui/screenscmd"
	"github.com/meta-ui-labs/cuppa/apps/cuppa-tui/shell"
	"github.com/meta-ui-labs/cuppa/libs/catalog/standard"
)

// version is stamped at build time via -ldflags "-X main.version=...".
var version = "dev"

func main() {
	if len(os.Args) > 1 && (os.Args[1] == "--version" || os.Args[1] == "-v") {
		fmt.Println("cuppa", version)
		return
	}
	if len(os.Args) > 1 && os.Args[1] == "pack" {
		os.Exit(packcmd.Run(os.Args[2:], os.Stdout, os.Stderr))
	}
	if len(os.Args) > 1 && os.Args[1] == "run" {
		os.Exit(runcmd.Run(os.Args[2:], os.Stdout, os.Stderr))
	}
	if len(os.Args) > 1 && os.Args[1] == "screens" {
		os.Exit(screenscmd.Run(os.Args[2:], os.Stdout, os.Stderr))
	}
	if len(os.Args) > 1 && os.Args[1] == "design" {
		os.Exit(designcmd.Run(os.Args[2:], os.Stdout, os.Stderr))
	}
	if len(os.Args) > 1 && os.Args[1] == "component" {
		os.Exit(componentcmd.Run(os.Args[2:], os.Stdout, os.Stderr))
	}
	app := shell.New(standard.Default())
	app.RestoreLayout()
	app.LoadUserPacks()
	args, connect := os.Args[1:], ""
	for i := 0; i < len(args); i++ {
		if args[i] == "--connect" && i+1 < len(args) {
			connect = args[i+1]
			args = append(args[:i], args[i+2:]...)
			break
		}
	}
	if len(args) > 0 {
		if err := app.OpenFile(args[0]); err != nil {
			fmt.Fprintln(os.Stderr, "cuppa:", err)
			os.Exit(1)
		}
	}
	if connect != "" {
		app.Connect(connect)
	}
	app.Welcome()
	app.ReportPackProblems()
	if _, err := tea.NewProgram(app).Run(); err != nil {
		fmt.Fprintln(os.Stderr, "cuppa:", err)
		os.Exit(1)
	}
}
