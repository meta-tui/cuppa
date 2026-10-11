// Package disk reads and writes .cuppa files on the file system.
package disk

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/meta-ui-labs/cuppa/libs/cuppafile/format"
	"github.com/meta-ui-labs/cuppa/libs/document/design"
)

// WithExtension appends .cuppa unless path already ends in it.
func WithExtension(path string) string {
	if strings.EqualFold(filepath.Ext(path), format.Extension) {
		return path
	}
	return path + format.Extension
}

// Save writes doc to path (with the .cuppa extension added if missing) and
// returns the path used. The write is atomic: a crash never leaves a half
// written file in place of the old one.
func Save(path string, doc design.Document) (string, error) {
	path = WithExtension(path)
	data, err := format.Encode(doc)
	if err != nil {
		return "", err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".cuppa-*")
	if err != nil {
		return "", err
	}
	defer func() { _ = os.Remove(tmp.Name()) }()
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return "", err
	}
	if err := tmp.Close(); err != nil {
		return "", err
	}
	if err := os.Rename(tmp.Name(), path); err != nil {
		return "", err
	}
	return path, nil
}
