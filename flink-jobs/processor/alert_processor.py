from config import CPU_ALERT_THRESHOLD, MEMORY_ALERT_THRESHOLD

def get_alert_query(source_view, target_table):
    return f"""
    INSERT INTO {target_table}
    SELECT *
    FROM {source_view}
    WHERE avg_cpu > {CPU_ALERT_THRESHOLD}
       OR avg_memory > {MEMORY_ALERT_THRESHOLD}
    """
