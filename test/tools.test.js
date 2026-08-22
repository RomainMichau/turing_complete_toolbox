// A readable pass over the behaviour the README promises. The exhaustive
// checking lives in golden.test.js — this file is here to be read.

import test from "node:test";
import assert from "node:assert/strict";

import { run } from "../docs/tools/index.js";

const values = (id, inputs) =>
  Object.fromEntries(run(id, inputs).fields.map((f) => [f.label, f.value]));

test("the number converter guesses how to read what you type", () => {
  // Nothing but 0 and 1 is taken as bits, at the width it was typed.
  assert.equal(values("number", { value: "1100 1101" })["Read as"], "binary, 8 bits (1 byte)");
  // A prefix settles it.
  assert.equal(values("number", { value: "0xCD" })["Read as"], "hex, 8 bits (1 byte)");
  // Anything else is decimal, at the narrowest whole number of bytes.
  assert.equal(values("number", { value: "205" })["Read as"], "decimal, 8 bits (1 byte)");
  // And the Read as buttons override the guess.
  assert.equal(values("number", { value: "1100 1101", base: "dec" })["Read as"],
    "decimal, 24 bits (3 bytes)");
});

test("separators are ignored and negatives come out in two's complement", () => {
  assert.equal(values("number", { value: "1100_1101" }).Decimal, "205");
  assert.equal(values("number", { value: "-5" }).Hex, "0xFB");
  assert.equal(values("number", { value: "11111111" })["Signed (8-bit two's complement)"], "-1");
});

test("a half typed value is not an error", () => {
  assert.deepEqual(run("number", { value: "" }), { fields: [], error: "" });
  assert.deepEqual(run("number", { value: "0x" }), { fields: [], error: "" });
});

test("the encoder lays the fields out into a word", () => {
  // mode 3 (RAM), opcode 6 (store_32).
  const got = values("symphony-instruction",
    { mode: "11", imm: "0", op: "0110", dest: "0000", argA: "0000", argB: "0000" });
  assert.equal(got.Int, "1711276032");
  assert.equal(got.Hex, "0x66000000");
  assert.equal(got.Bits, "01100110 00000000 00000000 00000000");
});

test("the immediate flag hands argument B the last two bytes", () => {
  const got = values("symphony-instruction",
    { mode: "11", imm: "1", op: "0110", dest: "0000", argA: "0000", argB: "0000000000000011" });
  assert.equal(got.Mode, "3 · RAM · IMM");
  assert.equal(got.Int, String(0x76000003));
});

test("the decoder reads an assembler line, letters and all", () => {
  const got = values("symphony-decode",
    { word: "01110110 0000vvvv aaaaaaaa aaaaaaaa", read: "bits" });
  assert.equal(got.Mode, "3 · RAM");
  assert.equal(got.Opcode, "6 · store_32");
  assert.equal(got["Arg A"], "variable v");
  assert.equal(got["Arg B"], "variable a (16 bits)");
  assert.equal(got.Instruction, "32 bits of reg A (v) → program memory at literal a");
});

test("the decoder takes a number as readily as bits", () => {
  const bits = values("symphony-decode", { word: "0x2A120300", read: "number" });
  assert.equal(bits.Bits, "00101010000100100000001100000000");
});

test("a short pattern is read as the low bits of the word", () => {
  // Only eight characters, so they land in the last byte and the fields above
  // them are all zero — the same way a small number would be read.
  const got = values("symphony-decode", { word: "0000vvvv", read: "bits" });
  assert.equal(got.Bits, "000000000000000000000000" + "0000vvvv");
  assert.equal(got.Mode, "0 · IO");
});

test("a field the decoder cannot pin down holds the whole answer back", () => {
  // The mode sits at bits 1-2 of the word, so it takes a full width pattern to
  // leave it unknown.
  const got = values("symphony-decode", { word: "0mm00110" + "0".repeat(24), read: "bits" });
  assert.equal(got.Mode, "variable m");
  assert.equal(got.Instruction, "needs the mode, the immediate flag and the opcode to be known");
});

test("what is not a bit is pointed at", () => {
  assert.equal(run("symphony-decode", { word: "0000!!!!", read: "bits" }).error,
    `"!" at position 5 is neither a bit nor a variable`);
});
