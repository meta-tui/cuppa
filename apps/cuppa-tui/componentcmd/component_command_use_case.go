// Package componentcmd is the "cuppa component" command line: read the
// description a Go module publishes for its component (cuppa.component.json)
// from a file, an address or a repository, and say what it is and whether it is
// in order.
package componentcmd

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/meta-ui-labs/cuppa/libs/catalog/component"
)

const usage = `usage:
  cuppa component check <file | https://address | github.com/owner/repo[@ref]>

check reads the component description of a module (cuppa.component.json),
prints what it describes and lists anything that is wrong with it. A repository
is read from its default branch, or from the tag, branch or commit after @.
`

// Run executes "cuppa component" with the arguments after the word component
// and returns the process exit code.
func Run(args []string, stdout, stderr io.Writer) int {
	if len(args) != 2 || args[0] != "check" {
		_, _ = fmt.Fprint(stderr, usage)
		return 2
	}
	data, from, err := read(args[1])
	if err != nil {
		_, _ = fmt.Fprintln(stderr, "cuppa component:", err)
		return 1
	}
	d, issues, err := component.Decode(data)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "cuppa component: %s: %v\n", from, err)
		return 1
	}
	_, _ = fmt.Fprint(stdout, summary(d, from))
	if len(issues) > 0 {
		_, _ = fmt.Fprintf(stdout, "\n%d problem(s):\n", len(issues))
		for _, issue := range issues {
			_, _ = fmt.Fprintln(stdout, "  -", issue)
		}
		return 1
	}
	_, _ = fmt.Fprintln(stdout, "\nthe description is in order")
	return 0
}

func summary(d component.Description, from string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s  %q  (%s)\n", d.ID(), d.Title, from)
	fmt.Fprintf(&b, "  %s\n", d.Description)
	fmt.Fprintf(&b, "  go: import %s, %s.%s(...), Bubble Tea v%d, licence %s\n", d.Go.Import, d.Go.Package, d.Go.Constructor, d.Go.BubbleTea, d.License)
	fmt.Fprintf(&b, "  size: %dx%d (smallest %dx%d)\n", d.Size.Default.W, d.Size.Default.H, d.Size.Min.W, d.Size.Min.H)
	for _, p := range d.Props {
		extra := ""
		if len(p.Choices) > 0 {
			extra = " one of " + strings.Join(p.Choices, ", ")
		}
		fmt.Fprintf(&b, "    %-10s %-7s default %q%s\n", p.Key, p.Kind, p.Default, extra)
	}
	for _, e := range d.Events {
		fmt.Fprintf(&b, "  event %s: %s\n", e.Name, e.Go.Message)
	}
	return b.String()
}

// Location says where to read a description from: a file, or an address.
// "github.com/owner/repo[@ref]" is the description at the root of the
// repository; a github.com address of a repository page is the same.
func Location(arg string) (address string, isFile bool) {
	if _, err := os.Stat(arg); err == nil {
		return arg, true
	}
	rest := strings.TrimPrefix(strings.TrimPrefix(arg, "https://"), "http://")
	if strings.HasPrefix(rest, "github.com/") {
		path, ref, _ := strings.Cut(strings.TrimPrefix(rest, "github.com/"), "@")
		parts := strings.Split(strings.Trim(path, "/"), "/")
		if len(parts) == 2 {
			if ref == "" {
				ref = "HEAD"
			}
			return "https://raw.githubusercontent.com/" + parts[0] + "/" + parts[1] + "/" + ref + "/" + component.FileName, false
		}
	}
	if strings.HasPrefix(arg, "https://") || strings.HasPrefix(arg, "http://") {
		return arg, false
	}
	return arg, true
}

func read(arg string) (data []byte, from string, err error) {
	address, isFile := Location(arg)
	if isFile {
		data, err = os.ReadFile(address)
		return data, address, err
	}
	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Get(address)
	if err != nil {
		return nil, address, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, address, fmt.Errorf("%s: %s (is there a %s at the root of the repository?)", address, resp.Status, component.FileName)
	}
	data, err = io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	return data, address, err
}
