package tools

import "testing"

func TestEncodeSymphony(t *testing.T) {
	cases := []struct {
		name        string
		in          map[string]string
		word        string
		hex         string
		instruction string
	}{
		{
			name:        "empty word",
			in:          map[string]string{},
			word:        "0",
			hex:         "0x00000000",
			instruction: "nothing",
		},
		{
			name:        "io: input to reg 5",
			in:          map[string]string{"mode": "00", "op": "0001", "dest": "0101"},
			word:        "22020096",
			hex:         "0x01500000",
			instruction: "input → reg 5",
		},
		{
			name:        "io: arg B to out",
			in:          map[string]string{"mode": "00", "op": "0010", "argB": "0111"},
			word:        "33556224",
			hex:         "0x02000700",
			instruction: "arg B (7) → out",
		},
		{
			name:        "ram: store_16 with everything else empty",
			in:          map[string]string{"mode": "11", "op": "0101"},
			word:        "1694498816",
			hex:         "0x65000000",
			instruction: "16 bits of reg A (0) → program memory at arg B (0)",
		},
		{
			name:        "alu: add",
			in:          map[string]string{"mode": "01", "op": "0100", "dest": "0010", "argA": "0110", "argB": "0001"},
			word:        "606470400",
			hex:         "0x24260100",
			instruction: "reg 2 ← arg A (6) ADD arg B (1)",
		},
		{
			name:        "alu: cmp spells its flags out",
			in:          map[string]string{"mode": "01", "op": "1010", "dest": "0001", "argA": "0010", "argB": "0011"},
			word:        "705823488",
			hex:         "0x2A120300",
			instruction: "reg 1 ← compare arg A (2) with arg B (3) — bit 0: equal, bit 1: below (unsigned), bit 2: less (signed)",
		},
		{
			name: "imm: argument B is a 16 bit literal over the last two bytes",
			in: map[string]string{
				"mode": "01", "imm": "1", "op": "0100", "dest": "0010", "argA": "0011",
				"argB": "0000000100000000", // 256
			},
			word:        "874709248",
			hex:         "0x34230100",
			instruction: "reg 2 ← arg A (3) ADD literal 256",
		},
		{
			name:        "imm: io sends a literal out",
			in:          map[string]string{"mode": "00", "imm": "1", "op": "0010", "argB": "0000000000000111"},
			word:        "301989895",
			hex:         "0x12000007",
			instruction: "literal 7 → out",
		},
		{
			name:        "ram: load_16 into a register",
			in:          map[string]string{"mode": "11", "op": "0001", "dest": "0010", "argB": "0101"},
			word:        "1629488384",
			hex:         "0x61200500",
			instruction: "reg 2 ← 16 bits of program memory at arg B (5)",
		},
		{
			name:        "ram: store_32 out of a register",
			in:          map[string]string{"mode": "11", "op": "0110", "argA": "0011", "argB": "0111"},
			word:        "1711474432",
			hex:         "0x66030700",
			instruction: "32 bits of reg A (3) → program memory at arg B (7)",
		},
		{
			name:        "ram: pstore reaches persistent storage",
			in:          map[string]string{"mode": "11", "op": "0111", "argA": "0101", "argB": "0010"},
			word:        "1728381440",
			hex:         "0x67050200",
			instruction: "32 bits of reg A (5) → persistent storage at arg B (2)",
		},
		{
			name:        "ram: pload reaches persistent storage",
			in:          map[string]string{"mode": "11", "op": "0011", "dest": "0110", "argB": "0010"},
			word:        "1667236352",
			hex:         "0x63600200",
			instruction: "reg 6 ← 32 bits of persistent storage at arg B (2)",
		},
		{
			name:        "jump: condition read off the flags register",
			in:          map[string]string{"mode": "10", "op": "0011", "argA": "0010", "argB": "0111"},
			word:        "1124206336",
			hex:         "0x43020700",
			instruction: "jbe · jump to arg B (7) if below or equal (unsigned) — flags in reg A (2)",
		},
		{
			name:        "jump: opcode 8 is unconditional",
			in:          map[string]string{"mode": "10", "op": "1000", "argB": "1001"},
			word:        "1207961856",
			hex:         "0x48000900",
			instruction: "jump to arg B (9)",
		},
		{
			name: "every field at once",
			in:   map[string]string{"mode": "01", "op": "1111", "dest": "1010", "argA": "0011", "argB": "1100"},
			// 00101111 10100011 00001100 00000000
			word:        "799214592",
			hex:         "0x2FA30C00",
			instruction: "ALU opcode 15 — not documented yet", // past the end of the opcode list
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			fields, err := encodeSymphony(c.in)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			for label, want := range map[string]string{
				"Int": c.word, "Hex": c.hex, "Instruction": c.instruction,
			} {
				if got := fieldValue(t, fields, label); got != want {
					t.Errorf("%s = %q, want %q", label, got, want)
				}
			}
		})
	}
}

func TestEncodeSymphonyRejectsOversizedField(t *testing.T) {
	if _, err := encodeSymphony(map[string]string{"mode": "101"}); err == nil {
		t.Fatal("expected an error for a 3 bit mode")
	}
}

func TestSymBits(t *testing.T) {
	// mode 1, opcode 1, dest 1: 0 01 0 0001 00010000 00000000 00000000
	got := symBits(1<<symModeShift | 1<<symOpShift | 1<<symDestShift)
	want := "00100001 00010000 00000000 00000000"
	if got != want {
		t.Errorf("symBits = %q, want %q", got, want)
	}
}
