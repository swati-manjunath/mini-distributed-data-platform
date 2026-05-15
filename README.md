# mini-distributed-data-platform

A simple Go-based in-memory key-value store with JSON POST ingestion and HTTP query support.

This project exposes a lightweight HTTP API for storing and retrieving string values by key. Incoming write requests are also appended to `data.log` as a write-ahead log.

## Features

- `POST /put` accepts a JSON payload and stores `key` → `value`
- `GET /get?key=<key>` returns the stored value for a given key
- In-memory map storage with thread-safe access via a mutex
- Write-ahead logging to `data.log` for simple persistence and debugging

## Build and Run

```bash
cd c:/Users/vidya/OneDrive/Documents/Swati/mini-distributed-data-platform
go run .
```

Or build a binary:

```bash
go build -o mini-kv-store .
./mini-kv-store
```

The server listens on port `8080`.

## API

### Store a value

`POST /put`

Request body:

```json
{
  "key": "cpu",
  "value": "80"
}
```

Example using PowerShell:

```powershell
Invoke-WebRequest -Uri "http://localhost:8080/put" -Method Post -ContentType "application/json" -Body '{"key":"cpu","value":"80"}'
```

### Retrieve a value

`GET /get?key=<key>`

Example:

```powershell
Invoke-WebRequest -Uri "http://localhost:8080/get?key=cpu" -Method Get
```

## Notes

- The server currently stores values in memory, so data is lost when the process exits.
- `data.log` is used as a write-ahead log for received POST requests.
- If `data.log` contains malformed lines, the loader skips them and continues.

## File layout

- `main.go` — server setup and route registration
- `handlers.go` — request handlers for `/put` and `/get`
- `store.go` — shared in-memory store and mutex
- `wal.go` — write-ahead logging helper

## Requirements

- Go 1.26 or later
