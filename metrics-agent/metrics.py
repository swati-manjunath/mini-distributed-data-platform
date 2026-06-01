import psutil
import json
import time
import socket


# Get the local computer name
SYSTEM_NAME = socket.gethostname()

 # 1. Collect raw data
def get_cpu_usage():
    return psutil.cpu_percent()

def get_memory_usage():
    return psutil.virtual_memory().percent

# 2. Build the dictionary
def collect_metrics():  
    metrics = {
                "host": SYSTEM_NAME,
                "ts": int(time.time()),
                "cpu_percent": get_cpu_usage(),
                "memory_percent": get_memory_usage()
            }
    
    return metrics