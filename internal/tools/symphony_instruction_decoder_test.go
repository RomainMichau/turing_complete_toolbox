package tools

import "testing"

func TestDecodeSymphony(t *testing.T) {
	cases := []struct {
		name        string
		word, read  string
		mode        string
		imm         string
		opcode      string
		dest        string
		argA        string
		argB        string
		instruction string
	}{
		{
			name: "a plain ALU word",
			word: "00100100 00100110 00000001 00000000", read: readBits,
			mode: "1 · ALU", imm: "0 — a plain word", opcode: "4 · ADD",
			dest: "reg 2", argA: "reg 6", argB: "reg 1",
			instruction: "reg 2 ← arg A (6) ADD arg B (1)",
		},
		{
			name: "an immediate word takes the last two bytes",
			word: "00110100 00100011 00000001 00000000", read: readBits,
			mode: "1 · ALU", imm: "1 — argument B is a 16 bit literal over the last two bytes",
			opcode: "4 · ADD", dest: "reg 2", argA: "reg 3", argB: "literal 256 (0x0100)",
			instruction: "reg 2 ← arg A (3) ADD literal 256",
		},
		{
			name: "a number instead of bits",
			word: "705823488", read: readNumber,
			mode: "1 · ALU", imm: "0 — a plain word", opcode: "10 · CMP",
			dest: "reg 1", argA: "reg 2", argB: "reg 3",
			instruction: "reg 1 ← compare arg A (2) with arg B (3) — " + cmpFlags,
		},
		{
			name: "hex is a number too",
			word: "0x48000900", read: readNumber,
			mode: "2 · JUMP", imm: "0 — a plain word", opcode: "8 · jmp",
			dest: "reg 0", argA: "reg 0", argB: "reg 9",
			instruction: "jump to arg B (9)",
		},
		{
			name: "a store lifted out of the assembler definition",
			word: "01110110 0000vvvv aaaaaaaa aaaaaaaa", read: readBits,
			mode: "3 · RAM", imm: "1 — argument B is a 16 bit literal over the last two bytes",
			opcode: "6 · store_32", dest: "reg 0", argA: "variable v", argB: "variable a (16 bits)",
			instruction: "32 bits of reg A (v) → program memory at literal a",
		},
		{
			name: "the layout itself decodes to variables",
			word: "XMMIOOOO DDDDAAAA XXXXBBBB XXXXXXXX", read: readBits,
			mode: "variable M", imm: "variable I — argument B could be either shape",
			opcode: "variable O", dest: "variable D", argA: "variable A", argB: "variable B (4 bits)",
			instruction: "needs the mode, the immediate flag and the opcode to be known",
		},
		{
			// Right aligned like a number, so these four bits land in the
			// trailing unused byte rather than in argument B.
			name: "a short pattern is read as the low bits of the word",
			word: "1010", read: readBits,
			mode: "0 · IO", imm: "0 — a plain word", opcode: "0 · nothing",
			dest: "reg 0", argA: "reg 0", argB: "reg 0",
			instruction: "nothing",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			fields, err := decodeSymphony(map[string]string{"word": c.word, "read": c.read})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			for label, want := range map[string]string{
				"Mode": c.mode, "Immediate": c.imm, "Opcode": c.opcode,
				"Dest": c.dest, "Arg A": c.argA, "Arg B": c.argB,
				"Instruction": c.instruction,
			} {
				if got := fieldValue(t, fields, label); got != want {
					t.Errorf("%s = %q, want %q", label, got, want)
				}
			}
		})
	}
}

// What the encoder builds, the decoder must read back.
func TestDecodeUndoesEncode(t *testing.T) {
	in := map[string]string{
		"mode": "01", "imm": "1", "op": "0100", "dest": "0010", "argA": "0011",
		"argB": "0000000100000000",
	}
	encoded, err := encodeSymphony(in)
	if err != nil {
		t.Fatal(err)
	}
	word := fieldValue(t, encoded, "Int")

	decoded, err := decodeSymphony(map[string]string{"word": word, "read": readNumber})
	if err != nil {
		t.Fatal(err)
	}
	if got, want := fieldValue(t, decoded, "Instruction"), fieldValue(t, encoded, "Instruction"); got != want {
		t.Errorf("decoded %q, encoder said %q", got, want)
	}
}

func TestDecodeSymphonyBadInput(t *testing.T) {
	if fields, err := decodeSymphony(map[string]string{"word": "  ", "read": readBits}); err != nil || fields != nil {
		t.Fatalf("blank input = %v, %v; want nil, nil", fields, err)
	}

	cases := []struct{ word, read, why string }{
		{"0100 1101 2", readBits, "2 is neither a bit nor a variable"},
		{strings1(33), readBits, "longer than a word"},
		{"twelve", readNumber, "not a number"},
		{"4294967296", readNumber, "wider than 32 bits"},
	}
	for _, c := range cases {
		if _, err := decodeSymphony(map[string]string{"word": c.word, "read": c.read}); err == nil {
			t.Errorf("decoding %q as %s should have failed: %s", c.word, c.read, c.why)
		}
	}
}

func strings1(n int) string {
	out := make([]byte, n)
	for i := range out {
		out[i] = '1'
	}
	return string(out)
}
