package design

import (
	"fmt"
	"strconv"
	"strings"
)

// Document is a design: a fixed-size canvas and the nodes placed on it.
// Nodes are ordered back to front, so the last node is drawn on top.
type Document struct {
	Name   string `json:"name"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
	Nodes  []Node `json:"nodes"`
	// Background is the canvas colour ("" for the terminal's own, "0" to "255"
	// or "#rrggbb"). It fills every cell no component colours itself.
	Background string `json:"background,omitempty"`
	// Theme holds the design's colours for the roles components take their
	// colours from: text, muted, border and secondary. Together with Background it is
	// the whole theme. A component colour that is not set on the component
	// follows the theme; one that is set stays, whatever the theme says.
	Theme Theme `json:"theme,omitzero"`
	// HideGrid turns off the dotted grid shown on an empty canvas in the editor.
	// The grid is never part of an export.
	HideGrid bool `json:"hideGrid,omitempty"`
	// Profile is how many colours the target terminal has: "" for true colour,
	// or "256", "16" or "none". Colours are reduced to it in the editor and in
	// every export, so the design shows how it will look there.
	Profile string `json:"profile,omitempty"`
	// Light previews the design on a light terminal in the editor. It is not
	// part of an export: an export keeps the terminal's own colours.
	Light bool `json:"light,omitempty"`
	Effects  Effects `json:"effects,omitempty"`
	// Keys are the keys of the screen: each raises an event when the design is
	// exported as a screen (ADR 0007).
	Keys []KeyBinding `json:"keys,omitempty"`
	// Embedded are copies of the user-made components the design uses, so it
	// still draws on a computer without their packs.
	Embedded []Embedded `json:"components,omitempty"`
	// Pages split the design into the screens of an application; Page is the
	// one being edited, whose components are Nodes. Empty for a design with
	// one page. See ADR 0012.
	// Samples give the inputs that only a layout expression or a condition
	// reads a value to preview with, so the designer and the exported
	// DefaultProps show the screen as it is meant to look: input name to a
	// number ("38"). An input a property is bound to keeps that property's
	// value.
	Samples map[string]string `json:"samples,omitempty"`
	// Defaults are the values the exported Default<Screen>Props starts from,
	// input name to text, where the application wants another than the one the
	// design shows (an optional part hidden, a layout-only input neutral). The
	// designer keeps previewing the design's own values.
	Defaults map[string]string `json:"defaults,omitempty"`
	Pages    []Page            `json:"pages,omitempty"`
	// PerPage says each page keeps its own samples and keys. A design saved
	// without it had one set of each for every page.
	PerPage bool `json:"perPage,omitempty"`
	// Export says where the Go source of the design goes; empty until the
	// person exports it.
	Export ExportSettings `json:"export,omitzero"`
	// Folder is the folder of the page this screen is, set by Screens; it is
	// not part of a design's file.
	Folder string `json:"-"`
	Page  int    `json:"page,omitempty"`
	// Seq is the counter behind generated node ids; it only ever grows.
	Seq int `json:"seq"`
}

// Theme is the colours of a design for the roles components can take colours
// from ("" for none set: the component's own default shows). The canvas colour
// is Document.Background.
type Theme struct {
	// Text is the colour of text.
	Text string `json:"text,omitempty"`
	// Muted is the colour of quieter text: placeholders, descriptions, labels
	// and hints.
	Muted string `json:"muted,omitempty"`
	// Border is the colour of frames and borders.
	Border string `json:"border,omitempty"`
	// Secondary is the accent colour: highlights, prompts, bars.
	Secondary string `json:"secondary,omitempty"`
	// Palette is the design's named colours. A colour property set to "@Name"
	// uses the swatch of that name, and a screen exported from the design
	// exposes each swatch so the program can change it (ADR 0007).
	Palette []Swatch `json:"palette,omitempty"`
}

// IsZero reports whether no theme colour and no palette colour is set.
func (t Theme) IsZero() bool {
	return t.Text == "" && t.Muted == "" && t.Border == "" && t.Secondary == "" && len(t.Palette) == 0
}

// Swatch is one named colour of a palette.
type Swatch struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

// TokenPrefix starts a colour property value that names a palette swatch.
const TokenPrefix = "@"

// Colour returns the colour of the swatch called name.
func (t Theme) Colour(name string) (string, bool) {
	for _, s := range t.Palette {
		if s.Name == name {
			return s.Color, true
		}
	}
	return "", false
}

// Resolve returns the colour a property value stands for: the swatch's colour
// for "@Name" (empty when the swatch does not exist), the value itself
// otherwise.
func (t Theme) Resolve(value string) string {
	name, ok := strings.CutPrefix(value, TokenPrefix)
	if !ok {
		return value
	}
	c, _ := t.Colour(name)
	return c
}

// Embedded is a copy of a pack component kept inside a design. ID is the
// catalog id the nodes use, such as "tea-shop.card".
type Embedded struct {
	ID        string    `json:"id"`
	Composite Composite `json:"component"`
}

// Colour profiles a document can target; the empty string is true colour.
const (
	Profile256  = "256"
	Profile16   = "16"
	ProfileNone = "none"
)

// ValidProfile reports whether p names a colour profile ("" included).
func ValidProfile(p string) bool {
	return p == "" || p == Profile256 || p == Profile16 || p == ProfileNone
}

// Effects are looks applied over the whole canvas, in the editor and in exports.
type Effects struct {
	// Shadow drops a one-cell shadow below and to the right of every component.
	Shadow bool `json:"shadow,omitempty"`
	// Scanlines dims every other row, like a CRT.
	Scanlines bool `json:"scanlines,omitempty"`
	// Vignette dims the edges of the canvas.
	Vignette bool `json:"vignette,omitempty"`
}

// Canvas size limits, in cells.
const (
	MaxWidth  = 400
	MaxHeight = 200
)

// NewDocument returns an empty document of the given canvas size.
func NewDocument(name string, width, height int) Document {
	return Document{Name: name, Width: width, Height: height}
}

// Bounds is the canvas as a rectangle.
func (d Document) Bounds() Rect { return Rect{W: d.Width, H: d.Height} }

// Clone returns a deep copy, so edits to it never reach the original.
func (d Document) Clone() Document {
	nodes := make([]Node, len(d.Nodes))
	for i, n := range d.Nodes {
		nodes[i] = n.Clone()
	}
	d.Nodes = nodes
	if d.Samples != nil {
		samples := make(map[string]string, len(d.Samples))
		for k, v := range d.Samples {
			samples[k] = v
		}
		d.Samples = samples
	}
	d.Defaults = cloneSamples(d.Defaults)
	if d.Pages != nil {
		pages := make([]Page, len(d.Pages))
		for i, p := range d.Pages {
			pages[i] = Page{Name: p.Name, Folder: p.Folder, Samples: cloneSamples(p.Samples), Defaults: cloneSamples(p.Defaults), Keys: append([]KeyBinding(nil), p.Keys...), Master: p.Master, Hide: append([]NodeID(nil), p.Hide...), Values: cloneSamples(p.Values)}
			for _, n := range p.Nodes {
				pages[i].Nodes = append(pages[i].Nodes, n.Clone())
			}
		}
		d.Pages = pages
	}
	d.Keys = append([]KeyBinding(nil), d.Keys...)
	d.Theme.Palette = append([]Swatch(nil), d.Theme.Palette...)
	if d.Embedded != nil {
		emb := make([]Embedded, len(d.Embedded))
		for i, e := range d.Embedded {
			emb[i] = Embedded{ID: e.ID, Composite: e.Composite.Clone()}
		}
		d.Embedded = emb
	}
	return d
}

// NewID returns an id no node has had in this document.
func (d *Document) NewID() NodeID {
	d.Seq++
	return NodeID(fmt.Sprintf("n%d", d.Seq))
}

// Add appends the node on top. A node without an id gets a fresh one.
// The stored node is returned.
func (d *Document) Add(n Node) Node {
	if n.ID == "" {
		n.ID = d.NewID()
	}
	d.Nodes = append(d.Nodes, n)
	return n
}

// Index is the position of the node in z-order (0 is the back), or -1.
func (d Document) Index(id NodeID) int {
	for i, n := range d.Nodes {
		if n.ID == id {
			return i
		}
	}
	return -1
}

// Get returns a copy of the node with the given id.
func (d Document) Get(id NodeID) (Node, bool) {
	if i := d.Index(id); i >= 0 {
		return d.Nodes[i], true
	}
	return Node{}, false
}

// Update applies fn to the node with the given id and reports whether it exists.
func (d *Document) Update(id NodeID, fn func(*Node)) bool {
	i := d.Index(id)
	if i < 0 {
		return false
	}
	fn(&d.Nodes[i])
	return true
}

// Remove deletes the node and reports whether it existed.
func (d *Document) Remove(id NodeID) bool {
	i := d.Index(id)
	if i < 0 {
		return false
	}
	d.Nodes = append(d.Nodes[:i], d.Nodes[i+1:]...)
	return true
}

// Raise moves the node one step toward the front.
func (d *Document) Raise(id NodeID) bool { return d.moveTo(id, d.Index(id)+1) }

// Lower moves the node one step toward the back.
func (d *Document) Lower(id NodeID) bool { return d.moveTo(id, d.Index(id)-1) }

// ToFront moves the node above every other node.
func (d *Document) ToFront(id NodeID) bool { return d.moveTo(id, len(d.Nodes)-1) }

// ToBack moves the node below every other node.
func (d *Document) ToBack(id NodeID) bool { return d.moveTo(id, 0) }

// MoveToIndex places the node at position to in z-order (0 is the back) and
// reports whether the order changed.
func (d *Document) MoveToIndex(id NodeID, to int) bool { return d.moveTo(id, to) }

// moveTo places the node at position to in z-order and reports whether the
// order changed.
func (d *Document) moveTo(id NodeID, to int) bool {
	from := d.Index(id)
	if from < 0 || to < 0 || to >= len(d.Nodes) || to == from {
		return false
	}
	n := d.Nodes[from]
	d.Nodes = append(d.Nodes[:from], d.Nodes[from+1:]...)
	d.Nodes = append(d.Nodes[:to], append([]Node{n}, d.Nodes[to:]...)...)
	return true
}

// SampleNumber reads a sample value: a number, or true and false as 1 and 0.
func SampleNumber(value string) (float64, bool) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "true":
		return 1, true
	case "false":
		return 0, true
	}
	f, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
	return f, err == nil
}

func cloneSamples(m map[string]string) map[string]string {
	if m == nil {
		return nil
	}
	out := make(map[string]string, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}
