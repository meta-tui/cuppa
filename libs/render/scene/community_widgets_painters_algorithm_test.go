package scene

import (
	"strings"
	"testing"

	"github.com/meta-ui-labs/cuppa/libs/catalog/standard"
	"github.com/meta-ui-labs/cuppa/libs/document/design"
)

func widgetText(id string, w, h int, props map[string]string) string {
	n := design.Node{Component: id, Rect: design.Rect{W: w, H: h}, Props: props}
	return text(RenderNode(n, standard.Default()).Lines())
}

func TestCommunityWidgetsDrawWhatTheirLibrariesShow(t *testing.T) {
	cases := []struct {
		name  string
		id    string
		w, h  int
		props map[string]string
		want  []string
	}{
		{"dropdown closed", "community.dropdown", 26, 7, map[string]string{"open": "false"}, []string{"[ Sencha ▼ ]"}},
		{"dropdown placeholder", "community.dropdown", 26, 7, map[string]string{"open": "false", "selected": "-1"}, []string{"[ Pick a tea ▼ ]"}},
		{"dropdown open", "community.dropdown", 26, 7, nil, []string{"╭───────────╮", "│ Earl Grey │", "│ Matcha    │"}},
		{"prompt input", "community.promptinput", 36, 2, map[string]string{"value": "Sencha"}, []string{"Tea: Sencha█"}},
		{"prompt placeholder", "community.promptinput", 36, 2, nil, []string{"Tea: █Earl Grey"}},
		{"prompt hidden", "community.promptinput", 36, 2, map[string]string{"value": "abc", "hidden": "true"}, []string{"Tea: •••█"}},
		{"prompt error", "community.promptinput", 36, 2, map[string]string{"error": "too short"}, []string{"Error: too short"}},
		{"select", "community.promptselect", 30, 6, nil, []string{"Pick a tea:", "▸ Sencha", "  Earl Grey"}},
		{"select filter", "community.promptselect", 30, 6, map[string]string{"filter": "ma", "selected": "0"}, []string{"Pick a tea: ma█", "▸ Matcha"}},
		{"select scrolls", "community.promptselect", 30, 4, map[string]string{"selected": "6"}, []string{"⇡ Chai", "▸ Pu-erh"}},
		{"data tree", "community.datatree", 34, 9, nil, []string{"name: cuppa", "packs", "├─ lipgloss: 9", "└─ bubbles: 13", "└─ russo: admin"}},
		{"pdf text", "community.pdfview", 44, 14, nil, []string{"Tea menu", "Earl Grey      4.50", "tea-menu.pdf · page 3/12 · text mode"}},
		{"pdf image", "community.pdfview", 44, 14, map[string]string{"mode": "image"}, []string{"▬▬▬ ▬▬▬▬", "page 3/12 · image mode"}},
		{"chart3d surface", "ntcharts.chart3d", 44, 14, nil, []string{"Steep temperature", "● surface", "▓", "└────"}},
		{"chart3d scatter", "ntcharts.chart3d", 44, 14, map[string]string{"kind": "scatter"}, []string{"●", "z"}},
		{"chart3d bar", "ntcharts.chart3d", 44, 14, map[string]string{"kind": "bar"}, []string{"██▒"}},
		{"chart3d vector", "ntcharts.chart3d", 44, 14, map[string]string{"kind": "vector"}, []string{"→"}},
	}
	for _, c := range cases {
		got := widgetText(c.id, c.w, c.h, c.props)
		for _, want := range c.want {
			if !strings.Contains(got, want) {
				t.Errorf("%s: %q not found in\n%s", c.name, want, got)
			}
		}
	}
}

func TestCommunityWidgetsSurviveAnySizeAndHostileInput(t *testing.T) {
	cat := standard.Default()
	props := []map[string]string{
		nil,
		{"options": "", "selected": "99", "open": "true"},
		{"options": "a,b", "selected": "-5", "placeholder": ""},
		{"value": "ünï ☕", "hidden": "true", "error": strings.Repeat("e", 300)},
		{"choices": "", "selected": "7", "filter": "zzz"},
		{"choices": "a,b,c,d,e,f", "selected": "-3", "filter": "A"},
		{"data": "||    deep: x|k: v: w|  child|      far", "text": "|||", "page": "-4", "pages": "0", "mode": "nonsense"},
		{"kind": "nonsense", "title": "", "legend": "true"},
	}
	ids := []string{"community.dropdown", "community.promptinput", "community.promptselect", "community.datatree", "community.pdfview", "ntcharts.chart3d"}
	for _, id := range ids {
		for _, pr := range props {
			for _, size := range []design.Rect{{W: 0, H: 0}, {W: 1, H: 1}, {W: 3, H: 2}, {W: 12, H: 4}, {W: 90, H: 40}} {
				n := design.Node{Component: id, Rect: size, Props: pr}
				_ = RenderNode(n, cat) // must not panic
			}
		}
	}
}
