package server

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	fiberlogger "github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"github.com/Delpak/shahrah/internal/auth"
	"github.com/Delpak/shahrah/internal/cache"
	"github.com/Delpak/shahrah/internal/config"
	"github.com/Delpak/shahrah/internal/proxy"
	"github.com/Delpak/shahrah/internal/router"

	"go.uber.org/zap"
)

// NewFiberApp creates and configures a new Fiber application.
func NewFiberApp(cfg *config.Config, llmRouter *router.Router, c cache.Cache, logger *zap.Logger) *fiber.App {
	app := fiber.New(fiber.Config{
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
		ErrorHandler: customErrorHandler,
	})

	// Middleware chain
	app.Use(fiberlogger.New(fiberlogger.Config{
		Format: "${time} | ${status} | ${latency} | ${ip} | ${method} | ${path}\n",
	}))
	app.Use(recover.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET, POST, OPTIONS",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization, api-key, x-api-key",
	}))

	// Auth middleware
	keyStore := auth.NewInMemoryKeyStore() // Replace with DB-backed store
	authMiddleware := auth.NewAuthMiddleware(keyStore, cfg.GeneralSettings.MasterKey, logger)
	app.Use(authMiddleware.Middleware())

	// Register routes
	registerRoutes(app, llmRouter, c, logger)

	return app
}

func registerRoutes(app *fiber.App, llmRouter *router.Router, c cache.Cache, logger *zap.Logger) {
	// Health check
	app.Get("/health", healthHandler)
	app.Get("/ready", healthHandler)

	// OpenAI-compatible endpoints
	v1 := app.Group("/v1")

	// Chat completions
	v1.Post("/chat/completions", proxy.NewChatCompletionHandler(llmRouter, c, logger).Handle)

	// Model list
	v1.Get("/models", listModelsHandler)
}

func healthHandler(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status":  "healthy",
		"version": "0.1.0",
	})
}

func listModelsHandler(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"object": "list",
		"data":   []fiber.Map{}, // Would return registered models
	})
}

func customErrorHandler(ctx *fiber.Ctx, err error) error {
	// Default 500 status code
	code := fiber.StatusInternalServerError
	msg := "Internal Server Error"

	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
		msg = e.Message
	}

	return ctx.Status(code).JSON(fiber.Map{
		"error": map[string]interface{}{
			"message": msg,
			"type":    "server_error",
		},
	})
}

