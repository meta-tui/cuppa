package editor

import "testing"

// A layout expression can read the place of a component before it.
func TestLayoutReadsThePlaceOfOtherComponentsInTheDesigner(t *testing.T) {
	e := newEditor()
	top := mustAdd(t, e, "lipgloss.box", 0, 2)
	e.Rename(top, "Playlists")
	must(t, e.SetLayout(top, AxisH, "min(8, (100% - 4) / 3)"))
	if got := rectOf(e, top).H; got != 6 {
		t.Errorf("playlists is %d high, want 6", got)
	}

	under := mustAdd(t, e, "lipgloss.box", 0, 12)
	e.Rename(under, "Entries")
	must(t, e.SetLayout(under, AxisY, `below("Playlists")`))
	if got, want := rectOf(e, under).Y, rectOf(e, top).Y+6; got != want {
		t.Errorf("entries at %d, want %d", got, want)
	}

	// A bad expression is refused with a reason.
	if err := e.SetLayout(top, AxisH, "$"); err == nil {
		t.Error("a lone $ was accepted")
	}
}
