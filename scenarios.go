package mustache

// Scenario defines a contextual behavior in Mustache templates
// that affects whitespace handling and indentation behavior
type Scenario struct {
	ID          string
	Name        string
	Description string
	YesExample  string
	NoExample   string
}

// scenarios enumerate different scenarios for Mustache template processing
var scenarios = []Scenario{
	{
		ID:          "TI",
		Name:        "Tag Indentation",
		Description: "Opening tag has whitespace before it",
		YesExample:  "    {{<parent}}",
		NoExample:   "{{<parent}}",
	},
	{
		ID:          "II",
		Name:        "Inline Indentation",
		Description: "Content on same line as opening tag has additional whitespace",
		YesExample:  "{{$block}}    content",
		NoExample:   "{{$block}}content",
	},
	{
		ID:          "MI",
		Name:        "Multiline Indentation",
		Description: "Content on subsequent lines has consistent leading whitespace",
		YesExample:  "{{$block}}\n    content\n    more",
		NoExample:   "{{$block}}\ncontent\nmore",
	},
	{
		ID:          "MC",
		Name:        "Multiline Content",
		Description: "Content spans multiple lines",
		YesExample:  "{{$block}}content\nmore{{/block}}",
		NoExample:   "{{$block}}content{{/block}}",
	},
	{
		ID:          "TN",
		Name:        "Trailing Newline",
		Description: "Newline after closing tag",
		YesExample:  "{{/parent}}\n",
		NoExample:   "{{/parent}}",
	},
	{
		ID:          "IC",
		Name:        "Indented Closing",
		Description: "Closing tag has whitespace before it",
		YesExample:  "content\n    {{/block}}",
		NoExample:   "content\n{{/block}}",
	},
	{
		ID:          "PI",
		Name:        "Parent Indentation",
		Description: "Parent template has indentation at block expansion point",
		YesExample:  "  {{>partial}}",
		NoExample:   "{{>partial}}",
	},
	{
		ID:          "ST",
		Name:        "Standalone Tag",
		Description: "Tag alone on line with only whitespace before/after",
		YesExample:  "    {{tag}}\n",
		NoExample:   "content {{tag}} more",
	},
}
