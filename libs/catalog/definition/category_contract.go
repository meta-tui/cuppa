package definition

import "strings"

// Category groups components by what they are for, whoever provides them:
// every list is under Data, every spinner under Feedback. The palette is
// organised by category; the Family (the provider) is shown beside a name and
// in the About window.
type Category string

// Component categories, in the order the palette lists them.
const (
	CategoryLayout     Category = "layout"
	CategoryBox        Category = "box"
	CategoryText       Category = "text"
	CategoryInput      Category = "input"
	CategorySelect     Category = "select"
	CategoryData       Category = "data"
	CategoryCharts     Category = "charts"
	CategoryFeedback   Category = "feedback"
	CategoryNavigation Category = "navigation"
	CategoryFiles      Category = "files"
	CategoryColour     Category = "colour"
	CategoryMedia      Category = "media"
	// CategoryOther holds what no other category fits, such as a component
	// from an installed pack that names none.
	CategoryOther Category = "other"
)

// CategoryOrder is the display order of the categories.
var CategoryOrder = []Category{
	CategoryBox, CategoryLayout, CategoryText, CategoryInput, CategorySelect, CategoryData,
	CategoryCharts, CategoryFeedback, CategoryNavigation, CategoryFiles, CategoryColour,
	CategoryMedia, CategoryOther,
}

// Title is the human-readable category name.
func (c Category) Title() string {
	if c == "" {
		return CategoryOther.Title()
	}
	return strings.ToUpper(string(c)[:1]) + string(c)[1:]
}

// ParseCategory returns the category a name stands for, ignoring case, and
// whether it is one of CategoryOrder.
func ParseCategory(name string) (Category, bool) {
	c := Category(strings.ToLower(strings.TrimSpace(name)))
	for _, known := range CategoryOrder {
		if c == known {
			return c, true
		}
	}
	return CategoryOther, false
}
