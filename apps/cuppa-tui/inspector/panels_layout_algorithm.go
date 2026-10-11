package inspector

// panelHeights decides how many body rows each panel gets in a bar with avail
// rows for bodies (the headers are not counted). A closed panel gets none; an
// open one gets the height it asked for, except that the last open panel takes
// whatever is left, so the stack always fills the bar. When the bar is too
// short, the panels give up rows from the bottom up, never going under their
// minimum until nothing else is left to give.
func panelHeights(avail int, wants, mins []int, open []bool) []int {
	h := make([]int, len(wants))
	last, others := -1, 0
	for i := range wants {
		if !open[i] {
			continue
		}
		last = i
		h[i] = max(wants[i], mins[i])
		others += h[i]
	}
	if last < 0 {
		return h
	}
	others -= h[last]
	h[last] = max(avail-others, mins[last])
	over := others + h[last] - avail
	for i := last - 1; i >= 0 && over > 0; i-- {
		if open[i] {
			cut := min(over, h[i]-mins[i])
			h[i] -= cut
			over -= cut
		}
	}
	for i := last; i >= 0 && over > 0; i-- {
		if open[i] {
			cut := min(over, h[i])
			h[i] -= cut
			over -= cut
		}
	}
	return h
}

// thumb is where the scroll bar's thumb sits in a body of rows rows that shows
// a slice of total lines starting at scroll: its first row and its length.
func thumb(rows, total, scroll int) (top, length int) {
	if total <= rows || rows <= 0 {
		return 0, 0
	}
	length = max(rows*rows/total, 1)
	span := total - rows
	top = scroll * (rows - length) / span
	return top, length
}
