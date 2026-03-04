package models

import "time"

type Command struct {
	ID        string     `json:"id"`
	DeviceID  string     `json:"device_id"`
	Command   string     `json:"command"`
	Payload   string     `json:"payload"`
	Status    string     `json:"status"`
	CreatedAt time.Time  `json:"created_at"`
	SentAt    *time.Time `json:"sent_at"`
	AckedAt   *time.Time `json:"acked_at"`
}
