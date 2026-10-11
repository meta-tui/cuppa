package format

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"errors"
	"strings"

	"github.com/meta-tui/cuppa/libs/document/design"
)

// GoDirective starts the line of a generated Go file that carries the design
// the application was made from (ADR 0012).
const GoDirective = "//cuppa:design "

// ErrNoDesign says a Go file has no design in it.
var ErrNoDesign = errors.New("no Cuppa design in this Go source")

// FromGo reads the design embedded in the source of an application Cuppa
// exported, so the application can be opened again from its Go.
func FromGo(src []byte) (design.Document, error) {
	sc := bufio.NewScanner(bytes.NewReader(src))
	sc.Buffer(make([]byte, 0, 64*1024), MaxDecodedSize*2)
	for sc.Scan() {
		text, ok := strings.CutPrefix(sc.Text(), GoDirective)
		if !ok {
			continue
		}
		data, err := base64.StdEncoding.DecodeString(strings.TrimSpace(text))
		if err != nil {
			return design.Document{}, ErrCorrupt
		}
		return Decode(data)
	}
	return design.Document{}, ErrNoDesign
}
