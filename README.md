# Turing Complete Toolbox

A tiny local web app with helpers for playing [Turing Complete](https://turingcomplete.game/).
All the logic lives in Go; the page is a thin client that posts each keystroke to the backend.

## Install

```sh
go install .        # -> $GOPATH/bin/turing_complete_toolbox (already on PATH)
```

The web assets are embedded, so the installed binary is self contained and can be
run from anywhere.

## Run

```sh
turing_complete_toolbox             # serves on :8080 and opens a browser tab
turing_complete_toolbox -addr :9000
turing_complete_toolbox -open=false # no browser tab

go run .                            # from the source tree, without installing
```

## Tools

Tools are grouped by family. Every card folds, only the first one is open on a
fresh browser, and your own folding choices are remembered in local storage.
Click any result value to copy it.

### General

| Tool | What it does |
| --- | --- |
| Number Converter | One box for every base. Type a value, read it back as decimal, hex, octal and bits at once — no direction to pick. |

The converter guesses how to read what you type: a `0x`, `0b` or `0o` prefix
settles it, and a string of nothing but `0` and `1` is taken as bits. The **Read
as** buttons override the guess, and the answer always says which reading it
used. Separators (spaces, `_`, `|`) are ignored, negatives come out in two's
complement, and the bit rows alternate colour byte by byte — wrapping only on a
byte boundary, never inside one. Bits are taken at the width they were typed,
anything else at the narrowest whole number of bytes; a narrower value also gets
spelled out over a full 32 bit word.

### Symphony

| Tool | What it does |
| --- | --- |
| Symphony Word | The word layout, what each field means, the IMM variant, and the four modes. |
| Symphony IO | Mode 0 opcodes, and how the keyboard's two pins are merged. |
| Symphony ALU | Mode 1 opcodes and the flags CMP writes. |
| Symphony JUMP | Mode 2 conditions and the mask rule behind them. |
| Symphony RAM | Mode 3 opcodes, program memory and persistent storage. |
| Instruction Encoder | One colour coded box per instruction field, all on a single line, and the integer to feed the machine underneath. |
| Instruction Decoder | The way back: a word, or a pattern with letters standing for variables, read out field by field. |

The reference is split one card per mode so only what you are building stays
open; the folding is remembered per card.

The instruction word is `XMMIOOOO DDDDAAAA XXXXBBBB XXXXXXXX` — mode, immediate
flag, opcode, destination register, argument A, argument B, and unused bits.
When the immediate flag on bit 4 is set the word is read as
`XMM1OOOO DDDDAAAA BBBBBBBB BBBBBBBB` instead: argument B becomes a 16 bit
literal, and the encoder relayouts to match. Every bit starts at 0; typing
overwrites the bit under the caret and walks into the next field, so a whole
word can be typed in one run, and the encoder decodes what it spells out.

The decoder takes bits or a number, and any letter in a bit pattern stands for
a variable — so an encoding line lifted straight out of the game's assembler
definition works as it is:

```
01110110 0000vvvv aaaaaaaa aaaaaaaa
→ RAM · store_32 · 32 bits of reg A (v) → program memory at literal a
```

Fields made only of bits are decoded; fields holding a letter are reported as
that variable, and the spelled out instruction uses the letter in place of the
value.

Everything is documented except ALU opcodes 11-15. To document more, add to the
cards in [internal/tools/symphony_doc.go](internal/tools/symphony_doc.go) and to
`describeSymphony` in
[internal/tools/symphony_instruction.go](internal/tools/symphony_instruction.go)
— a test fails if a documented opcode has no decoding.

## Adding a tool

Drop a new file in `internal/tools/` and register it from `init()`:

```go
func init() {
    Register(&Tool{
        ID:          "my-tool",
        Name:        "My Tool",
        Family:      FamilyGeneral,
        Description: "What it does.",
        Inputs:      []Input{{ID: "bits", Placeholder: "1010", Format: "bits"}},
        Run: func(in map[string]string) ([]Field, error) {
            return []Field{{Label: "Result", Value: in["bits"]}}, nil
        },
    })
}
```

It shows up automatically as a card in the UI — nothing to change in the
frontend. Give `Inputs` several entries with a `Width` to get the segmented,
colour coded row instead of a single free input; an entry without an `ID` is
filler the user cannot edit.

## Test

```sh
go test ./...
```
