// Package screenscmd is the "cuppa screens" command line: turn a folder of
// designs into a Go package of data-agnostic screens, the way a program that
// already exists takes its views from Cuppa (ADR 0007).
package screenscmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/meta-ui-labs/cuppa/libs/catalog/cupp"
	"github.com/meta-ui-labs/cuppa/libs/catalog/standard"
	"github.com/meta-ui-labs/cuppa/libs/cuppafile/disk"
	"github.com/meta-ui-labs/cuppa/libs/cuppafile/format"
	"github.com/meta-ui-labs/cuppa/libs/document/design"
	"github.com/meta-ui-labs/cuppa/libs/export/gosource"
)

const usage = `usage:
  cuppa screens <design.cuppa|folder>... -o <folder> [-p <package>] [--check]
  cuppa screens <design.cuppa> -o <folder> --app
  cuppa screens <design.cuppa> -o <folder> --layout slices [-m <module path>]
  cuppa screens <design.cuppa> -o <folder> --layout slices --library -m <import path of the folder> [--router]

Writes one contract and one view file per design, and the shared runtime, into
<folder> (a package of its own). Files it wrote earlier are replaced and the
files of designs that are gone are removed; any other file stops the export.

A design with pages gives one screen per page and a router file with the props
of the pages and which one is shown. With --app the folder is a whole
application instead: go.mod, main.go that runs the pages, the screens and the
router, in an empty folder; "cuppa main.go" opens its design again.

With --layout slices the application is written as a folder and a package per
page (pages/<page>/), the router in app/, the runtime in internal/cuppa/ and
main.go at the root of the module -m names (the design's name by default).

With --library the slices are written into a folder of a module that exists:
no go.mod, main.go or README.md; each page in the folder its page names (or
pages/<page>), the runtime in internal/cuppa and the components in
internal/components. Everything under the folder's parent of internal/ can
import them, and every page declares aliases of the runtime types it uses. The
router (app/) is written only with --router. -m is the import path of <folder>.

With --check nothing is written: the command exits 1 and lists the files that
are missing, differ from what the designs export, or are left over, so a CI job
can tell that the generated package is out of date.
`

// Run executes "cuppa screens" with the arguments after the word screens and
// returns the process exit code.
func Run(args []string, stdout, stderr io.Writer) int {
	var inputs []string
	out, pkg, layout, module := "", "", gosource.LayoutFlat, ""
	check, app, library, router := false, false, false, false
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--app":
			app = true
		case "--check":
			check = true
		case "--library":
			library = true
		case "--router":
			router = true
		case "--layout":
			i++
			if i < len(args) {
				layout = args[i]
			}
		case "-m", "--module":
			i++
			if i < len(args) {
				module = args[i]
			}
		case "-o", "--out":
			i++
			if i < len(args) {
				out = args[i]
			}
		case "-p", "--package":
			i++
			if i < len(args) {
				pkg = args[i]
			}
		default:
			inputs = append(inputs, args[i])
		}
	}
	if out == "" || len(inputs) == 0 {
		_, _ = fmt.Fprint(stderr, usage)
		return 2
	}
	if layout != gosource.LayoutFlat && layout != gosource.LayoutSlices {
		_, _ = fmt.Fprintf(stderr, "cuppa screens: --layout is %s or %s\n", gosource.LayoutFlat, gosource.LayoutSlices)
		return 2
	}
	if library && layout != gosource.LayoutSlices {
		_, _ = fmt.Fprintln(stderr, "cuppa screens: --library needs --layout slices")
		return 2
	}
	if library && module == "" {
		_, _ = fmt.Fprintln(stderr, "cuppa screens: --library needs -m, the import path of the output folder")
		return 2
	}
	if layout == gosource.LayoutSlices {
		app = true
	}
	if app {
		if err := runApp(inputs, out, layout, module, library, router, check, stdout); err != nil {
			_, _ = fmt.Fprintln(stderr, "cuppa screens:", err)
			return 1
		}
		return 0
	}
	if err := run(inputs, out, pkg, check, stdout); err != nil {
		_, _ = fmt.Fprintln(stderr, "cuppa screens:", err)
		return 1
	}
	return 0
}

func run(inputs []string, out, pkg string, check bool, stdout io.Writer) error {
	paths, err := designFiles(inputs)
	if err != nil {
		return err
	}
	var docs []design.Document
	var embedded []design.Embedded
	for _, path := range paths {
		doc, err := disk.Load(path)
		if err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
		docs = append(docs, doc)
		embedded = append(embedded, doc.Embedded...)
	}
	if pkg == "" {
		pkg = filepath.Base(filepath.Clean(out))
	}
	cat := cupp.Adopt(standard.Default(), embedded)
	var opts gosource.Options
	if len(docs) == 1 {
		// A single design travels with its screens, so it can be opened again
		// from them.
		if data, err := format.Encode(docs[0]); err == nil {
			opts.Source = data
		}
	}
	project := gosource.GenerateScreensWith(docs, cat, pkg, opts)
	if check {
		diffs := gosource.CheckScreens(out, project)
		if len(diffs) == 0 {
			_, _ = fmt.Fprintf(stdout, "%s is current: %d designs, %d files\n", out, len(docs), len(project.Files))
			return nil
		}
		for _, d := range diffs {
			_, _ = fmt.Fprintln(stdout, "  "+d)
		}
		return fmt.Errorf("%s is out of date with %d design(s): run cuppa screens again", out, len(docs))
	}
	if err := gosource.ReplaceScreens(out, project); err != nil {
		return err
	}
	names := make([]string, 0, len(project.Files))
	for name := range project.Files {
		names = append(names, name)
	}
	sort.Strings(names)
	_, _ = fmt.Fprintf(stdout, "%d designs, %d files written to %s (package %s)\n", len(docs), len(names), out, project.Module)
	for _, n := range project.Notes {
		_, _ = fmt.Fprintln(stdout, "note:", n)
	}
	for _, r := range project.Requires {
		_, _ = fmt.Fprintln(stdout, "needs in go.mod:", r)
	}
	return nil
}

// designFiles expands folders into the designs they hold, in name order.
func designFiles(inputs []string) ([]string, error) {
	var paths []string
	for _, in := range inputs {
		info, err := os.Stat(in)
		if err != nil {
			return nil, err
		}
		if !info.IsDir() {
			paths = append(paths, in)
			continue
		}
		entries, err := os.ReadDir(in)
		if err != nil {
			return nil, err
		}
		for _, e := range entries {
			if !e.IsDir() && strings.EqualFold(filepath.Ext(e.Name()), format.Extension) {
				paths = append(paths, filepath.Join(in, e.Name()))
			}
		}
	}
	if len(paths) == 0 {
		return nil, fmt.Errorf("no %s designs found", format.Extension)
	}
	return paths, nil
}

// runApp writes one design as a whole application: the screens of its pages,
// the router, main.go and go.mod.
func runApp(inputs []string, out, layout, module string, library, router, check bool, stdout io.Writer) error {
	paths, err := designFiles(inputs)
	if err != nil {
		return err
	}
	if len(paths) != 1 {
		return fmt.Errorf("--app takes one design, got %d", len(paths))
	}
	doc, err := disk.Load(paths[0])
	if err != nil {
		return fmt.Errorf("%s: %w", paths[0], err)
	}
	opts := gosource.Options{Module: module, Library: library, Router: router}
	if data, err := format.Encode(doc); err == nil {
		opts.Source = data
	}
	cat := cupp.Adopt(standard.Default(), doc.Embedded)
	project := gosource.GenerateWith(doc, cat, opts)
	if layout == gosource.LayoutSlices {
		if project, err = gosource.GenerateSlices(doc, cat, opts); err != nil {
			return err
		}
	}
	if layout == gosource.LayoutSlices {
		if check {
			diffs := gosource.CheckProject(out, project)
			if len(diffs) == 0 {
				_, _ = fmt.Fprintf(stdout, "%s is current: %d pages, %d files\n", out, doc.PageCount(), len(project.Files))
				return nil
			}
			for _, d := range diffs {
				_, _ = fmt.Fprintln(stdout, "  "+d)
			}
			return fmt.Errorf("%s is out of date with the design: run cuppa screens again", out)
		}
		if err := gosource.ReplaceProject(out, project); err != nil {
			return err
		}
	} else if err := gosource.Write(out, project); err != nil {
		return err
	}
	_, _ = fmt.Fprintf(stdout, "%d pages, %d files written to %s\nNext: cd %s && go mod tidy && go run .\n", doc.PageCount(), len(project.Files), out, out)
	for _, n := range project.Notes {
		_, _ = fmt.Fprintln(stdout, "note:", n)
	}
	return nil
}
