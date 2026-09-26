/**
 * PrismaTec Control Studio (Phase 13)
 * Observes and controls a live Core node via AIP v1 — real data only.
 */

const $ = (s) => document.querySelector(s);
const $$ = (s) => [...document.querySelectorAll(s)];

let es = null;
let selectedId = null;
let organisms = [];
let pulseN = 0;

function base() {
  const v = ($("#baseUrl").value || "").trim();
  return v ? v.replace(/\/$/, "") : "";
}

async function api(path, opts = {}) {
  const res = await fetch(base() + path, {
    ...opts,
    headers: { "Content-Type": "application/json", ...(opts.headers || {}) },
  });
  const data = await res.json().catch(() => ({}));
  if (!res.ok) throw new Error(data.error || res.statusText);
  return data;
}

function setConn(on) {
  const b = $("#connBadge");
  b.textContent = on ? "ONLINE" : "OFFLINE";
  b.className = "badge " + (on ? "on" : "off");
}

function msg(el, text, ok) {
  el.textContent = text || "";
  el.className = "msg " + (ok === true ? "ok" : ok === false ? "err" : "");
}

function setF(name, val) {
  const el = document.querySelector(`[data-f="${name}"]`);
  if (!el) return;
  const v = val == null || val === "" ? "—" : String(val);
  if (el.textContent !== v) el.textContent = v;
}

function renderOrg(o) {
  if (!o) return;
  selectedId = o.id;
  setF("id", o.id);
  setF("root_cid", o.root_cid);
  setF("node_id", o.node_id);
  setF("status", (o.status || "").toUpperCase());
  setF(
    "memory",
    `${o.working_keys || 0} working · ${o.episodes || 0} ep · ${o.semantic_keys || 0} sem`
  );
  setF("capabilities", (o.capabilities || []).join(", ") || "0");
  setF("current_action", o.current_action || "—");
  const dot = $("#stDot");
  dot.className = "dot " + (o.status || "");
  const ul = $("#polList");
  ul.innerHTML = "";
  const policy = o.policy || {};
  const caps = o.capabilities || [];
  const keys = new Set([...caps, ...Object.keys(policy)]);
  for (const k of keys) {
    const allowed = policy[k] !== false;
    const li = document.createElement("li");
    li.className = allowed ? "ok" : "no";
    li.textContent = `${allowed ? "✓" : "✕"} ${k}`;
    ul.appendChild(li);
  }
  $$("#orgList li").forEach((li) => {
    li.classList.toggle("sel", li.dataset.id === o.id);
  });
}

function renderList(list) {
  organisms = list || [];
  const ul = $("#orgList");
  ul.innerHTML = "";
  if (!organisms.length) {
    ul.innerHTML = "<li class='st'>(no organisms)</li>";
    return;
  }
  for (const o of organisms) {
    const li = document.createElement("li");
    li.dataset.id = o.id;
    li.innerHTML = `<div>${o.name || o.id}</div><div class="st">${o.status} · ${o.id}</div>`;
    if (o.id === selectedId) li.classList.add("sel");
    li.onclick = () => {
      selectedId = o.id;
      renderOrg(o);
      refreshSelected();
    };
    ul.appendChild(li);
  }
}

function pushPulse(p) {
  pulseN++;
  $("#pulseCount").textContent = String(pulseN);
  const li = document.createElement("li");
  const t = p.timestamp ? new Date(p.timestamp).toLocaleTimeString() : "";
  li.innerHTML = `<strong>${p.type || "?"}</strong> → ${p.target || "?"} <span>${t}</span>`;
  $("#pulseLog").prepend(li);
  const log = $("#pulseLog");
  while (log.children.length > 100) log.removeChild(log.lastChild);

  if (p.type && p.type.startsWith("organism.")) {
    refreshOrganisms();
  }
  if (selectedId && p.target === selectedId) {
    refreshSelected();
  }
  if (p.type && p.type.startsWith("node.")) {
    refreshInfo();
  }
}

async function refreshInfo() {
  const info = await api("/aip/v1/info");
  $("#nodeInfo").textContent = JSON.stringify(info, null, 2);
  $("#hAip").textContent = info.aip || "v1";
  $("#hOrgs").textContent = String(info.organisms ?? "—");
  $("#hPulse").textContent = String(info.pulse_seq ?? info.pulses ?? "—");
  $("#healthBadge").textContent = `node ${info.status || "—"}`;
  return info;
}

async function refreshOrganisms() {
  const data = await api("/aip/v1/organisms");
  const list = data.organisms || [];
  renderList(list);
  if (!list.length) return;
  if (!selectedId || !list.find((x) => x.id === selectedId)) {
    selectedId = list[0].id;
  }
  const o = list.find((x) => x.id === selectedId) || list[0];
  renderOrg(o);
}

async function refreshSelected() {
  if (!selectedId) return;
  try {
    const o = await api("/aip/v1/organisms/" + encodeURIComponent(selectedId));
    renderOrg(o);
  } catch {
    /* ignore */
  }
}

async function command(action, params) {
  return api("/aip/v1/commands", {
    method: "POST",
    body: JSON.stringify({ aip: "v1", action, params }),
  });
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
  es.onopen = () => setConn(true);
  es.onerror = () => setConn(false);
}

async function connect() {
  try {
    const h = await fetch(base() + "/healthz");
    $("#hHttp").textContent = h.ok ? "ok" : "fail";
    await refreshInfo();
    await refreshOrganisms();
    connectSSE();
    setConn(true);
  } catch (e) {
    $("#hHttp").textContent = "fail";
    setConn(false);
    msg($("#qMsg"), String(e.message || e), false);
  }
}

// Navigation
$$(".nav").forEach((btn) => {
  btn.onclick = () => {
    $$(".nav").forEach((b) => b.classList.remove("active"));
    btn.classList.add("active");
    const v = btn.dataset.view;
    $$(".view").forEach((s) => s.classList.remove("active"));
    $("#view-" + v).classList.add("active");
    $("#viewTitle").textContent =
      { overview: "Overview", organisms: "Organisms", pulse: "Pulse", execution: "Execution", policy: "Policy" }[v] ||
      v;
  };
});

$("#btnConnect").onclick = () => connect();
$("#btnRefresh").onclick = async () => {
  try {
    await refreshInfo();
    await refreshOrganisms();
  } catch (e) {
    msg($("#qMsg"), String(e.message || e), false);
  }
};

$("#btnQuickCreate").onclick = async () => {
  try {
    const name = $("#qName").value.trim() || "agent";
    const res = await command("create", {
      name,
      capabilities: ["memory.read", "memory.write", "inference"],
    });
    if (res.organism) {
      selectedId = res.organism.id;
      renderOrg(res.organism);
    }
    await refreshOrganisms();
    msg($("#qMsg"), "Created " + (res.organism && res.organism.id), true);
  } catch (e) {
    msg($("#qMsg"), String(e.message || e), false);
  }
};

$("#btnStart").onclick = async () => {
  try {
    if (!selectedId) throw new Error("select organism");
    const res = await command("start", { id: selectedId });
    if (res.organism) renderOrg(res.organism);
    msg($("#orgMsg"), "Started", true);
  } catch (e) {
    msg($("#orgMsg"), String(e.message || e), false);
  }
};

$("#btnStop").onclick = async () => {
  try {
    if (!selectedId) throw new Error("select organism");
    const res = await command("stop", { id: selectedId });
    if (res.organism) renderOrg(res.organism);
    msg($("#orgMsg"), "Stopped", true);
  } catch (e) {
    msg($("#orgMsg"), String(e.message || e), false);
  }
};

$("#btnMem").onclick = async () => {
  try {
    if (!selectedId) throw new Error("select organism");
    const res = await command("memory.set", {
      id: selectedId,
      key: $("#memKey").value,
      value: $("#memVal").value,
    });
    if (res.organism) renderOrg(res.organism);
    msg($("#orgMsg"), "Memory updated", true);
  } catch (e) {
    msg($("#orgMsg"), String(e.message || e), false);
  }
};

$("#btnExec").onclick = async () => {
  try {
    if (!selectedId) throw new Error("select organism first");
    const res = await command("execute", {
      id: selectedId,
      entry: $("#exEntry").value || "ping",
      input: $("#exInput").value || "",
    });
    $("#exOut").textContent = JSON.stringify(res, null, 2);
  } catch (e) {
    $("#exOut").textContent = String(e.message || e);
  }
};

$("#btnInfer").onclick = async () => {
  try {
    const res = await command("infer", { prompt: $("#infPrompt").value });
    $("#infOut").textContent = JSON.stringify(res, null, 2);
  } catch (e) {
    $("#infOut").textContent = String(e.message || e);
  }
};

$("#btnPolCheck").onclick = async () => {
  try {
    if (!selectedId) throw new Error("select organism first");
    const res = await command("policy.check", {
      id: selectedId,
      action: $("#polAction").value || "memory.write",
    });
    $("#polOut").textContent = JSON.stringify(res, null, 2);
  } catch (e) {
    $("#polOut").textContent = String(e.message || e);
  }
};

connect().catch(() => {});
