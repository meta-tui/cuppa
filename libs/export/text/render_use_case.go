// Package text exports a design as terminal text: with colours (ANSI) or plain.
package text

import (
	"strings"

	"github.com/meta-ui-labs/cuppa/libs/document/design"
	"github.com/meta-ui-labs/cuppa/libs/render/scene"
)

// ANSI renders the design with colour escape codes, one line per canvas row.
func ANSI(doc design.Document, cat scene.Catalog) string {
	return scene.Render(doc, cat).String() + "\n"
}

// Plain renders the design as unstyled text, trimming trailing spaces.
func Plain(doc design.Document, cat scene.Catalog) string {
	g := scene.Render(doc, cat)
	lines := make([]string, g.H)
	for y := range lines {
		var b strings.Builder
		for x := 0; x < g.W; x++ {
			ch := g.At(x, y).Ch
			if ch == 0 {
				ch = ' '
			}
			b.WriteRune(ch)
		}
		lines[y] = strings.TrimRight(b.String(), " ")
	}
	return strings.Join(lines, "\n") + "\n"
}
