"""
End-to-end benchmark script.

Sends metric messages to Kafka topic `system-metrics`, then polls the mini-kv-store
`/latest?key=<host>` endpoint until the latest value changes. Prints one line
per test: the elapsed time in seconds (microsecond precision) or `ERROR` when
not observed within timeout.

Requirements:
  pip install kafka-python requests

Configuration variables are at the top of the file.
"""

WINDOW_SIZE = 1.0  # seconds: matches Flink window size
MAX_SEND_ATTEMPTS = 3
SEND_RETRY_DELAY = 0.5
import time
import json
import socket
import argparse
from uuid import uuid4
import requests

try:
    from kafka import KafkaProducer
except Exception as e:
    raise SystemExit("kafka-python is required: pip install kafka-python")

# Configuration (edit as needed)
KAFKA_BOOTSTRAP = "localhost:9092"
KAFKA_TOPIC = "system-metrics"
MINI_KV_BASE = "http://127.0.0.1:8080"
TOTAL_TESTS = 50
MESSAGES_PER_TEST = 5
POLL_TIMEOUT = 30.0  # seconds
POLL_INTERVAL = 0.5  # seconds

HOSTNAME = socket.gethostname()


def make_producer(bootstrap_servers):
    return KafkaProducer(
        bootstrap_servers=bootstrap_servers,
        value_serializer=lambda v: json.dumps(v).encode("utf-8"),
        linger_ms=5,
    )


DEBUG = False
KEY_NAME = None


def get_latest(host):
    # Retry a few times for transient connection issues
    for attempt in range(1, 4):
        try:
            resp = requests.get(f"{MINI_KV_BASE}/latest?key={host}", timeout=3.0)
            if resp.status_code != 200:
                if DEBUG:
                    print(f"DEBUG: /latest returned {resp.status_code}: {resp.text}")
                return None
            j = resp.json()
            return j.get("latest")
        except Exception as exc:
            if DEBUG:
                print(f"DEBUG: /latest attempt {attempt} error: {exc}")
            if attempt < 3:
                time.sleep(0.2)
            else:
                return None


def send_messages(producer, topic, host, marker, count):
    # Send `count` metric messages matching the live agent schema.
    for attempt in range(1, MAX_SEND_ATTEMPTS + 1):
        try:
            for i in range(count):
                payload = {
                    "host": host,
                    "ts": int(time.time()),
                    "cpu_percent": float(i),
                    "memory_percent": float(i),
                }
                producer.send(topic, payload)
            producer.flush()
            return True
        except Exception as exc:
            if DEBUG:
                print(f"DEBUG: Kafka send attempt {attempt} failed: {exc}")
            if attempt < MAX_SEND_ATTEMPTS:
                time.sleep(SEND_RETRY_DELAY)
            else:
                return False


def run_test(producer, test_idx, results):
    key = KEY_NAME or HOSTNAME

    # Read pre-existing latest value
    initial = get_latest(key)

    # Use a unique marker so you can correlate messages in logs if needed
    marker = f"e2e-{test_idx}-{uuid4().hex[:8]}"

    # Send a small burst to ensure Flink window has data
    if not send_messages(producer, KAFKA_TOPIC, HOSTNAME, marker, MESSAGES_PER_TEST):
        if DEBUG:
            print("DEBUG: failed to send messages after retries")
        print("ERROR")
        return

    # start measuring after messages have been flushed to Kafka
    start = time.perf_counter()
    deadline = start + POLL_TIMEOUT
    while time.perf_counter() < deadline:
        latest = get_latest(key)
        # Success if latest changes (or appears when previously missing)
        if (initial is None and latest is not None) or (initial is not None and latest != initial):
            elapsed = time.perf_counter() - start
            results.append(elapsed)
            running_avg = sum(results) / len(results)
            # Print elapsed and running average
            print(f"{elapsed:.6f} {running_avg:.6f}")
            return
        time.sleep(POLL_INTERVAL)

    if DEBUG:
        print(f"DEBUG: timed out waiting for /latest?key={key}")
    print("ERROR")


def main():
    global KAFKA_TOPIC, MINI_KV_BASE, TOTAL_TESTS, MESSAGES_PER_TEST, POLL_TIMEOUT, DEBUG, KEY_NAME

    parser = argparse.ArgumentParser(description="End-to-end benchmark: Kafka -> Flink -> mini-kv-store")
    parser.add_argument("--bootstrap", default=KAFKA_BOOTSTRAP)
    parser.add_argument("--topic", default=KAFKA_TOPIC)
    parser.add_argument("--kv", default=MINI_KV_BASE)
    parser.add_argument("--tests", type=int, default=TOTAL_TESTS)
    parser.add_argument("--per", type=int, default=MESSAGES_PER_TEST)
    parser.add_argument("--timeout", type=float, default=POLL_TIMEOUT)
    parser.add_argument("--key", default=None, help="Override the host key used for /latest polling")
    parser.add_argument("--debug", action="store_true", help="Print debug output for /latest and Kafka send failures")
    args = parser.parse_args()

    producer = make_producer(args.bootstrap)

    KAFKA_TOPIC = args.topic
    MINI_KV_BASE = args.kv
    TOTAL_TESTS = args.tests
    MESSAGES_PER_TEST = args.per
    POLL_TIMEOUT = args.timeout
    DEBUG = args.debug
    if args.key:
        KEY_NAME = args.key

    if DEBUG:
        print(f"DEBUG: bootstrap={args.bootstrap} topic={args.topic} kv={args.kv} key={KEY_NAME or HOSTNAME}")

    print(f"Running {TOTAL_TESTS} tests, {MESSAGES_PER_TEST} messages/test -> topic {args.topic}")

    results = []
    errors = 0
    for i in range(TOTAL_TESTS):
        before = len(results)
        run_test(producer, i, results)
        if len(results) == before:
            # no new result appended => error for this test
            errors += 1

    if len(results) > 0:
        overall_avg = sum(results) / len(results)
        print(f"COMPLETED: {len(results)} successes, {errors} errors, overall_avg={overall_avg:.6f}s")
    else:
        print(f"COMPLETED: 0 successes, {errors} errors")


if __name__ == '__main__':
    main()
