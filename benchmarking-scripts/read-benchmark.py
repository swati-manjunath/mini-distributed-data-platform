import argparse
import socket
import time
from collections import Counter
from concurrent.futures import ThreadPoolExecutor

import requests

from benchmark_utils import print_summary


BASE_URL = "http://127.0.0.1:8080"
TOTAL_REQUESTS = 1000
MAX_WORKERS = 50
TIMEOUT_SECONDS = 5.0
SYSTEM_NAME = socket.gethostname()


def preload_keys(base_url, key_prefix, count, timeout):
    for i in range(count):
        payload = {
            "key": f"{key_prefix}-{i}",
            "value": f"benchmark-value-{i}",
        }
        response = requests.post(f"{base_url}/put", json=payload, timeout=timeout)
        response.raise_for_status()


def key_for_request(key_prefix, request_id, key_count, mode):
    if mode == "single":
        return f"{key_prefix}-0"
    return f"{key_prefix}-{request_id % key_count}"


def send_get(base_url, endpoint, key, timeout):
    start = time.perf_counter()
    try:
        response = requests.get(
            f"{base_url}/{endpoint}",
            params={"key": key},
            timeout=timeout,
        )
        elapsed_ms = (time.perf_counter() - start) * 1000
        return response.status_code, elapsed_ms
    except requests.RequestException:
        return "ERROR", None


def benchmark_name(endpoint):
    if endpoint == "history":
        return "History Query Latency"
    if endpoint == "latest":
        return "Latest Query Latency"
    return "KV Store Read Latency"


def main():
    parser = argparse.ArgumentParser(
        description="Benchmark KV Store Read, Latest, or History query latency"
    )
    parser.add_argument("--base-url", default=BASE_URL)
    parser.add_argument("--endpoint", choices=["get", "history", "latest"], default="get")
    parser.add_argument("--requests", type=int, default=TOTAL_REQUESTS)
    parser.add_argument("--workers", type=int, default=MAX_WORKERS)
    parser.add_argument("--timeout", type=float, default=TIMEOUT_SECONDS)
    parser.add_argument("--key-prefix", default=f"{SYSTEM_NAME}-READ-BENCH")
    parser.add_argument("--keys", type=int, default=1000, help="Number of keys to preload/read")
    parser.add_argument(
        "--mode",
        choices=["single", "spread"],
        default="spread",
        help="Read one key repeatedly or spread reads across preloaded keys",
    )
    parser.add_argument(
        "--skip-preload",
        action="store_true",
        help="Use existing keys instead of writing test data before the read benchmark",
    )
    args = parser.parse_args()

    if args.keys < 1:
        raise SystemExit("--keys must be at least 1")

    if not args.skip_preload:
        preload_keys(args.base_url, args.key_prefix, args.keys, args.timeout)

    started_at = time.perf_counter()
    latencies_ms = []
    status_counts = Counter()
    errors = 0

    with ThreadPoolExecutor(max_workers=args.workers) as executor:
        futures = []
        for i in range(args.requests):
            key = key_for_request(args.key_prefix, i, args.keys, args.mode)
            futures.append(
                executor.submit(send_get, args.base_url, args.endpoint, key, args.timeout)
            )

        for future in futures:
            status, latency_ms = future.result()
            status_counts[status] += 1
            if latency_ms is None:
                errors += 1
            else:
                latencies_ms.append(latency_ms)

    print_summary(
        benchmark=benchmark_name(args.endpoint),
        target=f"{args.base_url}/{args.endpoint}",
        total_requests=args.requests,
        concurrency=args.workers,
        started_at=started_at,
        latencies_ms=latencies_ms,
        status_counts=status_counts,
        errors=errors,
        extra={
            "uses_flink": False,
            "endpoint": args.endpoint,
            "key_prefix": args.key_prefix,
            "preloaded_keys": args.keys,
            "read_mode": args.mode,
            "preload_skipped": args.skip_preload,
        },
    )


if __name__ == "__main__":
    main()
