package scheme

import "github.com/meta-ui-labs/cuppa/libs/color/space"

// ThemeColours are the five colours a design's theme takes from a scheme.
type ThemeColours struct {
	Background, Text, Muted, Border, Secondary space.RGB
}

// Theme picks a design theme from the scheme: the page and text colours as
// they are, the muted colour from bright black (or half way between text and
// page when the scheme's bright black is no different from its page), the
// border from blue and the secondary colour from purple.
func (s Scheme) Theme() ThemeColours {
	t := ThemeColours{
		Background: s.Background,
		Text:       s.Foreground,
		Muted:      s.ANSI[8],
		Border:     s.ANSI[4],
		Secondary:  s.ANSI[5],
	}
	if t.Muted == t.Background || t.Muted == t.Text {
		t.Muted = mix(t.Text, t.Background)
	}
	return t
}

func mix(a, b space.RGB) space.RGB {
	return space.RGB{R: uint8((int(a.R) + int(b.R)) / 2), G: uint8((int(a.G) + int(b.G)) / 2), B: uint8((int(a.B) + int(b.B)) / 2)}
}
