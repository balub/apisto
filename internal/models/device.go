package models

import "time"

type Device struct {
	ID              string     `json:"id"`
	ProjectID       string     `json:"project_id"`
	Token           string     `json:"token,omitempty"`
	Name            string     `json:"name"`
	Description     string     `json:"description"`
	Metadata        []byte     `json:"metadata,omitempty"`
	IsOnline        bool       `json:"is_online"`
	LastSeenAt      *time.Time `json:"last_seen_at"`
	FirstSeenAt     *time.Time `json:"first_seen_at"`
	FirmwareVersion string     `json:"firmware_version"`
	IPAddress       string     `json:"ip_address"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type DeviceKey struct {
	DeviceID    string    `json:"device_id"`
	Key         string    `json:"key"`
	ValueType   string    `json:"value_type"`
	WidgetType  string    `json:"widget_type"`
	DisplayName string    `json:"display_name"`
	Unit        string    `json:"unit"`
	SortOrder   int       `json:"sort_order"`
	FirstSeenAt time.Time `json:"first_seen_at"`
	LastSeenAt  time.Time `json:"last_seen_at"`
}
