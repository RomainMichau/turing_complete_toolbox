package tools

import "testing"

func fieldValue(t *testing.T, fields []Field, prefix string) string {
	t.Helper()
	for _, f := range fields {
		if len(f.Label) >= len(prefix) && f.Label[:len(prefix)] == prefix {
			return f.Value
		}
	}
	t.Fatalf("no field with label prefix %q in %v", prefix, fields)
	return ""
}

func TestConvertNumber(t *testing.T) {
	cases := []struct {
		name        string
		value, base string
		read        string
		decimal     string
		hex         string
		binary      string
		signed      string // empty when no signed row is expected
	}{
		{
			name: "bits are guessed without a prefix", value: "0100 1101", base: "auto",
			read: "binary, 8 bits (1 byte)", decimal: "77", hex: "0x4D", binary: "01001101",
		},
		{
			name: "a decimal cannot be mistaken for bits", value: "205", base: "auto",
			read: "decimal, 8 bits (1 byte)", decimal: "205", hex: "0xCD", binary: "11001101",
			signed: "-51",
		},
		{
			name: "hex by prefix", value: "0x4DF0", base: "auto",
			read: "hex, 16 bits (2 bytes)", decimal: "19952", hex: "0x4DF0", binary: "0100110111110000",
		},
		{
			name: "octal by prefix", value: "0o315", base: "auto",
			read: "octal, 8 bits (1 byte)", decimal: "205", hex: "0xCD", binary: "11001101",
			signed: "-51",
		},
		{
			name: "the choice wins over the guess", value: "1010", base: "dec",
			read: "decimal, 16 bits (2 bytes)", decimal: "1010", hex: "0x03F2", binary: "0000001111110010",
		},
		{
			name: "hex without its prefix", value: "CD", base: "hex",
			read: "hex, 8 bits (1 byte)", decimal: "205", hex: "0xCD", binary: "11001101",
			signed: "-51",
		},
		{
			name: "a negative is shown in two's complement", value: "-51", base: "dec",
			read: "decimal, 8 bits (1 byte)", decimal: "-51", hex: "0xCD", binary: "11001101",
		},
		{
			name: "bits keep the width they were typed at", value: "0001", base: "auto",
			read: "binary, 4 bits", decimal: "1", hex: "0x1", binary: "0001",
		},
		{
			name: "wider than a machine word", value: "0xFFFFFFFFFFFFFFFFFF", base: "auto",
			read:    "hex, 72 bits (9 bytes)",
			decimal: "4722366482869645213695", hex: "0xFFFFFFFFFFFFFFFFFF",
			binary: "111111111111111111111111111111111111111111111111111111111111111111111111",
			signed: "-1",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			fields, err := convertNumber(map[string]string{"value": c.value, "base": c.base})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			for label, want := range map[string]string{
				"Read as": c.read, "Decimal": c.decimal, "Hex": c.hex, "Binary": c.binary,
			} {
				if got := fieldValue(t, fields, label); got != want {
					t.Errorf("%s = %q, want %q", label, got, want)
				}
			}

			signed := ""
			for _, f := range fields {
				if len(f.Label) > 6 && f.Label[:6] == "Signed" {
					signed = f.Value
				}
			}
			if signed != c.signed {
				t.Errorf("signed reading = %q, want %q", signed, c.signed)
			}
		})
	}
}

func TestConvertNumberPadsToAWord(t *testing.T) {
	fields, err := convertNumber(map[string]string{"value": "205", "base": "dec"})
	if err != nil {
		t.Fatal(err)
	}
	if got := fieldValue(t, fields, "32 bits"); got != "00000000000000000000000011001101" {
		t.Errorf("padded value = %q", got)
	}

	wide, _ := convertNumber(map[string]string{"value": "0xD1000000", "base": "auto"})
	for _, f := range wide {
		if f.Label == "32 bits" {
			t.Error("a 32 bit value should not be padded to 32 bits again")
		}
	}
}

func TestConvertNumberBitsAreColoured(t *testing.T) {
	fields, _ := convertNumber(map[string]string{"value": "205", "base": "dec"})
	for _, f := range fields {
		if (f.Label == "Binary" || f.Label == "32 bits") && f.Format != "bits" {
			t.Errorf("%s asks for %q formatting, want %q", f.Label, f.Format, "bits")
		}
	}
}

func TestConvertNumberQuietAndLoudFailures(t *testing.T) {
	quiet := []string{"", "   ", "0x", "0b"}
	for _, in := range quiet {
		fields, err := convertNumber(map[string]string{"value": in, "base": "auto"})
		if err != nil || fields != nil {
			t.Errorf("convertNumber(%q) = %v, %v; want nil, nil while still typing", in, fields, err)
		}
	}

	loud := []struct{ value, base string }{
		{"twelve", "auto"},
		{"12", "bin"},
		{"0xZZ", "auto"},
	}
	for _, c := range loud {
		if _, err := convertNumber(map[string]string{"value": c.value, "base": c.base}); err == nil {
			t.Errorf("convertNumber(%q, %q) should have failed", c.value, c.base)
		}
	}
}

func TestDescribeWidth(t *testing.T) {
	cases := map[int]string{
		3:  "3 bits",
		8:  "8 bits (1 byte)",
		12: "12 bits (1 byte + 4 bits)",
		16: "16 bits (2 bytes)",
	}
	for in, want := range cases {
		if got := describeWidth(in); got != want {
			t.Errorf("describeWidth(%d) = %q, want %q", in, got, want)
		}
	}
}
