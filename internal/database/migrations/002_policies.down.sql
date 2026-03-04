DROP MATERIALIZED VIEW IF EXISTS telemetry_daily CASCADE;
DROP MATERIALIZED VIEW IF EXISTS telemetry_hourly CASCADE;
SELECT remove_compression_policy('telemetry', if_exists => true);
SELECT remove_retention_policy('telemetry', if_exists => true);
