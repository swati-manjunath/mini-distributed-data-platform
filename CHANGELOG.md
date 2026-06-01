# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased]
- Added `flink-jobs/` Flink SQL pipeline to consume `system-metrics` from Kafka, aggregate metrics, and send results to `mini-kv-store` via HTTP.
- Updated `kafka/docker-compose.yaml` listener configuration to support both Docker-internal broker access (`broker:9092`) and host-local access (`localhost:9092`).
- Updated `flink-jobs/config.py` to use `host.docker.internal:8082` for containerized Flink to reach the host `mini-kv-store` service.
- Fixed `flink-jobs/sinks/kv_store_sink.py` to send valid `key` / `value` JSON payloads to the key-value store API.
- Added Kafka broker orchestration under `kafka/docker-compose.yaml` for a local KRaft broker.
- Added `metrics-agent/` Python producer to publish host CPU/memory metrics to Kafka topic `system-metrics`.
- Added `consumer-debug/` Python Kafka consumer to inspect the `system-metrics` stream.
- Updated root `README.md` with documentation for the Flink job, Kafka networking, metrics agent, consumer-debug, and the key-value store.
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
