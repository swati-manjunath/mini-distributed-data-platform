"""
End-to-end benchmark for:

Kafka -> Flink window aggregation -> mini-kv-store /latest

The benchmark sends metric events for the local host, then polls
`/latest?key=<host>` until Flink writes a newer window result.

Important: the Flink source uses a 5 second watermark delay. For quick 1 second
windows, this script keeps sending small heartbeat events while polling so the
watermark advances and the window can close.

Requirements:
  pip install kafka-python requests
"""

import argparse
import json
import socket
import time

import requests

from benchmark_utils import summarize_latencies_ms

try:
    from kafka import KafkaProducer
except Exception:
    raise SystemExit("kafka-python is required: pip install kafka-python")


KAFKA_BOOTSTRAP = "localhost:9092"
KAFKA_TOPIC = "system-metrics"
MINI_KV_BASE = "http://127.0.0.1:8082"

TOTAL_TESTS = 10
MESSAGES_PER_TEST = 5
POLL_TIMEOUT = 20.0
POLL_INTERVAL = 0.25
HEARTBEAT_INTERVAL = 0.5

HOSTNAME = socket.gethostname()
DEBUG = False


def make_producer(bootstrap_servers):
    return KafkaProducer(
        bootstrap_servers=bootstrap_servers,
        value_serializer=lambda v: json.dumps(v).encode("utf-8"),
        linger_ms=5,
    )


def make_metric(host, sample_id):
    # Vary values slightly so each aggregate is easy to distinguish in /latest.
    value = float(sample_id % 100)
    return {
        "host": host,
        "ts": int(time.time()),
        "cpu_percent": value,
        "memory_percent": value,
    }


def send_metrics(producer, topic, host, count, sample_start):
    for i in range(count):
        producer.send(topic, make_metric(host, sample_start + i))
    producer.flush()


def parse_latest_value(raw_value):
    if not raw_value:
        return None

    try:
        return json.loads(raw_value)
    except json.JSONDecodeError:
        return None


def get_latest_window(base_url, host):
    try:
        resp = requests.get(f"{base_url}/latest", params={"key": host}, timeout=3.0)
    except requests.RequestException as exc:
        if DEBUG:
            print(f"DEBUG: /latest request failed: {exc}")
        return None

    if resp.status_code != 200:
        if DEBUG:
            print(f"DEBUG: /latest returned {resp.status_code}: {resp.text}")
        return None

    payload = resp.json()
    latest = parse_latest_value(payload.get("latest"))
    if not latest:
        if DEBUG:
            print(f"DEBUG: latest value was not JSON: {payload.get('latest')!r}")
        return None

    try:
        return int(latest["window_start_ms"])
    except (KeyError, TypeError, ValueError):
        return None


def run_test(producer, args, test_idx, results):
    initial_window = get_latest_window(args.kv, args.key)
    sample_id = test_idx * 1000

    send_metrics(producer, args.topic, args.key, args.per, sample_id)

    start = time.perf_counter()
    start_epoch = int(time.time())
    deadline = start + args.timeout
    next_heartbeat = start + HEARTBEAT_INTERVAL
    heartbeat_id = sample_id + args.per

    while time.perf_counter() < deadline:
        now = time.perf_counter()

        if now >= next_heartbeat:
            send_metrics(producer, args.topic, args.key, 1, heartbeat_id)
            heartbeat_id += 1
            next_heartbeat = now + HEARTBEAT_INTERVAL

        latest_window = get_latest_window(args.kv, args.key)
        if (
            latest_window is not None
            and latest_window != initial_window
            and latest_window >= start_epoch
        ):
            elapsed = time.perf_counter() - start
            results.append(elapsed)
            running_avg = sum(results) / len(results)
            print(f"{elapsed:.6f} {running_avg:.6f}")
            return True

        time.sleep(POLL_INTERVAL)

    if DEBUG:
        print(
            "DEBUG: timed out waiting for a newer window "
            f"for key={args.key}; initial_window={initial_window}"
        )
    print("TIMEOUT")
    return False


def main():
    global DEBUG

    parser = argparse.ArgumentParser(
        description="Benchmark Kafka -> Flink -> mini-kv-store latency"
    )
    parser.add_argument("--bootstrap", default=KAFKA_BOOTSTRAP)
    parser.add_argument("--topic", default=KAFKA_TOPIC)
    parser.add_argument("--kv", default=MINI_KV_BASE)
    parser.add_argument("--tests", type=int, default=TOTAL_TESTS)
    parser.add_argument("--per", type=int, default=MESSAGES_PER_TEST)
    parser.add_argument("--timeout", type=float, default=POLL_TIMEOUT)
    parser.add_argument("--key", default=HOSTNAME, help="Host key used for /latest polling")
    parser.add_argument("--debug", action="store_true")
    args = parser.parse_args()

    DEBUG = args.debug

    if DEBUG:
        print(
            "DEBUG: "
            f"bootstrap={args.bootstrap} topic={args.topic} kv={args.kv} key={args.key}"
        )

    producer = make_producer(args.bootstrap)
    results = []
    timeouts = 0

    print(f"Running {args.tests} tests, {args.per} messages/test -> topic {args.topic}")
    try:
        for i in range(args.tests):
            if not run_test(producer, args, i, results):
                timeouts += 1
    finally:
        producer.close()

    latencies_ms = [value * 1000 for value in results]
    summary = {
        "benchmark": "End-to-End Pipeline Latency",
        "uses_flink": True,
        "bootstrap": args.bootstrap,
        "topic": args.topic,
        "kv": args.kv,
        "tests": args.tests,
        "messages_per_test": args.per,
        "successes": len(results),
        "timeouts": timeouts,
        "latency_ms": summarize_latencies_ms(latencies_ms),
    }
    print(json.dumps(summary, indent=2))


if __name__ == "__main__":
    main()
