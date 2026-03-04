# Apisto — Contributor Onboarding Guide

This document takes you from "I barely know Go" to "I can add a feature without help." It is specific to this codebase. Every concept is explained in terms of code that actually exists here.

---

## Table of Contents

1. [What the repo does in one paragraph](#1-what-the-repo-does)
2. [The tech stack — what each piece is and why](#2-the-tech-stack)
3. [Go language concepts you will encounter](#3-go-concepts-you-will-encounter)
4. [External services and protocols](#4-external-services-and-protocols)
5. [How the codebase is structured](#5-how-the-codebase-is-structured)
6. [A full data flow walkthrough](#6-a-full-data-flow-walkthrough)
7. [Line-by-line explanations of key patterns](#7-line-by-line-explanations-of-key-patterns)
8. [How to make a change — worked examples](#8-how-to-make-a-change)
9. [Running the stack locally](#9-running-the-stack-locally)
10. [Where to read next](#10-where-to-read-next)

---

## 1. What the repo does

An ESP32 board sends sensor readings (temperature, humidity, etc.) over MQTT to a Mosquitto broker. The Go server subscribes to Mosquitto, validates the device token, writes the data into TimescaleDB (a time-series Postgres extension), and pushes it over WebSocket to any browser that has the dashboard open. The browser renders auto-generated widgets. You can also send commands back to the device from the browser. Everything runs with `docker compose up`.

---

## 2. The tech stack

### Go (the server language)
Go is a compiled, statically typed language made by Google. It is fast, has excellent concurrency primitives built in, and compiles to a single binary with no runtime dependencies. That last point is why it was chosen — the whole server ships as one file.

**Docs:** https://go.dev/doc/

**The one thing to understand first:** Go has no classes. Instead you define `struct` types and attach functions (called *methods*) to them. Everything else flows from that.

---

### Fiber v2 (HTTP framework)
Fiber is the web framework. It handles incoming HTTP requests, routes them to the right handler function, and manages middleware (like CORS and logging). It is similar to Express.js if you have used Node.

**Docs:** https://docs.gofiber.io/

**Where it lives in this repo:**
- `internal/api/router.go` — where all routes are registered
- `internal/api/*.go` — the handler functions
- `internal/api/middleware.go` — CORS, logging, error handling

**Key concept — a handler function signature:**
```go
func (h *projectHandlers) create(c *fiber.Ctx) error {
    // c is the request/response context
    // return nil = success, return an error = Fiber sends 500
}
```

---

### pgx v5 + pgxpool (PostgreSQL driver)
pgx is how Go talks to PostgreSQL. pgxpool manages a *connection pool* — a set of reusable database connections so we are not opening and closing a connection on every request.

**Docs:** https://pkg.go.dev/github.com/jackc/pgx/v5

**Where it lives:** `internal/database/postgres.go` (pool setup), used throughout all services as `s.db.Pool.Query(...)` or `s.db.Pool.Exec(...)`.

**Key concept — parameterised queries:**
```go
s.db.Pool.Exec(ctx, `UPDATE devices SET is_online = true WHERE id = $1`, deviceID)
// $1, $2, $3... are placeholders — never build SQL with string concatenation
```

---

### TimescaleDB (time-series database)
TimescaleDB is a PostgreSQL *extension* — it adds time-series superpowers to regular Postgres. You write normal SQL. It handles partitioning data by time automatically.

**Docs:** https://docs.timescale.com/

**Why it matters for this project:**
- The `telemetry` table is a *hypertable* — Timescale automatically splits it into daily chunks for fast inserts and queries
- `time_bucket('1 hour', time)` is a Timescale function that rounds timestamps to the nearest hour — this is how the chart aggregations work
- Continuous aggregates (`telemetry_hourly`, `telemetry_daily`) are pre-computed summaries that make long-range queries instant
- Retention policy automatically drops data older than 30 days

**Where it lives:** `internal/database/migrations/001_init.up.sql`, `002_policies.up.sql`

---

### golang-migrate (database migrations)
Migrations are versioned SQL files that describe how to evolve the database schema. golang-migrate runs them in order and tracks which ones have already been applied.

**Docs:** https://github.com/golang-migrate/migrate

**Where it lives:** `internal/database/postgres.go` (`RunMigrations`), files in `internal/database/migrations/`

**How it works:** On every server start, `RunMigrations` runs. It checks a `schema_migrations` table in Postgres to see which migrations have run, then applies any new ones. If nothing is new, it does nothing.

---

### paho.mqtt.golang (MQTT client)
The Go server does not *run* an MQTT broker — Mosquitto does that. Instead the server connects to Mosquitto *as a client* and subscribes to device topics. Paho is the library that implements the MQTT client protocol.

**Docs:** https://pkg.go.dev/github.com/eclipse/paho.mqtt.golang

**Where it lives:** `internal/mqtt/client.go`

---

### gofiber/contrib/websocket (WebSocket)
WebSocket is a protocol that keeps a persistent two-way connection open between the browser and the server. Unlike HTTP (request → response → done), a WebSocket stays alive so the server can push data whenever it wants. This is how the dashboard updates in real time without polling.

**Docs:** https://github.com/gofiber/contrib/tree/main/websocket

**Where it lives:** `internal/api/websocket.go`, `internal/services/realtime_service.go`

---

### Docker + Docker Compose
Docker packages applications into containers. Docker Compose orchestrates multiple containers together.

The `docker-compose.yml` defines three services:
- `timescaledb` — the database
- `mosquitto` — the MQTT broker
- `apisto` — the Go server (built from `Dockerfile`)

**Docs:**
- https://docs.docker.com/compose/
- https://docs.docker.com/reference/dockerfile/

---

### React + Vite + Tailwind (frontend)
The frontend is a standard React SPA. Vite is the build tool (replaces Create React App, much faster). Tailwind is a utility-first CSS framework — instead of writing CSS files you add classes like `text-sm font-bold text-zinc-400` directly to elements.

**Docs:**
- https://react.dev/
- https://vitejs.dev/
- https://tailwindcss.com/docs

**Where it lives:** `web/src/`

---

## 3. Go concepts you will encounter

This section explains every Go pattern you will see in this codebase, with the actual file and line as context.

### Packages

Every `.go` file starts with `package something`. Files in the same directory must share the same package name. The package name determines how other packages import and use it.

```go
// internal/config/config.go
package config     // ← this file belongs to the "config" package

// internal/server/server.go
import "github.com/balub/apisto/internal/config"  // ← import it like this
cfg := config.Load()                               // ← use it like this
```

The top-level module name (`github.com/balub/apisto`) is defined in `go.mod`. Every internal import starts with that prefix.

---

### Structs — the main building block

A struct is a collection of named fields. Think of it like an object with data but no behaviour yet.

```go
// internal/models/device.go
type Device struct {
    ID        string    `json:"id"`
    ProjectID string    `json:"project_id"`
    IsOnline  bool      `json:"is_online"`
    LastSeenAt *time.Time `json:"last_seen_at"`
}
```

The backtick annotations (`json:"id"`) are called *struct tags*. They tell the `encoding/json` package what name to use when converting this struct to/from JSON. So `ID` becomes `"id"` in JSON.

---

### Methods — behaviour attached to structs

A method is a function with a *receiver* — the struct it belongs to. This is Go's version of a class method.

```go
// internal/services/device_service.go
func (s *DeviceService) GetByID(ctx context.Context, id string) (*Device, error) {
//   ↑ receiver: "s" is the DeviceService instance this method runs on
}
```

`(s *DeviceService)` means: this method belongs to `DeviceService`, and `s` is how you refer to the instance inside the function (like `this` or `self` in other languages). The `*` means it's a pointer — explained below.

---

### Pointers

A pointer holds the *memory address* of a value rather than the value itself. `*T` is a pointer to a value of type `T`.

```go
var x int = 5
var p *int = &x   // p points to x. & means "address of"
*p = 10           // * means "dereference" — follow the pointer. Now x == 10
```

**Why you need to know this for this codebase:**

1. **Methods on structs use pointer receivers** (`*DeviceService` not `DeviceService`) so the method can modify the struct and so Go doesn't copy the entire struct on every call.

2. **Nullable fields use pointers.** `LastSeenAt *time.Time` means it can be `nil` (not set). `LastSeenAt time.Time` (no `*`) can never be nil. You'll see this a lot in the models.

3. **`New` functions return pointers:**
```go
func NewDeviceService(db *database.DB, realtime *RealtimeService) *DeviceService {
    return &DeviceService{db: db, realtime: realtime}
    //     ↑ & creates a pointer to the new struct
}
```

---

### Error handling

Go does not have exceptions. Functions that can fail return two values: the result and an error. You always check it.

```go
device, err := deviceSvc.GetByID(ctx, id)
if err != nil {
    return errResponse(c, 404, "device not found", "NOT_FOUND")
}
// if we get here, device is valid
```

**Wrapping errors** with `fmt.Errorf("...: %w", err)` adds context while keeping the original error:
```go
return nil, fmt.Errorf("get device: %w", err)
// If err was "no rows", this becomes "get device: no rows"
```

---

### Context (`context.Context`)

`context.Context` carries deadlines, cancellation signals, and request-scoped values. Almost every function that does I/O (database, HTTP, MQTT) takes a `ctx` as its first argument.

```go
func (s *DeviceService) GetByID(ctx context.Context, id string) (*Device, error) {
    s.db.Pool.QueryRow(ctx, `SELECT ...`, id)
    //                ↑ if ctx is cancelled, the query aborts automatically
}
```

In `server.go`:
```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel()
// When cancel() is called (on shutdown), all in-progress operations abort cleanly
```

---

### Goroutines — concurrent execution

A goroutine is a lightweight thread. You start one with the `go` keyword:

```go
go func() {
    <-quit          // wait for shutdown signal
    app.Shutdown()  // then shut down
}()
// This runs concurrently — the line after this executes immediately
```

This is how the server listens for shutdown signals without blocking the main thread.

---

### Channels — communication between goroutines

A channel is a typed pipe between goroutines. `chan T` is a channel that carries values of type `T`.

```go
// internal/mqtt/client.go
telemetryCh: make(chan TelemetryMessage, 512)
//                                      ↑ buffered: can hold 512 messages before blocking
```

**Sending and receiving:**
```go
telemetryCh <- msg   // send msg into the channel (blocks if full)
msg := <-telemetryCh // receive from channel (blocks until something is there)
```

**Select** — wait on multiple channels simultaneously:
```go
select {
case <-ctx.Done():      // if context cancelled → shut down
    return
case msg := <-ch:       // if message arrives → process it
    process(msg)
}
```

This is the core of the MQTT worker pool in `internal/mqtt/client.go`. Four goroutines all read from the same `telemetryCh` channel — whichever is free picks up the next message.

---

### Defer

`defer` runs a function call when the surrounding function returns, no matter how it returns (normal return, error, panic).

```go
rows, err := db.Pool.Query(ctx, `SELECT ...`)
defer rows.Close()  // ← always runs when function exits, even if an error happens below
```

It's used for cleanup: closing database rows, unlocking mutexes, disconnecting clients.

---

### sync.RWMutex — safe concurrent access

If multiple goroutines read and write the same data simultaneously, you get data races (corruption, crashes). A mutex prevents this.

```go
// internal/services/realtime_service.go
type RealtimeService struct {
    mu   sync.RWMutex
    hubs map[string]map[*websocket.Conn]bool
}

func (s *RealtimeService) Register(deviceID string, conn *websocket.Conn) {
    s.mu.Lock()         // only one goroutine can write at a time
    defer s.mu.Unlock() // always unlock when function exits
    s.hubs[deviceID][conn] = true
}

func (s *RealtimeService) Broadcast(deviceID string, msg WSMessage) {
    s.mu.RLock()         // multiple goroutines can read simultaneously
    defer s.mu.RUnlock()
    // ... read hubs ...
}
```

`Lock` for writes, `RLock` for reads. Reads don't block each other; a write blocks everything.

---

### sync.Map — concurrent-safe map

A regular Go `map` crashes if multiple goroutines read/write it simultaneously. `sync.Map` is a built-in concurrent-safe version:

```go
// internal/mqtt/client.go
tokenCache sync.Map

// Store a value
tokenCache.Store("abc123", cacheEntry{deviceID: "xyz", expiresAt: ...})

// Load a value (returns interface{}, bool)
if entry, ok := tokenCache.Load("abc123"); ok {
    e := entry.(tokenCacheEntry)  // type assertion — explained below
}

// Delete
tokenCache.Delete("abc123")
```

---

### Type assertions

`sync.Map` stores `interface{}` (any type). To get back the specific type, use a type assertion:

```go
entry, ok := tokenCache.Load(token)
if ok {
    e := entry.(tokenCacheEntry)  // "I assert entry is of type tokenCacheEntry"
    // if wrong type: panic. Use two-value form to be safe:
    e, ok := entry.(tokenCacheEntry)
}
```

---

### go:embed — bake files into the binary

```go
// internal/database/postgres.go
//go:embed migrations/*.sql
var migrations embed.FS
```

This compiler directive tells Go to read all `*.sql` files in `migrations/` at compile time and bundle them into the binary. The migration files don't need to exist on the server at runtime — they're inside the binary itself.

---

### Anonymous functions and closures

Functions are first-class values in Go. You can create them inline and pass them around:

```go
// internal/server/server.go
lookupToken := func(ctx context.Context, token string) (string, error) {
    device, err := deviceSvc.GetByToken(ctx, token)
    // ...
    return device.ID, nil
}
// lookupToken is now a variable holding a function
mqttClient, _ := mqtt.NewClient(..., lookupToken, ...)
```

This is a *closure* — the anonymous function captures `deviceSvc` from the surrounding scope. It remembers it even after `Run()` returns.

---

### Interfaces

An interface defines a set of method signatures. Any type that has those methods satisfies the interface automatically — no explicit `implements` keyword.

```go
// In telemetry_service.go, scanAggResults accepts an interface:
func scanAggResults(rows interface {
    Next() bool
    Scan(...interface{}) error
    Close()
}) ([]*models.TelemetryQueryResult, error) { ... }
```

This lets the function work with any type that has `Next()`, `Scan()`, and `Close()` — both `pgx.Rows` and other row types work without changes.

---

## 4. External services and protocols

### MQTT

MQTT is a lightweight publish/subscribe messaging protocol designed for IoT devices. Instead of making HTTP requests, devices *publish* messages to *topics* (like `apisto/abc123/telemetry`). Other clients *subscribe* to topics and receive those messages.

**Key terms:**
- **Broker** — the message router. Mosquitto is the broker here.
- **Client** — anything that connects to the broker. Both the ESP32 and the Go server are clients.
- **Topic** — a string path like `apisto/{token}/telemetry`. Uses `/` as separator, `+` as single-level wildcard, `#` as multi-level wildcard.
- **QoS** — Quality of Service. `0` = fire and forget, `1` = at least once (we use this), `2` = exactly once.

**In this repo:** The Go server subscribes to `apisto/+/telemetry` — the `+` matches any device token. When a message arrives, it extracts the token from the topic string.

**Interactive MQTT explorer:** https://mqtt-explorer.com/ — useful for debugging

---

### WebSocket

WebSocket is an HTTP upgrade that turns a request-response connection into a persistent two-way channel. Once upgraded, the server can push data to the browser at any time without the browser asking.

**How it works in this repo:**
1. Browser opens `ws://localhost:8080/api/v1/devices/{id}/ws`
2. Fiber upgrades the connection and calls `ws.handle`
3. `RealtimeService.Register` adds this connection to `hubs[deviceID]`
4. When MQTT telemetry arrives → `Broadcast` → data pushed to all connections for that device

---

### TimescaleDB key concepts

**Hypertable:** A regular Postgres table declared as a hypertable via `SELECT create_hypertable('telemetry', 'time', ...)`. TimescaleDB then automatically splits the data into *chunks* by time (1 day chunks here). Inserts are fast because they always go into the current chunk. Time-range queries are fast because TimescaleDB only reads the relevant chunks.

**time_bucket:** A TimescaleDB SQL function that floors a timestamp to a boundary:
```sql
time_bucket('1 hour', time)  -- 10:47:22 → 10:00:00
time_bucket('1 day', time)   -- 2024-03-15 14:30 → 2024-03-15 00:00
```
Used in `queryAggregated` in `telemetry_service.go` to group readings into hourly averages for charts.

**Continuous aggregates:** Pre-computed materialised views that TimescaleDB updates automatically in the background. `telemetry_hourly` and `telemetry_daily` are queried for longer time ranges instead of re-aggregating raw data every time.

**Retention policy:** `SELECT add_retention_policy('telemetry', INTERVAL '30 days')` tells TimescaleDB to drop chunks older than 30 days. No cron jobs needed.

---

## 5. How the codebase is structured

```
apisto/
├── cmd/apisto/main.go          ← binary entry point (tiny — just calls server.Run)
│
├── internal/                   ← all private application code
│   ├── config/config.go        ← reads environment variables, returns Config struct
│   ├── database/
│   │   ├── postgres.go         ← creates the connection pool, runs migrations
│   │   └── migrations/         ← SQL files applied in order on startup
│   ├── models/                 ← plain Go structs (no database logic, no methods)
│   │   ├── project.go
│   │   ├── device.go
│   │   ├── telemetry.go
│   │   └── command.go
│   ├── services/               ← business logic (database queries + side effects)
│   │   ├── device_service.go
│   │   ├── telemetry_service.go
│   │   ├── command_service.go
│   │   └── realtime_service.go ← WebSocket hub
│   ├── mqtt/client.go          ← MQTT connection, subscriptions, worker pool
│   ├── api/                    ← HTTP handlers (thin — delegate to services)
│   │   ├── router.go           ← route definitions
│   │   ├── middleware.go       ← CORS, logging, error handler
│   │   ├── projects.go
│   │   ├── devices.go
│   │   ├── telemetry.go
│   │   ├── commands.go
│   │   ├── websocket.go
│   │   └── shares.go
│   ├── auth/token.go           ← generates random hex tokens
│   └── server/server.go        ← wires everything together, starts HTTP server
│
├── web/src/                    ← React frontend
│   ├── lib/api.js              ← all fetch() calls to the backend
│   ├── hooks/
│   │   ├── useWebSocket.js     ← WebSocket connection with auto-reconnect
│   │   └── useDeviceData.js    ← fetches device + streams live data
│   ├── pages/                  ← one file per route
│   └── components/             ← reusable UI pieces
│       └── widgets/            ← dashboard widgets (ValueCard, LineChart, etc.)
│
├── sdk/arduino/                ← ESP32/ESP8266 library
├── mosquitto/mosquitto.conf    ← MQTT broker config
├── docker-compose.yml
├── Dockerfile
└── go.mod                      ← Go module definition + dependencies
```

### The layering rule

Code only flows *downward*. Higher layers call lower layers, never the reverse:

```
cmd (entry point)
  └── server (orchestrator)
        ├── api (HTTP handlers)  → call services
        ├── mqtt (client)        → calls services
        └── services             → call database
              └── database       → calls PostgreSQL
```

The `api` layer never talks to the database directly. The `database` layer never knows about HTTP. This makes it easy to change one layer without breaking others.

---

## 6. A full data flow walkthrough

Let's follow a single temperature reading from ESP32 to the browser chart.

### Step 1 — ESP32 publishes to MQTT

```
ESP32 calls device.send("temperature", 24.5)
→ Arduino library publishes JSON to MQTT topic: apisto/abc123def456.../telemetry
→ Payload: {"temperature": 24.5}
→ Mosquitto broker receives it and delivers to all subscribers
```

### Step 2 — Go server receives the MQTT message

In `internal/mqtt/client.go`, `messageHandler` is called by paho:

```go
func (c *Client) messageHandler(_ mqtt.Client, msg mqtt.Message) {
    parts := strings.Split(msg.Topic(), "/")
    token := parts[1]              // "abc123def456..."
    msgType := parts[2]            // "telemetry"

    deviceID, err := c.resolveToken(c.ctx, token)
    // resolveToken checks the in-memory cache first, then hits the DB
    // if token is invalid → log and return (message silently dropped)

    case c.telemetryCh <- TelemetryMessage{DeviceID: deviceID, Payload: msg.Payload()}:
    // puts the message into the buffered channel — non-blocking
    // one of 4 worker goroutines will pick it up
}
```

### Step 3 — Worker goroutine picks it up

```go
func (c *Client) telemetryWorker() {
    for {
        select {
        case msg := <-c.telemetryCh:
            c.handlers.OnTelemetry(c.ctx, msg.DeviceID, msg.Payload)
            // OnTelemetry is set in server.go to call telemetrySvc.Ingest
        }
    }
}
```

### Step 4 — TelemetryService.Ingest

In `internal/services/telemetry_service.go`:

```go
func (s *TelemetryService) Ingest(ctx context.Context, deviceID string, payload []byte) error {
    var raw map[string]interface{}
    json.Unmarshal(payload, &raw)
    // raw = {"temperature": 24.5}

    for key, val := range raw {
        // key = "temperature", val = 24.5 (float64 in Go's JSON decoder)
        // detect type: float64 → "number"
        // INSERT INTO telemetry (time, device_id, key, value_numeric, value_type)
        // VALUES (now, deviceID, "temperature", 24.5, "number")
    }

    s.devices.UpsertDeviceKeys(ctx, deviceID, keyTypes)
    // makes sure "temperature" appears in device_keys so the dashboard knows about it

    s.devices.MarkOnline(ctx, deviceID, "")
    // UPDATE devices SET is_online = true, last_seen_at = NOW() WHERE id = $1

    s.realtime.Broadcast(deviceID, WSMessage{Type: "telemetry", Data: payload, ...})
    // pushes to all open browser WebSocket connections for this device
}
```

### Step 5 — RealtimeService fans out to browsers

In `internal/services/realtime_service.go`:

```go
func (s *RealtimeService) Broadcast(deviceID string, msg WSMessage) {
    s.mu.RLock()
    conns := ... // all WebSocket connections for this deviceID
    s.mu.RUnlock()

    data, _ := json.Marshal(msg)
    for _, conn := range conns {
        conn.WriteMessage(1, data)
        // WebSocket message type 1 = text frame
    }
}
```

### Step 6 — Browser receives the WebSocket message

In `web/src/hooks/useDeviceData.js`:

```js
useWebSocket(`/devices/${deviceId}/ws`, (msg) => {
    if (msg.type === 'telemetry') {
        setLatest((prev) => {
            // update the "temperature" entry with the new value
        })
    }
})
```

### Step 7 — React re-renders the widget

The `latest` state update triggers React to re-render `AutoDashboard`, which passes the new value to `ValueCard`, which displays `24.5°C` and adds it to the sparkline.

**Total time from ESP32 to browser:** typically under 100ms.

---

## 7. Line-by-line explanations of key patterns

### Pattern 1 — The service constructor

```go
// internal/services/telemetry_service.go

type TelemetryService struct {
    db       *database.DB       // pointer to database connection pool
    realtime *RealtimeService   // pointer to WebSocket hub
    devices  *DeviceService     // pointer to device service (to call UpsertDeviceKeys)
}

func NewTelemetryService(db *database.DB, realtime *RealtimeService, devices *DeviceService) *TelemetryService {
    return &TelemetryService{db: db, realtime: realtime, devices: devices}
}
```

This is *dependency injection*. The service doesn't create its dependencies — it receives them. This makes testing easier (you can pass a fake database) and makes the dependency graph explicit.

In `server.go` you can see all services being constructed in order:
```go
realtime := services.NewRealtimeService()           // no deps
deviceSvc := services.NewDeviceService(db, realtime)
telemetrySvc := services.NewTelemetryService(db, realtime, deviceSvc)
commandSvc := services.NewCommandService(db, realtime, mqttClient.Publish)
```

---

### Pattern 2 — A Fiber HTTP handler

```go
// internal/api/projects.go

func (h *projectHandlers) create(c *fiber.Ctx) error {
    // 1. Parse the request body
    var body struct {
        Name        string `json:"name"`
        Description string `json:"description"`
    }
    if err := c.BodyParser(&body); err != nil {
        return errResponse(c, 400, "invalid request body", "INVALID_INPUT")
    }

    // 2. Validate
    if body.Name == "" {
        return errResponse(c, 400, "name is required", "INVALID_INPUT")
    }

    // 3. Hit the database
    var p models.Project
    err := h.db.Pool.QueryRow(context.Background(), `
        INSERT INTO projects (name, description)
        VALUES ($1, $2)
        RETURNING id, name, description, created_at, updated_at`,
        body.Name, body.Description,
    ).Scan(&p.ID, &p.Name, &p.Description, &p.CreatedAt, &p.UpdatedAt)

    // 4. Handle error
    if err != nil {
        return errResponse(c, 500, "failed to create project", "INTERNAL_ERROR")
    }

    // 5. Return JSON
    return c.Status(201).JSON(p)
}
```

`RETURNING id, ...` is a PostgreSQL feature that returns the inserted row's columns. `.Scan(...)` copies those columns into the Go struct fields. The order of `.Scan()` arguments must match the order of columns in `RETURNING`.

---

### Pattern 3 — Querying multiple rows

```go
// Pattern used throughout the services

rows, err := s.db.Pool.Query(ctx, `SELECT id, name FROM devices WHERE project_id = $1`, projectID)
if err != nil {
    return nil, err
}
defer rows.Close()    // always close, even if we return early due to an error

var devices []*models.Device
for rows.Next() {     // rows.Next() advances the cursor and returns false when done
    var d models.Device
    if err := rows.Scan(&d.ID, &d.Name); err != nil {
        return nil, err
    }
    devices = append(devices, &d)  // append to the slice
}
return devices, nil
```

`append(devices, &d)` grows the slice. Starting as `nil` is fine — `append` creates the slice on first use.

---

### Pattern 4 — The MQTT worker pool

```go
// internal/mqtt/client.go

// On startup, create 4 worker goroutines
for i := 0; i < workerCount; i++ {
    c.wg.Add(1)      // tell the WaitGroup to expect 4 goroutines
    go c.telemetryWorker()
}

func (c *Client) telemetryWorker() {
    defer c.wg.Done()    // tell WaitGroup this goroutine finished
    for {
        select {
        case <-c.ctx.Done():
            // context cancelled = shutdown signal
            // drain remaining messages before exiting
            for {
                select {
                case msg := <-c.telemetryCh:
                    c.handlers.OnTelemetry(context.Background(), msg.DeviceID, msg.Payload)
                default:
                    return  // channel empty, exit
                }
            }
        case msg := <-c.telemetryCh:
            c.handlers.OnTelemetry(c.ctx, msg.DeviceID, msg.Payload)
        }
    }
}

// On shutdown:
func (c *Client) Disconnect() {
    c.cancel()    // cancels the context → workers see ctx.Done() and exit
    c.wg.Wait()   // blocks until all 4 workers have called wg.Done()
    c.client.Disconnect(500)
}
```

`sync.WaitGroup` is a counter. `wg.Add(4)`, four goroutines each call `wg.Done()`, and `wg.Wait()` blocks until the counter hits zero. This guarantees all in-flight telemetry is processed before shutdown.

---

### Pattern 5 — The WebSocket hub

```go
// internal/services/realtime_service.go

// The hub is a nested map: device ID → set of connections
// map[*websocket.Conn]bool is Go's idiomatic "set" — the bool is always true
type RealtimeService struct {
    mu   sync.RWMutex
    hubs map[string]map[*websocket.Conn]bool
}

// When a browser opens /devices/:id/ws
func (s *RealtimeService) Register(deviceID string, conn *websocket.Conn) {
    s.mu.Lock()
    defer s.mu.Unlock()
    if s.hubs[deviceID] == nil {
        s.hubs[deviceID] = make(map[*websocket.Conn]bool)
    }
    s.hubs[deviceID][conn] = true
}

// When the browser closes the tab
func (s *RealtimeService) Unregister(deviceID string, conn *websocket.Conn) {
    s.mu.Lock()
    defer s.mu.Unlock()
    delete(s.hubs[deviceID], conn)
    if len(s.hubs[deviceID]) == 0 {
        delete(s.hubs, deviceID)  // clean up empty map
    }
}
```

---

## 8. How to make a change

### Example A — Add a new field to an existing model

**Goal:** Add a `location` string field to devices.

**Step 1 — Add a migration**

Create `internal/database/migrations/003_device_location.up.sql`:
```sql
ALTER TABLE devices ADD COLUMN location TEXT DEFAULT '';
```

Create `internal/database/migrations/003_device_location.down.sql`:
```sql
ALTER TABLE devices DROP COLUMN location;
```

The migration runs automatically next time the server starts.

**Step 2 — Add the field to the model**

`internal/models/device.go`:
```go
type Device struct {
    // ... existing fields ...
    Location string `json:"location"`
}
```

**Step 3 — Update the queries that scan into Device**

Any `rows.Scan(...)` or `QueryRow(...).Scan(...)` that reads a device needs the new column added in the right position. Search for `Scan(&d.ID` in the codebase.

In `internal/services/device_service.go`, update the SELECT queries to include `location` and add `&d.Location` to the corresponding `.Scan(...)` call.

**Step 4 — Update the UPDATE handler if the field is editable**

In `internal/api/devices.go`, add `Location` to the body struct and include it in the UPDATE query.

---

### Example B — Add a new API endpoint

**Goal:** Add `GET /api/v1/devices/:id/stats` that returns the total number of telemetry rows for a device.

**Step 1 — Add the handler function**

In `internal/api/telemetry.go`:
```go
func (h *telemetryHandlers) stats(c *fiber.Ctx) error {
    id := c.Params("id")
    var count int64
    err := h.telemetrySvc.DB().Pool.QueryRow(context.Background(),
        `SELECT COUNT(*) FROM telemetry WHERE device_id = $1`, id,
    ).Scan(&count)
    if err != nil {
        return errResponse(c, 500, "query failed", "INTERNAL_ERROR")
    }
    return c.JSON(fiber.Map{"device_id": id, "total_rows": count})
}
```

**Step 2 — Register the route**

In `internal/api/router.go`, add one line:
```go
v1.Get("/devices/:id/stats", telemetry.stats)
```

That's it. The route is live next time the server starts.

---

### Example C — Add a new service method

**Goal:** Get the count of online devices in a project.

In `internal/services/device_service.go`:
```go
func (s *DeviceService) CountOnline(ctx context.Context, projectID string) (int, error) {
    var count int
    err := s.db.Pool.QueryRow(ctx,
        `SELECT COUNT(*) FROM devices WHERE project_id = $1 AND is_online = true`,
        projectID,
    ).Scan(&count)
    return count, err
}
```

Then call it from a handler:
```go
online, err := h.deviceSvc.CountOnline(ctx, projectID)
```

---

### What to do when you add a feature — checklist

- [ ] Does it need a DB schema change? → Add a migration file (`00N_name.up.sql` and `.down.sql`)
- [ ] Does it add new data? → Add/update the model in `internal/models/`
- [ ] Does it have business logic? → Add a method to the relevant service
- [ ] Does it need an HTTP endpoint? → Add handler in `internal/api/`, register in `router.go`
- [ ] Does the frontend need to call it? → Add a function to `web/src/lib/api.js`
- [ ] Does it need a new UI component? → Add to `web/src/components/` or `web/src/pages/`

---

## 9. Running the stack locally

### Full stack via Docker (recommended)

```bash
docker compose up -d              # start all 3 services
docker compose logs -f apisto    # watch server logs
docker compose down              # stop everything
docker compose down -v           # stop + wipe all data
```

### Rebuild after a Go code change

```bash
docker compose build apisto
docker compose up -d apisto
```

### Split dev (faster iteration)

```bash
# Terminal 1 — DB and MQTT only
docker compose up timescaledb mosquitto -d

# Terminal 2 — Go server (restart manually after changes)
go run ./cmd/apisto/

# Terminal 3 — Frontend with hot reload
cd web && npm run dev
# → http://localhost:5173 (proxies /api to :8080)
```

### Useful commands

```bash
# Test the API manually
curl -X POST http://localhost:8080/api/v1/projects \
  -H "Content-Type: application/json" \
  -d '{"name": "test"}'

# Simulate a device over MQTT
mosquitto_pub -h localhost -p 1883 \
  -t "apisto/TOKEN/telemetry" \
  -m '{"temperature": 24.5}'

# Connect to the database directly
docker exec -it apisto-db psql -U apisto -d apisto

# Inside psql — check telemetry data
SELECT * FROM telemetry ORDER BY time DESC LIMIT 10;
SELECT * FROM device_keys;
```

---

## 10. Where to read next

**Go language:**
- https://go.dev/tour/ — the official interactive tour, takes ~4 hours, covers everything in section 3 above
- https://go.dev/doc/effective_go — idiomatic Go patterns
- https://gobyexample.com/ — short examples for specific concepts (goroutines, channels, mutexes, etc.)

**Fiber:**
- https://docs.gofiber.io/ — especially "Context" (the `c *fiber.Ctx` object) and "Routing"

**pgx:**
- https://pkg.go.dev/github.com/jackc/pgx/v5 — focus on `Query`, `QueryRow`, `Exec`, `Scan`
- https://pkg.go.dev/github.com/jackc/pgx/v5/pgxpool — the pool specifically

**TimescaleDB:**
- https://docs.timescale.com/getting-started/latest/ — skim the "Hypertables" and "Continuous Aggregates" sections
- https://docs.timescale.com/api/latest/hyperfunctions/time_bucket/ — `time_bucket` function reference

**MQTT:**
- https://mqtt.org/mqtt-specification/ — the protocol (skim)
- https://www.emqx.com/en/blog/the-easiest-guide-to-getting-started-with-mqtt — practical introduction
- https://mqtt-explorer.com/ — GUI tool for exploring MQTT topics

**Docker Compose:**
- https://docs.docker.com/compose/gettingstarted/ — 10-minute intro

**React (if you want to touch the frontend):**
- https://react.dev/learn — official tutorial
- https://tailwindcss.com/docs/utility-first — Tailwind's mental model

---

### The fastest path to your first contribution

1. Run `docker compose up -d` and verify it works
2. Open http://localhost:8080, create a project and a device
3. Send fake telemetry with `mosquitto_pub` or curl and see it appear
4. Read `internal/server/server.go` — it's the spine of the whole application
5. Read `internal/mqtt/client.go` and `internal/services/telemetry_service.go` — the most interesting backend code
6. Read `web/src/hooks/useDeviceData.js` and `web/src/components/AutoDashboard.jsx` — the most interesting frontend code
7. Make a small change from Example A or B above and watch it work
