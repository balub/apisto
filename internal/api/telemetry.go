package api

import (
	"context"
	"time"

	"github.com/balub/apisto/internal/services"
	"github.com/gofiber/fiber/v2"
)

type telemetryHandlers struct {
	telemetrySvc *services.TelemetryService
	deviceSvc    *services.DeviceService
}

// ingest handles HTTP telemetry ingestion via device token
func (h *telemetryHandlers) ingest(c *fiber.Ctx) error {
	token := c.Params("token")

	device, err := h.deviceSvc.GetByToken(context.Background(), token)
	if err != nil {
		return errResponse(c, 401, "invalid device token", "INVALID_TOKEN")
	}

	if err := h.telemetrySvc.Ingest(context.Background(), device.ID, c.Body()); err != nil {
		return errResponse(c, 400, "failed to ingest telemetry: "+err.Error(), "INVALID_INPUT")
	}

	return c.Status(202).JSON(fiber.Map{"status": "accepted"})
}

// query handles time-series telemetry queries
func (h *telemetryHandlers) query(c *fiber.Ctx) error {
	id := c.Params("id")

	params := services.TelemetryQueryParams{
		DeviceID:  id,
		Key:       c.Query("key"),
		Aggregate: c.Query("aggregate"),
		Interval:  c.Query("interval", "1h"),
		Order:     c.Query("order", "asc"),
		Limit:     c.QueryInt("limit", 1000),
	}

	if fromStr := c.Query("from"); fromStr != "" {
		if t, err := time.Parse(time.RFC3339, fromStr); err == nil {
			params.From = t
		}
	}
	if toStr := c.Query("to"); toStr != "" {
		if t, err := time.Parse(time.RFC3339, toStr); err == nil {
			params.To = t
		}
	}

	data, err := h.telemetrySvc.Query(context.Background(), params)
	if err != nil {
		return errResponse(c, 500, "query failed: "+err.Error(), "INTERNAL_ERROR")
	}

	return c.JSON(fiber.Map{
		"device_id": id,
		"key":       params.Key,
		"aggregate": params.Aggregate,
		"interval":  params.Interval,
		"data":      data,
	})
}

// latest returns the most recent value for each key
func (h *telemetryHandlers) latest(c *fiber.Ctx) error {
	id := c.Params("id")

	data, err := h.telemetrySvc.Latest(context.Background(), id)
	if err != nil {
		return errResponse(c, 500, "query failed", "INTERNAL_ERROR")
	}

	// Transform slice to map keyed by key name
	result := make(map[string]interface{})
	for _, lt := range data {
		result[lt.Key] = fiber.Map{
			"value":      lt.Value,
			"value_type": lt.ValueType,
			"time":       lt.Time,
		}
	}

	return c.JSON(fiber.Map{
		"device_id": id,
		"data":      result,
	})
}
