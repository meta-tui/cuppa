// Package scene paints a whole document onto a cell grid.
package scene

import (
	"github.com/meta-ui-labs/cuppa/libs/document/container"
	"github.com/meta-ui-labs/cuppa/libs/catalog/definition"
	"github.com/meta-ui-labs/cuppa/libs/document/design"
	"github.com/meta-ui-labs/cuppa/libs/document/drawlayer"
	"github.com/meta-ui-labs/cuppa/libs/render/grid"
)

// Catalog is what the renderer needs to know about components.
type Catalog interface {
	Get(id string) (definition.Definition, bool)
}

type painter func(g *grid.Grid, p Props)

// painters maps a component id to the function that paints it.
var painters = map[string]painter{
	"draw.layer":     paintDrawLayer,
	"lipgloss.box":   paintBox,
	"lipgloss.label": paintLabel,
	"lipgloss.list":  paintList,
	"lipgloss.tabs":  paintTabs,
	"lipgloss.tree":  paintTree,
	"lipgloss.table": paintTable,
	"lipgloss.joinh": paintJoinH,
	"lipgloss.joinv": paintJoinV,
	"lipgloss.place": paintPlace,

	"bubbles.textinput":  paintTextInput,
	"bubbles.textarea":   paintTextArea,
	"bubbles.list":       paintBubbleList,
	"bubbles.table":      paintBubbleTable,
	"bubbles.tree":       paintBubbleTree,
	"bubbles.viewport":   paintViewport,
	"bubbles.paginator":  paintPaginator,
	"bubbles.filepicker": paintFilePicker,
	"bubbles.spinner":    paintSpinner,
	"bubbles.progress":   paintProgress,
	"bubbles.timer":      paintTimer,
	"bubbles.stopwatch":  paintStopwatch,
	"bubbles.help":       paintHelp,

	"huh.input":       paintHuhInput,
	"huh.text":        paintHuhText,
	"huh.select":      paintHuhSelect,
	"huh.multiselect": paintHuhMultiSelect,
	"huh.confirm":     paintHuhConfirm,
	"huh.note":        paintHuhNote,
	"huh.spinner":     paintHuhSpinner,
	"huh.filepicker":  paintHuhFilePicker,
	"huh.form":        paintHuhForm,

	"glamour.markdown": paintMarkdown,

	"ntcharts.barchart":   paintBarChart,
	"ntcharts.linechart":  paintLineChart,
	"ntcharts.sparkline":  paintSparkline,
	"ntcharts.streamline": paintStreamline,
	"ntcharts.timeseries": paintTimeSeries,
	"ntcharts.heatmap":    paintHeatMap,
	"ntcharts.canvas":     paintChartCanvas,

	"community.bubbletable":   paintBubbleTableCommunity,
	"community.flexbox":       paintFlexBox,
	"community.boxer":         paintBoxer,
	"community.datepicker":    paintDatePicker,
	"community.overlay":       paintOverlay,
	"community.statusbar":     paintStatusBar,
	"community.filetree":      paintFileTree,
	"community.frame":         paintTitledFrame,
	"community.dialog":        paintDialog,
	"community.statusmessage": paintStatusMessage,
	"lipgloss.swatch":         paintSwatch,
	"community.bigtext":      paintBigText,
	"community.qrcode":       paintQRCode,
	"community.image":        paintImage,
	"community.toast":         paintToast,
	"lipgloss.rows":           paintRows,
	"lipgloss.keybar":         paintKeybar,
	"lipgloss.progressline":   paintProgressLine,
	"lipgloss.slot":           paintSlot,
	"cuppa.tabs":              paintTabPages,
	"cuppa.panels":            paintPanelStack,
	"cuppa.menubar":           paintMenuBar,
	"lipgloss.colourpicker":   paintColourPicker,
	"community.dropdown":      paintDropdown,
	"community.promptinput":  paintPromptInput,
	"community.promptselect": paintPromptSelect,
	"community.datatree":     paintDataTree,
	"community.pdfview":      paintPDFView,
	"ntcharts.chart3d":       paintChart3D,
}

// Render paints every visible node of doc, back to front, onto a new grid the
// size of the canvas, then applies the document's background and effects.
// Hidden nodes are skipped, so they are missing from every export too. Cells no
// node covers stay unpainted (a zero character) but carry the background.
func Render(doc design.Document, cat Catalog) *grid.Grid { return RenderWith(doc, cat, nil) }

// RenderWith is Render, except that a node for which override returns a grid
// is drawn from that grid instead of by its painter. The preview uses it to
// show components that are running as real models. A nil override changes
// nothing.
func RenderWith(doc design.Document, cat Catalog, override func(design.Node) *grid.Grid) *grid.Grid {
	out := grid.New(doc.Width, doc.Height)
	if doc.Background != "" {
		for y := 0; y < out.H; y++ {
			for x := 0; x < out.W; x++ {
				out.Set(x, y, grid.Cell{Style: grid.Style{Bg: doc.Background}})
			}
		}
	}
	for _, i := range container.Order(doc) {
		n := doc.Nodes[i]
		if n.Hidden || !container.Visible(doc, n) {
			continue
		}
		if doc.Effects.Shadow && n.Component != drawlayer.Component {
			castShadow(out, n.Rect, doc.Background)
		}
		var drawn *grid.Grid
		if override != nil {
			drawn = override(n)
		}
		if drawn == nil {
			drawn = renderNode(n, cat, 0, doc.Theme, doc.Background)
		}
		out.Blit(drawn, n.Rect.X, n.Rect.Y)
	}
	paintBackground(out, doc.Background)
	applyEffects(out, doc.Effects)
	applyProfile(out, doc.Profile)
	return out
}

// RenderNode paints one node onto its own grid, sized like the node, with no
// theme: a colour the node does not set shows the component's default.
func RenderNode(n design.Node, cat Catalog) *grid.Grid {
	return renderNode(n, cat, 0, design.Theme{}, "")
}

// RenderNodeThemed is RenderNode with the design's theme: a colour the node
// does not set follows the theme.
func RenderNodeThemed(n design.Node, cat Catalog, theme design.Theme, background string) *grid.Grid {
	return renderNode(n, cat, 0, theme, background)
}

func renderNode(n design.Node, cat Catalog, depth int, theme design.Theme, background string) *grid.Grid {
	g := grid.New(n.Rect.W, n.Rect.H)
	def, known := cat.Get(n.Component)
	props := Props{}
	if known {
		for k, v := range def.Effective(n.Props, theme, background) {
			props[k] = v
		}
	} else {
		for k, v := range n.Props {
			props[k] = v
		}
	}
	if n.IsGroup() {
		g.Fill(design.Rect{W: g.W, H: g.H}, ' ', grid.Style{})
		paintComposite(g, design.Composite{ID: "group", Name: n.Name, W: n.BaseW, H: n.BaseH, Nodes: n.Children}, props, cat, depth, theme, background)
		return g
	}
	if known && def.Inner != nil {
		g.Fill(design.Rect{W: g.W, H: g.H}, ' ', grid.Style{})
		paintComposite(g, *def.Inner, props, cat, depth, theme, background)
		return g
	}
	if n.Component == drawlayer.Component {
		// The drawing is transparent where nothing was painted, so it does
		// not clear its rectangle first.
		paintDrawLayer(g, props)
		return g
	}
	if p, ok := painters[n.Component]; ok {
		// Every painter owns the whole node rectangle: clear it first so
		// nodes below never show through.
		g.Fill(design.Rect{W: g.W, H: g.H}, ' ', grid.Style{})
		p(g, props)
		return g
	}
	g.Fill(design.Rect{W: g.W, H: g.H}, ' ', grid.Style{})
	name := n.Name
	if known {
		name = def.Name
	}
	paintGeneric(g, name)
	return g
}

// Painted reports whether the component has a dedicated painter.
func Painted(componentID string) bool {
	_, ok := painters[componentID]
	return ok
}
