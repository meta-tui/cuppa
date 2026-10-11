package disk

import (
	"encoding/base64"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/meta-tui/cuppa/libs/cuppafile/format"
	"github.com/meta-tui/cuppa/libs/document/design"
)

func TestSaveThenLoad(t *testing.T) {
	doc := design.NewDocument("Demo", 40, 10)
	doc.Add(design.Node{Component: "lipgloss.box", Name: "Box", Rect: design.Rect{W: 5, H: 3}})
	dir := t.TempDir()
	path, err := Save(filepath.Join(dir, "demo"), doc)
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(path) != "demo.cuppa" {
		t.Fatalf("extension not added: %s", path)
	}
	got, err := Load(path)
	if err != nil || !reflect.DeepEqual(got, doc) {
		t.Fatalf("load: %v %+v", err, got)
	}
	entries, _ := os.ReadDir(dir)
	if len(entries) != 1 {
		t.Fatalf("temp file left behind: %v", entries)
	}
}

func TestLoadErrors(t *testing.T) {
	dir := t.TempDir()
	if _, err := Load(filepath.Join(dir, "missing.cuppa")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("missing: %v", err)
	}
	bad := filepath.Join(dir, "bad.cuppa")
	_ = os.WriteFile(bad, []byte("hello"), 0o644)
	if _, err := Load(bad); !errors.Is(err, format.ErrNotCuppa) {
		t.Fatalf("bad: %v", err)
	}
}

func TestSaveKeepsExistingExtension(t *testing.T) {
	if got := WithExtension("a/B.CUPPA"); got != "a/B.CUPPA" {
		t.Fatal(got)
	}
}

// An application's main.go has no design in it, but the router beside it does.
func TestLoadReadsTheDesignOutOfAnApplicationsGo(t *testing.T) {
	doc := design.NewDocument("App", 40, 10)
	doc.Add(design.Node{Component: "lipgloss.box", Name: "Box", Rect: design.Rect{W: 5, H: 3}})
	data, err := format.Encode(doc)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main\n\nfunc main() {}\n"), 0o644)
	if _, err := Load(filepath.Join(dir, "main.go")); !errors.Is(err, format.ErrNoDesign) {
		t.Fatalf("no router yet: %v", err)
	}
	router := format.GoDirective + base64.StdEncoding.EncodeToString(data) + "\n\npackage main\n"
	_ = os.WriteFile(filepath.Join(dir, "app_app.go"), []byte(router), 0o644)
	got, err := Load(filepath.Join(dir, "main.go"))
	if err != nil || got.Name != "App" || len(got.Nodes) != 1 {
		t.Fatalf("load: %v %+v", err, got)
	}
	if !IsGoSource(filepath.Join(dir, "main.go")) || IsGoSource("a.cuppa") {
		t.Error("IsGoSource")
	}
}
