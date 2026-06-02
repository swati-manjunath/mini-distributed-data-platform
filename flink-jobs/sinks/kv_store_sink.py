import requests
import json
from pyflink.table import DataTypes
from pyflink.table.udf import udf
from config import KV_STORE_API_URL

# 1. METRICS UDF (sends everything)
@udf(input_types=[DataTypes.STRING(), DataTypes.DOUBLE(), DataTypes.DOUBLE(), DataTypes.BIGINT()], 
     result_type=DataTypes.STRING()) 
def send_metrics_to_microservice(host, avg_cpu, avg_memory, window_start_ms):
    metric_payload = {
        "host": host, 
        "cpu": avg_cpu, 
        "memory": avg_memory, 
        "type": "metric",
        "window_start_ms": window_start_ms
    }
    payload = {
        "key": f"{host}_{window_start_ms}",
        "value": json.dumps(metric_payload)
    }
    try:
        # Timeout is crucial to prevent blocking the stream indefinitely
        response = requests.post(KV_STORE_API_URL, json=payload, timeout=2)
        return f"Metric Sent: {response.status_code}"
    except Exception as e:
        return f"Metric Error: {str(e)}"

# 2. ALERTS UDF (sends only high usage)
@udf(input_types=[DataTypes.STRING(), DataTypes.DOUBLE(), DataTypes.DOUBLE(), DataTypes.BIGINT()], 
     result_type=DataTypes.STRING()) 
def send_alerts_to_microservice(host, avg_cpu, avg_memory, window_start_ms):
    alert_payload = {
        "host": host, 
        "cpu": avg_cpu, 
        "memory": avg_memory, 
        "type": "ALERT",
        "window_start_ms": window_start_ms
    }
    payload = {
        "key": f"{host}_{window_start_ms}",
        "value": json.dumps(alert_payload)
    }
    try:
        response = requests.post(KV_STORE_API_URL, json=payload, timeout=2)
        return f"Alert Sent: {response.status_code}"
    except Exception as e:
        return f"Alert Error: {str(e)}"
