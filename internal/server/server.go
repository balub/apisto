package server

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/balub/apisto/internal/api"
	"github.com/balub/apisto/internal/config"
	"github.com/balub/apisto/internal/database"
	apistomqtt "github.com/balub/apisto/internal/mqtt"
	"github.com/balub/apisto/internal/services"
	"github.com/gofiber/fiber/v2"
)

func Run(cfg *config.Config) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Connect to database
	log.Println("server: connecting to database...")
	db, err := database.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("db connect: %w", err)
	}
	defer db.Close()
	log.Println("server: database connected")

	// Run migrations
	log.Println("server: running migrations...")
	if err := db.RunMigrations(cfg.DatabaseURL); err != nil {
		return fmt.Errorf("migrations: %w", err)
	}
	log.Println("server: migrations complete")

	// Initialize services
	realtime := services.NewRealtimeService()
	deviceSvc := services.NewDeviceService(db, realtime)
	telemetrySvc := services.NewTelemetryService(db, realtime, deviceSvc)
	commandSvc := services.NewCommandService(db, realtime, nil) // MQTT publish set after client is created

	// MQTT token lookup function
	lookupToken := func(ctx context.Context, token string) (string, error) {
		device, err := deviceSvc.GetByToken(ctx, token)
		if err != nil {
			return "", err
		}
		return device.ID, nil
	}

	// MQTT handlers
	mqttHandlers := apistomqtt.Handlers{
		OnTelemetry: func(ctx context.Context, deviceID string, payload []byte) {
			if err := telemetrySvc.Ingest(ctx, deviceID, payload); err != nil {
				log.Printf("mqtt: telemetry ingest error for device %s: %v", deviceID, err)
			}
		},
		OnStatus: func(ctx context.Context, deviceID string, payload []byte) {
			if err := deviceSvc.MarkOnline(ctx, deviceID, ""); err != nil {
				log.Printf("mqtt: mark online error for device %s: %v", deviceID, err)
			}
		},
		OnCommandAck: func(ctx context.Context, commandID string) {
			commandSvc.HandleAck(ctx, commandID)
		},
	}

	// Connect to MQTT
	log.Println("server: connecting to MQTT broker...")
	mqttClient, err := apistomqtt.NewClient(cfg.MQTTBrokerURL, cfg.MQTTClientID, lookupToken, mqttHandlers)
	if err != nil {
		return fmt.Errorf("mqtt connect: %w", err)
	}
	defer mqttClient.Disconnect()
	log.Println("server: MQTT connected")

	// Wire MQTT publish into command service
	commandSvc = services.NewCommandService(db, realtime, mqttClient.Publish)

	// Start heartbeat checker
	deviceSvc.StartHeartbeatChecker(ctx, cfg.HeartbeatTimeout)

	// Set up HTTP server
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"error": err.Error(),
				"code":  "INTERNAL_ERROR",
			})
		},
	})

	api.SetupRoutes(app, api.Services{
		DB:           db,
		DeviceSvc:    deviceSvc,
		TelemetrySvc: telemetrySvc,
		CommandSvc:   commandSvc,
		Realtime:     realtime,
	}, cfg.CORSOrigins)

	// Serve frontend static files (only if built dist exists)
	if _, err := os.Stat("./web/dist"); err == nil {
		app.Static("/", "./web/dist", fiber.Static{
			Index: "index.html",
		})
		// SPA fallback: serve index.html for all non-API routes
		app.Get("/*", func(c *fiber.Ctx) error {
			return c.SendFile("./web/dist/index.html")
		})
		log.Println("server: serving frontend from ./web/dist")
	} else {
		log.Println("server: web/dist not found — API only mode. Run 'cd web && npm run build' or use 'npm run dev' for the frontend.")
		app.Get("/", func(c *fiber.Ctx) error {
			return c.Status(200).SendString("Apisto API running. Frontend not built. See /api/v1/ for API endpoints.")
		})
	}

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		log.Println("server: shutting down...")
		cancel()

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()
		app.ShutdownWithContext(shutdownCtx)
	}()

	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("server: listening on %s", addr)
	return app.Listen(addr)
}
