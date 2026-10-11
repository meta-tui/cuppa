// Package packcmd is the "cuppa pack" command line: turn a .cupp file into
// editable JSON, build one from JSON, and check either. It is how a pack is
// authored by hand and how a pack is reviewed before it is contributed.
package packcmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/meta-ui-labs/cuppa/libs/catalog/cupp"
	"github.com/meta-ui-labs/cuppa/libs/catalog/standard"
)

const usage = `usage:
  cuppa pack json  <file.cupp>              print a pack as editable JSON
  cuppa pack build <file.json> [out.cupp]   check the JSON and write the pack
  cuppa pack check <file.cupp|file.json>    check a pack and list its components
  cuppa pack catalog [text]                 list the components a pack can use, with property keys
`

// Run executes "cuppa pack" with the arguments after the word pack and returns
// the process exit code.
func Run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		_, _ = fmt.Fprint(stderr, usage)
		return 2
	}
	var err error
	switch args[0] {
	case "json":
		err = need(args, 1, 1, func() error { return toJSON(args[1], stdout) })
	case "build":
		err = need(args, 1, 2, func() error { return build(args[1:], stdout) })
	case "check":
		err = need(args, 1, 1, func() error { return check(args[1], stdout) })
	case "catalog":
		err = need(args, 0, 1, func() error { return catalog(args[1:], stdout) })
	default:
		_, _ = fmt.Fprint(stderr, usage)
		return 2
	}
	if err != nil {
		_, _ = fmt.Fprintln(stderr, "cuppa pack:", err)
		return 1
	}
	return 0
}

func need(args []string, min, max int, do func() error) error {
	if n := len(args) - 1; n < min || n > max {
		return fmt.Errorf("wrong number of arguments\n%s", usage)
	}
	return do()
}

// catalog lists the built-in components with the keys and kinds of their
// properties: what the "component" and "targetProp" fields of a pack refer to.
func catalog(args []string, out io.Writer) error {
	query := ""
	if len(args) == 1 {
		query = args[0]
	}
	for _, d := range standard.Default().Search(query) {
		_, _ = fmt.Fprintf(out, "%s  %q  %dx%d (min %dx%d)\n", d.ID, d.Name, d.DefaultSize.W, d.DefaultSize.H, d.MinSize.W, d.MinSize.H)
		for _, p := range d.Props {
			extra := ""
			if len(p.Choices) > 0 {
				extra = " one of " + strings.Join(p.Choices, ", ")
			}
			_, _ = fmt.Fprintf(out, "    %-14s %-7s default %q%s\n", p.Key, p.Kind, p.Default, extra)
		}
	}
	return nil
}

func toJSON(path string, out io.Writer) error {
	p, issues, err := cupp.LoadFile(path)
	if err != nil {
		return err
	}
	for _, i := range issues {
		_, _ = fmt.Fprintln(os.Stderr, "warning:", i)
	}
	data, err := cupp.ToJSON(p)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(out, string(data))
	return err
}

func build(args []string, out io.Writer) error {
	src := args[0]
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	p, err := cupp.FromJSON(data)
	if err != nil {
		return fmt.Errorf("%s: %w", filepath.Base(src), err)
	}
	dst := strings.TrimSuffix(src, filepath.Ext(src)) + cupp.Extension
	if len(args) == 2 {
		dst = args[1]
	}
	if !strings.EqualFold(filepath.Ext(dst), cupp.Extension) {
		return fmt.Errorf("the output must end in %s", cupp.Extension)
	}
	bytes, err := cupp.Encode(p)
	if err != nil {
		return err
	}
	if err := os.WriteFile(dst, bytes, 0o644); err != nil {
		return err
	}
	_, _ = fmt.Fprintf(out, "wrote %s: pack %q with %d components\n", dst, p.ID, len(p.Components))
	return nil
}

func check(path string, out io.Writer) error {
	var p cupp.Pack
	if strings.EqualFold(filepath.Ext(path), ".json") {
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if p, err = cupp.FromJSON(data); err != nil {
			return err
		}
	} else {
		loaded, issues, err := cupp.LoadFile(path)
		if err != nil {
			return err
		}
		if len(issues) > 0 {
			return fmt.Errorf("%d component(s) cannot be used: %v", len(issues), issues[0])
		}
		p = loaded
	}
	_, _ = fmt.Fprintf(out, "ok: pack %q (%s) version %q\n", p.ID, p.Name, p.Version)
	for _, c := range p.Components {
		_, _ = fmt.Fprintf(out, "  %s.%s  %q  %dx%d  %d parts, %d properties\n", p.ID, c.ID, c.Name, c.W, c.H, len(c.Nodes), len(c.Props))
	}
	return nil
}
