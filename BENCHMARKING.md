# Benchmarking

This project uses a small, personal-project benchmark suite to separate storage-layer performance from full pipeline behavior.

## Benchmark Scope

| Benchmark | What is tested? | Uses Flink? |
| --- | --- | --- |
| KV Store Write Latency | Storage layer `/put` writes, including WAL append and in-memory update | No |
| KV Store Read Latency | Storage layer `/get` point reads | No |
| History Query Latency | Query layer `/history` reads over indexed history | No |
| Recovery Time | WAL replay and server startup recovery | No |
| End-to-End Pipeline Latency | Kafka to Flink to KV store to `/latest` visibility | Yes |

## Latest Benchmark Results

Environment: local development machine, local Kafka/Flink/KV store setup.

| Benchmark | Requests / Tests | Concurrency | Success Rate | Throughput | Avg Latency | p50 | p95 | p99 | Max |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| KV Write Latency | 1,000 | 20 | 100% | 123.30 req/s | 158.48 ms | 147.10 ms | 253.04 ms | 504.43 ms | 522.88 ms |
| KV Write Latency | 5,000 | 20 | 100% | 117.47 req/s | 169.49 ms | 146.69 ms | 344.20 ms | 466.14 ms | 571.01 ms |
| KV Write Latency | 10,000 | 20 | 76.05% | 142.13 req/s | 174.70 ms | 162.82 ms | 291.93 ms | 382.49 ms | 528.19 ms |
| KV Read Latency | 1,000 | 50 | 100% | 692.55 req/s | 56.57 ms | 54.35 ms | 116.37 ms | 136.73 ms | 204.86 ms |
| E2E Pipeline Latency | 10 tests | N/A | 100% | N/A | 6093.27 ms | 5969.34 ms | 7052.05 ms | 7052.05 ms | 7052.05 ms |
| E2E Pipeline Latency | 100 tests | N/A | 99% | N/A | 6812.85 ms | 6128.28 ms | 8058.77 ms | 8129.74 ms | 8285.93 ms |

## Observations

The key-value store handled write workloads reliably up to 5,000 requests at concurrency 20, with a 100% success rate. Write throughput stayed around 117-123 requests per second, and median write latency remained close to 147 ms.

At 10,000 write requests, the system completed 7,605 successful writes and reported 2,395 request errors. This result is useful as a stress boundary: it suggests the local setup begins to hit sustained write pressure from request timeouts, local resource limits, WAL overhead, or connection pressure.

Read performance was significantly faster than write performance. For 1,000 spread reads at concurrency 50, the store reached 692.55 requests per second with a 100% success rate. Median read latency was 54.35 ms, with p95 at 116.37 ms.

The end-to-end benchmark measured this path:

```text
Kafka -> Flink window aggregation -> KV store write -> /latest visibility
```

The 10-test run completed successfully with an average latency of about 6.09 seconds. The 100-test run completed 99 out of 100 tests, with an average latency of about 6.81 seconds and p95 latency around 8.06 seconds.

End-to-end latency includes Flink windowing and watermark delay, so it should not be interpreted as only raw message-processing time.

## How To Run

Start the key-value store before running storage-layer benchmarks:

```powershell
go run ./mini-kv-store -port 8080 -node-id 1
```

Run write latency:

```powershell
python benchmarking-scripts/write-to-store.py --requests 1000 --workers 20
```

Run read latency:

```powershell
python benchmarking-scripts/read-benchmark.py --endpoint get --requests 1000 --workers 50
```

Run history query latency:

```powershell
python benchmarking-scripts/read-benchmark.py --endpoint history --keys 1000 --requests 100 --workers 10
```

Run recovery time:

```powershell
python benchmarking-scripts/recovery-benchmark.py --generate-wal 10000 --overwrite
```

For the end-to-end benchmark, start Kafka, the key-value store, and the Flink job first. Then run:

```powershell
python benchmarking-scripts/read-e2e-benchmark.py --kv http://127.0.0.1:8082 --tests 10 --per 5
```

For faster local E2E benchmark feedback, use a short Flink window, such as:

```python
WINDOW_SIZE = "1"
```

The E2E benchmark still includes the configured Flink watermark delay.
