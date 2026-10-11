// Package modal is the contract every dialog box fulfils, so the shell can
// route input to whichever one is open without knowing what it is.
package modal

import (
	"github.com/meta-ui-labs/cuppa/apps/cuppa-tui/pointer"
	"github.com/meta-ui-labs/cuppa/libs/document/design"
)

// Outcome is how a dialog ended.
type Outcome struct {
	// Button is the label of the button that was pressed.
	Button string
	// Path is the file the user picked, for file dialogs.
	Path string
	// Value is what the user chose in a value dialog, such as a colour.
	Value string
	// Canceled is true when the user dismissed the dialog (Esc, Cancel).
	Canceled bool
}

// Modal is a box centred over the app that takes all input until it ends.
type Modal interface {
	// Place centres the box on a screen of the given size.
	Place(screenW, screenH int)
	// Rect is where the box sits on screen after Place.
	Rect() design.Rect
	// Lines renders the box: Rect().H lines of Rect().W cells.
	Lines() []string
	// Handle takes a pointer event in screen coordinates.
	Handle(e pointer.Event)
	// Key takes typed text and the editing keys.
	Key(text string, back, enter, esc bool)
	// Outcome reports whether the dialog has ended and how.
	Outcome() (Outcome, bool)
}
