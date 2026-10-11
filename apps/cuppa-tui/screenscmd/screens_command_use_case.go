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

Writes one contract and one view file per design, and the shared runtime, into
<folder> (a package of its own). Files it wrote earlier are replaced and the
files of designs that are gone are removed; any other file stops the export.

A design with pages gives one screen per page and a router file with the props
of the pages and which one is shown. With --app the folder is a whole
application instead: go.mod, main.go that runs the pages, the screens and the
router, in an empty folder; "cuppa main.go" opens its design again.

With --check nothing is written: the command exits 1 and lists the files that
are missing, differ from what the designs export, or are left over, so a CI job
can tell that the generated package is out of date.
`

// Run executes "cuppa screens" with the arguments after the word screens and
// returns the process exit code.
func Run(args []string, stdout, stderr io.Writer) int {
	var inputs []string
	out, pkg := "", ""
	check, app := false, false
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--app":
			app = true
		case "--check":
			check = true
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
	if app {
		if err := runApp(inputs, out, stdout); err != nil {
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
func runApp(inputs []string, out string, stdout io.Writer) error {
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
	opts := gosource.Options{}
	if data, err := format.Encode(doc); err == nil {
		opts.Source = data
	}
	project := gosource.GenerateWith(doc, cupp.Adopt(standard.Default(), doc.Embedded), opts)
	if err := gosource.Write(out, project); err != nil {
		return err
	}
	_, _ = fmt.Fprintf(stdout, "%d pages, %d files written to %s\nNext: cd %s && go mod tidy && go run .\n", doc.PageCount(), len(project.Files), out, out)
	for _, n := range project.Notes {
		_, _ = fmt.Fprintln(stdout, "note:", n)
	}
	return nil
}
