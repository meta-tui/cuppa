package fileflow

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/meta-ui-labs/cuppa/libs/document/design"
)

// A design that no longer needs a module its go.mod has asks before go.mod is
// changed, and Skip leaves it.
func TestExportAsksBeforeChangingGoMod(t *testing.T) {
	f, ed, dir := setup(t)
	if err := ed.AddPage("Settings"); err != nil {
		t.Fatal(err)
	}
	app := filepath.Join(dir, "shop")
	if err := ed.SetExportSettings(design.ExportSettings{Layout: design.ExportSlices, Dir: app, Module: "example.com/shop"}); err != nil {
		t.Fatal(err)
	}
	f.ExportGoSource()
	press(f, true, false) // Export
	press(f, true, false) // OK
	mod := filepath.Join(app, "go.mod")
	data, err := os.ReadFile(mod)
	if err != nil {
		t.Fatal(err)
	}
	// An earlier design needed the figlet module; this one does not.
	stale := strings.Replace(string(data), "\n)\n", "\n\tgithub.com/common-nighthawk/go-figure v0.0.0-20210622060536-734e95fb86be\n)\n", 1)
	if err := os.WriteFile(mod, []byte(stale), 0o644); err != nil {
		t.Fatal(err)
	}

	f.ExportGoSource()
	press(f, true, false) // Export
	text := stripped(strings.Join(f.Modal().Lines(), "\n"))
	if !strings.Contains(text, "Update go.mod") || !strings.Contains(text, "remove  github.com/common-nighthawk/go-figure") {
		t.Fatalf("the change to go.mod is listed: %s", text)
	}
	if now, _ := os.ReadFile(mod); string(now) != stale {
		t.Fatal("go.mod changed before the answer")
	}
	clickButton(t, f, "Skip")
	if now, _ := os.ReadFile(mod); string(now) != stale {
		t.Fatal("Skip changed go.mod")
	}
	press(f, true, false) // OK of the notice

	f.ExportGoSource()
	press(f, true, false)
	press(f, true, false) // Update go.mod, the first button
	now, _ := os.ReadFile(mod)
	if strings.Contains(string(now), "go-figure") {
		t.Fatalf("Update left the requirement:\n%s", now)
	}
}
