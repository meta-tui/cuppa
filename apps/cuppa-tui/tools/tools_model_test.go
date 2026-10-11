package tools

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/meta-ui-labs/cuppa/apps/cuppa-tui/pointer"
	"github.com/meta-ui-labs/cuppa/libs/canvas/shape"
)

func TestShortcutLettersChooseTheTools(t *testing.T) {
	m := New()
	for r, want := range map[rune]Tool{'U': Rectangle, 'p': Path, 'B': Brush, 'e': Erase, 'T': Text, 'v': Select} {
		if !m.ChooseKey(r) || m.Tool() != want {
			t.Fatalf("%q: tool = %v, want %v", r, m.Tool(), want)
		}
	}
	if m.ChooseKey('x') {
		t.Fatal("x is not a tool")
	}
}

func TestEveryToolShowsItsOwnOptions(t *testing.T) {
	m := New()
	want := map[Tool][]string{
		Select:    {"Click a component"},
		Rectangle: {"Fill", "Border", "Corners", "╭", "┌", "╱"},
		Path:      {"Line", "End", "Thickness", "Colour"},
		Brush:     {"Char", "Thickness", "Colour"},
		Erase:     {"only works on the drawing"},
		Text:      {"Colour", "Click the canvas"},
	}
	for tool, parts := range want {
		m.Choose(tool)
		line := ansi.Strip(m.BarLine(400))
		for _, p := range parts {
			if !strings.Contains(line, p) {
				t.Errorf("%v bar %q lacks %q", tool, line, p)
			}
		}
	}
}

func TestTheBarChangesTheOptions(t *testing.T) {
	m := New()
	m.Choose(Rectangle)
	line := ansi.Strip(m.BarLine(400))
	at := ansi.StringWidth(line[:strings.Index(line, "╭")])
	m.HandleBar(pointer.Event{X: at, Y: 0, Phase: pointer.Down, Left: true})
	if m.Opts.Corners != shape.CornersRound {
		t.Fatalf("corners = %q, want round", m.Opts.Corners)
	}
	m.Choose(Path)
	line = ansi.Strip(m.BarLine(400))
	plus := ansi.StringWidth(line[:strings.Index(line, "[+]")])
	m.HandleBar(pointer.Event{X: plus + 1, Y: 0, Phase: pointer.Down, Left: true})
	if m.Opts.Thickness != 2 {
		t.Fatalf("thickness = %d, want 2", m.Opts.Thickness)
	}
}

func TestBrushAndPathKeepTheirOwnCharacters(t *testing.T) {
	m := New()
	m.Opts.BrushChar = "▓"
	if m.LineOptions().Char != shape.AutoChar {
		t.Fatal("the brush character leaked into the path")
	}
	if m.BrushOptions().Char != "▓" {
		t.Fatal("the brush lost its character")
	}
}
