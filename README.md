# Turing Complete Toolbox

A small web toolbox with helpers for playing [Turing Complete](https://turingcomplete.game/).

**→ [romainmichau.github.io/turing_complete_toolbox](https://romainmichau.github.io/turing_complete_toolbox/)**

Everything runs in the browser: there is no backend, no build step and nothing to
install. The site is the `docs/` folder exactly as it sits in the repository.

## Run it

```sh
npm run serve        # http://localhost:8080
```

Any static server will do — `npx serve docs`, or whatever you already have. You
do need one: the page is made of ES modules, and browsers refuse to load those
over `file://`, so opening `index.html` by double clicking it will show a blank
page.

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
| Instruction Doc | The word layout and every opcode, one fold per mode: IO, ALU, JUMP and RAM. |
| Instruction Encoder | One colour coded box per instruction field, all on a single line, and the integer to feed the machine underneath. |
| Instruction Decoder | The way back: a word, or a pattern with letters standing for variables, read out field by field. |

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

Everything is documented except ALU opcodes 11-15.

## How it is put together

```
docs/
  index.html      the page, and the import map naming the vendored modules
  registry.js     what every tool looks like: its inputs, and its doc sections
  tools/          what every tool computes
  components/     how it is drawn, in Preact
  lib/            preact, preact/hooks and htm, vendored
test/             node --test, no dependencies
```

`registry.js` is data: the inputs a tool shows, the colours its fields wear and
the reference sections it prints. `tools/index.js` maps a tool id to the
function that answers it. A tool with no entry there is pure documentation.

The UI is [Preact](https://preactjs.com/) with
[htm](https://github.com/developit/htm), so the templates are tagged template
literals and there is nothing to transpile — the vendored modules are loaded by
name through the import map in `index.html`.

One part is deliberately not declarative. `components/segments-dom.js` builds
the segmented word with plain DOM, because every keystroke in it is about the
caret — where it sits, which box it moves to next, where it lands again after
the immediate flag changes the row's shape — and a caret is not state worth
handing to a renderer. It owns its subtree and reports values back up.

## Adding a tool

Add a descriptor to `docs/registry.js`:

```js
{
  id: "my-tool",
  name: "My Tool",
  family: "General",
  description: "What it does.",
  inputs: [{ id: "bits", placeholder: "1010", format: "bits" }],
}
```

and the function answering it to `docs/tools/index.js`:

```js
export const RUNNERS = {
  "my-tool": (inputs) => [{ label: "Result", value: inputs.bits }],
};
```

It shows up as a card on its own — nothing to change in the components. Give
`inputs` several entries with a `width` to get the segmented, colour coded row
instead of a single free input; an entry without an `id` is filler the user
cannot edit. Throw from the function to put a message under the card.

## Test

```sh
npm test
```

`test/golden.json` holds 434 recorded cases. This toolbox began as a Go program
that computed its answers on the server, and the golden file is what that Go
returned for every one of them — captured before it was removed, and matched
byte for byte by the port that replaced it. It stays the spec. The Go itself is
in the history, at `6c8b634`.

## Hosting

The site is served from `docs/` on `main` by GitHub Pages — Settings → Pages →
"Deploy from a branch", branch `main`, folder `/docs`. There is nothing to
build, so a push is a deploy.
