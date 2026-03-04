package api

import (
	"log"

	"github.com/balub/apisto/internal/services"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

type wsHandlers struct {
	realtime  *services.RealtimeService
	deviceSvc *services.DeviceService
}

func (h *wsHandlers) upgrade(c *fiber.Ctx) error {
	if websocket.IsWebSocketUpgrade(c) {
		return c.Next()
	}
	return fiber.ErrUpgradeRequired
}

func (h *wsHandlers) handle(c *websocket.Conn) {
	deviceID := c.Params("id")

	h.realtime.Register(deviceID, c)
	defer func() {
		h.realtime.Unregister(deviceID, c)
		c.Close()
	}()

	log.Printf("ws: client connected for device %s", deviceID)

	// Keep connection alive; reads are used only to detect disconnects
	for {
		if _, _, err := c.ReadMessage(); err != nil {
			log.Printf("ws: client disconnected from device %s: %v", deviceID, err)
			break
		}
	}
}
