# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased]
- Added `CHANGELOG.md` for project history tracking.
- Upgraded README with detailed mini key-value store description and usage.
- Fixed startup ordering so cluster config is parsed before WAL file initialization, ensuring `data-<node-id>.log` uses the correct local node ID.

## [2026-05-20]
- Add clustering support (basic node config and key routing).
- Load `NUMBER_OF_NODES` from `.env` and print on startup.
- Fix forwarding logic: use node index → node mapping and avoid double-writes on the origin node.
- Fix WAL handling: read request body once, append to `data.log`, and decode JSON from the preserved bytes; skip malformed WAL lines on load.

## [2026-05-15]
- Initialized mini KV store server with JSON `POST /put` and `GET /get` API.
- Added write-ahead logging to `data.log`.
- Implemented thread-safe in-memory store with mutex.
- Added `handlers.go`, `store.go`, `wal.go`, and `README.md` documentation.
