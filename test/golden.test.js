// The golden file was produced by the Go implementation this toolbox was
// ported from: every case is an input and the exact answer the Go gave for it.
// It is the spec, and it is why the port could be trusted — the whole corpus
// matched byte for byte on the day the Go was removed.
//
// A handful of symphony-decode cases with a 0x/0b/0o-prefixed word under
// "Read as: Bits" were updated after that day: the decoder now shares its
// word-parsing with riscv-toolbox's, which auto-detects that prefix
// regardless of the Read-as toggle — a real behavior change, adopted
// deliberately, not a Go-parity bug.

import test from "node:test";
import assert from "node:assert/strict";
import { readFileSync } from "node:fs";

import { run } from "../docs/tools/index.js";

const golden = JSON.parse(readFileSync(new URL("./golden.json", import.meta.url), "utf8"));

// The Go left `format` off a field that had none, so drop it here too before
// comparing.
const strip = (fields) =>
  fields.map((f) => (f.format ? { label: f.label, value: f.value, format: f.format }
                              : { label: f.label, value: f.value }));

test(`every one of the ${golden.length} recorded cases still answers as the Go did`, () => {
  for (const { tool, inputs, expect } of golden) {
    const got = run(tool, inputs);
    assert.deepEqual(
      { error: got.error, fields: strip(got.fields) },
      expect,
      `${tool} ${JSON.stringify(inputs)}`,
    );
  }
});
