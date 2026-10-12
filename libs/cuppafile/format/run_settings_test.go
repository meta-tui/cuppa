package format

import (
	"testing"

	"github.com/meta-ui-labs/cuppa/libs/document/design"
)

func TestRunSettingsSurviveSavingAndAnOldFileHasNone(t *testing.T) {
	doc := design.NewDocument("App", 40, 10)
	doc.Run = design.RunSettings{
		Build:  []string{"go", "build", "-o", "app.exe", "./cmd/app"},
		Run:    []string{"app.exe", "--demo"},
		Env:    map[string]string{"APPDATA": "scratch"},
		Dir:    "..",
		Width:  100,
		Height: 30,
		OnSave: true,
	}
	data, err := Encode(doc)
	if err != nil {
		t.Fatal(err)
	}
	got, err := Decode(data)
	if err != nil {
		t.Fatal(err)
	}
	if !got.Run.Equal(doc.Run) {
		t.Fatalf("got %+v", got.Run)
	}
	old, _ := Encode(design.NewDocument("Old", 40, 10))
	got, _ = Decode(old)
	if got.Run.IsSet() {
		t.Fatal("a design with no run block has one")
	}
}
