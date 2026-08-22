// The encoder: one box per instruction field, and the integer to feed the
// machine underneath.

import { cleanBits } from "./bits.js";
import {
  SHIFT, ARG_B_WIDTH, ARG_B_WIDTH_IMM, MODE_NAMES, describeSymphony,
} from "./symphony.js";

// bitsField reads one named bit field, defaulting to 0 when it is missing.
function bitsField(input, id, width) {
  const raw = input[id] || "";
  const bits = cleanBits(raw);
  if (bits === "") return 0;
  if (bits.length > width) {
    throw new Error(`${id}: ${bits.length} bits given, field is ${width} bits wide`);
  }
  if (!/^[01]+$/.test(bits)) {
    throw new Error(`${id}: ${JSON.stringify(raw)} is not a binary value`);
  }
  return parseInt(bits, 2);
}

// symBits renders the whole word as bits, one group per byte.
function symBits(word) {
  let out = "";
  for (let i = 31; i >= 0; i--) {
    out += (word >>> i) & 1;
    if (i % 8 === 0 && i !== 0) out += " ";
  }
  return out;
}

export function encodeSymphony(input) {
  const imm = bitsField(input, "imm", 1);
  // The immediate flag decides how much room argument B gets, and where.
  const bWidth = imm === 1 ? ARG_B_WIDTH_IMM : ARG_B_WIDTH;
  const bShift = imm === 1 ? 0 : SHIFT.argB;

  const mode = bitsField(input, "mode", 2);
  const op = bitsField(input, "op", 4);
  const dest = bitsField(input, "dest", 4);
  const argA = bitsField(input, "argA", 4);
  const argB = bitsField(input, "argB", bWidth);

  const word = ((mode << SHIFT.mode) | (imm << SHIFT.imm) | (op << SHIFT.op) |
    (dest << SHIFT.dest) | (argA << SHIFT.argA) | (argB << bShift)) >>> 0;

  let modeText = `${mode} · ${MODE_NAMES[mode]}`;
  if (imm === 1) modeText += " · IMM";

  return [
    { label: "Int", value: String(word) },
    { label: "Hex", value: "0x" + word.toString(16).toUpperCase().padStart(8, "0") },
    { label: "Bits", value: symBits(word) },
    { label: "Mode", value: modeText },
    {
      label: "Instruction",
      value: describeSymphony(mode, op, String(dest), String(argA), String(argB), imm === 1),
    },
  ];
}
