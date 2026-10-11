package scene

import (
	"strconv"
	"strings"

	"github.com/meta-ui-labs/cuppa/libs/render/grid"
)

// sgrRun is text with the style the escape sequences before it gave it.
type sgrRun struct {
	text  string
	style grid.Style
}

// parseSGR splits text that may carry SGR escape sequences (colours, bold,
// dim, reverse) into runs of one style, starting from base. Other escape
// sequences are dropped. Text with no escape is one run.
func parseSGR(text string, base grid.Style) []sgrRun {
	if !strings.ContainsRune(text, 0x1b) {
		return []sgrRun{{text, base}}
	}
	var runs []sgrRun
	cur := base
	var b strings.Builder
	flush := func() {
		if b.Len() > 0 {
			runs = append(runs, sgrRun{b.String(), cur})
			b.Reset()
		}
	}
	rs := []rune(text)
	for i := 0; i < len(rs); i++ {
		if rs[i] != 0x1b {
			b.WriteRune(rs[i])
			continue
		}
		// ESC [ params final: only the final byte m changes the style.
		if i+1 < len(rs) && rs[i+1] == '[' {
			j := i + 2
			for j < len(rs) && (rs[j] < 0x40 || rs[j] > 0x7e) {
				j++
			}
			if j < len(rs) {
				if rs[j] == 'm' {
					flush()
					cur = applySGR(cur, string(rs[i+2:j]))
				}
				i = j
				continue
			}
		}
		// ESC ] ... (OSC, such as a link) up to BEL or ESC \ is dropped.
		if i+1 < len(rs) && rs[i+1] == ']' {
			j := i + 2
			for j < len(rs) && rs[j] != 0x07 && (rs[j] != 0x1b || j+1 >= len(rs) || rs[j+1] != '\\') {
				j++
			}
			if j < len(rs) && rs[j] == 0x1b {
				j++
			}
			i = j
			continue
		}
	}
	flush()
	return runs
}

// applySGR changes style by the parameters of one SGR sequence. A reset (0, 39
// or 49) means the terminal's default colours, not the style the text started
// from: text that carries its own sequences is in charge of its colours.
func applySGR(style grid.Style, params string) grid.Style {
	if params == "" {
		return grid.Style{}
	}
	parts := strings.Split(params, ";")
	num := func(i int) (int, bool) {
		if i >= len(parts) {
			return 0, false
		}
		n, err := strconv.Atoi(parts[i])
		return n, err == nil
	}
	colour := func(i int) (string, int) { // the colour after 38 or 48 and how many parameters it took
		mode, _ := num(i + 1)
		switch mode {
		case 5:
			n, _ := num(i + 2)
			return strconv.Itoa(n), 3
		case 2:
			r, _ := num(i + 2)
			g, _ := num(i + 3)
			bl, _ := num(i + 4)
			return "#" + hex2(r) + hex2(g) + hex2(bl), 5
		}
		return "", 1
	}
	for i := 0; i < len(parts); i++ {
		n, ok := num(i)
		if !ok {
			continue
		}
		switch {
		case n == 0:
			style = grid.Style{}
		case n == 1:
			style.Bold = true
		case n == 2:
			style.Dim = true
		case n == 22:
			style.Bold, style.Dim = false, false
		case n == 7:
			style.Reverse = true
		case n == 27:
			style.Reverse = false
		case n >= 30 && n <= 37:
			style.Fg = strconv.Itoa(n - 30)
		case n >= 90 && n <= 97:
			style.Fg = strconv.Itoa(n - 90 + 8)
		case n >= 40 && n <= 47:
			style.Bg = strconv.Itoa(n - 40)
		case n >= 100 && n <= 107:
			style.Bg = strconv.Itoa(n - 100 + 8)
		case n == 39:
			style.Fg = ""
		case n == 49:
			style.Bg = ""
		case n == 38 || n == 48:
			c, used := colour(i)
			if c != "" {
				if n == 38 {
					style.Fg = c
				} else {
					style.Bg = c
				}
			}
			i += used - 1
		}
	}
	return style
}

func hex2(n int) string {
	const digits = "0123456789abcdef"
	n = min(max(n, 0), 255)
	return string([]byte{digits[n>>4], digits[n&15]})
}

// runesWidth is the number of cells of the text of the runs.
func runesWidth(runs []sgrRun) int {
	n := 0
	for _, r := range runs {
		n += grid.Width(r.text)
	}
	return n
}

// cutRuns keeps the first w cells of the runs; with ellipsis a cut ends in
// "…" in the style of the last run kept.
func cutRuns(runs []sgrRun, w int, ellipsis bool) []sgrRun {
	if w <= 0 {
		return nil
	}
	total := runesWidth(runs)
	if total <= w {
		return runs
	}
	keep := w
	if ellipsis {
		keep = w - 1
	}
	var out []sgrRun
	var last grid.Style
	for _, r := range runs {
		if keep <= 0 {
			break
		}
		var kept []rune
		for _, c := range r.text {
			w := grid.RuneWidth(c)
			if w > keep {
				keep = 0
				break
			}
			kept = append(kept, c)
			keep -= w
		}
		out = append(out, sgrRun{string(kept), r.style})
		last = r.style
	}
	if ellipsis {
		out = append(out, sgrRun{"…", last})
	}
	return out
}
