package api

import (
	"context"

	"github.com/balub/apisto/internal/services"
	"github.com/gofiber/fiber/v2"
)

type commandHandlers struct {
	commandSvc *services.CommandService
}

func (h *commandHandlers) send(c *fiber.Ctx) error {
	deviceID := c.Params("id")
	var body struct {
		Command string `json:"command"`
		Payload string `json:"payload"`
	}
	if err := c.BodyParser(&body); err != nil {
		return errResponse(c, 400, "invalid request body", "INVALID_INPUT")
	}
	if body.Command == "" {
		return errResponse(c, 400, "command is required", "INVALID_INPUT")
	}

	cmd, err := h.commandSvc.Send(context.Background(), deviceID, body.Command, body.Payload)
	if err != nil {
		return errResponse(c, 500, "failed to send command: "+err.Error(), "INTERNAL_ERROR")
	}
	return c.Status(201).JSON(cmd)
}

func (h *commandHandlers) list(c *fiber.Ctx) error {
	deviceID := c.Params("id")
	limit := c.QueryInt("limit", 50)

	cmds, err := h.commandSvc.List(context.Background(), deviceID, limit)
	if err != nil {
		return errResponse(c, 500, "failed to list commands", "INTERNAL_ERROR")
	}
	if cmds == nil {
		return c.JSON([]interface{}{})
	}
	return c.JSON(cmds)
}
