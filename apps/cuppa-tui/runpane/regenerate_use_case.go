package runpane

import (
	"fmt"

	"github.com/meta-ui-labs/cuppa/libs/catalog/cupp"
	"github.com/meta-ui-labs/cuppa/libs/catalog/standard"
	"github.com/meta-ui-labs/cuppa/libs/cuppafile/disk"
	"github.com/meta-ui-labs/cuppa/libs/cuppafile/format"
	"github.com/meta-ui-labs/cuppa/libs/export/gosource"
)

// Regenerate writes the Go source of the design saved at path to where the
// design says it goes, before the program is built. A file that no longer
// belongs to the design is kept, never deleted, and go.mod is left alone:
// both are questions for the person, asked by the export itself. It does
// nothing, and returns false, when the design does not say where its source
// goes.
func Regenerate(path string) (bool, error) {
	doc, err := disk.Load(path)
	if err != nil {
		return false, fmt.Errorf("regenerate: %w", err)
	}
	if !doc.Export.IsSet() {
		return false, nil
	}
	var source []byte
	if data, err := format.Encode(doc); err == nil {
		source = data
	}
	project, err := gosource.GenerateExport(doc, cupp.Adopt(standard.Default(), doc.Embedded), doc.Export, source)
	if err != nil {
		return true, fmt.Errorf("regenerate: %w", err)
	}
	dir := gosource.ExportDir(doc.Export, path)
	keep := gosource.ObsoleteExport(dir, doc.Export, project)
	if err := gosource.WriteExportKeeping(dir, doc.Export, project, keep); err != nil {
		return true, fmt.Errorf("regenerate: %w", err)
	}
	return true, nil
}
