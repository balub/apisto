package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/balub/apisto/internal/database"
	"github.com/balub/apisto/internal/models"
)

type TelemetryService struct {
	db       *database.DB
	realtime *RealtimeService
	devices  *DeviceService
}

func NewTelemetryService(db *database.DB, realtime *RealtimeService, devices *DeviceService) *TelemetryService {
	return &TelemetryService{db: db, realtime: realtime, devices: devices}
}

// Ingest parses a raw JSON payload, detects types, stores in DB, and fans out to WebSocket.
func (s *TelemetryService) Ingest(ctx context.Context, deviceID string, payload []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(payload, &raw); err != nil {
		return fmt.Errorf("parse payload: %w", err)
	}

	now := time.Now()
	keyTypes := make(map[string]string)

	for key, val := range raw {
		var numVal *float64
		var textVal *string
		var boolVal *bool
		var valType string

		switch v := val.(type) {
		case float64:
			valType = "number"
			numVal = &v
		case bool:
			valType = "boolean"
			boolVal = &v
		case string:
			valType = "string"
			textVal = &v
		default:
			// JSON object/array: store as string
			valType = "json"
			b, _ := json.Marshal(v)
			s := string(b)
			textVal = &s
		}

		keyTypes[key] = valType

		_, err := s.db.Pool.Exec(ctx, `
			INSERT INTO telemetry (time, device_id, key, value_numeric, value_text, value_bool, value_type)
			VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			now, deviceID, key, numVal, textVal, boolVal, valType)
		if err != nil {
			return fmt.Errorf("insert telemetry key %s: %w", key, err)
		}
	}

	// Upsert device keys for auto-dashboard
	if err := s.devices.UpsertDeviceKeys(ctx, deviceID, keyTypes); err != nil {
		return fmt.Errorf("upsert device keys: %w", err)
	}

	// Update device online status
	if err := s.devices.MarkOnline(ctx, deviceID, ""); err != nil {
		return fmt.Errorf("mark online: %w", err)
	}

	// Fan out to WebSocket subscribers
	ts := now
	s.realtime.Broadcast(deviceID, WSMessage{
		Type:      "telemetry",
		DeviceID:  deviceID,
		Data:      payload,
		Timestamp: &ts,
	})

	return nil
}

type TelemetryQueryParams struct {
	DeviceID  string
	Key       string
	From      time.Time
	To        time.Time
	Limit     int
	Order     string
	Aggregate string
	Interval  string
}

func (s *TelemetryService) Query(ctx context.Context, p TelemetryQueryParams) (interface{}, error) {
	if p.Limit <= 0 {
		p.Limit = 1000
	}
	if p.Order != "asc" && p.Order != "desc" {
		p.Order = "asc"
	}
	if p.From.IsZero() {
		p.From = time.Now().Add(-time.Hour)
	}
	if p.To.IsZero() {
		p.To = time.Now()
	}

	duration := p.To.Sub(p.From)

	// Use continuous aggregates for long time ranges
	if p.Aggregate != "" && duration > 7*24*time.Hour {
		return s.queryFromDaily(ctx, p)
	}
	if p.Aggregate != "" && duration > 24*time.Hour {
		return s.queryFromHourly(ctx, p)
	}
	if p.Aggregate != "" {
		return s.queryAggregated(ctx, p)
	}
	return s.queryRaw(ctx, p)
}

func (s *TelemetryService) queryRaw(ctx context.Context, p TelemetryQueryParams) ([]*models.TelemetryPoint, error) {
	query := `
		SELECT time, device_id, key, value_numeric, value_text, value_bool, value_type
		FROM telemetry
		WHERE device_id = $1 AND time BETWEEN $2 AND $3`
	args := []interface{}{p.DeviceID, p.From, p.To}
	argIdx := 4

	if p.Key != "" {
		query += fmt.Sprintf(" AND key = $%d", argIdx)
		args = append(args, p.Key)
		argIdx++
	}

	query += fmt.Sprintf(" ORDER BY time %s LIMIT $%d", p.Order, argIdx)
	args = append(args, p.Limit)

	rows, err := s.db.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var points []*models.TelemetryPoint
	for rows.Next() {
		var pt models.TelemetryPoint
		if err := rows.Scan(&pt.Time, &pt.DeviceID, &pt.Key,
			&pt.ValueNumeric, &pt.ValueText, &pt.ValueBool, &pt.ValueType); err != nil {
			return nil, err
		}
		points = append(points, &pt)
	}
	return points, nil
}

func intervalToSQL(interval string) string {
	switch interval {
	case "1m":
		return "1 minute"
	case "5m":
		return "5 minutes"
	case "15m":
		return "15 minutes"
	case "1h":
		return "1 hour"
	case "1d":
		return "1 day"
	default:
		return "1 hour"
	}
}

func (s *TelemetryService) queryAggregated(ctx context.Context, p TelemetryQueryParams) ([]*models.TelemetryQueryResult, error) {
	intervalSQL := intervalToSQL(p.Interval)
	aggFunc := "AVG"
	switch p.Aggregate {
	case "min":
		aggFunc = "MIN"
	case "max":
		aggFunc = "MAX"
	case "count":
		aggFunc = "COUNT"
	}

	query := fmt.Sprintf(`
		SELECT
			time_bucket('%s', time) AS bucket,
			%s(value_numeric) AS value,
			MIN(value_numeric) AS min_val,
			MAX(value_numeric) AS max_val,
			COUNT(*) AS cnt
		FROM telemetry
		WHERE device_id = $1 AND key = $2 AND time BETWEEN $3 AND $4 AND value_type = 'number'
		GROUP BY bucket
		ORDER BY bucket %s
		LIMIT $5`, intervalSQL, aggFunc, p.Order)

	rows, err := s.db.Pool.Query(ctx, query, p.DeviceID, p.Key, p.From, p.To, p.Limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanAggResults(rows)
}

func (s *TelemetryService) queryFromHourly(ctx context.Context, p TelemetryQueryParams) ([]*models.TelemetryQueryResult, error) {
	rows, err := s.db.Pool.Query(ctx, `
		SELECT bucket, avg_value, min_value, max_value, sample_count
		FROM telemetry_hourly
		WHERE device_id = $1 AND key = $2 AND bucket BETWEEN $3 AND $4
		ORDER BY bucket ASC LIMIT $5`,
		p.DeviceID, p.Key, p.From, p.To, p.Limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanAggResults(rows)
}

func (s *TelemetryService) queryFromDaily(ctx context.Context, p TelemetryQueryParams) ([]*models.TelemetryQueryResult, error) {
	rows, err := s.db.Pool.Query(ctx, `
		SELECT bucket, avg_value, min_value, max_value, sample_count
		FROM telemetry_daily
		WHERE device_id = $1 AND key = $2 AND bucket BETWEEN $3 AND $4
		ORDER BY bucket ASC LIMIT $5`,
		p.DeviceID, p.Key, p.From, p.To, p.Limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanAggResults(rows)
}

func scanAggResults(rows interface{ Next() bool; Scan(...interface{}) error; Close() }) ([]*models.TelemetryQueryResult, error) {
	defer rows.Close()
	var results []*models.TelemetryQueryResult
	for rows.Next() {
		var r models.TelemetryQueryResult
		if err := rows.Scan(&r.Time, &r.Value, &r.Min, &r.Max, &r.Count); err != nil {
			return nil, err
		}
		results = append(results, &r)
	}
	return results, nil
}

// Latest returns the most recent value for each key for a device.
func (s *TelemetryService) Latest(ctx context.Context, deviceID string) ([]*models.LatestTelemetry, error) {
	rows, err := s.db.Pool.Query(ctx, `
		SELECT DISTINCT ON (key)
			key, value_numeric, value_text, value_bool, value_type, time
		FROM telemetry
		WHERE device_id = $1
		ORDER BY key, time DESC`, deviceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*models.LatestTelemetry
	for rows.Next() {
		var key, valueType string
		var numVal *float64
		var textVal *string
		var boolVal *bool
		var t time.Time

		if err := rows.Scan(&key, &numVal, &textVal, &boolVal, &valueType, &t); err != nil {
			return nil, err
		}

		lt := &models.LatestTelemetry{Key: key, ValueType: valueType, Time: t}
		switch valueType {
		case "number":
			if numVal != nil {
				lt.Value = *numVal
			}
		case "boolean":
			if boolVal != nil {
				lt.Value = *boolVal
			}
		default:
			if textVal != nil {
				lt.Value = *textVal
			}
		}
		results = append(results, lt)
	}
	return results, nil
}
