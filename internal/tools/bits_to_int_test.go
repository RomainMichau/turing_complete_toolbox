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

func TestBitsToInt(t *testing.T) {
	cases := []struct {
		input            string
		unsigned, signed string
		hex              string
	}{
		{"0100 1101", "77", "77", "0x4D"},
		{"1111 1111", "255", "-1", "0xFF"},
		{"1000 0000", "128", "-128", "0x80"},
		{"\t0111_1111 ", "127", "127", "0x7F"},
		{"1", "1", "-1", "0x1"},
		{"0", "0", "0", "0x0"},
		{"1111111111111111111111111111111111111111111111111111111111111111",
			"18446744073709551615", "-1", "0xFFFFFFFFFFFFFFFF"},
	}

	for _, c := range cases {
		fields, err := bitsToInt(c.input)
		if err != nil {
			t.Fatalf("bitsToInt(%q): unexpected error %v", c.input, err)
		}
		if got := fieldValue(t, fields, "Unsigned"); got != c.unsigned {
			t.Errorf("bitsToInt(%q) unsigned = %s, want %s", c.input, got, c.unsigned)
		}
		if got := fieldValue(t, fields, "Signed"); got != c.signed {
			t.Errorf("bitsToInt(%q) signed = %s, want %s", c.input, got, c.signed)
		}
		if got := fieldValue(t, fields, "Hex"); got != c.hex {
			t.Errorf("bitsToInt(%q) hex = %s, want %s", c.input, got, c.hex)
		}
	}
}

func TestBitsToIntEmpty(t *testing.T) {
	fields, err := bitsToInt("   ")
	if err != nil || fields != nil {
		t.Fatalf("bitsToInt(blank) = %v, %v; want nil, nil", fields, err)
	}
}

func TestBitsToIntInvalid(t *testing.T) {
	if _, err := bitsToInt("10102"); err == nil {
		t.Fatal("bitsToInt(\"10102\"): expected an error")
	}
}

func TestGroupBits(t *testing.T) {
	cases := map[string]string{
		"1":            "1",
		"10110":        "1 0110",
		"010011010101": "0100 1101 0101",
	}
	for in, want := range cases {
		if got := groupBits(in); got != want {
			t.Errorf("groupBits(%q) = %q, want %q", in, got, want)
		}
	}
}
