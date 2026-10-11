package screenscmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/meta-ui-labs/cuppa/libs/canvas/editor"
	"github.com/meta-ui-labs/cuppa/libs/catalog/cupp"
	"github.com/meta-ui-labs/cuppa/libs/catalog/standard"
	"github.com/meta-ui-labs/cuppa/libs/cuppafile/disk"
	"github.com/meta-ui-labs/cuppa/libs/document/design"
	"github.com/meta-ui-labs/cuppa/libs/export/gosource"
	"github.com/meta-ui-labs/cuppa/libs/export/text"
)

// goldenSize is one size a page is drawn at.
type goldenSize struct{ w, h int }

// goldenDrawing is one page drawn at one size.
type goldenDrawing struct {
	page string
	size goldenSize
	view string
}

// goldenFile is the name of a drawing's file: <page>_<w>x<h>.ansi.
func goldenFile(d goldenDrawing) string {
	return fmt.Sprintf("%s_%dx%d.ansi", gosource.Slug(d.page), d.size.w, d.size.h)
}

// parseSizes reads "80x24,120x30".
func parseSizes(list string) ([]goldenSize, error) {
	var out []goldenSize
	for _, part := range strings.Split(list, ",") {
		w, h, ok := strings.Cut(strings.TrimSpace(strings.ToLower(part)), "x")
		wn, werr := strconv.Atoi(w)
		hn, herr := strconv.Atoi(h)
		if !ok || werr != nil || herr != nil || wn < 1 || hn < 1 || wn > 500 || hn > 200 {
			return nil, fmt.Errorf("%q is not a size: use WxH, for example 80x24", part)
		}
		out = append(out, goldenSize{wn, hn})
	}
	return out, nil
}

// runGolden writes the drawing of each page of a design at each size into dir,
// or, with check, compares them with the files there. By default the pages are
// drawn by Cuppa's own renderer with their samples; with generated they are
// drawn by the Go the export writes (built and run with the go tool), which is
// what an application ships.
func runGolden(file, dir, sizes string, generated, check bool, stdout io.Writer) error {
	doc, err := disk.Load(file)
	if err != nil {
		return fmt.Errorf("%s: %w", file, err)
	}
	list := []goldenSize{{doc.Width, doc.Height}}
	if sizes != "" {
		if list, err = parseSizes(sizes); err != nil {
			return err
		}
	}
	cat := cupp.Adopt(standard.Default(), doc.Embedded)
	var drawings []goldenDrawing
	if generated {
		if drawings, err = drawGenerated(doc, cat, list); err != nil {
			return err
		}
	} else {
		for _, page := range doc.Screens() {
			for _, size := range list {
				ed := editor.New(cat, page)
				if err := ed.SetCanvasSize(size.w, size.h); err != nil {
					return fmt.Errorf("%s at %dx%d: %w", page.Name, size.w, size.h, err)
				}
				drawings = append(drawings, goldenDrawing{page.Name, size, text.ANSI(ed.Document(), cat)})
			}
		}
	}
	sort.Slice(drawings, func(i, j int) bool { return goldenFile(drawings[i]) < goldenFile(drawings[j]) })

	if check {
		var bad []string
		for _, d := range drawings {
			data, err := os.ReadFile(filepath.Join(dir, goldenFile(d)))
			switch {
			case err != nil:
				bad = append(bad, goldenFile(d)+": missing")
			case strings.ReplaceAll(string(data), "\r\n", "\n") != strings.ReplaceAll(d.view, "\r\n", "\n"):
				bad = append(bad, goldenFile(d)+": differs from the drawing")
			}
		}
		if len(bad) == 0 {
			_, _ = fmt.Fprintf(stdout, "%s is current: %d drawings\n", dir, len(drawings))
			return nil
		}
		for _, b := range bad {
			_, _ = fmt.Fprintln(stdout, "  "+b)
		}
		return fmt.Errorf("%s is out of date with the design: run cuppa screens --golden again", dir)
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	for _, d := range drawings {
		if err := os.WriteFile(filepath.Join(dir, goldenFile(d)), []byte(d.view), 0o644); err != nil {
			return err
		}
	}
	_, _ = fmt.Fprintf(stdout, "%d drawings written to %s\n", len(drawings), dir)
	return nil
}

// goldenProgram is the program that draws each page of the generated project
// at each size and prints one JSON line per drawing.
const goldenProgram = `package main

import (
	"encoding/json"
	"os"

%s
)

type drawing struct {
	Page string
	W, H int
	View string
}

func main() {
	enc := json.NewEncoder(os.Stdout)
	for _, size := range [][2]int{%s} {
%s
	}
}
`

// drawGenerated exports the design as the slice layout into a temporary
// module, builds a program that calls each page's View with its Default props
// and reads the drawings back.
func drawGenerated(doc design.Document, cat gosource.Catalog, sizes []goldenSize) ([]goldenDrawing, error) {
	goBin, err := exec.LookPath("go")
	if err != nil {
		return nil, fmt.Errorf("--generated needs the go tool: %w", err)
	}
	const module = "goldens"
	project, err := gosource.GenerateSlices(doc, cat, gosource.Options{Module: module})
	if err != nil {
		return nil, err
	}
	tmp, err := os.MkdirTemp("", "cuppa-golden-")
	if err != nil {
		return nil, err
	}
	defer func() { _ = os.RemoveAll(tmp) }()
	if err := gosource.ReplaceProject(tmp, project); err != nil {
		return nil, err
	}

	var imports, calls, pairs []string
	for i, page := range doc.Screens() {
		files := project.PageFiles[page.Name]
		if len(files) == 0 {
			return nil, fmt.Errorf("page %s has no file in the export", page.Name)
		}
		alias := fmt.Sprintf("p%d", i)
		imports = append(imports, fmt.Sprintf("\t%s %q", alias, module+"/"+path.Dir(files[0])))
		calls = append(calls, fmt.Sprintf("\t\t_ = enc.Encode(drawing{Page: %q, W: size[0], H: size[1], View: %s.View(%s.Default(), size[0], size[1]).View})", page.Name, alias, alias))
	}
	for _, s := range sizes {
		pairs = append(pairs, fmt.Sprintf("{%d, %d}", s.w, s.h))
	}
	dir := filepath.Join(tmp, "internal", "goldens")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	src := fmt.Sprintf(goldenProgram, strings.Join(imports, "\n"), strings.Join(pairs, ", "), strings.Join(calls, "\n"))
	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte(src), 0o644); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, goBin, "run", "./internal/goldens")
	cmd.Dir = tmp
	cmd.Env = append(os.Environ(), "GOFLAGS=-mod=mod", "GOWORK=off")
	var out, errs bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errs
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("the generated pages did not build or run: %w\n%s", err, strings.TrimSpace(errs.String()))
	}
	var drawings []goldenDrawing
	dec := json.NewDecoder(&out)
	for dec.More() {
		var d struct {
			Page string
			W, H int
			View string
		}
		if err := dec.Decode(&d); err != nil {
			return nil, err
		}
		drawings = append(drawings, goldenDrawing{d.Page, goldenSize{d.W, d.H}, d.View})
	}
	return drawings, nil
}
