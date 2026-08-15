// Package tools holds the toolbox: every tool is a small self contained unit
// registered here, exposed over HTTP and rendered as a card in the web UI.
package tools

// Field is one labelled result line shown under a tool.
type Field struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

// Tool is a single converter/helper of the toolbox.
type Tool struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Placeholder string `json:"placeholder"`
	// Run turns the raw user input into result fields.
	Run func(input string) ([]Field, error) `json:"-"`
}

var registry []*Tool

// Register adds a tool to the toolbox. Order of registration is the display order.
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
