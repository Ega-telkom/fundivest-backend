// cmd/api/app.go
package main

import (
	"time"

	swaggo "github.com/gofiber/contrib/v3/swaggo"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/helmet"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"go.uber.org/zap"

	"github.com/Ega-telkom/fundivest-backend/internal/config"
	"github.com/Ega-telkom/fundivest-backend/internal/handler"
	"github.com/Ega-telkom/fundivest-backend/internal/pubsub"
	repoPostgres "github.com/Ega-telkom/fundivest-backend/internal/repository/postgres"
	repoValkey "github.com/Ega-telkom/fundivest-backend/internal/repository/valkey"
	"github.com/Ega-telkom/fundivest-backend/internal/service"
)

func SetupApp(cfg *config.Config, infra *Infrastructure, logger *zap.Logger) *fiber.App {
    // Init repositories
    certRepo := repoPostgres.NewCertificateRepo(infra.DB)
    sessionRepo := repoValkey.NewSessionRepo(infra.ValkeyClient, cfg.SessionTTL)

    // Init services
    certSvc := service.NewCertificateService(
        certRepo,
        sessionRepo,
        infra.QueuePublisher,
        infra.FileStorage,
    )
    sessionSvc := service.NewSessionService(sessionRepo, cfg.SessionTTL)

    pubsub := pubsub.NewRedisPubSub(infra.ValkeyClient)
    
    // Init handlers
    certHandler := handler.NewCertificateHandler(certSvc, pubsub, logger)
    sessionHandler := handler.NewSessionHandler(sessionSvc)

    // Setup Fiber
    app := fiber.New(fiber.Config{
        AppName:      "fundivest REST API",
        ErrorHandler: customErrorHandler(logger),
    })

    // Middlewares
    app.Use(recover.New())
    app.Use(cors.New(cors.Config{
       	AllowOrigins: []string{cfg.AllowedOrigins},
    }))
    app.Use(helmet.New())
    app.Use(fiberLogger(logger))
   
    if !cfg.IsProduction() {
    	app.Get("/swagger/*", swaggo.HandlerDefault)
    	logger.Info("Swagger UI is enabled, set environment to production to disable")
    }

    // Routes
    setupRoutes(app, sessionHandler, certHandler)

    return app
}

func setupRoutes(
    app *fiber.App,
    sessionHandler *handler.SessionHandler,
    certHandler *handler.CertificateHandler,
) {
    // Health check
    app.Get("/health", func(c fiber.Ctx) error {
        return c.JSON(fiber.Map{"status": "ok"})
    })

    // API routes
    api := app.Group("/api")

    // v1
    v1 := api.Group("/v1")

    // Session routes
    sess := v1.Group("/sessions")
    sess.Post("/", sessionHandler.CreateSession)
    sess.Post("/:id/chapters/:chapter/complete", sessionHandler.CompleteChapter)
    sess.Post("/:id/certificate", certHandler.RequestCertificate)

    // Certificate routes
    cert := v1.Group("/certificates")
    cert.Get("/:certificate_id/status", certHandler.GetStatus)
    cert.Get("/:certificate_id/stream", certHandler.StreamStatus)
    cert.Get("/:certificate_id/download", certHandler.Download)
    cert.Get("/:certificate_id/verify", certHandler.Verify)
}

func customErrorHandler(logger *zap.Logger) fiber.ErrorHandler {
    return func(c fiber.Ctx, err error) error {
        code := fiber.StatusInternalServerError
        if e, ok := err.(*fiber.Error); ok {
            code = e.Code
        }

        logger.Error("Request error",
            zap.String("Path", c.Path()),
            zap.String("Method", c.Method()),
            zap.Int("Status", code),
            zap.Error(err),
        )

        return c.Status(code).JSON(fiber.Map{
            "error": err.Error(),
        })
    }
}

func fiberLogger(logger *zap.Logger) fiber.Handler {
    return func(c fiber.Ctx) error {
        start := time.Now()

        err := c.Next()

        logger.Info("Request",
            zap.String("Method", c.Method()),
            zap.String("Path", c.Path()),
            zap.Int("Status", c.Response().StatusCode()),
            zap.Duration("Latency", time.Since(start)),
            zap.String("IP", c.IP()),
        )

        return err
    }
}
