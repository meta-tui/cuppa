package gosource

import (
	"encoding/base64"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/meta-ui-labs/cuppa/libs/document/design"
)

// DesignDirective starts the line of a generated file that carries the whole
// design, so an application can be opened again in the designer from its Go:
// "//cuppa:design <base64 of the .cuppa file>". See ADR 0012.
const DesignDirective = "//cuppa:design "

// Options are what an export may add that the document does not say.
type Options struct {
	// Source is the design as the bytes of a .cuppa file. An application with
	// pages embeds it next to its router.
	Source []byte
	// Module is the import path of the generated module; the slice layout
	// needs it for the imports between its packages. Empty means the name of
	// the design.
	Module string
	// Library writes the slice layout as packages of a module that exists:
	// no go.mod, main.go or README.md, the runtime in internal/cuppa and the
	// components in internal/components under the folder, and the router only
	// when Router asks for it.
	Library bool
	// Router writes app/app.go in a library export; an application always has
	// it.
	Router bool
	// master is the master of the design generated as a screen of its own,
	// when the pages take its parts from a package (the slice layout).
	master *masterShare
	// masterNotes collects why a page keeps its own copy of the master's parts.
	masterNotes *[]string
}

// gotoEvent is the name of the event a component raises when a click on it
// goes to a page: the router recognises it and changes the page, and the
// program sees it like any other event.
func gotoEvent(page string) string { return "Go to " + page }

// pageField is the field of the application's props that holds a page's props.
func pageField(page string) string {
	f := pascal(page)
	if f == "Page" || f == "Theme" || f == "" {
		f += "Screen"
	}
	return f
}

// routerSource is the router of a design with pages: the pages' props together
// with the one that is shown, the function that draws it and the one that
// turns a click or a key into what happened (a click on a component that goes
// to a page changes the page by itself).
func routerSource(doc design.Document, screens []design.Document, opts Options) string {
	app := pascal(doc.Name)
	if app == "" {
		app = "App"
	}
	var b strings.Builder
	fmt.Fprintf(&b, "%s\n\n", GeneratedHeader)
	if len(opts.Source) > 0 {
		fmt.Fprintf(&b, "%s%s\n\n", DesignDirective, base64.StdEncoding.EncodeToString(opts.Source))
	}
	b.WriteString("package PKG\n\nimport tea \"charm.land/bubbletea/v2\"\n\n")

	idents := make([]string, len(screens))
	for i, s := range screens {
		idents[i] = pascal(s.Name)
	}
	fmt.Fprintf(&b, "// %sPage names a page of the application.\ntype %sPage string\n\n// The pages of %s, in the order they were designed.\nconst (\n", app, app, doc.Name)
	for i, s := range screens {
		fmt.Fprintf(&b, "\t%sPage%s %sPage = %s\n", app, idents[i], app, strconv.Quote(s.Name))
	}
	b.WriteString(")\n\n")

	fmt.Fprintf(&b, "// %sPages are the pages in order, for a menu or a tab bar of the program's own.\nvar %sPages = []%sPage{", app, app, app)
	for i := range screens {
		if i > 0 {
			b.WriteString(", ")
		}
		fmt.Fprintf(&b, "%sPage%s", app, idents[i])
	}
	b.WriteString("}\n\n")

	fmt.Fprintf(&b, "// %sProps is everything the pages take from the program, and which page is\n// shown.\ntype %sProps struct {\n\t// Page is the page shown: set it to go to another, read it to know where the\n\t// person is.\n\tPage %sPage\n", app, app, app)
	for i, s := range screens {
		fmt.Fprintf(&b, "\t// %s holds the inputs of the page %s.\n\t%s %sProps\n", pageField(s.Name), strconv.Quote(s.Name), pageField(s.Name), idents[i])
	}
	b.WriteString("}\n\n")

	fmt.Fprintf(&b, "// Default%sProps is the application as it was designed: the first page is\n// shown and every page holds the values it had.\nfunc Default%sProps() %sProps {\n\treturn %sProps{\n\t\tPage: %sPage%s,\n", app, app, app, app, app, idents[0])
	for i, s := range screens {
		fmt.Fprintf(&b, "\t\t%s: Default%sProps(),\n", pageField(s.Name), idents[i])
	}
	b.WriteString("\t}\n}\n\n")

	fmt.Fprintf(&b, "// %s draws the page that is shown in an area of w by h cells.\nfunc %s(p %sProps, w, h int) Frame {\n\tswitch p.Page {\n", app, app, app)
	for i, s := range screens[1:] {
		fmt.Fprintf(&b, "\tcase %sPage%s:\n\t\treturn %s(p.%s, w, h)\n", app, idents[i+1], idents[i+1], pageField(s.Name))
	}
	fmt.Fprintf(&b, "\t}\n\treturn %s(p.%s, w, h)\n}\n\n", idents[0], pageField(screens[0].Name))

	fmt.Fprintf(&b, "// %sSize is the size of area the first page was designed for.\nfunc %sSize() (w, h int) { return %sSize() }\n\n", app, app, idents[0])

	fmt.Fprintf(&b, "// %sEvent is what happened on the page that is shown: the page's own event\n// (a %s... struct of that page), or nothing when the click only went to\n// another page.\ntype %sEvent struct {\n\t// Page is the page it happened on.\n\tPage %sPage\n\t// Event is the page's event; the go to events are not raised.\n\tEvent any\n\t// Went is true when the click changed the page, and To is the new one.\n\tWent bool\n\tTo   %sPage\n}\n\n", app, "<Page>", app, app, app)

	fmt.Fprintf(&b, "// %sHandle turns a click or a key into what happened on the page that is\n// shown. A click on a component that goes to a page changes p.Page. f is the\n// frame %s last returned.\nfunc %sHandle(p *%sProps, f Frame, msg tea.Msg) (%sEvent, bool) {\n\tswitch p.Page {\n", app, app, app, app, app)
	for i, s := range screens {
		fmt.Fprintf(&b, "\tcase %sPage%s:\n\t\tev, ok := %sHandle(f, msg)\n\t\tif !ok {\n\t\t\treturn %sEvent{}, false\n\t\t}\n", app, idents[i], idents[i], app)
		targets := gotoTargets(s)
		if len(targets) > 0 {
			b.WriteString("\t\tswitch ev.(type) {\n")
			for _, to := range targets {
				j := indexOfScreen(screens, to)
				if j < 0 {
					continue
				}
				fmt.Fprintf(&b, "\t\tcase %s%s:\n\t\t\tp.Page = %sPage%s\n\t\t\treturn %sEvent{Page: %sPage%s, Went: true, To: %sPage%s}, true\n",
					idents[i], pascal(gotoEvent(to)), app, idents[j], app, app, idents[i], app, idents[j])
			}
			b.WriteString("\t\t}\n")
		}
		fmt.Fprintf(&b, "\t\treturn %sEvent{Page: %sPage%s, Event: ev}, true\n", app, app, idents[i])
	}
	fmt.Fprintf(&b, "\t}\n\treturn %sEvent{}, false\n}\n", app)

	// Press: the containers of the pages that have some.
	var withPress []int
	for i, s := range screens {
		if hasContainer(s) {
			withPress = append(withPress, i)
		}
	}
	if len(withPress) > 0 {
		fmt.Fprintf(&b, "\n// %sPress lets the tabs and panels of the page that is shown answer a click.\n// It reports whether the click was one.\nfunc %sPress(p *%sProps, f Frame, msg tea.Msg) bool {\n\tswitch p.Page {\n", app, app, app)
		for _, i := range withPress {
			fmt.Fprintf(&b, "\tcase %sPage%s:\n\t\treturn %sPress(&p.%s, f, msg)\n", app, idents[i], idents[i], pageField(screens[i].Name))
		}
		b.WriteString("\t}\n\treturn false\n}\n")
	}
	return b.String()
}

// gotoTargets are the pages the components of a screen go to, in order and
// without repeats.
func gotoTargets(doc design.Document) []string {
	var out []string
	seen := map[string]bool{}
	var walk func(nodes []design.Node)
	walk = func(nodes []design.Node) {
		for _, n := range nodes {
			if n.Goto != "" && !n.Hidden && !seen[n.Goto] {
				seen[n.Goto] = true
				out = append(out, n.Goto)
			}
			walk(n.Children)
		}
	}
	walk(doc.Nodes)
	sort.Strings(out)
	return out
}

func indexOfScreen(screens []design.Document, name string) int {
	for i, s := range screens {
		if s.Name == name {
			return i
		}
	}
	return -1
}

func hasContainer(doc design.Document) bool {
	for _, n := range doc.Nodes {
		if n.IsContainer() {
			return true
		}
	}
	return false
}

// mainSource is the program that runs an application with pages: it keeps the
// props of the pages, draws the page that is shown, and lets tabs, panels and
// the components that go to a page work with no code of its own. It is a
// starting point: the screens are the part that comes from the design.
func mainSource(doc design.Document, screens []design.Document) string {
	app := pascal(doc.Name)
	if app == "" {
		app = "App"
	}
	first := pascal(screens[0].Name)
	return fmt.Sprintf(`%s

// main.go runs the application %s. Every page is a screen of this package
// (see the %s... files); %s.go is the router. Replace this program with
// your own when the application has behaviour: draw %s(props, w, h) and give
// %sHandle the messages.
package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
)

type model struct {
	props %sProps
	w, h  int
	frame Frame
}

func newModel() model {
	w, h := %sSize()
	return model{props: Default%sProps(), w: w, h: h}
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.w, m.h = msg.Width, msg.Height
	case tea.KeyPressMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	}
	m.frame = %s(m.props, m.w, m.h)
	%s
	return m, nil
}

func (m model) View() tea.View {
	v := tea.NewView(%s(m.props, m.w, m.h).View)
	v.AltScreen = true
	v.MouseMode = tea.MouseModeCellMotion
	return v
}

func main() {
	if _, err := tea.NewProgram(newModel()).Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
`, GeneratedHeader, strconv.Quote(doc.Name), first, snake(doc.Name)+"_app", app, app, app, app, app, app, pressAndHandle(app, screens), app)
}

// pressAndHandle is the update's use of the router.
func pressAndHandle(app string, screens []design.Document) string {
	var press string
	for _, s := range screens {
		if hasContainer(s) {
			press = fmt.Sprintf("if %sPress(&m.props, m.frame, msg) {\n\t\treturn m, nil\n\t}\n\t", app)
			break
		}
	}
	return press + fmt.Sprintf("%sHandle(&m.props, m.frame, msg)", app)
}
