// The tool descriptors: static data describing what each tool shows. This
// started as the Go registry the old GET /api/tools endpoint served, dumped
// out at migration time, and is maintained here by hand since.

import { NUMBER_TOOL } from "./tools/number.js";
import { sendableFromBits } from "./tools/bits.js";
import { WORD_BITS, wordFields } from "./tools/symphony.js";

export const TOOLS = [
  NUMBER_TOOL,
  {
    "id": "symphony-doc",
    "name": "Instruction Doc",
    "family": "Symphony",
    "description": "The word layout and every opcode, one fold per mode.",
    "doc": [
      {
        "title": "Word",
        "kind": "group",
        "text": "How a word is laid out, what each field means, and the four modes.",
        "sections": [
          {
            "title": "Layout",
            "kind": "layout",
            "text": "XMMIOOOO DDDDAAAA XXXXBBBB XXXXXXXX",
            "colors": {
              "A": "green",
              "B": "blue",
              "D": "yellow",
              "I": "purple",
              "M": "orange",
              "O": "red",
              "X": "muted"
            }
          },
          {
            "title": "Fields",
            "kind": "legend",
            "rows": [
              {
                "key": "M",
                "value": "Mode",
                "color": "orange"
              },
              {
                "key": "I",
                "value": "Immediate flag — bit 4",
                "color": "purple"
              },
              {
                "key": "O",
                "value": "OpCode",
                "color": "red"
              },
              {
                "key": "D",
                "value": "Destination register",
                "color": "yellow"
              },
              {
                "key": "A",
                "value": "Argument A — a register address",
                "color": "green"
              },
              {
                "key": "B",
                "value": "Argument B — a register address, a program memory address in RAM mode, or the value itself in IMM mode",
                "color": "blue"
              },
              {
                "key": "X",
                "value": "Unused",
                "color": "muted"
              }
            ]
          },
          {
            "kind": "note",
            "text": "Three stores: the registers, which the destination and the arguments address; program memory, which the RAM loads and stores reach; and persistent storage, reached only by pload and pstore."
          },
          {
            "title": "In IMM mode",
            "kind": "layout",
            "text": "XMM1OOOO DDDDAAAA BBBBBBBB BBBBBBBB",
            "colors": {
              "1": "purple",
              "A": "green",
              "B": "blue",
              "D": "yellow",
              "I": "purple",
              "M": "orange",
              "O": "red",
              "X": "muted"
            }
          },
          {
            "kind": "note",
            "text": "Everything in the mode folds holds as written unless bit 4 is set. In IMM mode argument B is the value itself instead of an address, and it grows to 16 bits over the last two bytes."
          },
          {
            "title": "Modes",
            "kind": "table",
            "rows": [
              {
                "key": "0",
                "value": "IO",
                "color": "orange"
              },
              {
                "key": "1",
                "value": "ALU",
                "color": "orange"
              },
              {
                "key": "2",
                "value": "JUMP",
                "color": "orange"
              },
              {
                "key": "3",
                "value": "RAM",
                "color": "orange"
              }
            ]
          }
        ]
      },
      {
        "title": "IO",
        "kind": "group",
        "text": "Mode 0: input, output, keyboard, screen, time and counter.",
        "sections": [
          {
            "title": "Opcodes",
            "kind": "table",
            "rows": [
              {
                "key": "0",
                "value": "nothing",
                "color": "red"
              },
              {
                "key": "1",
                "value": "send input to dest",
                "color": "red"
              },
              {
                "key": "2",
                "value": "send argument B to out",
                "color": "red"
              },
              {
                "key": "3",
                "value": "keyboard — read the keyboard into dest",
                "color": "red"
              },
              {
                "key": "4",
                "value": "screen — set the screen settings from arg A and arg B",
                "color": "red"
              },
              {
                "key": "5",
                "value": "time_0 — read the low half of time into dest, and store the high half for time_1",
                "color": "red"
              },
              {
                "key": "6",
                "value": "time_1 — read the high half time_0 stored — 0 until time_0 has run",
                "color": "red"
              },
              {
                "key": "7",
                "value": "counter — read the counter into dest",
                "color": "red"
              }
            ],
            "compact": true
          },
          {
            "title": "Keyboard result",
            "kind": "layout",
            "text": "XXXXXXXD VVVVVVVV",
            "colors": {
              "D": "purple",
              "V": "green",
              "X": "muted"
            }
          },
          {
            "kind": "legend",
            "rows": [
              {
                "key": "D",
                "value": "key down",
                "color": "purple"
              },
              {
                "key": "V",
                "value": "key value",
                "color": "green"
              }
            ]
          },
          {
            "kind": "note",
            "text": "The keyboard has two output pins, key down and key value. They are merged into the single word the keyboard opcode reads into the destination register."
          },
          {
            "kind": "note",
            "text": "The time component gives the nanoseconds since 1 January 1970 as a 64 bit number, and this architecture only handles 32, so reading it takes two instructions. Only time_0 actually reads the component: it hands back the low half and stores the high half for time_1 to pick up afterwards. Halves taken from two different cycles would belong to two different times, which is why time_1 reads the stored half instead — and reads 0 until time_0 has run."
          }
        ]
      },
      {
        "title": "ALU",
        "kind": "group",
        "text": "Mode 1: arithmetic, logic and the compare that feeds the jumps.",
        "sections": [
          {
            "title": "Opcodes",
            "kind": "table",
            "rows": [
              {
                "key": "0",
                "value": "NAND",
                "color": "red"
              },
              {
                "key": "1",
                "value": "OR",
                "color": "red"
              },
              {
                "key": "2",
                "value": "AND",
                "color": "red"
              },
              {
                "key": "3",
                "value": "NOR",
                "color": "red"
              },
              {
                "key": "4",
                "value": "ADD",
                "color": "red"
              },
              {
                "key": "5",
                "value": "SUB",
                "color": "red"
              },
              {
                "key": "6",
                "value": "XOR",
                "color": "red"
              },
              {
                "key": "7",
                "value": "LSL",
                "color": "red"
              },
              {
                "key": "8",
                "value": "LSR",
                "color": "red"
              },
              {
                "key": "9",
                "value": "ASR",
                "color": "red"
              },
              {
                "key": "10",
                "value": "CMP",
                "color": "red"
              }
            ],
            "compact": true
          },
          {
            "kind": "note",
            "text": "CMP writes flags into the destination register — bit 0: equal, bit 1: below (unsigned), bit 2: less (signed). There is no greater flag: greater is reached by inverting a jump condition."
          }
        ]
      },
      {
        "title": "JUMP",
        "kind": "group",
        "text": "Mode 2: the opcode is a condition tested against a flags register.",
        "sections": [
          {
            "title": "Opcodes",
            "kind": "table",
            "rows": [
              {
                "key": "0",
                "value": "never — never",
                "color": "red"
              },
              {
                "key": "1",
                "value": "je — equal",
                "color": "red"
              },
              {
                "key": "2",
                "value": "jb — below (unsigned)",
                "color": "red"
              },
              {
                "key": "3",
                "value": "jbe — below or equal (unsigned)",
                "color": "red"
              },
              {
                "key": "4",
                "value": "jl — less (signed)",
                "color": "red"
              },
              {
                "key": "5",
                "value": "jle — less or equal (signed)",
                "color": "red"
              },
              {
                "key": "6",
                "value": "below or less",
                "color": "red"
              },
              {
                "key": "7",
                "value": "equal, below or less",
                "color": "red"
              },
              {
                "key": "8",
                "value": "jmp — always",
                "color": "red"
              },
              {
                "key": "9",
                "value": "jne — not equal",
                "color": "red"
              },
              {
                "key": "10",
                "value": "jae — above or equal (unsigned)",
                "color": "red"
              },
              {
                "key": "11",
                "value": "ja — above (unsigned)",
                "color": "red"
              },
              {
                "key": "12",
                "value": "jge — greater or equal (signed)",
                "color": "red"
              },
              {
                "key": "13",
                "value": "jg — greater (signed)",
                "color": "red"
              },
              {
                "key": "14",
                "value": "neither below nor less",
                "color": "red"
              },
              {
                "key": "15",
                "value": "no flag set — above and greater",
                "color": "red"
              }
            ],
            "compact": true
          },
          {
            "kind": "note",
            "text": "The opcode is the condition. Its low three bits are matched pairwise against the flags register — the register at argument A, usually written by a CMP — and the jump fires if any bit matches. Bit 3 inverts that answer, which is why opcode 8 jumps unconditionally. The target is argument B."
          }
        ]
      },
      {
        "title": "RAM",
        "kind": "group",
        "text": "Mode 3: moving values between the registers, program memory and persistent storage.",
        "sections": [
          {
            "title": "Opcodes",
            "kind": "table",
            "rows": [
              {
                "key": "0",
                "value": "load_8",
                "color": "red"
              },
              {
                "key": "1",
                "value": "load_16",
                "color": "red"
              },
              {
                "key": "2",
                "value": "load_32",
                "color": "red"
              },
              {
                "key": "3",
                "value": "pload",
                "color": "red"
              },
              {
                "key": "4",
                "value": "store_8",
                "color": "red"
              },
              {
                "key": "5",
                "value": "store_16",
                "color": "red"
              },
              {
                "key": "6",
                "value": "store_32",
                "color": "red"
              },
              {
                "key": "7",
                "value": "pstore",
                "color": "red"
              }
            ],
            "compact": true
          },
          {
            "kind": "note",
            "text": "load reads the value from program memory at address argument B and stores it in the destination register. store writes the value in the argument A register to program memory at address argument B. The numerical suffix is how many bits are read or written, starting at that address."
          },
          {
            "kind": "note",
            "text": "pload and pstore reach persistent storage — the 3D memory — instead of program memory, and always move 32 bits."
          }
        ]
      }
    ]
  },
  {
    "id": "symphony-instruction",
    "name": "Instruction Encoder",
    "family": "Symphony",
    "description": "Build an instruction bit by bit and read the integer to feed the machine. The doc above spells out the layout and every opcode.",
    "inputs": [
      {
        "width": 1,
        "value": "0"
      },
      {
        "id": "mode",
        "label": "mode",
        "width": 2,
        "color": "orange"
      },
      {
        "id": "imm",
        "label": "imm",
        "width": 1,
        "color": "purple"
      },
      {
        "id": "op",
        "label": "opcode",
        "width": 4,
        "color": "red"
      },
      {
        "id": "dest",
        "label": "dest",
        "width": 4,
        "color": "yellow"
      },
      {
        "id": "argA",
        "label": "arg A",
        "width": 4,
        "color": "green"
      },
      {
        "width": 4,
        "value": "0000"
      },
      {
        "id": "argB",
        "label": "arg B",
        "width": 4,
        "color": "blue"
      },
      {
        "width": 8,
        "value": "00000000"
      }
    ],
    "variants": [
      {
        "when": {
          "input": "imm",
          "equals": "1"
        },
        "inputs": [
          {
            "width": 1,
            "value": "0"
          },
          {
            "id": "mode",
            "label": "mode",
            "width": 2,
            "color": "orange"
          },
          {
            "id": "imm",
            "label": "imm",
            "width": 1,
            "color": "purple"
          },
          {
            "id": "op",
            "label": "opcode",
            "width": 4,
            "color": "red"
          },
          {
            "id": "dest",
            "label": "dest",
            "width": 4,
            "color": "yellow"
          },
          {
            "id": "argA",
            "label": "arg A",
            "width": 4,
            "color": "green"
          },
          {
            "id": "argB",
            "label": "arg B — value",
            "width": 16,
            "color": "blue"
          }
        ]
      }
    ]
  },
  {
    "id": "symphony-decode",
    "name": "Instruction Decoder",
    "family": "Symphony",
    "description": "Read a word back into its fields. Letters stand for variables, so a pattern like 0000vvvv aaaaaaaa works as well as plain bits.",
    "sendTo": "symphony-instruction",
    "sendLabel": "Edit in encoder →",
    "extractSendable": (res) => sendableFromBits(res, WORD_BITS, wordFields),
    "inputs": [
      {
        "id": "word",
        "placeholder": "01110110 0000vvvv aaaaaaaa aaaaaaaa",
        "format": "bits"
      },
      {
        "id": "read",
        "label": "Read as",
        "kind": "choice",
        "options": [
          {
            "id": "bits",
            "label": "Bits"
          },
          {
            "id": "number",
            "label": "Number"
          },
          {
            "id": "hex",
            "label": "Hex"
          }
        ],
        "value": "bits"
      }
    ]
  }
];
