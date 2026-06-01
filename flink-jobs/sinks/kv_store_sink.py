import requests
import json
from pyflink.table import DataTypes
from pyflink.table.udf import udf
from config import KV_STORE_API_URL

# 1. METRICS UDF (sends everything)
@udf(input_types=[DataTypes.STRING(), DataTypes.DOUBLE(), DataTypes.DOUBLE()], 
     result_type=DataTypes.STRING()) 
def send_metrics_to_microservice(host, avg_cpu, avg_memory):
    metric_payload = {"host": host, "cpu": avg_cpu, "memory": avg_memory, "type": "metric"}
    payload = {
        "key": host,
        "value": json.dumps(metric_payload)
    }
    try:
        # Timeout is crucial to prevent blocking the stream indefinitely
        response = requests.post(KV_STORE_API_URL, json=payload, timeout=2)
        return f"Metric Sent: {response.status_code}"
    except Exception as e:
        return f"Metric Error: {str(e)}"

# 2. ALERTS UDF (sends only high usage)
# FIX: Added the missing @udf decorator here
@udf(input_types=[DataTypes.STRING(), DataTypes.DOUBLE(), DataTypes.DOUBLE()], 
     result_type=DataTypes.STRING()) 
def send_alerts_to_microservice(host, avg_cpu, avg_memory):
    alert_payload = {"host": host, "cpu": avg_cpu, "memory": avg_memory, "type": "ALERT"}
    payload = {
        "key": host,
        "value": json.dumps(alert_payload)
    }
    try:
        response = requests.post(KV_STORE_API_URL, json=payload, timeout=2)
        return f"Alert Sent: {response.status_code}"
    except Exception as e:
        return f"Alert Error: {str(e)}"
