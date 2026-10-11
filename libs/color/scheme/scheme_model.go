// Package scheme lists terminal colour schemes (Dracula, Nord, Solarized and
// hundreds more) so a design can pick colours from them. A scheme is only a set
// of colours; picking one stores a plain hex colour, so no file format changes.
package scheme

import "github.com/meta-ui-labs/cuppa/libs/color/space"

// Scheme is one terminal colour scheme.
type Scheme struct {
	// Name is the display name, for example "Tokyo Night".
	Name string
	// Dark is true when the background is dark.
	Dark bool
	// ANSI holds the 16 terminal colours in the usual order: black, red, green,
	// yellow, blue, purple, cyan, white, then the same eight bright.
	ANSI [16]space.RGB
	// Foreground and Background are the scheme's default text and page colours.
	Foreground, Background space.RGB
	// Cursor and Selection are optional: most schemes leave them out.
	Cursor, Selection       space.RGB
	HasCursor, HasSelection bool
}

// ANSINames are the names of the 16 ANSI colours, in the order of Scheme.ANSI.
var ANSINames = [16]string{
	"Black", "Red", "Green", "Yellow", "Blue", "Purple", "Cyan", "White",
	"Bright black", "Bright red", "Bright green", "Bright yellow",
	"Bright blue", "Bright purple", "Bright cyan", "Bright white",
}
