// Package tools holds the toolbox: every tool is a small self contained unit
// registered here, exposed over HTTP and rendered as a card in the web UI.
package tools

// Families group the tools in the UI, in first-registration order.
const (
	FamilyGeneral  = "General"
	FamilySymphony = "Symphony"
)

// Input describes one editable box of a tool.
//
// A tool either has a single free input (Width 0, any length), or a row of
// fixed-width segments making up a wider word. A segment with an empty ID is
// filler the user cannot edit — the unused bits of an instruction, say.
// Option is one choice of a "choice" input.
type Option struct {
	ID    string `json:"id"`
	Label string `json:"label"`
}

type Input struct {
	ID          string `json:"id,omitempty"`
	Label       string `json:"label,omitempty"`
	Placeholder string `json:"placeholder,omitempty"`
	// Kind is "choice" for a row of options to pick from, empty for a text
	// box.
	Kind    string   `json:"kind,omitempty"`
	Options []Option `json:"options,omitempty"`
	// Width in bits, 0 for a free length input.
	Width int `json:"width,omitempty"`
	// Colour name the UI paints the segment with: orange, red, yellow,
	// green, blue or muted. Anything else falls back to muted.
	Color string `json:"color,omitempty"`
	// Value is the initial content; the fixed content for filler segments.
	Value string `json:"value,omitempty"`
	// Format asks the UI for extra typing behaviour. "bits" colours the
	// input byte by byte as it is typed.
	Format string `json:"format,omitempty"`
}

// Condition is met when the named input holds exactly Equals.
type Condition struct {
	Input  string `json:"input"`
	Equals string `json:"equals"`
}

// Variant is an alternative input layout, used while its condition holds. It
// lets a word change shape as it is typed — the Symphony immediate flag giving
// argument B the last two bytes, say. Values of inputs present in both layouts
// survive the switch.
type Variant struct {
	When   Condition `json:"when"`
	Inputs []Input   `json:"inputs"`
}

// Field is one labelled result line shown under a tool.
type Field struct {
	Label string `json:"label"`
	Value string `json:"value"`
	// Format asks the UI to draw the value specially. "bits" paints it byte
	// by byte in alternating colours, like a bit input.
	Format string `json:"format,omitempty"`
}

// DocRow is one entry of a legend or of a table of values.
type DocRow struct {
	Key   string `json:"key"`
	Value string `json:"value"`
	Color string `json:"color,omitempty"`
}

// DocSection is a block of reference documentation. Kind picks how the UI
// draws it:
//
//	layout — Text drawn one character at a time, coloured through Colors
//	legend — Rows as "key — value", the key wearing its colour
//	table  — Rows as two aligned columns
//	note   — Text as a plain remark
//	group  — Sections behind a fold of their own, Text as its subtitle
type DocSection struct {
	Title    string            `json:"title,omitempty"`
	Kind     string            `json:"kind"`
	Text     string            `json:"text,omitempty"`
	Colors   map[string]string `json:"colors,omitempty"`
	Rows     []DocRow          `json:"rows,omitempty"`
	Sections []DocSection      `json:"sections,omitempty"`
	// Compact flows a long table over as many columns as fit, instead of
	// one row per line.
	Compact bool `json:"compact,omitempty"`
}

// Tool is a single converter/helper of the toolbox. A tool either takes input
// and computes something with Run, or is pure reference documentation.
type Tool struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Family      string       `json:"family"`
	Description string       `json:"description"`
	Inputs      []Input      `json:"inputs,omitempty"`
	Variants    []Variant    `json:"variants,omitempty"`
	Doc         []DocSection `json:"doc,omitempty"`
	// Run turns the raw user input, keyed by input ID, into result fields.
	// Nil for a documentation-only tool.
	Run func(inputs map[string]string) ([]Field, error) `json:"-"`
}

var registry []*Tool

// Register adds a tool to the toolbox. Order of registration is the display
// order, and files of a package are initialised in alphabetical order, so a
// tool's file name decides where it lands inside its family.
func Register(t *Tool) {
	registry = append(registry, t)
}

// All returns every registered tool.
func All() []*Tool {
	return registry
}

// Get returns the tool with the given id, or nil.
func Get(id string) *Tool {
	for _, t := range registry {
		if t.ID == id {
			return t
		}
	}
	return nil
}
