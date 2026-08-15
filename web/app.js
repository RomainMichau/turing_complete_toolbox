// Renders one card per tool exposed by the Go backend; every keystroke is sent
// to /api/tools/<id> so all the logic stays server side.

async function main() {
  const tools = await (await fetch("api/tools")).json();
  const container = document.getElementById("tools");
  tools.forEach((tool) => container.appendChild(card(tool)));
}

function card(tool) {
  const el = document.createElement("section");
  el.className = "tool";

  const title = document.createElement("h2");
  title.textContent = tool.name;

  const desc = document.createElement("p");
  desc.className = "desc";
  desc.textContent = tool.description;

  const input = document.createElement("input");
  input.type = "text";
  input.spellcheck = false;
  input.autocomplete = "off";
  input.placeholder = tool.placeholder || "";

  const results = document.createElement("div");
  results.className = "results";

  const error = document.createElement("p");
  error.className = "error";
  error.hidden = true;

  let pending = 0;
  input.addEventListener("input", async () => {
    const seq = ++pending;
    const res = await run(tool.id, input.value);
    if (seq !== pending) return; // a newer keystroke already won
    render(results, error, res);
  });

  el.append(title, desc, input, results, error);
  return el;
}

async function run(id, value) {
  try {
    const resp = await fetch("api/tools/" + encodeURIComponent(id), {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ input: value }),
    });
    return await resp.json();
  } catch (e) {
    return { error: String(e) };
  }
}

function render(results, error, res) {
  results.replaceChildren();
  error.hidden = !res.error;
  error.textContent = res.error || "";

  for (const field of res.fields || []) {
    const row = document.createElement("div");
    row.className = "row";

    const label = document.createElement("span");
    label.className = "label";
    label.textContent = field.label;

    const value = document.createElement("span");
    value.className = "value";
    value.textContent = field.value;
    value.title = "Click to copy";
    value.addEventListener("click", () => {
      navigator.clipboard?.writeText(field.value);
      row.classList.add("copied");
      setTimeout(() => row.classList.remove("copied"), 600);
    });

    row.append(label, value);
    results.appendChild(row);
  }
}

main();
