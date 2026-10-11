package gosource

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/meta-ui-labs/cuppa/libs/catalog/standard"
	"github.com/meta-ui-labs/cuppa/libs/document/design"
	"github.com/meta-ui-labs/cuppa/libs/render/scene"
)

// dumpTest is added to the generated project: it prints what every placed
// component draws, as JSON, with the colour codes removed.
const dumpTest = `package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func TestDumpViews(t *testing.T) {
	out := map[string][]string{}
	for _, c := range layout {
		out[c.Name] = strings.Split(ansi.Strip(build(c).view()), "\n")
	}
	data, _ := json.Marshal(out)
	fmt.Println("VIEWS=" + string(data))
}
`

// TestPreviewMatchesTheGeneratedProgram compares, for every component of the
// catalog at its default size, what the designer draws with what the generated
// program draws, and reports how much of the drawn area is the same. It is a
// measuring tool: it never fails on a difference, it prints the report with
// -v. It needs the go tool and the module proxy, so -short skips it.
func TestPreviewMatchesTheGeneratedProgram(t *testing.T) {
	if testing.Short() {
		t.Skip("builds a project with the go tool")
	}
	cat := standard.Default()
	doc := design.NewDocument("fidelity", 200, 200)
	type item struct {
		id   string
		node design.Node
	}
	var items []item
	for _, def := range cat.List() {
		if def.Inner != nil {
			continue
		}
		n := design.Node{Component: def.ID, Name: def.ID, Rect: design.Rect{W: def.DefaultSize.W, H: def.DefaultSize.H}}
		doc.Add(n)
		items = append(items, item{def.ID, n})
	}
	dir := filepath.Join(t.TempDir(), "app")
	if err := Write(dir, Generate(doc, cat)); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "dump_test.go"), []byte(dumpTest), 0o644); err != nil {
		t.Fatal(err)
	}
	goBin, err := exec.LookPath("go")
	if err != nil {
		t.Skip("no go tool")
	}
	env := append(os.Environ(), "GOFLAGS=-mod=mod", "GOWORK=off")
	tidy := exec.Command(goBin, "mod", "tidy")
	tidy.Dir, tidy.Env = dir, env
	if out, err := tidy.CombinedOutput(); err != nil {
		t.Fatalf("go mod tidy: %v\n%s", err, out)
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
	if len(views) == 0 {
		t.Fatalf("no views in the output:\n%s", out)
	}

	type row struct {
		id    string
		score float64
	}
	var rows []row
	for _, it := range items {
		want := views[it.id]
		got := strings.Split(ansi.Strip(strings.Join(scene.RenderNode(it.node, cat).Lines(), "\n")), "\n")
		score := similarity(got, want, it.node.Rect.W, it.node.Rect.H)
		rows = append(rows, row{it.id, score})
		if report := os.Getenv("CUPPA_FIDELITY_SHOW"); report != "" && (strings.Contains(it.id, report) || report == "low" && score < 0.95) {
			fmt.Printf("== %s (%.0f%%)\n-- designer\n%s\n-- generated\n%s\n", it.id, score*100,
				strings.Join(trimmed(got, it.node.Rect.H), "\n"), strings.Join(trimmed(want, it.node.Rect.H), "\n"))
		}
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].score < rows[j].score })
	var sum float64
	for _, r := range rows {
		sum += r.score
		t.Logf("%5.1f%%  %s", r.score*100, r.id)
	}
	mean := sum / float64(len(rows))
	t.Logf("mean %.1f%% over %d components", mean*100, len(rows))
	// The painters were checked against the generated program one by one; the
	// components below the line are sketches of something that depends on the
	// machine (files, folders) or that is drawn differently on purpose. A
	// painter that drifts from its component drags the mean down.
	if mean < 0.88 {
		t.Errorf("the designer's drawings match the generated program %.1f%% (want at least 88%%): run with -v and CUPPA_FIDELITY_SHOW=low to see", mean*100)
	}
}

// similarity is the share of the cells either drawing uses (not blank) that are
// the same character in both.
func similarity(a, b []string, w, h int) float64 {
	cell := func(lines []string, x, y int) rune {
		if y >= len(lines) {
			return ' '
		}
		r := []rune(lines[y])
		if x >= len(r) {
			return ' '
		}
		return r[x]
	}
	used, same := 0, 0
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			ra, rb := cell(a, x, y), cell(b, x, y)
			if ra == ' ' && rb == ' ' {
				continue
			}
			used++
			if ra == rb {
				same++
			}
		}
	}
	if used == 0 {
		return 1
	}
	return float64(same) / float64(used)
}

func trimmed(lines []string, h int) []string {
	out := make([]string, 0, h)
	for i := 0; i < h && i < len(lines); i++ {
		out = append(out, strings.TrimRight(lines[i], " "))
	}
	return out
}
