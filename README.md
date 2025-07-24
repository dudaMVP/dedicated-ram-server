# Dedicated-RAM-Server

A lightweight Go HTTP API side-project where I explored building REST endpoints from scratch.

## What It Does

- **/memory**: GET and POST handlers to view or set a simulated RAM allocation value  
- **/health**: Simple health check (returns HTTP 200)  
- Built using Go’s `net/http` package — no frameworks  
- Includes request parsing, JSON serialization, and basic error handling

## Why I Built It

I wanted hands-on experience designing and implementing REST APIs in Go — from route setup and request handling to structured JSON responses — without the overhead of a framework.

## Tech Stack

- **Go** (standard library)
- JSON I/O with `encoding/json`
- Built-in router via `http.HandleFunc`
- Runs standalone — no DB or external dependencies

## Getting Started

1. Clone the repo  
   ```bash
   git clone https://github.com/dudaMVP/dedicated-ram-server.git
   cd dedicated-ram-server
   ```

2. Build and run  
   ```bash
   go build -o ram-server
   ./ram-server
   ```

3. Try it out:
   ```bash
   # Check health
   curl http://localhost:8080/health
   # {"status":"ok"}

   # Get current RAM value
   curl http://localhost:8080/memory
   # {"ram":512}

   # Update RAM value
   curl -X POST http://localhost:8080/memory \
     -H "Content-Type: application/json" \
     -d '{"ram":1024}'
   # {"ram":1024}
   ```

## Code Structure

```
./
├── main.go        # Hook up routes and start server
└── handlers.go    # HTTP handlers (memory & health endpoints)
```

Everything lives in one package for simplicity—great for small-scale API learning.

## What's Next?

- Add PUT/PATCH support or DELETE to experiment with other verbs  
- Add input validation (e.g., enforce min/max values)  
- Implement unit and integration tests (using Go’s `testing` package or `httptest`)  
- Extract handlers into their own package(s) as app grows  
- Swap simulated RAM with real metrics via `/metrics` endpoint, using `runtime.MemStats`
