package designcmd

import (
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"

	"github.com/meta-ui-labs/cuppa/libs/cuppafile/disk"
	"github.com/meta-ui-labs/cuppa/libs/document/design"
)

const mergeUsage = `usage:
  cuppa design merge <a.cuppa> <b.cuppa>... -o <out.cuppa> [--name <design name>] [--folder <page>=<folder>]...
                     [--master [--master-min <pages>]]

Builds one design with a page per input design, in the order given. With
--master a single design that has pages is enough: its pages are lifted. A page is
named after its design, and keeps its components, its samples and its keys; a
design that has pages gives one page each. The canvas size is the largest of
the designs, and the theme, palette and background are the first design's.
--folder says where an export keeps a page (setup/preferences).
--master lifts the components that are the same on at least --master-min pages
(3 by default) to a master page the pages share (ADR 0015); a page that did not
have one hides it. It prints what it lifted and what it left.
`

// merge builds a multi-page design from several, so an application made of
// separate designs can be exported with one runtime and opened as a whole.
func merge(args []string, stdout io.Writer) error {
	var inputs []string
	out, name := "", ""
	lift, liftMin := false, 3
	folders := map[string]string{}
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-o", "--out":
			i++
			if i < len(args) {
				out = args[i]
			}
		case "--master":
			lift = true
		case "--master-min":
			i++
			if i >= len(args) {
				return fmt.Errorf("--master-min takes a number")
			}
			n, err := strconv.Atoi(args[i])
			if err != nil || n < 2 {
				return fmt.Errorf("--master-min takes a number of at least 2, got %q", args[i])
			}
			liftMin = n
		case "--name":
			i++
			if i < len(args) {
				name = args[i]
			}
		case "--folder":
			i++
			if i >= len(args) {
				return fmt.Errorf("--folder takes <page>=<folder>")
			}
			page, folder, ok := strings.Cut(args[i], "=")
			if !ok {
				return fmt.Errorf("--folder takes <page>=<folder>, got %q", args[i])
			}
			if err := design.ValidFolder(folder); err != nil {
				return fmt.Errorf("--folder %s: %w", page, err)
			}
			folders[page] = folder
		default:
			inputs = append(inputs, args[i])
		}
	}
	if out == "" || len(inputs) == 0 || len(inputs) < 2 && !lift {
		return fmt.Errorf("%s", mergeUsage)
	}
	var base design.Document
	var screens []design.Document
	var notes []string
	for i, path := range inputs {
		doc, err := disk.Load(path)
		if err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
		if i == 0 {
			base = doc
		} else {
			if doc.Width > base.Width {
				base.Width = doc.Width
			}
			if doc.Height > base.Height {
				base.Height = doc.Height
			}
			if !reflect.DeepEqual(doc.Theme, design.Document{}.Theme) && !reflect.DeepEqual(doc.Theme, screens[0].Theme) {
				notes = append(notes, fmt.Sprintf("%s: its theme differs from the first design's; the pages share one", path))
			}
			for _, e := range doc.Embedded {
				if !hasEmbedded(base.Embedded, e.ID) {
					base.Embedded = append(base.Embedded, e)
				}
			}
		}
		screens = append(screens, doc.Screens()...)
	}
	seen := map[string]bool{}
	pages := make([]design.Page, len(screens))
	for i, s := range screens {
		if err := design.ValidPageName(s.Name); err != nil {
			return fmt.Errorf("the design %q cannot be a page: %w", s.Name, err)
		}
		if seen[s.Name] {
			return fmt.Errorf("two designs are called %q; a page needs a name of its own", s.Name)
		}
		seen[s.Name] = true
		folder := s.Folder
		if f, ok := folders[s.Name]; ok {
			folder = f
		}
		pages[i] = design.Page{Name: s.Name, Nodes: s.Nodes, Samples: s.Samples, Keys: s.Keys, Folder: folder}
		if s.Seq > base.Seq {
			base.Seq = s.Seq
		}
	}
	for page := range folders {
		if !seen[page] {
			return fmt.Errorf("--folder: there is no page called %q", page)
		}
	}
	if lift {
		master, report := liftMaster(pages, liftMin, &base)
		for _, l := range report {
			_, _ = fmt.Fprintln(stdout, l)
		}
		if len(master) > 0 {
			pages = append(pages, design.Page{Name: "Master", Master: true, Nodes: master})
		}
	}
	first := pages[0]
	if name != "" {
		base.Name = name
	}
	// The first page is the one shown, as in every saved design.
	base.Nodes, base.Samples, base.Keys = first.Nodes, first.Samples, first.Keys
	pages[0].Nodes, pages[0].Samples, pages[0].Keys = nil, nil, nil
	base.Pages, base.Page, base.PerPage = pages, 0, true
	base.Effects = screens[0].Effects
	path, err := disk.Save(out, base)
	if err != nil {
		return err
	}
	_, _ = fmt.Fprintf(stdout, "%d pages written to %s\n", len(pages), path)
	for _, n := range notes {
		_, _ = fmt.Fprintln(stdout, "note:", n)
	}
	return nil
}

func hasEmbedded(list []design.Embedded, id string) bool {
	for _, e := range list {
		if e.ID == id {
			return true
		}
	}
	return false
}
