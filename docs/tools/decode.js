// The decoder reads a word back. Bits it cannot know are welcome: any letter
// stands for a variable, so an encoding line lifted straight out of the game's
// assembler definition — 01110110 0000vvvv aaaaaaaa aaaaaaaa — decodes into the
// fields it does pin down, and names the letters for the rest.

import { cleanBits, patternValue, variableName, wordPattern } from "./bits.js";
import {
  WORD_BITS, opcodeName, describeSymphony, MODE_NAMES,
  MODE_AT, IMM_AT, OP_AT, DEST_AT, ARG_A_AT, ARG_B_AT,
} from "./symphony.js";

// operand is how a field reads inside a sentence: its value when known, and
// otherwise the letter standing for it.
function operand(slice) {
  const [value, ok] = patternValue(slice);
  return ok ? String(value) : variableName(slice);
}

function describeMode(slice, mode, known) {
  if (!known) return `variable ${variableName(slice)}`;
  return `${mode} · ${MODE_NAMES[mode]}`;
}

function describeIMM(slice, imm, known) {
  if (!known) return `variable ${variableName(slice)} — argument B could be either shape`;
  if (imm === 1) return "1 — argument B is a 16 bit literal over the last two bytes";
  return "0 — a plain word";
}

function describeOpcode(slice, mode, op, modeKnown, opKnown) {
  if (!opKnown) return `variable ${variableName(slice)}`;
  if (!modeKnown) return `${op} — the mode decides what it means`;
  return `${op} · ${opcodeName(mode, op)}`;
}

function describeRegister(slice) {
  const [value, ok] = patternValue(slice);
  return ok ? `reg ${value}` : `variable ${variableName(slice)}`;
}

function describeArgB(slice, imm) {
  const [value, ok] = patternValue(slice);
  if (!ok) return `variable ${variableName(slice)} (${slice.length} bits)`;
  if (imm) return `literal ${value} (0x${value.toString(16).toUpperCase().padStart(4, "0")})`;
  return `reg ${value}`;
}

export function decodeSymphony(input) {
  const text = cleanBits(input.word || "");
  if (text === "") return [];

  const pattern = wordPattern(text, input.read, WORD_BITS);

  const [mode, modeKnown] = patternValue(pattern.slice(MODE_AT, MODE_AT + 2));
  const [imm, immKnown] = patternValue(pattern.slice(IMM_AT, IMM_AT + 1));
  const [op, opKnown] = patternValue(pattern.slice(OP_AT, OP_AT + 4));

  // Argument B is four bits, unless the immediate flag hands it the last two
  // bytes. With the flag itself unknown, the narrow reading is the one shown.
  const isIMM = immKnown && imm === 1;
  const [argBFrom, argBTo] = isIMM ? [16, WORD_BITS] : [ARG_B_AT, ARG_B_AT + 4];

  const fields = [
    { label: "Bits", value: pattern, format: "bits" },
    { label: "Mode", value: describeMode(pattern.slice(MODE_AT, MODE_AT + 2), mode, modeKnown) },
    { label: "Immediate", value: describeIMM(pattern.slice(IMM_AT, IMM_AT + 1), imm, immKnown) },
    { label: "Opcode", value: describeOpcode(pattern.slice(OP_AT, OP_AT + 4), mode, op, modeKnown, opKnown) },
    { label: "Dest", value: describeRegister(pattern.slice(DEST_AT, DEST_AT + 4)) },
    { label: "Arg A", value: describeRegister(pattern.slice(ARG_A_AT, ARG_A_AT + 4)) },
    { label: "Arg B", value: describeArgB(pattern.slice(argBFrom, argBTo), isIMM) },
  ];

  if (modeKnown && immKnown && opKnown) {
    fields.push({
      label: "Instruction",
      value: describeSymphony(mode, op,
        operand(pattern.slice(DEST_AT, DEST_AT + 4)),
        operand(pattern.slice(ARG_A_AT, ARG_A_AT + 4)),
        operand(pattern.slice(argBFrom, argBTo)),
        imm === 1),
    });
  } else {
    fields.push({
      label: "Instruction",
      value: "needs the mode, the immediate flag and the opcode to be known",
    });
  }
  return fields;
}
