package format

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/meta-ui-labs/cuppa/libs/document/design"
)

// ToJSON is the document as the readable JSON that a .cuppa file holds inside
// its compressed body: the form to read, review and write by hand.
func ToJSON(doc design.Document) ([]byte, error) {
	return json.MarshalIndent(envelope{Document: doc}, "", "  ")
}

// FromJSON reads that JSON. Unlike Decode it refuses a field it does not know,
// so a misspelt property of a hand-written design is reported, not dropped.
// What it accepts is repaired the way a file is when it is opened.
func FromJSON(data []byte) (design.Document, error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	var env envelope
	if err := dec.Decode(&env); err != nil {
		return design.Document{}, fmt.Errorf("format: %w", err)
	}
	if env.Document.Width < 1 || env.Document.Height < 1 {
		return design.Document{}, fmt.Errorf("format: the document needs a width and a height")
	}
	return normalise(env.Document), nil
}
