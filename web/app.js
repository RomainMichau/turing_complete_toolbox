// Renders one collapsible card per tool exposed by the Go backend, grouped by
// family. Every keystroke is sent to /api/tools/<id> so all the logic stays
// server side.

const COLORS = ["orange", "red", "yellow", "green", "blue", "purple", "muted"];
const STORE_KEY = "tct.collapsed";

async function main() {
  const tools = await (await fetch("api/tools")).json();
  const container = document.getElementById("tools");
  loadFolds();

  let family = null;
  tools.forEach((tool, i) => {
    if (tool.family !== family) {
      family = tool.family;
      container.append(familyHeading(family));
    }
    container.append(card(tool, i > 0));
  });
}

function familyHeading(name) {
  const h = document.createElement("h2");
  h.className = "family";
  h.textContent = name;
  return h;
}

// --- folds ------------------------------------------------------------------
// Cards and the groups inside them fold the same way: everything but the first
// one starts closed, and the user's own choices win and are kept in the browser
// under a key of its own.

let folds = {};

function loadFolds() {
  try {
    folds = JSON.parse(localStorage.getItem(STORE_KEY)) || {};
  } catch {
    folds = {};
  }
}

function saveFolds() {
  try {
    localStorage.setItem(STORE_KEY, JSON.stringify(folds));
  } catch {
    /* private mode, storage full: the UI still works, it just forgets */
  }
}

// foldable wires a head button to a body, remembering the choice under key.
function foldable(el, head, body, key, closedByDefault) {
  const setOpen = (open) => {
    head.setAttribute("aria-expanded", String(open));
    body.hidden = !open;
    el.classList.toggle("folded", !open);
  };
  setOpen(!(key in folds ? Boolean(folds[key]) : closedByDefault));

  head.addEventListener("click", () => {
    const open = body.hidden;
    setOpen(open);
    folds[key] = !open;
    saveFolds();
  });
}

function chevron() {
  const el = document.createElement("span");
  el.className = "chevron";
  el.textContent = "▸";
  return el;
}

// --- cards ------------------------------------------------------------------

function card(tool, closedByDefault) {
  const el = document.createElement("section");
  el.className = "tool";

  const head = document.createElement("button");
  head.type = "button";
  head.className = "tool-head";
  head.append(chevron(), text("span", "tool-title", tool.name));

  const body = document.createElement("div");
  body.className = "tool-body";
  body.append(text("p", "desc", tool.description));

  const results = document.createElement("div");
  results.className = "results";
  const error = text("p", "error", "");
  error.hidden = true;
  const hint = text("p", "hint", "Click a value to copy it.");
  hint.hidden = true;

  let pending = 0;
  const update = async (inputs) => {
    const seq = ++pending;
    const res = await run(tool.id, inputs());
    if (seq !== pending) return; // a newer keystroke already won
    render(results, error, hint, res);
  };

  const inputs = tool.inputs || []; // documentation cards have none
  // Only a word cut into fixed width fields gets the segmented row; several
  // plain inputs are just a form.
  const segmented = inputs.some((i) => i.width);
  if (tool.doc) {
    el.classList.add("wide");
    body.append(docBlock(tool.doc, tool.id));
  } else if (segmented) {
    el.classList.add("wide");
    body.append(segmentedField(tool, update)); // all zeroes is already a valid word
  } else {
    body.append(inputForm(inputs, update));
  }
  if (!tool.doc) body.append(results, error, hint);

  foldable(el, head, body, tool.id, closedByDefault);
  el.append(head, body);
  return el;
}

function text(tag, className, content) {
  const el = document.createElement(tag);
  el.className = className;
  el.textContent = content;
  return el;
}

// --- documentation ----------------------------------------------------------
// Reference cards have no input: the server hands over sections and the colours
// to paint them with, matching the field colours of the encoders.

function docBlock(sections, path) {
  const el = document.createElement("div");
  el.className = "doc";
  let groups = 0;
  for (const section of sections) {
    if (section.kind === "group") {
      el.append(docGroup(section, path, groups++ > 0));
      continue;
    }
    if (section.title) el.append(text("p", "doc-title", section.title));
    el.append(docSection(section));
  }
  return el;
}

// docGroup folds a run of sections away behind its own heading, remembered
// under the card it belongs to.
function docGroup(section, path, closedByDefault) {
  const el = document.createElement("section");
  el.className = "doc-group";

  const head = document.createElement("button");
  head.type = "button";
  head.className = "doc-group-head";
  head.append(chevron(), text("span", "doc-group-title", section.title));

  const body = document.createElement("div");
  body.className = "doc-group-body";
  if (section.text) body.append(text("p", "desc", section.text));
  body.append(docBlock(section.sections || [], `${path}/${section.title}`));

  foldable(el, head, body, `${path}/${section.title}`, closedByDefault);
  el.append(head, body);
  return el;
}

function docSection(section) {
  switch (section.kind) {
    case "layout": {
      const line = document.createElement("p");
      line.className = "doc-layout";
      for (const ch of section.text) {
        const span = text("span", "", ch);
        const color = (section.colors || {})[ch];
        if (color) span.style.color = colorVar(color);
        line.append(span);
      }
      return line;
    }
    case "legend":
    case "table":
      return section.compact ? compactTable(section) : docTable(section);
    default:
      return text("p", "doc-note", section.text);
  }
}

function docTable(section) {
  const list = document.createElement("dl");
  list.className = section.kind === "legend" ? "doc-legend" : "doc-table";
  for (const row of section.rows || []) {
    const key = text("dt", "doc-key", row.key);
    if (row.color) key.style.color = colorVar(row.color);
    list.append(key, text("dd", "doc-value", row.value));
  }
  return list;
}

// compactTable flows a long list over as many columns as fit. The column width
// follows the longest entry, so short opcode names pack tight while spelled out
// jump conditions still get a readable column.
function compactTable(section) {
  const rows = section.rows || [];
  const el = document.createElement("div");
  el.className = "doc-compact";

  const longest = rows.reduce((n, r) => Math.max(n, r.key.length + r.value.length), 0);
  el.style.columnWidth = `${Math.max(7, longest * 0.52).toFixed(1)}rem`;

  for (const row of rows) {
    const entry = document.createElement("p");
    entry.className = "doc-entry";
    const key = text("span", "doc-key", row.key);
    if (row.color) key.style.color = colorVar(row.color);
    entry.append(key, text("span", "doc-value", row.value));
    el.append(entry);
  }
  return el;
}

function colorVar(name) {
  return `var(--c-${COLORS.includes(name) ? name : "muted"})`;
}

// --- plain inputs -----------------------------------------------------------
// A tool without a segmented word gets a small form instead: a text box, and
// any number of choices to go with it.

function inputForm(inputs, update) {
  const el = document.createElement("div");
  el.className = "form";

  const readers = [];
  const changed = () => update(() => Object.assign({}, ...readers.map((read) => read())));

  for (const spec of inputs) {
    const field = spec.kind === "choice" ? choiceInput(spec, changed) : freeInput(spec, changed);
    readers.push(field.value);
    el.append(field.el);
  }
  changed(); // defaults alone are already worth an answer
  return el;
}

function freeInput(spec, changed) {
  const input = document.createElement("input");
  input.type = "text";
  input.spellcheck = false;
  input.autocomplete = "off";
  input.placeholder = spec.placeholder || "";

  const value = () => ({ [spec.id]: input.value });
  if (spec.format !== "bits") {
    input.addEventListener("input", changed);
    return { el: input, value };
  }

  const field = bitField(input);
  input.addEventListener("input", () => {
    field.refresh();
    changed();
  });
  return { el: field.el, value };
}

// choiceInput is a row of buttons acting as radio buttons: one is always on,
// and clicking another swaps it.
function choiceInput(spec, changed) {
  const el = document.createElement("div");
  el.className = "choice";
  el.setAttribute("role", "radiogroup");
  if (spec.label) {
    el.append(text("span", "choice-label", spec.label));
    el.setAttribute("aria-label", spec.label);
  }

  const options = spec.options || [];
  let picked = spec.value || (options[0] || {}).id || "";

  const buttons = options.map((option) => {
    const button = document.createElement("button");
    button.type = "button";
    button.className = "choice-option";
    button.textContent = option.label;
    button.setAttribute("role", "radio");
    button.addEventListener("click", () => {
      if (picked === option.id) return;
      picked = option.id;
      buttons.forEach((b) => b.setAttribute("aria-checked", String(b === button)));
      changed();
    });
    button.setAttribute("aria-checked", String(option.id === picked));
    el.append(button);
    return button;
  });

  return { el, value: () => ({ [spec.id]: picked }) };
}

// --- byte colouring ---------------------------------------------------------
// No separator characters: the bits are typed as one run and consecutive bytes
// alternate colour instead. A layer of coloured spans is drawn on top of a
// transparent input, so the native caret, selection and scrolling are kept.

const SEPARATOR = /[\s_|]/g;
const ZERO_WIDTH_SPACE = "\u200B";

// bitField wraps the input in the colouring layer. Returns the element to put
// in the card plus the refresh hook to call on every keystroke.
function bitField(input) {
  const el = document.createElement("div");
  el.className = "bitfield";

  const layer = document.createElement("div");
  layer.className = "bits-layer";
  layer.setAttribute("aria-hidden", "true");

  el.append(input, layer);
  input.addEventListener("scroll", () => (layer.scrollLeft = input.scrollLeft));

  const refresh = () => {
    stripSeparators(input);
    paintBytes(layer, input.value);
    layer.scrollLeft = input.scrollLeft;
  };
  return { el, refresh };
}

// stripSeparators removes anything a paste may have brought in between the
// bits, keeping the caret on the same bit. Characters that are neither bits nor
// separators are left in place so the server can point at them.
function stripSeparators(input) {
  const raw = input.value;
  const clean = raw.replace(SEPARATOR, "");
  if (clean === raw) return;
  const caret = raw.slice(0, input.selectionStart).replace(SEPARATOR, "").length;
  input.value = clean;
  input.setSelectionRange(caret, caret);
}

// paintBytes re-draws the coloured copy of the text, one span per byte, cut
// from the right so the low byte is always a whole one. With breakable set, a
// zero width space goes between the bytes: a long value then wraps on byte
// boundaries only, never inside one. The overlay of an input must not do that,
// or the copy would stop lining up with the text underneath.
function paintBytes(layer, text, breakable = false) {
  layer.replaceChildren();
  // Letters are welcome: a decoder pattern like 0000vvvv is still a word, and
  // reads better with its bytes told apart. Anything else is not a bit string
  // (yet), so show it plain and let the error line do the talking.
  if (/[^01a-zA-Z]/.test(text)) {
    layer.textContent = text;
    return;
  }
  for (let end = text.length, i = 0; end > 0; end -= 8, i++) {
    const span = document.createElement("span");
    span.className = i % 2 ? "byte-odd" : "byte-even";
    span.textContent = text.slice(Math.max(end - 8, 0), end);
    if (breakable && i > 0) layer.prepend(document.createTextNode(ZERO_WIDTH_SPACE));
    layer.prepend(span);
  }
}

// --- segmented input --------------------------------------------------------
// One fixed width box per field of the word, laid out on a single line and
// colour coded. Typing overwrites the bit under the caret and walks forward
// into the next field, so a word can be filled in one go.

// segmentedField keeps the row in step with the layout the values ask for: the
// Symphony immediate flag, say, hands the last two bytes to argument B. Values
// carry over the switch, so what has been typed survives it.
function segmentedField(tool, update) {
  const holder = document.createElement("div");
  const values = {}; // id -> bits, shared by every rebuild
  let layout = null;

  const wanted = () => {
    for (const variant of tool.variants || []) {
      if (values[variant.when.input] === variant.when.equals) return variant.inputs;
    }
    return tool.inputs;
  };

  const build = (focus) => {
    layout = wanted();
    const row = segmentRow(layout, values, changed);
    holder.replaceChildren(row.el);
    if (focus) row.restore(focus);
  };

  // changed runs after every keystroke: relayout if the word changed shape,
  // then hand the values to the server.
  const changed = () => {
    if (wanted() !== layout) build(activeSegment());
    update(() => ({ ...values }));
  };

  build(null);
  update(() => ({ ...values }));
  return holder;
}

// activeSegment remembers where the caret is, to put it back after a rebuild.
function activeSegment() {
  const el = document.activeElement;
  return el && el.dataset && el.dataset.id ? { id: el.dataset.id, caret: el.selectionStart } : null;
}

// fit trims or zero pads a value to a width, keeping the low bits — argument B
// holds the same number whether it is 4 or 16 bits wide.
function fit(bits, width) {
  return bits.length > width ? bits.slice(bits.length - width) : bits.padStart(width, "0");
}

function segmentRow(specs, values, changed) {
  const el = document.createElement("div");
  el.className = "segments";

  const editable = [];
  const ids = [];
  let bits = 0;

  specs.forEach((spec, i) => {
    const cell = document.createElement("label");
    cell.className = spec.id ? "seg" : "seg seg-fixed";
    const color = COLORS.includes(spec.color) ? spec.color : "muted";
    cell.style.setProperty("--seg", `var(--c-${color})`);

    const box = document.createElement("input");
    box.type = "text";
    box.spellcheck = false;
    box.autocomplete = "off";
    box.value = spec.value || "0".repeat(spec.width);
    // 1ch per bit is not enough: the tracking between the digits counts too,
    // and the box still needs its padding on top of that.
    box.style.width = `calc(${spec.width} * (1ch + var(--seg-track)) + 1.5rem)`;
    box.dataset.width = spec.width;
    if (spec.id) {
      box.dataset.id = spec.id;
      if (spec.id in values) box.value = fit(values[spec.id], spec.width);
      values[spec.id] = box.value;
      editable.push(box);
      ids.push(spec.id);
    } else {
      box.readOnly = true;
      box.tabIndex = -1;
    }

    cell.append(box, text("span", "seg-label", spec.label || "×"));
    el.append(cell);

    bits += spec.width;
    if (bits % 8 === 0 && i < specs.length - 1) {
      el.append(text("span", "byte-gap", ""));
    }
  });

  wireSegments(editable, () => {
    ids.forEach((id, i) => (values[id] = editable[i].value));
    changed();
  });

  const restore = ({ id, caret }) => {
    const box = editable[ids.indexOf(id)];
    if (!box) return;
    box.focus();
    const at = Math.min(caret, box.value.length);
    box.setSelectionRange(at, at);
  };
  return { el, restore };
}

function wireSegments(boxes, changed) {
  const width = (box) => Number(box.dataset.width);

  const focusAt = (i, pos) => {
    const box = boxes[i];
    box.focus();
    box.setSelectionRange(pos, pos);
  };

  const write = (i, pos, bit) => {
    const box = boxes[i];
    box.value = box.value.slice(0, pos) + bit + box.value.slice(pos + 1);
    changed();
  };

  // typeBit overwrites the bit at pos and moves on, spilling into the next
  // field when this one is full.
  const typeBit = (i, pos, bit) => {
    if (pos >= width(boxes[i])) {
      if (i + 1 >= boxes.length) return;
      return typeBit(i + 1, 0, bit);
    }
    write(i, pos, bit);
    if (pos + 1 < width(boxes[i])) focusAt(i, pos + 1);
    else if (i + 1 < boxes.length) focusAt(i + 1, 0);
    else focusAt(i, pos + 1);
  };

  boxes.forEach((box, i) => {
    box.addEventListener("keydown", (e) => {
      if (e.ctrlKey || e.metaKey || e.altKey) return;
      const pos = box.selectionStart;

      switch (e.key) {
        case "0":
        case "1":
          e.preventDefault();
          typeBit(i, pos, e.key);
          return;
        case "Backspace":
          e.preventDefault();
          if (pos > 0) {
            write(i, pos - 1, "0");
            focusAt(i, pos - 1);
          } else if (i > 0) {
            const prev = width(boxes[i - 1]) - 1;
            write(i - 1, prev, "0");
            focusAt(i - 1, prev);
          }
          return;
        case "Delete":
          e.preventDefault();
          if (pos < width(box)) write(i, pos, "0");
          return;
        case "ArrowLeft":
          if (pos === 0 && i > 0) {
            e.preventDefault();
            focusAt(i - 1, width(boxes[i - 1]));
          }
          return;
        case "ArrowRight":
          if (pos === width(box) && i + 1 < boxes.length) {
            e.preventDefault();
            focusAt(i + 1, 0);
          }
          return;
      }
      if (e.key.length === 1) e.preventDefault(); // bits only
    });

    // Pasting a run of bits fills this field and the ones after it.
    box.addEventListener("paste", (e) => {
      e.preventDefault();
      const bits = (e.clipboardData || window.clipboardData).getData("text").replace(/[^01]/g, "");
      let at = i;
      let pos = box.selectionStart;
      for (const bit of bits) {
        if (pos >= width(boxes[at])) {
          if (++at >= boxes.length) break;
          pos = 0;
        }
        write(at, pos++, bit);
      }
      focusAt(at, Math.min(pos, width(boxes[at])));
    });
  });
}

// --- results ----------------------------------------------------------------

async function run(id, inputs) {
  try {
    const resp = await fetch("api/tools/" + encodeURIComponent(id), {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ inputs }),
    });
    return await resp.json();
  } catch (e) {
    return { error: String(e) };
  }
}

function render(results, error, hint, res) {
  results.replaceChildren();
  error.hidden = !res.error;
  error.textContent = res.error || "";
  hint.hidden = !(res.fields || []).length;

  for (const field of res.fields || []) {
    const row = document.createElement("div");
    row.className = "row";

    const value = text("span", "value", field.value);
    if (field.format === "bits") {
      // Same byte cycle the bit inputs use, so a result reads like one.
      value.classList.add("bits-value");
      paintBytes(value, field.value, true);
    }
    value.title = "Click to copy";
    value.addEventListener("click", () => {
      navigator.clipboard?.writeText(field.value);
      row.classList.add("copied");
      setTimeout(() => row.classList.remove("copied"), 900);
    });

    row.append(text("span", "label", field.label), value);
    results.appendChild(row);
  }
}

main();
