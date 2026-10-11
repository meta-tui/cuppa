package gosource

import (
	"encoding/base64"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"sort"
	"strconv"
	"strings"

	"github.com/meta-ui-labs/cuppa/libs/document/design"
)

// Layouts of an exported application.
const (
	// LayoutFlat is every screen and the runtime in one package (the default).
	LayoutFlat = "flat"
	// LayoutSlices is a folder and a package per page, an app router, the
	// runtime in internal/cuppa and main.go at the module root (ADR 0013).
	LayoutSlices = "slices"
)

// goKeywords cannot be a package name.
var goKeywords = map[string]bool{
	"break": true, "case": true, "chan": true, "const": true, "continue": true, "default": true, "defer": true,
	"else": true, "fallthrough": true, "for": true, "func": true, "go": true, "goto": true, "if": true,
	"import": true, "interface": true, "map": true, "package": true, "range": true, "return": true,
	"select": true, "struct": true, "switch": true, "type": true, "var": true,
}

// pagePackages gives each screen the name of its package and folder: its
// words joined in lower case, with "page" added when that is a keyword or a
// name the application already has, or is taken by another page.
func pagePackages(screens []design.Document, hasMaster bool) []string {
	taken := map[string]bool{"app": true, "cuppa": true, "main": true, "tea": true, "internal": true, "components": true}
	if hasMaster {
		taken[masterPackage] = true
	}
	out := make([]string, len(screens))
	for i, s := range screens {
		name := strings.ToLower(strings.Join(words(s.Name), ""))
		if name == "" || name[0] >= '0' && name[0] <= '9' {
			name = "page" + name
		}
		if goKeywords[name] || taken[name] {
			name += "page"
		}
		for taken[name] {
			name += "x"
		}
		taken[name] = true
		out[i] = name
	}
	return out
}

// GenerateSlices exports a design as an application in the slice layout:
//
//	go.mod, main.go
//	app/app.go              the router
//	pages/<page>/<page>.go  one package per page
//	internal/cuppa/         the runtime they share
//
// The pages are the flat layout's screens with the names they use from the
// runtime exported and qualified, so what they draw is the same by
// construction (ADR 0013).
func GenerateSlices(doc design.Document, cat Catalog, opts Options) (Project, error) {
	module := opts.Module
	if module == "" {
		module = Slug(doc.Name)
	}
	screens := doc.Screens()
	// The master's parts are written once, as a package the pages take them
	// from (ADR 0015).
	shared, masterNotes := masterScreen(doc, cat)
	var whyNot []string
	flat := GenerateScreensWith(screens, cat, "cuppa", Options{master: shared, masterNotes: &whyNot})
	p := Project{Module: module, Files: map[string]string{}, Requires: flat.Requires, Notes: append(append(flat.Notes, masterNotes...), whyNot...)}

	// What each component file (a library of kinds) becomes: a package under
	// components/. The rest of the cuppa_ files is the runtime they share.
	componentPkg := map[string]string{}
	for _, e := range extensions {
		componentPkg["cuppa_"+e.file] = strings.TrimSuffix(strings.ReplaceAll(strings.TrimPrefix(e.file, "community_"), "_", ""), ".go")
	}
	kindPkg := map[string]string{}
	for kind, e := range extensions {
		kindPkg[kind] = componentPkg["cuppa_"+e.file]
	}
	runtime := map[string]string{}
	components := map[string]string{}
	for name, src := range flat.Files {
		if !strings.HasPrefix(name, "cuppa_") || strings.Contains(name, "_screen_") {
			continue
		}
		if _, ok := componentPkg[name]; ok {
			components[name] = src
		} else {
			runtime[name] = src
		}
	}
	sf, err := surfaceOf(runtime)
	if err != nil {
		return p, fmt.Errorf("slices: reading the runtime: %w", err)
	}
	// A component file may declare methods on the runtime's types; those
	// belong to the runtime's package.
	ownOf := map[string]map[string]bool{}
	for name, src := range components {
		rest, hoisted, err := hoistMethods(src, sf.tops)
		if err != nil {
			return p, fmt.Errorf("slices: component %s: %w", name, err)
		}
		components[name] = rest
		if hoisted != "" {
			runtime["cuppa_hoisted_"+componentPkg[name]+".go"] = hoisted
		}
		own, err := ownMembers(rest)
		if err != nil {
			return p, fmt.Errorf("slices: component %s: %w", name, err)
		}
		ownOf[name] = own
	}
	if sf, err = surfaceOf(runtime); err != nil {
		return p, fmt.Errorf("slices: reading the runtime: %w", err)
	}

	pkgs := pagePackages(screens, shared != nil)
	// A page with a folder of its own keeps its package name from the folder.
	dirs := make([]string, len(screens))
	pkgNames := make([]string, len(screens))
	for i, s := range screens {
		dirs[i], pkgNames[i] = "pages/"+pkgs[i], pkgs[i]
		if s.Folder != "" {
			dirs[i] = s.Folder
			pkgNames[i] = packageOfFolder(s.Folder)
		}
	}
	componentsDir, runtimeDir := "components", "internal/cuppa"
	if opts.Library {
		componentsDir = "internal/components"
	}
	use := &pageUse{tops: map[string]bool{}, members: map[string]bool{}, pkgs: map[string]bool{}, qualify: map[string]string{}}
	// The exported names a component package declares are used by pages as
	// pkg.Name (the colour picker's type, for one).
	for name, src := range components {
		csf, err := surfaceOf(map[string]string{name: src})
		if err != nil {
			return p, fmt.Errorf("slices: component %s: %w", name, err)
		}
		for top := range csf.tops {
			if !isUnexported(top) && !sf.tops[top] {
				use.qualify[top] = componentPkg[name]
			}
		}
	}
	hasContainers := make([]bool, len(screens))
	usesMaster := false
	for i, s := range screens {
		stem := snake(s.Name)
		view, contract := flat.Files[stem+"_screen_view.go"], flat.Files[stem+"_screen_contract.go"]
		if view == "" || contract == "" {
			return p, fmt.Errorf("slices: no screen was generated for the page %q (%v)", s.Name, flat.Notes)
		}
		// One file per page: the view's imports, then the contract's types.
		body := contract
		if at := strings.Index(body, "package cuppa"); at >= 0 {
			body = body[at+len("package cuppa"):]
		}
		combined := strings.TrimRight(view, "\n") + "\n" + body
		events := map[string]bool{}
		for _, l := range leaves(s, cat) {
			if l.Event != "" {
				events[pascal(l.Event)] = true
			}
		}
		for _, k := range s.Keys {
			events[pascal(k.Event)] = true
		}
		use.pkgs = map[string]bool{}
		out, err := rewritePage(combined, pkgNames[i], "cuppa", sf, pageNamer{ident: pascal(s.Name), lower: camel(s.Name), events: events}, use)
		pageUses := use.pkgs
		takesMaster := strings.Contains(combined, masterPackage+".PartsFor(")
		if takesMaster {
			pageUses[masterPackage] = true
			usesMaster = true
		}
		if err != nil {
			return p, fmt.Errorf("slices: page %q: %w", s.Name, err)
		}
		out = addImport(out, "cuppa", module+"/"+runtimeDir)
		out = addRuntimeAliases(out, sf)
		// The page imports the packages of the components it uses, so only
		// those are in the application and their kinds are registered.
		needed := map[string]bool{}
		for _, pkg := range componentsUsed(leaves(s, cat), kindPkg) {
			needed[pkg] = true
		}
		for pkg := range pageUses {
			needed[pkg] = true
		}
		var order []string
		for pkg := range needed {
			order = append(order, pkg)
		}
		sort.Strings(order)
		for _, pkg := range order {
			alias := "_"
			if pageUses[pkg] {
				alias = pkg
			}
			out = addImport(out, alias, module+"/"+componentsDir+"/"+pkg)
		}
		p.Files[dirs[i]+"/"+pkgNames[i]+generatedSuffix] = tidy(out, &p)
		hasContainers[i] = hasContainer(s)
	}

	if shared != nil && usesMaster {
		out, err := rewritePage(shared.source(), masterPackage, "cuppa", sf, pageNamer{ident: masterIdent, lower: masterLower, events: shared.events}, use)
		if err != nil {
			return p, fmt.Errorf("slices: the master: %w", err)
		}
		out = addImport(out, "cuppa", module+"/"+runtimeDir)
		p.Files[componentsDir+"/"+masterPackage+"/"+masterPackage+generatedSuffix] = tidy(out, &p)
	}

	componentFiles := make([]string, 0, len(components))
	for name := range components {
		componentFiles = append(componentFiles, name)
	}
	sort.Strings(componentFiles)
	for _, name := range componentFiles {
		pkg := componentPkg[name]
		out, err := rewritePage(components[name], pkg, "cuppa", sf, pageNamer{off: true, own: ownOf[name]}, use)
		if err != nil {
			return p, fmt.Errorf("slices: component %s: %w", pkg, err)
		}
		out = addImport(out, "cuppa", module+"/"+runtimeDir)
		p.Files[componentsDir+"/"+pkg+"/"+pkg+generatedSuffix] = tidy(out, &p)
	}

	names := map[string]bool{}
	for n := range use.tops {
		names[n] = true
	}
	for n := range use.members {
		names[n] = true
	}
	for name, src := range exportRuntime(runtime, names) {
		p.Files[runtimeDir+"/"+strings.TrimPrefix(name, "cuppa_")] = tidy(src, &p)
	}

	if !opts.Library || opts.Router {
		p.Files["app/app.go"] = tidy(sliceRouter(doc, screens, pkgs, dirs, hasContainers, module, opts.Source), &p)
	}
	if !opts.Library {
		p.Files["main.go"] = tidy(sliceMain(doc, module, hasContainers), &p)
		p.Files["go.mod"] = goMod(module, p.Requires)
		p.Files["README.md"] = sliceReadme(doc, screens, pkgs, dirs, p.Notes)
	}
	return p, nil
}

// addImport adds an import to the first import block of a source.
func addImport(src, name, path string) string {
	line := "\t" + name + " " + strconv.Quote(path) + "\n"
	if at := strings.Index(src, "import (\n"); at >= 0 {
		at += len("import (\n")
		return src[:at] + line + src[at:]
	}
	// No import block: add one after the package clause.
	pkg := strings.Index(src, "package ")
	at := pkg + strings.Index(src[pkg:], "\n")
	return src[:at+1] + "\nimport " + name + " " + strconv.Quote(path) + "\n" + src[at+1:]
}

// sliceRouter is app/app.go: the props of the pages and the page shown.
func sliceRouter(doc design.Document, screens []design.Document, pkgs, dirs []string, containers []bool, module string, source []byte) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s\n\n", GeneratedHeader)
	if len(source) > 0 {
		fmt.Fprintf(&b, "%s%s\n\n", DesignDirective, base64.StdEncoding.EncodeToString(source))
	}
	fmt.Fprintf(&b, "// Package app is the router of %s: the props of every page and which one is\n// shown. Pages go to each other by a click that names a page; the router\n// changes the page, so no page imports another.\npackage app\n\n", strconv.Quote(doc.Name))
	b.WriteString("import (\n\ttea \"charm.land/bubbletea/v2\"\n\n\t\"" + module + "/internal/cuppa\"\n")
	for i, pkg := range pkgs {
		fmt.Fprintf(&b, "\t%s \"%s/%s\"\n", pkg, module, dirs[i])
	}
	b.WriteString(")\n\n")

	idents := make([]string, len(screens))
	for i, s := range screens {
		idents[i] = pascal(s.Name)
	}
	fmt.Fprintf(&b, "// Page names a page of the application.\ntype Page string\n\n// The pages, in the order they were designed.\nconst (\n")
	for i, s := range screens {
		fmt.Fprintf(&b, "\tPage%s Page = %s\n", idents[i], strconv.Quote(s.Name))
	}
	b.WriteString(")\n\n// Pages are the pages in order, for a menu or a tab bar of the program's own.\nvar Pages = []Page{")
	for i := range screens {
		if i > 0 {
			b.WriteString(", ")
		}
		fmt.Fprintf(&b, "Page%s", idents[i])
	}
	b.WriteString("}\n\n")

	b.WriteString("// Props is everything the pages take from the program, and which page is shown.\ntype Props struct {\n\t// Page is the page shown: set it to go to another, read it to know where the\n\t// person is.\n\tPage Page\n")
	for i, s := range screens {
		fmt.Fprintf(&b, "\t// %s holds the inputs of the page %s.\n\t%s %s.Props\n", pageField(s.Name), strconv.Quote(s.Name), pageField(s.Name), pkgs[i])
	}
	b.WriteString("}\n\n")

	fmt.Fprintf(&b, "// Default is the application as it was designed: the first page is shown and\n// every page holds the values it had.\nfunc Default() Props {\n\treturn Props{\n\t\tPage: Page%s,\n", idents[0])
	for i, s := range screens {
		fmt.Fprintf(&b, "\t\t%s: %s.Default(),\n", pageField(s.Name), pkgs[i])
	}
	b.WriteString("\t}\n}\n\n")

	b.WriteString("// View draws the page that is shown in an area of w by h cells.\nfunc View(p Props, w, h int) cuppa.Frame {\n\tswitch p.Page {\n")
	for i := 1; i < len(screens); i++ {
		fmt.Fprintf(&b, "\tcase Page%s:\n\t\treturn %s.View(p.%s, w, h)\n", idents[i], pkgs[i], pageField(screens[i].Name))
	}
	fmt.Fprintf(&b, "\t}\n\treturn %s.View(p.%s, w, h)\n}\n\n", pkgs[0], pageField(screens[0].Name))
	fmt.Fprintf(&b, "// Size is the size of area the first page was designed for.\nfunc Size() (w, h int) { return %s.Size() }\n\n", pkgs[0])

	b.WriteString("// Event is what happened on the page that is shown: the page's own event (a\n// struct of that page's package), or nothing when the click only went to\n// another page.\ntype Event struct {\n\t// Page is the page it happened on.\n\tPage Page\n\t// Event is the page's event; the go to events are not raised.\n\tEvent any\n\t// Went is true when the click changed the page, and To is the new one.\n\tWent bool\n\tTo   Page\n}\n\n")

	b.WriteString("// Handle turns a click or a key into what happened on the page that is shown.\n// A click on a component that goes to a page changes p.Page. f is the frame\n// View last returned.\nfunc Handle(p *Props, f cuppa.Frame, msg tea.Msg) (Event, bool) {\n\tswitch p.Page {\n")
	for i, s := range screens {
		fmt.Fprintf(&b, "\tcase Page%s:\n\t\tev, ok := %s.Handle(f, msg)\n\t\tif !ok {\n\t\t\treturn Event{}, false\n\t\t}\n", idents[i], pkgs[i])
		if targets := gotoTargets(s); len(targets) > 0 {
			b.WriteString("\t\tswitch ev.(type) {\n")
			for _, to := range targets {
				j := indexOfScreen(screens, to)
				if j < 0 {
					continue
				}
				fmt.Fprintf(&b, "\t\tcase %s.%sEvent:\n\t\t\tp.Page = Page%s\n\t\t\treturn Event{Page: Page%s, Went: true, To: Page%s}, true\n",
					pkgs[i], pascal(gotoEvent(to)), idents[j], idents[i], idents[j])
			}
			b.WriteString("\t\t}\n")
		}
		fmt.Fprintf(&b, "\t\treturn Event{Page: Page%s, Event: ev}, true\n", idents[i])
	}
	b.WriteString("\t}\n\treturn Event{}, false\n}\n")

	var withPress []int
	for i, c := range containers {
		if c {
			withPress = append(withPress, i)
		}
	}
	if len(withPress) > 0 {
		b.WriteString("\n// Press lets the tabs and panels of the page that is shown answer a click. It\n// reports whether the click was one.\nfunc Press(p *Props, f cuppa.Frame, msg tea.Msg) bool {\n\tswitch p.Page {\n")
		for _, i := range withPress {
			fmt.Fprintf(&b, "\tcase Page%s:\n\t\treturn %s.Press(&p.%s, f, msg)\n", idents[i], pkgs[i], pageField(screens[i].Name))
		}
		b.WriteString("\t}\n\treturn false\n}\n")
	}
	return b.String()
}

// sliceMain is main.go: a program that runs the application.
func sliceMain(doc design.Document, module string, containers []bool) string {
	press := ""
	for _, c := range containers {
		if c {
			press = "if app.Press(&m.props, f, msg) {\n\t\treturn m, nil\n\t}\n\t"
			break
		}
	}
	return fmt.Sprintf(`%s

// main.go runs the application %s. Every page is a package under pages/;
// app/ is the router. Replace this program with your own when the application
// has behaviour: draw app.View(props, w, h) and give app.Handle the messages.
package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"

	"%s/app"
)

type model struct {
	props app.Props
	w, h  int
}

func newModel() model {
	w, h := app.Size()
	return model{props: app.Default(), w: w, h: h}
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
	f := app.View(m.props, m.w, m.h)
	%sapp.Handle(&m.props, f, msg)
	return m, nil
}

func (m model) View() tea.View {
	v := tea.NewView(app.View(m.props, m.w, m.h).View)
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
`, GeneratedHeader, strconv.Quote(doc.Name), module, press)
}

func sliceReadme(doc design.Document, screens []design.Document, pkgs, dirs []string, notes []string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# %s\n\nA Bubble Tea v2 application generated by Cuppa, one package per page.\n\n", doc.Name)
	b.WriteString("```sh\ngo mod tidy\ngo run .\n```\n\n")
	b.WriteString("| Folder | What it is |\n|---|---|\n")
	b.WriteString("| `main.go` | the program that runs the application |\n")
	b.WriteString("| `app/` | the router: `Props`, `Page`, `View`, `Handle` |\n")
	for i, s := range screens {
		fmt.Fprintf(&b, "| `%s/` | the page %q: `Props`, `Default`, `View`, `Handle` |\n", dirs[i], s.Name)
	}
	b.WriteString("| `internal/cuppa/` | the runtime the pages share |\n\n")
	b.WriteString("Files Cuppa wrote start with `// Code generated by Cuppa`; add your own files beside them (a page's\nupdate logic, tests) and they are left alone when the application is exported again.\nOpen the design again with `cuppa main.go`.\n")
	if len(notes) > 0 {
		sort.Strings(notes)
		b.WriteString("\n## Not generated yet\n\n")
		for _, n := range notes {
			b.WriteString("- " + n + "\n")
		}
	}
	return b.String()
}

// componentsUsed are the packages of the components the leaves are, in name
// order and without repeats.
func componentsUsed(ls []leaf, kindPkg map[string]string) []string {
	seen := map[string]bool{}
	var out []string
	for _, l := range ls {
		if pkg, ok := kindPkg[l.Kind]; ok && !seen[pkg] {
			seen[pkg] = true
			out = append(out, pkg)
		}
	}
	sort.Strings(out)
	return out
}

// packageOfFolder is the package name of a page kept in folder: its last
// element in lower case, with "page" added when that is a keyword.
func packageOfFolder(folder string) string {
	name := folder[strings.LastIndex(folder, "/")+1:]
	var b strings.Builder
	for _, r := range strings.ToLower(name) {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	name = b.String()
	if name == "" || name[0] >= '0' && name[0] <= '9' {
		name = "page" + name
	}
	if goKeywords[name] {
		name += "page"
	}
	return name
}

// aliasedTypes and aliasedRowStyles are the names of the runtime a page's
// Props and Frame use. Each page declares them as aliases, so the program that
// holds the page names pages/<page>.CellStyle and never imports the runtime.
var (
	aliasedTypes     = []string{"CellStyle", "RowStyle", "Frame", "Region", "Theme"}
	aliasedRowStyles = []string{"RowNormal", "RowSelected", "RowDim", "RowAccent"}
)

// addRuntimeAliases appends the aliases of the runtime names the page does
// not declare itself.
func addRuntimeAliases(src string, sf surface) string {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "page.go", src, parser.SkipObjectResolution)
	if err != nil {
		return src
	}
	declared := map[string]bool{}
	for _, d := range f.Decls {
		switch d := d.(type) {
		case *ast.FuncDecl:
			if d.Recv == nil {
				declared[d.Name.Name] = true
			}
		case *ast.GenDecl:
			for _, spec := range d.Specs {
				switch spec := spec.(type) {
				case *ast.TypeSpec:
					declared[spec.Name.Name] = true
				case *ast.ValueSpec:
					for _, n := range spec.Names {
						declared[n.Name] = true
					}
				}
			}
		}
	}
	var types, rows []string
	for _, n := range aliasedTypes {
		if sf.tops[n] && !declared[n] {
			types = append(types, "\t"+n+" = cuppa."+n)
		}
	}
	for _, n := range aliasedRowStyles {
		if sf.tops[n] && !declared[n] {
			rows = append(rows, "\t"+n+" = cuppa."+n)
		}
	}
	var b strings.Builder
	b.WriteString(strings.TrimRight(src, "\n") + "\n")
	if len(types) > 0 {
		b.WriteString("\n// The runtime's types a page's Props and Frame are made of, so the program that\n")
		b.WriteString("// holds the page does not import the runtime.\ntype (\n" + strings.Join(types, "\n") + "\n)\n")
	}
	if len(rows) > 0 {
		b.WriteString("\n// The styles of a row.\nconst (\n" + strings.Join(rows, "\n") + "\n)\n")
	}
	return b.String()
}
