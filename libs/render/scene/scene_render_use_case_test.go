package scene

import (
	"strings"
	"testing"

	"github.com/meta-ui-labs/cuppa/libs/catalog/standard"
	"github.com/meta-ui-labs/cuppa/libs/document/design"
	"github.com/meta-ui-labs/cuppa/libs/render/grid"
)

func text(lines []string) string {
	var b strings.Builder
	for _, l := range lines {
		esc := false
		for _, r := range l {
			switch {
			case r == 0x1b:
				esc = true
			case esc && r == 'm':
				esc = false
			case !esc:
				b.WriteRune(r)
			}
		}
		b.WriteByte('\n')
	}
	return b.String()
}

func TestEveryShippedComponentRendersAtDefaultAndMinimumSize(t *testing.T) {
	cat := standard.Default()
	for _, def := range cat.List() {
		for _, sz := range []struct{ w, h int }{{def.DefaultSize.W, def.DefaultSize.H}, {def.MinSize.W, def.MinSize.H}} {
			n := design.Node{Component: def.ID, Name: def.Name, Rect: design.Rect{W: sz.w, H: sz.h}}
			g := RenderNode(n, cat) // must not panic at any legal size
			if g.W != sz.w || g.H != sz.h {
				t.Errorf("%s: grid %dx%d, want %dx%d", def.ID, g.W, g.H, sz.w, sz.h)
			}
		}
		if def.Status == "supported" && !Painted(def.ID) {
			t.Errorf("%s is marked supported but has no painter", def.ID)
		}
	}
}

func TestBoxWithTitle(t *testing.T) {
	cat := standard.Default()
	n := design.Node{Component: "lipgloss.box", Rect: design.Rect{W: 12, H: 3}, Props: map[string]string{"title": "Tea", "border": "normal"}}
	got := text(RenderNode(n, cat).Lines())
	want := "┌─ Tea ────┐\n│          │\n└──────────┘\n"
	if got != want {
		t.Fatalf("got\n%s\nwant\n%s", got, want)
	}
}

func TestRenderOrdersBackToFront(t *testing.T) {
	cat := standard.Default()
	doc := design.NewDocument("t", 10, 1)
	doc.Add(design.Node{Component: "lipgloss.label", Rect: design.Rect{W: 10, H: 1}, Props: map[string]string{"text": "back"}})
	doc.Add(design.Node{Component: "lipgloss.label", Rect: design.Rect{X: 2, W: 4, H: 1}, Props: map[string]string{"text": "TOP"}})
	got := text(Render(doc, cat).Lines())
	if got != "baTOP     \n" && got != "baTOP\n" && !strings.HasPrefix(got, "baTOP") {
		t.Fatalf("got %q", got)
	}
}

func TestUnknownComponentFallsBackToPlaceholder(t *testing.T) {
	cat := standard.Default()
	n := design.Node{Component: "mystery", Name: "Mystery", Rect: design.Rect{W: 12, H: 3}}
	if got := text(RenderNode(n, cat).Lines()); !strings.Contains(got, "Mystery") {
		t.Fatalf("placeholder missing name: %q", got)
	}
}

func TestNewComponentsDrawRecognisableContent(t *testing.T) {
	cat := standard.Default()
	cases := []struct {
		id    string
		w, h  int
		props map[string]string
		want  string
	}{
		{"bubbles.paginator", 10, 1, map[string]string{"total": "3", "page": "2"}, "○•○"},
		{"bubbles.timer", 20, 1, map[string]string{"value": "01:23", "label": "Steep"}, "Steep 01:23"},
		{"huh.confirm", 30, 3, map[string]string{"title": "Sure?"}, "Sure?"},
		{"ntcharts.sparkline", 8, 1, map[string]string{"values": "1,2,3,4,5,6,7,8"}, "▁"},
		{"glamour.markdown", 30, 3, map[string]string{"markdown": "# Hi|**bold** word"}, "bold word"},
		{"community.statusbar", 40, 1, map[string]string{"left": "main.go"}, "main.go"},
		{"bubbles.help", 40, 1, map[string]string{"bindings": "q:quit"}, "q quit"},
		{"huh.spinner", 30, 1, map[string]string{"title": "Steeping", "style": "line"}, "Steeping"},
		{"bubbles.tree", 40, 7, map[string]string{"root": "cuppa", "items": "apps/|  tui/|go.work", "selected": "0"}, "→ ▼ cuppa"},
		{"bubbles.tree", 40, 7, map[string]string{"root": "cuppa", "items": "apps/|  tui/|go.work", "selected": "0"}, "├──▼ apps/"},
		{"bubbles.tree", 40, 7, map[string]string{"root": "cuppa", "items": "apps/|  tui/|go.work", "selected": "0"}, "│  └──tui/"},
		{"bubbles.tree", 40, 7, map[string]string{"root": "cuppa", "items": "apps/|  tui/|go.work", "selected": "2"}, "→ │  └──tui/"},
		{"bubbles.tree", 40, 7, map[string]string{"root": "cuppa", "items": "apps/|go.work", "show_help": "true"}, "↓/j down • ↑/k up"},
		{"community.frame", 20, 3, map[string]string{"title": "Menu", "position": "center", "border": "normal"}, "┌────── Menu ──────┐"},
		{"community.dialog", 30, 7, map[string]string{"text": "Pour?", "buttons": "Yes,No"}, "Pour?"},
		{"community.dialog", 30, 7, map[string]string{"text": "Pour?", "buttons": "Yes,No"}, "  Yes    No  "},
		{"community.statusmessage", 30, 1, map[string]string{"text": "Steeping", "spinner": "true"}, "Steeping ⣾"},
		{"community.toast", 24, 3, map[string]string{"kind": "error", "message": "Spilled", "symbols": "ascii"}, "[!!] Spilled"},
		{"community.toast", 24, 3, map[string]string{"kind": "info", "message": "Ready"}, "ⓘ Ready"},
		{"community.bigtext", 40, 6, map[string]string{"text": "Hi", "font": "standard"}, "_   _"},
		{"community.bigtext", 40, 6, map[string]string{"text": "Hi", "font": "no-such-font"}, "_   _"},
		{"community.qrcode", 29, 15, map[string]string{"content": "cuppa"}, "█"},
		{"community.image", 20, 5, map[string]string{"path": "tea.png", "fit": "fill"}, "tea.png · fill"},
	}
	for _, c := range cases {
		n := design.Node{Component: c.id, Rect: design.Rect{W: c.w, H: c.h}, Props: c.props}
		if got := text(RenderNode(n, cat).Lines()); !strings.Contains(got, c.want) {
			t.Errorf("%s: %q not found in\n%s", c.id, c.want, got)
		}
	}
}

func TestTabsOpenTheActiveTabIntoTheWindow(t *testing.T) {
	cat := standard.Default()
	n := design.Node{Component: "lipgloss.tabs", Rect: design.Rect{W: 36, H: 6}, Props: map[string]string{"tabs": "One,Two,Three", "active": "2"}}
	var got []string
	for _, l := range strings.Split(strings.TrimSuffix(text(RenderNode(n, cat).Lines()), "\n"), "\n") {
		got = append(got, strings.TrimRight(l, " "))
	}
	want := []string{
		"╭─────╮╭─────╮╭───────╮",
		"│ One ││ Two ││ Three │",
		"├─────┴┘     └────────┴────────────╮",
		"│                                  │",
		"│                                  │",
		"╰──────────────────────────────────╯",
	}
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Fatalf("got:\n%s\nwant:\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
}

func TestTabsThatDoNotFitAreCutAndTooSmallFallsBack(t *testing.T) {
	cat := standard.Default()
	narrow := design.Node{Component: "lipgloss.tabs", Rect: design.Rect{W: 14, H: 5}, Props: map[string]string{"tabs": "Alpha,Beta,Gamma,Delta"}}
	for _, l := range strings.Split(text(RenderNode(narrow, cat).Lines()), "\n") {
		if len([]rune(l)) > 14 {
			t.Fatalf("a tab spilled past the component: %q", l)
		}
	}
	tiny := design.Node{Component: "lipgloss.tabs", Rect: design.Rect{W: 6, H: 3}}
	if g := RenderNode(tiny, cat); g.W != 6 || g.H != 3 {
		t.Fatalf("grid %dx%d", g.W, g.H)
	}
}

func TestHiddenNodesAreNotPaintedAndLockedOnesAre(t *testing.T) {
	cat := standard.Default()
	doc := design.NewDocument("t", 20, 5)
	doc.Add(design.Node{Component: "lipgloss.box", Name: "Box", Rect: design.Rect{X: 0, Y: 0, W: 6, H: 3}, Hidden: true})
	doc.Add(design.Node{Component: "lipgloss.box", Name: "Box", Rect: design.Rect{X: 10, Y: 0, W: 6, H: 3}, Locked: true})
	out := text(Render(doc, cat).Lines())
	lines := strings.Split(out, "\n")
	if strings.TrimSpace(lines[0][:8]) != "" {
		t.Fatalf("the hidden box was painted: %q", lines[0])
	}
	if !strings.ContainsAny(lines[0][10:], "╭┌") {
		t.Fatalf("the locked box must still be painted: %q", lines[0])
	}
}

func TestRenderWithDrawsAnOverrideInPlaceOfThePainter(t *testing.T) {
	doc := design.NewDocument("t", 20, 6)
	doc.Add(design.Node{Component: "lipgloss.box", Name: "Box", Rect: design.Rect{X: 1, Y: 1, W: 5, H: 3}})
	doc.Add(design.Node{Component: "lipgloss.box", Name: "Other", Rect: design.Rect{X: 10, Y: 1, W: 5, H: 3}})
	custom := grid.New(5, 3)
	custom.Fill(design.Rect{W: 5, H: 3}, 'Z', grid.Style{})
	out := RenderWith(doc, standard.Default(), func(n design.Node) *grid.Grid {
		if n.Name == "Box" {
			return custom
		}
		return nil
	})
	if out.At(1, 1).Ch != 'Z' || out.At(5, 3).Ch != 'Z' {
		t.Fatal("the overridden node is drawn from the override")
	}
	if out.At(10, 1).Ch == 'Z' || out.At(10, 1).Ch == 0 {
		t.Fatal("other nodes still use their painters")
	}
	if Render(doc, standard.Default()).At(1, 1).Ch == 'Z' {
		t.Fatal("Render itself is unchanged")
	}
}

func TestMediaComponentsSurviveHostileInput(t *testing.T) {
	cat := standard.Default()
	cases := []struct {
		id    string
		props map[string]string
	}{
		{"community.bigtext", map[string]string{"text": "tea ☕ ünïcode\nline", "font": "doom"}},
		{"community.bigtext", map[string]string{"text": "", "font": "big"}},
		{"community.qrcode", map[string]string{"content": strings.Repeat("x", 9000)}},
		{"community.qrcode", map[string]string{"content": ""}},
		{"community.qrcode", map[string]string{"content": "cuppa", "level": "nonsense"}},
		{"community.image", map[string]string{"color": "not-a-colour"}},
	}
	for _, c := range cases {
		for _, size := range []design.Rect{{W: 1, H: 1}, {W: 12, H: 4}, {W: 80, H: 30}} {
			n := design.Node{Component: c.id, Rect: size, Props: c.props}
			_ = RenderNode(n, cat) // must not panic or exit
		}
	}
}
