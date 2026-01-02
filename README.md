# 📈 CryptoStream — Real-Time Crypto Trading Dashboard Built with Go and Server-Sent Events

A lightweight **real-time cryptocurrency trading dashboard** written entirely in Go.  
Live price updates, trades, alerts, and statistics are streamed instantly using **Server-Sent Events (SSE)** — no WebSockets, no external services, no JavaScript frameworks.

![Go](https://img.shields.io/badge/Go-1.21-00ADD8?style=for-the-badge&logo=go&logoColor=white)
![Tailwind CSS](https://img.shields.io/badge/Tailwind_CSS-CDN-38B2AC?style=for-the-badge&logo=tailwind-css&logoColor=white)
![Render](https://img.shields.io/badge/Render-Deployed-46E3B7?style=for-the-badge&logo=render&logoColor=white)

This is a single-file, self-contained demo project perfect for:
- learning Go concurrency and HTTP handling,
- exploring Server-Sent Events in production-like scenarios,
- serving as an MVP template for real-time dashboards.

All market data is **simulated** in-memory for demonstration purposes.

---

## 🚀 Features

- Live price updates every ~800ms
- Simulated trade feed (buy/sell)
- Real-time market alerts
- Global statistics (total trades, volume, active viewers)
- Responsive Tailwind CSS UI with dark mode and animations
- Price sparklines for each asset
- Visual feedback on price changes (green/red pulse)
- Thread-safe client broadcasting via SSE
- No external dependencies or database

---

## 🧱 Tech Stack

- **Go 1.21+**
- Standard library only (`net/http`, `sync`, etc.)
- **Server-Sent Events (SSE)**
- Tailwind CSS (via CDN)
- In-memory state

Zero third-party Go packages.  
Pure vanilla JavaScript for the frontend.

---

## 📂 Project Structure

```
go-crypto-trading-dashboard/
├── main.go            # Complete application (server, simulation, SSE, HTML)
├── render.yaml        # Render deployment configuration
├── go.mod
├── .gitignore
└── README.md
```

Single-file design — everything lives in `main.go`.

---

## ▶️ Run Locally

```bash
go run .
```

Open your browser:

```
http://localhost:8080
```

You’ll immediately see simulated market activity streaming in real time.

---

## 🌐 HTTP Endpoints

- `/` — Main dashboard (HTML page with embedded Tailwind + JS)
- `/events` — Server-Sent Events stream (real-time updates)

No additional API routes — all communication happens via SSE.

---

## 📡 Real-Time Architecture

- Clients connect to `/events` using the native browser `EventSource`
- A background simulation generates price changes, trades, alerts, and stats
- Every event is broadcast to **all connected clients** instantly
- Clients are tracked and cleaned up automatically on disconnect
- Thread-safe with `sync.RWMutex`

---

## 🧠 Ideas for Improvement

- Connect to a real crypto API (e.g., Binance, Coinbase)
- Add persistent storage (Redis/PostgreSQL)
- Implement user authentication
- Add charting library (Chart.js, Lightweight Charts)
- Support multiple markets or watchlists
- Dockerize the application
- Add sound notifications for alerts

---

## ☁️ Deployment (Render)

The project is configured to deploy instantly on **Render** using the provided `render.yaml`.

```yaml
services:
  - type: web
    name: go-crypto-trading-dashboard
    env: go
    plan: free
    buildCommand: |
      if [ ! -f go.mod ]; then
        go mod init app
      fi
      go mod tidy
      go build -o app .
    startCommand: ./app
```

---

## 🚀 Deploy in 10 Seconds

[![Deploy to Render](https://render.com/images/deploy-to-render-button.svg)](https://render.com/deploy)
