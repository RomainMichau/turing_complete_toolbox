package tools

import "fmt"

// The reference cards for the Symphony instruction set: one for the shape of a
// word, then one per mode. They sit first in the family — see Register on why
// the file name matters — and are meant to grow as the modes get documented.

func init() {
	Register(&Tool{
		ID:          "symphony-doc",
		Name:        "Instruction Doc",
		Family:      FamilySymphony,
		Description: "The word layout and every opcode, one fold per mode.",
		Doc: []DocSection{
			docGroup("Word", "How a word is laid out, what each field means, and the four modes.", wordDoc()),
			docGroup("IO", "Mode 0: input, output, keyboard, screen, time and counter.", ioDoc()),
			docGroup("ALU", "Mode 1: arithmetic, logic and the compare that feeds the jumps.", aluDoc()),
			docGroup("JUMP", "Mode 2: the opcode is a condition tested against a flags register.", jumpDoc()),
			docGroup("RAM", "Mode 3: moving values between the registers, program memory and persistent storage.", ramDoc()),
		},
	})
}

// docGroup wraps sections in a fold of their own inside the card.
func docGroup(title, subtitle string, sections []DocSection) DocSection {
	return DocSection{Kind: "group", Title: title, Text: subtitle, Sections: sections}
}

func wordDoc() []DocSection {
	return []DocSection{
		{
			Kind:   "layout",
			Title:  "Layout",
			Text:   symLayout,
			Colors: symDocColors,
		},
		{
			Kind:  "legend",
			Title: "Fields",
			Rows: []DocRow{
				{Key: "M", Value: "Mode", Color: colorMode},
				{Key: "I", Value: "Immediate flag — bit 4", Color: colorIMM},
				{Key: "O", Value: "OpCode", Color: colorOp},
				{Key: "D", Value: "Destination register", Color: colorDest},
				{Key: "A", Value: "Argument A — a register address", Color: colorArgA},
				{Key: "B", Value: "Argument B — a register address, a program memory address in RAM mode, or the value itself in IMM mode", Color: colorArgB},
				{Key: "X", Value: "Unused", Color: colorUnused},
			},
		},
		{
			Kind: "note",
			Text: "Three stores: the registers, which the destination and the arguments address; program memory, which the RAM loads and stores reach; and persistent storage, reached only by pload and pstore.",
		},
		{
			Kind:   "layout",
			Title:  "In IMM mode",
			Text:   symLayoutIMM,
			Colors: symDocColorsIMM,
		},
		{
			Kind: "note",
			Text: "Everything in the mode folds holds as written unless bit 4 is set. In IMM mode argument B is the value itself instead of an address, and it grows to 16 bits over the last two bytes.",
		},
		{
			Kind:  "table",
			Title: "Modes",
			Rows:  symModeRows(),
		},
	}
}

func ioDoc() []DocSection {
	return []DocSection{
		{
			Kind:    "table",
			Title:   "Opcodes",
			Rows:    symIORows(),
			Compact: true,
		},
		{
			Kind:   "layout",
			Title:  "Keyboard result",
			Text:   keyboardLayout,
			Colors: symKeyboardColors,
		},
		{
			Kind: "legend",
			Rows: []DocRow{
				{Key: "D", Value: "key down", Color: colorIMM},
				{Key: "V", Value: "key value", Color: colorArgA},
			},
		},
		{
			Kind: "note",
			Text: "The keyboard has two output pins, key down and key value. They are merged into the single word the keyboard opcode reads into the destination register.",
		},
		{
			Kind: "note",
			Text: "The time component gives the nanoseconds since 1 January 1970 as a 64 bit number, and this architecture only handles 32, so reading it takes two instructions. Only time_0 actually reads the component: it hands back the low half and stores the high half for time_1 to pick up afterwards. Halves taken from two different cycles would belong to two different times, which is why time_1 reads the stored half instead — and reads 0 until time_0 has run.",
		},
	}
}

func aluDoc() []DocSection {
	return []DocSection{
		{
			Kind:    "table",
			Title:   "Opcodes",
			Rows:    symALURows(),
			Compact: true,
		},
		{
			Kind: "note",
			Text: "CMP writes flags into the destination register — " + cmpFlags + ". There is no greater flag: greater is reached by inverting a jump condition.",
		},
	}
}

func jumpDoc() []DocSection {
	return []DocSection{
		{
			Kind:    "table",
			Title:   "Opcodes",
			Rows:    symJumpRows(),
			Compact: true,
		},
		{
			Kind: "note",
			Text: "The opcode is the condition. Its low three bits are matched pairwise against the flags register — the register at argument A, usually written by a CMP — and the jump fires if any bit matches. Bit 3 inverts that answer, which is why opcode 8 jumps unconditionally. The target is argument B.",
		},
	}
}

func ramDoc() []DocSection {
	return []DocSection{
		{
			Kind:    "table",
			Title:   "Opcodes",
			Rows:    symRAMRows(),
			Compact: true,
		},
		{
			Kind: "note",
			Text: "load reads the value from program memory at address argument B and stores it in the destination register. store writes the value in the argument A register to program memory at address argument B. The numerical suffix is how many bits are read or written, starting at that address.",
		},
		{
			Kind: "note",
			Text: "pload and pstore reach persistent storage — the 3D memory — instead of program memory, and always move 32 bits.",
		},
	}
}

// The two shapes a word can take. Bit 4 is the immediate flag: when it is set
// argument B swallows the last two bytes and holds a literal value.
const (
	symLayout    = "XMMIOOOO DDDDAAAA XXXXBBBB XXXXXXXX"
	symLayoutIMM = "XMM1OOOO DDDDAAAA BBBBBBBB BBBBBBBB"
)

// symDocColors paints each letter of the layout with its field colour.
var symDocColors = map[string]string{
	"M": colorMode,
	"I": colorIMM,
	"O": colorOp,
	"D": colorDest,
	"A": colorArgA,
	"B": colorArgB,
	"X": colorUnused,
}

// symKeyboardColors paints the merged keyboard word.
var symKeyboardColors = map[string]string{
	"D": colorIMM,
	"V": colorArgA,
	"X": colorUnused,
}

// symDocColorsIMM is the same, with the flag spelled as the literal 1 it is.
var symDocColorsIMM = func() map[string]string {
	colors := map[string]string{"1": colorIMM}
	for k, v := range symDocColors {
		colors[k] = v
	}
	return colors
}()

func symModeRows() []DocRow {
	rows := make([]DocRow, 0, len(symModeNames))
	for i, name := range symModeNames {
		rows = append(rows, DocRow{Key: fmt.Sprint(i), Value: name, Color: colorMode})
	}
	return rows
}

func symALURows() []DocRow {
	return opcodeRows(symALUOps[:])
}

func symRAMRows() []DocRow {
	return opcodeRows(symRAMOps[:])
}

func symJumpRows() []DocRow {
	return opcodeRows(spellOut(symJumpConds[:]))
}

func symIORows() []DocRow {
	return opcodeRows(spellOut(symIOOps[:]))
}

// spellOut turns opcodes into the "mnemonic — what it does" lines of the doc.
func spellOut(ops []opDesc) []string {
	lines := make([]string, 0, len(ops))
	for _, op := range ops {
		if op.Name == "" {
			lines = append(lines, op.Desc)
			continue
		}
		lines = append(lines, op.Name+" — "+op.Desc)
	}
	return lines
}

func opcodeRows(names []string) []DocRow {
	rows := make([]DocRow, 0, len(names))
	for i, name := range names {
		rows = append(rows, DocRow{Key: fmt.Sprint(i), Value: name, Color: colorOp})
	}
	return rows
}
