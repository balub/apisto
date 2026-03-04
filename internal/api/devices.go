package api

import (
	"context"

	"github.com/balub/apisto/internal/models"
	"github.com/balub/apisto/internal/services"
	"github.com/gofiber/fiber/v2"
)

type deviceHandlers struct {
	deviceSvc *services.DeviceService
}

func (h *deviceHandlers) create(c *fiber.Ctx) error {
	projectID := c.Params("id")
	var body struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := c.BodyParser(&body); err != nil {
		return errResponse(c, 400, "invalid request body", "INVALID_INPUT")
	}
	if body.Name == "" {
		body.Name = "Unnamed Device"
	}

	device, err := h.deviceSvc.Create(context.Background(), projectID, body.Name, body.Description)
	if err != nil {
		return errResponse(c, 500, "failed to create device", "INTERNAL_ERROR")
	}
	return c.Status(201).JSON(device)
}

func (h *deviceHandlers) listByProject(c *fiber.Ctx) error {
	projectID := c.Params("id")
	devices, err := h.deviceSvc.ListByProject(context.Background(), projectID)
	if err != nil {
		return errResponse(c, 500, "failed to list devices", "INTERNAL_ERROR")
	}
	if devices == nil {
		devices = []*models.Device{}
	}
	return c.JSON(devices)
}

func (h *deviceHandlers) get(c *fiber.Ctx) error {
	id := c.Params("id")
	device, err := h.deviceSvc.GetByID(context.Background(), id)
	if err != nil {
		return errResponse(c, 404, "device not found", "NOT_FOUND")
	}
	// Don't expose token in get response (except on creation)
	device.Token = ""
	return c.JSON(device)
}

func (h *deviceHandlers) update(c *fiber.Ctx) error {
	id := c.Params("id")
	var body struct {
		Name            string `json:"name"`
		Description     string `json:"description"`
		FirmwareVersion string `json:"firmware_version"`
	}
	if err := c.BodyParser(&body); err != nil {
		return errResponse(c, 400, "invalid request body", "INVALID_INPUT")
	}

	device, err := h.deviceSvc.Update(context.Background(), id, body.Name, body.Description, body.FirmwareVersion)
	if err != nil {
		return errResponse(c, 404, "device not found", "NOT_FOUND")
	}
	device.Token = ""
	return c.JSON(device)
}

func (h *deviceHandlers) delete(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := h.deviceSvc.Delete(context.Background(), id); err != nil {
		return errResponse(c, 404, "device not found", "NOT_FOUND")
	}
	return c.SendStatus(204)
}

func (h *deviceHandlers) getKeys(c *fiber.Ctx) error {
	id := c.Params("id")
	keys, err := h.deviceSvc.GetKeys(context.Background(), id)
	if err != nil {
		return errResponse(c, 500, "failed to get keys", "INTERNAL_ERROR")
	}
	if keys == nil {
		keys = []*models.DeviceKey{}
	}
	return c.JSON(keys)
}
