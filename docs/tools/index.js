// The toolbox: every tool that computes something, keyed by the id its
// descriptor carries in registry.js. A tool missing from here is pure
// reference documentation.

import { convertNumber } from "./number.js";
import { encodeSymphony } from "./encode.js";
import { decodeSymphony } from "./decode.js";

export const RUNNERS = {
  "number": convertNumber,
  "symphony-instruction": encodeSymphony,
  "symphony-decode": decodeSymphony,
};

// run answers a tool with either its fields or the message explaining why it
// could not. It never throws: a half typed value is a normal state here.
export function run(id, inputs) {
  const runner = RUNNERS[id];
  if (!runner) return { fields: [], error: "" };
  try {
    return { fields: runner(inputs) || [], error: "" };
  } catch (e) {
    return { fields: [], error: e.message };
  }
}
