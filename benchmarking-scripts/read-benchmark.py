import time
import requests
from concurrent.futures import ThreadPoolExecutor
import socket

# Configuration
BASE_URL = "http://127.0.0.1:8080"
TOTAL_REQUESTS = 1000
MAX_WORKERS = 50
SYSTEM_NAME = socket.gethostname()
KEY_TEMPLATE = f"{SYSTEM_NAME}-BENCH-0"  # default key; change as needed


def send_get(request_id):
    key = KEY_TEMPLATE
    try:
        start = time.perf_counter()
        resp = requests.get(f"{BASE_URL}/history?key={key}", timeout=5.0)
        elapsed = time.perf_counter() - start
        elapsed_ms = elapsed * 1000
        # Print only the elapsed time in seconds with microsecond precision
        print(f"{elapsed_ms:.3f}")
    except Exception:
        # Print a sentinel value for errors so you can filter them later
        print("ERROR")


def main():
    with ThreadPoolExecutor(max_workers=MAX_WORKERS) as ex:
        list(ex.map(send_get, range(TOTAL_REQUESTS)))


if __name__ == '__main__':
    main()
