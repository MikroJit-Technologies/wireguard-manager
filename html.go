package main

import (
	"encoding/json"
	"html/template"
	"io"
)

var pageTmpl = template.Must(template.New("page").Funcs(template.FuncMap{
	"json": func(v any) (template.JS, error) {
		b, err := json.Marshal(v)
		return template.JS(b), err
	},
}).Parse(pageHTML))

func renderHTML(w io.Writer, data map[string]any) {
	pageTmpl.Execute(w, data)
}

const pageHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8"/>
<meta name="viewport" content="width=device-width,initial-scale=1"/>
<title>WireGuard Manager</title>
<style>
*{box-sizing:border-box;margin:0;padding:0}
:root{
  --bg:#0d1117;--surface:#161b22;--border:#30363d;--border2:#21262d;
  --fg:#c9d1d9;--fg2:#8b949e;--bright:#f0f6fc;
  --green:#3fb950;--blue:#388bfd;--red:#f85149;--orange:#e3b341;
  --green-bg:#1a3622;--blue-bg:#1f3a5f;--red-bg:#2d1b1b;
}
body{font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',sans-serif;background:var(--bg);color:var(--fg);min-height:100vh;font-size:14px}
a{color:var(--blue)}

/* ── Nav ── */
nav{background:var(--surface);border-bottom:1px solid var(--border);padding:0 20px;display:flex;align-items:center;height:52px;gap:12px;position:sticky;top:0;z-index:50}
.nav-brand{font-weight:700;font-size:15px;color:var(--bright);display:flex;align-items:center;gap:8px;white-space:nowrap}
.iface-badge{background:var(--bg);border:1px solid var(--border);border-radius:5px;padding:3px 9px;color:var(--fg2);font-size:12px;font-family:monospace}
.server-key{font-size:11px;color:var(--fg2);display:flex;align-items:center;gap:5px}
.server-key code{font-family:monospace;background:var(--bg);border:1px solid var(--border);border-radius:4px;padding:2px 6px}
.nav-right{margin-left:auto;display:flex;align-items:center;gap:8px}
.refresh-btn{background:none;border:1px solid var(--border);border-radius:6px;padding:5px 10px;color:var(--fg2);cursor:pointer;font-size:13px;display:flex;align-items:center;gap:5px;transition:all .15s}
.refresh-btn:hover{border-color:var(--blue);color:var(--bright)}
.refresh-btn.spinning svg{animation:spin .6s linear infinite}
@keyframes spin{to{transform:rotate(360deg)}}

/* ── Stats bar ── */
.stats-bar{padding:16px 20px;display:flex;gap:12px;flex-wrap:wrap}
.stat{background:var(--surface);border:1px solid var(--border);border-radius:10px;padding:14px 18px;min-width:120px;flex:1;max-width:200px}
.stat-label{font-size:11px;color:var(--fg2);text-transform:uppercase;letter-spacing:.6px;margin-bottom:5px}
.stat-value{font-size:26px;font-weight:700;color:var(--bright);line-height:1}
.stat-value.c-green{color:var(--green)}
.stat-value.c-blue{color:var(--blue)}
.stat-value.c-orange{color:var(--orange)}

/* ── Content ── */
.content{padding:0 20px 40px}
.section-head{display:flex;align-items:center;gap:12px;margin-bottom:14px;flex-wrap:wrap}
.section-title{font-size:15px;font-weight:600;color:var(--bright)}
.peer-count{background:var(--border2);border-radius:10px;padding:1px 7px;font-size:12px;color:var(--fg2)}
.search-wrap{flex:1;max-width:300px;position:relative}
.search-wrap input{width:100%;background:var(--bg);border:1px solid var(--border);border-radius:6px;padding:6px 10px 6px 30px;color:var(--fg);font-size:13px;outline:none}
.search-wrap input:focus{border-color:var(--blue)}
.search-wrap svg{position:absolute;left:8px;top:50%;transform:translateY(-50%);color:var(--fg2);pointer-events:none}
.ml-auto{margin-left:auto}

/* ── Table ── */
.table-wrap{background:var(--surface);border:1px solid var(--border);border-radius:10px;overflow:hidden}
table{width:100%;border-collapse:collapse}
th{text-align:left;padding:10px 14px;font-size:11px;font-weight:600;color:var(--fg2);text-transform:uppercase;letter-spacing:.5px;border-bottom:1px solid var(--border2);white-space:nowrap}
td{padding:12px 14px;border-bottom:1px solid var(--border2);vertical-align:middle}
tr:last-child td{border-bottom:none}
tbody tr:hover td{background:#1a2030}
.peer-name{font-weight:500;color:var(--bright);margin-bottom:2px}
.peer-ip{font-size:12px;color:var(--fg2);font-family:monospace}
.key-wrap{display:flex;align-items:center;gap:5px}
.key-mono{font-family:monospace;font-size:12px;color:var(--fg2);background:var(--bg);padding:2px 6px;border-radius:4px}
.copy-btn{background:none;border:none;cursor:pointer;color:var(--fg2);padding:2px;border-radius:3px;display:flex;align-items:center;transition:color .15s}
.copy-btn:hover{color:var(--bright)}
.copy-btn.copied{color:var(--green)}
.traffic-up{color:var(--fg2);font-size:12px}
.traffic-up span{display:inline-block;width:12px;text-align:center;margin-right:3px;color:var(--fg2)}
.badge{display:inline-flex;align-items:center;gap:5px;padding:3px 8px;border-radius:10px;font-size:12px;font-weight:500;white-space:nowrap}
.badge-online{background:var(--green-bg);color:var(--green)}
.badge-offline{background:var(--red-bg);color:var(--red)}
.dot{width:6px;height:6px;border-radius:50%;flex-shrink:0}
.dot-green{background:var(--green)}
.dot-red{background:var(--red)}
.actions{display:flex;gap:6px;align-items:center}
.empty-cell{text-align:center;padding:48px 20px}
.empty-icon{font-size:36px;margin-bottom:10px;opacity:.4}
.empty-title{color:var(--bright);margin-bottom:4px;font-size:15px}
.empty-sub{color:var(--fg2);font-size:13px}

/* ── Buttons ── */
.btn{display:inline-flex;align-items:center;gap:6px;padding:6px 12px;border-radius:6px;font-size:13px;font-weight:500;cursor:pointer;border:none;white-space:nowrap;transition:filter .15s}
.btn:hover{filter:brightness(1.15)}
.btn:disabled{opacity:.5;cursor:default;filter:none}
.btn-primary{background:#238636;color:#fff}
.btn-blue{background:#1f6feb;color:#fff}
.btn-ghost{background:transparent;border:1px solid var(--border);color:var(--fg)}
.btn-ghost:hover{border-color:var(--blue);color:var(--bright)}
.btn-danger{background:transparent;border:1px solid var(--border);color:var(--red)}
.btn-danger:hover{background:var(--red-bg);border-color:var(--red)}
.btn-confirm{background:#da3633;color:#fff;border:none}
.btn-sm{padding:4px 9px;font-size:12px}

/* ── Modal ── */
.overlay{display:none;position:fixed;inset:0;background:rgba(0,0,0,.75);z-index:200;align-items:flex-start;justify-content:center;padding:60px 16px;overflow-y:auto}
.overlay.open{display:flex}
.modal{background:var(--surface);border:1px solid var(--border);border-radius:12px;width:100%;max-width:500px;padding:22px;position:relative;margin:auto}
.modal-lg{max-width:560px}
.modal h3{font-size:16px;font-weight:600;color:var(--bright);margin-bottom:18px}
.modal-close{position:absolute;top:16px;right:16px;background:none;border:none;color:var(--fg2);cursor:pointer;padding:2px;border-radius:4px;line-height:1}
.modal-close:hover{color:var(--bright)}
.form-group{margin-bottom:14px}
.form-label{display:block;font-size:12px;font-weight:500;color:var(--fg2);margin-bottom:5px;text-transform:uppercase;letter-spacing:.4px}
.form-input{width:100%;background:var(--bg);border:1px solid var(--border);border-radius:6px;padding:8px 10px;color:var(--fg);font-size:13px;outline:none;font-family:inherit}
.form-input:focus{border-color:var(--blue)}
.input-row{display:flex;gap:7px}
.input-row .form-input{flex:1;min-width:0}
.form-hint{font-size:11px;color:var(--fg2);margin-top:4px}
.modal-footer{display:flex;justify-content:flex-end;gap:8px;margin-top:18px;padding-top:16px;border-top:1px solid var(--border2)}
.config-preview{display:none;margin-top:14px}
.config-preview .label-row{display:flex;align-items:center;justify-content:space-between;margin-bottom:6px}
.config-box{background:var(--bg);border:1px solid var(--border);border-radius:6px;padding:10px 12px;font-family:monospace;font-size:11px;color:var(--fg2);white-space:pre;overflow-x:auto;max-height:150px;overflow-y:auto;line-height:1.5}
.generated-tag{display:inline-block;background:var(--green-bg);color:var(--green);font-size:11px;padding:2px 7px;border-radius:8px;margin-left:8px}

/* ── QR Modal ── */
.qr-body{display:flex;flex-direction:column;align-items:center;gap:14px}
.qr-img-wrap{background:#fff;border-radius:10px;padding:10px;display:flex}
.qr-img-wrap img{width:220px;height:220px;display:block}
.qr-placeholder{width:220px;height:220px;display:flex;align-items:center;justify-content:center;color:#999;font-size:13px;text-align:center;padding:20px}
.qr-config{width:100%}
.qr-config .config-box{max-height:160px}

/* ── Toast ── */
#toast-container{position:fixed;top:16px;right:16px;z-index:999;display:flex;flex-direction:column;gap:8px;pointer-events:none}
.toast{background:var(--surface);border:1px solid var(--border);border-radius:8px;padding:10px 14px;font-size:13px;display:flex;align-items:center;gap:8px;min-width:240px;max-width:340px;pointer-events:auto;animation:slideIn .2s ease;box-shadow:0 4px 16px rgba(0,0,0,.4)}
.toast.success{border-color:var(--green);background:#0d2318}
.toast.error{border-color:var(--red);background:var(--red-bg)}
.toast.info{border-color:var(--blue);background:var(--blue-bg)}
.toast-icon{width:16px;height:16px;flex-shrink:0}
.toast-msg{flex:1;color:var(--bright)}
.toast-close{background:none;border:none;color:var(--fg2);cursor:pointer;font-size:16px;line-height:1;padding:0 2px}
@keyframes slideIn{from{transform:translateX(20px);opacity:0}to{transform:none;opacity:1}}
@keyframes slideOut{to{transform:translateX(20px);opacity:0}}

/* ── Loading ── */
.spinner{display:inline-block;width:14px;height:14px;border:2px solid var(--border);border-top-color:var(--blue);border-radius:50%;animation:spin .7s linear infinite}

@media(max-width:700px){
  .stats-bar{gap:8px}
  .stat{min-width:calc(50% - 4px);max-width:none}
  th:nth-child(3),td:nth-child(3),th:nth-child(4),td:nth-child(4){display:none}
  nav .server-key{display:none}
}
</style>
</head>
<body>

<nav>
  <div class="nav-brand">
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="var(--blue)" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
      <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/>
    </svg>
    wireguard-manager
  </div>
  <span class="iface-badge" id="nav-iface">{{.Interface}}</span>
  <span class="server-key">
    Server&nbsp;
    <code id="nav-server-pub">{{.ServerPub}}</code>
  </span>
  <div class="nav-right">
    <button class="refresh-btn" id="refresh-btn" onclick="loadPeers(true)" title="Refresh peers">
      <svg width="14" height="14" viewBox="0 0 16 16" fill="currentColor">
        <path d="M8 3a5 5 0 104.546 2.914.5.5 0 01.908-.417A6 6 0 118 2v1z"/>
        <path d="M8 4.466V.534a.25.25 0 01.41-.192l2.36 1.966c.12.1.12.284 0 .384L8.41 4.658A.25.25 0 018 4.466z"/>
      </svg>
      Refresh
    </button>
  </div>
</nav>

<div class="stats-bar" id="stats-bar">
  <div class="stat"><div class="stat-label">Total</div><div class="stat-value c-blue" id="stat-total">—</div></div>
  <div class="stat"><div class="stat-label">Online</div><div class="stat-value c-green" id="stat-online">—</div></div>
  <div class="stat"><div class="stat-label">Offline</div><div class="stat-value" id="stat-offline">—</div></div>
  <div class="stat"><div class="stat-label">Total RX</div><div class="stat-value c-orange" id="stat-rx" style="font-size:18px;padding-top:4px">—</div></div>
  <div class="stat"><div class="stat-label">Total TX</div><div class="stat-value c-orange" id="stat-tx" style="font-size:18px;padding-top:4px">—</div></div>
</div>

<div class="content">
  <div class="section-head">
    <span class="section-title">Peers</span>
    <span class="peer-count" id="peer-count">0</span>
    <div class="search-wrap">
      <svg width="14" height="14" viewBox="0 0 16 16" fill="currentColor">
        <path d="M11.742 10.344a6.5 6.5 0 10-1.397 1.398h-.001c.03.04.062.078.098.115l3.85 3.85a1 1 0 001.415-1.414l-3.85-3.85a1.007 1.007 0 00-.115-.099zm-5.242 1.656a5.5 5.5 0 110-11 5.5 5.5 0 010 11z"/>
      </svg>
      <input type="text" id="search" placeholder="Search by name, key, IP…" oninput="onSearch(this.value)"/>
    </div>
    <button class="btn btn-primary ml-auto" onclick="openAdd()">
      <svg width="13" height="13" viewBox="0 0 16 16" fill="currentColor"><path d="M8 2a.75.75 0 01.75.75v4.5h4.5a.75.75 0 010 1.5h-4.5v4.5a.75.75 0 01-1.5 0v-4.5h-4.5a.75.75 0 010-1.5h4.5v-4.5A.75.75 0 018 2z"/></svg>
      Add Peer
    </button>
  </div>

  <div class="table-wrap">
    <table>
      <thead>
        <tr>
          <th>Name / IP</th>
          <th>Public Key</th>
          <th>Last Handshake</th>
          <th>Traffic</th>
          <th>Status</th>
          <th></th>
        </tr>
      </thead>
      <tbody id="peer-tbody">
        <tr><td colspan="6" class="empty-cell"><div class="spinner"></div></td></tr>
      </tbody>
    </table>
  </div>
</div>

<!-- Add Peer Modal -->
<div class="overlay" id="add-modal">
  <div class="modal modal-lg">
    <button class="modal-close" onclick="closeAdd()">✕</button>
    <h3>Add Peer</h3>
    <div class="form-group">
      <label class="form-label">Name</label>
      <input class="form-input" id="p-name" placeholder="e.g. phone, laptop, office"/>
    </div>
    <div class="form-group">
      <label class="form-label">Allowed IPs</label>
      <div class="input-row">
        <input class="form-input" id="p-ips" placeholder="10.0.0.2/32"/>
        <button class="btn btn-ghost btn-sm" id="suggest-btn" onclick="suggestIP()" title="Auto-fill next available IP">Suggest</button>
      </div>
      <div class="form-hint">The IP(s) this peer is allowed to use on the VPN</div>
    </div>
    <div class="form-group">
      <label class="form-label">Public Key</label>
      <div class="input-row">
        <input class="form-input" id="p-pub" placeholder="Peer public key (base64)"/>
        <button class="btn btn-blue btn-sm" id="gen-btn" onclick="generateKeys()">Generate</button>
      </div>
    </div>
    <div class="form-group">
      <label class="form-label">Preshared Key <span style="color:var(--fg2);font-size:11px;text-transform:none;font-weight:400">(optional, auto-filled on generate)</span></label>
      <input class="form-input" id="p-psk" placeholder="Preshared key (base64)"/>
    </div>

    <div class="config-preview" id="config-preview">
      <div class="label-row">
        <span class="form-label" style="margin:0">Client Config <span class="generated-tag">generated</span></span>
        <button class="btn btn-ghost btn-sm" onclick="copyConfig('p-config-text', this)">Copy</button>
      </div>
      <pre class="config-box" id="p-config-text"></pre>
      <div class="form-hint" style="margin-top:6px">Save this now — the private key will not be shown again</div>
    </div>

    <div class="modal-footer">
      <button class="btn btn-ghost" onclick="closeAdd()">Cancel</button>
      <button class="btn btn-primary" id="add-btn" onclick="submitPeer()">Add Peer</button>
    </div>
  </div>
</div>

<!-- QR Modal -->
<div class="overlay" id="qr-modal">
  <div class="modal" style="max-width:420px">
    <button class="modal-close" onclick="closeQR()">✕</button>
    <h3 id="qr-title">Client Config</h3>
    <div class="qr-body">
      <div class="qr-img-wrap" id="qr-img-wrap">
        <div class="qr-placeholder" id="qr-placeholder"><div class="spinner"></div></div>
        <img id="qr-img" src="" alt="QR" style="display:none" width="220" height="220"/>
      </div>
      <div class="qr-config" style="width:100%">
        <div class="label-row" style="margin-bottom:6px">
          <span class="form-label" style="margin:0">Config</span>
          <button class="btn btn-ghost btn-sm" onclick="copyConfig('qr-config-text', this)">Copy</button>
        </div>
        <pre class="config-box" id="qr-config-text" style="max-height:160px"></pre>
      </div>
    </div>
    <div class="modal-footer">
      <button class="btn btn-ghost" id="qr-download-btn">Download .conf</button>
      <button class="btn btn-primary" onclick="closeQR()">Done</button>
    </div>
  </div>
</div>

<div id="toast-container"></div>

<script>
const APP = {{json .}};

let peers = [];
let searchQuery = '';
let refreshTimer = null;

// ── Helpers ────────────────────────────────────────────────────────────────
function esc(s) {
  return String(s||'').replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;').replace(/"/g,'&quot;');
}
function shortKey(s) {
  if (!s) return '';
  return s.length > 12 ? s.slice(0,8) + '…' + s.slice(-4) : s;
}
function humanBytes(b) {
  if (!b || b < 1024) return (b||0) + ' B';
  const units = ['KB','MB','GB','TB'];
  let i = -1, v = b;
  do { v /= 1024; i++; } while (v >= 1024 && i < units.length - 1);
  return v.toFixed(1) + ' ' + units[i];
}
function timeAgo(ts) {
  if (!ts) return 'never';
  const d = Math.floor(Date.now()/1000) - ts;
  if (d < 5)   return 'just now';
  if (d < 60)  return d + 's ago';
  if (d < 3600) return Math.floor(d/60) + 'm ago';
  if (d < 86400) return Math.floor(d/3600) + 'h ago';
  return Math.floor(d/86400) + 'd ago';
}

// ── Toast ──────────────────────────────────────────────────────────────────
function toast(msg, type='info') {
  const icons = {
    success: '<svg class="toast-icon" viewBox="0 0 16 16" fill="var(--green)"><path d="M8 16A8 8 0 108 0a8 8 0 000 16zm3.78-9.72l-5 5a.75.75 0 01-1.06 0l-2-2a.75.75 0 111.06-1.06l1.47 1.47 4.47-4.47a.75.75 0 111.06 1.06z"/></svg>',
    error:   '<svg class="toast-icon" viewBox="0 0 16 16" fill="var(--red)"><path d="M8 16A8 8 0 108 0a8 8 0 000 16zm-.75-4.75h1.5v1.5h-1.5v-1.5zm0-6.5h1.5v5h-1.5v-5z"/></svg>',
    info:    '<svg class="toast-icon" viewBox="0 0 16 16" fill="var(--blue)"><path d="M8 16A8 8 0 108 0a8 8 0 000 16zm0-11.5a1 1 0 110 2 1 1 0 010-2zm-.5 3.5h1v4h-1v-4z"/></svg>',
  };
  const el = document.createElement('div');
  el.className = 'toast ' + type;
  el.innerHTML = (icons[type]||icons.info) + '<span class="toast-msg">' + esc(msg) + '</span>' +
    '<button class="toast-close" onclick="this.parentNode.remove()">×</button>';
  document.getElementById('toast-container').appendChild(el);
  setTimeout(() => { el.style.animation = 'slideOut .2s ease forwards'; setTimeout(() => el.remove(), 200); }, 3500);
}

// ── Copy ───────────────────────────────────────────────────────────────────
async function copyText(text, btn) {
  try {
    await navigator.clipboard.writeText(text);
    if (btn) { btn.classList.add('copied'); setTimeout(() => btn.classList.remove('copied'), 1500); }
    toast('Copied to clipboard', 'success');
  } catch { toast('Copy failed', 'error'); }
}
async function copyConfig(elId, btn) {
  const text = document.getElementById(elId).textContent;
  copyText(text, null);
  if (btn) { const orig = btn.textContent; btn.textContent = 'Copied!'; setTimeout(() => btn.textContent = orig, 1500); }
}

// ── Load & render peers ────────────────────────────────────────────────────
async function loadPeers(manual) {
  const btn = document.getElementById('refresh-btn');
  btn.classList.add('spinning');
  try {
    const r = await fetch('/api/peers');
    if (!r.ok) throw new Error('HTTP ' + r.status);
    peers = (await r.json()) || [];
    renderStats();
    renderTable();
    if (manual) toast('Peers refreshed', 'success');
  } catch(e) {
    toast('Failed to load peers: ' + e.message, 'error');
  } finally {
    btn.classList.remove('spinning');
  }
}

function renderStats() {
  const online = peers.filter(p => p.online).length;
  const totalRx = peers.reduce((s,p) => s + (p.rx_bytes||0), 0);
  const totalTx = peers.reduce((s,p) => s + (p.tx_bytes||0), 0);
  document.getElementById('stat-total').textContent   = peers.length;
  document.getElementById('stat-online').textContent  = online;
  document.getElementById('stat-offline').textContent = peers.length - online;
  document.getElementById('stat-rx').textContent      = humanBytes(totalRx);
  document.getElementById('stat-tx').textContent      = humanBytes(totalTx);
  document.getElementById('peer-count').textContent   = peers.length;
}

function onSearch(q) {
  searchQuery = q.toLowerCase().trim();
  renderTable();
}

function renderTable() {
  const filtered = searchQuery
    ? peers.filter(p =>
        (p.name||'').toLowerCase().includes(searchQuery) ||
        (p.public_key||'').toLowerCase().includes(searchQuery) ||
        (p.allowed_ips||'').includes(searchQuery))
    : peers;

  const tbody = document.getElementById('peer-tbody');
  if (filtered.length === 0) {
    if (searchQuery) {
      tbody.innerHTML = '<tr><td colspan="6" class="empty-cell"><div style="color:var(--fg2)">No peers match "<strong>' + esc(searchQuery) + '</strong>"</div></td></tr>';
    } else {
      tbody.innerHTML = '<tr><td colspan="6" class="empty-cell"><div class="empty-icon">🛡</div><div class="empty-title">No peers yet</div><div class="empty-sub">Click "Add Peer" to add the first one</div></td></tr>';
    }
    return;
  }

  tbody.innerHTML = filtered.map(p => {
    const name = esc(p.name || '—');
    const ip   = esc(p.allowed_ips || '—');
    const key  = esc(p.public_key  || '');
    const hs   = timeAgo(p.last_handshake);
    const rx   = humanBytes(p.rx_bytes);
    const tx   = humanBytes(p.tx_bytes);
    const online = p.online;
    const pname  = esc((p.name || p.public_key.slice(0,8)));

    return '<tr>' +
      '<td><div class="peer-name">' + name + '</div><div class="peer-ip">' + ip + '</div></td>' +
      '<td><div class="key-wrap"><span class="key-mono" title="' + key + '">' + shortKey(p.public_key) + '</span>' +
        '<button class="copy-btn" onclick="copyText(\'' + key + '\',this)" title="Copy public key">' +
        '<svg width="12" height="12" viewBox="0 0 16 16" fill="currentColor"><path d="M0 6.75C0 5.784.784 5 1.75 5h1.5a.75.75 0 010 1.5h-1.5a.25.25 0 00-.25.25v7.5c0 .138.112.25.25.25h7.5a.25.25 0 00.25-.25v-1.5a.75.75 0 011.5 0v1.5A1.75 1.75 0 019.25 16h-7.5A1.75 1.75 0 010 14.25v-7.5z"/><path d="M5 1.75C5 .784 5.784 0 6.75 0h7.5C15.216 0 16 .784 16 1.75v7.5A1.75 1.75 0 0114.25 11h-7.5A1.75 1.75 0 015 9.25v-7.5zm1.75-.25a.25.25 0 00-.25.25v7.5c0 .138.112.25.25.25h7.5a.25.25 0 00.25-.25v-7.5a.25.25 0 00-.25-.25h-7.5z"/></svg>' +
        '</button></div></td>' +
      '<td style="color:var(--fg2);font-size:12px">' + esc(hs) + '</td>' +
      '<td><div class="traffic-up"><span>↓</span>' + rx + '</div><div class="traffic-up"><span>↑</span>' + tx + '</div></td>' +
      '<td><span class="badge ' + (online ? 'badge-online' : 'badge-offline') + '">' +
        '<span class="dot ' + (online ? 'dot-green' : 'dot-red') + '"></span>' +
        (online ? 'Online' : 'Offline') + '</span></td>' +
      '<td><div class="actions">' +
        (p.has_config
          ? '<button class="btn btn-ghost btn-sm" onclick="showQR(\'' + key + '\',\'' + pname + '\')">QR / Config</button>'
          : '') +
        '<button class="btn btn-danger btn-sm" data-state="idle" data-pubkey="' + key + '" data-name="' + pname + '" onclick="confirmDelete(this)">Delete</button>' +
      '</div></td>' +
    '</tr>';
  }).join('');
}

// ── Delete (inline confirm) ────────────────────────────────────────────────
function confirmDelete(btn) {
  if (btn.dataset.state === 'idle') {
    btn.dataset.state = 'confirm';
    btn.className = 'btn btn-confirm btn-sm';
    btn.textContent = 'Sure?';
    setTimeout(() => {
      if (btn.dataset.state === 'confirm') {
        btn.dataset.state = 'idle';
        btn.className = 'btn btn-danger btn-sm';
        btn.textContent = 'Delete';
      }
    }, 3000);
  } else {
    doDelete(btn.dataset.pubkey, btn.dataset.name);
  }
}
async function doDelete(pubkey, name) {
  try {
    const r = await fetch('/api/peers/' + encodeURIComponent(pubkey), {method:'DELETE'});
    if (!r.ok) { const d=await r.json(); throw new Error(d.error); }
    toast('Removed peer: ' + name, 'success');
    await loadPeers(false);
  } catch(e) { toast('Delete failed: ' + e.message, 'error'); }
}

// ── Add Peer modal ─────────────────────────────────────────────────────────
let generatedClientConfig = '';

function openAdd() {
  document.getElementById('p-name').value = '';
  document.getElementById('p-ips').value  = '';
  document.getElementById('p-pub').value  = '';
  document.getElementById('p-psk').value  = '';
  document.getElementById('p-config-text').textContent = '';
  document.getElementById('config-preview').style.display = 'none';
  document.getElementById('add-btn').disabled = false;
  generatedClientConfig = '';
  document.getElementById('add-modal').classList.add('open');
  setTimeout(() => document.getElementById('p-name').focus(), 50);
}
function closeAdd() { document.getElementById('add-modal').classList.remove('open'); }

async function suggestIP() {
  const btn = document.getElementById('suggest-btn');
  btn.disabled = true;
  btn.textContent = '…';
  try {
    const r = await fetch('/api/suggest-ip');
    const d = await r.json();
    if (d.error) throw new Error(d.error);
    document.getElementById('p-ips').value = d.ip;
    toast('Suggested: ' + d.ip, 'info');
  } catch(e) { toast('Suggest IP failed: ' + e.message, 'error'); }
  finally { btn.disabled = false; btn.textContent = 'Suggest'; }
}

async function generateKeys() {
  const btn = document.getElementById('gen-btn');
  btn.disabled = true; btn.textContent = '…';
  try {
    const r = await fetch('/api/generate', {method:'POST'});
    const d = await r.json();
    if (d.error) throw new Error(d.error);

    document.getElementById('p-pub').value = d.public_key;
    document.getElementById('p-psk').value = d.preshared_key;

    const ips = document.getElementById('p-ips').value || d.allowed_ips || '0.0.0.0/0, ::/0';
    const name = document.getElementById('p-name').value || '<your-device>';

    generatedClientConfig = buildClientConfig(d.private_key, d.preshared_key, d.server_pub,
      d.server_endpoint, d.dns, ips, name);

    document.getElementById('p-config-text').textContent = generatedClientConfig;
    document.getElementById('config-preview').style.display = 'block';
    toast('Keys generated', 'success');
  } catch(e) { toast('Generate failed: ' + e.message, 'error'); }
  finally { btn.disabled = false; btn.textContent = 'Generate'; }
}

function buildClientConfig(priv, psk, serverPub, endpoint, dns, allowedIPs, address) {
  let c = '[Interface]\n';
  c += 'PrivateKey = ' + priv + '\n';
  c += 'Address = ' + (address && address !== '<your-device>' ? allowedIPs : '10.x.x.x/32') + '\n';
  if (dns) c += 'DNS = ' + dns + '\n';
  c += '\n[Peer]\n';
  if (serverPub) c += 'PublicKey = ' + serverPub + '\n';
  if (psk) c += 'PresharedKey = ' + psk + '\n';
  c += 'AllowedIPs = 0.0.0.0/0, ::/0\n';
  if (endpoint) c += 'Endpoint = ' + endpoint + '\n';
  c += 'PersistentKeepalive = 25\n';
  return c;
}

async function submitPeer() {
  const name   = document.getElementById('p-name').value.trim();
  const pub    = document.getElementById('p-pub').value.trim();
  const psk    = document.getElementById('p-psk').value.trim();
  const ips    = document.getElementById('p-ips').value.trim() || APP.DefaultIPs;
  if (!pub) { toast('Public key is required', 'error'); return; }

  const btn = document.getElementById('add-btn');
  btn.disabled = true; btn.textContent = 'Adding…';
  try {
    const r = await fetch('/api/peers', {
      method: 'POST',
      headers: {'Content-Type':'application/json'},
      body: JSON.stringify({name, public_key: pub, preshared_key: psk,
        allowed_ips: ips, client_config: generatedClientConfig})
    });
    const d = await r.json();
    if (!r.ok) throw new Error(d.error || 'Unknown error');
    toast('Peer "' + (name||pub.slice(0,8)) + '" added', 'success');
    closeAdd();
    await loadPeers(false);
  } catch(e) { toast('Add failed: ' + e.message, 'error'); }
  finally { btn.disabled = false; btn.textContent = 'Add Peer'; }
}

// ── QR / Config modal ──────────────────────────────────────────────────────
async function showQR(pubkey, name) {
  document.getElementById('qr-title').textContent = name || pubkey.slice(0,8);
  document.getElementById('qr-config-text').textContent = '';
  document.getElementById('qr-img').style.display = 'none';
  document.getElementById('qr-placeholder').style.display = 'flex';
  document.getElementById('qr-placeholder').innerHTML = '<div class="spinner"></div>';
  document.getElementById('qr-modal').classList.add('open');

  // download button
  const dlBtn = document.getElementById('qr-download-btn');
  dlBtn.onclick = () => {
    const a = document.createElement('a');
    a.href = '/api/peers/' + encodeURIComponent(pubkey) + '/config';
    a.download = (name || 'peer') + '.conf';
    a.click();
  };

  // load config text
  try {
    const cr = await fetch('/api/peers/' + encodeURIComponent(pubkey) + '/config');
    if (cr.ok) {
      document.getElementById('qr-config-text').textContent = await cr.text();
    }
  } catch {}

  // load QR image
  const img = document.getElementById('qr-img');
  img.onload = () => {
    document.getElementById('qr-placeholder').style.display = 'none';
    img.style.display = 'block';
  };
  img.onerror = () => {
    document.getElementById('qr-placeholder').innerHTML = '<div style="color:var(--fg2);font-size:12px;text-align:center">QR not available</div>';
  };
  img.src = '/api/peers/' + encodeURIComponent(pubkey) + '/qr?' + Date.now();
}
function closeQR() { document.getElementById('qr-modal').classList.remove('open'); }

// ── Close overlays on bg click ─────────────────────────────────────────────
document.querySelectorAll('.overlay').forEach(el => {
  el.addEventListener('click', e => { if (e.target === el) el.classList.remove('open'); });
});
document.addEventListener('keydown', e => {
  if (e.key === 'Escape') document.querySelectorAll('.overlay.open').forEach(el => el.classList.remove('open'));
});

// ── Auto-refresh every 30s ─────────────────────────────────────────────────
loadPeers(false);
refreshTimer = setInterval(() => loadPeers(false), 30000);
</script>
</body>
</html>`
