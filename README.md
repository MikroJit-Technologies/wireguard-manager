# wireguard-manager

> Lightweight web UI for managing WireGuard peers — part of the [MikroJit Technologies](https://github.com/MikroJit-Technologies) toolchain.

[![CI](https://github.com/MikroJit-Technologies/wireguard-manager/actions/workflows/ci.yml/badge.svg)](https://github.com/MikroJit-Technologies/wireguard-manager/actions/workflows/ci.yml)
[![Go 1.22](https://img.shields.io/badge/Go-1.22-00ADD8?logo=go)](https://go.dev)
[![Docker](https://img.shields.io/badge/ghcr.io-wireguard--manager-0d1117?logo=docker)](https://ghcr.io/mikrojit-technologies/wireguard-manager)
[![License: MIT](https://img.shields.io/badge/License-MIT-388bfd)](LICENSE)

---

## Features

- **Peer list** — live status (online/offline), RX/TX, last handshake
- **Add peers** — generate key pair + preshared key in one click, or paste your own public key
- **QR code** — scan with WireGuard mobile app instantly
- **Delete peers** — removes from config + running interface atomically
- **Dark UI** — native `#0d1117` theme, no external dependencies
- **Optional basic auth** — single username/password gate
- Multi-arch Docker image: `amd64`, `arm64`, `arm/v7`

## Quick Start

```bash
# Docker Compose
cp config.example.yml config.yml
# edit config.yml — set endpoint, dns, auth
docker compose up -d
```

Open `http://localhost:8080`

## Configuration

```yaml
listen_addr: ":8080"
interface: "wg0"
wg_config: "/etc/wireguard/wg0.conf"

server:
  endpoint: "vpn.yourdomain.com:51820"
  dns: "1.1.1.1"

defaults:
  allowed_ips: "0.0.0.0/0, ::/0"
  mtu: 1420

auth:
  username: "admin"
  password: "changeme"
```

| Key | Default | Description |
|---|---|---|
| `listen_addr` | `:8080` | HTTP listen address |
| `interface` | `wg0` | WireGuard interface name |
| `wg_config` | `/etc/wireguard/wg0.conf` | Path to wg0.conf |
| `server.endpoint` | — | Shown in generated client configs |
| `server.dns` | — | DNS for clients |
| `defaults.allowed_ips` | `0.0.0.0/0, ::/0` | Default allowed IPs for new peers |
| `auth.username/password` | — | Basic auth (empty = no auth) |

## API

| Method | Path | Description |
|---|---|---|
| `GET` | `/api/peers` | List all peers with live stats |
| `POST` | `/api/peers` | Add peer `{name, public_key, preshared_key, allowed_ips}` |
| `DELETE` | `/api/peers/{pubkey}` | Remove peer |
| `POST` | `/api/generate` | Generate new key pair + PSK |
| `GET` | `/api/peers/{pubkey}/qr` | QR PNG for client config |
| `GET` | `/health` | Health check |

## Docker

```bash
docker run -d \
  --name wireguard-manager \
  --cap-add NET_ADMIN \
  --network host \
  -v /etc/wireguard:/etc/wireguard \
  -v $(pwd)/config.yml:/app/config.yml:ro \
  ghcr.io/mikrojit-technologies/wireguard-manager:latest
```

## Requirements

- WireGuard tools (`wg` binary) installed and `wg0.conf` existing
- `NET_ADMIN` capability (or root) to call `wg set` / `wg addconf`

---

*Part of [MikroJit Technologies](https://github.com/MikroJit-Technologies) — network tooling & infrastructure automation*
