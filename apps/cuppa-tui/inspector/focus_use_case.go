package inspector

// SetFocused marks the bar as the area the keyboard is on; its title shows it.
func (m *Model) SetFocused(on bool) {
	m.focused = on
	if !on {
		m.menu = nil
	}
}
