# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased]
- Added Kafka broker orchestration under `kafka/docker-compose.yaml` for a local KRaft broker.
- Added `metrics-agent/` Python producer to publish host CPU/memory metrics to Kafka topic `system-metrics`.
- Added `consumer-debug/` Python Kafka consumer to inspect the `system-metrics` stream.
- Updated root `README.md` with documentation for Kafka, metrics-agent, consumer-debug, and the key-value store.
- Added `ARCHITECTURE.md` with an end-to-end architecture overview.
- Included Python requirements for metrics and Kafka integration.

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
