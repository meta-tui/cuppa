package standard

import "github.com/meta-ui-labs/cuppa/libs/catalog/definition"

const lipglossImport = "charm.land/lipgloss/v2"

var borders = []string{"rounded", "normal", "thick", "double", "ascii", "hidden"}

func lipglossEntries() []definition.Definition {
	return []definition.Definition{
		{
			ID: "lipgloss.box", Name: "Box", Family: definition.FamilyLipgloss,
			Description: "A bordered block with an optional title.",
			DefaultSize: size(24, 6), MinSize: size(3, 3), Import: lipglossImport, Status: definition.StatusSupported,
			Props: []definition.PropSpec{
				textProp("title", "Title", ""),
				choiceProp("border", "Border", "rounded", borders...),
				colorProp("color", "Border color", "212"),
				colorProp("gradient", "Gradient to (blend the border)", ""),
				textProp("right", "Right label (in the top edge)", ""),
				colorProp("titleColor", "Title colour (empty: the border's)", ""),
				boolProp("titleBold", "Title in bold", true),
				colorProp("rightColor", "Right label colour", "240"),
				boolProp("focused", "Has focus (false: the border takes the blurred colour)", true),
				colorProp("blurColor", "Border colour without focus", "240"),
			},
		},
		{
			ID: "lipgloss.label", Name: "Label", Family: definition.FamilyLipgloss,
			Description: "A line of styled text.",
			DefaultSize: size(14, 1), MinSize: size(1, 1), Import: lipglossImport, Status: definition.StatusSupported,
			Props: []definition.PropSpec{
				textProp("text", "Text", "Hello, Cuppa"),
				colorProp("color", "Foreground", "255"),
				colorProp("background", "Background", ""),
				boolProp("bold", "Bold", false),
				choiceProp("align", "Align", "left", "left", "center", "right"),
			},
		},
		{
			ID: "lipgloss.swatch", Name: "Colour swatch", Family: definition.FamilyLipgloss,
			Description: "A colour shown as a block with its value and a label, like a row of a colour list. Bind the colour so a program can change it.",
			DefaultSize: size(30, 1), MinSize: size(4, 1), Import: lipglossImport, Status: definition.StatusSupported,
			Props: []definition.PropSpec{
				textProp("label", "Label", "Accent"),
				intProp("labelWidth", "Label width (0 fits the text)", 0, 0, 60),
				colorProp("color", "Colour", "#ff007f"),
				intProp("swatch", "Swatch width", 2, 1, 8),
				boolProp("showValue", "Show the value", true),
				colorProp("textColor", "Text colour", "255"),
			},
		},
		{
			ID: "lipgloss.colourpicker", Name: "Colour picker", Family: definition.FamilyLipgloss,
			Description: "Cuppa's colour dialog as a component: tabs for the 16 and 256 palettes and for RGB and HSL sliders, a preview and the value. From github.com/meta-ui-labs/bubble-colourpicker, which any Bubble Tea v2 program can use. Bind the value so the program owns it: the screen then takes a ColourPicker it updates with the keys and the mouse. The width is 46; the 256 tab needs 23 rows.",
			DefaultSize: size(46, 10), MinSize: size(46, 4), Import: "github.com/meta-ui-labs/bubble-colourpicker", Status: definition.StatusSupported,
			Props: []definition.PropSpec{
				colorProp("value", "Colour (empty for none)", "#ff007f"),
				textProp("tabs", "Tabs (16, 256, RGB, HSL)", "16,256,RGB,HSL"),
				choiceProp("tab", "Tab shown", "RGB", "16", "256", "RGB", "HSL"),
				intProp("slide", "Slider the keyboard is on (RGB, HSL)", 0, 0, 2),
				colorProp("color", "Accent", "212"),
			},
		},
		{
			ID: "lipgloss.rows", Name: "Rows", Family: definition.FamilyLipgloss,
			Description: "A repeated row: the design names the columns, the program gives the rows and the style of each (normal, selected, dim, accent). Bind rows to make a typed list in the screen's contract.",
			DefaultSize: size(44, 5), MinSize: size(8, 1), Import: lipglossImport, Status: definition.StatusSupported,
			Props: []definition.PropSpec{
				textProp("columns", "Columns (Name:width, add :colour for colours)", "Cursor:2,Name:14,Swatch:6:colour,Value"),
				textProp("rows", "Sample rows (; between rows, , between cells)", "▸,Accent,#ff007f,#ff007f;,Focus,#00f0ff,#00f0ff;,Success,#3ddc84,#3ddc84;,Error,#ff4d4d,#ff4d4d"),
				textProp("styles", "Row styles (normal, selected, dim, accent)", "selected,normal,normal,dim"),
				textProp("cellstyles", "Cell styles (given by the program)", ""),
				intProp("gap", "Spaces between columns", 0, 0, 8),
				colorProp("color", "Colour", "212"),
				boolProp("focused", "Has focus (false: a selected row is accent text, not inverted)", true),
			},
		},
		{
			ID: "lipgloss.progressline", Name: "Progress line", Family: definition.FamilyLipgloss,
			Description: "A progress bar followed by a text, such as 34.2% of 112.4MiB at 8.1MiB/s. Bind the value (a number with a fraction) and the suffix.",
			DefaultSize: size(70, 1), MinSize: size(4, 1), Import: lipglossImport, Status: definition.StatusSupported,
			Props: []definition.PropSpec{
				floatProp("value", "Done (0 to 100)", 34.2, 0, 100),
				intProp("barWidth", "Bar width (0: what the text leaves)", 32, 0, 400),
				textProp("full", "Full cell", "█"),
				textProp("empty", "Empty cell", "░"),
				colorProp("fullColor", "Full colour", "220"),
				colorProp("emptyColor", "Empty colour", "240"),
				textProp("suffix", "Text after the bar", "34.2% of 112.4MiB at 8.1MiB/s ETA 0:09"),
				colorProp("suffixColor", "Text colour", "255"),
				intProp("gap", "Spaces between", 3, 0, 16),
			},
		},
		{
			ID: "lipgloss.keybar", Name: "Key bar", Family: definition.FamilyLipgloss,
			Description: "A line of key hints: the key in bold, its label after it. Bind the hints and the program says which ones show; a click raises the component's event with the column.",
			DefaultSize: size(60, 1), MinSize: size(4, 1), Import: lipglossImport, Status: definition.StatusSupported,
			Props: []definition.PropSpec{
				textProp("hints", "Hints (key:label, comma separated)", "↑↓:move,enter:edit,s:save,esc:back"),
				colorProp("keyColor", "Key colour", "220"),
				colorProp("labelColor", "Label colour", "240"),
				intProp("gap", "Spaces between hints", 2, 0, 8),
			},
		},
		{
			ID: "cuppa.tabs", Name: "Tab pages", Family: definition.FamilyLipgloss,
			Description: "A bar of tabs whose pages hold components: place components on the shown page and they appear only while their tab is chosen. Pressing a tab shows its page, in the designer and in the program, with no code. Bind the active page to read or set it.",
			DefaultSize: size(40, 12), MinSize: size(12, 4), Import: lipglossImport, Status: definition.StatusSupported,
			Props: []definition.PropSpec{
				textProp("pages", "Pages (comma separated)", "Page 1,Page 2"),
				intProp("active", "Active page (from 0)", 0, 0, 40),
				colorProp("color", "Active tab colour", "212"),
				colorProp("mutedColor", "Other tabs colour", "245"),
				colorProp("ruleColor", "Rule colour", "240"),
			},
		},
		{
			ID: "cuppa.panels", Name: "Panels", Family: definition.FamilyLipgloss,
			Description: "A stack of collapsible sections, like the side bar of an image editor. Components placed on a section show only while it is open; pressing its header opens or closes it, in the designer and in the program. With Exclusive it is an accordion. Bind the open sections to read or set them.",
			DefaultSize: size(40, 16), MinSize: size(12, 3), Import: lipglossImport, Status: definition.StatusSupported,
			Props: []definition.PropSpec{
				textProp("pages", "Sections (comma separated)", "Details,Properties,Layers"),
				textProp("open", "Open sections (numbers from 0)", "0,1"),
				boolProp("exclusive", "Only one open (accordion)", false),
				colorProp("color", "Open header colour", "212"),
				colorProp("mutedColor", "Closed header colour", "245"),
				colorProp("ruleColor", "Rule colour", "240"),
			},
		},
		{
			ID: "lipgloss.slot", Name: "Model view", Family: definition.FamilyLipgloss,
			Description: "The place of a view the program draws itself: bind the view and give it what a Bubbles model of the program rendered (a text area with its real cursor, a viewport). The design places and cuts it; its text is shown here only as a sample.",
			DefaultSize: size(40, 6), MinSize: size(1, 1), Import: lipglossImport, Status: definition.StatusSupported,
			Props: []definition.PropSpec{
				textProp("view", "Sample view (| starts a line)", "The program's own view|is drawn here"),
			},
		},
		{
			ID: "lipgloss.list", Name: "List", Family: definition.FamilyLipgloss,
			Description: "An enumerated list (lipgloss/list).",
			DefaultSize: size(20, 5), MinSize: size(6, 1), Import: lipglossImport + "/list", Status: definition.StatusSupported,
			Props: []definition.PropSpec{
				textProp("items", "Items (comma separated)", "Tea,Milk,Sugar"),
				choiceProp("enumerator", "Enumerator", "bullet", "bullet", "arabic", "alphabet", "dash", "none"),
				colorProp("color", "Color", "255"),
			},
		},
		{
			ID: "lipgloss.tabs", Name: "Tabs", Family: definition.FamilyLipgloss,
			Description: "Tabs whose active tab opens into the window below (the Lip Gloss layout example).",
			DefaultSize: size(44, 9), MinSize: size(10, 4), Import: lipglossImport, Status: definition.StatusSupported,
			Props: []definition.PropSpec{
				textProp("tabs", "Tabs (comma separated)", "Lip Gloss,Blush,Eye Shadow,Mascara"),
				intProp("active", "Active tab (1 is the first)", 1, 1, 12),
				colorProp("color", "Border color", "99"),
			},
		},
		{
			ID: "lipgloss.table", Name: "Table", Family: definition.FamilyLipgloss,
			Description: "A static table (lipgloss/table).",
			DefaultSize: size(32, 7), MinSize: size(8, 3), Import: lipglossImport + "/table", Status: definition.StatusSupported,
			Props: []definition.PropSpec{
				textProp("headers", "Headers (comma separated)", "Name,Kind,Qty"),
				textProp("rows", "Rows (; and , separated)", "Earl Grey,Black,2;Sencha,Green,1"),
				choiceProp("border", "Border", "normal", borders...),
				colorProp("color", "Border color", "240"),
			},
		},
		{
			ID: "lipgloss.tree", Name: "Tree", Family: definition.FamilyLipgloss,
			Description: "A tree of items (lipgloss/tree).",
			DefaultSize: size(22, 6), MinSize: size(8, 2), Import: lipglossImport + "/tree", Status: definition.StatusSupported,
			Props: []definition.PropSpec{
				textProp("root", "Root", "Pantry"),
				textProp("items", "Items (comma separated)", "Tea,Milk,Biscuits"),
				colorProp("color", "Color", "255"),
			},
		},
		{
			ID: "lipgloss.joinh", Name: "Join horizontal", Family: definition.FamilyLipgloss,
			Description: "Lays out blocks side by side (JoinHorizontal).",
			DefaultSize: size(30, 5), MinSize: size(6, 1), Import: lipglossImport, Status: definition.StatusPlaceholder,
			Props: []definition.PropSpec{
				intProp("columns", "Columns", 3, 2, 8),
				choiceProp("align", "Align", "top", "top", "center", "bottom"),
			},
		},
		{
			ID: "lipgloss.joinv", Name: "Join vertical", Family: definition.FamilyLipgloss,
			Description: "Stacks blocks (JoinVertical).",
			DefaultSize: size(20, 7), MinSize: size(4, 2), Import: lipglossImport, Status: definition.StatusPlaceholder,
			Props: []definition.PropSpec{
				intProp("rows", "Rows", 3, 2, 8),
				choiceProp("align", "Align", "left", "left", "center", "right"),
			},
		},
		{
			ID: "lipgloss.place", Name: "Place", Family: definition.FamilyLipgloss,
			Description: "Positions content inside a larger area (Place).",
			DefaultSize: size(24, 7), MinSize: size(5, 3), Import: lipglossImport, Status: definition.StatusPlaceholder,
			Props: []definition.PropSpec{
				choiceProp("horizontal", "Horizontal", "center", "left", "center", "right"),
				choiceProp("vertical", "Vertical", "center", "top", "center", "bottom"),
				textProp("text", "Content", "centered"),
			},
		},
	}
}
