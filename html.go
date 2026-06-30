package main

import (
	"fmt"
	"html/template"
	"io"
	"time"
)

var pageTmpl = template.Must(template.New("page").Funcs(template.FuncMap{
	"humanBytes": humanBytes,
	"timeAgo":    timeAgo,
	"shortKey":   func(s string) string {
		if len(s) > 12 {
			return s[:8] + "…" + s[len(s)-4:]
		}
		return s
	},
	"sub": func(a, b int) int { return a - b },
}).Parse(pageHTML))

func renderHTML(w io.Writer, data map[string]any) {
	pageTmpl.Execute(w, data)
}

func humanBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

func timeAgo(ts int64) string {
	if ts == 0 {
		return "never"
	}
	d := time.Now().Unix() - ts
	switch {
	case d < 60:
		return fmt.Sprintf("%ds ago", d)
	case d < 3600:
		return fmt.Sprintf("%dm ago", d/60)
	case d < 86400:
		return fmt.Sprintf("%dh ago", d/3600)
	default:
		return fmt.Sprintf("%dd ago", d/86400)
	}
}

const pageHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8"/>
<meta name="viewport" content="width=device-width,initial-scale=1"/>
<title>WireGuard Manager</title>
<style>
*{box-sizing:border-box;margin:0;padding:0}
body{font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',sans-serif;background:#0d1117;color:#c9d1d9;min-height:100vh}
a{color:#388bfd;text-decoration:none}

/* Nav */
nav{background:#161b22;border-bottom:1px solid #30363d;padding:0 24px;display:flex;align-items:center;height:56px;gap:16px}
.nav-brand{font-weight:700;font-size:16px;color:#f0f6fc;display:flex;align-items:center;gap:8px}
.nav-brand svg{width:20px;height:20px}
.nav-iface{background:#0d1117;border:1px solid #30363d;border-radius:6px;padding:4px 10px;color:#8b949e;font-size:13px}
.nav-right{margin-left:auto;display:flex;gap:10px;align-items:center}

/* Stats bar */
.stats-bar{padding:20px 24px;display:flex;gap:16px;flex-wrap:wrap}
.stat-card{background:#161b22;border:1px solid #30363d;border-radius:10px;padding:16px 20px;min-width:140px}
.stat-label{font-size:12px;color:#8b949e;text-transform:uppercase;letter-spacing:.5px;margin-bottom:6px}
.stat-value{font-size:28px;font-weight:700;color:#f0f6fc}
.stat-value.green{color:#3fb950}
.stat-value.blue{color:#388bfd}

/* Peers table */
.section{padding:0 24px 32px}
.section-header{display:flex;align-items:center;justify-content:space-between;margin-bottom:16px}
.section-title{font-size:16px;font-weight:600;color:#f0f6fc}
.table-wrap{background:#161b22;border:1px solid #30363d;border-radius:10px;overflow:hidden}
table{width:100%;border-collapse:collapse}
th{text-align:left;padding:12px 16px;font-size:12px;font-weight:600;color:#8b949e;text-transform:uppercase;letter-spacing:.5px;border-bottom:1px solid #21262d}
td{padding:14px 16px;font-size:14px;border-bottom:1px solid #21262d;vertical-align:middle}
tr:last-child td{border-bottom:none}
tr:hover td{background:#1c2128}

.badge{display:inline-flex;align-items:center;gap:5px;padding:3px 9px;border-radius:12px;font-size:12px;font-weight:500}
.badge-online{background:#1a3622;color:#3fb950}
.badge-offline{background:#2a1f1f;color:#f85149}
.dot{width:6px;height:6px;border-radius:50%}
.dot-green{background:#3fb950}
.dot-red{background:#f85149}

.key-mono{font-family:'SFMono-Regular',Consolas,monospace;font-size:12px;color:#8b949e;background:#0d1117;padding:2px 6px;border-radius:4px}
.actions{display:flex;gap:6px}

/* Buttons */
.btn{display:inline-flex;align-items:center;gap:6px;padding:7px 14px;border-radius:6px;font-size:13px;font-weight:500;cursor:pointer;border:none;transition:opacity .15s}
.btn:hover{opacity:.85}
.btn-primary{background:#238636;color:#fff}
.btn-blue{background:#1f6feb;color:#fff}
.btn-sm{padding:4px 10px;font-size:12px}
.btn-ghost{background:transparent;border:1px solid #30363d;color:#c9d1d9}
.btn-danger{background:transparent;border:1px solid #30363d;color:#f85149}
.btn-danger:hover{background:#2a1f1f;border-color:#f85149}

/* Modal */
.modal-overlay{display:none;position:fixed;inset:0;background:rgba(0,0,0,.7);z-index:100;align-items:center;justify-content:center}
.modal-overlay.open{display:flex}
.modal{background:#161b22;border:1px solid #30363d;border-radius:12px;width:100%;max-width:520px;padding:24px;position:relative}
.modal h3{font-size:16px;font-weight:600;color:#f0f6fc;margin-bottom:20px}
.form-group{margin-bottom:16px}
label{display:block;font-size:13px;color:#8b949e;margin-bottom:6px;font-weight:500}
input,select{width:100%;background:#0d1117;border:1px solid #30363d;border-radius:6px;padding:8px 12px;color:#c9d1d9;font-size:14px;outline:none}
input:focus,select:focus{border-color:#388bfd}
.form-row{display:grid;grid-template-columns:1fr 1fr;gap:12px}
.modal-footer{display:flex;justify-content:flex-end;gap:10px;margin-top:20px}

/* QR Modal */
.qr-wrap{display:flex;flex-direction:column;align-items:center;gap:16px}
.qr-wrap img{width:240px;height:240px;border-radius:8px;background:#fff;padding:8px}
.config-box{width:100%;background:#0d1117;border:1px solid #30363d;border-radius:6px;padding:12px;font-family:monospace;font-size:12px;color:#8b949e;white-space:pre;overflow-x:auto;max-height:180px;overflow-y:auto}

/* Empty */
.empty{padding:48px;text-align:center;color:#8b949e}
.empty svg{width:40px;height:40px;margin-bottom:12px;opacity:.4}
</style>
</head>
<body>

<nav>
  <div class="nav-brand">
    <svg viewBox="0 0 24 24" fill="none" stroke="#388bfd" stroke-width="2">
      <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/>
    </svg>
    wireguard-manager
  </div>
  <span class="nav-iface">{{.Interface}}</span>
  <div class="nav-right">
    <span style="font-size:12px;color:#8b949e">Server: <span class="key-mono">{{shortKey .ServerPub}}</span></span>
  </div>
</nav>

<div class="stats-bar">
  <div class="stat-card">
    <div class="stat-label">Total Peers</div>
    <div class="stat-value blue">{{.Total}}</div>
  </div>
  <div class="stat-card">
    <div class="stat-label">Online</div>
    <div class="stat-value green">{{.Online}}</div>
  </div>
  <div class="stat-card">
    <div class="stat-label">Offline</div>
    <div class="stat-value">{{sub .Total .Online}}</div>
  </div>
</div>

<div class="section">
  <div class="section-header">
    <span class="section-title">Peers</span>
    <button class="btn btn-primary" onclick="openAdd()">
      <svg width="14" height="14" viewBox="0 0 16 16" fill="currentColor"><path d="M8 2a.75.75 0 01.75.75v4.5h4.5a.75.75 0 010 1.5h-4.5v4.5a.75.75 0 01-1.5 0v-4.5h-4.5a.75.75 0 010-1.5h4.5v-4.5A.75.75 0 018 2z"/></svg>
      Add Peer
    </button>
  </div>

  {{if .Peers}}
  <div class="table-wrap">
    <table>
      <thead>
        <tr>
          <th>Name</th>
          <th>Public Key</th>
          <th>Allowed IPs</th>
          <th>Last Handshake</th>
          <th>RX / TX</th>
          <th>Status</th>
          <th></th>
        </tr>
      </thead>
      <tbody>
        {{range .Peers}}
        <tr>
          <td><strong>{{if .Name}}{{.Name}}{{else}}—{{end}}</strong></td>
          <td><span class="key-mono">{{shortKey .PublicKey}}</span></td>
          <td><code style="font-size:12px">{{.AllowedIPs}}</code></td>
          <td style="color:#8b949e;font-size:13px">{{timeAgo .LastHandshake}}</td>
          <td style="font-size:13px;color:#8b949e">{{humanBytes .RxBytes}} / {{humanBytes .TxBytes}}</td>
          <td>
            {{if .Online}}
            <span class="badge badge-online"><span class="dot dot-green"></span>Online</span>
            {{else}}
            <span class="badge badge-offline"><span class="dot dot-red"></span>Offline</span>
            {{end}}
          </td>
          <td>
            <div class="actions">
              <button class="btn btn-ghost btn-sm" onclick="showQR('{{.PublicKey}}','{{.Name}}')">QR</button>
              <button class="btn btn-danger btn-sm" onclick="deletePeer('{{.PublicKey}}','{{.Name}}')">Delete</button>
            </div>
          </td>
        </tr>
        {{end}}
      </tbody>
    </table>
  </div>
  {{else}}
  <div class="table-wrap">
    <div class="empty">
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/></svg>
      <div>No peers yet</div>
      <div style="font-size:13px;margin-top:4px">Click "Add Peer" to create the first one</div>
    </div>
  </div>
  {{end}}
</div>

<!-- Add Peer Modal -->
<div class="modal-overlay" id="addModal">
  <div class="modal">
    <h3>Add Peer</h3>
    <div class="form-group">
      <label>Name</label>
      <input id="peerName" placeholder="e.g. phone, laptop" />
    </div>
    <div class="form-group">
      <label>Public Key</label>
      <div style="display:flex;gap:8px">
        <input id="peerPub" placeholder="peer public key" style="flex:1"/>
        <button class="btn btn-ghost btn-sm" onclick="generateKeys()" style="white-space:nowrap">Generate</button>
      </div>
    </div>
    <div class="form-group">
      <label>Preshared Key (optional)</label>
      <input id="peerPsk" placeholder="preshared key" />
    </div>
    <div class="form-group">
      <label>Allowed IPs</label>
      <input id="peerIPs" value="0.0.0.0/0, ::/0" />
    </div>
    <div id="generatedConfig" style="display:none">
      <div class="form-group">
        <label>Client Config (save before closing)</label>
        <div class="config-box" id="configText"></div>
      </div>
    </div>
    <div class="modal-footer">
      <button class="btn btn-ghost" onclick="closeAdd()">Cancel</button>
      <button class="btn btn-primary" onclick="submitPeer()">Add Peer</button>
    </div>
  </div>
</div>

<!-- QR Modal -->
<div class="modal-overlay" id="qrModal">
  <div class="modal" style="max-width:380px">
    <h3 id="qrTitle">QR Code</h3>
    <div class="qr-wrap">
      <img id="qrImg" src="" alt="QR Code"/>
      <p style="font-size:13px;color:#8b949e;text-align:center">Scan with WireGuard mobile app</p>
    </div>
    <div class="modal-footer">
      <button class="btn btn-ghost" onclick="closeQR()">Close</button>
    </div>
  </div>
</div>

<script>
let generatedPrivKey = '';
let serverPub = '{{.ServerPub}}';

function openAdd() { document.getElementById('addModal').classList.add('open'); }
function closeAdd() {
  document.getElementById('addModal').classList.remove('open');
  document.getElementById('generatedConfig').style.display='none';
  generatedPrivKey='';
}
function closeQR() { document.getElementById('qrModal').classList.remove('open'); }

async function generateKeys() {
  const r = await fetch('/api/generate', {method:'POST'});
  const d = await r.json();
  document.getElementById('peerPub').value = d.public_key;
  document.getElementById('peerPsk').value = d.preshared_key;
  generatedPrivKey = d.private_key;
  serverPub = d.server_pub;

  const cfg = buildClientConfig(d.private_key, d.preshared_key, d.server_pub,
    d.server_endpoint, d.dns, document.getElementById('peerIPs').value || d.allowed_ips);
  document.getElementById('configText').textContent = cfg;
  document.getElementById('generatedConfig').style.display='block';
}

function buildClientConfig(priv, psk, serverPub, endpoint, dns, allowedIPs) {
  let c = '[Interface]\n';
  c += 'PrivateKey = ' + priv + '\n';
  c += 'Address = <assign-ip>/32\n';
  if (dns) c += 'DNS = ' + dns + '\n';
  c += '\n[Peer]\n';
  c += 'PublicKey = ' + serverPub + '\n';
  if (psk) c += 'PresharedKey = ' + psk + '\n';
  c += 'AllowedIPs = ' + (allowedIPs || '0.0.0.0/0') + '\n';
  if (endpoint) c += 'Endpoint = ' + endpoint + '\n';
  c += 'PersistentKeepalive = 25\n';
  return c;
}

async function submitPeer() {
  const name = document.getElementById('peerName').value;
  const pub  = document.getElementById('peerPub').value;
  const psk  = document.getElementById('peerPsk').value;
  const ips  = document.getElementById('peerIPs').value;

  if (!pub) { alert('Public key is required'); return; }

  const r = await fetch('/api/peers', {
    method: 'POST',
    headers: {'Content-Type':'application/json'},
    body: JSON.stringify({name, public_key: pub, preshared_key: psk, allowed_ips: ips})
  });
  if (r.ok) { closeAdd(); location.reload(); }
  else { const d = await r.json(); alert('Error: ' + d.error); }
}

async function deletePeer(pubkey, name) {
  if (!confirm('Delete peer ' + (name||pubkey.slice(0,8)) + '?')) return;
  await fetch('/api/peers/' + encodeURIComponent(pubkey), {method:'DELETE'});
  location.reload();
}

function showQR(pubkey, name) {
  document.getElementById('qrTitle').textContent = 'QR — ' + (name || pubkey.slice(0,8));
  document.getElementById('qrImg').src = '/api/peers/' + encodeURIComponent(pubkey) + '/qr';
  document.getElementById('qrModal').classList.add('open');
}

// auto-refresh stats every 30s
setInterval(() => location.reload(), 30000);

// close modals on overlay click
document.querySelectorAll('.modal-overlay').forEach(el => {
  el.addEventListener('click', e => { if(e.target===el) el.classList.remove('open'); });
});
</script>
</body>
</html>`
