# Architecture

This repository defines a small distributed data platform with two main subsystems:

1. **Key-value store subsystem**
   - `mini-kv-store/` provides a Go HTTP API for storing and retrieving string values by key.
   - Writes are persisted to a node-specific write-ahead log (`data-<node-id>.log`).
   - In cluster mode, the service routes each key to its owning node and forwards requests accordingly.

2. **Metrics and Flink pipeline subsystem**
   - `kafka/` contains a Docker Compose configuration for a local Kafka KRaft broker.
   - `metrics-agent/` collects system CPU and memory metrics and publishes them to the Kafka topic `system-metrics`.
   - `flink-jobs/` consumes the `system-metrics` stream, performs aggregation, and forwards selected results to `mini-kv-store` via HTTP.
   - `consumer-debug/` reads from the `system-metrics` topic and prints each message for debugging.

## High-level flow

```mermaid
graph TB
  MetricsAgent["metrics-agent"] -->|publish system metrics| Kafka["Kafka broker"]
  Kafka -->|topic: system-metrics| Flink["flink-jobs"]
  Flink -->|HTTP writes| KVStore["mini-kv-store"]
  Kafka -->|topic: system-metrics| Consumer["consumer-debug"]
  Client["Client"] -->|HTTP /put,/get| KVStore["mini-kv-store"]
  KVStore -->|persist to WAL| WAL["data-<node-id>.log"]
```

## Component details

### `mini-kv-store`

- Built in Go.
- Exposes:
  - `POST /put` for writing key/value pairs.
  - `GET /get?key=<key>` for reading values.
- Uses an in-memory map protected by a mutex.
- Appends writes to a local WAL file on every successful write.
- Supports optional cluster mode where keys are routed to an owning node.

### `kafka`

- Local Kafka broker is provided via `docker-compose.yaml`.
- Uses Kafka KRaft mode, so there is no Zookeeper dependency.
- Exposes port `9092` for producer and consumer connections.

### `metrics-agent`

- Python app that gathers:
  - `cpu_percent`
  - `memory_percent`
  - host name
  - timestamp
- Sends metrics to Kafka topic `system-metrics`.
- Uses `psutil` to collect system statistics and `kafka-python` to publish.

### `flink-jobs`

- Flink SQL application that reads from Kafka topic `system-metrics`.
- Performs windowed aggregation and emits both metric and alert payloads.
- Sends aggregated results to `mini-kv-store` using HTTP POST with `key`/`value` JSON payloads.
- Uses `host.docker.internal` in Docker to reach the host `mini-kv-store` service when running inside containers.

### `consumer-debug`

- Python Kafka consumer that subscribes to the `system-metrics` topic.
- Decodes incoming JSON payloads and prints them.
- Useful for manual validation of the metrics pipeline.

## Deployment notes

- Start Kafka first with `docker compose up -d` in `kafka/`.
- Then run `metrics-agent/agent.py` to begin publishing data.
- Start `consumer-debug/consume-metrics.py` to confirm data delivery.
- Run `mini-kv-store` separately to exercise the key-value service.

## Notes

- The metrics pipeline is intentionally separate from the key-value store.
- The current architecture demonstrates both a clustered Go datastore and a Kafka-based telemetry flow in one repository.
