"use strict";

const state = {
  offset: 0,
  limit: 100,
  total: 0,
  rows: [],
  selectedSequence: null,
  selectedEvent: null,
  requestVersion: 0,
  live: false,
};

const elements = Object.fromEntries([
  "capture-status", "capture-id", "event-count", "blob-count", "duration", "generator",
  "search", "direction", "packet", "kind", "channel", "problems", "reset",
  "previous", "next", "page-label", "event-rows", "empty-state", "event-details",
  "copy-json", "byte-status", "hex-view", "viewer-status", "follow", "follow-container",
].map((id) => [id, document.getElementById(id)]));

const filterSelectKeys = ["direction", "packet", "kind", "channel"];
const customSelects = new Map();
let searchTimer;
let pollTimer;

function closeCustomSelect(select) {
  const custom = customSelects.get(select);
  if (!custom) return;
  custom.menu.hidden = true;
  custom.trigger.setAttribute("aria-expanded", "false");
}

function closeAllCustomSelects(except) {
  for (const select of customSelects.keys()) {
    if (select !== except) closeCustomSelect(select);
  }
}

function updateCustomSelectActive(select) {
  const custom = customSelects.get(select);
  if (!custom) return;
  for (const option of custom.menu.querySelectorAll(".select-option")) {
    option.classList.toggle("active", Number(option.dataset.index) === custom.activeIndex);
  }
  custom.menu.querySelector(`[data-index="${custom.activeIndex}"]`)?.scrollIntoView({ block: "nearest" });
}

function syncCustomSelect(select) {
  const custom = customSelects.get(select);
  if (!custom) return;
  const options = Array.from(select.options);
  const selectedIndex = Math.max(0, select.selectedIndex);
  custom.activeIndex = selectedIndex;
  custom.trigger.textContent = options[selectedIndex]?.textContent || "";
  custom.menu.replaceChildren();
  for (const [index, option] of options.entries()) {
    const item = document.createElement("div");
    item.className = "select-option";
    item.dataset.index = String(index);
    item.setAttribute("role", "option");
    item.setAttribute("aria-selected", String(index === selectedIndex));
    item.textContent = option.textContent;
    item.addEventListener("pointerdown", (event) => {
      event.preventDefault();
      selectCustomOption(select, index);
    });
    item.addEventListener("mouseenter", () => {
      custom.activeIndex = index;
      updateCustomSelectActive(select);
    });
    custom.menu.append(item);
  }
  updateCustomSelectActive(select);
}

function openCustomSelect(select) {
  const custom = customSelects.get(select);
  if (!custom) return;
  closeAllCustomSelects(select);
  custom.activeIndex = Math.max(0, select.selectedIndex);
  custom.menu.hidden = false;
  custom.trigger.setAttribute("aria-expanded", "true");
  updateCustomSelectActive(select);
}

function selectCustomOption(select, index) {
  const custom = customSelects.get(select);
  const option = select.options[index];
  if (!custom || !option) return;
  select.selectedIndex = index;
  select.dispatchEvent(new Event("change", { bubbles: true }));
  syncCustomSelect(select);
  closeCustomSelect(select);
  custom.trigger.focus();
}

function moveCustomSelectActive(select, delta) {
  const custom = customSelects.get(select);
  if (!custom || !select.options.length) return;
  const count = select.options.length;
  custom.activeIndex = (custom.activeIndex + delta + count) % count;
  updateCustomSelectActive(select);
}

function setupCustomSelects() {
  for (const key of filterSelectKeys) {
    const select = elements[key];
    const label = select.closest("label");
    const labelText = label?.querySelector(":scope > span")?.textContent.trim() || key;
    const wrapper = document.createElement("span");
    wrapper.className = "select-control";
    select.replaceWith(wrapper);
    wrapper.append(select);
    select.classList.add("native-select");
    select.tabIndex = -1;
    select.setAttribute("aria-hidden", "true");

    const trigger = document.createElement("button");
    trigger.type = "button";
    trigger.className = "select-trigger";
    trigger.setAttribute("aria-label", labelText);
    trigger.setAttribute("aria-haspopup", "listbox");
    trigger.setAttribute("aria-expanded", "false");

    const menu = document.createElement("div");
    menu.className = "select-menu";
    menu.id = `${key}-menu`;
    menu.hidden = true;
    menu.setAttribute("role", "listbox");
    trigger.setAttribute("aria-controls", menu.id);
    wrapper.append(trigger, menu);

    customSelects.set(select, { wrapper, trigger, menu, activeIndex: 0 });
    select.addEventListener("change", () => syncCustomSelect(select));
    trigger.addEventListener("click", () => {
      if (menu.hidden) openCustomSelect(select);
      else closeCustomSelect(select);
    });
    trigger.addEventListener("keydown", (event) => {
      if (event.key === "ArrowDown" || event.key === "ArrowUp") {
        event.preventDefault();
        if (menu.hidden) openCustomSelect(select);
        else moveCustomSelectActive(select, event.key === "ArrowDown" ? 1 : -1);
        return;
      }
      if (event.key === "Home" || event.key === "End") {
        if (menu.hidden) return;
        event.preventDefault();
        const custom = customSelects.get(select);
        custom.activeIndex = event.key === "Home" ? 0 : select.options.length - 1;
        updateCustomSelectActive(select);
        return;
      }
      if (event.key === "Enter" || event.key === " ") {
        event.preventDefault();
        if (menu.hidden) openCustomSelect(select);
        else selectCustomOption(select, customSelects.get(select).activeIndex);
        return;
      }
      if (event.key === "Escape" && !menu.hidden) {
        event.preventDefault();
        closeCustomSelect(select);
      }
    });
    syncCustomSelect(select);
  }
  document.addEventListener("pointerdown", (event) => {
    for (const [select, custom] of customSelects) {
      if (!custom.wrapper.contains(event.target)) closeCustomSelect(select);
    }
  });
}

setupCustomSelects();

async function requestJSON(path) {
  const response = await fetch(path, { headers: { Accept: "application/json" } });
  const value = await response.json();
  if (!response.ok) {
    throw new Error(value.error || `Request failed with status ${response.status}`);
  }
  return value;
}

function formatCount(value) {
  return new Intl.NumberFormat().format(value);
}

function formatDuration(nanoseconds) {
  const milliseconds = Number(BigInt(nanoseconds || "0") / 1000000n);
  if (milliseconds < 1000) return `${milliseconds} ms`;
  const seconds = milliseconds / 1000;
  if (seconds < 60) return `${seconds.toFixed(seconds < 10 ? 2 : 1)} s`;
  const minutes = Math.floor(seconds / 60);
  return `${minutes}m ${(seconds % 60).toFixed(1)}s`;
}

function eventClock(value) {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return value || "-";
  return date.toLocaleTimeString([], { hour12: false, hour: "2-digit", minute: "2-digit", second: "2-digit", fractionalSecondDigits: 3 });
}

function flowLabel(direction) {
  if (direction === "client_to_server") return "C → S";
  if (direction === "server_to_client") return "S → C";
  if (direction === "internal") return "LOCAL";
  return "?";
}

function syncOptions(select, values) {
  const selected = select.value;
  while (select.options.length > 1) select.remove(1);
  for (const value of values || []) {
    const option = document.createElement("option");
    option.value = value;
    option.textContent = value;
    select.append(option);
  }
  if ((values || []).includes(selected)) select.value = selected;
  syncCustomSelect(select);
}

async function loadSession() {
  const session = await requestJSON("api/session");
  const manifest = session.manifest;
  state.live = session.mode === "live" && manifest.status === "open";
  const displayStatus = state.live ? "live" : manifest.status === "failed" || session.verification === "failed" ? "failed" : session.verification === "verified" ? "verified" : manifest.status;
  elements["capture-status"].textContent = displayStatus;
  elements["capture-status"].className = `status-pill ${state.live ? "live" : displayStatus === "failed" ? "failed" : ""}`;
  elements["follow-container"].hidden = !state.live;
  elements["capture-id"].textContent = manifest.capture_id;
  elements["event-count"].textContent = formatCount(session.events);
  elements["blob-count"].textContent = formatCount(session.blobs);
  const start = Date.parse(manifest.started_at);
  const end = state.live ? Date.now() : Date.parse(manifest.ended_at);
  elements.duration.textContent = Number.isNaN(start) || Number.isNaN(end) ? "-" : formatDuration(String((end - start) * 1000000));
  const generator = manifest.generator;
  elements.generator.textContent = [generator.name, generator.version].filter(Boolean).join(" ") || "-";
  syncOptions(elements.direction, session.filters.directions);
  syncOptions(elements.packet, session.filters.packets);
  syncOptions(elements.kind, session.filters.kinds);
  syncOptions(elements.channel, session.filters.channels);
  return session;
}

function filterQuery() {
  const query = new URLSearchParams({ offset: String(state.offset), limit: String(state.limit) });
  for (const key of ["search", "direction", "packet", "kind", "channel"]) {
    const value = elements[key].value.trim();
    if (value) query.set(key === "search" ? "q" : key, value);
  }
  if (elements.problems.checked) query.set("problems", "true");
  if (state.live && elements.follow.checked) query.set("tail", "true");
  return query.toString();
}

async function loadEvents({ preserveSelection = true } = {}) {
  const version = ++state.requestVersion;
  elements["viewer-status"].textContent = "Loading events...";
  const result = await requestJSON(`api/events?${filterQuery()}`);
  if (version !== state.requestVersion) return;
  state.offset = result.offset;
  state.total = result.total;
  state.rows = result.rows;
  renderRows();
  updatePager();
  if (!preserveSelection) {
    clearSelection();
  } else {
    markSelectedRow();
  }
  const start = state.total === 0 ? 0 : state.offset + 1;
  const end = Math.min(state.offset + state.rows.length, state.total);
  const prefix = state.live ? "Live" : "Showing";
  elements["viewer-status"].textContent = `${prefix} ${formatCount(start)}-${formatCount(end)} of ${formatCount(state.total)} matching events`;
}

function renderRows() {
  elements["event-rows"].replaceChildren();
  elements["empty-state"].hidden = state.rows.length !== 0;
  for (const row of state.rows) {
    const tr = document.createElement("tr");
    tr.dataset.sequence = String(row.sequence);
    tr.tabIndex = 0;
    if (row.problem) tr.classList.add("problem");

    const sequence = cell(`#${row.sequence}`, "sequence");
    const time = cell(eventClock(row.time), "time");
    const flow = cell(flowLabel(row.direction), `flow ${row.direction || "unknown"}`);
    const eventCell = document.createElement("td");
    const eventName = document.createElement("span");
    eventName.className = "event-name";
    eventName.textContent = row.packet?.name || row.kind;
    const eventKind = document.createElement("span");
    eventKind.className = "event-kind";
    eventKind.textContent = row.packet ? `${row.kind} · 0x${row.packet.id.toString(16).padStart(2, "0")}` : row.channel || row.stage || "event";
    eventCell.append(eventName, eventKind);
    const details = cell(row.summary || "-", row.problem ? "problem-text" : "");
    tr.append(sequence, time, flow, eventCell, details);
    tr.addEventListener("click", () => selectEvent(row.sequence));
    tr.addEventListener("keydown", (event) => {
      if (event.key === "Enter" || event.key === " ") {
        event.preventDefault();
        selectEvent(row.sequence);
      }
    });
    elements["event-rows"].append(tr);
  }
  markSelectedRow();
}

function cell(text, className) {
  const td = document.createElement("td");
  td.textContent = text;
  if (className) td.className = className;
  td.title = text;
  return td;
}

function updatePager() {
  const page = state.total === 0 ? 0 : Math.floor(state.offset / state.limit) + 1;
  const pages = state.total === 0 ? 0 : Math.ceil(state.total / state.limit);
  elements["page-label"].textContent = `${page} / ${pages}`;
  elements.previous.disabled = state.offset === 0;
  elements.next.disabled = state.offset + state.limit >= state.total;
}

function markSelectedRow() {
  for (const tr of elements["event-rows"].querySelectorAll("tr")) {
    tr.classList.toggle("selected", Number(tr.dataset.sequence) === state.selectedSequence);
  }
}

async function selectEvent(sequence) {
  state.selectedSequence = sequence;
  markSelectedRow();
  elements["event-details"].textContent = "Loading event details...";
  elements["hex-view"].textContent = "Loading blob preview...";
  try {
    const result = await requestJSON(`api/events/${sequence}`);
    if (state.selectedSequence !== sequence) return;
    state.selectedEvent = result.event;
    renderEvent(result);
  } catch (error) {
    if (state.selectedSequence !== sequence) return;
    elements["event-details"].textContent = error.message;
    elements["hex-view"].textContent = "Blob preview unavailable.";
    elements["viewer-status"].textContent = error.message;
  }
}

function renderEvent(result) {
  const event = result.event;
  elements["copy-json"].disabled = false;
  const container = document.createDocumentFragment();
  const meta = document.createElement("div");
  meta.className = "meta-grid";
  const fields = [
    ["Sequence", `#${event.sequence}`], ["Elapsed", formatDuration(event.elapsed_nano)],
    ["Direction", flowLabel(event.direction)], ["Kind", event.kind],
    ["Channel", event.channel || "-"], ["Hop", event.hop ? String(event.hop) : "-"],
    ["Packet", event.packet?.name || "-"], ["Decode", event.packet?.decode_status || "-"],
  ];
  for (const [label, value] of fields) {
    const item = document.createElement("div");
    item.className = "meta-item";
    const name = document.createElement("span");
    name.textContent = label;
    const content = document.createElement("strong");
    content.textContent = value;
    content.title = value;
    item.append(name, content);
    meta.append(item);
  }
  container.append(meta);

  const tree = document.createElement("div");
  tree.className = "json-tree";
  const budget = { remaining: 4000, clipped: false };
  tree.append(renderJSON("event", event, budget));
  if (budget.clipped) {
    const note = document.createElement("p");
    note.className = "muted";
    note.textContent = "The decoded tree display stopped after 4,000 nodes. The canonical event remains complete.";
    tree.append(note);
  }
  container.append(tree);
  elements["event-details"].replaceChildren(container);
  renderHex(result.blob_preview, result.blob_preview_truncated, event.blob);
}

function renderJSON(key, value, budget) {
  if (budget.remaining-- <= 0) {
    budget.clipped = true;
    return textValue(key, "display limit reached", "json-null");
  }
  if (value !== null && typeof value === "object") {
    const details = document.createElement("details");
    details.open = key === "event" || key === "data";
    const summary = document.createElement("summary");
    const keySpan = document.createElement("span");
    keySpan.className = "json-key";
    keySpan.textContent = key;
    const count = Array.isArray(value) ? value.length : Object.keys(value).length;
    summary.append(keySpan, document.createTextNode(Array.isArray(value) ? ` [${count}]` : ` {${count}}`));
    details.append(summary);
    const entries = Array.isArray(value) ? value.entries() : Object.entries(value);
    for (const [childKey, childValue] of entries) {
      if (budget.remaining <= 0) {
        budget.clipped = true;
        break;
      }
      details.append(renderJSON(String(childKey), childValue, budget));
    }
    return details;
  }
  const className = value === null ? "json-null" : `json-${typeof value}`;
  const formatted = typeof value === "string" ? JSON.stringify(value) : String(value);
  return textValue(key, formatted, className);
}

function textValue(key, value, className) {
  const line = document.createElement("div");
  line.className = "json-value";
  const keySpan = document.createElement("span");
  keySpan.className = "json-key";
  keySpan.textContent = `${key}: `;
  const valueSpan = document.createElement("span");
  valueSpan.className = className;
  valueSpan.textContent = value;
  line.append(keySpan, valueSpan);
  return line;
}

function renderHex(encoded, truncated, blob) {
  if (!encoded || !blob) {
    elements["byte-status"].textContent = "No blob selected";
    elements["hex-view"].textContent = "This event has no content-addressed blob.";
    return;
  }
  const binary = atob(encoded);
  const bytes = Uint8Array.from(binary, (character) => character.charCodeAt(0));
  const lines = [];
  for (let offset = 0; offset < bytes.length; offset += 16) {
    const slice = bytes.slice(offset, offset + 16);
    const hex = Array.from(slice, (byte) => byte.toString(16).padStart(2, "0")).join(" ").padEnd(47, " ");
    const ascii = Array.from(slice, (byte) => byte >= 32 && byte <= 126 ? String.fromCharCode(byte) : ".").join("");
    lines.push(`${offset.toString(16).padStart(8, "0")}  ${hex}  |${ascii}|`);
  }
  elements["hex-view"].textContent = lines.join("\n") || "Blob is empty.";
  const suffix = truncated ? ` of ${formatCount(blob.size)} bytes` : " bytes";
  elements["byte-status"].textContent = `${formatCount(bytes.length)}${suffix} · ${blob.representation || blob.media_type || "blob"}`;
}

function clearSelection() {
  state.selectedSequence = null;
  state.selectedEvent = null;
  markSelectedRow();
  elements["copy-json"].disabled = true;
  elements["event-details"].innerHTML = '<div class="selection-prompt">Select an event to inspect its metadata and decoded fields.</div>';
  elements["byte-status"].textContent = "No blob selected";
  elements["hex-view"].textContent = "Select an event with a blob reference.";
}

function resetFilters() {
  elements.search.value = "";
  for (const key of filterSelectKeys) {
    elements[key].value = "";
    syncCustomSelect(elements[key]);
  }
  elements.problems.checked = false;
  state.offset = 0;
  loadEvents({ preserveSelection: false }).catch(showFatal);
}

function showFatal(error) {
  elements["viewer-status"].textContent = error.message;
  elements["event-rows"].replaceChildren();
  elements["empty-state"].hidden = false;
  elements["empty-state"].textContent = error.message;
}

function schedulePoll() {
  clearTimeout(pollTimer);
  if (!state.live) return;
  pollTimer = setTimeout(async () => {
    try {
      await loadSession();
      await loadEvents();
    } catch (error) {
      elements["viewer-status"].textContent = `Live refresh failed: ${error.message}`;
    } finally {
      schedulePoll();
    }
  }, 1000);
}

for (const key of ["direction", "packet", "kind", "channel", "problems"]) {
  elements[key].addEventListener("change", () => {
    state.offset = 0;
    loadEvents({ preserveSelection: false }).catch(showFatal);
  });
}

elements.search.addEventListener("input", () => {
  clearTimeout(searchTimer);
  searchTimer = setTimeout(() => {
    state.offset = 0;
    loadEvents({ preserveSelection: false }).catch(showFatal);
  }, 180);
});

elements.reset.addEventListener("click", resetFilters);
elements.previous.addEventListener("click", () => {
  elements.follow.checked = false;
  state.offset = Math.max(0, state.offset - state.limit);
  loadEvents().catch(showFatal);
});
elements.next.addEventListener("click", () => {
  elements.follow.checked = false;
  state.offset += state.limit;
  loadEvents().catch(showFatal);
});
elements.follow.addEventListener("change", () => {
  if (elements.follow.checked) state.offset = 0;
  loadEvents().catch(showFatal);
});
elements["copy-json"].addEventListener("click", async () => {
  if (!state.selectedEvent) return;
  await navigator.clipboard.writeText(JSON.stringify(state.selectedEvent, null, 2));
  elements["viewer-status"].textContent = `Copied event #${state.selectedEvent.sequence} as JSON`;
});

document.addEventListener("keydown", (event) => {
  if ((event.ctrlKey || event.metaKey) && event.key.toLowerCase() === "f") {
    event.preventDefault();
    elements.search.focus();
    elements.search.select();
    return;
  }
  if (event.target.matches("input, select, button")) return;
  if (event.key !== "ArrowDown" && event.key !== "ArrowUp") return;
  const current = state.rows.findIndex((row) => row.sequence === state.selectedSequence);
  const next = event.key === "ArrowDown" ? Math.min(state.rows.length - 1, current + 1) : Math.max(0, current <= 0 ? 0 : current - 1);
  if (state.rows[next]) {
    event.preventDefault();
    selectEvent(state.rows[next].sequence);
    elements["event-rows"].querySelector(`[data-sequence="${state.rows[next].sequence}"]`)?.scrollIntoView({ block: "nearest" });
  }
});

(async () => {
  try {
    await loadSession();
    await loadEvents({ preserveSelection: false });
    schedulePoll();
  } catch (error) {
    showFatal(error);
  }
})();
