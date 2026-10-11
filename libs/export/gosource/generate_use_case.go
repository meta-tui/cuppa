package gosource

import (
	"encoding/base64"
	_ "embed"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/meta-tui/cuppa/libs/document/design"
	"github.com/meta-tui/cuppa/libs/layout/expr"
)

//go:embed runtime_go.txt
var runtimeBase string

//go:embed containers_logic_go.txt
var containersLogic string

// runtimeSource is the runtime, with the arithmetic of containers that the
// model uses to show a page and move its members.
var runtimeSource = runtimeBase + "\n" + containersLogic

//go:embed app_go.txt
var appSource string

//go:embed widgets_go.txt
var widgetsSource string

//go:embed bigtext_go.txt
var bigTextSource string

//go:embed qrcode_go.txt
var qrCodeSource string

//go:embed image_go.txt
var imageSource string

//go:embed community_widgets_go.txt
var communityWidgetsBase string

//go:embed containers_go.txt
var containersSource string

// communityWidgetsSource is community_widgets.go: the widgets and the
// container logic they share the grid with.
var communityWidgetsSource = communityWidgetsBase + "\n" + containersSource

//go:embed huh_go.txt
var huhSource string

//go:embed glamour_go.txt
var glamourSource string

//go:embed charts_go.txt
var chartsSource string

//go:embed bubbletable_go.txt
var bubbleTableSource string

// generated lists the components that have real code behind them. The
// community widgets are drawn with Lip Gloss to look like the library they
// stand for, so the generated project needs no extra module for them.
var generated = map[string]bool{
	"lipgloss.box": true, "lipgloss.label": true, "lipgloss.list": true, "lipgloss.swatch": true, "lipgloss.rows": true, "lipgloss.colourpicker": true, "lipgloss.keybar": true, "lipgloss.slot": true, "lipgloss.progressline": true, design.TabsComponent: true, design.PanelsComponent: true,
	"bubbles.textinput": true, "bubbles.textarea": true, "bubbles.list": true, "bubbles.table": true,
	"bubbles.viewport": true, "bubbles.paginator": true, "bubbles.spinner": true, "bubbles.progress": true,
	"bubbles.stopwatch": true, "bubbles.timer": true, "bubbles.tree": true,
	"huh.spinner":     true,
	"community.frame": true, "community.dialog": true, "community.statusmessage": true, "community.toast": true,
	"community.bigtext": true, "community.qrcode": true, "community.image": true,
	"lipgloss.table": true, "lipgloss.tree": true, "lipgloss.tabs": true,
	"lipgloss.joinh": true, "lipgloss.joinv": true, "lipgloss.place": true,
	"bubbles.help": true, "bubbles.filepicker": true,
	"community.flexbox": true, "community.boxer": true, "community.datepicker": true,
	"community.overlay": true, "community.statusbar": true, "community.filetree": true,
	"huh.input": true, "huh.text": true, "huh.select": true, "huh.multiselect": true,
	"huh.confirm": true, "huh.note": true, "huh.filepicker": true, "huh.form": true,
	"glamour.markdown":   true,
	"ntcharts.sparkline": true, "ntcharts.barchart": true, "ntcharts.linechart": true,
	"ntcharts.streamline": true, "ntcharts.timeseries": true, "ntcharts.heatmap": true, "ntcharts.canvas": true,
	"community.bubbletable": true,
	"community.dropdown":    true, "community.promptinput": true, "community.promptselect": true,
	"community.datatree":    true, "community.pdfview": true, "ntcharts.chart3d": true,
	// drawRun is what the drawing layer is expanded into.
	drawRunKind: true,
}

// extension is a component that needs a file of its own and, for some, a
// module the rest of the project does not use. Several components can share a
// file; the file and its module are added once.
type extension struct {
	file, text string
	// require is a go.mod requirement line, or empty for the standard library.
	require string
}

const (
	figletRequire  = "github.com/common-nighthawk/go-figure v0.0.0-20210622060536-734e95fb86be"
	qrRequire      = "github.com/skip2/go-qrcode v0.0.0-20200617195104-da1b6568686e"
	huhRequire     = "charm.land/huh/v2 v2.0.3"
	glamourRequire = "charm.land/glamour/v2 v2.0.1"
	chartsRequire  = "github.com/NimbleMarkets/ntcharts/v2 v2.7.2"
	tableRequire   = "github.com/evertras/bubble-table v0.23.0"
)

var extensions = map[string]extension{
	"community.bigtext": {"bigtext.go", bigTextSource, figletRequire},
	"community.qrcode":  {"qrcode.go", qrCodeSource, qrRequire},
	"community.image":   {"image.go", imageSource, ""},

	"lipgloss.rows":          {"community_widgets.go", communityWidgetsSource, ""},
	"lipgloss.keybar":        {"community_widgets.go", communityWidgetsSource, ""},
	"lipgloss.slot":          {"community_widgets.go", communityWidgetsSource, ""},
	"lipgloss.progressline":  {"community_widgets.go", communityWidgetsSource, ""},
	design.TabsComponent:     {"community_widgets.go", communityWidgetsSource, ""},
	design.PanelsComponent:   {"community_widgets.go", communityWidgetsSource, ""},
	"lipgloss.colourpicker":  {"colourpicker.go", colourPickerSource(), ""},
	"community.dropdown":     {"community_widgets.go", communityWidgetsSource, ""},
	"community.promptinput":  {"community_widgets.go", communityWidgetsSource, ""},
	"community.promptselect": {"community_widgets.go", communityWidgetsSource, ""},
	"community.datatree":     {"community_widgets.go", communityWidgetsSource, ""},
	"community.pdfview":      {"community_widgets.go", communityWidgetsSource, ""},
	"ntcharts.chart3d":       {"community_widgets.go", communityWidgetsSource, ""},

	"huh.input":       {"huh.go", huhSource, huhRequire},
	"huh.text":        {"huh.go", huhSource, huhRequire},
	"huh.select":      {"huh.go", huhSource, huhRequire},
	"huh.multiselect": {"huh.go", huhSource, huhRequire},
	"huh.confirm":     {"huh.go", huhSource, huhRequire},
	"huh.note":        {"huh.go", huhSource, huhRequire},
	"huh.filepicker":  {"huh.go", huhSource, huhRequire},

	"glamour.markdown": {"glamour.go", glamourSource, glamourRequire},

	"ntcharts.sparkline":  {"charts.go", chartsSource, chartsRequire},
	"ntcharts.barchart":   {"charts.go", chartsSource, chartsRequire},
	"ntcharts.linechart":  {"charts.go", chartsSource, chartsRequire},
	"ntcharts.streamline": {"charts.go", chartsSource, chartsRequire},
	"ntcharts.timeseries": {"charts.go", chartsSource, chartsRequire},
	"ntcharts.heatmap":    {"charts.go", chartsSource, chartsRequire},
	"ntcharts.canvas":     {"charts.go", chartsSource, chartsRequire},

	"community.bubbletable": {"bubbletable.go", bubbleTableSource, tableRequire},
}

// Generate turns doc into a Go project. The same document always gives the
// same files.
func Generate(doc design.Document, cat Catalog) Project {
	return GenerateWith(doc, cat, Options{})
}

// GenerateWith is Generate with options. A design with pages is an
// application: its screens, the router and a main.go that runs them.
func GenerateWith(doc design.Document, cat Catalog, opts Options) Project {
	if doc.IsPaged() {
		return generateApp(doc, cat, opts)
	}
	module := Slug(doc.Name)
	p := Project{Module: module, Files: map[string]string{}}
	placed := leaves(doc, cat)
	var requires []string
	files := map[string]bool{}
	needs := map[string]bool{}
	for _, l := range placed {
		e, ok := extensions[l.Kind]
		if !ok || files[e.file] {
			continue
		}
		files[e.file] = true
		p.Files[e.file] = e.text
		if e.require != "" && !needs[e.require] {
			needs[e.require] = true
			requires = append(requires, e.require)
		}
	}
	sort.Strings(requires)
	p.Files["go.mod"] = goMod(module, requires)
	p.Files["main.go"] = appSource
	p.Files["runtime.go"] = runtimeSource
	p.Files["widgets.go"] = widgetsSource
	p.Files["layout.go"] = layoutSource(doc, placed)
	if len(opts.Source) > 0 {
		// The design travels with the program, so Cuppa can open it from here.
		p.Files["layout.go"] = strings.Replace(p.Files["layout.go"], "package main\n\n", "package main\n\n"+DesignDirective+base64.StdEncoding.EncodeToString(opts.Source)+"\n\n", 1)
	}
	for _, l := range placed {
		if !generated[l.Kind] {
			p.Notes = append(p.Notes, fmt.Sprintf("%s (%s) is not generated yet: it is drawn as an empty frame", l.Name, l.Kind))
		}
	}
	for _, n := range doc.Nodes {
		if !n.Layout.IsZero() && !n.Hidden && (n.IsGroup() || isComposite(n, cat)) {
			p.Notes = append(p.Notes, fmt.Sprintf("%s is a group or pack component: its layout expressions are not followed in the program, only the size it has in the design", n.Name))
		}
	}
	p.Files["README.md"] = readme(doc, p.Notes)
	return p
}

// Slug turns a design name into a module name: lowercase letters, digits and
// dashes.
func Slug(name string) string {
	var b strings.Builder
	dash := false
	for _, r := range strings.ToLower(name) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			if dash && b.Len() > 0 {
				b.WriteByte('-')
			}
			dash = false
			b.WriteRune(r)
		} else {
			dash = true
		}
	}
	if b.Len() == 0 {
		return "cuppa-design"
	}
	return b.String()
}

func goMod(module string, requires []string) string {
	var extra strings.Builder
	for _, r := range requires {
		extra.WriteString("\t" + r + "\n")
	}
	return fmt.Sprintf("module %s\n\ngo 1.24\n\nrequire (\n\tcharm.land/bubbles/v2 %s\n\tcharm.land/bubbletea/v2 %s\n\tcharm.land/lipgloss/v2 %s\n%s)\n",
		module, bubblesVersion, bubbleteaVersion, lipglossVersion, extra.String())
}

// layoutSource is the part that comes from the design: the canvas and where
// each component sits, with its properties.
func layoutSource(doc design.Document, placed []leaf) string {
	var b strings.Builder
	b.WriteString("package main\n\n")
	b.WriteString("// Generated by Cuppa from the design " + strconv.Quote(doc.Name) + ".\n")
	b.WriteString("// Components are listed back to front.\n\n")
	fmt.Fprintf(&b, "const (\n\tcanvasW    = %d\n\tcanvasH    = %d\n\tbackground = %s\n\t// responsive is true when a component follows the size of the window.\n\tresponsive = %t\n)\n\n", doc.Width, doc.Height, strconv.Quote(doc.Background), followsWindow(placed))
	b.WriteString("var layout = []placed{\n")
	for _, l := range placed {
		fmt.Fprintf(&b, "\t{Kind: %s, Name: %s, X: %d, Y: %d, W: %d, H: %d, Props: map[string]string{", strconv.Quote(l.Kind), strconv.Quote(l.Name), l.Rect.X, l.Rect.Y, l.Rect.W, l.Rect.H)
		keys := make([]string, 0, len(l.Props))
		for k := range l.Props {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for i, k := range keys {
			if i > 0 {
				b.WriteString(", ")
			}
			fmt.Fprintf(&b, "%s: %s", strconv.Quote(k), strconv.Quote(l.Props[k]))
		}
		b.WriteString("}")
		b.WriteString(behaviourSource(l))
		b.WriteString("},\n")
	}
	b.WriteString("}\n")
	return b.String()
}

// behaviourSource is the fields of a placed component that come from how it
// is used rather than from where it sits: the function that follows the
// window, and the drag and resize the person using the program gets. It is
// empty for a plain fixed component.
func behaviourSource(l leaf) string {
	var parts []string
	if l.In != "" {
		// A member of a container: shown only while its page is.
		parts = append(parts, fmt.Sprintf("Member: true, Of: %d, Page: %d", l.Of, l.Page))
	}
	if l.Layout.IsZero() && !l.Drag && !l.Resize {
		if len(parts) == 0 {
			return ""
		}
		return ", " + strings.Join(parts, ", ")
	}
	if l.Drag {
		parts = append(parts, "Drag: true")
	}
	if l.Resize {
		parts = append(parts, "Resize: true")
	}
	parts = append(parts, fmt.Sprintf("MinW: %d, MinH: %d", l.MinW, l.MinH))
	if fit := fitSource(l); fit != "" {
		parts = append(parts, fit)
	}
	return ", " + strings.Join(parts, ", ")
}

// fitSource is the Fit field of a component whose size or position follows
// the window ("" for a fixed one).
func fitSource(l leaf) string {
	if l.Layout.IsZero() {
		return ""
	}
	axis := func(src string, fixed int, parent string) string {
		if src == "" {
			return strconv.Itoa(fixed)
		}
		e, err := expr.Parse(src)
		if err != nil || e.UsesEnv() {
			return strconv.Itoa(fixed)
		}
		return "cells(" + e.GoSource(parent) + ")"
	}
	return fmt.Sprintf("Fit: func(w, h int) (x, y, cw, ch int) {\n\t\treturn %s, %s, %s, %s\n\t}",
		axis(l.Layout.X, l.Rect.X, "w"), axis(l.Layout.Y, l.Rect.Y, "h"),
		axis(l.Layout.W, l.Rect.W, "w"), axis(l.Layout.H, l.Rect.H, "h"))
}

// layoutReadsEnv reports whether any axis of the component's layout reads an
// input or the place of another component.
func layoutReadsEnv(l leaf) bool {
	for _, src := range []string{l.Layout.X, l.Layout.Y, l.Layout.W, l.Layout.H} {
		if src == "" {
			continue
		}
		if e, err := expr.Parse(src); err == nil && e.UsesEnv() {
			return true
		}
	}
	return false
}

// fitEnvSource is the FitEnv field of a component of a screen whose place reads
// inputs or other components.
func fitEnvSource(l leaf) string {
	axis := func(src string, fixed int, parent string) string {
		if src == "" {
			return strconv.Itoa(fixed)
		}
		e, err := expr.Parse(src)
		if err != nil {
			return strconv.Itoa(fixed)
		}
		return "cells(" + e.GoSource(parent) + ")"
	}
	return fmt.Sprintf("FitEnv: func(e fitEnv) (x, y, cw, ch int) {\n\t\tw, h := e.W, e.H\n\t\t_, _ = w, h\n\t\treturn %s, %s, %s, %s\n\t}",
		axis(l.Layout.X, l.Rect.X, "w"), axis(l.Layout.Y, l.Rect.Y, "h"),
		axis(l.Layout.W, l.Rect.W, "w"), axis(l.Layout.H, l.Rect.H, "h"))
}

func readme(doc design.Document, notes []string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# %s\n\nA Bubble Tea v2 program generated by Cuppa.\n\n", doc.Name)
	b.WriteString("```sh\ngo mod tidy\ngo run .\n```\n\n")
	b.WriteString("Tab and Shift+Tab move between components that take input, clicks focus them, Esc quits.\n\n")
	b.WriteString("- `layout.go` is the design: where each component sits and its properties.\n")
	b.WriteString("- `runtime.go` wraps the real Bubbles models and draws boxes, labels, lists and the look of the community widgets with Lip Gloss. Change it freely.\n")
	b.WriteString("- `widgets.go` draws the Lip Gloss tables, trees and tabs, the help and file picker, and the look of the community widgets.\n")
	b.WriteString("- `huh.go`, `glamour.go`, `charts.go`, `bubbletable.go`, `bigtext.go`, `qrcode.go` and `image.go`, when present, hold the components that need a library of their own.\n")
	b.WriteString("- `main.go` is the Bubble Tea model that runs them.\n")
	if len(notes) > 0 {
		b.WriteString("\n## Not generated yet\n\n")
		for _, n := range notes {
			b.WriteString("- " + n + "\n")
		}
	}
	return b.String()
}

// followsWindow reports whether any placed component takes its place from the
// size of the window.
func followsWindow(placed []leaf) bool {
	for _, l := range placed {
		if !l.Layout.IsZero() {
			return true
		}
	}
	return false
}

