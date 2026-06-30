<div align="center">

<img src="https://capsule-render.vercel.app/api?type=waving&color=7c3aed&height=120&section=header&text=wireguard-manager&fontSize=34&fontColor=ffffff&fontAlignY=38&desc=Web%20UI%20for%20WireGuard%20peer%20management&descAlignY=62&descColor=c9d1d9" width="100%"/>

[![CI](https://github.com/MikroJit-Technologies/wireguard-manager/actions/workflows/ci.yml/badge.svg)](https://github.com/MikroJit-Technologies/wireguard-manager/actions/workflows/ci.yml)
[![Go 1.22](https://img.shields.io/badge/Go-1.22-00ADD8?style=flat-square&logo=go&logoColor=white)](https://go.dev)
[![Docker](https://img.shields.io/badge/ghcr.io%2Fwireguard--manager-latest-2496ED?style=flat-square&logo=docker&logoColor=white)](https://ghcr.io/mikrojit-technologies/wireguard-manager)
[![License: MIT](https://img.shields.io/badge/License-MIT-388bfd?style=flat-square)](LICENSE)
[![Release](https://img.shields.io/github/v/release/MikroJit-Technologies/wireguard-manager?style=flat-square&color=3fb950)](https://github.com/MikroJit-Technologies/wireguard-manager/releases)

**Add, remove, and monitor WireGuard peers from your browser — no editing config files by hand.**

[Quick Start](#quick-start) · [Configuration](#configuration) · [Features](#features) · [API](#api) · [Docker](#docker)

</div>

---

## Features

| | |
|---|---|
| **Add peers** | Name, auto-suggested IP, DNS, AllowedIPs — keys generated in-browser, never leave client |
| **QR codes** | Scan with WireGuard mobile app — one tap to configure |
| **Client config download** | `.conf` file ready to import into any WireGuard client |
| **Live stats** | Peer table with last handshake, endpoint, RX/TX, online badge — auto-refreshes |
| **Auto IP suggestion** | Scans existing peers and suggests the next free `/32` in the interface subnet |
| **Search** | Filter peer table by name or public key |
| **No-reload UI** | All CRUD via `fetch()` — table updates without page reload, toast notifications |
| **Copy buttons** | One-click copy for public key, client config, or any field |
| **Delete with confirm** | Two-click delete with 3-second confirmation timeout |
| **Auth** | Optional HTTP Basic Auth to protect the UI |

---

## Quick Start

**Docker (recommended)**

```bash
docker run -d \
  --name wireguard-manager \
  --cap-add NET_ADMIN \
  --network host \
  -v /etc/wireguard:/etc/wireguard \
  -v $(pwd)/config.yml:/app/config.yml:ro \
  ghcr.io/mikrojit-technologies/wireguard-manager:latest
```

Open **http://localhost:8080**

**Binary**

```bash
curl -Lo wireguard-manager https://github.com/MikroJit-Technologies/wireguard-manager/releases/latest/download/wireguard-manager-linux-amd64
chmod +x wireguard-manager

cp config.example.yml config.yml
# edit config.yml
sudo ./wireguard-manager
```

> **Note:** wireguard-manager calls `wg` and `wg-quick` — it must run with access to the WireGuard interface (`NET_ADMIN` cap or root).

---

## Configuration

```yaml
listen_addr: ":8080"
interface: wg0                                    # WireGuard interface name
wg_config: /etc/wireguard/wg0.conf               # Path to wg0.conf

server:
  endpoint: "vpn.example.com:51820"              # Clients connect here
  public_key: "SERVER_PUBLIC_KEY_HERE"           # Server's public key
  dns: "1.1.1.1, 8.8.8.8"                       # DNS pushed to clients

defaults:
  allowed_ips: "0.0.0.0/0, ::/0"               # Route all traffic
  mtu: 1420

auth:
  username: admin
  password: "changeme"                           # Remove block to disable auth
```

### Reference

| Key | Default | Description |
|---|---|---|
| `listen_addr` | `:8080` | HTTP listen address |
| `interface` | `wg0` | WireGuard interface (`ip link`) |
| `wg_config` | `/etc/wireguard/wg0.conf` | Config file path for peer persistence |
| `server.endpoint` | — | Public hostname:port clients connect to |
| `server.public_key` | — | Server's WireGuard public key |
| `server.dns` | `1.1.1.1` | DNS servers pushed to clients |
| `defaults.allowed_ips` | `0.0.0.0/0, ::/0` | Default AllowedIPs for new peers |
| `defaults.mtu` | `1420` | MTU in generated client configs |
| `auth.username` | — | HTTP Basic Auth username |
| `auth.password` | — | HTTP Basic Auth password |

---

## How Keys Work

Private keys **never leave the browser**. The key generation flow:

```
Browser
  │
  ├─ generateKeys()       — calls POST /api/generate → server generates keypair
  │     returns: { private_key, public_key }
  │
  ├─ suggestIP()          — calls GET /api/suggest-ip → scans used IPs, returns next free
  │
  ├─ buildClientConfig()  — assembles .conf in JS using:
  │     [Interface] PrivateKey = <private_key>
  │     [Interface] Address = <suggested_ip>
  │     [Interface] DNS = <server.dns>
  │     [Peer] PublicKey = <server.public_key>
  │     [Peer] Endpoint = <server.endpoint>
  │     [Peer] AllowedIPs = <defaults.allowed_ips>
  │
  └─ submitPeer()         — sends POST /api/peers with { name, public_key, allowed_ips, client_config }
                            server: wg set wg0 peer <pubkey> allowed-ips <ips> + appends to wg0.conf
                                    stores client_config for QR generation
```

The server stores `client_config` in memory (24h TTL). QR and config download use this stored value.

---

## Dashboard

```
┌─────────────────────────────────────────────────────────────────────┐
│  wireguard-manager    [wg0]  [10.0.0.0/24]          Search peers…  │
├─────────────────────────────────────────────────────────────────────┤
│  + Add Peer                                                          │
│  Name ___________  IP 10.0.0.X/32  DNS ______  AllowedIPs ______   │
│  [Generate Keys]                                [Save Peer]         │
├─────────────────────────────────────────────────────────────────────┤
│  Name         Public Key          Endpoint         RX/TX    Status  │
│  laptop       xLgJk3N...Q=        203.0.113.1      12M/3M   ● UP   │
│  phone        Yt8p2K...A=         (none)            0B/204B  ○ —    │
│  server-eu    mP9qL...Z=          185.42.1.1       88M/21M  ● UP   │
│                                                   [QR] [Config] [✕] │
└─────────────────────────────────────────────────────────────────────┘
```

- **Green dot** = handshake within last 3 minutes
- **QR button** = modal with scannable QR code
- **Config button** = download `.conf` file
- **✕** = delete (two-click confirmation, 3s timeout)

---

## API

| Method | Path | Description |
|---|---|---|
| `GET` | `/api/peers` | List all peers with live stats |
| `POST` | `/api/peers` | Add a new peer |
| `DELETE` | `/api/peers/{pubkey}` | Remove a peer |
| `POST` | `/api/generate` | Generate a new keypair |
| `GET` | `/api/suggest-ip` | Get next available IP in subnet |
| `GET` | `/api/peers/{pubkey}/qr` | QR code PNG for client config |
| `GET` | `/api/peers/{pubkey}/config` | Download client `.conf` file |

**Add peer:**

```bash
curl -X POST http://localhost:8080/api/peers \
  -H 'Content-Type: application/json' \
  -d '{"name":"laptop","public_key":"xLgJk3N...Q=","allowed_ips":"10.0.0.5/32","client_config":"[Interface]\n..."}'
```

**List peers:**

```bash
curl http://localhost:8080/api/peers | jq '.[].name, .[].endpoint, .[].online'
```

---

## Docker

**docker-compose.yml:**

```yaml
services:
  wireguard-manager:
    image: ghcr.io/mikrojit-technologies/wireguard-manager:latest
    container_name: wireguard-manager
    restart: unless-stopped
    cap_add:
      - NET_ADMIN
    network_mode: host
    volumes:
      - /etc/wireguard:/etc/wireguard
      - ./config.yml:/app/config.yml:ro
    environment:
      - TZ=Asia/Bangkok
```

```bash
docker compose up -d
```

---

## Part of MikroJit Technologies

<div align="center">

| Tool | Description |
|---|---|
| [mikrotik-exporter](https://github.com/MikroJit-Technologies/mikrotik-exporter) | Prometheus exporter for MikroTik RouterOS |
| [mikrotik-backup](https://github.com/MikroJit-Technologies/mikrotik-backup) | Auto-backup RouterOS configs to Git + Telegram diffs |
| **wireguard-manager** | **Web UI for WireGuard peer management** |
| [netmon](https://github.com/MikroJit-Technologies/netmon) | Uptime monitor — HTTP / ping / TCP + Telegram alerts |
| [routeros-cli](https://github.com/MikroJit-Technologies/routeros-cli) | CLI tool for multi-device RouterOS command execution |

</div>

---

<div align="center">

<img src="https://capsule-render.vercel.app/api?type=waving&color=7c3aed&height=80&section=footer" width="100%"/>

MIT License · [MikroJit Technologies](https://github.com/MikroJit-Technologies) · Bangkok, Thailand

</div>
