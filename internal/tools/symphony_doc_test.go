package tools

import (
	"strconv"
	"strings"
	"testing"
)

// The doc comes first in the family, one fold per mode, and the encoder closes
// the family.
func TestSymphonyFamilyOrder(t *testing.T) {
	var got []string
	for _, tool := range All() {
		if tool.Family == FamilySymphony {
			got = append(got, tool.ID)
		}
	}
	want := []string{"symphony-doc", "symphony-instruction", "symphony-decode"}
	if len(got) != len(want) {
		t.Fatalf("Symphony family holds %v, want %v", got, want)
	}
	for i, id := range want {
		if got[i] != id {
			t.Fatalf("Symphony family holds %v, want %v", got, want)
		}
	}

	doc := Get("symphony-doc")
	if doc.Run != nil {
		t.Error("the doc is a reference card, it should have nothing to run")
	}
	wantGroups := []string{"Word", "IO", "ALU", "JUMP", "RAM"}
	if len(doc.Doc) != len(wantGroups) {
		t.Fatalf("the doc holds %d sections, want %d groups", len(doc.Doc), len(wantGroups))
	}
	for i, title := range wantGroups {
		group := doc.Doc[i]
		if group.Kind != "group" || group.Title != title {
			t.Errorf("section %d is %q/%q, want a group titled %q", i, group.Kind, group.Title, title)
		}
		if len(group.Sections) == 0 {
			t.Errorf("group %q is empty", title)
		}
	}
}

// docSection finds a section by title inside a named group of the doc.
func docSection(t *testing.T, group, title string) *DocSection {
	t.Helper()
	for _, tool := range All() {
		for _, g := range tool.Doc {
			if g.Title != group {
				continue
			}
			for i, s := range g.Sections {
				if s.Title == title {
					return &g.Sections[i]
				}
			}
		}
	}
	t.Fatalf("no %q section in the %q group", title, group)
	return nil
}

func TestSymphonyLayouts(t *testing.T) {
	for _, layout := range []string{symLayout, symLayoutIMM} {
		bits := strings.ReplaceAll(layout, " ", "")
		if len(bits) != 32 {
			t.Errorf("%q describes %d bits, a word is 32", layout, len(bits))
		}
		// Bit 4 of the first byte is the immediate flag in both shapes.
		if flag := bits[3]; flag != 'I' && flag != '1' {
			t.Errorf("%q has %q at bit 4, want the immediate flag", layout, flag)
		}
	}
	// The mode and opcode fields must sit at the same place in both shapes.
	for i, want := range symLayout {
		if want == 'M' || want == 'O' || want == 'D' || want == 'A' {
			if got := rune(symLayoutIMM[i]); got != want {
				t.Errorf("position %d is %q in the plain layout, %q in IMM", i, want, got)
			}
		}
	}
}

// Every opcode the doc card lists must decode to something real: a doc entry
// with "not documented yet" underneath it would be a contradiction.
func TestEveryDocumentedOpcodeDecodes(t *testing.T) {
	modes := map[int]int{
		modeIO:   len(symIOOps),
		modeALU:  len(symALUOps),
		modeJump: len(symJumpConds),
		modeRAM:  len(symRAMOps),
	}
	for mode, count := range modes {
		for op := range count {
			got := describeSymphony(mode, op, "1", "2", "3", false)
			if strings.Contains(got, "not documented") {
				t.Errorf("%s opcode %d is listed in the doc but decodes as %q", symModeNames[mode], op, got)
			}
		}
	}
}

func TestSymphonyDocMatchesTheEncoder(t *testing.T) {
	cases := []struct {
		group, title string
		names        []string
	}{
		{"Word", "Modes", symModeNames[:]},
		{"ALU", "Opcodes", symALUOps[:]},
		{"RAM", "Opcodes", symRAMOps[:]},
	}

	for _, c := range cases {
		t.Run(c.group+"/"+c.title, func(t *testing.T) {
			rows := docSection(t, c.group, c.title).Rows
			if len(rows) != len(c.names) {
				t.Fatalf("doc lists %d rows, the encoder knows %d", len(rows), len(c.names))
			}
			for i, row := range rows {
				if row.Value != c.names[i] {
					t.Errorf("%d is %q in the doc, %q in the encoder", i, row.Value, c.names[i])
				}
				if row.Key != strconv.Itoa(i) {
					t.Errorf("row %d is keyed %q", i, row.Key)
				}
			}
		})
	}
}
