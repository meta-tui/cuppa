package disk

import (
	"os"

	"github.com/meta-ui-labs/cuppa/libs/cuppafile/format"
	"github.com/meta-ui-labs/cuppa/libs/document/design"
)

// Load reads the design stored at path, a .cuppa file.
func Load(path string) (design.Document, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return design.Document{}, err
	}
	return format.Decode(data)
}
