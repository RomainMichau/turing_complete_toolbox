// The decoder reads a word back. Bits it cannot know are welcome: any letter
// stands for a variable, so an encoding line lifted straight out of the game's
// assembler definition — 01110110 0000vvvv aaaaaaaa aaaaaaaa — decodes into the
// fields it does pin down, and names the letters for the rest.

import { cleanBits, parseDigits } from "./bits.js";
import { WORD_BITS, opcodeName, describeSymphony, MODE_NAMES } from "./symphony.js";

// The fields of a word, as slices of the 32 character pattern.
const MODE_AT = 1;
const IMM_AT = 3;
const OP_AT = 4;
const DEST_AT = 8;
const ARG_A_AT = 12;
const ARG_B_AT = 20; // 16 in IMM mode, where argument B takes the last two bytes

const UINT64_MAX = (1n << 64n) - 1n;
const UINT32_MAX = (1n << 32n) - 1n;

// parseNumber reads a literal the way Go's strconv.ParseUint does with base 0:
// the prefix picks the base, a bare leading zero means octal, and no sign is
// allowed. Returns null when the text is not a number at all.
function parseNumber(text) {
  if (text === "") return null;
  let radix = 10;
  let digits = text;
  if (text.startsWith("0x")) [radix, digits] = [16, text.slice(2)];
  else if (text.startsWith("0b")) [radix, digits] = [2, text.slice(2)];
  else if (text.startsWith("0o")) [radix, digits] = [8, text.slice(2)];
  else if (text.length > 1 && text[0] === "0") [radix, digits] = [8, text.slice(1)];

  const value = parseDigits(digits, radix);
  if (value === null || value > UINT64_MAX) return null;
  return value;
}

// wordPattern turns the input into exactly 32 characters of bits and variables.
function wordPattern(text, read) {
  if (read === "number") {
    const value = parseNumber(text.toLowerCase());
    if (value === null) {
      throw new Error(`${JSON.stringify(text)} is not a number — try 705823488 or 0x2A120300`);
    }
    if (value > UINT32_MAX) {
      throw new Error(`${text} does not fit in a 32 bit word`);
    }
    return value.toString(2).padStart(WORD_BITS, "0");
  }

  for (let i = 0; i < text.length; i++) {
    const r = text[i];
    const bit = r === "0" || r === "1";
    const letter = (r >= "a" && r <= "z") || (r >= "A" && r <= "Z");
    if (!bit && !letter) {
      throw new Error(`${JSON.stringify(r)} at position ${i + 1} is neither a bit nor a variable`);
    }
  }
  if (text.length > WORD_BITS) {
    throw new Error(`${text.length} bits given, a word is ${WORD_BITS}`);
  }
  // A short pattern is read as the low bits of the word, like a number is.
  return "0".repeat(WORD_BITS - text.length) + text;
}

// patternValue reads a run of the pattern, and says whether it was all bits.
function patternValue(slice) {
  if (!/^[01]+$/.test(slice)) return [0, false];
  return [parseInt(slice, 2), true];
}

// variableName is the letter a field is filled with, or the raw run when it is
// a mix of bits and letters.
function variableName(slice) {
  const first = slice[0];
  return [...slice].every((c) => c === first) ? first : slice;
}

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

  const pattern = wordPattern(text, input.read);

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
