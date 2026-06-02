# Kafka Configs
KAFKA_BROKER = "broker:9092"
INPUT_TOPIC = "system-metrics"
CONSUMER_GROUP = "flink-metrics"

# Thresholds
CPU_ALERT_THRESHOLD = 90.0
MEMORY_ALERT_THRESHOLD = 90.0

# Window Configs
WINDOW_SIZE = "30"  # Seconds

# External Services
# Use host.docker.internal inside Docker containers on Windows/macOS to reach services running on the host.
KV_STORE_API_URL = "http://host.docker.internal:8082/put"