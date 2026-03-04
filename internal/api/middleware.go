package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func setupMiddleware(app *fiber.App, corsOrigins string) {
	app.Use(recover.New())
	app.Use(logger.New(logger.Config{
		Format: "${time} ${method} ${path} ${status} ${latency}\n",
	}))
	app.Use(cors.New(cors.Config{
		AllowOrigins: corsOrigins,
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Content-Type,Authorization",
	}))
}

func errResponse(c *fiber.Ctx, status int, message, code string) error {
	return c.Status(status).JSON(fiber.Map{
		"error": message,
		"code":  code,
	})
}
