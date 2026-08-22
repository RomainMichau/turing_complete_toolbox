package tools

import (
	"fmt"
	"strconv"
	"strings"
)

// The decoder reads a word back. Bits it cannot know are welcome: any letter
// stands for a variable, so an encoding line lifted straight out of the game's
// assembler definition — 01110110 0000vvvv aaaaaaaa aaaaaaaa — decodes into the
// fields it does pin down, and names the letters for the rest.

func init() {
	Register(&Tool{
		ID:          "symphony-decode",
		Name:        "Instruction Decoder",
		Family:      FamilySymphony,
		Description: "Read a word back into its fields. Letters stand for variables, so a pattern like 0000vvvv aaaaaaaa works as well as plain bits.",
		Inputs: []Input{
			{
				ID:          "word",
				Placeholder: "01110110 0000vvvv aaaaaaaa aaaaaaaa",
				Format:      "bits",
			},
			{
				ID:    "read",
				Kind:  "choice",
				Label: "Read as",
				Value: readBits,
				Options: []Option{
					{ID: readBits, Label: "Bits"},
					{ID: readNumber, Label: "Number"},
				},
			},
		},
		Run: decodeSymphony,
	})
}

const (
	readBits   = "bits"
	readNumber = "number"

	wordBits = 32
)

// The fields of a word, as slices of the 32 character pattern.
const (
	modeAt = 1
	immAt  = 3
	opAt   = 4
	destAt = 8
	argAAt = 12
	argBAt = 20 // 16 in IMM mode, where argument B takes the last two bytes
)

func decodeSymphony(in map[string]string) ([]Field, error) {
	text := cleanBits(in["word"])
	if text == "" {
		return nil, nil
	}

	pattern, err := wordPattern(text, in["read"])
	if err != nil {
		return nil, err
	}

	mode, modeKnown := patternValue(pattern[modeAt : modeAt+2])
	imm, immKnown := patternValue(pattern[immAt : immAt+1])
	op, opKnown := patternValue(pattern[opAt : opAt+4])

	// Argument B is four bits, unless the immediate flag hands it the last two
	// bytes. With the flag itself unknown, the narrow reading is the one shown.
	argBFrom, argBTo := argBAt, argBAt+4
	if immKnown && imm == 1 {
		argBFrom, argBTo = 16, wordBits
	}

	fields := []Field{
		{Label: "Bits", Value: pattern, Format: "bits"},
		{Label: "Mode", Value: describeMode(pattern[modeAt:modeAt+2], mode, modeKnown)},
		{Label: "Immediate", Value: describeIMM(pattern[immAt:immAt+1], imm, immKnown)},
		{Label: "Opcode", Value: describeOpcode(pattern[opAt:opAt+4], mode, op, modeKnown, opKnown)},
		{Label: "Dest", Value: describeRegister(pattern[destAt : destAt+4])},
		{Label: "Arg A", Value: describeRegister(pattern[argAAt : argAAt+4])},
		{Label: "Arg B", Value: describeArgB(pattern[argBFrom:argBTo], immKnown && imm == 1)},
	}

	if modeKnown && immKnown && opKnown {
		fields = append(fields, Field{
			Label: "Instruction",
			Value: describeSymphony(mode, op,
				operand(pattern[destAt:destAt+4]),
				operand(pattern[argAAt:argAAt+4]),
				operand(pattern[argBFrom:argBTo]),
				imm == 1),
		})
	} else {
		fields = append(fields, Field{
			Label: "Instruction",
			Value: "needs the mode, the immediate flag and the opcode to be known",
		})
	}
	return fields, nil
}

// wordPattern turns the input into exactly 32 characters of bits and variables.
func wordPattern(text, read string) (string, error) {
	if read == readNumber {
		value, err := strconv.ParseUint(strings.ToLower(text), 0, 64)
		if err != nil {
			return "", fmt.Errorf("%q is not a number — try 705823488 or 0x2A120300", text)
		}
		if value > 1<<wordBits-1 {
			return "", fmt.Errorf("%s does not fit in a 32 bit word", text)
		}
		return fmt.Sprintf("%032b", value), nil
	}

	for i, r := range text {
		bit := r == '0' || r == '1'
		letter := (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
		if !bit && !letter {
			return "", fmt.Errorf("%q at position %d is neither a bit nor a variable", string(r), i+1)
		}
	}
	if len(text) > wordBits {
		return "", fmt.Errorf("%d bits given, a word is %d", len(text), wordBits)
	}
	// A short pattern is read as the low bits of the word, like a number is.
	return strings.Repeat("0", wordBits-len(text)) + text, nil
}

// patternValue reads a run of the pattern, and says whether it was all bits.
func patternValue(slice string) (int, bool) {
	value, err := strconv.ParseInt(slice, 2, 32)
	if err != nil {
		return 0, false
	}
	return int(value), true
}

// operand is how a field reads inside a sentence: its value when known, and
// otherwise the letter standing for it.
func operand(slice string) string {
	if value, ok := patternValue(slice); ok {
		return strconv.Itoa(value)
	}
	return variableName(slice)
}

// variableName is the letter a field is filled with, or the raw run when it is
// a mix of bits and letters.
func variableName(slice string) string {
	if strings.Trim(slice, slice[:1]) == "" {
		return slice[:1] // all the same character, like "aaaa"
	}
	return slice
}

func describeMode(slice string, mode int, known bool) string {
	if !known {
		return fmt.Sprintf("variable %s", variableName(slice))
	}
	return fmt.Sprintf("%d · %s", mode, symModeNames[mode])
}

func describeIMM(slice string, imm int, known bool) string {
	switch {
	case !known:
		return fmt.Sprintf("variable %s — argument B could be either shape", variableName(slice))
	case imm == 1:
		return "1 — argument B is a 16 bit literal over the last two bytes"
	default:
		return "0 — a plain word"
	}
}

func describeOpcode(slice string, mode, op int, modeKnown, opKnown bool) string {
	if !opKnown {
		return fmt.Sprintf("variable %s", variableName(slice))
	}
	if !modeKnown {
		return fmt.Sprintf("%d — the mode decides what it means", op)
	}
	return fmt.Sprintf("%d · %s", op, opcodeName(mode, op))
}

func describeRegister(slice string) string {
	if value, ok := patternValue(slice); ok {
		return fmt.Sprintf("reg %d", value)
	}
	return fmt.Sprintf("variable %s", variableName(slice))
}

func describeArgB(slice string, imm bool) string {
	value, ok := patternValue(slice)
	if !ok {
		return fmt.Sprintf("variable %s (%d bits)", variableName(slice), len(slice))
	}
	if imm {
		return fmt.Sprintf("literal %d (0x%04X)", value, value)
	}
	return fmt.Sprintf("reg %d", value)
}

// opcodeName is what an opcode is called in its mode.
func opcodeName(mode, op int) string {
	switch mode {
	case modeIO:
		if op < len(symIOOps) {
			return opLabel(symIOOps[op])
		}
	case modeALU:
		if op < len(symALUOps) {
			return symALUOps[op]
		}
	case modeJump:
		return opLabel(symJumpConds[op&0xF])
	case modeRAM:
		if op < len(symRAMOps) {
			return symRAMOps[op]
		}
	}
	return "not documented"
}

func opLabel(op opDesc) string {
	if op.Name != "" {
		return op.Name
	}
	return op.Desc
}
