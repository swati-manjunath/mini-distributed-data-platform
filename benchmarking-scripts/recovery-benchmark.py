import argparse
import json
import os
import subprocess
import time

import requests


BASE_URL = "http://127.0.0.1:8080"
SERVER_COMMAND = "go run ./mini-kv-store -port 8080 -node-id 1"
WAL_PATH = "data-1.log"
KEY_PREFIX = "RECOVERY-BENCH"
TIMEOUT_SECONDS = 30.0


def generate_wal(path, entries, key_prefix, overwrite):
    if os.path.exists(path) and not overwrite:
        raise SystemExit(
            f"{path} already exists. Use --overwrite to replace it, "
            "or omit --generate-wal to benchmark the existing WAL."
        )

    with open(path, "w", encoding="utf-8") as file:
        for i in range(entries):
            payload = {
                "key": f"{key_prefix}-{i}",
                "value": f"benchmark-value-{i}",
            }
            file.write(json.dumps(payload) + "\n")


def wait_until_recovered(base_url, key, timeout):
    started_at = time.perf_counter()
    deadline = started_at + timeout
    last_error = None

    while time.perf_counter() < deadline:
        try:
            response = requests.get(
                f"{base_url}/get",
                params={"key": key},
                timeout=1.0,
            )
            if response.status_code == 200:
                return time.perf_counter() - started_at
            last_error = f"HTTP {response.status_code}"
        except requests.RequestException as exc:
            last_error = str(exc)

        time.sleep(0.1)

    raise TimeoutError(f"server did not recover within {timeout}s; last_error={last_error}")


def main():
    parser = argparse.ArgumentParser(
        description="Benchmark KV Store Recovery Time by starting the server and polling /get"
    )
    parser.add_argument("--base-url", default=BASE_URL)
    parser.add_argument("--server-command", default=SERVER_COMMAND)
    parser.add_argument("--cwd", default=".", help="Directory where the server command runs")
    parser.add_argument("--wal", default=WAL_PATH)
    parser.add_argument("--generate-wal", type=int, default=0, help="Generate this many WAL entries first")
    parser.add_argument("--overwrite", action="store_true", help="Allow replacing an existing WAL")
    parser.add_argument("--key-prefix", default=KEY_PREFIX)
    parser.add_argument("--key", default=None, help="Key to poll after startup")
    parser.add_argument("--timeout", type=float, default=TIMEOUT_SECONDS)
    parser.add_argument("--keep-running", action="store_true")
    args = parser.parse_args()

    if args.generate_wal > 0:
        generate_wal(args.wal, args.generate_wal, args.key_prefix, args.overwrite)

    recovery_key = args.key or f"{args.key_prefix}-0"
    wal_size_bytes = os.path.getsize(args.wal) if os.path.exists(args.wal) else 0

    process = subprocess.Popen(
        args.server_command,
        cwd=args.cwd,
        shell=True,
        stdout=subprocess.DEVNULL,
        stderr=subprocess.DEVNULL,
    )

    try:
        recovery_seconds = wait_until_recovered(args.base_url, recovery_key, args.timeout)
        summary = {
            "benchmark": "Recovery Time",
            "uses_flink": False,
            "server_command": args.server_command,
            "wal": args.wal,
            "wal_entries_generated": args.generate_wal,
            "wal_size_bytes": wal_size_bytes,
            "recovery_key": recovery_key,
            "recovery_seconds": recovery_seconds,
        }
        print(json.dumps(summary, indent=2))
    finally:
        if not args.keep_running:
            process.terminate()
            try:
                process.wait(timeout=5)
            except subprocess.TimeoutExpired:
                process.kill()


if __name__ == "__main__":
    main()
