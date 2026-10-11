package standard

import "github.com/meta-ui-labs/cuppa/libs/catalog/definition"

// Community entries are drawn as faithful approximations. Which Bubble Tea
// major version each library targets is recorded in docs/catalog/community-components.md.
func communityEntries() []definition.Definition {
	accent := colorProp("color", "Color", "212")
	return []definition.Definition{
		{
			ID: "community.bubbletable", Name: "Bubble table", Family: definition.FamilyCommunity,
			Description: "Evertras/bubble-table: paginated, sortable table with a footer.",
			DefaultSize: size(44, 9), MinSize: size(16, 5), Import: "github.com/evertras/bubble-table/table", Status: definition.StatusPlaceholder,
			Props: []definition.PropSpec{
				textProp("columns", "Columns (comma separated)", "Name,Kind,Qty"),
				textProp("rows", "Rows (; and , separated)", "Earl Grey,Black,2;Sencha,Green,1;Rooibos,Herbal,4"),
				textProp("footer", "Footer", "Page 1/3"),
				accent,
			},
		},
		{
			ID: "community.flexbox", Name: "Flex box", Family: definition.FamilyCommunity,
			Description: "76creates/stickers FlexBox: responsive grid of cells (Bubble Tea v1).",
			DefaultSize: size(40, 8), MinSize: size(10, 3), Import: "github.com/76creates/stickers/flexbox", Status: definition.StatusPlaceholder,
			Props: []definition.PropSpec{intProp("rows", "Rows", 2, 1, 6), intProp("columns", "Columns", 3, 1, 6), accent},
		},
		{
			ID: "community.boxer", Name: "Boxer layout", Family: definition.FamilyCommunity,
			Description: "treilik/bubbleboxer: a layout tree of side-by-side models (Bubble Tea v1-era).",
			DefaultSize: size(44, 10), MinSize: size(12, 3), Import: "github.com/treilik/bubbleboxer", Status: definition.StatusPlaceholder,
			Props: []definition.PropSpec{choiceProp("orientation", "Orientation", "horizontal", "horizontal", "vertical"), intProp("children", "Children", 2, 2, 5), accent},
		},
		{
			ID: "community.datepicker", Name: "Date picker", Family: definition.FamilyCommunity,
			Description: "CameronJHall/bubble-datepicker (v2 fork of EthanEFung's): a month calendar.",
			DefaultSize: size(26, 9), MinSize: size(22, 8), Import: "github.com/CameronJHall/bubble-datepicker/v2", Status: definition.StatusPlaceholder,
			Props: []definition.PropSpec{textProp("month", "Month label", "October 2026"), intProp("day", "Selected day", 9, 1, 31), accent},
		},
		{
			ID: "community.overlay", Name: "Overlay / modal", Family: definition.FamilyCommunity,
			Description: "rmhubbert/bubbletea-overlay: a modal window (Bubble Tea v1; v2 users composite with Lip Gloss).",
			DefaultSize: size(36, 7), MinSize: size(12, 3), Import: "github.com/rmhubbert/bubbletea-overlay", Status: definition.StatusPlaceholder,
			Props: []definition.PropSpec{textProp("title", "Title", "Are you sure?"), textProp("body", "Body", "This cannot be undone."), accent},
		},
		{
			ID: "community.statusbar", Name: "Status bar", Family: definition.FamilyCommunity,
			Description: "knipferrc/teacup statusbar: four colored segments (Bubble Tea v1-era).",
			DefaultSize: size(60, 1), MinSize: size(12, 1), Import: "github.com/knipferrc/teacup/statusbar", Status: definition.StatusPlaceholder,
			Props: []definition.PropSpec{
				textProp("left", "Left", "main.go"), textProp("middle", "Middle", "UTF-8"),
				textProp("right", "Right", "Ln 12, Col 4"), textProp("end", "End", "100%"), accent,
			},
		},
		{
			ID: "community.filetree", Name: "File tree", Family: definition.FamilyCommunity,
			Description: "knipferrc/teacup filetree (Bubble Tea v1-era; the Bubbles tree supersedes it).",
			DefaultSize: size(28, 10), MinSize: size(10, 3), Import: "github.com/knipferrc/teacup/filetree", Status: definition.StatusPlaceholder,
			Props: []definition.PropSpec{
				textProp("entries", "Entries (comma separated, / = dir)", "apps/,libs/,docs/,go.work,README.md"),
				intProp("selected", "Selected", 1, 0, 99), accent,
			},
		},
		{
			ID: "community.frame", Name: "Titled frame", Family: definition.FamilyCommunity,
			Description: "clambin/bubbles frame: a border with the title set into the top edge.",
			DefaultSize: size(36, 5), MinSize: size(6, 3), Import: "codeberg.org/clambin/bubbles/frame", Status: definition.StatusPlaceholder,
			Props: []definition.PropSpec{
				textProp("title", "Title", "Tea menu"),
				choiceProp("position", "Title position", "left", "left", "center", "right"),
				choiceProp("border", "Border", "normal", borders...),
				textProp("content", "Content", "Earl Grey, Sencha, Rooibos"),
				accent,
			},
		},
		{
			ID: "community.dialog", Name: "Dialog", Family: definition.FamilyCommunity,
			Description: "clambin/bubbles dialog: centred text and a row of buttons in a rounded box.",
			DefaultSize: size(40, 7), MinSize: size(14, 5), Import: "codeberg.org/clambin/bubbles/dialog", Status: definition.StatusPlaceholder,
			Props: []definition.PropSpec{
				textProp("text", "Text", "Pour the tea?"),
				textProp("buttons", "Buttons (comma separated)", "OK,Cancel"),
				intProp("active", "Active button", 0, 0, 5),
				colorProp("color", "Border", "212"),
			},
		},
		{
			ID: "community.statusmessage", Name: "Status message", Family: definition.FamilyCommunity,
			Description: "clambin/bubbles statusbar: one message line, styled by level, with an optional spinner.",
			DefaultSize: size(40, 1), MinSize: size(8, 1), Import: "codeberg.org/clambin/bubbles/statusbar", Status: definition.StatusPlaceholder,
			Props: []definition.PropSpec{
				textProp("text", "Message", "Steeping the tea"),
				choiceProp("level", "Level", "info", "info", "warning", "error"),
				boolProp("spinner", "Show spinner", true),
			},
		},
		{
			ID: "community.toast", Name: "Toast", Family: definition.FamilyCommunity,
			Description: "DaltonSW/BubbleUp v2: a bordered notification coloured by kind, with a symbol prefix.",
			DefaultSize: size(34, 3), MinSize: size(12, 3), Import: "go.dalton.dog/bubbleup/v2", Status: definition.StatusPlaceholder,
			Props: []definition.PropSpec{
				choiceProp("kind", "Kind", "info", "info", "warn", "error", "debug"),
				textProp("message", "Message", "Tea is ready"),
				choiceProp("symbols", "Symbols", "unicode", "unicode", "ascii"),
			},
		},
		{
			ID: "community.bigtext", Name: "Big text", Family: definition.FamilyCommunity,
			Description: "common-nighthawk/go-figure: banner text in a FIGlet font.",
			DefaultSize: size(40, 6), MinSize: size(8, 3), Import: "github.com/common-nighthawk/go-figure", Status: definition.StatusPlaceholder,
			Props: []definition.PropSpec{
				textProp("text", "Text", "Cuppa"),
				choiceProp("font", "Font", "standard", "standard", "big", "small", "mini", "slant", "shadow", "doom", "banner", "block", "lean"),
				accent,
			},
		},
		{
			ID: "community.qrcode", Name: "QR code", Family: definition.FamilyCommunity,
			Description: "skip2/go-qrcode: a QR code drawn with half blocks.",
			DefaultSize: size(29, 15), MinSize: size(10, 5), Import: "github.com/skip2/go-qrcode", Status: definition.StatusPlaceholder,
			Props: []definition.PropSpec{
				textProp("content", "Content", "cuppa"),
				choiceProp("level", "Error correction", "medium", "low", "medium", "high", "highest"),
			},
		},
		{
			ID: "community.image", Name: "Image", Family: definition.FamilyCommunity,
			Description: "blacktop/go-termimg: an image shown with Kitty, Sixel, iTerm2 or half blocks.",
			DefaultSize: size(32, 10), MinSize: size(6, 3), Import: "github.com/blacktop/go-termimg", Status: definition.StatusPlaceholder,
			Props: []definition.PropSpec{
				textProp("path", "File", "photo.png"),
				choiceProp("fit", "Fit", "fit", "fit", "fill", "stretch"),
				accent,
			},
		},
	}
}
