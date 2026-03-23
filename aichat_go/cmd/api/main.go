package main

import (
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/gofiber/fiber/v3/middleware/requestid"
	"github.com/joho/godotenv"

	"aichat_go/internal/api"
	"aichat_go/internal/config"
	"aichat_go/internal/observability"
)

func main() {
	_ = godotenv.Load()
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	logger := observability.NewLogger(cfg.AppEnv)

	app := fiber.New(fiber.Config{
		ErrorHandler:      api.ErrorHandler(logger),
		StreamRequestBody: true,
	})

	app.Use(recover.New())
	app.Use(requestid.New())
	app.Use(observability.RequestLogger(logger))
	app.Use(cors.New(cors.Config{
		AllowOrigins:     strings.Split(cfg.CORSAllowOrigins, ","),
		AllowMethods:     []string{"GET", "POST", "OPTIONS"},
		AllowHeaders:     []string{"Authorization", "Content-Type", "X-Tenant-ID", "X-Workspace-ID"},
		AllowCredentials: true,
	}))

	app.Get("/health", api.Health)

	v1 := app.Group("/v1")
	api.RegisterV1(v1, cfg, logger)

	addr := ":" + cfg.Port

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		logger.Info("shutting down")
		if err := app.Shutdown(); err != nil {
			logger.Error("shutdown error", "error", err)
		}
	}()

	logger.Info("listening", "addr", addr, "env", cfg.AppEnv)
	if err := app.Listen(addr, fiber.ListenConfig{
		DisableStartupMessage: cfg.AppEnv != "development",
	}); err != nil {
		log.Fatalf("listen: %v", err)
	}
}
