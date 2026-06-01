from pyflink.table import TableEnvironment

def register_window_aggregates(t_env: TableEnvironment, slide_sec, size_sec):
    """
    Registers the core sliding window view.
    """
    query = f"""
    CREATE VIEW metric_aggregates AS
    SELECT
        host,
        AVG(cpu_percent) AS avg_cpu,
        AVG(memory_percent) AS avg_memory,
        window_start,
        window_end
    FROM TABLE(
        HOP(
            TABLE metrics,
            DESCRIPTOR(event_time),
            INTERVAL '{slide_sec}' SECOND,
            INTERVAL '{size_sec}' SECOND
        )
    )
    GROUP BY
        host,
        window_start,
        window_end
    """
    t_env.execute_sql(query)
