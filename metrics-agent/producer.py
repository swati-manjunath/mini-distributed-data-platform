import kafka
import time
import json
from metrics import collect_metrics
from metrics import SYSTEM_NAME

TOPIC_NAME = 'system-metrics'

def create_producer():
    producer = kafka.KafkaProducer(
        bootstrap_servers='localhost:9092',
        value_serializer=lambda x: json.dumps(x).encode('utf-8')
    )
    return producer

def send_metric(producer, metric):
    producer.send(TOPIC_NAME, metric)

def close_producer(producer):
    producer.flush()
    producer.close()


