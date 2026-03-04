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

type DeviceService struct {
	db       *database.DB
	realtime *RealtimeService
}

func NewDeviceService(db *database.DB, realtime *RealtimeService) *DeviceService {
	return &DeviceService{db: db, realtime: realtime}
}

func (s *DeviceService) Create(ctx context.Context, projectID, name, description string) (*models.Device, error) {
	var d models.Device
	err := s.db.Pool.QueryRow(ctx, `
		INSERT INTO devices (project_id, name, description)
		VALUES ($1, $2, $3)
		RETURNING id, project_id, token, name, description, is_online, last_seen_at, first_seen_at, firmware_version, ip_address, created_at, updated_at`,
		projectID, name, description,
	).Scan(&d.ID, &d.ProjectID, &d.Token, &d.Name, &d.Description,
		&d.IsOnline, &d.LastSeenAt, &d.FirstSeenAt, &d.FirmwareVersion, &d.IPAddress,
		&d.CreatedAt, &d.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("create device: %w", err)
	}
	return &d, nil
}

func (s *DeviceService) GetByID(ctx context.Context, id string) (*models.Device, error) {
	var d models.Device
	err := s.db.Pool.QueryRow(ctx, `
		SELECT id, project_id, token, name, description, metadata, is_online, last_seen_at, first_seen_at, firmware_version, ip_address, created_at, updated_at
		FROM devices WHERE id = $1`, id,
	).Scan(&d.ID, &d.ProjectID, &d.Token, &d.Name, &d.Description, &d.Metadata,
		&d.IsOnline, &d.LastSeenAt, &d.FirstSeenAt, &d.FirmwareVersion, &d.IPAddress,
		&d.CreatedAt, &d.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get device: %w", err)
	}
	return &d, nil
}

func (s *DeviceService) GetByToken(ctx context.Context, token string) (*models.Device, error) {
	var d models.Device
	err := s.db.Pool.QueryRow(ctx, `
		SELECT id, project_id, token, name, description, is_online, last_seen_at, first_seen_at, firmware_version, ip_address, created_at, updated_at
		FROM devices WHERE token = $1`, token,
	).Scan(&d.ID, &d.ProjectID, &d.Token, &d.Name, &d.Description,
		&d.IsOnline, &d.LastSeenAt, &d.FirstSeenAt, &d.FirmwareVersion, &d.IPAddress,
		&d.CreatedAt, &d.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get device by token: %w", err)
	}
	return &d, nil
}

func (s *DeviceService) ListByProject(ctx context.Context, projectID string) ([]*models.Device, error) {
	rows, err := s.db.Pool.Query(ctx, `
		SELECT id, project_id, name, description, is_online, last_seen_at, first_seen_at, firmware_version, ip_address, created_at, updated_at
		FROM devices WHERE project_id = $1 ORDER BY created_at ASC`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var devices []*models.Device
	for rows.Next() {
		var d models.Device
		if err := rows.Scan(&d.ID, &d.ProjectID, &d.Name, &d.Description,
			&d.IsOnline, &d.LastSeenAt, &d.FirstSeenAt, &d.FirmwareVersion, &d.IPAddress,
			&d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, err
		}
		devices = append(devices, &d)
	}
	return devices, nil
}

func (s *DeviceService) Update(ctx context.Context, id, name, description, firmwareVersion string) (*models.Device, error) {
	var d models.Device
	err := s.db.Pool.QueryRow(ctx, `
		UPDATE devices SET name = $2, description = $3, firmware_version = $4, updated_at = NOW()
		WHERE id = $1
		RETURNING id, project_id, name, description, is_online, last_seen_at, first_seen_at, firmware_version, ip_address, created_at, updated_at`,
		id, name, description, firmwareVersion,
	).Scan(&d.ID, &d.ProjectID, &d.Name, &d.Description,
		&d.IsOnline, &d.LastSeenAt, &d.FirstSeenAt, &d.FirmwareVersion, &d.IPAddress,
		&d.CreatedAt, &d.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("update device: %w", err)
	}
	return &d, nil
}

func (s *DeviceService) Delete(ctx context.Context, id string) error {
	_, err := s.db.Pool.Exec(ctx, `DELETE FROM devices WHERE id = $1`, id)
	return err
}

func (s *DeviceService) GetKeys(ctx context.Context, deviceID string) ([]*models.DeviceKey, error) {
	rows, err := s.db.Pool.Query(ctx, `
		SELECT device_id, key, value_type, widget_type, display_name, unit, sort_order, first_seen_at, last_seen_at
		FROM device_keys WHERE device_id = $1 ORDER BY sort_order ASC, key ASC`, deviceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []*models.DeviceKey
	for rows.Next() {
		var k models.DeviceKey
		if err := rows.Scan(&k.DeviceID, &k.Key, &k.ValueType, &k.WidgetType,
			&k.DisplayName, &k.Unit, &k.SortOrder, &k.FirstSeenAt, &k.LastSeenAt); err != nil {
			return nil, err
		}
		keys = append(keys, &k)
	}
	return keys, nil
}

func (s *DeviceService) MarkOnline(ctx context.Context, deviceID string, ipAddress string) error {
	now := time.Now()
	_, err := s.db.Pool.Exec(ctx, `
		UPDATE devices SET is_online = true, last_seen_at = $2, ip_address = COALESCE(NULLIF($3, ''), ip_address),
		first_seen_at = COALESCE(first_seen_at, $2), updated_at = NOW()
		WHERE id = $1`, deviceID, now, ipAddress)
	if err != nil {
		return err
	}

	online := true
	s.realtime.Broadcast(deviceID, WSMessage{
		Type:     "status",
		DeviceID: deviceID,
		IsOnline: &online,
		LastSeen: &now,
	})
	return nil
}

func (s *DeviceService) MarkOffline(ctx context.Context, deviceID string) error {
	_, err := s.db.Pool.Exec(ctx, `
		UPDATE devices SET is_online = false, updated_at = NOW() WHERE id = $1`, deviceID)
	if err != nil {
		return err
	}

	offline := false
	s.realtime.Broadcast(deviceID, WSMessage{
		Type:     "status",
		DeviceID: deviceID,
		IsOnline: &offline,
	})
	return nil
}

// StartHeartbeatChecker marks devices offline if last_seen_at is older than timeout seconds.
func (s *DeviceService) StartHeartbeatChecker(ctx context.Context, timeoutSeconds int) {
	ticker := time.NewTicker(30 * time.Second)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				s.checkHeartbeats(ctx, timeoutSeconds)
			}
		}
	}()
}

func (s *DeviceService) checkHeartbeats(ctx context.Context, timeoutSeconds int) {
	rows, err := s.db.Pool.Query(ctx, `
		SELECT id FROM devices
		WHERE is_online = true
		AND last_seen_at < NOW() - $1::interval`, fmt.Sprintf("%d seconds", timeoutSeconds))
	if err != nil {
		log.Printf("heartbeat: query error: %v", err)
		return
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err == nil {
			ids = append(ids, id)
		}
	}

	for _, id := range ids {
		if err := s.MarkOffline(ctx, id); err != nil {
			log.Printf("heartbeat: mark offline %s: %v", id, err)
		} else {
			log.Printf("heartbeat: marked %s offline", id)
		}
	}
}

// UpsertDeviceKeys upserts discovered telemetry keys for a device.
func (s *DeviceService) UpsertDeviceKeys(ctx context.Context, deviceID string, keys map[string]string) error {
	for key, valueType := range keys {
		_, err := s.db.Pool.Exec(ctx, `
			INSERT INTO device_keys (device_id, key, value_type, last_seen_at)
			VALUES ($1, $2, $3, NOW())
			ON CONFLICT (device_id, key) DO UPDATE SET
				last_seen_at = NOW(),
				value_type = EXCLUDED.value_type`,
			deviceID, key, valueType)
		if err != nil {
			return fmt.Errorf("upsert device key %s: %w", key, err)
		}
	}
	return nil
}

// unused import guard
var _ = json.Marshal
