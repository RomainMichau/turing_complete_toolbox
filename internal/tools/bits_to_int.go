package tools

import (
	"fmt"
	"math/big"
	"strings"
	"unicode"
)

func init() {
	Register(&Tool{
		ID:          "bits-to-int",
		Name:        "Bits → Int",
		Description: "Convert a binary string into its integer value. Spaces, tabs and underscores are ignored.",
		Placeholder: "0100 1101",
		Run:         bitsToInt,
	})
}

// cleanBits strips the separators a human naturally types between bits.
func cleanBits(input string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) || r == '_' {
			return -1
		}
		return r
	}, input)
}

// groupBits re-groups a bit string in nibbles of 4, from the right.
func groupBits(bits string) string {
	var parts []string
	for end := len(bits); end > 0; end -= 4 {
		start := end - 4
		if start < 0 {
			start = 0
		}
		parts = append([]string{bits[start:end]}, parts...)
	}
	return strings.Join(parts, " ")
}

func bitsToInt(input string) ([]Field, error) {
	bits := cleanBits(input)
	if bits == "" {
		return nil, nil
	}
	for i, r := range bits {
		if r != '0' && r != '1' {
			return nil, fmt.Errorf("invalid character %q at position %d: only 0 and 1 are allowed", r, i+1)
		}
	}

	width := len(bits)
	unsigned, _ := new(big.Int).SetString(bits, 2)

	// Two's complement reading over the width actually typed.
	signed := new(big.Int).Set(unsigned)
	if bits[0] == '1' {
		signed.Sub(signed, new(big.Int).Lsh(big.NewInt(1), uint(width)))
	}

	hexDigits := (width + 3) / 4
	return []Field{
		{Label: "Unsigned", Value: unsigned.String()},
		{Label: fmt.Sprintf("Signed (%d-bit two's complement)", width), Value: signed.String()},
		{Label: "Hex", Value: fmt.Sprintf("0x%0*X", hexDigits, unsigned)},
		{Label: "Octal", Value: "0o" + unsigned.Text(8)},
		{Label: "Width", Value: fmt.Sprintf("%d bits", width)},
		{Label: "Normalized", Value: groupBits(bits)},
	}, nil
}
