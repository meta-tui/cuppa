// Package message is the wire format of the live connection (ADR 0016): one
// JSON object per line, in UTF-8. Any language can speak it.
package message

import "encoding/json"

// Protocol is the version a hello carries. It changes only when a message
// changes meaning.
const Protocol = 1

// The types of message.
const (
	// Hello opens a connection: who is speaking and which pages it knows.
	Hello = "hello"
	// State is the full set of inputs of one page, sent by the application.
	State = "state"
	// Event is what a click or key on the design raised, sent by the designer.
	Event = "event"
)

// Message is every message; Type says which fields it uses.
type Message struct {
	Type     string `json:"type"`
	Protocol int    `json:"protocol,omitempty"`
	// App is the application's name and Design the design's, in a hello.
	App    string   `json:"app,omitempty"`
	Design string   `json:"design,omitempty"`
	Pages  []string `json:"pages,omitempty"`
	// Page names the page a state or event is about; Current says that the
	// application is showing that page.
	Page    string `json:"page,omitempty"`
	Current bool   `json:"current,omitempty"`
	// Inputs is a state's values by the input names of the design.
	Inputs map[string]json.RawMessage `json:"inputs,omitempty"`
	// Name is an event's name; X and Y are the cell inside the component that
	// was clicked, and 0 for a key.
	Name string `json:"name,omitempty"`
	X    int    `json:"x,omitempty"`
	Y    int    `json:"y,omitempty"`
}
