package screenscmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/meta-ui-labs/cuppa/libs/cuppafile/disk"
	"github.com/meta-ui-labs/cuppa/libs/document/design"
)

// go.mod belongs to the person: a requirement no component needs any more is
// only listed, and --update-mod removes it.
func TestGoModIsReportedAndChangedOnlyWhenAsked(t *testing.T) {
	dir := t.TempDir()
	doc := design.NewDocument("App", 60, 12)
	doc.Add(design.Node{Component: "lipgloss.box", Name: "Frame", Rect: design.Rect{X: 1, Y: 1, W: 20, H: 4}})
	doc.Export = design.ExportSettings{Layout: design.ExportSlices, Dir: "app", Module: "example.com/app"}
	file, err := disk.Save(filepath.Join(dir, "app"), doc)
	if err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	if code := Run([]string{file}, &stdout, &stderr); code != 0 {
		t.Fatalf("exit %d: %s%s", code, stdout.String(), stderr.String())
	}
	mod := filepath.Join(dir, "app", "go.mod")
	data, err := os.ReadFile(mod)
	if err != nil {
		t.Fatal(err)
	}
	const figlet = "github.com/common-nighthawk/go-figure v0.0.0-20210622060536-734e95fb86be"
	stale := strings.Replace(string(data), "\n)", "\n\t"+figlet+"\n)", 1)
	if err := os.WriteFile(mod, []byte(stale), 0o644); err != nil {
		t.Fatal(err)
	}

	stdout.Reset()
	if code := Run([]string{file}, &stdout, &stderr); code != 0 {
		t.Fatalf("exit %d: %s%s", code, stdout.String(), stderr.String())
	}
	if !strings.Contains(stdout.String(), "go.mod no longer needs "+figlet) {
		t.Errorf("the requirement is listed: %s", stdout.String())
	}
	if now, _ := os.ReadFile(mod); string(now) != stale {
		t.Error("go.mod changed without --update-mod")
	}

	stdout.Reset()
	if code := Run([]string{file, "--update-mod"}, &stdout, &stderr); code != 0 {
		t.Fatalf("exit %d: %s%s", code, stdout.String(), stderr.String())
	}
	if now, _ := os.ReadFile(mod); strings.Contains(string(now), "go-figure") {
		t.Errorf("--update-mod left it:\n%s", now)
	}
}
