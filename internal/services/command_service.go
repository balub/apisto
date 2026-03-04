package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/balub/apisto/internal/database"
	"github.com/balub/apisto/internal/models"
)

type CommandService struct {
	db       *database.DB
	realtime *RealtimeService
	publish  func(topic string, payload []byte) error
}

func NewCommandService(db *database.DB, realtime *RealtimeService, publish func(topic string, payload []byte) error) *CommandService {
	return &CommandService{db: db, realtime: realtime, publish: publish}
}

type commandMQTTPayload struct {
	ID      string `json:"id"`
	Command string `json:"command"`
	Payload string `json:"payload"`
}

func (s *CommandService) Send(ctx context.Context, deviceID, command, payload string) (*models.Command, error) {
	// Get device token for MQTT topic
	var token string
	err := s.db.Pool.QueryRow(ctx, `SELECT token FROM devices WHERE id = $1`, deviceID).Scan(&token)
	if err != nil {
		return nil, fmt.Errorf("get device token: %w", err)
	}

	// Insert command record
	var cmd models.Command
	err = s.db.Pool.QueryRow(ctx, `
		INSERT INTO commands (device_id, command, payload, status)
		VALUES ($1, $2, $3, 'pending')
		RETURNING id, device_id, command, payload, status, created_at, sent_at, acked_at`,
		deviceID, command, payload,
	).Scan(&cmd.ID, &cmd.DeviceID, &cmd.Command, &cmd.Payload, &cmd.Status,
		&cmd.CreatedAt, &cmd.SentAt, &cmd.AckedAt)
	if err != nil {
		return nil, fmt.Errorf("insert command: %w", err)
	}

	// Publish to MQTT
	mqttPayload, _ := json.Marshal(commandMQTTPayload{
		ID:      cmd.ID,
		Command: command,
		Payload: payload,
	})

	topic := fmt.Sprintf("apisto/%s/commands", token)
	if err := s.publish(topic, mqttPayload); err != nil {
		log.Printf("command: mqtt publish error: %v", err)
		s.db.Pool.Exec(ctx, `UPDATE commands SET status = 'failed' WHERE id = $1`, cmd.ID)
		return nil, fmt.Errorf("publish command: %w", err)
	}

	// Mark as sent
	now := time.Now()
	s.db.Pool.Exec(ctx, `UPDATE commands SET status = 'sent', sent_at = $2 WHERE id = $1`, cmd.ID, now)
	cmd.Status = "sent"
	cmd.SentAt = &now

	return &cmd, nil
}

func (s *CommandService) HandleAck(ctx context.Context, commandID string) {
	now := time.Now()
	_, err := s.db.Pool.Exec(ctx, `
		UPDATE commands SET status = 'acknowledged', acked_at = $2 WHERE id = $1`,
		commandID, now)
	if err != nil {
		log.Printf("command ack: update error: %v", err)
		return
	}

	// Notify WebSocket clients
	var deviceID string
	s.db.Pool.QueryRow(ctx, `SELECT device_id FROM commands WHERE id = $1`, commandID).Scan(&deviceID)
	if deviceID != "" {
		s.realtime.Broadcast(deviceID, WSMessage{
			Type:      "command_ack",
			DeviceID:  deviceID,
			CommandID: commandID,
			Status:    "acknowledged",
		})
	}
}

func (s *CommandService) List(ctx context.Context, deviceID string, limit int) ([]*models.Command, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := s.db.Pool.Query(ctx, `
		SELECT id, device_id, command, payload, status, created_at, sent_at, acked_at
		FROM commands WHERE device_id = $1 ORDER BY created_at DESC LIMIT $2`,
		deviceID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cmds []*models.Command
	for rows.Next() {
		var cmd models.Command
		if err := rows.Scan(&cmd.ID, &cmd.DeviceID, &cmd.Command, &cmd.Payload, &cmd.Status,
			&cmd.CreatedAt, &cmd.SentAt, &cmd.AckedAt); err != nil {
			return nil, err
		}
		cmds = append(cmds, &cmd)
	}
	return cmds, nil
}
