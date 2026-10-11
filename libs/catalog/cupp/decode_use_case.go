package cupp

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"

	"github.com/meta-ui-labs/cuppa/libs/document/design"
)

// Decode reads a .cupp file. Components that cannot be used are dropped and
// reported in the returned problems, so one bad component never costs the pack.
func Decode(data []byte) (Pack, []error, error) {
	if !bytes.HasPrefix(data, []byte(Magic)) {
		return Pack{}, nil, ErrNotCupp
	}
	rest := data[len(Magic):]
	if len(rest) < 2 {
		return Pack{}, nil, ErrCorrupt
	}
	version := binary.BigEndian.Uint16(rest)
	if version > CurrentVersion {
		return Pack{}, nil, ErrTooNew
	}
	if version == 0 {
		return Pack{}, nil, ErrCorrupt
	}
	zr, err := gzip.NewReader(bytes.NewReader(rest[2:]))
	if err != nil {
		return Pack{}, nil, ErrCorrupt
	}
	body, err := io.ReadAll(io.LimitReader(zr, MaxDecodedSize+1))
	if err != nil || len(body) > MaxDecodedSize {
		return Pack{}, nil, ErrCorrupt
	}
	var env envelope
	if json.Unmarshal(body, &env) != nil {
		return Pack{}, nil, ErrCorrupt
	}
	p := env.Pack
	if !design.ValidID(p.ID) || p.Name == "" {
		return Pack{}, nil, ErrCorrupt
	}
	var problems []error
	var kept []design.Composite
	seen := map[string]bool{}
	for _, c := range p.Components {
		if err := c.Check(); err != nil {
			problems = append(problems, err)
			continue
		}
		if seen[c.ID] {
			problems = append(problems, fmt.Errorf("component %q appears twice", c.ID))
			continue
		}
		seen[c.ID] = true
		kept = append(kept, c)
	}
	p.Components = kept
	return p, problems, nil
}

// check is what Encode requires before it writes.
func (p Pack) check() error {
	if !design.ValidID(p.ID) {
		return fmt.Errorf("pack id %q must be lowercase letters, digits and dashes", p.ID)
	}
	if p.Name == "" {
		return fmt.Errorf("pack %q has no name", p.ID)
	}
	seen := map[string]bool{}
	for _, c := range p.Components {
		if err := c.Check(); err != nil {
			return err
		}
		if seen[c.ID] {
			return fmt.Errorf("component %q appears twice", c.ID)
		}
		seen[c.ID] = true
	}
	return nil
}
