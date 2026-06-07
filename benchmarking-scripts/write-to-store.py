import socket
import time
import psutil
import requests
from concurrent.futures import ThreadPoolExecutor

# 1. Configuration
BASE_URL = "http://127.0.0.1:8080"
TOTAL_REQUESTS = 1000
MAX_WORKERS = 20

SYSTEM_NAME = socket.gethostname()

def get_base_metrics():
    """Gathers system metrics exactly once before flooding the database."""
    cpu = psutil.cpu_percent(interval=0.1)
    memory = psutil.virtual_memory().percent
    return cpu, memory

# Fetch the metrics payload upfront so psutil doesn't slow down the loop
INITIAL_CPU, INITIAL_MEMORY = get_base_metrics()

def send_benchmark_payload(request_id):
    """Sends the metrics payload structured specifically for your Go PutRequest struct."""
    
    # FIX: Matching what your Go 'PutRequest' validation expects.
    # Appending request_id ensures unique keys to test your hash ring routing!
    payload = {
        "key": f"{SYSTEM_NAME}-BENCH-{request_id}",
        "value": f"cpu:{INITIAL_CPU}|mem:{INITIAL_MEMORY}"
    }
    
    try:
        response = requests.post(f"{BASE_URL}/put", json=payload, timeout=2.0)
        return response.status_code
    except Exception:
        return "ERROR"

def main():
    print(f"🚀 Launching Benchmark: Sending {TOTAL_REQUESTS} payloads matching PutRequest struct to {BASE_URL}/put")
    print(f"🧵 Thread Pool Size: {MAX_WORKERS} (Zero sleep delay between requests)\n")
    
    start_time = time.perf_counter()
    
    # Use parallel threads to smash the Go endpoint without waiting
    with ThreadPoolExecutor(max_workers=MAX_WORKERS) as executor:
        results = list(executor.map(send_benchmark_payload, range(TOTAL_REQUESTS)))
        
    duration = time.perf_counter() - start_time
    
    # Calculate performance metrics
    successes = results.count(200)
    errors = results.count("ERROR")
    bad_requests = results.count(400)
    other_errors = TOTAL_REQUESTS - successes - errors - bad_requests
    
    print("📊 BENCHMARK METRICS:")
    print("=========================================")
    print(f"Total Test Duration : {duration:.3f} seconds")
    print(f"Throughput Rate     : {TOTAL_REQUESTS / duration:.2f} requests/sec")
    print(f"Successful (200 OK) : {successes} / {TOTAL_REQUESTS}")
    print(f"Bad Requests (400)  : {bad_requests} / {TOTAL_REQUESTS}")
    print(f"Network Failures    : {errors}")
    print(f"Other HTTP Errors   : {other_errors}")
    print("=========================================")

if __name__ == "__main__":
    main()
