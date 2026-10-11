package gosource

import (
	_ "embed"
	"regexp"
	"strings"
)

// colourPickerLibrary is colourpicker.go of github.com/meta-ui-labs/bubble-colourpicker,
// byte for byte (a test compares it with the module). The exported program
// carries a renamed copy of it, so the component the designer shows and the
// one in the program are the same code.
//
//go:embed colourpicker_go.txt
var colourPickerLibrary string

// colourPickerVersion is the release of the library the copy is of.
const colourPickerVersion = "v0.1.3"

// colourPickerNames are the exported names of the library. In a program they
// share a package with the rest of the generated code, so each is prefixed.
var colourPickerNames = []string{
	"WithValue", "WithTabs", "WithTab", "WithAccent", "WithOrigin",
	"ChangedMsg", "Option", "Model", "New", "Tab256", "Tab16", "TabRGB", "TabHSL",
}

var colourPickerWord = func() *regexp.Regexp {
	return regexp.MustCompile(`(^|[^.\w]|\.\.\.)(` + strings.Join(colourPickerNames, "|") + `)\b`)
}()

// colourPickerSource is the library as a file of the generated program: the
// package is main (the screens export changes it), the exported names are
// prefixed (Model becomes ColourPicker, New becomes NewColourPicker, the
// options ColourPickerWithValue and so on) and the component is registered
// under its catalog id.
func colourPickerSource() string {
	src := strings.ReplaceAll(colourPickerLibrary, "\r\n", "\n")
	if i := strings.Index(src, "package colourpicker"); i >= 0 {
		src = src[i:]
	}
	src = strings.Replace(src, "package colourpicker", "package main", 1)
	src = colourPickerWord.ReplaceAllStringFunc(src, func(m string) string {
		loc := colourPickerWord.FindStringSubmatch(m)
		name := loc[2]
		switch name {
		case "Model":
			name = "ColourPicker"
		case "New":
			name = "NewColourPicker"
		default:
			name = "ColourPicker" + name
		}
		return loc[1] + name
	})
	header := "// colourpicker.go is the colour picker component, a copy of colourpicker.go of\n" +
		"// github.com/meta-ui-labs/bubble-colourpicker " + colourPickerVersion + ", with the exported names prefixed\n" +
		"// (Model is ColourPicker, New is NewColourPicker) so they can share a package.\n" +
		"// Use the module itself for a program of your own.\n\n"
	return header + src + colourPickerGlue
}

// colourPickerGlue draws the component of a design from its properties.
const colourPickerGlue = `
func init() { extra["lipgloss.colourpicker"] = colourPickerPart }

// colourPickerPart draws the colour picker of the design: the same drawing as
// the picker a program owns, from the value, tabs, tab and slider of its
// properties.
func colourPickerPart(c placed, p props) *part {
	m := NewColourPicker(
		ColourPickerWithValue(p.str("value")),
		ColourPickerWithTabs(split(p.str("tabs"), ",")...),
		ColourPickerWithTab(p.str("tab")),
		ColourPickerWithAccent(orDefault(p.str("color"), "212")),
	).SetSlide(p.integer("slide", 0))
	return staticPart(m.View())
}
`
