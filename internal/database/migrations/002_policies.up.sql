-- Automatic data retention: drop telemetry chunks older than 30 days
SELECT add_retention_policy('telemetry', INTERVAL '30 days');

-- Enable compression on chunks older than 7 days
ALTER TABLE telemetry SET (
    timescaledb.compress,
    timescaledb.compress_segmentby = 'device_id, key',
    timescaledb.compress_orderby = 'time DESC'
);

SELECT add_compression_policy('telemetry', INTERVAL '7 days');

-- Continuous aggregate: pre-compute hourly averages
CREATE MATERIALIZED VIEW telemetry_hourly
WITH (timescaledb.continuous) AS
SELECT
    time_bucket('1 hour', time) AS bucket,
    device_id,
    key,
    AVG(value_numeric) AS avg_value,
    MIN(value_numeric) AS min_value,
    MAX(value_numeric) AS max_value,
    COUNT(*) AS sample_count,
    last(value_numeric, time) AS last_value
FROM telemetry
WHERE value_type = 'number'
GROUP BY bucket, device_id, key
WITH NO DATA;

SELECT add_continuous_aggregate_policy('telemetry_hourly',
    start_offset => INTERVAL '3 hours',
    end_offset => INTERVAL '1 hour',
    schedule_interval => INTERVAL '1 hour'
);

-- Daily aggregate for long-range dashboards
CREATE MATERIALIZED VIEW telemetry_daily
WITH (timescaledb.continuous) AS
SELECT
    time_bucket('1 day', time) AS bucket,
    device_id,
    key,
    AVG(value_numeric) AS avg_value,
    MIN(value_numeric) AS min_value,
    MAX(value_numeric) AS max_value,
    COUNT(*) AS sample_count,
    last(value_numeric, time) AS last_value
FROM telemetry
WHERE value_type = 'number'
GROUP BY bucket, device_id, key
WITH NO DATA;

SELECT add_continuous_aggregate_policy('telemetry_daily',
    start_offset => INTERVAL '3 days',
    end_offset => INTERVAL '1 day',
    schedule_interval => INTERVAL '1 day'
);
