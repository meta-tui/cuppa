package scene

import (
	"strings"

	figure "github.com/common-nighthawk/go-figure"
	qrcode "github.com/skip2/go-qrcode"

	"github.com/meta-ui-labs/cuppa/libs/color/space"
	"github.com/meta-ui-labs/cuppa/libs/render/grid"
)

// figletFonts are the fonts the catalog offers. go-figure stops the whole
// program on a font it does not know, so only these names reach it.
var figletFonts = map[string]bool{
	"standard": true, "big": true, "small": true, "mini": true, "slant": true,
	"shadow": true, "doom": true, "banner": true, "block": true, "lean": true,
}

// paintBigText draws the text as FIGlet rows, cut to the node.
func paintBigText(g *grid.Grid, p Props) {
	font := p.Str("font")
	if !figletFonts[font] {
		font = "standard"
	}
	text := strings.TrimSpace(p.Str("text"))
	if text == "" {
		return
	}
	style := grid.Style{Fg: p.Str("color")}
	for y, row := range figure.NewFigure(text, font, false).Slicify() {
		if y >= g.H {
			break
		}
		g.Text(0, y, row, style, g.W)
	}
}

var qrLevels = map[string]qrcode.RecoveryLevel{
	"low": qrcode.Low, "medium": qrcode.Medium, "high": qrcode.High, "highest": qrcode.Highest,
}

// paintQRCode draws a QR code two modules to a row with half blocks, dark on a
// light ground so that it scans on a dark terminal too. A node smaller than the
// code shows its top left corner.
func paintQRCode(g *grid.Grid, p Props) {
	content := p.Str("content")
	level, ok := qrLevels[p.Str("level")]
	if !ok {
		level = qrcode.Medium
	}
	code, err := qrcode.New(content, level)
	if content == "" || err != nil {
		paintGeneric(g, "QR code")
		return
	}
	bits := code.Bitmap()
	dark := func(x, y int) bool { return y < len(bits) && x < len(bits[y]) && bits[y][x] }
	style := grid.Style{Fg: "0", Bg: "255"}
	for y := 0; y < g.H; y++ {
		for x := 0; x < g.W; x++ {
			top, bottom := dark(x, 2*y), dark(x, 2*y+1)
			ch := ' '
			switch {
			case top && bottom:
				ch = '█'
			case top:
				ch = '▀'
			case bottom:
				ch = '▄'
			}
			g.Set(x, y, grid.Cell{Ch: ch, Style: style})
		}
	}
}

// paintImage draws a stand-in for a picture: a colour wash from a deep blue to
// the accent colour in half blocks, with the file name on a bar at the bottom.
func paintImage(g *grid.Grid, p Props) {
	from, _ := space.Resolve("#1d2b53")
	to, ok := space.Resolve(p.Str("color"))
	if !ok {
		to, _ = space.Resolve("#ff5fd7")
	}
	for y := 0; y < g.H; y++ {
		for x := 0; x < g.W; x++ {
			top := mix(from, to, float64(x+2*y)/float64(g.W+2*g.H))
			bottom := mix(from, to, float64(x+2*y+1)/float64(g.W+2*g.H))
			g.Set(x, y, grid.Cell{Ch: '▀', Style: grid.Style{Fg: top.Hex(), Bg: bottom.Hex()}})
		}
	}
	if g.H >= 3 && g.W >= 6 {
		bar := grid.Style{Fg: "255", Bg: "0", Bold: true}
		for x := 0; x < g.W; x++ {
			g.Set(x, g.H-1, grid.Cell{Ch: ' ', Style: bar})
		}
		g.Text(1, g.H-1, p.Str("path")+" · "+p.Str("fit"), bar, g.W-2)
	}
}
