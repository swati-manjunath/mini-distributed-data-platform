import json
import time


def percentile(values, percentile_value):
    if not values:
        return None

    sorted_values = sorted(values)
    index = round((len(sorted_values) - 1) * percentile_value / 100)
    return sorted_values[index]


def summarize_latencies_ms(latencies_ms):
    if not latencies_ms:
        return {
            "avg": None,
            "p50": None,
            "p95": None,
            "p99": None,
            "max": None,
        }

    return {
        "avg": sum(latencies_ms) / len(latencies_ms),
        "p50": percentile(latencies_ms, 50),
        "p95": percentile(latencies_ms, 95),
        "p99": percentile(latencies_ms, 99),
        "max": max(latencies_ms),
    }


def print_summary(
    benchmark,
    target,
    total_requests,
    concurrency,
    started_at,
    latencies_ms,
    status_counts,
    errors,
    extra=None,
):
    duration = time.perf_counter() - started_at
    successes = sum(
        count for status, count in status_counts.items()
        if isinstance(status, int) and 200 <= status < 300
    )

    summary = {
        "benchmark": benchmark,
        "target": target,
        "requests": total_requests,
        "concurrency": concurrency,
        "duration_seconds": duration,
        "throughput_per_second": total_requests / duration if duration > 0 else None,
        "successes": successes,
        "errors": errors,
        "status_counts": {str(k): v for k, v in sorted(status_counts.items(), key=lambda item: str(item[0]))},
        "latency_ms": summarize_latencies_ms(latencies_ms),
    }

    if extra:
        summary.update(extra)

    print(json.dumps(summary, indent=2))
