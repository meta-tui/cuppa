// Package cupp reads and writes .cupp files: a pack of user-made components.
// The format is a short binary header followed by gzip-compressed JSON, like
// .cuppa. See docs/spec/cupp-format.md.
package cupp

import (
	"errors"

	"github.com/meta-ui-labs/cuppa/libs/document/design"
)

// Magic opens every .cupp file.
const Magic = "CUPP\n"

// CurrentVersion is the format version this build writes and the newest it reads.
const CurrentVersion uint16 = 1

// Extension is the file extension of a pack, with its dot.
const Extension = ".cupp"

// MaxDecodedSize bounds the JSON a file may inflate to.
const MaxDecodedSize = 16 << 20

// Errors a decode can return; test with errors.Is.
var (
	ErrNotCupp = errors.New("not a .cupp file")
	ErrTooNew  = errors.New("file was written by a newer version of Cuppa")
	ErrCorrupt = errors.New("file is damaged")
)

// Pack is the content of one .cupp file.
type Pack struct {
	// ID is lowercase letters, digits and dashes; it prefixes every component id.
	ID          string             `json:"id"`
	Name        string             `json:"name"`
	Version     string             `json:"version,omitempty"`
	Description string             `json:"description,omitempty"`
	Components  []design.Composite `json:"components"`
	// Path is where the pack was loaded from; it is not part of the file.
	Path string `json:"-"`
}

// envelope is the JSON body of the current version.
type envelope struct {
	Pack Pack `json:"pack"`
}
