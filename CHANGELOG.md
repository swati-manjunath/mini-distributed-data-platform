# Changelog

All notable changes to this project are documented here.

## [2026-06-13]

- Cleaned up repository metadata for GitHub publication.
- Stopped requiring a `.env` file for the Go server to start.
- Fixed forwarded read requests so response bodies are relayed to the caller.
- Fixed analytics replica fallback to forward to `/history` and `/latest` instead of `/get`.
- Added `.env.example` for local configuration.

## [2026-06-07]

- Documented `GET /history` and `GET /latest` analytics endpoints in `README.md` and `ARCHITECTURE.md`.
- Added `ARCHITECTURAL_DECISIONS.md` to capture architectural decisions and design rationale for the distribution and analytics pipeline.
- Clarified the `mini-kv-store/analytics-handlers.go` implementation of analytics read endpoints returning structured JSON responses.

## [2026-06-03]

- Changed `mini-kv-store` routing logic: target node is now computed by hashing the host only to ensure consistent ownership regardless of per-window timestamps.
- Added `GET /history` and `GET /latest` endpoints to `mini-kv-store` for analytics reads.
- Added `ARCHITECTURE.md` with an end-to-end architecture overview.
- Included Python requirements for metrics and Kafka integration.

## [2026-06-01]

- Updated `flink-jobs/sinks/kv_store_sink.py` to include `window_start_ms` in the payload and use `key = "{host}_{window_start_ms}"` for per-window uniqueness.
- Updated `flink-jobs/main.py` to pass the window start timestamp into the UDFs.
- Updated `kafka/docker-compose.yaml` listener configuration to support both Docker-internal broker access (`broker:9092`) and host-local access (`localhost:9092`).
- Updated `flink-jobs/config.py` to use `host.docker.internal:8082` for containerized Flink to reach the host `mini-kv-store` service.
- Fixed `flink-jobs/sinks/kv_store_sink.py` to send valid `key` and `value` JSON payloads to the key-value store API.

## [2026-05-26]

- Added `metrics-agent/` Python producer to publish host CPU/memory metrics to Kafka topic `system-metrics`.
- Added `consumer-debug/` Python Kafka consumer to inspect the `system-metrics` stream.
- Added Kafka broker orchestration under `kafka/docker-compose.yaml` for a local KRaft broker.

## [2026-05-24]

- Added `flink-jobs/` Flink SQL pipeline to consume `system-metrics` from Kafka, aggregate metrics, and send results to `mini-kv-store` via HTTP.
- Converted Flink windowing to use a tumbling window for non-overlapping aggregates.
- Flink aggregates now include `window_start`; the pipeline uses `UNIX_TIMESTAMP(CAST(window_start AS STRING))` when sending to the sink.

## [2026-05-20]

- Added clustering support with basic node config and key routing.
- Loaded `NUMBER_OF_NODES` from `.env` and printed it on startup.
- Fixed forwarding logic to use node mapping and avoid double writes on the origin node.
- Fixed WAL handling to read request bodies once, append to `data.log`, decode JSON from preserved bytes, and skip malformed WAL lines on load.

## [2026-05-15]

- Initialized mini KV store server with JSON `POST /put` and `GET /get` API.
- Added write-ahead logging to `data.log`.
- Implemented thread-safe in-memory store with mutex.
- Added `handlers.go`, `store.go`, `wal.go`, and README documentation.
