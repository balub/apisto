package api

import (
	"context"
	"log"

	"github.com/balub/apisto/internal/database"
	"github.com/balub/apisto/internal/services"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

type shareHandlers struct {
	db           *database.DB
	telemetrySvc *services.TelemetryService
	deviceSvc    *services.DeviceService
	realtime     *services.RealtimeService
}

func (h *shareHandlers) create(c *fiber.Ctx) error {
	deviceID := c.Params("id")

	var share struct {
		ID          string `json:"id"`
		DeviceID    string `json:"device_id"`
		ShareToken  string `json:"share_token"`
		IsActive    bool   `json:"is_active"`
		CreatedAt   string `json:"created_at"`
	}

	err := h.db.Pool.QueryRow(context.Background(), `
		INSERT INTO dashboard_shares (device_id)
		VALUES ($1)
		RETURNING id, device_id, share_token, is_active, created_at::text`,
		deviceID,
	).Scan(&share.ID, &share.DeviceID, &share.ShareToken, &share.IsActive, &share.CreatedAt)
	if err != nil {
		return errResponse(c, 500, "failed to create share", "INTERNAL_ERROR")
	}

	return c.Status(201).JSON(fiber.Map{
		"id":          share.ID,
		"share_token": share.ShareToken,
		"is_active":   share.IsActive,
		"created_at":  share.CreatedAt,
	})
}

func (h *shareHandlers) revoke(c *fiber.Ctx) error {
	shareToken := c.Params("share_token")
	cmd, err := h.db.Pool.Exec(context.Background(),
		`UPDATE dashboard_shares SET is_active = false WHERE share_token = $1`, shareToken)
	if err != nil || cmd.RowsAffected() == 0 {
		return errResponse(c, 404, "share not found", "NOT_FOUND")
	}
	return c.SendStatus(204)
}

func (h *shareHandlers) publicData(c *fiber.Ctx) error {
	shareToken := c.Params("share_token")

	var deviceID string
	err := h.db.Pool.QueryRow(context.Background(), `
		SELECT device_id FROM dashboard_shares
		WHERE share_token = $1 AND is_active = true`, shareToken,
	).Scan(&deviceID)
	if err != nil {
		return errResponse(c, 404, "share not found or expired", "NOT_FOUND")
	}

	device, err := h.deviceSvc.GetByID(context.Background(), deviceID)
	if err != nil {
		return errResponse(c, 404, "device not found", "NOT_FOUND")
	}
	device.Token = ""

	latest, err := h.telemetrySvc.Latest(context.Background(), deviceID)
	if err != nil {
		latest = nil
	}

	keys, err := h.deviceSvc.GetKeys(context.Background(), deviceID)
	if err != nil {
		keys = nil
	}

	return c.JSON(fiber.Map{
		"device":  device,
		"latest":  latest,
		"keys":    keys,
	})
}

func (h *shareHandlers) publicWS(c *fiber.Ctx) error {
	if websocket.IsWebSocketUpgrade(c) {
		shareToken := c.Params("share_token")
		var deviceID string
		err := h.db.Pool.QueryRow(context.Background(), `
			SELECT device_id FROM dashboard_shares
			WHERE share_token = $1 AND is_active = true`, shareToken,
		).Scan(&deviceID)
		if err != nil {
			return errResponse(c, 404, "share not found", "NOT_FOUND")
		}
		c.Locals("device_id", deviceID)
		return c.Next()
	}
	return fiber.ErrUpgradeRequired
}

func (h *shareHandlers) publicWSHandle(c *websocket.Conn) {
	deviceID, _ := c.Locals("device_id").(string)
	if deviceID == "" {
		c.Close()
		return
	}

	h.realtime.Register(deviceID, c)
	defer func() {
		h.realtime.Unregister(deviceID, c)
		c.Close()
	}()

	log.Printf("ws: public client connected for device %s", deviceID)
	for {
		if _, _, err := c.ReadMessage(); err != nil {
			break
		}
	}
}
