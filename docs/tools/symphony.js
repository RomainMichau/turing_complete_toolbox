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

import { sliceFields, shiftToSlice } from "./bits.js";

export const WORD_BITS = 32;

// Where each field sits in the word: how far left it is shifted from the
// LSB, and how wide. Everything else here — the encoder's packing, the
// decoder's slice offsets below — is derived from this one table, instead
// of each hand-maintaining its own copy of the same layout.
export const FIELDS = {
  mode: { shift: 29, width: 2 },
  imm: { shift: 28, width: 1 },
  op: { shift: 24, width: 4 },
  dest: { shift: 20, width: 4 },
  argA: { shift: 16, width: 4 },
  argB: { shift: 8, width: 4 },
};
// With the immediate flag set, argument B is a 16 bit literal over the low
// two bytes instead.
const ARG_B_IMM = { shift: 0, width: 16 };

// layout is FIELDS in word order, argB swapped for its IMM-mode shape when
// asked — what the encoder packs a word from.
export function layout(imm) {
  return [
    ["mode", FIELDS.mode], ["imm", FIELDS.imm], ["op", FIELDS.op],
    ["dest", FIELDS.dest], ["argA", FIELDS.argA],
    ["argB", imm ? ARG_B_IMM : FIELDS.argB],
  ];
}

// The same fields, as the start of their slice offset into the 32 character
// pattern (MSB first) the decoder and the encoder's segmented boxes read —
// each one's own width completes the range at its use site.
const at = (id) => shiftToSlice(FIELDS[id].shift, FIELDS[id].width, WORD_BITS)[0];
export const MODE_AT = at("mode");
export const IMM_AT = at("imm");
export const OP_AT = at("op");
export const DEST_AT = at("dest");
export const ARG_A_AT = at("argA");
export const ARG_B_AT = at("argB"); // 16 in IMM mode, where argument B takes the last two bytes

// wordFields cuts a full 32 bit pattern — every bit known, none of it a
// decoder's variable letter — into the named fields the encoder's boxes
// hold, so a decoded word can be handed straight to the encoder to tweak.
export function wordFields(pattern) {
  const isIMM = pattern[IMM_AT] === "1";
  return sliceFields(pattern, [
    ["mode", [MODE_AT, MODE_AT + 2]],
    ["imm", [IMM_AT, IMM_AT + 1]],
    ["op", [OP_AT, OP_AT + 4]],
    ["dest", [DEST_AT, DEST_AT + 4]],
    ["argA", [ARG_A_AT, ARG_A_AT + 4]],
    ["argB", isIMM ? [16, WORD_BITS] : [ARG_B_AT, ARG_B_AT + 4]],
  ]);
}

// Modes, in the order of their encoding.
export const MODE_IO = 0;
export const MODE_ALU = 1;
export const MODE_JUMP = 2;
export const MODE_RAM = 3;

export const MODE_NAMES = ["IO", "ALU", "JUMP", "RAM"];

// The keyboard has two output pins, merged into the single value the keyboard
// opcode reads: D is key down, V is the key value.
const KEYBOARD_LAYOUT = "XXXXXXXD VVVVVVVV";

// IO opcodes, in the order of their encoding.
export const IO_OPS = [
  { name: "", desc: "nothing" },
  { name: "", desc: "send input to dest" },
  { name: "", desc: "send argument B to out" },
  { name: "keyboard", desc: "read the keyboard into dest" },
  { name: "screen", desc: "set the screen settings from arg A and arg B" },
  { name: "time_0", desc: "read the low half of time into dest, and store the high half for time_1" },
  { name: "time_1", desc: "read the high half time_0 stored — 0 until time_0 has run" },
  { name: "counter", desc: "read the counter into dest" },
];

// ALU opcodes, in the order of their encoding.
export const ALU_OPS = ["NAND", "OR", "AND", "NOR", "ADD", "SUB", "XOR", "LSL", "LSR", "ASR", "CMP"];

// aluCMP writes flags rather than a value, so it reads differently.
const ALU_CMP = 10;

// RAM opcodes, in the order of their encoding. The first half loads from
// memory, the second half stores to it, and the suffix is how many bits move.
export const RAM_OPS = [
  "load_8", "load_16", "load_32", "pload",
  "store_8", "store_16", "store_32", "pstore",
];

const RAM_STORE = 4; // opcodes from here on store instead of load
const RAM_PLOAD = 3;
const RAM_PSTORE = 7;

// cmpFlags is how the result of a CMP is laid out. Note that there is no
// "greater" flag: greater is reached by inverting a jump condition.
const CMP_FLAGS = "bit 0: equal, bit 1: below (unsigned), bit 2: less (signed)";

// A JUMP opcode is a condition. Its low three bits are matched against the
// flags register, the jump firing when any bit matches, and bit 3 inverts the
// answer — so opcode 8, the inverse of "no flag at all", is the plain jump.
const JUMP_INVERT = 8;
const JUMP_NEVER = 0;

export const JUMP_CONDS = [
  { name: "never", desc: "never" },
  { name: "je", desc: "equal" },
  { name: "jb", desc: "below (unsigned)" },
  { name: "jbe", desc: "below or equal (unsigned)" },
  { name: "jl", desc: "less (signed)" },
  { name: "jle", desc: "less or equal (signed)" },
  { name: "", desc: "below or less" },
  { name: "", desc: "equal, below or less" },
  { name: "jmp", desc: "always" },
  { name: "jne", desc: "not equal" },
  { name: "jae", desc: "above or equal (unsigned)" },
  { name: "ja", desc: "above (unsigned)" },
  { name: "jge", desc: "greater or equal (signed)" },
  { name: "jg", desc: "greater (signed)" },
  { name: "", desc: "neither below nor less" },
  { name: "", desc: "no flag set — above and greater" },
];

export const opLabel = (op) => (op.name !== "" ? op.name : op.desc);

// opcodeName is what an opcode is called in its mode.
export function opcodeName(mode, op) {
  switch (mode) {
    case MODE_IO:
      if (op < IO_OPS.length) return opLabel(IO_OPS[op]);
      break;
    case MODE_ALU:
      if (op < ALU_OPS.length) return ALU_OPS[op];
      break;
    case MODE_JUMP:
      return opLabel(JUMP_CONDS[op & 0xf]);
    case MODE_RAM:
      if (op < RAM_OPS.length) return RAM_OPS[op];
      break;
  }
  return "not documented";
}

// describeSymphony spells the instruction out. Only the JUMP conditions and the
// IO, ALU and RAM opcodes documented so far are named; anything else says it is
// not documented. The operands are text, not numbers, so a decoder can hand
// over the letter of a field it does not know.
export function describeSymphony(mode, op, dest, argA, argB, imm) {
  // With the immediate flag set, argument B carries a value, not an address.
  const b = imm ? `literal ${argB}` : `arg B (${argB})`;

  switch (mode) {
    case MODE_IO:
      switch (op) {
        case 0: return "nothing";
        case 1: return `input → reg ${dest}`;
        case 2: return `${b} → out`;
        case 3: return `keyboard → reg ${dest} (${KEYBOARD_LAYOUT})`;
        case 4: return `screen setting arg A (${argA}) ← ${b}`;
        case 5: return `reg ${dest} ← low half of time, high half stored for time_1`;
        case 6: return `reg ${dest} ← high half of time, as stored by time_0`;
        case 7: return `reg ${dest} ← counter`;
      }
      break;
    case MODE_ALU:
      if (op === ALU_CMP) {
        return `reg ${dest} ← compare arg A (${argA}) with ${b} — ${CMP_FLAGS}`;
      }
      if (op < ALU_OPS.length) {
        return `reg ${dest} ← arg A (${argA}) ${ALU_OPS[op]} ${b}`;
      }
      break;
    case MODE_JUMP: {
      const cond = JUMP_CONDS[op & 0xf];
      if (op === JUMP_NEVER) return "never jumps";
      if (op === JUMP_INVERT) return `jump to ${b}`; // the plain, unconditional jump
      const line = `jump to ${b} if ${cond.desc} — flags in reg A (${argA})`;
      return cond.name !== "" ? `${cond.name} · ${line}` : line;
    }
    case MODE_RAM:
      if (op < RAM_OPS.length) {
        // The persistent pair always moves a full 32 bits, and reaches
        // the 3D memory rather than program memory.
        if (op === RAM_PLOAD) return `reg ${dest} ← 32 bits of persistent storage at ${b}`;
        if (op === RAM_PSTORE) return `32 bits of reg A (${argA}) → persistent storage at ${b}`;
        const width = 8 << (op % RAM_STORE); // 8, 16 or 32 bits
        if (op < RAM_STORE) return `reg ${dest} ← ${width} bits of program memory at ${b}`;
        return `${width} bits of reg A (${argA}) → program memory at ${b}`;
      }
      break;
  }
  return `${MODE_NAMES[mode]} opcode ${op} — not documented yet`;
}
