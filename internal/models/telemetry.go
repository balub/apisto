package models

import "time"

type TelemetryPoint struct {
	Time         time.Time  `json:"time"`
	DeviceID     string     `json:"device_id"`
	Key          string     `json:"key"`
	ValueNumeric *float64   `json:"value_numeric,omitempty"`
	ValueText    *string    `json:"value_text,omitempty"`
	ValueBool    *bool      `json:"value_bool,omitempty"`
	ValueType    string     `json:"value_type"`
}

type TelemetryQueryResult struct {
	Time  time.Time `json:"time"`
	Value float64   `json:"value"`
	Min   *float64  `json:"min,omitempty"`
	Max   *float64  `json:"max,omitempty"`
	Count *int64    `json:"count,omitempty"`
}

type LatestTelemetry struct {
	Key       string      `json:"key"`
	Value     interface{} `json:"value"`
	ValueType string      `json:"value_type"`
	Time      time.Time   `json:"time"`
}

type TelemetryIngest struct {
	DeviceID  string
	Timestamp time.Time
	Values    map[string]interface{}
}
