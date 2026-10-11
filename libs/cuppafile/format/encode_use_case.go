package format

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"encoding/json"

	"github.com/meta-ui-labs/cuppa/libs/document/design"
)

// Encode returns the bytes of a .cuppa file holding doc.
func Encode(doc design.Document) ([]byte, error) {
	body, err := json.Marshal(envelope{Document: doc})
	if err != nil {
		return nil, err
	}
	var out bytes.Buffer
	out.WriteString(Magic)
	_ = binary.Write(&out, binary.BigEndian, CurrentVersion)
	zw := gzip.NewWriter(&out)
	if _, err := zw.Write(body); err != nil {
		return nil, err
	}
	if err := zw.Close(); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}
