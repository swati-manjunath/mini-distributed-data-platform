def get_sink_ddl(table_name):
    # Using 'print' connector for demonstration as per original code
    # Replace with 'jdbc' or 'hbase' for real KV Store
    return f"""
    CREATE TABLE {table_name} (
        host STRING,
        avg_cpu DOUBLE,
        avg_memory DOUBLE,
        window_start TIMESTAMP(3),
        window_end TIMESTAMP(3)
    ) WITH (
        'connector' = 'print'
    )
    """
