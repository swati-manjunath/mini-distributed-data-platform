from kafka import KafkaConsumer
import json

try:
    consumer = KafkaConsumer(
        'system-metrics',
        bootstrap_servers='localhost:9092',
        auto_offset_reset='earliest',
        enable_auto_commit=True,
        group_id='metrics-consumers',
        value_deserializer=lambda x: json.loads(x.decode('utf-8')))

    print("Waiting for messages...")
    for message in consumer:
        print(f"Value: {message.value}")

except KeyboardInterrupt:
    consumer.close()
    print("\nConsumer stopped by user.")