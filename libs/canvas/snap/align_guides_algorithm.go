// Package snap finds the alignment a moving rectangle should snap to.
package snap

import "github.com/meta-ui-labs/cuppa/libs/document/design"

// Guide is an alignment line: a vertical line at column Pos when Vertical,
// otherwise a horizontal line at row Pos.
type Guide struct {
	Vertical bool
	Pos      int
}

// Result is the correction to apply to the moving rectangle and the guides
// that explain it.
type Result struct {
	DX, DY int
	Guides []Guide
}

// lines are the three alignment lines of a rectangle along one axis: first
// cell, middle cell and last cell.
func lines(start, length int) [3]int {
	return [3]int{start, start + (length-1)/2, start + length - 1}
}

// Snap returns how far moving should shift so one of its edges or its centre
// lines up with an edge or centre of another rectangle (or of the canvas),
// when that is within threshold cells. Axes snap independently.
func Snap(moving design.Rect, others []design.Rect, canvas design.Rect, threshold int) Result {
	targets := append([]design.Rect{canvas}, others...)
	var res Result
	res.DX, res.Guides = snapAxis(lines(moving.X, moving.W), threshold, true, func(yield func(int)) {
		for _, t := range targets {
			for _, v := range lines(t.X, t.W) {
				yield(v)
			}
		}
	}, res.Guides)
	var dy int
	dy, res.Guides = snapAxis(lines(moving.Y, moving.H), threshold, false, func(yield func(int)) {
		for _, t := range targets {
			for _, v := range lines(t.Y, t.H) {
				yield(v)
			}
		}
	}, res.Guides)
	res.DY = dy
	return res
}

// snapAxis picks the smallest correction within threshold and records a guide
// for every target line the corrected rectangle then shares.
func snapAxis(mine [3]int, threshold int, vertical bool, each func(func(int)), guides []Guide) (int, []Guide) {
	best, found := 0, false
	each(func(target int) {
		for _, m := range mine {
			d := target - m
			if abs(d) > threshold {
				continue
			}
			if !found || abs(d) < abs(best) {
				best, found = d, true
			}
		}
	})
	if !found {
		return 0, guides
	}
	seen := map[int]bool{}
	each(func(target int) {
		for _, m := range mine {
			if m+best == target && !seen[target] {
				seen[target] = true
				guides = append(guides, Guide{Vertical: vertical, Pos: target})
			}
		}
	})
	return best, guides
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}
