from metrics import collect_metrics
from producer import create_producer, send_metric, close_producer
import time

try: 
    producer = create_producer()
    while True:
        json_data = collect_metrics()
        send_metric(producer, json_data)
        time.sleep(5)

except KeyboardInterrupt:
    close_producer(producer)
    print("\nPublishing stopped by user.")
