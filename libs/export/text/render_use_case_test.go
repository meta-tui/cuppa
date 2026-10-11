package text

import (
	"strings"
	"testing"

	"github.com/meta-ui-labs/cuppa/libs/catalog/standard"
	"github.com/meta-ui-labs/cuppa/libs/document/design"
)

func demo() design.Document {
	d := design.NewDocument("t", 20, 5)
	d.Add(design.Node{Component: "lipgloss.box", Name: "Box", Rect: design.Rect{X: 1, Y: 1, W: 8, H: 3}})
	return d
}

func TestPlainHasNoEscapes(t *testing.T) {
	out := Plain(demo(), standard.Default())
	if strings.Contains(out, "\x1b") {
		t.Fatalf("escape codes in plain output: %q", out)
	}
	lines := strings.Split(strings.TrimSuffix(out, "\n"), "\n")
	if len(lines) != 5 {
		t.Fatalf("want 5 lines, got %d", len(lines))
	}
	for _, l := range lines {
		if l != strings.TrimRight(l, " ") {
			t.Fatalf("trailing spaces in %q", l)
		}
	}
	if !strings.ContainsAny(out, "╭┌│") {
		t.Fatalf("box not drawn:\n%s", out)
	}
}

func TestANSIKeepsColour(t *testing.T) {
	d := design.NewDocument("t", 20, 3)
	d.Add(design.Node{Component: "huh.confirm", Name: "Confirm", Rect: design.Rect{W: 20, H: 3}})
	if out := ANSI(d, standard.Default()); !strings.Contains(out, "\x1b[") {
		t.Fatalf("no colour in ANSI output: %q", out)
	}
}
