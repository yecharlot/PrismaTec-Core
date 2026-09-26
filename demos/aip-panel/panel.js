/**
 * PrismaTec Core — AIP Panel (Phase 6 / Demo 1)
 * Consumes real Core data via AIP v1. No mock organism state.
 *
 * PIN-style updates: only the fields that change are written into the DOM
 * (localized resonance idea from Alset-JS, minimal implementation).
 */

const $ = (sel) => document.querySelector(sel);

let es = null;
let selectedId = null;
let pulseCount = 0;

function base() {
  const v = ($("#baseUrl").value || "").trim();
  if (v) return v.replace(/\/$/, "");
  // same-origin when served from Core /demo/
  return "";
}

function setConn(on) {
  const el = $("#connStatus");
  el.textContent = on ? "ONLINE" : "OFFLINE";
  el.className = "badge " + (on ? "on" : "off");
  $("#btnConnect").disabled = on;
  $("#btnDisconnect").disabled = !on;
}

function msg(text, ok) {
  const el = $("#actionMsg");
  el.textContent = text || "";
  el.className = "msg " + (ok === true ? "ok" : ok === false ? "err" : "");
}

async function api(path, opts) {
  const res = await fetch(base() + path, {
    ...opts,
    headers: {
      "Content-Type": "application/json",
      ...(opts && opts.headers),
    },
  });
  const data = await res.json().catch(() => ({}));
  if (!res.ok) {
    throw new Error(data.error || res.statusText || "request failed");
  }
  return data;
}

function setField(name, value) {
  const el = document.querySelector(`[data-field="${name}"]`);
  if (!el) return;
  const v = value == null || value === "" ? "—" : String(value);
  if (el.textContent !== v) el.textContent = v; // localized: only if changed
}

function renderOrganism(o) {
  if (!o) return;
  selectedId = o.id;
  setField("id", o.id);
  setField("root_cid", o.root_cid);
  setField("node_id", o.node_id);
  setField("status", (o.status || "").toUpperCase());
  setField(
    "memory",
    `${o.working_keys || 0} working · ${o.episodes || 0} episodes · ${o.semantic_keys || 0} semantic`
  );
  setField("capabilities", (o.capabilities || []).length);
  setField("current_action", o.current_action || "—");

  const dot = $("#statusDot");
  dot.className = "dot " + (o.status || "");

  const ul = $("#policyList");
  ul.innerHTML = "";
  const policy = o.policy || {};
  const caps = o.capabilities || [];
  const keys = new Set([...caps, ...Object.keys(policy)]);
  for (const k of keys) {
    const allowed = policy[k] !== false && (caps.includes(k) ? policy[k] !== false : policy[k] === true);
    const li = document.createElement("li");
    li.className = allowed ? "ok" : "no";
    li.textContent = `${allowed ? "✓" : "✕"} ${k}`;
    ul.appendChild(li);
  }
}

function renderNode(info) {
  $("#nodeInfo").textContent = JSON.stringify(info, null, 2);
}

function pushPulse(p) {
  pulseCount += 1;
  $("#pulseCount").textContent = String(pulseCount);
  const ul = $("#pulseLog");
  const li = document.createElement("li");
  const t = p.timestamp ? new Date(p.timestamp).toLocaleTimeString() : "";
  li.innerHTML = `<strong>${p.type || "?"}</strong> → ${p.target || "?"} <span>${t}</span>`;
  ul.prepend(li);
  while (ul.children.length > 80) ul.removeChild(ul.lastChild);

  // Localized resonance: if pulse targets selected organism, refresh that PIN only
  if (selectedId && p.target === selectedId) {
    refreshSelectedOrganism();
  }
  if (p.type && p.type.startsWith("node.")) {
    refreshInfo();
  }
  if (p.type === "organism.created" && p.data && p.data.id) {
    selectedId = p.data.id;
    refreshSelectedOrganism();
    refreshInfo();
  }
}

async function refreshInfo() {
  const info = await api("/aip/v1/info");
  renderNode(info);
}

async function refreshOrganisms() {
  const data = await api("/aip/v1/organisms");
  const list = data.organisms || [];
  if (!list.length) return;
  if (!selectedId || !list.find((o) => o.id === selectedId)) {
    selectedId = list[0].id;
  }
  const o = list.find((x) => x.id === selectedId) || list[0];
  renderOrganism(o);
}

async function refreshSelectedOrganism() {
  if (!selectedId) return;
  try {
    const o = await api("/aip/v1/organisms/" + encodeURIComponent(selectedId));
    renderOrganism(o);
  } catch {
    /* ignore if deleted */
  }
}

function connectSSE() {
  if (es) es.close();
  es = new EventSource(base() + "/aip/v1/pulse");
  es.addEventListener("aip.hello", () => setConn(true));
  es.addEventListener("pulse", (ev) => {
    try {
      pushPulse(JSON.parse(ev.data));
    } catch {
      /* ignore */
    }
  });
  es.onerror = () => {
    setConn(false);
  };
  es.onopen = () => setConn(true);
}

function disconnect() {
  if (es) {
    es.close();
    es = null;
  }
  setConn(false);
}

async function connect() {
  msg("Conectando…");
  try {
    await refreshInfo();
    await refreshOrganisms();
    connectSSE();
    msg("Conectado a AIP v1", true);
  } catch (e) {
    msg(String(e.message || e), false);
    setConn(false);
  }
}

async function command(action, params) {
  const data = await api("/aip/v1/commands", {
    method: "POST",
    body: JSON.stringify({ aip: "v1", action, params }),
  });
  if (data.organism) {
    renderOrganism(data.organism);
    selectedId = data.organism.id;
  }
  await refreshInfo();
  return data;
}

$("#btnConnect").onclick = () => connect();
$("#btnDisconnect").onclick = () => {
  disconnect();
  msg("Desconectado");
};

$("#btnCreate").onclick = async () => {
  try {
    const name = $("#orgName").value.trim() || "research-agent-01";
    await command("create", {
      name,
      capabilities: ["memory.read", "inference"],
    });
    msg("Organism created", true);
  } catch (e) {
    msg(String(e.message || e), false);
  }
};

$("#btnStart").onclick = async () => {
  try {
    if (!selectedId) throw new Error("no organism selected");
    await command("start", { id: selectedId });
    msg("Started", true);
  } catch (e) {
    msg(String(e.message || e), false);
  }
};

$("#btnStop").onclick = async () => {
  try {
    if (!selectedId) throw new Error("no organism selected");
    await command("stop", { id: selectedId });
    msg("Stopped", true);
  } catch (e) {
    msg(String(e.message || e), false);
  }
};

$("#btnMemory").onclick = async () => {
  try {
    if (!selectedId) throw new Error("no organism selected");
    await command("memory.set", {
      id: selectedId,
      key: $("#memKey").value || "focus",
      value: $("#memVal").value || "",
    });
    msg("Memory updated", true);
  } catch (e) {
    msg(String(e.message || e), false);
  }
};

// Auto-connect if core is already up
connect().catch(() => {});
