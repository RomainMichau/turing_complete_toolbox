package tools

import (
	"fmt"
	"math/big"
	"strings"
	"unicode"
)

// The one number box: type a value in any base and read it back in all of
// them. There is nothing to pick on the way out — every representation is
// shown at once — so the only choice is how the input should be read, and even
// that guesses well on its own.

func init() {
	Register(&Tool{
		ID:          "number",
		Name:        "Number Converter",
		Family:      FamilyGeneral,
		Description: "Binary, decimal, hex and octal, all at once. Prefixes and separators are understood, and negatives come out in two's complement.",
		Inputs: []Input{
			{
				ID:          "value",
				Placeholder: "1100 1101, 205, 0xCD…",
			},
			{
				ID:    "base",
				Kind:  "choice",
				Label: "Read as",
				Value: baseAuto,
				Options: []Option{
					{ID: baseAuto, Label: "Auto"},
					{ID: baseBin, Label: "Binary"},
					{ID: baseDec, Label: "Decimal"},
					{ID: baseHex, Label: "Hex"},
					{ID: baseOct, Label: "Octal"},
				},
			},
		},
		Run: convertNumber,
	})
}

const (
	baseAuto = "auto"
	baseBin  = "bin"
	baseDec  = "dec"
	baseHex  = "hex"
	baseOct  = "oct"
)

type numberBase struct {
	radix  int
	name   string
	prefix string
}

var numberBases = map[string]numberBase{
	baseBin: {2, "binary", "0b"},
	baseDec: {10, "decimal", ""},
	baseHex: {16, "hex", "0x"},
	baseOct: {8, "octal", "0o"},
}

// cleanBits strips the separators a human naturally types inside a number.
func cleanBits(input string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) || r == '_' || r == '|' {
			return -1
		}
		return r
	}, input)
}

// describeWidth spells a bit count out in bytes, e.g. "12 bits (1 byte + 4 bits)".
func describeWidth(width int) string {
	bytes, rest := width/8, width%8
	unit := "bytes"
	if bytes == 1 {
		unit = "byte"
	}
	switch {
	case bytes == 0:
		return fmt.Sprintf("%d bits", width)
	case rest == 0:
		return fmt.Sprintf("%d bits (%d %s)", width, bytes, unit)
	default:
		return fmt.Sprintf("%d bits (%d %s + %d bits)", width, bytes, unit, rest)
	}
}

// readAs decides which base the digits are in and hands back the digits with
// any prefix removed.
func readAs(choice, text string) (numberBase, string, error) {
	if base, ok := numberBases[choice]; ok {
		return base, strings.TrimPrefix(strings.ToLower(text), base.prefix), nil
	}

	// Auto: a prefix settles it, then a string of nothing but 0 and 1 is far
	// more likely to be bits than a decimal number in this toolbox.
	lower := strings.ToLower(text)
	for _, id := range []string{baseHex, baseBin, baseOct} {
		base := numberBases[id]
		if digits, found := strings.CutPrefix(lower, base.prefix); found {
			return base, digits, nil
		}
	}
	if strings.Trim(text, "01") == "" {
		return numberBases[baseBin], text, nil
	}
	return numberBases[baseDec], text, nil
}

func convertNumber(in map[string]string) ([]Field, error) {
	text := cleanBits(in["value"])
	if text == "" {
		return nil, nil
	}

	negative := strings.HasPrefix(text, "-")
	base, digits, err := readAs(in["base"], strings.TrimPrefix(text, "-"))
	if err != nil {
		return nil, err
	}
	if digits == "" {
		return nil, nil // a lone prefix: the value is still being typed
	}

	value, ok := new(big.Int).SetString(digits, base.radix)
	if !ok {
		return nil, fmt.Errorf("%q is not a %s number", text, base.name)
	}

	// Binary input is taken at the width it was typed; anything else gets the
	// narrowest whole number of bytes that holds it.
	width := byteWidth(value, negative)
	if base.radix == 2 {
		width = len(digits)
	}

	// pattern is what the bits actually hold, so a negative shows up as its
	// two's complement.
	pattern := new(big.Int).Set(value)
	if negative {
		pattern.Sub(new(big.Int).Lsh(big.NewInt(1), uint(width)), value)
	}

	fields := []Field{
		{Label: "Read as", Value: fmt.Sprintf("%s, %s", base.name, describeWidth(width))},
		{Label: "Decimal", Value: signedText(value, negative)},
	}
	if signed, differs := signedReading(pattern, width, negative); differs {
		fields = append(fields, Field{
			Label: fmt.Sprintf("Signed (%d-bit two's complement)", width),
			Value: signed,
		})
	}
	fields = append(fields,
		Field{Label: "Hex", Value: fmt.Sprintf("0x%0*X", (width+3)/4, pattern)},
		Field{Label: "Octal", Value: "0o" + pattern.Text(8)},
		Field{Label: "Binary", Value: bitsOf(pattern, width), Format: "bits"},
	)
	if width < 32 {
		fields = append(fields, Field{
			Label:  "32 bits",
			Value:  bitsOf(pattern, 32),
			Format: "bits",
		})
	}
	return fields, nil
}

// byteWidth is the narrowest whole number of bytes the value fits in.
func byteWidth(value *big.Int, negative bool) int {
	need := value.BitLen()
	if negative {
		// Two's complement reaches one further down than up.
		need = new(big.Int).Sub(value, big.NewInt(1)).BitLen() + 1
	}
	for width := 8; width < 1<<16; width += 8 {
		if need <= width {
			return width
		}
	}
	return need
}

// signedReading is the two's complement meaning of the pattern, and whether
// that differs from the plain reading.
func signedReading(pattern *big.Int, width int, negative bool) (string, bool) {
	if negative || pattern.Bit(width-1) == 0 {
		return "", false
	}
	signed := new(big.Int).Sub(pattern, new(big.Int).Lsh(big.NewInt(1), uint(width)))
	return signed.String(), true
}

func signedText(value *big.Int, negative bool) string {
	if negative {
		return "-" + value.String()
	}
	return value.String()
}

// bitsOf writes the low width bits of the pattern, most significant first.
func bitsOf(pattern *big.Int, width int) string {
	var b strings.Builder
	for i := width - 1; i >= 0; i-- {
		b.WriteByte('0' + byte(pattern.Bit(i)))
	}
	return b.String()
}
