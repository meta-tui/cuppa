// Package designcmd is the "cuppa design" command line: turn a .cuppa design
// into editable JSON, build one from JSON, and check one. It is how a design is
// written by a script or an agent and how a design is reviewed.
package designcmd

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/meta-ui-labs/cuppa/libs/catalog/standard"
	"github.com/meta-ui-labs/cuppa/libs/cuppafile/disk"
	"github.com/meta-ui-labs/cuppa/libs/cuppafile/format"
	"github.com/meta-ui-labs/cuppa/libs/document/design"
	"github.com/meta-ui-labs/cuppa/libs/export/gosource"
)

const usage = `usage:
  cuppa design json  <file.cuppa>             print a design as editable JSON
  cuppa design build <file.json> [out.cuppa]  check the JSON and write the design
  cuppa design check <file.cuppa|file.json>   list what is wrong with a design and what it exports

A design is {"document": {"name", "width", "height", "background", "theme",
"keys", "nodes": [...]}}. A node has "id", "component" (see "cuppa pack
catalog"), "name", "rect" {"X","Y","W","H"}, "props" (string values), and for
screens "bind" (property key -> input name), "showIf" and "event"; "layout"
holds expressions such as {"w": "100% - 4", "y": "100% - 2"}. Unknown fields
and unknown properties are refused.
`

// Run executes "cuppa design" with the arguments after the word design and
// returns the process exit code.
func Run(args []string, stdout, stderr io.Writer) int {
	if len(args) < 2 {
		_, _ = fmt.Fprint(stderr, usage)
		return 2
	}
	var err error
	switch args[0] {
	case "json":
		err = toJSON(args[1], stdout)
	case "build":
		out := ""
		if len(args) > 2 {
			out = args[2]
		}
		err = build(args[1], out, stdout)
	case "check":
		err = check(args[1], stdout)
	default:
		_, _ = fmt.Fprint(stderr, usage)
		return 2
	}
	if err != nil {
		_, _ = fmt.Fprintln(stderr, "cuppa design:", err)
		return 1
	}
	return 0
}

func toJSON(path string, stdout io.Writer) error {
	doc, err := disk.Load(path)
	if err != nil {
		return err
	}
	data, err := format.ToJSON(doc)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(stdout, string(data))
	return err
}

func build(path, out string, stdout io.Writer) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	doc, err := format.FromJSON(data)
	if err != nil {
		return err
	}
	if problems := problemsOf(doc); len(problems) > 0 {
		return fmt.Errorf("the design has problems:\n  - %s", strings.Join(problems, "\n  - "))
	}
	if out == "" {
		out = strings.TrimSuffix(path, ".json") + ".cuppa"
	}
	written, err := disk.Save(out, doc)
	if err != nil {
		return err
	}
	_, _ = fmt.Fprintf(stdout, "wrote %s: %q, %dx%d, %d components\n", written, doc.Name, doc.Width, doc.Height, count(doc.Nodes))
	return nil
}

func check(path string, stdout io.Writer) error {
	var doc design.Document
	var err error
	if strings.HasSuffix(strings.ToLower(path), ".json") {
		var data []byte
		if data, err = os.ReadFile(path); err == nil {
			doc, err = format.FromJSON(data)
		}
	} else {
		doc, err = disk.Load(path)
	}
	if err != nil {
		return err
	}
	problems := problemsOf(doc)
	cat := standard.Default()
	export := gosource.GenerateScreens([]design.Document{doc}, cat, "screens")
	_, _ = fmt.Fprintf(stdout, "%q, %dx%d, %d components\n", doc.Name, doc.Width, doc.Height, count(doc.Nodes))
	if len(problems) == 0 && len(export.Notes) == 0 {
		_, _ = fmt.Fprintln(stdout, "in order; exports as a screen without notes")
		return nil
	}
	for _, p := range problems {
		_, _ = fmt.Fprintln(stdout, "  problem:", p)
	}
	for _, n := range export.Notes {
		_, _ = fmt.Fprintln(stdout, "  note:", n)
	}
	if len(problems) > 0 {
		return fmt.Errorf("%d problem(s)", len(problems))
	}
	return nil
}

// problemsOf lists what makes a design wrong rather than merely unusual:
// components that are not in the catalog, properties they do not have,
// properties bound or values that do not fit, and sizes below the smallest.
func problemsOf(doc design.Document) []string {
	cat := standard.Default()
	var out []string
	var walk func(nodes []design.Node)
	walk = func(nodes []design.Node) {
		for _, n := range nodes {
			if n.IsGroup() {
				walk(n.Children)
				continue
			}
			def, ok := cat.Get(n.Component)
			if !ok {
				out = append(out, fmt.Sprintf("%s: unknown component %q (see: cuppa pack catalog)", n.Name, n.Component))
				continue
			}
			for key := range n.Props {
				if _, ok := def.Prop(key); !ok {
					out = append(out, fmt.Sprintf("%s: %s has no property %q", n.Name, n.Component, key))
				}
			}
			for key := range n.Bind {
				if _, ok := def.Prop(key); !ok {
					out = append(out, fmt.Sprintf("%s: cannot bind %q, %s has no such property", n.Name, key, n.Component))
				}
			}
			if n.Rect.W < def.MinSize.W || n.Rect.H < def.MinSize.H {
				out = append(out, fmt.Sprintf("%s: %dx%d is smaller than the %dx%d %s allows", n.Name, n.Rect.W, n.Rect.H, def.MinSize.W, def.MinSize.H, n.Component))
			}
		}
	}
	walk(doc.Nodes)
	return out
}

func count(nodes []design.Node) int {
	n := 0
	for _, node := range nodes {
		n++
		n += count(node.Children)
	}
	return n
}
