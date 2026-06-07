# Architecture Decisions

## Project Goal

This project demonstrates a simplified distributed data processing and storage platform. Rather than focusing on a single technology, it models the complete lifecycle of monitoring data:

```
Agent
    ↓
Kafka
    ↓
Flink
    ↓
Distributed Key-Value Store
    ↓
Query APIs
```

The objective is to show how distributed systems separate data collection, transport, computation, and storage while maintaining scalability and fault tolerance.

---

# Why use a monitoring agent?

A lightweight Python agent periodically collects CPU and memory utilization using `psutil`.

The agent publishes metrics every 5 seconds instead of writing directly to the database.

### Why?

* The agent should only be responsible for collecting data.
* It should not know how data is processed.
* It should not depend on the database implementation.
* Multiple downstream consumers should be able to process the same events.

This separation allows independent scaling of producers and consumers.

---

# Why use Kafka?

Kafka acts as a durable event log between the producer and the processing layer.

## Benefits

### Decoupling

The producer and consumer operate independently.

If Flink is temporarily unavailable, the producer can continue writing events.

---

### Durability

Events are stored for a configurable retention period.

This allows:

* replaying historical events,
* recomputing metrics if processing logic changes,
* recovering from downstream failures.

---

### Scalability

Kafka scales through partitions.

Increasing the number of partitions allows multiple consumers to process data in parallel.

A single partition limits parallelism because only one consumer in a consumer group can own that partition.

---

### Ordering

Kafka guarantees ordering within a partition.

By choosing an appropriate partition key, events for the same host can be processed in order.

---

# Why use Flink?

Raw metrics have limited value.

The useful information comes from continuous computation over streams of events.

Flink provides efficient stream processing without requiring custom logic for:

* offset management,
* windowing,
* continuous computation,
* stream state management.

---

# Why use tumbling windows?

The project computes aggregates every 30 seconds.

A tumbling window ensures:

* each event contributes exactly once,
* no repeated computation,
* fixed periodic summaries.

Short CPU spikes are less important than sustained trends, making a 30-second aggregation appropriate for this monitoring use case.

---

# Why not use sliding windows?

Sliding windows recompute overlapping windows continuously.

They are appropriate for dashboards that require near real-time updates.

However, they require maintaining overlapping state and repeatedly recalculating metrics, increasing computational cost.

Since this project only requires one summary every 30 seconds, tumbling windows are a better fit.

---

# Why create separate Metrics and Alerts tables?

The system separates normal aggregated metrics from alert events.

Metrics represent periodic summaries.

Alerts only contain events that exceed predefined thresholds.

Separating these workloads simplifies querying and allows different retention or processing strategies.

---

# Why use a distributed key-value store?

The project demonstrates horizontal scalability.

Instead of storing every key on every node, keys are partitioned across nodes.

Each node owns a subset of the key space.

This prevents every request from becoming a cluster-wide broadcast.

---

# Why hash keys?

Given a key:

```
hash(key)
        ↓
assigned node
```

Every node can independently determine which node owns the key.

If a request reaches the wrong node, it forwards the request to the correct node instead of querying every node.

This keeps lookup cost effectively constant rather than growing with cluster size.

---

# Why use consistent hashing?

Simple approaches such as:

```
hash(key) % N
```

cause large amounts of data movement whenever nodes are added or removed.

Consistent hashing minimizes redistribution by moving only a small portion of keys when cluster membership changes.

This makes horizontal scaling significantly more efficient.

---

# Why use replication?

Replication improves fault tolerance.

If one node becomes unavailable, another copy of the data exists.

The project uses asynchronous replication to prioritize lower write latency.

The trade-off is temporary inconsistency until replication completes.

---

# Why asynchronous replication instead of synchronous replication?

Synchronous replication would require waiting for replicas before responding to the client.

This increases write latency.

Asynchronous replication allows the client request to complete quickly while replicas are updated in the background.

The trade-off is eventual consistency rather than immediate consistency.

---

# Why use a Write-Ahead Log (WAL)?

Memory is volatile.

If the process crashes, in-memory state is lost.

Before updating the in-memory store, operations are written to the WAL.

After restart, the WAL is replayed to reconstruct state.

Without a WAL, committed writes could disappear after a crash.

---

# Why maintain indexes for History and Latest?

History queries require retrieving all values associated with a key.

Maintaining an index avoids scanning the complete database.

Latest queries require retrieving the newest value.

Instead of comparing timestamps across all records, an index allows direct access to the most recent value.

This significantly reduces query cost.

---

# Recovery strategy

If asynchronous replication fails:

1. The primary still records the operation in the WAL.
2. After restart, the WAL reconstructs the primary state.
3. A scheduled reconciliation job periodically retries failed replication requests.
4. The cluster eventually becomes consistent.

This design favors lower write latency while providing eventual consistency.

---

# Scaling strategy

If the number of monitored hosts increases significantly:

## Kafka

* Increase partitions.
* Increase consumer parallelism.

## Flink

* Increase job parallelism so multiple partitions are processed concurrently.

## Database

* Add additional nodes.
* Redistribute keys using consistent hashing.
* Increase storage and write throughput horizontally.

---

# Trade-offs

This project intentionally demonstrates engineering trade-offs rather than attempting to implement every production feature.

Examples include:

* lower latency through asynchronous replication,
* durability through WAL,
* scalable event ingestion through Kafka,
* efficient stream computation through Flink,
* horizontal scaling through partitioning and consistent hashing,
* eventual consistency instead of synchronous consensus.

The goal is to model how modern distributed data platforms separate ingestion, processing, and storage into independently scalable components.
