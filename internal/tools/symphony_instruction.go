package tools

import (
	"fmt"
	"strconv"
	"strings"
)

// A Symphony instruction is a 32 bit word laid out as
//
//	XMMIOOOO DDDDAAAA XXXXBBBB XXXXXXXX
//
// M = mode, I = immediate flag, O = opcode, D = destination register,
// A = argument A, B = argument B, X = unused. With the immediate flag set the
// word is read as
//
//	XMM1OOOO DDDDAAAA BBBBBBBB BBBBBBBB
//
// where argument B is a 16 bit literal instead of an address.
const (
	symModeShift = 29
	symIMMShift  = 28
	symOpShift   = 24
	symDestShift = 20
	symArgAShift = 16
	symArgBShift = 8

	argBWidth    = 4
	argBWidthIMM = 16
)

// Colours the instruction fields wear, in the encoder and in the doc alike.
const (
	colorMode   = "orange"
	colorOp     = "red"
	colorDest   = "yellow"
	colorArgA   = "green"
	colorArgB   = "blue"
	colorIMM    = "purple"
	colorUnused = "muted"
)

// Modes, in the order of their encoding.
const (
	modeIO = iota
	modeALU
	modeJump
	modeRAM
)

// The keyboard has two output pins, merged into the single value the keyboard
// opcode reads: D is key down, V is the key value.
const keyboardLayout = "XXXXXXXD VVVVVVVV"

// IO opcodes, in the order of their encoding.
var symIOOps = [...]opDesc{
	0: {"", "nothing"},
	1: {"", "send input to dest"},
	2: {"", "send argument B to out"},
	3: {"keyboard", "read the keyboard into dest"},
	4: {"screen", "set the screen settings from arg A and arg B"},
	5: {"time_0", "read the low half of time into dest, and store the high half for time_1"},
	6: {"time_1", "read the high half time_0 stored — 0 until time_0 has run"},
	7: {"counter", "read the counter into dest"},
}

var symModeNames = [...]string{modeIO: "IO", modeALU: "ALU", modeJump: "JUMP", modeRAM: "RAM"}

// ALU opcodes, in the order of their encoding.
var symALUOps = [...]string{"NAND", "OR", "AND", "NOR", "ADD", "SUB", "XOR", "LSL", "LSR", "ASR", "CMP"}

// aluCMP writes flags rather than a value, so it reads differently.
const aluCMP = 10

// RAM opcodes, in the order of their encoding. The first half loads from
// memory, the second half stores to it, and the suffix is how many bits move.
var symRAMOps = [...]string{
	"load_8", "load_16", "load_32", "pload",
	"store_8", "store_16", "store_32", "pstore",
}

const (
	ramStore  = 4 // opcodes from here on store instead of load
	ramPLoad  = 3
	ramPStore = 7
)

// cmpFlags is how the result of a CMP is laid out. Note that there is no
// "greater" flag: greater is reached by inverting a jump condition.
const cmpFlags = "bit 0: equal, bit 1: below (unsigned), bit 2: less (signed)"

// A JUMP opcode is a condition. Its low three bits are matched against the
// flags register, the jump firing when any bit matches, and bit 3 inverts the
// answer — so opcode 8, the inverse of "no flag at all", is the plain jump.
const (
	jumpInvert = 8
	jumpNever  = 0
)

// opDesc is an opcode as the doc card spells it out: a mnemonic when it has
// one, and what it does.
type opDesc struct {
	Name string
	Desc string
}

var symJumpConds = [16]opDesc{
	0:  {"never", "never"},
	1:  {"je", "equal"},
	2:  {"jb", "below (unsigned)"},
	3:  {"jbe", "below or equal (unsigned)"},
	4:  {"jl", "less (signed)"},
	5:  {"jle", "less or equal (signed)"},
	6:  {"", "below or less"},
	7:  {"", "equal, below or less"},
	8:  {"jmp", "always"},
	9:  {"jne", "not equal"},
	10: {"jae", "above or equal (unsigned)"},
	11: {"ja", "above (unsigned)"},
	12: {"jge", "greater or equal (signed)"},
	13: {"jg", "greater (signed)"},
	14: {"", "neither below nor less"},
	15: {"", "no flag set — above and greater"},
}

func init() {
	Register(&Tool{
		ID:          "symphony-instruction",
		Name:        "Instruction Encoder",
		Family:      FamilySymphony,
		Description: "Build an instruction bit by bit and read the integer to feed the machine. The doc above spells out the layout and every opcode.",
		Inputs: symInputs(
			Input{Width: 4, Value: "0000"},
			Input{ID: "argB", Label: "arg B", Width: argBWidth, Color: colorArgB},
			Input{Width: 8, Value: "00000000"},
		),
		// Setting the immediate flag hands the last two bytes to argument B.
		Variants: []Variant{{
			When: Condition{Input: "imm", Equals: "1"},
			Inputs: symInputs(
				Input{ID: "argB", Label: "arg B — value", Width: argBWidthIMM, Color: colorArgB},
			),
		}},
		Run: encodeSymphony,
	})
}

// symInputs builds the segment row: the head is the same in both layouts, only
// the last two bytes differ.
func symInputs(tail ...Input) []Input {
	head := []Input{
		{Width: 1, Value: "0"},
		{ID: "mode", Label: "mode", Width: 2, Color: colorMode},
		{ID: "imm", Label: "imm", Width: 1, Color: colorIMM},
		{ID: "op", Label: "opcode", Width: 4, Color: colorOp},
		{ID: "dest", Label: "dest", Width: 4, Color: colorDest},
		{ID: "argA", Label: "arg A", Width: 4, Color: colorArgA},
	}
	return append(head, tail...)
}

// bitsField reads one named bit field, defaulting to 0 when it is missing.
func bitsField(in map[string]string, id string, width int) (int, error) {
	bits := cleanBits(in[id])
	if bits == "" {
		return 0, nil
	}
	if len(bits) > width {
		return 0, fmt.Errorf("%s: %d bits given, field is %d bits wide", id, len(bits), width)
	}
	v, err := strconv.ParseInt(bits, 2, 32)
	if err != nil {
		return 0, fmt.Errorf("%s: %q is not a binary value", id, in[id])
	}
	return int(v), nil
}

func encodeSymphony(in map[string]string) ([]Field, error) {
	imm, err := bitsField(in, "imm", 1)
	if err != nil {
		return nil, err
	}
	// The immediate flag decides how much room argument B gets, and where.
	bWidth, bShift := argBWidth, symArgBShift
	if imm == 1 {
		bWidth, bShift = argBWidthIMM, 0
	}

	widths := []struct {
		id    string
		width int
	}{{"mode", 2}, {"op", 4}, {"dest", 4}, {"argA", 4}, {"argB", bWidth}}

	values := map[string]int{}
	for _, f := range widths {
		v, err := bitsField(in, f.id, f.width)
		if err != nil {
			return nil, err
		}
		values[f.id] = v
	}
	mode, op := values["mode"], values["op"]
	dest, argA, argB := values["dest"], values["argA"], values["argB"]

	word := uint32(mode)<<symModeShift |
		uint32(imm)<<symIMMShift |
		uint32(op)<<symOpShift |
		uint32(dest)<<symDestShift |
		uint32(argA)<<symArgAShift |
		uint32(argB)<<uint(bShift)

	modeText := fmt.Sprintf("%d · %s", mode, symModeNames[mode])
	if imm == 1 {
		modeText += " · IMM"
	}

	return []Field{
		{Label: "Int", Value: strconv.FormatUint(uint64(word), 10)},
		{Label: "Hex", Value: fmt.Sprintf("0x%08X", word)},
		{Label: "Bits", Value: symBits(word)},
		{Label: "Mode", Value: modeText},
		{Label: "Instruction", Value: describeSymphony(mode, op, itoa(dest), itoa(argA), itoa(argB), imm == 1)},
	}, nil
}

// symBits renders the whole word as bits, one group per byte.
func symBits(word uint32) string {
	var b strings.Builder
	for i := 31; i >= 0; i-- {
		b.WriteByte('0' + byte(word>>uint(i)&1))
		if i%8 == 0 && i != 0 {
			b.WriteByte(' ')
		}
	}
	return b.String()
}

// itoa keeps the encoder's calls to describeSymphony readable.
func itoa(v int) string { return strconv.Itoa(v) }

// describeSymphony spells the instruction out. Only the JUMP conditions and the
// IO, ALU and RAM opcodes documented so far are named; anything else says it is
// not documented. The operands are text, not numbers, so a decoder can hand
// over the letter of a field it does not know.
func describeSymphony(mode, op int, dest, argA, argB string, imm bool) string {
	// With the immediate flag set, argument B carries a value, not an address.
	b := fmt.Sprintf("arg B (%s)", argB)
	if imm {
		b = fmt.Sprintf("literal %s", argB)
	}

	switch mode {
	case modeIO:
		switch op {
		case 0:
			return "nothing"
		case 1:
			return fmt.Sprintf("input → reg %s", dest)
		case 2:
			return fmt.Sprintf("%s → out", b)
		case 3:
			return fmt.Sprintf("keyboard → reg %s (%s)", dest, keyboardLayout)
		case 4:
			return fmt.Sprintf("screen setting arg A (%s) ← %s", argA, b)
		case 5:
			return fmt.Sprintf("reg %s ← low half of time, high half stored for time_1", dest)
		case 6:
			return fmt.Sprintf("reg %s ← high half of time, as stored by time_0", dest)
		case 7:
			return fmt.Sprintf("reg %s ← counter", dest)
		}
	case modeALU:
		if op == aluCMP {
			return fmt.Sprintf("reg %s ← compare arg A (%s) with %s — %s", dest, argA, b, cmpFlags)
		}
		if op < len(symALUOps) {
			return fmt.Sprintf("reg %s ← arg A (%s) %s %s", dest, argA, symALUOps[op], b)
		}
	case modeJump:
		cond := symJumpConds[op&0xF]
		switch op {
		case jumpNever:
			return "never jumps"
		case jumpInvert: // the plain, unconditional jump
			return fmt.Sprintf("jump to %s", b)
		}
		line := fmt.Sprintf("jump to %s if %s — flags in reg A (%s)", b, cond.Desc, argA)
		if cond.Name != "" {
			line = cond.Name + " · " + line
		}
		return line
	case modeRAM:
		if op < len(symRAMOps) {
			// The persistent pair always moves a full 32 bits, and reaches
			// the 3D memory rather than program memory.
			switch op {
			case ramPLoad:
				return fmt.Sprintf("reg %s ← 32 bits of persistent storage at %s", dest, b)
			case ramPStore:
				return fmt.Sprintf("32 bits of reg A (%s) → persistent storage at %s", argA, b)
			}
			width := 8 << (op % ramStore) // 8, 16 or 32 bits
			if op < ramStore {
				return fmt.Sprintf("reg %s ← %d bits of program memory at %s", dest, width, b)
			}
			return fmt.Sprintf("%d bits of reg A (%s) → program memory at %s", width, argA, b)
		}
	}
	return fmt.Sprintf("%s opcode %d — not documented yet", symModeNames[mode], op)
}
