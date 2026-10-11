package format

import (
	"encoding/json"

	"github.com/meta-ui-labs/cuppa/libs/document/design"
)

// envelope is the JSON body of the current version.
type envelope struct {
	Document design.Document `json:"document"`
}

// rawEnvelope is a body of any version, migrated step by step before use.
type rawEnvelope = json.RawMessage
