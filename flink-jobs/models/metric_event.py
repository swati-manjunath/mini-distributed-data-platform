def get_source_ddl(topic, broker, group_id):
    return f"""
    CREATE TABLE metrics (
        host STRING,
        ts BIGINT,
        cpu_percent DOUBLE,
        memory_percent DOUBLE,
        event_time AS TO_TIMESTAMP_LTZ(ts * 1000, 3),
        WATERMARK FOR event_time AS event_time - INTERVAL '5' SECOND
    ) WITH (
        'connector' = 'kafka',
        'topic' = '{topic}',
        'properties.bootstrap.servers' = '{broker}',
        'properties.group.id' = '{group_id}',
        'scan.startup.mode' = 'latest-offset',
        'format' = 'json',
        'json.ignore-parse-errors' = 'false'
    )
    """
