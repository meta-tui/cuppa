package gosource

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/meta-ui-labs/cuppa/libs/catalog/standard"
	"github.com/meta-ui-labs/cuppa/libs/document/design"
	"github.com/meta-ui-labs/cuppa/libs/render/scene"
)

// TestWidgetsDrawTheSameInTheDesignerAndTheProgram builds a program with every
// variant of the community widgets and checks that each draws, character for
// character, what the designer draws. The two drawings are written twice (the
// painter and community_widgets_go.txt), so this is what keeps them one.
func TestWidgetsDrawTheSameInTheDesignerAndTheProgram(t *testing.T) {
	if testing.Short() {
		t.Skip("builds a project with the go tool")
	}
	goBin, err := exec.LookPath("go")
	if err != nil {
		t.Skip("no go tool")
	}
	cat := standard.Default()
	variants := []struct {
		name, id string
		w, h     int
		props    map[string]string
	}{
		{"dropdown open", "community.dropdown", 26, 7, nil},
		{"dropdown closed", "community.dropdown", 26, 7, map[string]string{"open": "false", "selected": "-1"}},
		{"dropdown scrolled", "community.dropdown", 20, 5, map[string]string{"selected": "3"}},
		{"input", "community.promptinput", 36, 2, map[string]string{"value": "Sencha"}},
		{"input placeholder", "community.promptinput", 36, 2, nil},
		{"input hidden", "community.promptinput", 36, 2, map[string]string{"value": "secret", "hidden": "true", "error": "too short"}},
		{"select", "community.promptselect", 30, 6, nil},
		{"select filtered", "community.promptselect", 30, 6, map[string]string{"filter": "e", "selected": "2"}},
		{"select scrolled", "community.promptselect", 30, 4, map[string]string{"selected": "6"}},
		{"tree", "community.datatree", 34, 9, nil},
		{"pdf text", "community.pdfview", 44, 14, nil},
		{"pdf image", "community.pdfview", 44, 14, map[string]string{"mode": "image"}},
		{"3d surface", "ntcharts.chart3d", 44, 14, nil},
		{"3d scatter", "ntcharts.chart3d", 44, 14, map[string]string{"kind": "scatter"}},
		{"3d bar", "ntcharts.chart3d", 44, 14, map[string]string{"kind": "bar"}},
		{"3d line", "ntcharts.chart3d", 44, 14, map[string]string{"kind": "line"}},
		{"3d vector", "ntcharts.chart3d", 44, 14, map[string]string{"kind": "vector", "legend": "false", "title": ""}},
		{"picker rgb", "lipgloss.colourpicker", 46, 10, nil},
		{"picker 16", "lipgloss.colourpicker", 46, 10, map[string]string{"tab": "16", "value": "9"}},
		{"picker 256", "lipgloss.colourpicker", 46, 23, map[string]string{"tab": "256", "value": "#00ff00"}},
		{"picker hsl", "lipgloss.colourpicker", 46, 10, map[string]string{"tab": "HSL", "value": "", "slide": "2"}},
		{"rows", "lipgloss.rows", 44, 5, nil},
		{"rows tokens", "lipgloss.rows", 40, 3, map[string]string{"columns": "Mark:2:fg=@Accent,Label:12:selbg=#00005f,Swatch:2:glyph,Value:ellipsis:fg=#00f0ff", "rows": "▸,Parallel,#ff007f,a very long value that is cut;,Folder,#3ddc84,/tmp", "styles": "selected,normal", "cellstyles": "-,-,-,#112233|#445566|bd;-,-,-,-"}},
		{"rows sgr", "lipgloss.rows", 30, 2, map[string]string{"columns": "A:10,B", "rows": "\x1b[31mred\x1b[0m text,\x1b[1;38;5;220mbold\x1b[0m and more", "styles": "normal"}},
		{"box title and right", "lipgloss.box", 40, 3, map[string]string{"title": "Playlists", "right": "3", "color": "#6b7280", "titleColor": "#ff007f"}},
		{"box title cut", "lipgloss.box", 40, 3, map[string]string{"title": "90s UK Dance Hits > 05 Stardust - Music Sounds Better With You", "right": "follow"}},
		{"box title only", "lipgloss.box", 20, 3, map[string]string{"title": "A very long title indeed"}},
		{"box no room for right", "lipgloss.box", 14, 3, map[string]string{"title": "T", "right": "a long right label"}},
		{"progress line", "lipgloss.progressline", 70, 1, nil},
		{"progress line fit", "lipgloss.progressline", 40, 1, map[string]string{"barWidth": "0", "value": "99.5", "suffix": "done", "full": "=", "empty": "-"}},
		{"rows fill auto", "lipgloss.rows", 24, 2, map[string]string{"columns": "Mark:1,Sp:1,Title:fill:ellipsis,Gap:1,Count:auto,Sp:1,Glyph:1", "rows": "▸, ,90s UK Dance Hits and more, ,3/6, ,x;v, ,Short, ,listing, ,y", "styles": "normal,normal"}},
		{"rows fit", "lipgloss.rows", 60, 3, map[string]string{"columns": "Review:6,Artist:fit=4..30%,Song:fit=6..50%,Video:fill:min=4:ellipsis,Replaced:fit=0..25%:floor=6:hide-empty,Addr:8", "rows": "ok,Aphex,Windowlicker (extended mix),Aphex Twin - Windowlicker official video, ,url;,B,Hi,Another long video title here, ,url2", "gap": "2", "styles": "normal,selected"}},
		{"rows fit hidden", "lipgloss.rows", 40, 2, map[string]string{"columns": "A:fit=2..40%,B:fill,C:fit=0..30%:hide-empty", "rows": "x,wide text,;yy,more,", "gap": "1"}},
		{"rows wide", "lipgloss.rows", 40, 1, map[string]string{"columns": "G:1,Sp:1,Idx:2,Sp2:1,Title:fill:ellipsis,Sp3:1,Tag:auto", "rows": "x, ,05, ,Stardust - Music 日本語タイトル and more, , 34%"}},
		{"tabs", "cuppa.tabs", 40, 6, map[string]string{"pages": "Overview,Colours,Keys", "active": "1"}},
		{"tabs narrow", "cuppa.tabs", 14, 4, map[string]string{"pages": "Overview,Colours,Keys", "active": "2", "color": "#ff007f"}},
		{"panels", "cuppa.panels", 30, 12, map[string]string{"pages": "Details,Properties,Layers", "open": "0,2"}},
		{"panels all closed", "cuppa.panels", 20, 5, map[string]string{"pages": "A,B,C", "open": ""}},
		{"keybar", "lipgloss.keybar", 60, 1, nil},
		{"keybar clipped", "lipgloss.keybar", 20, 1, map[string]string{"hints": "enter:apply,esc:cancel,tab:next,s:save", "gap": "3"}},
		{"slot", "lipgloss.slot", 30, 3, map[string]string{"view": "first\x1b[7m cursor\x1b[0m line\nsecond line"}},
		{"rows spaces", "lipgloss.rows", 30, 3, map[string]string{"columns": "A:6,B:6,C", "rows": "  x,,z; y\\,w,,; q,r,s", "styles": "accent,dim,selected"}},
		{"rows colours", "lipgloss.rows", 30, 3, map[string]string{"columns": "Name:8,Swatch:5:colour,Rest", "rows": "Tea,#ff0000,hot;Milk,,cold", "styles": "selected"}},
	}
	doc := design.NewDocument("widgets", 200, 600)
	var nodes []design.Node
	for i, v := range variants {
		n := design.Node{Component: v.id, Name: v.name, Rect: design.Rect{X: 0, Y: i * 15, W: v.w, H: v.h}, Props: v.props}
		doc.Add(n)
		nodes = append(nodes, n)
	}
	dir := filepath.Join(t.TempDir(), "app")
	p := Generate(doc, cat)
	if len(p.Notes) != 0 {
		t.Fatalf("notes: %v", p.Notes)
	}
	if err := Write(dir, p); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "dump_test.go"), []byte(dumpTest), 0o644); err != nil {
		t.Fatal(err)
	}
	env := append(os.Environ(), "GOFLAGS=-mod=mod", "GOWORK=off")
	for _, args := range [][]string{{"mod", "tidy"}, {"vet", "."}} {
		cmd := exec.Command(goBin, args...)
		cmd.Dir, cmd.Env = dir, env
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("go %s: %v\n%s", strings.Join(args, " "), err, out)
		}
	}
	run := exec.Command(goBin, "test", "-run", "TestDumpViews", "-v", ".")
	run.Dir, run.Env = dir, env
	out, err := run.CombinedOutput()
	if err != nil {
		t.Fatalf("go test: %v\n%s", err, out)
	}
	var views map[string][]string
	for _, line := range strings.Split(string(out), "\n") {
		if rest, ok := strings.CutPrefix(line, "VIEWS="); ok {
			if err := json.Unmarshal([]byte(rest), &views); err != nil {
				t.Fatal(err)
			}
		}
	}
	for _, n := range nodes {
		want := trimRows(views[n.Name], n.Rect.H)
		got := trimRows(strings.Split(ansi.Strip(strings.Join(scene.RenderNode(n, cat).Lines(), "\n")), "\n"), n.Rect.H)
		if strings.Join(got, "\n") != strings.Join(want, "\n") {
			t.Errorf("%s differs\n-- designer\n%s\n-- program\n%s", n.Name, strings.Join(got, "\n"), strings.Join(want, "\n"))
		}
	}
}

func trimRows(lines []string, h int) []string {
	out := make([]string, h)
	for i := range out {
		if i < len(lines) {
			out[i] = strings.TrimRight(lines[i], " ")
		}
	}
	return out
}
