from pyflink.table import EnvironmentSettings, TableEnvironment
from config import INPUT_TOPIC, KAFKA_BROKER, CONSUMER_GROUP, CPU_ALERT_THRESHOLD, MEMORY_ALERT_THRESHOLD, WINDOW_SLIDE, WINDOW_SIZE
from models import metric_event
from processor import aggregator as cpu_aggregator
from sinks import kv_store_sink

def main():
    # 1. Setup
    env_settings = EnvironmentSettings.in_streaming_mode()
    t_env = TableEnvironment.create(env_settings)

    # 2. Register Source (Kafka)
    t_env.execute_sql(metric_event.get_source_ddl(
        INPUT_TOPIC, KAFKA_BROKER, CONSUMER_GROUP
    ))

    # 3. Register a "Dummy" Sink (Required to pull data through the UDFs)
    # The UDF does the HTTP POST; this sink just records the "Success/Fail" status text.
    t_env.execute_sql("""
    CREATE TABLE execution_log (
        status_msg STRING
    ) WITH (
        'connector' = 'print'
    )
    """)

    # 4. Register the UDFs
    t_env.create_temporary_system_function("send_metrics", kv_store_sink.send_metrics_to_microservice)
    t_env.create_temporary_system_function("send_alert", kv_store_sink.send_alerts_to_microservice)

    # 5. Create the View (Aggregation Logic)
    cpu_aggregator.register_window_aggregates(t_env, WINDOW_SLIDE, WINDOW_SIZE)

    # 6. Build the Execution Plan
    statement_set = t_env.create_statement_set()

    # --- PIPELINE 1: Send ALL metrics ---
    statement_set.add_insert_sql(f"""
    INSERT INTO execution_log
    SELECT 
        send_metrics(host, avg_cpu, avg_memory)
    FROM metric_aggregates
    """)

    # --- PIPELINE 2: Send ALERTS only (Filtered) ---
    # We apply the WHERE clause inside Flink SQL *before* calling the UDF.
    # This prevents the microservice from being spammed with non-alert data.
    statement_set.add_insert_sql(f"""
    INSERT INTO execution_log
    SELECT 
        send_alert(host, avg_cpu, avg_memory)
    FROM metric_aggregates
    WHERE avg_cpu > {CPU_ALERT_THRESHOLD} 
       OR avg_memory > {MEMORY_ALERT_THRESHOLD}
    """)

    # 7. Execute both pipelines
    statement_set.execute()

if __name__ == '__main__':
    main()
