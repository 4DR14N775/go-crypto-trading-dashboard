// Real-time Crypto Trading Dashboard
// A single-file Go application demonstrating SSE (Server-Sent Events)
// with a beautiful Tailwind CSS interface

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

// ============================================================================
// DATA MODELS
// ============================================================================

// Crypto represents a cryptocurrency with its current state
type Crypto struct {
	Symbol    string  `json:"symbol"`
	Name      string  `json:"name"`
	Price     float64 `json:"price"`
	Change24h float64 `json:"change24h"`
	Volume    float64 `json:"volume"`
	High24h   float64 `json:"high24h"`
	Low24h    float64 `json:"low24h"`
}

// Trade represents a single trade transaction
type Trade struct {
	ID        string  `json:"id"`
	Symbol    string  `json:"symbol"`
	Type      string  `json:"type"` // "buy" or "sell"
	Price     float64 `json:"price"`
	Amount    float64 `json:"amount"`
	Total     float64 `json:"total"`
	Timestamp string  `json:"timestamp"`
}

// Alert represents a market alert
type Alert struct {
	ID       string `json:"id"`
	Type     string `json:"type"` // "info", "warning", "success", "danger"
	Title    string `json:"title"`
	Message  string `json:"message"`
	Timestamp string `json:"timestamp"`
}

// SSEMessage wraps different event types for SSE
type SSEMessage struct {
	Event string      `json:"event"`
	Data  interface{} `json:"data"`
}

// ============================================================================
// GLOBAL STATE
// ============================================================================

var (
	// Current crypto prices
	cryptos = map[string]*Crypto{
		"BTC": {Symbol: "BTC", Name: "Bitcoin", Price: 43250.00, Change24h: 2.5, Volume: 28500000000, High24h: 44100, Low24h: 42800},
		"ETH": {Symbol: "ETH", Name: "Ethereum", Price: 2280.00, Change24h: -1.2, Volume: 15200000000, High24h: 2350, Low24h: 2250},
		"SOL": {Symbol: "SOL", Name: "Solana", Price: 98.50, Change24h: 5.8, Volume: 2100000000, High24h: 102, Low24h: 94},
		"ADA": {Symbol: "ADA", Name: "Cardano", Price: 0.52, Change24h: -0.8, Volume: 450000000, High24h: 0.55, Low24h: 0.50},
		"DOT": {Symbol: "DOT", Name: "Polkadot", Price: 7.25, Change24h: 1.3, Volume: 320000000, High24h: 7.50, Low24h: 7.10},
		"AVAX": {Symbol: "AVAX", Name: "Avalanche", Price: 35.80, Change24h: 3.2, Volume: 580000000, High24h: 37.00, Low24h: 34.50},
	}
	
	// Connected SSE clients
	clients   = make(map[chan SSEMessage]bool)
	clientsMu sync.RWMutex
	
	// Statistics
	totalTrades   int64
	totalVolume   float64
	activeTraders int
	statsMu       sync.RWMutex
)

// ============================================================================
// SSE CLIENT MANAGEMENT
// ============================================================================

// addClient registers a new SSE client
func addClient(ch chan SSEMessage) {
	clientsMu.Lock()
	clients[ch] = true
	clientsMu.Unlock()
	
	statsMu.Lock()
	activeTraders++
	statsMu.Unlock()
	
	log.Printf("Client connected. Total clients: %d", len(clients))
}

// removeClient unregisters an SSE client
func removeClient(ch chan SSEMessage) {
	clientsMu.Lock()
	delete(clients, ch)
	close(ch)
	clientsMu.Unlock()
	
	statsMu.Lock()
	activeTraders--
	statsMu.Unlock()
	
	log.Printf("Client disconnected. Total clients: %d", len(clients))
}

// broadcast sends a message to all connected clients
func broadcast(msg SSEMessage) {
	clientsMu.RLock()
	defer clientsMu.RUnlock()
	
	for ch := range clients {
		select {
		case ch <- msg:
		default:
			// Client buffer full, skip
		}
	}
}

// ============================================================================
// MARKET SIMULATION
// ============================================================================

// simulateMarket runs the market simulation in background
func simulateMarket() {
	priceUpdateTicker := time.NewTicker(800 * time.Millisecond)
	tradeTicker := time.NewTicker(1500 * time.Millisecond)
	alertTicker := time.NewTicker(8 * time.Second)
	statsTicker := time.NewTicker(2 * time.Second)
	
	go func() {
		for range priceUpdateTicker.C {
			updatePrices()
		}
	}()
	
	go func() {
		for range tradeTicker.C {
			generateTrade()
		}
	}()
	
	go func() {
		for range alertTicker.C {
			generateAlert()
		}
	}()
	
	go func() {
		for range statsTicker.C {
			broadcastStats()
		}
	}()
}

// updatePrices simulates price changes for all cryptos
func updatePrices() {
	symbols := []string{"BTC", "ETH", "SOL", "ADA", "DOT", "AVAX"}
	
	for _, symbol := range symbols {
		crypto := cryptos[symbol]
		
		// Random price change (-2% to +2%)
		changePercent := (rand.Float64() - 0.5) * 4
		priceChange := crypto.Price * (changePercent / 100)
		crypto.Price += priceChange
		
		// Update 24h change
		crypto.Change24h += (rand.Float64() - 0.5) * 0.5
		crypto.Change24h = math.Max(-20, math.Min(20, crypto.Change24h))
		
		// Update high/low
		if crypto.Price > crypto.High24h {
			crypto.High24h = crypto.Price
		}
		if crypto.Price < crypto.Low24h {
			crypto.Low24h = crypto.Price
		}
		
		// Update volume
		crypto.Volume += rand.Float64() * 10000000
	}
	
	// Broadcast price update
	broadcast(SSEMessage{
		Event: "prices",
		Data:  getCryptoList(),
	})
}

// generateTrade creates a random trade
func generateTrade() {
	symbols := []string{"BTC", "ETH", "SOL", "ADA", "DOT", "AVAX"}
	symbol := symbols[rand.Intn(len(symbols))]
	crypto := cryptos[symbol]
	
	tradeType := "buy"
	if rand.Float32() > 0.5 {
		tradeType = "sell"
	}
	
	amount := rand.Float64() * 10
	if symbol == "BTC" {
		amount = rand.Float64() * 2
	}
	
	trade := Trade{
		ID:        fmt.Sprintf("T%d", time.Now().UnixNano()),
		Symbol:    symbol,
		Type:      tradeType,
		Price:     crypto.Price,
		Amount:    math.Round(amount*10000) / 10000,
		Total:     math.Round(crypto.Price*amount*100) / 100,
		Timestamp: time.Now().Format("15:04:05"),
	}
	
	// Update stats
	statsMu.Lock()
	totalTrades++
	totalVolume += trade.Total
	statsMu.Unlock()
	
	broadcast(SSEMessage{
		Event: "trade",
		Data:  trade,
	})
}

// generateAlert creates random market alerts
func generateAlert() {
	alerts := []Alert{
		{Type: "success", Title: "Whale Alert", Message: "Large BTC transfer detected: 500 BTC moved to exchange"},
		{Type: "warning", Title: "High Volatility", Message: "SOL experiencing unusual price movement"},
		{Type: "info", Title: "Market Update", Message: "Trading volume up 25% in the last hour"},
		{Type: "danger", Title: "Price Alert", Message: "ETH dropped below key support level"},
		{Type: "success", Title: "New ATH", Message: "AVAX reached new all-time high!"},
		{Type: "info", Title: "Network Update", Message: "Ethereum gas fees at 3-month low"},
		{Type: "warning", Title: "Liquidation Alert", Message: "$50M in longs liquidated on BTC"},
		{Type: "success", Title: "Adoption News", Message: "Major institution announces crypto investment"},
	}
	
	alert := alerts[rand.Intn(len(alerts))]
	alert.ID = fmt.Sprintf("A%d", time.Now().UnixNano())
	alert.Timestamp = time.Now().Format("15:04:05")
	
	broadcast(SSEMessage{
		Event: "alert",
		Data:  alert,
	})
}

// broadcastStats sends current statistics
func broadcastStats() {
	statsMu.RLock()
	stats := map[string]interface{}{
		"totalTrades":   totalTrades,
		"totalVolume":   math.Round(totalVolume*100) / 100,
		"activeTraders": activeTraders,
		"timestamp":     time.Now().Format("15:04:05"),
	}
	statsMu.RUnlock()
	
	broadcast(SSEMessage{
		Event: "stats",
		Data:  stats,
	})
}

// getCryptoList returns current crypto data as a slice
func getCryptoList() []Crypto {
	result := make([]Crypto, 0, len(cryptos))
	for _, c := range cryptos {
		result = append(result, *c)
	}
	return result
}

// ============================================================================
// HTTP HANDLERS
// ============================================================================

// handleSSE handles Server-Sent Events connections
func handleSSE(w http.ResponseWriter, r *http.Request) {
	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	
	// Create client channel
	clientChan := make(chan SSEMessage, 10)
	addClient(clientChan)
	
	// Cleanup on disconnect
	defer removeClient(clientChan)
	
	// Send initial data
	initialData := SSEMessage{
		Event: "init",
		Data: map[string]interface{}{
			"cryptos": getCryptoList(),
			"message": "Connected to CryptoStream Live",
		},
	}
	sendSSE(w, initialData)
	
	// Flush the initial data
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}
	
	// Listen for messages
	for {
		select {
		case msg := <-clientChan:
			sendSSE(w, msg)
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
		case <-r.Context().Done():
			return
		}
	}
}

// sendSSE writes an SSE message to the response
func sendSSE(w http.ResponseWriter, msg SSEMessage) {
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}
	fmt.Fprintf(w, "event: message\n")
	fmt.Fprintf(w, "data: %s\n\n", data)
}

// handleHome serves the main HTML page
func handleHome(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, htmlTemplate)
}

// ============================================================================
// MAIN
// ============================================================================

func main() {
	// Seed random number generator
	rand.Seed(time.Now().UnixNano())
	
	// Start market simulation
	go simulateMarket()
	
	// Setup routes
	http.HandleFunc("/", handleHome)
	http.HandleFunc("/events", handleSSE)
	
	// Start server
	port := ":8080"
	log.Printf("🚀 CryptoStream Dashboard starting on http://localhost%s", port)
	log.Printf("📡 SSE endpoint: http://localhost%s/events", port)
	log.Fatal(http.ListenAndServe(port, nil))
}

// ============================================================================
// HTML TEMPLATE
// ============================================================================

const htmlTemplate = `<!DOCTYPE html>
<html lang="en" class="dark">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>CryptoStream - Real-time Trading Dashboard</title>
    <script src="https://cdn.tailwindcss.com"></script>
    <script>
        tailwind.config = {
            darkMode: 'class',
            theme: {
                extend: {
                    animation: {
                        'pulse-fast': 'pulse 0.5s cubic-bezier(0.4, 0, 0.6, 1)',
                        'slide-in': 'slideIn 0.3s ease-out',
                        'fade-in': 'fadeIn 0.5s ease-out',
                    }
                }
            }
        }
    </script>
    <style>
        @keyframes slideIn {
            from { transform: translateX(100%); opacity: 0; }
            to { transform: translateX(0); opacity: 1; }
        }
        @keyframes fadeIn {
            from { opacity: 0; transform: translateY(-10px); }
            to { opacity: 1; transform: translateY(0); }
        }
        .price-up { animation: pulse-green 0.5s ease-out; }
        .price-down { animation: pulse-red 0.5s ease-out; }
        @keyframes pulse-green {
            0%, 100% { background-color: transparent; }
            50% { background-color: rgba(34, 197, 94, 0.2); }
        }
        @keyframes pulse-red {
            0%, 100% { background-color: transparent; }
            50% { background-color: rgba(239, 68, 68, 0.2); }
        }
        .gradient-bg {
            background: linear-gradient(135deg, #1a1a2e 0%, #16213e 50%, #0f3460 100%);
        }
        .glass {
            background: rgba(255, 255, 255, 0.05);
            backdrop-filter: blur(10px);
            border: 1px solid rgba(255, 255, 255, 0.1);
        }
        .scrollbar-thin::-webkit-scrollbar { width: 6px; }
        .scrollbar-thin::-webkit-scrollbar-track { background: rgba(255,255,255,0.1); border-radius: 3px; }
        .scrollbar-thin::-webkit-scrollbar-thumb { background: rgba(255,255,255,0.3); border-radius: 3px; }
    </style>
</head>
<body class="gradient-bg min-h-screen text-white">
    <!-- Header -->
    <header class="glass sticky top-0 z-50">
        <div class="max-w-7xl mx-auto px-4 py-4">
            <div class="flex items-center justify-between">
                <div class="flex items-center space-x-3">
                    <div class="w-10 h-10 bg-gradient-to-br from-yellow-400 to-orange-500 rounded-xl flex items-center justify-center">
                        <svg class="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 7h8m0 0v8m0-8l-8 8-4-4-6 6"></path>
                        </svg>
                    </div>
                    <div>
                        <h1 class="text-xl font-bold bg-gradient-to-r from-yellow-400 to-orange-500 bg-clip-text text-transparent">
                            CryptoStream
                        </h1>
                        <p class="text-xs text-gray-400">Real-time Trading Dashboard</p>
                    </div>
                </div>
                
                <div class="flex items-center space-x-4">
                    <div id="connection-status" class="flex items-center space-x-2 px-3 py-1.5 rounded-full bg-gray-800">
                        <span class="w-2 h-2 rounded-full bg-red-500 animate-pulse" id="status-dot"></span>
                        <span class="text-sm text-gray-400" id="status-text">Connecting...</span>
                    </div>
                    <div class="text-right hidden sm:block">
                        <p class="text-xs text-gray-400">Last Update</p>
                        <p class="text-sm font-mono" id="last-update">--:--:--</p>
                    </div>
                </div>
            </div>
        </div>
    </header>

    <main class="max-w-7xl mx-auto px-4 py-6 space-y-6">
        <!-- Stats Cards -->
        <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
            <div class="glass rounded-2xl p-4">
                <div class="flex items-center justify-between">
                    <div>
                        <p class="text-xs text-gray-400 uppercase tracking-wide">Total Trades</p>
                        <p class="text-2xl font-bold mt-1" id="stat-trades">0</p>
                    </div>
                    <div class="w-12 h-12 bg-blue-500/20 rounded-xl flex items-center justify-center">
                        <svg class="w-6 h-6 text-blue-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"></path>
                        </svg>
                    </div>
                </div>
            </div>
            
            <div class="glass rounded-2xl p-4">
                <div class="flex items-center justify-between">
                    <div>
                        <p class="text-xs text-gray-400 uppercase tracking-wide">Volume (24h)</p>
                        <p class="text-2xl font-bold mt-1" id="stat-volume">$0</p>
                    </div>
                    <div class="w-12 h-12 bg-green-500/20 rounded-xl flex items-center justify-center">
                        <svg class="w-6 h-6 text-green-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path>
                        </svg>
                    </div>
                </div>
            </div>
            
            <div class="glass rounded-2xl p-4">
                <div class="flex items-center justify-between">
                    <div>
                        <p class="text-xs text-gray-400 uppercase tracking-wide">Active Traders</p>
                        <p class="text-2xl font-bold mt-1" id="stat-traders">0</p>
                    </div>
                    <div class="w-12 h-12 bg-purple-500/20 rounded-xl flex items-center justify-center">
                        <svg class="w-6 h-6 text-purple-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z"></path>
                        </svg>
                    </div>
                </div>
            </div>
            
            <div class="glass rounded-2xl p-4">
                <div class="flex items-center justify-between">
                    <div>
                        <p class="text-xs text-gray-400 uppercase tracking-wide">Market Status</p>
                        <p class="text-2xl font-bold mt-1 text-green-400">LIVE</p>
                    </div>
                    <div class="w-12 h-12 bg-yellow-500/20 rounded-xl flex items-center justify-center">
                        <svg class="w-6 h-6 text-yellow-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z"></path>
                        </svg>
                    </div>
                </div>
            </div>
        </div>

        <!-- Main Content Grid -->
        <div class="grid lg:grid-cols-3 gap-6">
            <!-- Crypto Prices Table -->
            <div class="lg:col-span-2 glass rounded-2xl overflow-hidden">
                <div class="p-4 border-b border-white/10">
                    <h2 class="text-lg font-semibold flex items-center">
                        <svg class="w-5 h-5 mr-2 text-yellow-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 7h8m0 0v8m0-8l-8 8-4-4-6 6"></path>
                        </svg>
                        Live Prices
                    </h2>
                </div>
                <div class="overflow-x-auto">
                    <table class="w-full">
                        <thead class="bg-white/5">
                            <tr class="text-left text-xs text-gray-400 uppercase tracking-wider">
                                <th class="px-4 py-3">Asset</th>
                                <th class="px-4 py-3">Price</th>
                                <th class="px-4 py-3">24h Change</th>
                                <th class="px-4 py-3 hidden sm:table-cell">24h High</th>
                                <th class="px-4 py-3 hidden sm:table-cell">24h Low</th>
                                <th class="px-4 py-3 hidden md:table-cell">Volume</th>
                            </tr>
                        </thead>
                        <tbody id="crypto-table" class="divide-y divide-white/5">
                            <!-- Crypto rows will be inserted here -->
                        </tbody>
                    </table>
                </div>
            </div>

            <!-- Alerts Panel -->
            <div class="glass rounded-2xl overflow-hidden flex flex-col">
                <div class="p-4 border-b border-white/10">
                    <h2 class="text-lg font-semibold flex items-center">
                        <svg class="w-5 h-5 mr-2 text-red-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9"></path>
                        </svg>
                        Market Alerts
                    </h2>
                </div>
                <div id="alerts-container" class="flex-1 p-4 space-y-3 overflow-y-auto max-h-80 scrollbar-thin">
                    <div class="text-center text-gray-500 py-8">
                        <svg class="w-12 h-12 mx-auto mb-2 opacity-50" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20 13V6a2 2 0 00-2-2H6a2 2 0 00-2 2v7m16 0v5a2 2 0 01-2 2H6a2 2 0 01-2-2v-5m16 0h-2.586a1 1 0 00-.707.293l-2.414 2.414a1 1 0 01-.707.293h-3.172a1 1 0 01-.707-.293l-2.414-2.414A1 1 0 006.586 13H4"></path>
                        </svg>
                        <p>Waiting for alerts...</p>
                    </div>
                </div>
            </div>
        </div>

        <!-- Live Trades Feed -->
        <div class="glass rounded-2xl overflow-hidden">
            <div class="p-4 border-b border-white/10">
                <h2 class="text-lg font-semibold flex items-center">
                    <svg class="w-5 h-5 mr-2 text-green-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7h12m0 0l-4-4m4 4l-4 4m0 6H4m0 0l4 4m-4-4l4-4"></path>
                    </svg>
                    Live Trade Feed
                    <span class="ml-2 px-2 py-0.5 text-xs bg-green-500/20 text-green-400 rounded-full animate-pulse">LIVE</span>
                </h2>
            </div>
            <div id="trades-container" class="p-4 overflow-x-auto">
                <div class="flex space-x-4 min-w-max" id="trades-feed">
                    <!-- Trade cards will be inserted here -->
                </div>
            </div>
        </div>

        <!-- Price History Charts -->
        <div class="glass rounded-2xl overflow-hidden">
            <div class="p-4 border-b border-white/10">
                <h2 class="text-lg font-semibold flex items-center">
                    <svg class="w-5 h-5 mr-2 text-blue-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 12l3-3 3 3 4-4M8 21l4-4 4 4M3 4h18M4 4h16v12a1 1 0 01-1 1H5a1 1 0 01-1-1V4z"></path>
                    </svg>
                    Price Sparklines
                </h2>
            </div>
            <div class="p-4 grid grid-cols-2 md:grid-cols-3 lg:grid-cols-6 gap-4" id="sparklines-container">
                <!-- Sparkline charts will be inserted here -->
            </div>
        </div>
    </main>

    <!-- Footer -->
    <footer class="glass mt-8">
        <div class="max-w-7xl mx-auto px-4 py-4 text-center text-sm text-gray-400">
            <p>Built with Go + SSE + Tailwind CSS | Real-time streaming demo</p>
            <p class="mt-1 text-xs">All data is simulated for demonstration purposes</p>
        </div>
    </footer>

    <script>
        // ========================================================================
        // STATE MANAGEMENT
        // ========================================================================
        
        const state = {
            prices: {},
            priceHistory: {},
            trades: [],
            connected: false
        };
        
        const MAX_TRADES = 20;
        const MAX_HISTORY = 50;
        
        // Crypto icons (emoji placeholders)
        const cryptoIcons = {
            'BTC': '₿',
            'ETH': 'Ξ',
            'SOL': '◎',
            'ADA': '₳',
            'DOT': '●',
            'AVAX': '▲'
        };
        
        const cryptoColors = {
            'BTC': 'from-orange-400 to-yellow-500',
            'ETH': 'from-blue-400 to-indigo-500',
            'SOL': 'from-purple-400 to-pink-500',
            'ADA': 'from-blue-500 to-cyan-400',
            'DOT': 'from-pink-500 to-rose-500',
            'AVAX': 'from-red-500 to-orange-500'
        };
        
        // ========================================================================
        // FORMATTING UTILITIES
        // ========================================================================
        
        function formatPrice(price) {
            if (price >= 1000) {
                return '$' + price.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 });
            } else if (price >= 1) {
                return '$' + price.toFixed(2);
            } else {
                return '$' + price.toFixed(4);
            }
        }
        
        function formatVolume(volume) {
            if (volume >= 1e9) return '$' + (volume / 1e9).toFixed(2) + 'B';
            if (volume >= 1e6) return '$' + (volume / 1e6).toFixed(2) + 'M';
            if (volume >= 1e3) return '$' + (volume / 1e3).toFixed(2) + 'K';
            return '$' + volume.toFixed(2);
        }
        
        function formatChange(change) {
            const sign = change >= 0 ? '+' : '';
            return sign + change.toFixed(2) + '%';
        }
        
        // ========================================================================
        // UI RENDERING
        // ========================================================================
        
        function updateCryptoTable(cryptos) {
            const table = document.getElementById('crypto-table');
            
            cryptos.forEach(crypto => {
                const oldPrice = state.prices[crypto.symbol]?.price || crypto.price;
                const priceChanged = oldPrice !== crypto.price;
                const priceUp = crypto.price > oldPrice;
                
                // Update price history
                if (!state.priceHistory[crypto.symbol]) {
                    state.priceHistory[crypto.symbol] = [];
                }
                state.priceHistory[crypto.symbol].push(crypto.price);
                if (state.priceHistory[crypto.symbol].length > MAX_HISTORY) {
                    state.priceHistory[crypto.symbol].shift();
                }
                
                state.prices[crypto.symbol] = crypto;
                
                let row = document.getElementById('crypto-' + crypto.symbol);
                const changeClass = crypto.change24h >= 0 ? 'text-green-400' : 'text-red-400';
                const changeIcon = crypto.change24h >= 0 ? '↑' : '↓';
                
                const rowHTML = ` + "`" + `
                    <td class="px-4 py-4">
                        <div class="flex items-center space-x-3">
                            <div class="w-10 h-10 rounded-xl bg-gradient-to-br ${cryptoColors[crypto.symbol]} flex items-center justify-center text-lg font-bold">
                                ${cryptoIcons[crypto.symbol]}
                            </div>
                            <div>
                                <p class="font-semibold">${crypto.symbol}</p>
                                <p class="text-xs text-gray-400">${crypto.name}</p>
                            </div>
                        </div>
                    </td>
                    <td class="px-4 py-4">
                        <p class="font-mono font-semibold text-lg">${formatPrice(crypto.price)}</p>
                    </td>
                    <td class="px-4 py-4">
                        <span class="${changeClass} font-semibold flex items-center">
                            ${changeIcon} ${formatChange(crypto.change24h)}
                        </span>
                    </td>
                    <td class="px-4 py-4 hidden sm:table-cell text-gray-400 font-mono">
                        ${formatPrice(crypto.high24h)}
                    </td>
                    <td class="px-4 py-4 hidden sm:table-cell text-gray-400 font-mono">
                        ${formatPrice(crypto.low24h)}
                    </td>
                    <td class="px-4 py-4 hidden md:table-cell text-gray-400">
                        ${formatVolume(crypto.volume)}
                    </td>
                ` + "`" + `;
                
                if (!row) {
                    row = document.createElement('tr');
                    row.id = 'crypto-' + crypto.symbol;
                    row.className = 'hover:bg-white/5 transition-all duration-200';
                    table.appendChild(row);
                }
                
                row.innerHTML = rowHTML;
                
                if (priceChanged) {
                    row.classList.add(priceUp ? 'price-up' : 'price-down');
                    setTimeout(() => {
                        row.classList.remove('price-up', 'price-down');
                    }, 500);
                }
            });
            
            updateSparklines();
        }
        
        function addTrade(trade) {
            state.trades.unshift(trade);
            if (state.trades.length > MAX_TRADES) {
                state.trades.pop();
            }
            
            const feed = document.getElementById('trades-feed');
            const isBuy = trade.type === 'buy';
            
            const card = document.createElement('div');
            card.className = ` + "`" + `flex-shrink-0 w-48 p-3 rounded-xl ${isBuy ? 'bg-green-500/10 border border-green-500/30' : 'bg-red-500/10 border border-red-500/30'} animate-slide-in` + "`" + `;
            
            card.innerHTML = ` + "`" + `
                <div class="flex items-center justify-between mb-2">
                    <span class="font-bold ${isBuy ? 'text-green-400' : 'text-red-400'}">${trade.type.toUpperCase()}</span>
                    <span class="text-xs text-gray-400">${trade.timestamp}</span>
                </div>
                <div class="flex items-center space-x-2 mb-2">
                    <div class="w-8 h-8 rounded-lg bg-gradient-to-br ${cryptoColors[trade.symbol]} flex items-center justify-center text-sm">
                        ${cryptoIcons[trade.symbol]}
                    </div>
                    <span class="font-semibold">${trade.symbol}</span>
                </div>
                <div class="text-sm space-y-1">
                    <p class="text-gray-400">Amount: <span class="text-white">${trade.amount}</span></p>
                    <p class="text-gray-400">Total: <span class="text-white font-mono">${formatPrice(trade.total)}</span></p>
                </div>
            ` + "`" + `;
            
            feed.insertBefore(card, feed.firstChild);
            
            // Remove old trades
            while (feed.children.length > MAX_TRADES) {
                feed.removeChild(feed.lastChild);
            }
        }
        
        function addAlert(alert) {
            const container = document.getElementById('alerts-container');
            
            // Remove placeholder if exists
            if (container.querySelector('.text-center')) {
                container.innerHTML = '';
            }
            
            const alertColors = {
                'success': 'bg-green-500/10 border-green-500/30 text-green-400',
                'warning': 'bg-yellow-500/10 border-yellow-500/30 text-yellow-400',
                'danger': 'bg-red-500/10 border-red-500/30 text-red-400',
                'info': 'bg-blue-500/10 border-blue-500/30 text-blue-400'
            };
            
            const alertIcons = {
                'success': '✓',
                'warning': '⚠',
                'danger': '✕',
                'info': 'ℹ'
            };
            
            const alertEl = document.createElement('div');
            alertEl.className = ` + "`" + `p-3 rounded-xl border ${alertColors[alert.type]} animate-fade-in` + "`" + `;
            
            alertEl.innerHTML = ` + "`" + `
                <div class="flex items-start space-x-3">
                    <span class="text-lg">${alertIcons[alert.type]}</span>
                    <div class="flex-1 min-w-0">
                        <div class="flex items-center justify-between">
                            <p class="font-semibold text-sm">${alert.title}</p>
                            <span class="text-xs text-gray-500">${alert.timestamp}</span>
                        </div>
                        <p class="text-xs text-gray-400 mt-1">${alert.message}</p>
                    </div>
                </div>
            ` + "`" + `;
            
            container.insertBefore(alertEl, container.firstChild);
            
            // Keep only last 10 alerts
            while (container.children.length > 10) {
                container.removeChild(container.lastChild);
            }
        }
        
        function updateStats(stats) {
            document.getElementById('stat-trades').textContent = stats.totalTrades.toLocaleString();
            document.getElementById('stat-volume').textContent = formatVolume(stats.totalVolume);
            document.getElementById('stat-traders').textContent = stats.activeTraders;
            document.getElementById('last-update').textContent = stats.timestamp;
        }
        
        function updateSparklines() {
            const container = document.getElementById('sparklines-container');
            container.innerHTML = '';
            
            Object.keys(state.priceHistory).forEach(symbol => {
                const history = state.priceHistory[symbol];
                if (history.length < 2) return;
                
                const crypto = state.prices[symbol];
                const sparkline = createSparkline(history, crypto.change24h >= 0);
                
                const card = document.createElement('div');
                card.className = 'glass rounded-xl p-3';
                card.innerHTML = ` + "`" + `
                    <div class="flex items-center space-x-2 mb-2">
                        <div class="w-6 h-6 rounded-lg bg-gradient-to-br ${cryptoColors[symbol]} flex items-center justify-center text-xs">
                            ${cryptoIcons[symbol]}
                        </div>
                        <span class="font-semibold text-sm">${symbol}</span>
                        <span class="text-xs ${crypto.change24h >= 0 ? 'text-green-400' : 'text-red-400'} ml-auto">
                            ${formatChange(crypto.change24h)}
                        </span>
                    </div>
                    ${sparkline}
                ` + "`" + `;
                
                container.appendChild(card);
            });
        }
        
        function createSparkline(data, isPositive) {
            const width = 120;
            const height = 40;
            const min = Math.min(...data);
            const max = Math.max(...data);
            const range = max - min || 1;
            
            const points = data.map((value, index) => {
                const x = (index / (data.length - 1)) * width;
                const y = height - ((value - min) / range) * height;
                return ` + "`" + `${x},${y}` + "`" + `;
            }).join(' ');
            
            const color = isPositive ? '#22c55e' : '#ef4444';
            
            return ` + "`" + `
                <svg width="100%" height="${height}" viewBox="0 0 ${width} ${height}" preserveAspectRatio="none">
                    <defs>
                        <linearGradient id="gradient-${isPositive}" x1="0%" y1="0%" x2="0%" y2="100%">
                            <stop offset="0%" style="stop-color:${color};stop-opacity:0.3" />
                            <stop offset="100%" style="stop-color:${color};stop-opacity:0" />
                        </linearGradient>
                    </defs>
                    <polygon points="0,${height} ${points} ${width},${height}" fill="url(#gradient-${isPositive})" />
                    <polyline points="${points}" fill="none" stroke="${color}" stroke-width="2" />
                </svg>
            ` + "`" + `;
        }
        
        function setConnectionStatus(connected) {
            const dot = document.getElementById('status-dot');
            const text = document.getElementById('status-text');
            
            state.connected = connected;
            
            if (connected) {
                dot.className = 'w-2 h-2 rounded-full bg-green-500';
                text.textContent = 'Connected';
                text.className = 'text-sm text-green-400';
            } else {
                dot.className = 'w-2 h-2 rounded-full bg-red-500 animate-pulse';
                text.textContent = 'Disconnected';
                text.className = 'text-sm text-red-400';
            }
        }
        
        // ========================================================================
        // SSE CONNECTION
        // ========================================================================
        
        function connectSSE() {
            console.log('Connecting to SSE...');
            const eventSource = new EventSource('/events');
            
            eventSource.onopen = function() {
                console.log('SSE Connected');
                setConnectionStatus(true);
            };
            
            eventSource.onerror = function(e) {
                console.error('SSE Error:', e);
                setConnectionStatus(false);
                
                // Reconnect after 3 seconds
                setTimeout(() => {
                    console.log('Attempting to reconnect...');
                    connectSSE();
                }, 3000);
            };
            
            eventSource.onmessage = function(e) {
                try {
                    const message = JSON.parse(e.data);
                    handleMessage(message);
                } catch (err) {
                    console.error('Error parsing message:', err);
                }
            };
        }
        
        function handleMessage(message) {
            switch (message.event) {
                case 'init':
                    console.log('Received initial data');
                    updateCryptoTable(message.data.cryptos);
                    break;
                    
                case 'prices':
                    updateCryptoTable(message.data);
                    break;
                    
                case 'trade':
                    addTrade(message.data);
                    break;
                    
                case 'alert':
                    addAlert(message.data);
                    break;
                    
                case 'stats':
                    updateStats(message.data);
                    break;
                    
                default:
                    console.log('Unknown event:', message.event);
            }
        }
        
        // ========================================================================
        // INITIALIZATION
        // ========================================================================
        
        document.addEventListener('DOMContentLoaded', function() {
            connectSSE();
        });
    </script>
</body>
</html>`
