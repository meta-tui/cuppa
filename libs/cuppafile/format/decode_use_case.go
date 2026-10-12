package format

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"encoding/json"
	"io"
	"strings"

	"github.com/meta-ui-labs/cuppa/libs/color/space"
	"github.com/meta-ui-labs/cuppa/libs/document/design"
)

// Decode reads a .cuppa file, migrating older versions to the current one.
func Decode(data []byte) (design.Document, error) {
	if !bytes.HasPrefix(data, []byte(Magic)) {
		return design.Document{}, ErrNotCuppa
	}
	rest := data[len(Magic):]
	if len(rest) < 2 {
		return design.Document{}, ErrCorrupt
	}
	version := binary.BigEndian.Uint16(rest)
	if version > CurrentVersion {
		return design.Document{}, ErrTooNew
	}
	if version == 0 {
		return design.Document{}, ErrCorrupt
	}
	zr, err := gzip.NewReader(bytes.NewReader(rest[2:]))
	if err != nil {
		return design.Document{}, ErrCorrupt
	}
	body, err := io.ReadAll(io.LimitReader(zr, MaxDecodedSize+1))
	if err != nil || len(body) > MaxDecodedSize {
		return design.Document{}, ErrCorrupt
	}
	body, err = migrate(body, version)
	if err != nil {
		return design.Document{}, err
	}
	var env envelope
	if err := json.Unmarshal(body, &env); err != nil {
		return design.Document{}, ErrCorrupt
	}
	return normalise(env.Document), nil
}

// normalise repairs what a hand-edited or damaged file can get wrong, so the
// editor never sees an impossible document.
func normalise(doc design.Document) design.Document {
	doc.Width = min(max(doc.Width, 1), design.MaxWidth)
	doc.Height = min(max(doc.Height, 1), design.MaxHeight)
	if bg, err := space.Normalise(doc.Background); err == nil {
		doc.Background = bg
	} else {
		doc.Background = ""
	}
	for _, c := range []*string{&doc.Theme.Text, &doc.Theme.Muted, &doc.Theme.Border, &doc.Theme.Secondary} {
		if v, err := space.Normalise(*c); err == nil {
			*c = v
		} else {
			*c = ""
		}
	}
	if !design.ValidProfile(doc.Profile) {
		doc.Profile = ""
	}
	doc.Theme.Palette = cleanPalette(doc.Theme.Palette)
	nodes := cleanNodes(doc.Nodes, 0)
	doc.Embedded = cleanEmbedded(doc.Embedded)
	doc.Nodes = nodes
	doc = cleanPages(doc)
	return doc
}

// maxPages bounds how many pages a file may carry.
const maxPages = 64

// cleanPages keeps the pages that have a usable, unused name, cleans the
// components of those that are not shown, and keeps the shown page inside the
// pages. A single page is no split at all.
func cleanPages(doc design.Document) design.Document {
	if len(doc.Pages) == 0 {
		doc.Page = 0
		return doc
	}
	seen := map[string]bool{}
	var pages []design.Page
	hasMaster := false
	active := -1
	for i, p := range doc.Pages {
		p.Name = strings.TrimSpace(p.Name)
		if design.ValidPageName(p.Name) != nil || seen[p.Name] || len(pages) >= maxPages {
			if i == doc.Page {
				active = len(pages)
			}
			continue
		}
		seen[p.Name] = true
		if i == doc.Page {
			active = len(pages)
		}
		p.Nodes = cleanNodes(p.Nodes, 0)
		if p.Master {
			if hasMaster {
				p.Master = false
			}
			hasMaster = true
			p.Hide = nil
		} else if len(p.Hide) > maxNodesPerPage {
			p.Hide = p.Hide[:maxNodesPerPage]
		}
		pages = append(pages, p)
	}
	if len(pages) == 0 {
		doc.Pages, doc.Page = nil, 0
		return doc
	}
	doc.Pages = pages
	doc.Page = min(max(active, 0), len(pages)-1)
	// The shown page keeps its components in Nodes.
	doc.Pages[doc.Page].Nodes = nil
	return doc
}

// maxEmbedded bounds how many component copies a file may carry.
const maxEmbedded = 256

// cleanEmbedded keeps the embedded components that are usable and not repeated.
func cleanEmbedded(in []design.Embedded) []design.Embedded {
	var out []design.Embedded
	seen := map[string]bool{}
	for _, e := range in {
		if e.ID == "" || seen[e.ID] || e.Composite.Check() != nil || len(out) >= maxEmbedded {
			continue
		}
		seen[e.ID] = true
		out = append(out, e)
	}
	return out
}

// maxGroupDepth bounds how deeply groups may nest in a file.
const maxGroupDepth = 16

// cleanNodes drops nodes with an empty or repeated id or a rect under 1×1, and
// repairs groups: children are cleaned the same way, and a group left with no
// children or no base size is dropped. Children on a plain node are ignored.
func cleanNodes(in []design.Node, depth int) []design.Node {
	seen := map[design.NodeID]bool{}
	out := make([]design.Node, 0, len(in))
	for _, n := range in {
		if n.ID == "" || seen[n.ID] || n.Rect.W < 1 || n.Rect.H < 1 {
			continue
		}
		seen[n.ID] = true
		if n.Component == design.GroupComponent {
			if depth >= maxGroupDepth || n.BaseW < 1 || n.BaseH < 1 {
				continue
			}
			n.Children = cleanNodes(n.Children, depth+1)
			if len(n.Children) == 0 {
				continue
			}
		} else {
			n.Children, n.BaseW, n.BaseH = nil, 0, 0
		}
		out = append(out, n)
	}
	return out
}

// cleanPalette drops swatches with an invalid name or colour and repeated
// names.
func cleanPalette(in []design.Swatch) []design.Swatch {
	var out []design.Swatch
	seen := map[string]bool{}
	for _, sw := range in {
		color, err := space.Normalise(sw.Color)
		if err != nil || color == "" || !design.ValidSwatchName(sw.Name) || seen[sw.Name] {
			continue
		}
		seen[sw.Name] = true
		out = append(out, design.Swatch{Name: sw.Name, Color: color})
	}
	return out
}

// maxNodesPerPage bounds the list of master components a page hides.
const maxNodesPerPage = 2048
