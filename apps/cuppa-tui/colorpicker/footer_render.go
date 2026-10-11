package colorpicker

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/meta-ui-labs/cuppa/apps/cuppa-tui/theme"
)

// footer is the preview, the value field, an error line and the buttons. It is
// four lines; their rows start right after the body.
func (m *Model) footer() []string {
	y := bodyTop + m.bodyHeight() + 1
	var lines []string

	preview := "  (none)  "
	if !m.empty {
		preview = lipgloss.NewStyle().Background(lipgloss.Color(m.previewColor())).Render("          ")
	}
	r := m.newRow(y).text(" ")
	r.text(preview + " ")
	name := "no colour"
	switch {
	case m.empty:
	case m.index >= 0:
		name = fmt.Sprintf("palette %d  %s", m.index, m.rgb.Hex())
	default:
		name = m.rgb.Hex() + "  true colour"
	}
	r.text(theme.Dim(name))
	lines = append(lines, r.String())

	field := string(m.text) + theme.Title("█")
	lines = append(lines, " "+theme.Dim("Value ")+field)

	msg := ""
	if m.errMsg != "" {
		msg = " " + theme.Dim(m.errMsg)
	}
	lines = append(lines, theme.Fit(msg, innerW))

	lines = append(lines, m.buttons(y+3))
	return lines
}

func (m *Model) previewColor() string {
	if m.index >= 0 {
		return fmt.Sprintf("%d", m.index)
	}
	return m.rgb.Hex()
}

// buttons draws [ None ] on the left and [ OK ] [ Cancel ] on the right.
func (m *Model) buttons(y int) string {
	r := m.newRow(y)
	none, ok, cancel := "[ None ]", "[ OK ]", "[ Cancel ]"
	gap := max(innerW-1-ansi.StringWidth(none)-ansi.StringWidth(ok)-ansi.StringWidth(cancel)-2, 1)
	r.text(" ").span(theme.Button(none, true), func(int) { m.empty = true; m.finish("None", false) }, nil)
	r.text(strings.Repeat(" ", gap)).span(theme.Button(ok, true), func(int) { m.finish("OK", false) }, nil)
	r.text(" ").span(theme.Button(cancel, true), func(int) { m.finish("", true) }, nil)
	return r.String()
}
