package shell

import (
	"strings"
)

// binding is one key and what it does.
type binding struct {
	keys string
	what string
}

// section is a group of bindings that work in one area.
type section struct {
	area     string
	bindings []binding
}

// shortcuts is the table of keys, the one place the Shortcuts dialog is built
// from. A test checks that the user guide mentions every key in it, so the
// two cannot drift apart. Ctrl reads as Cmd on a Mac.
var shortcuts = []section{
	{"Everywhere", []binding{
		{"Ctrl+N", "New"}, {"Ctrl+O", "Open"},
		{"Ctrl+S", "Save"}, {"Ctrl+Shift+S", "Save As"},
		{"Ctrl+Z", "Undo"}, {"Ctrl+Y", "Redo"},
		{"Ctrl+C", "Copy"}, {"Ctrl+V", "Paste"},
		{"Ctrl+D", "Duplicate"}, {"Delete", "Delete"},
		{"Ctrl+G", "Group"}, {"Ctrl+U", "Ungroup"},
		{"Ctrl+]", "Forward one"}, {"Ctrl+[", "Back one"},
		{"Ctrl+Shift+]", "To front"}, {"Ctrl+Shift+[", "To back"},
		{"Ctrl+A", "Select all"}, {"Ctrl+F", "Search"},
		{"Ctrl+P", "Preview"}, {"Ctrl+Q", "Quit"},
	}},
	{"Moving around", []binding{
		{"F6", "Next area"}, {"Shift+F6", "Previous area"},
		{"Alt+1", "Jump (to Alt+4)"}, {"F10", "Open menus"},
		{"Esc", "Back, deselect"},
	}},
	{"Palette", []binding{
		{"Up", "Choose"}, {"Enter", "Place"},
		{"Left", "Fold"}, {"/", "Search"},
	}},
	{"Canvas", []binding{
		{"Arrow keys", "Move"}, {"Alt + arrow keys", "Resize"},
		{"Tab", "Next layer"}, {"Space", "Add next layer"},
	}},
	{"Details bar", []binding{
		{"Tab", "Next control"}, {"Enter", "Press or edit"},
		{"Left", "Change"}, {"h", "Hide layer"},
		{"l", "Lock layer"}, {"Alt+Up", "Layer forward"},
	}},
	{"Dialogs", []binding{
		{"Tab", "Next button"}, {"Enter", "Default button"},
		{"Esc", "Cancel"}, {"Up", "Through lists"},
	}},
}

// shortcutsText is the Shortcuts dialog: the table by area, two keys to a
// line; Ctrl becomes Cmd on a Mac.
func shortcutsText(command bool) string {
	var b strings.Builder
	for _, s := range shortcuts {
		b.WriteString(s.area + "\n")
		for i := 0; i < len(s.bindings); i += 2 {
			b.WriteString("  " + cell(s.bindings[i], 16, 14))
			if i+1 < len(s.bindings) {
				b.WriteString(cell(s.bindings[i+1], 14, 0))
			}
			b.WriteString("\n")
		}
	}
	b.WriteString("Everything can also be done with the mouse.")
	text := b.String()
	if command {
		text = strings.ReplaceAll(text, "Ctrl", "Cmd")
	}
	return text
}

// cell is a binding as keys padded to keyW, then what padded to whatW.
func cell(bd binding, keyW, whatW int) string {
	if whatW == 0 {
		return padRight(bd.keys, keyW) + bd.what
	}
	return padRight(bd.keys, keyW) + padRight(bd.what, whatW)
}

func padRight(s string, n int) string {
	if len(s) >= n {
		return s + " "
	}
	return s + strings.Repeat(" ", n-len(s))
}
