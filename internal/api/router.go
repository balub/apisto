package api

import (
	"github.com/balub/apisto/internal/database"
	"github.com/balub/apisto/internal/services"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

type Services struct {
	DB           *database.DB
	DeviceSvc    *services.DeviceService
	TelemetrySvc *services.TelemetryService
	CommandSvc   *services.CommandService
	Realtime     *services.RealtimeService
}

func SetupRoutes(app *fiber.App, svc Services, corsOrigins string) {
	setupMiddleware(app, corsOrigins)

	projects := &projectHandlers{db: svc.DB}
	devices := &deviceHandlers{deviceSvc: svc.DeviceSvc}
	telemetry := &telemetryHandlers{telemetrySvc: svc.TelemetrySvc, deviceSvc: svc.DeviceSvc}
	commands := &commandHandlers{commandSvc: svc.CommandSvc}
	ws := &wsHandlers{realtime: svc.Realtime, deviceSvc: svc.DeviceSvc}
	shares := &shareHandlers{
		db:           svc.DB,
		telemetrySvc: svc.TelemetrySvc,
		deviceSvc:    svc.DeviceSvc,
		realtime:     svc.Realtime,
	}

	v1 := app.Group("/api/v1")

	// Projects
	v1.Post("/projects", projects.create)
	v1.Get("/projects", projects.list)
	v1.Get("/projects/:id", projects.get)
	v1.Put("/projects/:id", projects.update)
	v1.Delete("/projects/:id", projects.delete)

	// Devices
	v1.Post("/projects/:id/devices", devices.create)
	v1.Get("/projects/:id/devices", devices.listByProject)
	v1.Get("/devices/:id", devices.get)
	v1.Put("/devices/:id", devices.update)
	v1.Delete("/devices/:id", devices.delete)
	v1.Get("/devices/:id/keys", devices.getKeys)

	// Telemetry
	v1.Post("/devices/:token/telemetry", telemetry.ingest)
	v1.Get("/devices/:id/telemetry", telemetry.query)
	v1.Get("/devices/:id/telemetry/latest", telemetry.latest)

	// Commands
	v1.Post("/devices/:id/commands", commands.send)
	v1.Get("/devices/:id/commands", commands.list)

	// WebSocket
	v1.Get("/devices/:id/ws", ws.upgrade, websocket.New(ws.handle))

	// Dashboard shares
	v1.Post("/devices/:id/share", shares.create)
	v1.Delete("/shares/:share_token", shares.revoke)
	v1.Get("/public/:share_token", shares.publicData)
	v1.Get("/public/:share_token/ws", shares.publicWS, websocket.New(shares.publicWSHandle))
}
