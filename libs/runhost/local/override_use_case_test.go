package local

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/meta-ui-labs/cuppa/libs/document/design"
)

func TestNoOverrideFileLeavesTheBlockAsItIs(t *testing.T) {
	base := design.RunSettings{Run: []string{"app"}, Env: map[string]string{"A": "1"}}
	got, err := Load(filepath.Join(t.TempDir(), "x.cuppa"), base)
	if err != nil || !got.Equal(base) {
		t.Fatalf("%+v %v", got, err)
	}
}

func TestAnOverrideWinsAndMergesTheEnvironmentByName(t *testing.T) {
	path := filepath.Join(t.TempDir(), "x.cuppa")
	dir := "C:/scratch"
	if err := Save(path, Override{
		Env:    map[string]string{"B": "2", "A": "9"},
		Dir:    &dir,
		OnSave: new(true),
	}); err != nil {
		t.Fatal(err)
	}
	base := design.RunSettings{Run: []string{"app"}, Env: map[string]string{"A": "1", "C": "3"}, Dir: "x"}
	got, err := Load(path, base)
	if err != nil {
		t.Fatal(err)
	}
	if got.Dir != dir || !got.OnSave || got.Env["A"] != "9" || got.Env["B"] != "2" || got.Env["C"] != "3" {
		t.Fatalf("got %+v", got)
	}
	if len(got.Run) != 1 || got.Run[0] != "app" {
		t.Fatalf("the program changed: %v", got.Run)
	}
	if base.Env["A"] != "1" {
		t.Fatal("the design's own env was changed")
	}
}

func TestABrokenOverrideIsAnErrorNotSilentlyIgnored(t *testing.T) {
	path := filepath.Join(t.TempDir(), "x.cuppa")
	if err := os.WriteFile(PathOf(path), []byte("{not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(path, design.RunSettings{Run: []string{"app"}}); err == nil {
		t.Fatal("a broken file was accepted")
	}
}
