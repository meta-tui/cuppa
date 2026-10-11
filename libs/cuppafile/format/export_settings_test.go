package format

import (
	"testing"

	"github.com/meta-ui-labs/cuppa/libs/document/design"
)

func TestExportSettingsSurviveSavingAndAnOldFileHasNone(t *testing.T) {
	doc := design.NewDocument("App", 40, 10)
	doc.Export = design.ExportSettings{Layout: design.ExportSlices, Library: true, Dir: "../libs/core", Module: "m/core"}
	data, err := Encode(doc)
	if err != nil {
		t.Fatal(err)
	}
	got, err := Decode(data)
	if err != nil {
		t.Fatal(err)
	}
	if got.Export != doc.Export {
		t.Fatalf("got %+v", got.Export)
	}
	doc.Export = design.ExportSettings{Layout: "zip", Dir: "x"}
	data, _ = Encode(doc)
	got, _ = Decode(data)
	if got.Export.IsSet() {
		t.Fatalf("an unknown layout was kept: %+v", got.Export)
	}
	old, _ := Encode(design.NewDocument("Old", 40, 10))
	got, _ = Decode(old)
	if got.Export.IsSet() {
		t.Fatal("a design with no settings has some")
	}
}
