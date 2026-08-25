// The encoder: one box per instruction field, and the integer to feed the
// machine underneath.

import { bitsField, packWord, wordBits, hexWord } from "./bits.js";
import { WORD_BITS, layout, MODE_NAMES, describeSymphony } from "./symphony.js";

export function encodeSymphony(input) {
  // The immediate flag decides how much room argument B gets, and where —
  // read on its own first, since it picks which layout the rest packs with.
  const imm = bitsField(input, "imm", 1);
  const { word, fields: f } = packWord(layout(imm), input);

  let modeText = `${f.mode} · ${MODE_NAMES[f.mode]}`;
  if (imm === 1) modeText += " · IMM";

  return [
    { label: "Int", value: String(word) },
    { label: "Hex", value: hexWord(word, WORD_BITS) },
    { label: "Bits", value: wordBits(word, WORD_BITS) },
    { label: "Mode", value: modeText },
    {
      label: "Instruction",
      value: describeSymphony(f.mode, f.op, String(f.dest), String(f.argA), String(f.argB), imm === 1),
    },
  ];
}
