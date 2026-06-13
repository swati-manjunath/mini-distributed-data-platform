import argparse
import socket
import time
from collections import Counter
from concurrent.futures import ThreadPoolExecutor

import requests

from benchmark_utils import print_summary


BASE_URL = "http://127.0.0.1:8080"
TOTAL_REQUESTS = 1000
MAX_WORKERS = 20
TIMEOUT_SECONDS = 3.0
SYSTEM_NAME = socket.gethostname()


def send_put(base_url, key_prefix, request_id, timeout):
    payload = {
        "key": f"{key_prefix}-{request_id}",
        "value": f"benchmark-value-{request_id}",
    }

    start = time.perf_counter()
    try:
        response = requests.post(f"{base_url}/put", json=payload, timeout=timeout)
        elapsed_ms = (time.perf_counter() - start) * 1000
        return response.status_code, elapsed_ms
    except requests.RequestException:
        return "ERROR", None


def main():
    parser = argparse.ArgumentParser(description="Benchmark KV Store Write Latency via /put")
    parser.add_argument("--base-url", default=BASE_URL)
    parser.add_argument("--requests", type=int, default=TOTAL_REQUESTS)
    parser.add_argument("--workers", type=int, default=MAX_WORKERS)
    parser.add_argument("--timeout", type=float, default=TIMEOUT_SECONDS)
    parser.add_argument("--key-prefix", default=f"{SYSTEM_NAME}-WRITE-BENCH")
    args = parser.parse_args()

    started_at = time.perf_counter()
    latencies_ms = []
    status_counts = Counter()
    errors = 0

    with ThreadPoolExecutor(max_workers=args.workers) as executor:
        futures = [
            executor.submit(send_put, args.base_url, args.key_prefix, i, args.timeout)
            for i in range(args.requests)
        ]

        for future in futures:
            status, latency_ms = future.result()
            status_counts[status] += 1
            if latency_ms is None:
                errors += 1
            else:
                latencies_ms.append(latency_ms)

    print_summary(
        benchmark="KV Store Write Latency",
        target=f"{args.base_url}/put",
        total_requests=args.requests,
        concurrency=args.workers,
        started_at=started_at,
        latencies_ms=latencies_ms,
        status_counts=status_counts,
        errors=errors,
        extra={"uses_flink": False, "key_prefix": args.key_prefix},
    )


if __name__ == "__main__":
    main()
