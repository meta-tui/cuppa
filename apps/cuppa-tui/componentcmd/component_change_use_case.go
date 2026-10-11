package componentcmd

import (
	"fmt"
	"io"
	"strings"

	"github.com/meta-ui-labs/cuppa/libs/canvas/editor"
	"github.com/meta-ui-labs/cuppa/libs/catalog/standard"
	"github.com/meta-ui-labs/cuppa/libs/cuppafile/disk"
	"github.com/meta-ui-labs/cuppa/libs/document/design"
)

// runChange is "cuppa component options" and "cuppa component change".
func runChange(args []string, stdout, stderr io.Writer) int {
	var positional []string
	allowLoss, out := false, ""
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--allow-loss":
			allowLoss = true
		case "-o", "--out":
			i++
			if i < len(args) {
				out = args[i]
			}
		default:
			positional = append(positional, args[i])
		}
	}
	want := 3
	if positional[0] == "change" {
		want = 4
	}
	if len(positional) != want {
		_, _ = fmt.Fprint(stderr, usage)
		return 2
	}
	file, node := positional[1], positional[2]
	doc, err := disk.Load(file)
	if err != nil {
		_, _ = fmt.Fprintln(stderr, "cuppa component:", err)
		return 1
	}
	cat := standard.Default()
	ed := editor.New(cat, doc)
	id, err := findNode(doc, node)
	if err != nil {
		_, _ = fmt.Fprintln(stderr, "cuppa component:", err)
		return 1
	}

	if positional[0] == "options" {
		n, _ := doc.Get(id)
		options := ed.ChangeOptions(id)
		_, _ = fmt.Fprintf(stdout, "%s (%s) can be changed for:\n", n.Name, n.Component)
		for _, d := range options {
			var ports []string
			for _, p := range d.Ports() {
				ports = append(ports, string(p))
			}
			_, _ = fmt.Fprintf(stdout, "  %-26s %-20s ports: %s\n", d.ID, d.Name, strings.Join(ports, ", "))
		}
		if len(options) == 0 {
			_, _ = fmt.Fprintln(stdout, "  (nothing: no component carries the data bound to it)")
		}
		return 0
	}

	report, err := ed.ChangeComponent(id, positional[3], allowLoss)
	if err != nil {
		_, _ = fmt.Fprintln(stderr, "cuppa component:", strings.TrimPrefix(err.Error(), "editor: "))
		return 1
	}
	if out == "" {
		out = file
	}
	if _, err := disk.Save(out, ed.Document()); err != nil {
		_, _ = fmt.Fprintln(stderr, "cuppa component:", err)
		return 1
	}
	n, _ := ed.Document().Get(id)
	_, _ = fmt.Fprintf(stdout, "%s is now %s\n", n.Name, n.Component)
	for _, m := range report.Moves {
		_, _ = fmt.Fprintf(stdout, "  %s -> %s (%s)\n", m.From, m.To, m.Port)
	}
	for _, key := range report.Lost {
		_, _ = fmt.Fprintf(stdout, "  removed the variable bound to %s\n", key)
	}
	for _, key := range report.Added {
		_, _ = fmt.Fprintf(stdout, "  new, not bound: %s\n", key)
	}
	_, _ = fmt.Fprintf(stdout, "written to %s\n", out)
	return 0
}

// findNode finds a component of the design by id or by name; a name that two
// components share is refused.
func findNode(doc design.Document, ref string) (design.NodeID, error) {
	var found []design.Node
	var walk func(nodes []design.Node)
	walk = func(nodes []design.Node) {
		for _, n := range nodes {
			if string(n.ID) == ref || strings.EqualFold(n.Name, ref) {
				found = append(found, n)
			}
			walk(n.Children)
		}
	}
	walk(doc.Nodes)
	switch len(found) {
	case 0:
		return "", fmt.Errorf("no component called %q in the design", ref)
	case 1:
		return found[0].ID, nil
	}
	return "", fmt.Errorf("%d components are called %q; use an id", len(found), ref)
}
