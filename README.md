# ironclad

A bare-metal Linux fleet monitoring system. No Docker, no Kubernetes, no cloud.

## Architecture

┌─────────────────────────────┐     Unix Socket      ┌──────────────────────────────┐
│      ironclad-agent         │ ──────────────────► │    ironclad-controller        │
│         (Rust)              │                      │         (Go)                  │
│                             │                      │                               │
│ • Reads /proc via sysinfo   │                      │ • Receives JSON metrics       │
│ • Decrypts secrets w/ age   │                      │ • Sustained alerting engine   │
│ • Serializes to JSON        │                      │ • Prometheus /metrics :9100   │
│ • Streams over Unix socket  │                      │ • Multi-node hostname labels  │
│ • Runs as systemd service   │                      │                               │
└─────────────────────────────┘                      └──────────────────────────────┘

## Stack

- **Agent:** Rust, Tokio, sysinfo, serde, age encryption
- **Controller:** Go, Prometheus client
- **IPC:** Unix domain sockets
- **Service management:** systemd
- **Observability:** Prometheus-compatible `/metrics` endpoint
- **Secret management:** age-encrypted secrets file

## Features

- Live CPU and memory metrics read directly from the Linux kernel
- JSON serialization over Unix sockets — no HTTP, no network stack
- Sustained alerting: fires when CPU > 80% or MEM > 90% for 10+ seconds
- Age-encrypted secrets loaded at agent startup — plaintext never touches disk
- Agent runs as a systemd service: starts on boot, restarts on crash
- Prometheus `/metrics` endpoint with per-host labels for multi-node scraping
- Fully static binary (musl) — single file, zero dependencies, runs anywhere
- 1.3MB binary, 3.4MB RSS at idle

## Screenshots

  <img width="431" height="211" alt="hostname_SR" src="https://github.com/user-attachments/assets/6bb89e13-64b0-413a-9d14-35510b759406" />

  <img width="426" height="217" alt="Stress_test" src="https://github.com/user-attachments/assets/39833328-1cc4-4a53-87af-b92638149a03" />

  <img width="758" height="373" alt="Prom_SR" src="https://github.com/user-attachments/assets/34091067-e26b-4d31-9c2a-ec15706dcf18" />

## Running

**Start the agent (or install as systemd service):**
```bash
cd agent && cargo build --release
sudo cp systemd/ironclad-agent.service /etc/systemd/system/
sudo systemctl enable --now ironclad-agent
```

**Start the controller:**
```bash
cd controller && go run main.go
```

**View Prometheus metrics:**
http://localhost:9100/metrics


**Test alerting (simulate high CPU):**
```bash
stress-ng --cpu 16 --timeout 60s
```
Controller will fire `[ALERT]` after 10 seconds of sustained CPU above 80%.


**Run tests:**
```bash
cd agent && cargo test
cd controller && go test
```

## Secret Management

Secrets are encrypted with [age](https://age-encryption.org/) and loaded at agent startup:

```bash
age-keygen -o secrets/identity.txt
age -r <pubkey> -o secrets/secrets.age secrets/secrets.txt
```

The private key (`identity.txt`) is never committed to git.


## Philosophy

Minimal, understandable, close to the hardware. No abstractions you don't need. Owns the stack from kernel metrics to observability.

  


